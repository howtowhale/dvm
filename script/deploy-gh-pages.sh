#!/usr/bin/env bash
set -euo pipefail

rm -fr _deploy/ &> /dev/null
git clone --branch gh-pages --depth 1 git@github.com:howtowhale/dvm.git _deploy

cp -R bin/dvm/$VERSION _deploy/downloads/
cp -R bin/dvm/$PERMALINK _deploy/downloads/
cd _deploy/
git add downloads/

git config user.name "Travis CI"
git config user.email "travis@travis-ci.org"
git commit -m "Publish $VERSION from Travis Build #${TRAVIS_BUILD_NUMBER:-1}"
git push
