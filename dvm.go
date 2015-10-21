package main

import "fmt"
import "io"
import "net/http"
import "os"
import "path"
import "github.com/codegangsta/cli"

func main() {
  app := cli.NewApp()
  app.Name = "Docker Version Manager"
  app.Usage = "Manage multiple versions of the Docker client"
  app.EnableBashCompletion = true
  app.Commands = []cli.Command{
      {
          Name:    "install",
          Aliases: []string{"i"},
          Usage:   "Download and install a <version>. Uses $DOCKER_VERSION if available",
          Action: func(c *cli.Context) {
              install(c.Args().First())
          },
      },
      {
          Name:    "use",
          Usage:   "Modify PATH to use <version>. Uses $DOCKER_VERSION if available",
          Action: func(c *cli.Context) {
              println("completed task: ", c.Args().First())
          },
      },
  }

  app.Run(os.Args)

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
    fmt.Fprintf(os.Stderr, "%s is already installed\n", version)
    use(version)
    return
  }

  fmt.Printf("Installing %s\n", version)

  url := fmt.Sprintf("https://get.docker.com/builds/%s/%s/docker-%s", getOS(), getArch(), version)
  tmpPath := path.Join(getDvmDir(), ".tmpbin/docker", version, getBinaryName())
  downloadFile(url, tmpPath)
  binaryPath := path.Join(getDvmDir(), "bin/docker", version, getBinaryName())
  ensureParentDirectoryExists(binaryPath)
  err := os.Rename(tmpPath, binaryPath)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to copy %s to %s.\n%s\n", tmpPath, binaryPath, err)
  }
}

func getDvmDir() string {
  return "/Users/caro8994/.dvm"
}

func getBinaryName() string {
    if getOS() == "Windows" {
      return "docker.exe"
    }
    return "docker"
}

func use(version string) {
  // TODO: implement
  fmt.Printf("Using %s\n", version)
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

  fmt.Printf("Downloading %s\n", url)
  response, err := http.Get(url)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to download %s.\n%s\n", url, err)
    os.Exit(1)
  }
  defer response.Body.Close()

  bytes, err := io.Copy(destFile, response.Body)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to write to %s.\n%s\n", destPath, err)
    os.Exit(1)
  }

  fmt.Printf("Download complete (%d) to %s!\n", bytes, destPath)
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
  // TODO: implement
  return fmt.Sprintf("/Users/caro8994/.dvm/bin/docker/%s", version)
}

func getOS() string {
  // TODO: implement
  return "Darwin"
}

func getArch() string {
  // TODO: implement
  return "x86_64"
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
