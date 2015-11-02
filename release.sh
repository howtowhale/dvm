#!/usr/bin/env bash
set -euo pipefail

declare -xr ORG="carolynvs" # TODO: getcarina
declare -xr REPO="dvm"
declare -xr BINARY="dvm-helper"
declare -xr DESCRIPTION="The Docker Version Manager (dvm)"

function usage {
  echo 'usage: release.sh "{major}.{minor}.{patch}"'
}

function main {
  # Pick your own leveled up tag
  TAG=${1:-}
  NAME=$TAG

  if [ "$TAG" == "" ] || [ "$NAME" == "" ] || [ "${TAG:0:1}" == "v" ]; then
    usage
    exit 5
  fi


  echo "Releasing '$TAG' : $DESCRIPTION"


  if [ ! -e "$( which github-release )" ]; then
    echo "You need github-release installed."
    echo "go get github.com/aktau/github-release"
    exit 1
  fi

  BRANCH=$(git rev-parse --abbrev-ref HEAD 2> /dev/null)

  if [ "$BRANCH" != "master" ]; then
    echo "Must release from master branch"
    exit 2
  fi

  set +e
  git diff --exit-code > /dev/null
  if [ $? != 0 ]; then
    echo "Workspace is not clean. Exiting"
    exit 3
  fi
  set -e

  REMOTE="release"
  REMOTE_URL="git@github.com:${ORG}/${REPO}.git"

  #
  # Confirm that we have a remote named "release"
  #

  set +e
  git remote show ${REMOTE} &> /dev/null

  rc=$?

  if [[ $rc != 0 ]]; then
    echo "Remote \"${REMOTE}\" not found. Exiting."
    exit 4
  fi
  set -e

  #
  # Now confirm that we've got the proper remote URL
  #

  REMOTE_ACTUAL_URL=$(git remote show ${REMOTE} | grep Push | cut -d ":" -f2- | xargs)

  if [ "$REMOTE_URL" != "$REMOTE_ACTUAL_URL" ]; then
    echo -e "Remote \"${REMOTE}\" PUSH url incorrect.\nShould be ${REMOTE_URL}. Exiting."
    exit 5
  fi

  make clean
  # Build off master to make sure all is well
  make dvm-helper
  echo "Out with the old, in with the new"
  ./dvm-helper/dvm-helper --version
  echo "---------------------------------"

  # Build with the tag now for actual binary shipping
  git fetch --tags $REMOTE
  make build-tagged-for-release TAG=$TAG

  github-release release \
    --user "$ORG" \
    --repo "$REPO" \
    --tag "$TAG" \
    --name "$NAME" \
    --description "$DESCRIPTION"

  push_binaries
}

function push_binaries {
  github-release upload \
    --user "$ORG" \
    --repo "$REPO" \
    --tag "$TAG" \
    --name "${BINARY}-linux-amd64" \
    --file bin/${BINARY}-linux-amd64

  github-release upload \
    --user "$ORG" \
    --repo "$REPO" \
    --tag "$TAG" \
    --name "${BINARY}-darwin-amd64" \
    --file bin/${BINARY}-darwin-amd64

  github-release upload \
    --user "$ORG" \
    --repo "$REPO" \
    --tag "$TAG" \
    --name "${BINARY}-windows-amd64.exe" \
    --file bin/${BINARY}-windows-amd64.exe
}

main "$1"
