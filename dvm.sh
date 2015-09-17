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

}
