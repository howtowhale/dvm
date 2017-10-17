# Docker Version Manager wrapper for *nix
# Implemented as a POSIX-compliant function
# Should work on sh, dash, bash, ksh, zsh
# To use, source this file from your bash profile

{ # This ensures the entire script is downloaded

DVM_SCRIPT_SOURCE="$_"

__dvm_has() {
  type "$1" > /dev/null 2>&1
}

# Make zsh glob matching behave same as bash
# This fixes the "zsh: no matches found" errors
if __dvm_has "unsetopt"; then
  unsetopt nomatch 2>/dev/null
  DVM_CD_FLAGS="-q"
fi

# Default DVM_DIR to $HOME/.dvm when not set
if [ -z "$DVM_DIR" ]; then
  DVM_DIR="$HOME/.dvm"
fi

# Expect that dvm-helper is next to this script
if [ -n "$BASH_SOURCE" ]; then
  DVM_SCRIPT_SOURCE="${BASH_SOURCE[0]}"
fi
if command -v builtin >/dev/null 2>&1; then
  export DVM_HELPER="$(builtin cd $DVM_CD_FLAGS "$(dirname "${DVM_SCRIPT_SOURCE:-$0}")" > /dev/null && command pwd)/dvm-helper/dvm-helper"
else
  export DVM_HELPER="$(cd $DVM_CD_FLAGS "$(dirname "${DVM_SCRIPT_SOURCE:-$0}")" > /dev/null && command pwd)/dvm-helper/dvm-helper"
fi
unset DVM_SCRIPT_SOURCE 2> /dev/null

dvm() {
  if [ ! -f "$DVM_HELPER" ]; then
    echo "Installation corrupt: dvm-helper is missing. Please reinstall dvm."
    return 1
  fi

  # Pass dvm-helper output back to script via ~/.dvm/.tmp/dvm-output.sh
  DVM_OUTPUT="$DVM_DIR/.tmp/dvm-output.sh"
  command rm -f "$DVM_OUTPUT"

  $DVM_HELPER --shell sh $@

  # Execute any dvm-helper output
  if [ -e $DVM_OUTPUT ]; then
    source $DVM_OUTPUT
  fi
}

# Make the dvm function available to other scripts
if [ -n "$ZSH_NAME" ]; then
  autoload dvm
else
  export -f dvm
fi
}
