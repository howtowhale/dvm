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

dvm_tree_contains_path() {
  local TREE
  TREE=$1
  local DOCKER_PATH
  DOCKER_PATH="$2"

  if [ "_${TREE}" = "_" ] || [ "_${DOCKER_PATH}" = "_" ]; then
    >&2 echo "Both the tree and Docker path are required by dvm_tree_contains_path."
    return 2
  fi

  local PATHDIR
  PATHDIR=$(dirname "${DOCKER_PATH}")
  while [ "$PATHDIR" != "" ] && [ "$PATHDIR" != "." ] && [ "$PATHDIR" != "/" ] && [ "$PATHDIR" != "$TREE" ]; do
    PATHDIR=$(dirname "$PATHDIR")
  done
  [ "$PATHDIR" = "$TREE" ]
}

# Get the latest version of dvm
dvm_get_latest() {
  >&2 echo "Not implemented yet"
  return 1
}

dvm_version_path() {
  local VERSION
  VERSION="$1"

  if [ -z "${VERSION}" ]; then
    echo "version is required" >&2
    return 3
  else
    echo "${DVM_VERSION_DIR}/${VERSION}"
  fi
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

dvm_ensure_version_installed() {
  local PROVIDED_VERSION
  PROVIDED_VERSION="$1"
  local LOCAL_VERSION
  local EXIT_CODE

  LOCAL_VERSION="$(dvm_version "${PROVIDED_VERSION}")"
  EXIT_CODE="$?"

  local DVM_CHOSEN_DIR
  if [ "_${EXIT_CODE}" = "_0" ]; then
    DVM_CHOSEN_DIR="$(dvm_version_path "$LOCAL_VERSION")"
  fi
  if [ "_${EXIT_CODE}" != "_0" ] || [ ! -d "${DVM_CHOSEN_DIR}" ]; then
    VERSION="$(dvm_resolve_alias "$PROVIDED_VERSION")"
    if [ $? -eq 0 ]; then
      echo "N/A: version \"${PROVIDED_VERSION} -> ${VERSION}\" is not yet installed" >&2
    else
      echo "N/A: version \"${PROVIDED_VERSION}\" is not yet installed" >&2
    fi
    return 1
  fi
}

dvm_has_system_docker() {
  [ "$(dvm deactivate >/dev/null 2>&1 && command -v docker)" != '' ]
}

dvm_match_version() {
  local PROVIDED_VERSION
  PROVIDED_VERSION="$1"
  case "_${PROVIDED_VERSION}" in
    "_system" )
      echo "system"
      ;;
    * )
      echo "$(dvm_version ${PROVIDED_VERSION})"
      ;;
  esac
}

# Expand a version using the version cache.
dvm_version() {
  local PATTERN
  PATTERN=$1

  # The default pattern is the current one
  if [ -z "${PATTERN}" ]; then
    PATTERN='current'
  fi

  if [ "${PATTERN}" = "current" ]; then
    dvm_ls_current
    return $?
  fi

  VERSION="$(dvm_ls "${PATTERN}" | command tail -n1)"
  if [ -z "${VERSION}" ] || [ "_${VERSION}" = "_N/A" ]; then
    echo "N/A"
    return 3
  else
    echo "${VERSION}"
  fi
}

dvm_alias() {
  local ALIAS
  ALIAS="$1"

  if [ -z "$ALIAS" ]; then
    echo >&2 "An alias is required."
    return 1
  fi

  local DVM_ALIAS_PATH
  DVM_ALIAS_PATH="${DVM_ALIAS_DIR}/${ALIAS}"

  if [ ! -f "${DVM_ALIAS_PATH}"]; then
    echo >&2 "Alias does not exist."
    return 2
  fi

  cat "${DVM_ALIAS_PATH}"
}

dvm_resolve_alias() {
  if [ -z "$1" ]; then
    return 1
  fi

  local PATTERN
  PATTERN="$1"

  local ALIAS
  ALIAS="$PATTERN"
  local ALIAS_TEMP

  local SEEN_ALIASES
  SEEN_ALIASES="$ALIAS"

  while true; do
    ALIAS_TEMP="$(dvm_alias "$ALIAS" 2> /dev/null)"

    if [ -z "$ALIAS_TEMP" ]; then
      break
    fi

    if [ -n "$ALIAS_TEMP" ] \
      && command printf "$SEEN_ALIASES" | command grep -e "^$ALIAS_TEMP$" > /dev/null; then
      ALIAS="∞"
      break
    fi

    SEEN_ALIASES="${SEEN_ALIASES}\n${ALIAS_TEMP}"
    ALIAS="$ALIAS_TEMP"
  done

  if [ -n "$ALIAS" ] && [ "_${ALIAS}" != "_${PATTERN}" ]; then
    echo "$ALIAS"
    return 0
  fi

  return 2
}

dvm_resolve_local_alias() {
  if [ -z "$1" ]; then
    return 1
  fi

  local VERSION
  local EXIT_CODE
  VERSION="$(dvm_resolve_alias "$1")"
  EXIT_CODE=$?

  if [ -z "${VERSION}" ]; then
    return ${EXIT_CODE}
  fi
  if [ "_${VERSION}" != "_∞" ]; then
    dvm_version "${VERSION}"
  else
    echo "${VERSION}"
  fi
}

dvm_ls() {
  local PATTERN
  PATTERN="$1"
  local VERSIONS
  VERSIONS=''

  if [ "${PATTERNS}" = 'current' ]; then
    dvm_ls_current
    return
  fi

  if dvm_resolve_local_alias "${PATTERN}"; then
    return
  fi

  if [ -n "${PATTERN}" ] && [ -d "$(dvm_version_path "${PATTERN}")" ]; then
    VERSIONS="$PATTERN"
  fi

  if [ -z "$VERSIONS" ]; then
    VERSIONS="$(command find ${DVM_VERSION_DIR} -maxdepth 1 -type d -name "${PATTERN}*" \
      | command sed "
        s#^${DVM_VERSION_DIR}/##;
        \#^${DVM_VERSION_DIR}# d;
      " \
      | command sort -t.)"
  fi

  if dvm_has_system_docker; then
    if [ -z "$PATTERN" ]; then
      VERSIONS="$VERSIONS$(command printf '\n%s' 'system')"
    elif [ "$PATTERN" = 'system' ]; then
      VERSIONS="$(command printf '%s' 'system')"
    fi
  fi

  if [ -z "${VERSIONS}" ]; then
    echo "N/A"
    return 3
  fi

  echo "$VERSIONS"
}

dvm_ls_current() {
  local DVM_LS_CURRENT_DOCKER_PATH
  DVM_LS_CURRENT_DOCKER_PATH="$(command which docker 2> /dev/null)"
  if [ $? -ne 0 ]; then
    echo 'none'
  elif dvm_tree_contains_path "$DVM_DIR" "$DVM_LS_CURRENT_DOCKER_PATH"; then
    local VERSION

    VERSION="$(command docker version -f '{{.Client.Version}}' 2>/dev/null)"

    if [ $? -ne 0 ]; then
      # docker version -f is not supported in older docker clients.
      VERSION="$(command docker version |
        command grep 'Client version:' |
        command sed -e "s#^Client version: ##")"
    fi

    echo "${VERSION}"
  else
    echo "system"
  fi
}

dvm_strip_path() {
  echo "$1" | command sed \
    -e "s#${DVM_DIR}/[^/]*$2[^:]*:##g" \
    -e "s#:${DVM_DIR}/[^/]*$2[^:]*##g" \
    -e "s#${DVM_DIR}/[^/]*$2[^:]*##g"
}

dvm_prepend_path() {
  if [ -z "$1" ]; then
    echo "$2"
  else
    echo "$2:$1"
  fi
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

DVM_VERSION_DIR="${DVM_DIR}/bin/docker"
DVM_ALIAS_DIR="${DVM_DIR}/alias"

# Setup mirror location if not already set
if [ -z "$DVM_GET_DOCKER_MIRROR" ]; then
  export DVM_GET_DOCKER_MIRROR="https://get.docker.com/builds"
fi
if [ -z "$DVM_GET_EXPERIMENTAL_DOCKER_MIRROR" ]; then
  export DVM_GET_EXPERIMENTAL_DOCKER_MIRROR="https://experimental.docker.com/builds"
fi

dvm_get_os() {
  local DVM_UNAME
  DVM_UNAME="$(uname -a)"
  local DVM_OS
  case "$DVM_UNAME" in
    Linux\ *) DVM_OS=Linux ;;
    Darwin\ *) DVM_OS=Darwin ;;
    FreeBSD\ *) DVM_OS=FreeBSD ;; # Whoa, this is available actually
  esac
  echo "$DVM_OS"
}

dvm_get_arch() {
  local DVM_UNAME
  DVM_UNAME="$(uname -m)"
  local DVM_ARCH
  case "$DVM_UNAME" in
    x86_64) DVM_ARCH="x86_64" ;;
    i*86) DVM_ARCH="i386" ;;
    *) DVM_ARCH="$DVM_UNAME" ;;
  esac
  echo "$DVM_ARCH"
}


dvm_remote_version() {
  local PATTERN
  PATTERN="$1"
  local VERSION
  VERSION="$(dvm_ls_remote "$PATTERN" | tail -n1)"

  echo "$VERSION"
  if [ "_$VERSION" = '_N/A' ]; then
    return 3
  fi
}

dvm_version_greater() {
  local LHS
  LHS=$(dvm_normalize_version "$1")
  local RHS
  RHS=$(dvm_normalize_version "$2")
  [ $LHS -gt $RHS ];
}

dvm_version_greater_than_or_equal_to() {
  local LHS
  LHS=$(dvm_normalize_version "$1")
  local RHS
  RHS=$(dvm_normalize_version "$2")
  [ $LHS -ge $RHS ];
}

dvm_normalize_version() {
  echo "${1#v}" | command awk -F. '{ printf("%d%06d%06d\n", $1,$2,$3); }'
}

dvm_binary_available() {
  # If they're trying to install experimental, assume it's available
  if [ "$1" == "experimental" ]; then
    return 0
  fi

  # binaries started with docker 0.6.0
  # binaries + checksums started with docker 0.10.0
  local FIRST_VERSION_WITH_BINARY_AND_CHECKSUM
  FIRST_VERSION_WITH_BINARY_AND_CHECKSUM="0.10.0"
  dvm_version_greater_than_or_equal_to "$1" "$FIRST_VERSION_WITH_BINARY_AND_CHECKSUM"
}

dvm_install_docker_binary() {
  local VERSION
  VERSION="$1"

  local VERSION_PATH
  VERSION_PATH="$(dvm_version_path ${VERSION})"

  local DVM_OS
  DVM_OS="$(dvm_get_os)"
  if [ -n "$DVM_OS" ]; then
    if dvm_binary_available "$VERSION"; then
      local DVM_ARCH
      DVM_ARCH="$(dvm_get_arch)"

      local DOCKER_BINARY
      DOCKER_BINARY="docker"

      local url
      if [ "$VERSION" == "experimental" ]; then
        url="$DVM_GET_EXPERIMENTAL_DOCKER_MIRROR/$DVM_OS/$DVM_ARCH/docker-latest"
      else
        url="$DVM_GET_DOCKER_MIRROR/$DVM_OS/$DVM_ARCH/docker-$VERSION"
      fi

      local checksum_url
      checksum_url="${url}.sha256"

      local sum
      sum=`dvm_download -L -s $checksum_url -o - | command awk '{print $1}'`

      local binbin
      binbin="${VERSION_PATH}/${DOCKER_BINARY}"

      local tmpdir
      tmpdir="$DVM_DIR/.tmpbin/docker/${VERSION}"

      local tmpbin
      tmpbin="${tmpdir}/${DOCKER_BINARY}"

      local DVM_INSTALL_ERRORED

      command mkdir -p "$tmpdir" && \
        dvm_download -L -C - --progress-bar $url -o "$tmpbin" || \
        DVM_INSTALL_ERRORED=true

      if (
        [ "$DVM_INSTALL_ERRORED" != true ] && \
        dvm_checksum "$tmpbin" $sum && \
        command mkdir -p "$VERSION_PATH" && \
        command mv "$tmpbin" "$binbin"
        ); then
        chmod a+x $binbin
        return 0
      else
        echo >&2 "Binary download failed."
        return 1
      fi
    fi
  fi
}

dvm_checksum() {
  local DVM_CHECKSUM
  if dvm_has "sha256sum" && ! dvm_is_alias "sha256sum"; then
    DVM_CHECKSUM="$(command sha256sum "$1" | command awk '{print $1}')"
  elif dvm_has "shasum" && ! dvm_is_alias "shasum"; then
    DVM_CHECKSUM="$(shasum -a 256 "$1" | command awk '{print $1}')"
  else
    echo "Unaliased sha256sum, shasum not found." >&2
    return 2
  fi

  if [ "_$DVM_CHECKSUM" = "_$2" ]; then
    return
  elif [ -z "$2" ]; then
    echo 'Checksums empty'
    return
  else
    echo 'Checksums do not match.' >&2
    return 1
  fi
}

dvm_ls_remote() {
  local PATTERN
  PATTERN="$1"
  local VERSIONS
  local GREP_OPTIONS
  GREP_OPTIONS=''

  local RELEASES_URL
  RELEASES_URL="https://api.github.com/repos/docker/docker/tags?per_page=100"

  VERSIONS=`dvm_download -L -s $RELEASES_URL -o - \
            | \egrep -o 'v[0-9]+\.[0-9]+\.[0-9]+' \
            | grep -v "^v0\.[0-9]\." \
            | grep "^v${PATTERN}" \
            | sort -t. -u -k 1.2,1n -k 2,2n -k 3,3n \
            | cut -c 2-`

  # Sneak experimental in
  VERSIONS=`echo -e "${VERSIONS}\nexperimental" | grep "$PATTERN"`

  if [ -z "$VERSIONS" ]; then
    echo "N/A"
    return 3
  fi
  echo "$VERSIONS"
}

dvm_print_versions() {
  local VERSION
  local FORMAT
  local DVM_CURRENT

  DVM_CURRENT=$(dvm_ls_current)
  echo "$1" | while read VERSION; do
    if [ "_${VERSION}" = "_${DVM_CURRENT}" ]; then
      FORMAT='\033[0;32m-> %12s\033[0m'
    elif [ "$VERSION" = "system" ]; then
      FORMAT='\033[0;33m%15s\033[0m'
    elif [ -d "$(dvm_version_path "$VERSION" 2> /dev/null)" ]; then
      FORMAT='\033[0;34m%15s\033[0m'
    else
      FORMAT='%15s'
    fi
    command printf "$FORMAT\n" $VERSION
  done
}

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
      ;;

      "--version" | "-v" )
        echo "$(cd $DVM_DIR && command git describe --tags --always HEAD)"
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

      "ls" | "list" )
        local DVM_LS_OUTPUT
        local DVM_LS_EXIT_CODE

        DVM_LS_OUTPUT=$(dvm_ls "$2")
        DVM_LS_EXIT_CODE=$?

        dvm_print_versions "$DVM_LS_OUTPUT"

        if [ $# -eq 1 ]; then
          dvm alias
        fi
        return $DVM_LS_EXIT_CODE
      ;;

      "ls-remote" | "list-remote" )
        local PATTERN
        PATTERN="$2"

        dvm_ls_remote "$PATTERN"
      ;;

      "current" )
        dvm_version current
      ;;

      "install" | "i" )
        local version_not_provided
        version_not_provided=0

        local provided_version

        if ! dvm_has "curl" && ! dvm_has "wget"; then
          'dvm needs curl or wget to proceed.' >&2;
          return 1
        fi

        shift # remove "install"

        if [ $# -lt 1 ]; then
          # Support DOCKER_VERSION as the version
          if [ -n "$DOCKER_VERSION" ]; then
            provided_version=$DOCKER_VERSION
          else
            version_not_provided=1
            >&2 dvm help
            return 127
          fi
        else
          provided_version="$1"
          shift
        fi

        VERSION="$(dvm_remote_version ${provided_version})"

        if [ "_$VERSION" == "_N/A" ]; then
          echo "Version ${provided_version} not found - try \`dvm ls-remote\` to browse available versions." >&2
          return 3
        fi

        local VERSION_PATH
        VERSION_PATH="$(dvm_version_path ${VERSION})"


        if [ -d "$VERSION_PATH" ]; then
          if [ "$VERSION" == "experimental" ]; then
            # TODO: Pull the checksum to see if we've already downloaded it
            echo "$VERSION is already installed, but we'll assume you want the latest experimental"
          else
            echo "$VERSION is already installed." >&2
            dvm use "$VERSION"
            return $?
          fi
        fi

        local DVM_INSTALL_SUCCESS
        if dvm_binary_available "$VERSION"; then
          if dvm_install_docker_binary $VERSION; then
            DVM_INSTALL_SUCCESS=true
          fi
        fi

        if [ "$DVM_INSTALL_SUCCESS" = true ]; then
          dvm use "$VERSION"
        fi
        return $?
      ;;

    "uninstall" )
      if [ $# -ne 2 ]; then
        >&2 echo "dvm uninstall requires an argument."
        >&2 echo
        >&2 dvm help
        return 127
      fi

      local PATTERN
      local VERSION
      local INFERRED_MSG

      PATTERN="$2"
      VERSION="$(dvm_version "${PATTERN}")"

      INFERRED_MSG=""
      if [ "_${VERSION}" != "_${PATTERN}" ]; then
        INFERRED_MSG=" (inferred from ${PATTERN})"
      fi

      if [ "_${VERSION}" = "_$(dvm_ls_current)" ]; then
        echo "Cannot uninstall currently-active Docker version, ${VERSION}${INFERRED_MSG}." >&2
        return 1
      fi

      local VERSION_DIR
      VERSION_DIR="$(dvm_version_path "${VERSION}")"
      if [ ! -d "${VERSION_DIR}" ]; then
        echo "${VERSION}${INFERRED_MSG} is not installed." >&2
        return
      fi

      if [ -z "${VERSION_DIR}" ]; then
        echo "No version directory found." >&2
        return 1
      fi

      VERSION_DIR=$(cd ${VERSION_DIR} && pwd)

      # This keeps things like "dvm uninstall ../../.." from working.
      if ! dvm_tree_contains_path "${DVM_VERSION_DIR}" "${VERSION_DIR}" ; then
        echo "Version directory ${VERSION_DIR} is not beneath ${DVM_VERSION_DIR}." >&2
        return 1
      fi

      # Delete all files related to the target version
      command rm -rf "${VERSION_DIR}"
      echo "Uninstalled Docker version ${VERSION}${INFERRED_MSG}."
      ;;

    "use" )
      local PROVIDED_VERSION
      local DVM_USE_SILENT
      DVM_USE_SILENT=0

      shift # remove "use"

      while [ $# -ne 0 ]
      do
        case "$1" in
          --silent) DVM_USE_SILENT=1 ;;
          *)
            if [ -n "$1" ]; then
              PROVIDED_VERSION="$1"
            fi
          ;;
        esac
        shift
      done

      if [ -n "${PROVIDED_VERSION}" ]; then
        VERSION=$(dvm_match_version "${PROVIDED_VERSION}")
      else
        # Support DOCKER_VERSION as the version
        if [ -n "$DOCKER_VERSION" ]; then
          VERSION=$(dvm_match_version "$DOCKER_VERSION")
        fi
      fi

      if [ -z "${VERSION}" ]; then
        >&2 dvm help
        return 127
      fi

      if [ "_${VERSION}" = '_system' ]; then
        if dvm_has_system_docker && dvm deactivate >/dev/null 2>&1; then
          if [ $DVM_USE_SILENT -ne 1 ]; then
            echo "Now using system version of Docker: $(docker version 2>/dev/null)"
          fi
          return
        else
          if [ $DVM_USE_SILENT -ne 1 ]; then
            echo "System version of Docker not found." >&2
          fi
          return 127
        fi
      elif [ "_${VERSION}" = "_∞" ]; then
        if [ $DVM_USE_SILENT -ne 1 ]; then
          echo "The alias \"${PROVIDED_VERSION}\" loads to an infinite loop. Aborting." >&2
        fi
        return 8
      fi

      dvm_ensure_version_installed "${VERSION}"
      EXIT_CODE=$?
      if [ "${EXIT_CODE}" != "0" ]; then
        return ${EXIT_CODE}
      fi

      local DVM_CHOSEN_DIR
      DVM_CHOSEN_DIR="$(dvm_version_path ${VERSION})"

      # Strip other versions from the PATH
      PATH="$(dvm_strip_path "${PATH}" "/docker/")"
      # Prepend the current version
      PATH="$(dvm_prepend_path "${PATH}" "${DVM_CHOSEN_DIR}")"

      export PATH
      hash -r

      export DVM_BIN="${DVM_CHOSEN_DIR}"

      if [ "${DVM_SYMLINK_CURRENT}" = "true" ]; then
        command rm -f "${DVM_DIR/current}" && ln -s "${DVM_CHOSEN_DIR}" "${DVM_DIR}/current"
      fi

      if [ $DVM_USE_SILENT -ne 1 ]; then
        echo "Now using Docker ${VERSION}"
      fi
      ;;

    "alias" )
      command mkdir -p ${DVM_ALIAS_DIR}
      if [ $# -le 2 ]; then
        local DEST
        for ALIAS_PATH in "${DVM_ALIAS_DIR}"/"$2"*; do
          ALIAS=$(command basename "${ALIAS_PATH}")
          DEST=$(dvm_alias "$ALIAS" 2> /dev/null)
          if [ -n "${DEST}" ]; then
            VERSION="$(dvm_version "${DEST}")"
            if [ "_${DEST}" = "_${VERSION}" ]; then
              echo "${ALIAS} -> ${DEST}"
            else
              echo "${ALIAS} -> ${DEST} (-> ${VERSION})"
            fi
          fi
        done

        return
      fi

      if [ -z "$3" ]; then
        command rm -f "${DVM_ALIAS_DIR}/${2}"
        echo "${2} -> *poof*"
        return
      fi

      VERSION="$(dvm_version "${3}")"
      if [ $? -ne 0 ]; then
        echo "! WARNING: Version '${3}' does not exist." >&2
      fi
      echo "${3}" | tee "${DVM_ALIAS_DIR}/${2}" > /dev/null
      if [ "_${3}" != "_${VERSION}" ]; then
        echo "${2} -> ${3} (-> ${VERSION})"
      else
        echo "${2} -> ${3}"
      fi
      ;;

    "unalias" )
      command mkdir -p "${DVM_ALIAS_DIR}"
      if [ $# -ne 2 ]; then
        >&2 dvm help
        return 127
      fi

      local DVM_ALIAS_PATH
      DVM_ALIAS_PATH="${DVM_ALIAS_DIR}/${2}"

      if [ ! -f "${DVM_ALIAS_PATH}" ]; then
        echo "Alias ${2} doesn't exist!" >&2
        return
      fi

      local REAL_ALIAS_DIR
      REAL_ALIAS_DIR=$(cd $(dirname ${DVM_ALIAS_PATH}) && pwd)
      if [ "_${REAL_ALIAS_DIR}" != "_${DVM_ALIAS_DIR}" ] && ! dvm_tree_contains_path "${DVM_ALIAS_DIR}" "${REAL_ALIAS_DIR}})"; then
        echo "Alias path ${DVM_ALIAS_PATH} is not beneath ${DVM_ALIAS_DIR}." >&2
        return 1
      fi

      command rm -f "${DVM_ALIAS_PATH}"
      echo "Deleted alias ${2}."
      ;;

    "deactivate" )
      local NEWPATH
      NEWPATH="$(dvm_strip_path "$PATH" "/docker/")"
      if [ "_${PATH}" = "_${NEWPATH}" ]; then
        echo "Could not find ${DVM_DIR}/* in \$PATH" >&2
      else
        export PATH="$NEWPATH"
        hash -r
        echo "${DVM_DIR}/* removed from \$PATH"
      fi
      ;;

    "which" )
      >&2 echo ""
      >&2 echo "dvm $1 is not implemented - raise an issue and see if you can implement it ;)"
      >&2 dvm help
      return 126
      ;;


    * )
        >&2 echo ""
        >&2 echo "dvm $1 is not a command"
        >&2 dvm help
        return 127
      ;;

    esac
}


}
