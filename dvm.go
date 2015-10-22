package main

import "fmt"
import "io"
import "net/http"
import "os"
import "path"
import "regexp"
import "runtime"
import "github.com/codegangsta/cli"
import "github.com/kardianos/osext"

var shell string
var dvmDir string
var debug bool
var silent bool

func main() {
  app := cli.NewApp()
  app.Name = "Docker Version Manager"
  app.Usage = "Manage multiple versions of the Docker client"
  app.EnableBashCompletion = true
  app.Flags = []cli.Flag{
    cli.StringFlag { Name: "dvm-dir", EnvVar: "DVM_DIR", Usage: "Specify an alternate DVM home directory, defaults to the current directory." },
    cli.StringFlag { Name: "shell", EnvVar: "SHELL", Usage: "Specify the shell format in which environment variables should be output, e.g. powershell, cmd or sh/bash. Defaults to sh/bash."},
    cli.BoolFlag { Name: "debug", Usage: "Print additional debug information." },
    cli.BoolFlag{ Name: "silent", EnvVar: "DVM_SILENT", Usage: "Suppress output. Errors will still be displayed."},
  }
  app.Commands = []cli.Command{
      {
          Name:    "install",
          Aliases: []string{"i"},
          Usage:   "Download and install a <version>. Uses $DOCKER_VERSION if available",
          Action: func(c *cli.Context) {
            setGlobalVars(c)
            install(c.Args().First())
          },
      },
      {
          Name:    "use",
          Usage:   "Modify PATH to use <version>. Uses $DOCKER_VERSION if available",
          Action: func(c *cli.Context) {
            setGlobalVars(c)
            use(c.Args().First())
          },
      },
  }

  app.Run(os.Args)
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
      fmt.Fprintln(os.Stderr, "Unable to determine DVM home directory")
      os.Exit(1)
    }
  }
}

func install(version string) {
  if version == "" {
    version = os.Getenv("DOCKER_VERSION")
  }

  if version == "" {
    fmt.Fprintln(os.Stderr, "The install command requires that a version is specified or the DOCKER_VERSION environment variable is set")
    os.Exit(127)
  }

  if !versionExists(version) {
    fmt.Fprintf(os.Stderr, "Version %s not found - try `dvm ls-remote` to browse available versions.\n", version)
    os.Exit(3)
  }

  versionDir := getVersionDir(version)

  // TODO: Support experimental

  if _, err := os.Stat(versionDir); err == nil {
    if !silent { fmt.Fprintf(os.Stderr, "%s is already installed\n", version) }
    use(version)
    return
  }

  if !silent { fmt.Printf("Installing %s\n", version) }

  url := fmt.Sprintf("https://get.docker.com/builds/%s/%s/%s", getDockerOS(), getDockerArch(), getDockerBinaryName(version))
  tmpPath := path.Join(getDvmDir(), ".tmp/docker", version, getBinaryName())
  downloadFile(url, tmpPath)
  binaryPath := path.Join(getDvmDir(), "bin/docker", version, getBinaryName())
  ensureParentDirectoryExists(binaryPath)
  err := os.Rename(tmpPath, binaryPath)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to copy %s to %s.\n%s\n", tmpPath, binaryPath, err)
  }

  if debug {
    fmt.Printf("Installed Docker %s to %s", version, binaryPath)
  }
}

func use(version string) {
  if version == "" {
    version = os.Getenv("DOCKER_VERSION")
  }

  if version == "" {
    fmt.Fprintln(os.Stderr, "The use command requires that a version is specified or the DOCKER_VERSION environment variable is set")
    os.Exit(127)
  }

  if !versionExists(version) {
    fmt.Fprintf(os.Stderr, "Version %s not found - try `dvm ls-remote` to browse available versions.\n", version)
    os.Exit(3)
  }

  // TODO: Handle system version

  ensureVersionIsInstalled(version)
  removePreviousDvmVersionFromPath()
  prependDvmVersionToPath(version)

  if !silent {
    fmt.Printf("Now using Docker %s\n", version)
  }
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

func prependDvmVersionToPath(version string) {
    if runtime.GOOS == "windows" && shell == "" {
      fmt.Fprintf(os.Stderr, "The --shell flag or SHELL environment variable must be set when running on Windows. Available values are sh, powershell and cmd.")
      os.Exit(127)
    }
    versionDir := getVersionDir(version)
    path := os.Getenv("PATH")

    var shellScript string
    var shellScriptExt string
    if shell == "powershell" {
      shellScript = fmt.Sprintf(`$env:PATH="%s;%s"`, versionDir, path)
      shellScriptExt = "ps1"
    } else if shell == "cmd" {
      shellScript = fmt.Sprintf("PATH=%s;%s", versionDir, path)
      shellScriptExt = "cmd"
    } else { // default to bash
      shellScript = fmt.Sprintf("export PATH=%s:%s", versionDir, path)
      shellScriptExt = "sh"
    }

    writeShellScript(shellScript, shellScriptExt)
}

func writeShellScript(contents string, fileExtension string) {
    // Write to a shell script for the calling wrapper to execute
    scriptPath := path.Join(dvmDir, ".tmp", ("dvm-output." + fileExtension))

    if debug {
      fmt.Printf("Writing wrapper shell script to %s\n%s\n", scriptPath, contents)
    }

    ensureParentDirectoryExists(scriptPath)
    scriptFile, err := os.Create(scriptPath)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Unable to create %s\n%s\n", scriptPath, err)
      os.Exit(1)
    }

    _, err = io.WriteString(scriptFile, contents)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Unable to write to %s\n%s\n", scriptPath, err)
      os.Exit(1)
    }

    scriptFile.Close()
}

func removePreviousDvmVersionFromPath() {
  regex, _ := regexp.Compile(fmt.Sprintf(`%s/bin/docker/\d+\.\d+\.\d+:`, dvmDir))
  path := regex.ReplaceAllString(os.Getenv("PATH"), "")
  os.Setenv("PATH", path)
}

func ensureVersionIsInstalled(version string) {
    versionDir := getVersionDir(version)
    if _, err := os.Stat(versionDir); err == nil {
      return
    }

    if !silent {
      fmt.Printf("%s is not installed. Installing now...\n", version)
      install(version)
    }
}

func downloadFile(url string, destPath string) {
  ensureParentDirectoryExists(destPath)

  destFile, err := os.Create(destPath)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to create to %s.\n%s\n", destPath, err)
    os.Exit(1)
  }
  defer destFile.Close()
  os.Chmod(destPath, 0755)

  if debug { fmt.Printf("Downloading %s\n", url) }

  response, err := http.Get(url)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to download %s.\n%s\n", url, err)
    os.Exit(1)
  }
  if response.StatusCode != 200 {
    fmt.Fprintf(os.Stderr, "Unable to download %s.\nStatus Code: %d\n", url, response.StatusCode)
    os.Exit(1)
  }
  defer response.Body.Close()

  _, err = io.Copy(destFile, response.Body)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to write to %s.\n%s\n", destPath, err)
    os.Exit(1)
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
  dir := path.Dir(filePath)

  err := os.MkdirAll(dir, 0777)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to create directory %s\n", dir)
    os.Exit(1)
  }
}

func getAvailableVersions(userPattern string) []string {
  return []string {"1.8.2", "1.8.3"}
}

func getVersionDir(version string) string {
  return fmt.Sprintf("%s/bin/docker/%s", dvmDir, version)
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

  fmt.Fprintf(os.Stderr, "Unsupported OS: %s\n", runtime.GOOS)
  os.Exit(1)
  return ""
}

func getDockerArch() string {
  switch runtime.GOARCH {
    case "amd64":
      return "x86_64"
    case "386":
      return "i386"
    }

  fmt.Fprintf(os.Stderr, "Unsupported ARCH: %s\n", runtime.GOARCH)
  os.Exit(1)
  return ""
}
/*
func help() {
  echo
  echo "Docker Version Manager"
  echo
  echo "Usage:"
  echo "  dvm help                              Show this message"
  echo "  dvm --version                         Print out the latest released version of dvm"
  echo "  dvm install <version>                 Download and install a <version>. Uses \$DOCKER_VERSION if available"
  echo "  dvm uninstall <version>               Uninstall a version"
  echo "  dvm use <version>                     Modify PATH to use <version>. Uses \$DOCKER_VERSION if available"
  echo "  dvm current                           Display currently activated version"
  echo "  dvm ls                                List installed versions"
  echo "  dvm ls <version>                      List versions matching a given description"
  echo "  dvm ls-remote                         List remote versions available for install"
  echo "  dvm deactivate                        Undo effects of \`dvm\` on current shell"
  echo "  dvm alias [<pattern>]                 Show all aliases beginning with <pattern>"
  echo "  dvm alias <name> <version>            Set an alias named <name> pointing to <version>"
  echo "  dvm unalias <name>                    Deletes the alias named <name>"
  echo "  dvm unload                            Unload \`dvm\` from shell"
  echo "  dvm which [<version>]                 Display path to installed docker version."
  echo
  echo "Example:"
  echo "  dvm install 1.8.1                     Install a specific version number"
  echo "  dvm use 1.6                           Use the latest available 1.6.x release"
  echo "  dvm alias default 1.8.1               Set default Docker version on a shell"
  echo
  echo "Note:"
  echo "  to remove, delete, or uninstall dvm - remove ~/.dvm"
  echo
}*/