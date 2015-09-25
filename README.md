# Docker Version Manager

[![Build Status](https://travis-ci.org/rgbkrk/dvm.svg?branch=master)][https://travis-ci.org/rgbkrk/dvm]

Version management for your Docker clients. Heavily influenced by [nvm]() (ok, "borrowed" from)

Put together to deal with version drift between docker machine versions.

## Installation

First, you'll need to make sure your system has `curl` or `wget` available. If you go the manual install route, you'll need `git` too.

Note: `dvm` does not support Windows or Fish shell.

### Manual install

If you have `git` installed, then clone this repository to `~/.dvm`.

```
git clone https://github.com/rgbkrk/dvm.git ~/.dvm
# TODO: git checkout `git describe --abbrev=0 --tags`
```

To activate dvm, you need to source it from your shell:

. ~/.dvm/dvm.sh

Add this line to your `~/.bashrc`, `~/.profile`, or `~/.zshrc` file to have it automatically sourced upon login.

## Usage

To install the 1.8.2 release of docker, do this:

    dvm install 1.8.2

Now in any new shell use the installed version

    dvm use 1.8.2

If you want to see what versions are installed:

    dvm ls

If you want to see what versions are available to install:

    dvm ls-remote

To restore your PATH, you can deactivate it.

    dvm deactivate
