#!/usr/bin/env bash

# Docker Version Manager Bash Wrapper
# Implemented as a POSIX-compliant function
# Should work on sh, dash, bash, ksh, zsh
# To use, source this file from your bash profile

dvm() {
  DVM_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

  if [ ! -f "${DVM_DIR}/dvm" ]; then
    # TODO: Download the binary instead of grabbing it from a local build
    cp $DVM_DIR/bin/darwin/amd64/dvm $DVM_DIR
    chmod u+x $DVM_DIR/dvm
  fi

  DVM_OUTPUT="${DVM_DIR}/.tmp/dvm-output.sh"
  if [ -e $DVM_OUTPUT ]; then
    rm $DVM_OUTPUT
  fi

  DVM_DIR=$DVM_DIR $DVM_DIR/dvm --shell sh $@

  if [ -e $DVM_OUTPUT ]; then
    source $DVM_OUTPUT
  fi
}
