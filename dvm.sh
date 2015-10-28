#!/usr/bin/env bash

# Docker Version Manager Bash Wrapper
# Implemented as a POSIX-compliant function
# Should work on sh, dash, bash, ksh, zsh
# To use, source this file from your bash profile

dvm() {
  DVM_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

  if [ ! -f "${DVM_DIR}/dvm-helper" ]; then
    # Detect mac vs. linux and x86 vs. x64
    DVM_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    [ $(getconf LONG_BIT) == 64 ] && DVM_ARCH="amd64" || DVM_ARCH="386"

    # Download latest release
    LATEST_TAG=$(curl -s https://api.github.com/repos/getcarina/dvm/tags | grep name -m 1 | awk '{print $2}' | cut -d '"' -f2)
    curl -L -s -o $DVM_DIR/dvm-helper https://github.com/getcarina/dvm/releases/download/$LATEST_TAG/dvm-helper-$DVM_OS-$DVM_ARCH
    chmod u+x $DVM_DIR/dvm-helper
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
