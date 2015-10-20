package main

import "fmt"
import "os"
import "github.com/codegangsta/cli"

func main() {
  app := cli.NewApp()
  app.Name = "Docker Version Manager"
  app.Usage = "Manage multiple versions of the Docker client"
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
  fmt.Printf("Installing %s", version)
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
