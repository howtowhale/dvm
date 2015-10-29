#!/usr/bin/env bash

{ # this ensures the entire script is downloaded #

set -e

if [ -z "$DVM_DIR" ]; then
  DVM_DIR="$HOME/.dvm"
fi

if [ ! -d "$DVM_DIR" ]; then
  mkdir $DVM_DIR
fi

echo "Downloading dvm.sh..."
curl -s -o $DVM_DIR/dvm.sh https://raw.githubusercontent.com/getcarina/dvm/master/dvm.sh

echo "Docker Version Manager (dvm) has been installed to ${DVM_DIR}"
echo "Add the following command to your bash profile (e.g. ~/.bashrc or ~/.bash_profile) complete the installation:"
echo ""
echo "\tsource ${DVM_DIR}/dvm.sh"

}
