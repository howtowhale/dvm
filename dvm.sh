# Docker Version Manager
# Implemented as a POSIX-compliant function
# Should work on sh, dash, bash, ksh, zsh
# To use, source this file from your bash profile
#
# Implemented by Kyle Kelley <rgbkrk@gmail.com>
# Influenced by nvm, via Tim Caswell <tim@creationix.com>

{ # This ensures the entire script is downloaded

DVM_SCRIPT_SOURCE="$_"

dvm_has() {
  type "$1" > /dev/null 2>&1
}

dvm_is_alias() {
  # this is intentionally not "command alias" so it works in zsh.
  \alias "$1" > /dev/null 2>&1
}

dvm_download() {
  if dvm_has "curl"; then
    curl -q $*
  elif dvm_has "wget"; then
    # Emulate curl with wget
    ARGS=$(echo "$*" | command sed -e 's/--progress-bar /--progress=bar /' \
                           -e 's/-L //' \
                           -e 's/-I /--server-response /' \
                           -e 's/-s /-q /' \
                           -e 's/-o /-O /' \
                           -e 's/-C - /-c /')
    eval wget $ARGS
  fi
}

dvm_has_system_docker() {
  [ "$(dvm deactivate >/dev/null 2>&1 && command -v docker)" != '' ]
}

# Make zsh glob matching behave same as bash
# This fixes the "zsh: no matches found" errors
if dvm_has "unsetopt"; then
  unsetopt nomatch 2>/dev/null
  DVM_CD_FLAGS="-q"
fi

# Auto detect the DVM_DIR when not set
if [ -z "$DVM_DIR" ]; then
  if [ -n "$BASH_SOURCE" ]; then
    DVM_SCRIPT_SOURCE="${BASH_SOURCE[0]}"
  fi
  export DVM_DIR=$(cd $DVM_CD_FLAGS $(dirname "${DVM_SCRIPT_SOURCE:-$0}") > /dev/null && \pwd)
fi
unset DVM_SCRIPT_SOURCE 2> /dev/null

# Setup mirror location if not already set
if [ -z "$DVM_GET_DOCKER_MIRROR" ]; then
  export DVM_GET_DOCKER_MIRROR="https://get.docker.com/builds/"
fi

## Skippy skip


dvm() {
    if [ $# -lt 1 ]; then
      dvm help
      return
    fi

    local GREP_OPTIONS
    GREP_OPTIONS=''

    local VERSION
    local ADDITIONAL_PARAMETERS
    local ALIAS

    # TODO: help mentions .dvmrc, doesn't exist yet

    case $1 in
      "help" )
        echo
        echo "Docker Version Manager"
        echo
        echo "Usage:"
        echo "  dvm help                              Show this message"
        echo "  dvm --version                         Print out the latest released version of dvm"
        echo "  dvm install [-s] <version>            Download and install a <version>, [-s] from source. Uses .dvmrc if available"
        echo "  dvm uninstall <version>               Uninstall a version"
        echo "  dvm use <version>                     Modify PATH to use <version>. Uses .dvmrc if available"
        echo "  dvm run <version> [<args>]            Run <version> with <args> as arguments. Uses .dvmrc if available for <version>"
        echo "  dvm current                           Display currently activated version"
        echo "  dvm ls                                List installed versions"
        echo "  dvm ls <version>                      List versions matching a given description"
        echo "  dvm ls-remote                         List remote versions available for install"
        echo "  dvm deactivate                        Undo effects of \`dvm\` on current shell"
        echo "  dvm alias [<pattern>]                 Show all aliases beginning with <pattern>"
        echo "  dvm alias <name> <version>            Set an alias named <name> pointing to <version>"
        echo "  dvm unalias <name>                    Deletes the alias named <name>"
        echo "  dvm unload                            Unload \`dvm\` from shell"
        echo "  dvm which [<version>]                 Display path to installed docker version. Uses .dvmrc if available"
        echo
        echo "Example:"
        echo "  dvm install 1.8.1                     Install a specific version number"
        echo "  dvm use 1.6                           Use the latest available 1.6.x release"
        echo "  dvm run 1.6.1 -it ubuntu /bin/bash    Run an ubuntu container using Docker 1.6.1"
        echo "  dvm exec 1.6.1 nginx                  Run the nginx image the PATH pointing to Docker 1.6.1"
        echo "  dvm alias default 1.8.1               Set default Docker version on a shell"
        echo
        echo "Note:"
        echo "  to remove, delete, or uninstall dvm - remove ~/.dvm"
        echo
      ;;

      "debug" )
        echo >&2 "\$SHELL: $SHELL"
        echo >&2 "\$DVM_DIR: $(echo $DVM_DIR | sed "s#$HOME#\$HOME#g")"
        for DVM_DEBUG_COMMAND in 'dvm current' 'which docker'
        do
          local DVM_DEBUG_OUTPUT="$($DVM_DEBUG_COMMAND | sed "s#$DVM_DIR#\$DVM_DIR#g")"
          echo >&2 "$DVM_DEBUG_COMMAND: ${DVM_DEBUG_OUTPUT}"
        done
        return 42
      ;;

      "current" )
        dvm_version current
      ;;

      * )
        >&2 nvm help
        return 127
      ;;

    esac
}


}
