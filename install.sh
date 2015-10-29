#!/usr/bin/env bash

{ # this ensures the entire script is downloaded #

set -e

dvm_has() {
  type "$1" > /dev/null 2>&1
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

install_dvm_helper() {
  local url
  local bin

  # Detect mac vs. linux and x86 vs. x64
  DVM_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  [ $(getconf LONG_BIT) == 64 ] && DVM_ARCH="amd64" || DVM_ARCH="386"

  # Download latest release
  LATEST_TAG=$(curl -s https://api.github.com/repos/getcarina/dvm/tags | grep name -m 1 | awk '{print $2}' | cut -d '"' -f2)
  bin="$DVM_DIR/dvm-helper"
  url=https://github.com/getcarina/dvm/releases/download/$LATEST_TAG/dvm-helper-$DVM_OS-$DVM_ARCH
  dvm_download -L -C - --progress-bar $url -o "$bin"
  chmod u+x $DVM_DIR/dvm-helper
}


if [ -z "$DVM_DIR" ]; then
  DVM_DIR="$HOME/.dvm"
fi

if [ ! -d "$DVM_DIR" ]; then
  mkdir $DVM_DIR
fi

echo "Downloading dvm.sh..."
dvm_download -L -C - --progress-bar https://raw.githubusercontent.com/getcarina/dvm/master/dvm.sh -o $DVM_DIR/dvm.sh

if [ ! -f "${DVM_DIR}/dvm-helper" ]; then
  echo "Downloading dvm-helper..."
  install_dvm_helper
fi

echo "Docker Version Manager (dvm) has been installed to ${DVM_DIR}"
echo "Add the following command to your bash profile (e.g. ~/.bashrc or ~/.bash_profile) to complete the installation:"
echo ""
echo "\tsource ${DVM_DIR}/dvm.sh"

}
