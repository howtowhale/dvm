#!/usr/bin/env bash
set -euo pipefail

rm -fr _deploy/ &> /dev/null
git clone --branch gh-pages --depth 1 https://github.com/howtowhale/dvm.git _deploy
cp -R bin/dvm/$(VERSION) _deploy/downloads/
cp -R bin/dvm/$(PERMALINK) _deploy/downloads/
cd _deploy/
git add downloads/
git commit --author "Travis CI <travis@travis-ci.org>" -m "Publish $VERSION from Travis Build #$TRAVIS_BUILD_NUMBER"
#PKEY=.travis.deploy.pem git push
