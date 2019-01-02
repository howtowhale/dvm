package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	neturl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/codegangsta/cli"
	dockerclient "github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/howtowhale/dvm/dvm-helper/dockerversion"
	"github.com/howtowhale/dvm/dvm-helper/internal/config"
	"github.com/howtowhale/dvm/dvm-helper/url"
	"github.com/pkg/errors"
	"github.com/ryanuber/go-glob"
	"golang.org/x/oauth2"
)

// These are global command line variables
var opts = config.NewDvmOptions()

// This is a nasty global state flag that we flip. Bad Carolyn.
var useAfterInstall bool

// These are set during the build
var dvmVersion string
var dvmCommit string
var upgradeDisabled string // Allow package managers like homebrew to disable in-place upgrades
var githubUrlOverride string

const (
	retCodeInvalidArgument  = 127
	retCodeInvalidOperation = 3
	retCodeRuntimeError     = 1
	versionEnvVar           = "DOCKER_VERSION"
)

func main() {
	app := makeCliApp()
	app.Run(os.Args)
}

func makeCliApp() *cli.App {
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

				writeDebug("dvm detect")
				detect()
				return nil
			},
		},
		{
			Name:    "install",
			Aliases: []string{"i"},
			Usage:   "dvm install [<version>], dvm install edge\n\tInstall a Docker version, using $DOCKER_VERSION if the version is not specified.",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "mirror-url", EnvVar: "DVM_MIRROR_URL", Usage: "Specify an alternate URL from which to download the Docker client. Defaults to https://get.docker.com/builds"},
			},
			Action: func(c *cli.Context) error {
				setGlobalVars(c)

				value := c.Args().First()
				if value == "" {
					value = getDockerVersionVar()

					if value == "" {
						die("The install command requires that a version is specified or the DOCKER_VERSION environment variable is set.", nil, retCodeInvalidArgument)
					}
				}

				writeDebug("dvm install %s", value)
				install(dockerversion.Parse(value))
				return nil
			},
		},
		{
			Name:  "uninstall",
			Usage: "dvm uninstall <version>\n\tUninstall a Docker version.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)

				value := c.Args().First()
				if value == "" {
					die("The uninstall command requires that a version is specified.", nil, retCodeInvalidArgument)
				}

				writeDebug("dvm uninstall %s", value)
				uninstall(dockerversion.Parse(value))
				return nil
			},
		},
		{
			Name:  "use",
			Usage: "dvm use [<version>], dvm use system, dvm use edge\n\tUse a Docker version, using $DOCKER_VERSION if the version is not specified.",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "mirror-url", EnvVar: "DVM_MIRROR_URL", Usage: "Specify an alternate URL from which to download the Docker client. Defaults to https://get.docker.com/builds"},
				cli.BoolFlag{Name: "nocheck", EnvVar: "DVM_NOCHECK", Usage: "Do not check if version exists (use with caution)."},
			},
			Action: func(c *cli.Context) error {
				setGlobalVars(c)

				value := c.Args().First()
				if value == "" {
					value = getDockerVersionVar()

					if value == "" {
						die("The use command requires that a version is specified or the DOCKER_VERSION environment variable is set.", nil, retCodeInvalidOperation)
					}
				}

				writeDebug("dvm use %s", value)
				use(dockerversion.Parse(value))
				return nil
			},
		},
		{
			Name:  "deactivate",
			Usage: "dvm deactivate\n\tUndo the effects of `dvm` on current shell.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)

				writeDebug("dvm deactivate")
				deactivate()
				return nil
			},
		},
		{
			Name:  "current",
			Usage: "dvm current\n\tPrint the current Docker version.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)

				writeDebug("dvm current")
				current()
				return nil
			},
		},
		{
			Name:  "which",
			Usage: "dvm which\n\tPrint the path to the current Docker version.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)

				writeDebug("dvm which")
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
				value := c.Args().Get(1)
				if name == "" || value == "" {
					die("The alias command requires both an alias name and a version.", nil, retCodeInvalidArgument)
				}

				writeDebug("dvm alias %s %s", name, value)
				alias(name, value)
				return nil
			},
		},
		{
			Name:  "unalias",
			Usage: "dvm unalias <alias>\n\tRemove a Docker version alias.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)

				alias := c.Args().First()
				if alias == "" {
					die("The unalias command requires an alias alias.", nil, retCodeInvalidArgument)
				}

				writeDebug("dvm unalias %s", alias)
				unalias(alias)
				return nil
			},
		},
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "dvm list [<pattern>]\n\tList installed Docker versions.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)

				pattern := c.Args().First()

				writeDebug("dvm list %s", pattern)
				list(pattern)
				return nil
			},
		},
		{
			Name:    "list-remote",
			Aliases: []string{"ls-remote"},
			Usage:   "dvm list-remote [<prefix>]\n\tList available Docker versions.",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "pre", Usage: "Include pre-release versions"},
			},
			Action: func(c *cli.Context) error {
				setGlobalVars(c)

				pattern := c.Args().First()

				writeDebug("dvm list-remote %s", pattern)
				listRemote(pattern)
				return nil
			},
		},
		{
			Name:    "list-alias",
			Aliases: []string{"ls-alias"},
			Usage:   "dvm list-alias\n\tList Docker version aliases.",
			Action: func(c *cli.Context) error {
				setGlobalVars(c)

				writeDebug("dvm list-alias")
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

	return app
}

func setGlobalVars(c *cli.Context) {
	useAfterInstall = true

	opts.Debug = c.GlobalBool("debug")
	if opts.Debug {
		opts.Logger = log.New(color.Output, "", log.LstdFlags)
	} else {
		opts.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
	}

	opts.Token = c.GlobalString("github-token")
	opts.Shell = c.GlobalString("shell")
	validateShellFlag()

	opts.Silent = c.GlobalBool("silent")
	opts.MirrorURL = c.String("mirror-url")
	opts.IncludePrereleases = c.Bool("pre")

	opts.DvmDir = c.GlobalString("dvm-dir")
	if opts.DvmDir == "" {
		opts.DvmDir = filepath.Join(getUserHomeDir(), ".dvm")
	}
	writeDebug("The dvm home directory is: %s", opts.DvmDir)
}

func detect() {
	writeDebug("dvm detect")

	docker, err := dockerclient.NewEnvClient()
	if err != nil {
		die("Cannot build a docker client from environment variables", err, retCodeRuntimeError)
	}

	versionResult, err := docker.ServerVersion(context.Background())
	if err != nil {
		die("Unable to query docker version", err, retCodeRuntimeError)
	}

	writeDebug("Queried /version and got Version: %s", versionResult.Version)
	version := dockerversion.Parse(versionResult.Version)

	// Docker versions prior to 1.12 don't return a usable client version
	// Lookup the client version from the API version
	if version.IsEmpty() {
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
		clientRange, found := oldVersionMap[versionResult.APIVersion]
		if !found {
			die("Unable to detect the proper client version for Docker API version %s", nil, retCodeRuntimeError, versionResult.APIVersion)
		}

		// Find the highest version that satisfies the client version range
		availableVersions := getAvailableVersions("", true)
		for i := len(availableVersions) - 1; i >= 0; i-- {
			v := availableVersions[i]

			if ok, _ := v.InRange(clientRange); ok {
				version = v
				break
			}
		}
		if version.IsEmpty() {
			die("Unable to detect the proper client version for %s", nil, retCodeRuntimeError, clientRange)
		}
	}
	writeDebug("Detected client version: %s", version)

	os.Setenv(versionEnvVar, version.String())
	writeEnvironmentVariableScript(versionEnvVar)

	use(version)
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
	prefix := url.Join("https://howtowhale.github.io/dvm/downloads", version)
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
		if current.String() == version.String() {
			color.Green("->\t%s", version)
		} else {
			writeInfo("\t%s", version)
		}
	}
}

func install(version dockerversion.Version) {
	versionDir := getVersionDir(version)

	if version.IsEdge() {
		// Always install latest of edge build
		err := os.RemoveAll(versionDir)
		if err != nil {
			die("Unable to remove edge version at %s.", err, retCodeRuntimeError, versionDir)
		}
	}

	if _, err := os.Stat(versionDir); err == nil {
		writeWarning("%s is already installed", version)
		use(version)
		return
	}

	writeInfo("Installing %s...", version)

	downloadRelease(version)

	if useAfterInstall {
		use(version)
	}
}

func downloadRelease(version dockerversion.Version) {
	destPath := filepath.Join(getVersionDir(version), getBinaryName())
	err := version.Download(opts, destPath)
	if err != nil {
		die("", err, retCodeRuntimeError)
	}

	writeDebug("Downloaded Docker %s to %s", version, destPath)
}

func uninstall(version dockerversion.Version) {
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
	if version.IsAlias() && aliasExists(version.Name()) {
		aliasedVersion, _ := ioutil.ReadFile(getAliasPath(version.Name()))
		version = dockerversion.NewAlias(version.Name(), string(aliasedVersion))
		writeDebug("Using alias: %s", version)
	}

	useAfterInstall = false
	ensureVersionIsInstalled(version)

	if version.IsSystem() {
		version, _ = getSystemDockerVersion()
	} else if version.IsEdge() {
		version, _ = getEdgeDockerVersion()
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

func alias(alias string, value string) {
	version := dockerversion.NewAlias(alias, value)
	if !isVersionInstalled(version) {
		die("The aliased version, %s, is not installed.", nil, retCodeInvalidArgument, version)
	}

	aliasPath := getAliasPath(alias)
	if _, err := os.Stat(aliasPath); err == nil {
		writeDebug("Overwriting existing alias.")
	}

	writeFile(aliasPath, version.Value())
	writeInfo("Aliased %s to %s.", alias, value)
}

func unalias(alias string) {
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
	return filepath.Join(opts.DvmDir, "alias", alias)
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
	if opts.Shell == "powershell" {
		fileExtension = "ps1"
	} else if opts.Shell == "cmd" {
		fileExtension = "cmd"
	} else { // default to bash
		fileExtension = "sh"
	}
	return filepath.Join(opts.DvmDir, ".tmp", "dvm-output."+fileExtension)
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
		if version.Equals(availableVersion) {
			writeDebug("Version %s is installed", version)
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
	edgeVersionPath, _ := getEdgeDockerPath()

	isSystem := currentDockerPath == systemDockerPath
	isEdge := currentDockerPath == edgeVersionPath

	current, _ := getDockerVersion(currentDockerPath, isEdge)

	if isSystem {
		writeDebug("The current docker is the system installation")
		current.SetAsSystem()
	}

	if isEdge {
		writeDebug("The current docker is an edge version")
		current.SetAsEdge()
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

func getEdgeDockerPath() (string, error) {
	edgeVersionPath := filepath.Join(getVersionsDir(), dockerversion.EdgeAlias, getBinaryName())
	_, err := os.Stat(edgeVersionPath)
	return edgeVersionPath, err
}

func getEdgeDockerVersion() (dockerversion.Version, error) {
	edgeDockerpath, err := getEdgeDockerPath()
	if err != nil {
		return dockerversion.Version{}, err
	}
	version, err := getDockerVersion(edgeDockerpath, true)
	version.SetAsEdge()
	return version, err
}

func getDockerVersion(dockerPath string, includeBuild bool) (dockerversion.Version, error) {
	stdout, _ := exec.Command(dockerPath, "-v").Output()
	rawVersion := strings.TrimSpace(string(stdout))

	writeDebug("%s -v output: %s", dockerPath, rawVersion)

	versionRegex := regexp.MustCompile(`^Docker version (.+), build (.+)?`)
	match := versionRegex.FindStringSubmatch(rawVersion)
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

func listRemote(prefix string) {
	versions := getAvailableVersions(prefix, opts.IncludePrereleases)
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

		if version.IsEdge() {
			edgeDockerPath := filepath.Join(versionDir, getBinaryName())
			edgeVersion, err := getDockerVersion(edgeDockerPath, true)
			if err != nil {
				writeDebug("Unable to get version of installed edge version at %s.\n%s", versionDir, err)
				continue
			}
			version = dockerversion.NewAlias(dockerversion.EdgeAlias, edgeVersion.Value())
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

func getAvailableVersions(pattern string, includePrereleases bool) []dockerversion.Version {
	versions := make(map[string]dockerversion.Version)

	writeDebug("Retrieving legacy Docker releases")
	legacyVersions, err := listLegacyDockerVersions()
	if err != nil {
		die("", err, retCodeRuntimeError)
	}
	for _, v := range legacyVersions {
		if !includePrereleases && v.IsPrerelease() {
			continue
		}
		if strings.HasPrefix(v.Value(), pattern) {
			versions[v.String()] = v
		}
	}

	writeDebug("Retrieving Docker releases")
	stableVersions, err := dockerversion.ListVersions(opts.MirrorURL, dockerversion.Stable)
	if err != nil {
		die("", err, retCodeRuntimeError)
	}
	for _, v := range stableVersions {
		if strings.HasPrefix(v.Value(), pattern) {
			versions[v.String()] = v
		}
	}

	if includePrereleases {
		writeDebug("Retrieving Docker pre-releases")
		prereleaseVersions, err := dockerversion.ListVersions(opts.MirrorURL, dockerversion.Test)
		if err != nil {
			die("", err, retCodeRuntimeError)
		}
		for _, v := range prereleaseVersions {
			if strings.HasPrefix(v.Value(), pattern) {
				versions[v.String()] = v
			}
		}
	}

	results := make([]dockerversion.Version, 0, len(versions))
	for _, v := range versions {
		results = append(results, v)
	}

	dockerversion.Sort(results)

	return results
}

func listLegacyDockerVersions() ([]dockerversion.Version, error) {
	gh := buildGithubClient()
	options := &github.ListOptions{PerPage: 100}

	var allReleases []github.RepositoryRelease
	for {
		releases, response, err := gh.Repositories.ListReleases("moby", "moby", options)
		if err != nil {
			warnWhenRateLimitExceeded(err, response)
			return nil, errors.Wrap(err, "Unable to retrieve list of Docker releases from GitHub")
		}
		allReleases = append(allReleases, releases...)
		if response.StatusCode != 200 {
			return nil, errors.Errorf("Unable to retrieve list of Docker releases from GitHub (Status %v).", response.StatusCode)
		}
		if response.NextPage == 0 {
			break
		}
		options.Page = response.NextPage
	}

	var results []dockerversion.Version
	for _, release := range allReleases {
		version := *release.Name
		v := dockerversion.Parse(version)
		if v.IsEmpty() {
			writeDebug("Ignoring non-semver Docker release: %s", version)
			continue
		}
		results = append(results, v)
	}

	return results, nil
}

func isUpgradeAvailable() (bool, string) {
	gh := buildGithubClient()
	release, response, err := gh.Repositories.GetLatestRelease("howtowhale", "dvm")
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

	currentVersion, err := semver.NewVersion(dvmVersion)
	if err != nil {
		writeWarning("Unable to parse the current dvm version as a semantic version!")
		writeWarning("%s", err)
		return false, ""
	}
	latestVersion, err := semver.NewVersion(*release.TagName)
	if err != nil {
		writeWarning("Unable to parse the latest dvm version as a semantic version!")
		writeWarning("%s", err)
		return false, ""
	}

	return latestVersion.GreaterThan(currentVersion), *release.TagName
}

func getVersionsDir() string {
	return filepath.Join(opts.DvmDir, "bin", "docker")
}

func getVersionDir(version dockerversion.Version) string {
	versionPath := version.Slug()
	if version.IsEdge() {
		versionPath = dockerversion.EdgeAlias
	}
	return filepath.Join(getVersionsDir(), versionPath)
}

func getDockerVersionVar() string {
	return strings.TrimSpace(os.Getenv("DOCKER_VERSION"))
}

func buildGithubClient() *github.Client {
	if opts.Token != "" {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: opts.Token})
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
