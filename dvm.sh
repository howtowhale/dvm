#!/usr/bin/env bash

# Docker Version Manager Bash Wrapper
# Implemented as a POSIX-compliant function
# Should work on sh, dash, bash, ksh, zsh
# To use, source this file from your bash profile

dvm() {
  local DVM_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

  if [ ! -f "${DVM_DIR}/dvm-helper" ]; then
    echo "Installation corrupt: dvm-helper is missing. Please reinstall dvm."
    return 1
  fi

  DVM_OUTPUT="${DVM_DIR}/.tmp/dvm-output.sh"
  if [ -e $DVM_OUTPUT ]; then
    rm $DVM_OUTPUT
  fi

  DVM_DIR=$DVM_DIR $DVM_DIR/dvm-helper --shell sh $@

  if [ -e $DVM_OUTPUT ]; then
    source $DVM_OUTPUT
  fi
}
