package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	neturl "net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"context"

	"github.com/blang/semver"
	"github.com/codegangsta/cli"
	dockerclient "github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/howtowhale/dvm/dvm-helper/dockerversion"
	"github.com/howtowhale/dvm/dvm-helper/url"
	"github.com/ryanuber/go-glob"
	"golang.org/x/oauth2"
)

// These are global command line variables
var shell string
var dvmDir string
var mirrorURL string
var githubUrlOverride string
var debug bool
var silent bool
var nocheck bool
var token string

// These are set during the build
var dvmVersion string
var dvmCommit string
var upgradeDisabled string // Allow package managers like homebrew to disable in-place upgrades

const (
	retCodeInvalidArgument  = 127
	retCodeInvalidOperation = 3
	retCodeRuntimeError     = 1
	versionEnvVar           = "DOCKER_VERSION"
)

func main() {
	app := cli.NewApp()
	app.Name = "Docker Version Manager"
	app.Usage = "Manage multiple versions of the Docker client"
	app.Version = fmt.Sprintf("%s (%s)", dvmVersion, dvmCommit)
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "github-token", EnvVar: "GITHUB_TOKEN", Usage: "Increase the github api rate limit by specifying your github personal access token."},
		cli.StringFlag{Name: "dvm-dir", EnvVar: "DVM_DIR", Usage: "Specify an alternate DVM home directory, defaults to ~/.dvm."},
		cli.StringFlag{Name: "shell", EnvVar: "SHELL", Usage: "Specify the shell format in which environment variables should be output, e.g. powershell, cmd or sh/bash. Defaults to sh/bash."},
		cli.BoolFlag{Name: "debug", Usage: "Print additional debug information."},
		cli.BoolFlag{Name: "silent", EnvVar: "DVM_SILENT", Usage: "Suppress output. Errors will still be displayed."},
	}
	app.Commands = []cli.Command{
		{
			Name:  "detect",
			Usage: "Detect the appropriate Docker client version",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)
				detect()
				return nil
			},
		},
		{
			Name:    "install",
			Aliases: []string{"i"},
			Usage:   "dvm install [<version>], dvm install experimental\n\tInstall a Docker version, using $DOCKER_VERSION if the version is not specified.",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "mirror-url", EnvVar: "DVM_MIRROR_URL", Usage: "Specify an alternate URL from which to download the Docker client. Defaults to https://get.docker.com/builds"},
				cli.BoolFlag{Name: "nocheck", EnvVar: "DVM_NOCHECK", Usage: "Do not check if version exists (use with caution)."},
			},
			Action: func(c *cli.Context) error {
				setGlobalVars(c)

				versionName := c.Args().First()
				if versionName == "" {
					versionName = getDockerVersionVar()
				}
				version := dockerversion.Parse(versionName)

				install(version)
				return nil
			},
		},
		{
			Name:  "uninstall",
			Usage: "dvm uninstall <version>\n\tUninstall a Docker version.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)
				version := dockerversion.Parse(c.Args().First())
				uninstall(version)
				return nil
			},
		},
		{
			Name:  "use",
			Usage: "dvm use [<version>], dvm use system, dvm use experimental\n\tUse a Docker version, using $DOCKER_VERSION if the version is not specified.",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "mirror-url", EnvVar: "DVM_MIRROR_URL", Usage: "Specify an alternate URL from which to download the Docker client. Defaults to https://get.docker.com/builds"},
				cli.BoolFlag{Name: "nocheck", EnvVar: "DVM_NOCHECK", Usage: "Do not check if version exists (use with caution)."},
			},
			Action: func(c *cli.Context) error {
				setGlobalVars(c)

				versionName := c.Args().First()
				if versionName == "" {
					versionName = getDockerVersionVar()
				}
				version := dockerversion.Parse(versionName)

				use(version)
				return nil
			},
		},
		{
			Name:  "deactivate",
			Usage: "dvm deactivate\n\tUndo the effects of `dvm` on current shell.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)
				deactivate()
				return nil
			},
		},
		{
			Name:  "current",
			Usage: "dvm current\n\tPrint the current Docker version.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)
				current()
				return nil
			},
		},
		{
			Name:  "which",
			Usage: "dvm which\n\tPrint the path to the current Docker version.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)
				which()
				return nil
			},
		},
		{
			Name:  "alias",
			Usage: "dvm alias <alias> <version>\n\tCreate an alias to a Docker version.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)
				name := c.Args().Get(0)
				version := dockerversion.Parse(c.Args().Get(1))
				alias(name, version)
				return nil
			},
		},
		{
			Name:  "unalias",
			Usage: "dvm unalias <alias>\n\tRemove a Docker version alias.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)
				name := c.Args().First()
				unalias(name)
				return nil
			},
		},
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "dvm list [<pattern>]\n\tList installed Docker versions.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)
				list(c.Args().First())
				return nil
			},
		},
		{
			Name:    "list-remote",
			Aliases: []string{"ls-remote"},
			Usage:   "dvm list-remote [<pattern>]\n\tList available Docker versions.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)
				listRemote(c.Args().First())
				return nil
			},
		},
		{
			Name:    "list-alias",
			Aliases: []string{"ls-alias"},
			Usage:   "dvm list-alias\n\tList Docker version aliases.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)
				listAlias()
				return nil
			},
		},
	}

	if upgradeDisabled != "true" {
		app.Commands = append(app.Commands, cli.Command{
			Name:  "upgrade",
			Usage: "dvm upgrade\n\tUpgrade dvm to the latest release.",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "check", Usage: "Checks if an newer version of dvm is available, but does not perform the upgrade."},
				cli.StringFlag{Name: "version", Usage: "Upgrade to the specified version."},
			},
			Action: func(c *cli.Context) error {
				setGlobalVars(c)
				upgrade(c.Bool("check"), c.String("version"))
				return nil
			},
		})
	}

	app.Run(os.Args)
}

func setGlobalVars(c *cli.Context) {
	debug = c.GlobalBool("debug")
	nocheck = c.Bool("nocheck")
	token = c.GlobalString("github-token")
	shell = c.GlobalString("shell")
	validateShellFlag()

	silent = c.GlobalBool("silent")
	mirrorURL = c.String("mirror-url")

	dvmDir = c.GlobalString("dvm-dir")
	if dvmDir == "" {
		dvmDir = filepath.Join(getUserHomeDir(), ".dvm")
	}
	writeDebug("The dvm home directory is: %s", dvmDir)
}

func detect() {
	docker, err := dockerclient.NewEnvClient()
	if err != nil {
		die("Cannot build a docker client from environment variables", err, retCodeRuntimeError)
	}

	versionResult, err := docker.ServerVersion(context.Background())
	if err != nil {
		die("Unable to query docker version", err, retCodeRuntimeError)
	}

	writeDebug("Queried /version and got Version: %s", versionResult.Version)
	version, err := semver.Parse(versionResult.Version)

	// Docker versions prior to 1.12 don't return a usable client version
	// Lookup the client version from the API version
	if err != nil {
		writeDebug("Attempting to lookup a client version for API version: %s", versionResult.APIVersion)

		// api version -> client version range
		oldVersionMap := map[string]string{
			"1.23": "1.11.x",
			"1.22": "1.10.x",
			"1.21": "1.9.x",
			"1.20": "1.8.x",
			"1.19": "1.7.x",
			"1.18": "1.6.x",
		}
		clientVersion, found := oldVersionMap[versionResult.APIVersion]
		if !found {
			die("Unable to detect the proper client version for Docker API version %s", nil, retCodeRuntimeError, versionResult.APIVersion)
		}

		// Find the highest version that statisfies the client version range
		clientRange := semver.MustParseRange(clientVersion)
		availableVersions := getAvailableVersions("")
		for i := len(availableVersions) - 1; i >= 0; i-- {
			v := availableVersions[i]
			if clientRange(v.SemVer) {
				version = v.SemVer
				break
			}
		}
		if version.Equals(semver.Version{}) {
			die("Unable to detect the proper client version for Docker client version %s", nil, retCodeRuntimeError, clientVersion)
		}
	}
	writeDebug("Detected client version: %s", version)

	os.Setenv(versionEnvVar, version.String())
	writeEnvironmentVariableScript(versionEnvVar)
	use(dockerversion.New(version))
}

func upgrade(checkOnly bool, version string) {
	if version != "" && dvmVersion == version {
		writeWarning("dvm %s is already installed.", version)
		return
	}

	if version == "" {
		shouldUpgrade, latestVersion := isUpgradeAvailable()
		if !shouldUpgrade {
			writeInfo("The latest version of dvm is already installed.")
			return
		}

		version = latestVersion
	}

	if checkOnly {
		writeInfo("dvm %s is available. Run `dvm upgrade` to install the latest version.", version)
		return
	}

	writeInfo("Upgrading to dvm %s...", version)
	upgradeSelf(version)
}

func buildDvmReleaseURL(version string, elem ...string) string {
	prefix := url.Join("https://download.getcarina.com/dvm", version)
	suffix := url.Join(elem...)
	return url.Join(prefix, suffix)
}

func current() {
	current, err := getCurrentDockerVersion()
	if err != nil {
		writeWarning("N/A")
	} else {
		writeInfo(current.String())
	}
}

func list(pattern string) {
	pattern += "*"
	versions := getInstalledVersions(pattern)
	current, _ := getCurrentDockerVersion()

	for _, version := range versions {
		if current.Equals(version) {
			color.Green("->\t%s", version)
		} else {
			writeInfo("\t%s", version)
		}
	}
}

func install(version dockerversion.Version) {
	writeDebug("dvm install %s", version)

	if version.IsEmpty() {
		die("The install command requires that a version is specified or the DOCKER_VERSION environment variable is set.", nil, retCodeInvalidArgument)
	}

	if nocheck {
		writeDebug("Skipping version validation!")
	}

	if !nocheck && !versionExists(version) {
		die("Version %s not found - try `dvm ls-remote` to browse available versions.", nil, retCodeInvalidOperation, version)
	}

	versionDir := getVersionDir(version)

	if version.IsExperimental() && pathExists(versionDir) {
		// Always install latest of experimental build
		err := os.RemoveAll(versionDir)
		if err != nil {
			die("Unable to remove experimental version at %s.", err, retCodeRuntimeError, versionDir)
		}
	}

	if _, err := os.Stat(versionDir); err == nil {
		writeWarning("%s is already installed", version)
		use(version)
		return
	}

	writeInfo("Installing %s...", version)

	downloadRelease(version)
	use(version)
}

func buildDownloadURL(version dockerversion.Version) string {
	dockerVersion := version.SemVer.String()
	if version.IsExperimental() {
		dockerVersion = "latest"
	}

	if mirrorURL == "" {
		mirrorURL = "https://get.docker.com/builds"
		if version.IsExperimental() {
			writeDebug("Downloading from experimental builds mirror")
			mirrorURL = "https://experimental.docker.com/builds"
		}
	}

	// New Docker versions are released in a zip file, vs. the old way of releasing the client binary only
	if version.ShouldUseArchivedRelease() {
		return fmt.Sprintf("%s/%s/%s/docker-%s%s", mirrorURL, dockerOS, dockerArch, dockerVersion, archiveFileExt)
	}

	return fmt.Sprintf("%s/%s/%s/docker-%s%s", mirrorURL, dockerOS, dockerArch, dockerVersion, binaryFileExt)
}

func downloadRelease(version dockerversion.Version) {
	url := buildDownloadURL(version)
	binaryName := getBinaryName()
	binaryPath := filepath.Join(getVersionDir(version), binaryName)
	if version.ShouldUseArchivedRelease() {
		archivedFile := path.Join("docker", binaryName)
		downloadArchivedFileWithChecksum(url, archivedFile, binaryPath)
	} else {
		downloadFileWithChecksum(url, binaryPath)
	}
	writeDebug("Downloaded Docker %s to %s.", version, binaryPath)
}

func uninstall(version dockerversion.Version) {
	if version.IsEmpty() {
		die("The uninstall command requires that a version is specified.", nil, retCodeInvalidArgument)
	}

	current, _ := getCurrentDockerVersion()
	if current.Equals(version) {
		die("Cannot uninstall the currently active Docker version.", nil, retCodeInvalidOperation)
	}

	versionDir := getVersionDir(version)
	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		writeWarning("%s is not installed.", version)
		return
	}

	err := os.RemoveAll(versionDir)
	if err != nil {
		die("Unable to uninstall Docker version %s located in %s.", err, retCodeRuntimeError, version, versionDir)
	}

	writeInfo("Uninstalled Docker %s.", version)
}

func use(version dockerversion.Version) {
	writeDebug("dvm use %s", version)

	if version.IsEmpty() {
		die("The use command requires that a version is specified or the DOCKER_VERSION environment variable is set.", nil, retCodeInvalidOperation)
	}

	if version.HasAlias() && aliasExists(version.Alias) {
		aliasedVersion, _ := ioutil.ReadFile(getAliasPath(version.Alias))
		version.SemVer = semver.MustParse(string(aliasedVersion))
		writeDebug("Using alias: %s -> %s", version.Alias, version.SemVer)
	}

	ensureVersionIsInstalled(version)

	if version.IsSystem() {
		version, _ = getSystemDockerVersion()
	} else if version.IsExperimental() {
		version, _ = getExperimentalDockerVersion()
	}

	removePreviousDockerVersionFromPath()
	if !version.IsSystem() {
		prependDockerVersionToPath(version)
	}

	writeEnvironmentVariableScript(pathEnvVar)
	writeInfo("Now using Docker %s", version)
}

func which() {
	currentPath, err := getCurrentDockerPath()
	if err == nil {
		writeInfo(currentPath)
	}
}

func alias(alias string, version dockerversion.Version) {
	if alias == "" || version.IsEmpty() {
		die("The alias command requires both an alias name and a version.", nil, retCodeInvalidArgument)
	}

	if !isVersionInstalled(version) {
		die("The aliased version, %s, is not installed.", nil, retCodeInvalidArgument, version)
	}

	aliasPath := getAliasPath(alias)
	if _, err := os.Stat(aliasPath); err == nil {
		writeDebug("Overwriting existing alias.")
	}

	writeFile(aliasPath, version.SemVer.String())
	writeInfo("Aliased %s to %s.", alias, version)
}

func unalias(alias string) {
	if alias == "" {
		die("The unalias command requires an alias name.", nil, retCodeInvalidArgument)
	}

	if !aliasExists(alias) {
		writeWarning("%s is not an alias.", alias)
		return
	}

	aliasPath := getAliasPath(alias)
	err := os.Remove(aliasPath)
	if err != nil {
		die("Unable to remove alias %s at %s.", err, retCodeRuntimeError, alias, aliasPath)
	}

	writeInfo("Removed alias %s", alias)
}

func listAlias() {
	aliases := getAliases()
	for alias, version := range aliases {
		writeInfo("\t%s -> %s", alias, version)
	}
}

func aliasExists(alias string) bool {
	aliasPath := getAliasPath(alias)
	if _, err := os.Stat(aliasPath); err == nil {
		return true
	}

	return false
}

func getAliases() map[string]string {
	aliases, _ := filepath.Glob(getAliasPath("*"))

	results := make(map[string]string)
	for _, aliasPath := range aliases {
		alias := filepath.Base(aliasPath)
		version, err := ioutil.ReadFile(aliasPath)
		if err != nil {
			writeDebug("Excluding alias: %s.", err, retCodeRuntimeError, alias)
			continue
		}

		results[alias] = string(version)
	}

	return results
}

func getAliasPath(alias string) string {
	return filepath.Join(dvmDir, "alias", alias)
}

func getBinaryName() string {
	return "docker" + binaryFileExt
}

func deactivate() {
	removePreviousDockerVersionFromPath()
	writeEnvironmentVariableScript(pathEnvVar)
}

func prependDockerVersionToPath(version dockerversion.Version) {
	prependPath(getVersionDir(version))
}

func writeEnvironmentVariableScript(name string) {
	// Write to a shell script for the calling wrapper to execute which exports the environment variable
	scriptPath := buildDvmOutputScriptPath()
	contents := exportEnvironmentVariable(name)

	writeFile(scriptPath, contents)
}

func buildDvmOutputScriptPath() string {
	var fileExtension string
	if shell == "powershell" {
		fileExtension = "ps1"
	} else if shell == "cmd" {
		fileExtension = "cmd"
	} else { // default to bash
		fileExtension = "sh"
	}
	return filepath.Join(dvmDir, ".tmp", ("dvm-output." + fileExtension))
}

func removePreviousDockerVersionFromPath() {
	removePath(getCleanPathRegex())
}

func ensureVersionIsInstalled(version dockerversion.Version) {
	if isVersionInstalled(version) {
		return
	}

	writeInfo("%s is not installed. Installing now...", version)
	install(version)
}

func isVersionInstalled(version dockerversion.Version) bool {
	writeDebug("Checking if version is installed: %s", version)
	installedVersions := getInstalledVersions("*")

	for _, availableVersion := range installedVersions {
		writeDebug("Checking version: %s", availableVersion)
		if version.Equals(availableVersion) {
			writeDebug("Version %s is installed", version)
			return true
		}
	}

	return false
}

func versionExists(version dockerversion.Version) bool {
	if version.IsExperimental() {
		return true
	}

	availableVersions := getAvailableVersions(version.SemVer.String())

	for _, availableVersion := range availableVersions {
		if version.Equals(availableVersion) {
			return true
		}
	}
	return false
}

func getCurrentDockerPath() (string, error) {
	currentDockerPath, err := exec.LookPath("docker")
	return currentDockerPath, err
}

func getCurrentDockerVersion() (dockerversion.Version, error) {
	currentDockerPath, err := getCurrentDockerPath()
	if err != nil {
		return dockerversion.Version{}, err
	}

	systemDockerPath, _ := getSystemDockerPath()
	experimentalVersionPath, _ := getExperimentalDockerPath()

	isSystem := currentDockerPath == systemDockerPath
	isExperimental := currentDockerPath == experimentalVersionPath

	current, _ := getDockerVersion(currentDockerPath, isExperimental)

	if isSystem {
		writeDebug("The current docker is the system installation")
		current.SetAsSystem()
	}

	if isExperimental {
		writeDebug("The current docker is an experimental version")
		current.SetAsExperimental()
	}

	writeDebug("The current version is: %s", current)
	return current, nil
}

func getSystemDockerPath() (string, error) {
	originalPath := getPath()
	removePreviousDockerVersionFromPath()
	systemDockerPath, err := exec.LookPath("docker")
	setPath(originalPath)
	return systemDockerPath, err
}

func getSystemDockerVersion() (dockerversion.Version, error) {
	systemDockerPath, err := getSystemDockerPath()
	if err != nil {
		return dockerversion.Version{}, err
	}
	version, err := getDockerVersion(systemDockerPath, false)
	version.SetAsSystem()
	return version, err
}

func getExperimentalDockerPath() (string, error) {
	experimentalVersionPath := filepath.Join(getVersionsDir(), dockerversion.ExperimentalAlias, getBinaryName())
	_, err := os.Stat(experimentalVersionPath)
	return experimentalVersionPath, err
}

func getExperimentalDockerVersion() (dockerversion.Version, error) {
	experimentalDockerpath, err := getExperimentalDockerPath()
	if err != nil {
		return dockerversion.Version{}, err
	}
	version, err := getDockerVersion(experimentalDockerpath, true)
	version.SetAsExperimental()
	return version, err
}

func getDockerVersion(dockerPath string, includeBuild bool) (dockerversion.Version, error) {
	rawVersion, _ := exec.Command(dockerPath, "-v").Output()

	writeDebug("%s -v output: %s", dockerPath, rawVersion)

	versionRegex := regexp.MustCompile(`^Docker version (.+), build ([^,]+),?`)
	match := versionRegex.FindSubmatch(rawVersion)
	if len(match) < 2 {
		return dockerversion.Version{}, errors.New("Could not detect docker version.")
	}

	version := string(match[1][:])
	build := string(match[2][:])
	if includeBuild {
		version = fmt.Sprintf("%s+%s", version, build)
	}
	return dockerversion.Parse(version), nil
}

func listRemote(pattern string) {
	versions := getAvailableVersions(pattern)
	for _, version := range versions {
		writeInfo(version.String())
	}
}

func getInstalledVersions(pattern string) []dockerversion.Version {
	searchPath := filepath.Join(getVersionsDir(), pattern)
	versions, _ := filepath.Glob(searchPath)

	var results []dockerversion.Version
	for _, versionDir := range versions {
		version := dockerversion.Parse(filepath.Base(versionDir))

		if version.IsExperimental() {
			experimentalDockerPath := filepath.Join(versionDir, getBinaryName())
			experimentalVersion, err := getDockerVersion(experimentalDockerPath, true)
			if err != nil {
				writeDebug("Unable to get version of installed experimental version at %s.\n%s", versionDir, err)
				continue
			}
			version.SemVer = experimentalVersion.SemVer
		}

		results = append(results, version)
	}

	if glob.Glob(pattern, dockerversion.SystemAlias) {
		systemVersion, err := getSystemDockerVersion()
		if err == nil {
			results = append(results, systemVersion)
		}
	}

	dockerversion.Sort(results)
	return results
}

func getAvailableVersions(pattern string) []dockerversion.Version {
	writeDebug("Retrieving Docker releases")
	gh := buildGithubClient()
	options := &github.ListOptions{PerPage: 100}

	var allReleases []github.RepositoryRelease
	for {
		releases, response, err := gh.Repositories.ListReleases("docker", "docker", options)
		if err != nil {
			warnWhenRateLimitExceeded(err, response)
			die("Unable to retrieve list of Docker releases from GitHub", err, retCodeRuntimeError)
		}
		allReleases = append(allReleases, releases...)
		if response.StatusCode != 200 {
			die("Unable to retrieve list of Docker releases from GitHub (Status %s).", nil, retCodeRuntimeError, response.StatusCode)
		}
		if response.NextPage == 0 {
			break
		}
		options.Page = response.NextPage
	}

	versionRegex := regexp.MustCompile(`^v([1-9]+\.\d+\.\d+)$`)
	patternRegex, err := regexp.Compile(pattern)
	if err != nil {
		die("Invalid pattern.", err, retCodeInvalidOperation)
	}

	var results []dockerversion.Version
	for _, release := range allReleases {
		version := *release.Name
		match := versionRegex.FindStringSubmatch(version)
		if len(match) > 1 && patternRegex.MatchString(version) {
			results = append(results, dockerversion.Parse(match[1]))
		}
	}

	dockerversion.Sort(results)
	return results
}

func isUpgradeAvailable() (bool, string) {
	gh := buildGithubClient()
	release, response, err := gh.Repositories.GetLatestRelease("getcarina", "dvm")
	if err != nil {
		warnWhenRateLimitExceeded(err, response)
		writeWarning("Unable to query the latest dvm release from GitHub:")
		writeWarning("%s", err)
		return false, ""
	}
	if response.StatusCode != 200 {
		writeWarning("Unable to query the latest dvm release from GitHub (Status %s):", response.StatusCode)
		return false, ""
	}

	currentVersion, err := semver.Make(dvmVersion)
	if err != nil {
		writeWarning("Unable to parse the current dvm version as a semantic version!")
		writeWarning("%s", err)
		return false, ""
	}
	latestVersion, err := semver.Make(*release.TagName)
	if err != nil {
		writeWarning("Unable to parse the latest dvm version as a semantic version!")
		writeWarning("%s", err)
		return false, ""
	}

	return latestVersion.Compare(currentVersion) > 0, *release.TagName
}

func getVersionsDir() string {
	return filepath.Join(dvmDir, "bin", "docker")
}

func getVersionDir(version dockerversion.Version) string {
	versionPath := version.SemVer.String()
	if version.IsExperimental() {
		versionPath = dockerversion.ExperimentalAlias
	}
	return filepath.Join(getVersionsDir(), versionPath)
}

func getDockerVersionVar() string {
	return strings.TrimSpace(os.Getenv("DOCKER_VERSION"))
}

func buildGithubClient() *github.Client {
	if token != "" {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
		return github.NewClient(httpClient)
	}

	client := github.NewClient(nil)
	if githubUrlOverride != "" {
		var err error
		client.BaseURL, err = neturl.Parse(githubUrlOverride)
		if err != nil {
			die("Invalid github url override: %s", err, retCodeInvalidArgument, githubUrlOverride)
		}
	}
	return client
}

func warnWhenRateLimitExceeded(err error, response *github.Response) {
	if err == nil {
		return
	}

	if response != nil {
		if response.StatusCode == 403 {
			writeWarning("Your GitHub API rate limit has been exceeded. Set the GITHUB_TOKEN environment variable or use the --github-token parameter with your GitHub personal access token to authenticate and increase the rate limit.")
		}
	}
}
