package main

import "errors"
import "fmt"
import "io"
import "net/http"
import "os"
import "os/exec"
import "path/filepath"
import "regexp"
import "runtime"
import "sort"
import "strings"
import "github.com/fatih/color"
import "github.com/google/go-github/github"
import "github.com/codegangsta/cli"
import "github.com/kardianos/osext"

// These are global command line variables
var shell string
var dvmDir string
var debug bool
var silent bool

// These are set during the build
var dvmVersion string
var dvmCommit string

const (
    INVALID_ARGUMENT = 127
    INVALID_OPERATION = 3
    RUNTIME_ERROR = 1
)

func main() {
  app := cli.NewApp()
  app.Name = "Docker Version Manager"
  app.Usage = "Manage multiple versions of the Docker client"
  app.Version = fmt.Sprintf("%s (%s)", dvmVersion, dvmCommit)
  app.EnableBashCompletion = true
  app.Flags = []cli.Flag{
    cli.StringFlag { Name: "dvm-dir", EnvVar: "DVM_DIR", Usage: "Specify an alternate DVM home directory, defaults to the current directory." },
    cli.StringFlag { Name: "shell", EnvVar: "SHELL", Usage: "Specify the shell format in which environment variables should be output, e.g. powershell, cmd or sh/bash. Defaults to sh/bash."},
    cli.BoolFlag { Name: "debug", Usage: "Print additional debug information." },
    cli.BoolFlag{ Name: "silent", EnvVar: "DVM_SILENT", Usage: "Suppress output. Errors will still be displayed."},
  }
  app.Commands = []cli.Command{
    {
      Name: "install",
      Aliases: []string{"i"},
      Usage: "dvm install <version>. Uses $DOCKER_VERSION if available",
      Action: func(c *cli.Context) {
        setGlobalVars(c)
        install(c.Args().First())
      },
    },
    {
      Name: "uninstall",
      Usage: "dvm uninstall <version>",
      Action : func(c *cli.Context) {
        setGlobalVars(c)
        uninstall(c.Args().First())
      },
    },
    {
      Name: "use",
      Usage: "dvm use <version>. dvm use system reverts to the system installation of Docker. Uses $DOCKER_VERSION if available",
      Action: func(c *cli.Context) {
        setGlobalVars(c)
        use(c.Args().First())
      },
    },
    {
      Name: "current",
      Usage: "dvm current",
      Action: func(c *cli.Context) {
        setGlobalVars(c)
        current()
      },
    },
    {
      Name: "list",
      Aliases: []string{"ls"},
      Usage: "dvm list [<pattern>]",
      Action: func(c *cli.Context) {
        setGlobalVars(c)
        list(c.Args().First())
      },
    },
    {
      Name: "list-remote",
      Aliases: []string{"ls-remote"},
      Usage: "dvm list-remote [<pattern>]",
      Action: func(c *cli.Context) {
        setGlobalVars(c)
        listRemote(c.Args().First())
      },
    },
    {
      Name: "deactivate",
      Usage: "dvm deactivate. Undo effects of `dvm` on current shell",
      Action: func(c *cli.Context) {
        setGlobalVars(c)
        deactivate()
      },
    },
  }

  app.Run(os.Args)
}

func writeDebug(format string, a ...interface{}) {
  if !debug { return }

  color.Cyan(format, a...)
}

func writeInfo(format string, a ...interface{}) {
  if silent { return }

  color.White(format, a...)
}

func writeWarning(format string, a ...interface{}) {
  if silent { return }

  color.Yellow(format, a...)
}

func writeError(format string, err error, a ...interface{}) {
  color.Set(color.FgRed)
  fmt.Fprintf(os.Stderr, format + "\n", a...)
  if err != nil {
    fmt.Fprintln(os.Stderr, err)
  }
  color.Unset()
}

func die(format string, err error, exitCode int, a ...interface{}) {
  writeError(format, err, a...)
  os.Exit(exitCode)
}

func setGlobalVars(c *cli.Context) {
  debug = c.GlobalBool("debug")
  shell = c.GlobalString("shell")
  silent = c.GlobalBool("silent")
  dvmDir = c.GlobalString("dvm-dir")
  if dvmDir == "" {
    var err error
    dvmDir, err = osext.ExecutableFolder()
    if err != nil {
      die("Unable to determine DVM home directory", nil, 1)
    }
  }
}

func current() {
  current, err := getCurrentDockerVersion()
  if err != nil {
    writeWarning("N/A")
  } else {
    writeInfo(current)
  }
}

func list(pattern string) {
  versions := getInstalledVersions(pattern)
  current, _ := getCurrentDockerVersion()

  for _, version := range versions {
    if current == version {
      color.Green("->\t%s", version)
    } else {
      writeInfo("\t%s", version)
    }
  }
}

func install(version string) {
  if version == "" {
    version = getDockerVersionVar()
  }

  if version == "" {
    die("The install command requires that a version is specified or the DOCKER_VERSION environment variable is set.", nil, INVALID_ARGUMENT)
  }

  if !versionExists(version) {
    die("Version %s not found - try `dvm ls-remote` to browse available versions.", nil, INVALID_OPERATION, version)
  }

  versionDir := getVersionDir(version)

  // TODO: Support experimental

  if _, err := os.Stat(versionDir); err == nil {
    writeWarning("%s is already installed", version)
    use(version)
    return
  }

  writeInfo("Installing %s...", version)

  url := fmt.Sprintf("https://get.docker.com/builds/%s/%s/%s", getDockerOS(), getDockerArch(), getDockerBinaryName(version))
  tmpPath := filepath.Join(getDvmDir(), ".tmp/docker", version, getBinaryName())
  downloadFile(url, tmpPath)
  binaryPath := filepath.Join(getDvmDir(), "bin/docker", version, getBinaryName())
  ensureParentDirectoryExists(binaryPath)
  err := os.Rename(tmpPath, binaryPath)
  if err != nil {
    die("Unable to copy %s to %s.", err, RUNTIME_ERROR, tmpPath, binaryPath)
  }

  writeDebug("Installed Docker %s to %s.", version, binaryPath)
}

func uninstall(version string) {
    if version == "" {
      die("The uninstall command requires that a version is specified.", nil, INVALID_ARGUMENT)
    }

    current, _ := getCurrentDockerVersion()
    if current == version {
      die("Cannot uninstall the currently active Docker version.", nil, INVALID_OPERATION)
    }

    versionDir := getVersionDir(version)
    if _, err := os.Stat(versionDir); os.IsNotExist(err)  {
      writeWarning("%s is not installed.", version)
      return
    }

    err := os.RemoveAll(versionDir)
    if err != nil {
      die("Unable to uninstall Docker version %s located in %s.", err, RUNTIME_ERROR, version, versionDir)
    }

    writeInfo("Uninstalled Docker %s.", version)
}

func use(version string) {
  if version == "" {
    version = getDockerVersionVar()
  }

  if version == "" {
    die("The use command requires that a version is specified or the DOCKER_VERSION environment variable is set.", nil, INVALID_OPERATION)
  }

  // dvm use system undoes changes to the PATH and uses installed version of DOcker
  if version == "system" {
    systemDockerVersion, err := getSystemDockerVersion()
    if err != nil {
      die("System version of Docker not found.", nil, INVALID_OPERATION)
    }

    removePreviousDvmVersionFromPath()
    writeInfo("Now using system version of Docker: %s", systemDockerVersion)
    writeShellScript()
    return
  }

  if !versionExists(version) {
    die("Version %s not found - try `dvm ls-remote` to browse available versions.", nil, INVALID_OPERATION, version)
  }

  ensureVersionIsInstalled(version)
  removePreviousDvmVersionFromPath()
  prependDvmVersionToPath(version)
  writeShellScript()

  writeInfo("Now using Docker %s", version)
}

func getDvmDir() string {
  return dvmDir
}

func getDockerBinaryName(version string) string {
    if runtime.GOOS == "windows" {
      return fmt.Sprintf("docker-%s.exe", version)
    }
    return fmt.Sprintf("docker-%s", version)
}

func getBinaryName() string {
    if runtime.GOOS == "windows" {
      return "docker.exe"
    }
    return "docker"
}

func deactivate() {
  removePreviousDvmVersionFromPath()
  writeShellScript()
}

func prependDvmVersionToPath(version string) {
  var pathSep string
  if runtime.GOOS == "windows" { pathSep = ";" } else { pathSep = ":" }
  versionDir := getVersionDir(version)
  path := fmt.Sprintf("%s%s%s", versionDir, pathSep, os.Getenv("PATH"))
  os.Setenv("PATH", path)
}

func writeShellScript() {
  if runtime.GOOS == "windows" && shell == "" {
    die("The --shell flag or SHELL environment variable must be set when running on Windows. Available values are sh, powershell and cmd.", nil, INVALID_ARGUMENT)
  }

  path := os.Getenv("PATH")

  var contents string
  var fileExtension string
  if shell == "powershell" {
    contents = fmt.Sprintf(`$env:PATH="%s"`, path)
    fileExtension = "ps1"
  } else if shell == "cmd" {
    contents = fmt.Sprintf("PATH=%s", path)
    fileExtension = "cmd"
  } else { // default to bash
    contents = fmt.Sprintf("export PATH=%s", path)
    fileExtension = "sh"
  }

  // Write to a shell script for the calling wrapper to execute
  scriptPath := filepath.Join(dvmDir, ".tmp", ("dvm-output." + fileExtension))

  writeDebug("Writing wrapper shell script to %s", scriptPath, contents)
  writeDebug(contents)

  ensureParentDirectoryExists(scriptPath)
  scriptFile, err := os.Create(scriptPath)
  if err != nil {
    die("Unable to create %s", err, RUNTIME_ERROR, scriptPath)
  }

  _, err = io.WriteString(scriptFile, contents)
  if err != nil {
    die("Unable to write to %s", err, RUNTIME_ERROR, scriptPath)
  }

  scriptFile.Close()
}

func removePreviousDvmVersionFromPath() {
  versionDir := getVersionDir("")

  var pathRegex string
  if runtime.GOOS == "windows" {
    escapedVersionDir := strings.Replace(versionDir, `\`, `\\`, -1)
    pathRegex = escapedVersionDir + `\\\d+\.\d+\.\d+;`
  } else {
    pathRegex = versionDir + `/\d+\.\d+\.\d+:`
  }

  regex, _ := regexp.Compile(pathRegex)
  path := regex.ReplaceAllString(os.Getenv("PATH"), "")
  os.Setenv("PATH", path)
}

func ensureVersionIsInstalled(version string) {
    versionDir := getVersionDir(version)
    if _, err := os.Stat(versionDir); err == nil {
      return
    }

    writeInfo("%s is not installed. Installing now...", version)
    install(version)
}

func downloadFile(url string, destPath string) {
  ensureParentDirectoryExists(destPath)

  destFile, err := os.Create(destPath)
  if err != nil {
    die("Unable to create to %s.", err, RUNTIME_ERROR, destPath)
  }
  defer destFile.Close()
  os.Chmod(destPath, 0755)

  writeDebug("Downloading %s", url)

  response, err := http.Get(url)
  if err != nil {
    die("Unable to download %s.", err, RUNTIME_ERROR, url)
  }

  if response.StatusCode != 200 {
    die("Unable to download %s. (Status %d)", nil, RUNTIME_ERROR, url, response.StatusCode)
  }
  defer response.Body.Close()

  _, err = io.Copy(destFile, response.Body)
  if err != nil {
    die("Unable to write to %s.", err, RUNTIME_ERROR, destPath)
  }
}

func versionExists(version string) bool {
  availableVersions := getAvailableVersions(version)

  for _,availableVersion := range availableVersions {
      if version == availableVersion {
          return true
      }
  }
  return false
}

func ensureParentDirectoryExists(filePath string) {
  dir := filepath.Dir(filePath)

  err := os.MkdirAll(dir, 0777)
  if err != nil {
    die("Unable to create directory %s.", err, RUNTIME_ERROR, dir)
  }
}

func getCurrentDockerPath() (string, error) {
  currentDockerPath, err := exec.LookPath("docker")
  return currentDockerPath, err
}

func getCurrentDockerVersion() (string, error) {
  currentDockerPath, err := getCurrentDockerPath()
  if err != nil {
    return "", err
  }

  current, _ := getDockerVersion(currentDockerPath)
  systemDockerPath, _ := getSystemDockerPath()
  if currentDockerPath == systemDockerPath {
    current += " (system)"
  }

  return current, nil
}

func getSystemDockerPath() (string, error) {
  originalPath := os.Getenv("PATH")
  removePreviousDvmVersionFromPath()
  systemDockerPath, err := exec.LookPath("docker")
  os.Setenv("PATH", originalPath)
  return systemDockerPath, err
}

func getSystemDockerVersion() (string, error) {
  systemDockerPath, err := getSystemDockerPath()
  if err != nil {
    return "", err
  }
  return getDockerVersion(systemDockerPath)
}

func getDockerVersion(dockerPath string) (string, error) {
  rawVersion, _ := exec.Command(dockerPath, "-v").Output()

  writeDebug("%s -v output: %s", dockerPath, rawVersion)

  versionRegex := regexp.MustCompile(`^Docker version (.*),`)
  match := versionRegex.FindSubmatch(rawVersion)
  if len(match) < 2 {
    return "", errors.New("Could not detect docker version.")
  }

  return string(match[1][:]), nil
}

func listRemote(pattern string) {
    versions := getAvailableVersions(pattern)
    for _, version := range versions {
      writeInfo(version)
    }
}

func getInstalledVersions(pattern string) []string {
  versions, _ := filepath.Glob(getVersionDir(pattern + "*"))

  var results []string
  for _, versionDir := range versions {
    results = append(results, filepath.Base(versionDir))
  }

  systemVersion, err := getSystemDockerVersion()
  if err == nil {
    results = append(results, systemVersion + " (system)")
  }

  sort.Strings(results)
  return results
}

func getAvailableVersions(pattern string) []string {
  gh := github.NewClient(nil)
	tags, response, err := gh.Repositories.ListTags("docker", "docker", nil)
  if err != nil {
    die("Unable to retrieve list of Docker tags from GitHub", err, RUNTIME_ERROR)
  }
  if response.StatusCode != 200 {
    die("Unable to retrieve list of Docker tags from GitHub (Status %s).", nil, RUNTIME_ERROR, response.StatusCode)
  }

  versionRegex := regexp.MustCompile(`^v([1-9]+\.\d+\.\d+)$`)
  patternRegex, err := regexp.Compile(pattern)
  if err != nil {
    die("Invalid pattern.", err, INVALID_OPERATION)
  }

  var results []string
  for _, tag := range tags {
    version := *tag.Name
    match := versionRegex.FindStringSubmatch(version)
    if len(match) > 1 && patternRegex.MatchString(version) {
      results = append(results, match[1])
    }
  }

  sort.Strings(results)
  return results
}

func getVersionDir(version string) string {
  return filepath.Join(dvmDir, "bin", "docker", version)
}

func getDockerOS() string {
  switch runtime.GOOS {
    case "windows":
      return "Windows"
    case "darwin":
      return "Darwin"
    case "linux":
      return "Linux"
    }

  die("Unsupported OS: %s", nil, RUNTIME_ERROR, runtime.GOOS)
  return ""
}

func getDockerArch() string {
  switch runtime.GOARCH {
    case "amd64":
      return "x86_64"
    case "386":
      return "i386"
    }

  die("Unsupported ARCH: %s", nil, RUNTIME_ERROR, runtime.GOARCH)
  return ""
}

func getDockerVersionVar() string {
  return strings.TrimSpace(os.Getenv("DOCKER_VERSION"))
}
