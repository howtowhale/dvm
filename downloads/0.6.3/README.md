# Docker Version Manager

[![Build Status](https://travis-ci.org/getcarina/dvm.svg?branch=master)](https://travis-ci.org/getcarina/dvm)

Version management for your Docker clients. Heavily influenced by [nvm](https://github.com/creationix/nvm) (ok, "borrowed" from).
This tool modifies your current PATH to switch between different Docker clients.

![dvm-usage](https://cloud.githubusercontent.com/assets/1368985/10800443/d3f0f39a-7d7f-11e5-87b5-1bda5ffe4859.png)

Escape from this error for a little bit longer:

```
Error response from daemon: client and server don't have same version (client : 1.18, server: 1.16)
```

## Prerequisites
* Mac OS X and Linux: curl/wget or homebrew.
* Windows: PowerShell v4+

## Installation
1. Run the one of the installation scripts below.
2. Copy, paste and run the commands from the installation output to finalize the installation.

**Mac OS X with Homebrew**

```bash
$ brew update && brew install dvm
```

**Mac OS X and Linux**

```bash
$ curl -sL https://howtowhale.github.io/dvm/downloads/latest/install.sh | sh
```

**Windows**

Open a PowerShell command prompt and execute the following command. We use PowerShell 4 to do the initial
installation but you can use `dvm` with PowerShell or CMD once it's installed.

```powershell
> Invoke-WebRequest https://howtowhale.github.io/dvm/downloads/latest/install.ps1 -UseBasicParsing | Invoke-Expression
```

## Upgrading from previous dvm
**Mac OS X with Homebrew**

Homebrew users should use `brew upgrade dvm` to get the latest version, as `dvm upgrade` is disabled in homebrew builds.

**Mac OS X, Linux and Windows**
If you have dvm 0.2 or later, run `dvm upgrade` to install the latest version of dvm.

**Note**: If you have dvm 0.0.0, then you will need to reinstall. `dvm` used to *only* be only one shell script, relying on a git backed `~/.dvm`. This worked well for \*nix users and was not workable for Windows users. We've since switched over to small wrapper scripts and a go binary called `dvm-helper` to make cross platform simple and easy. To upgrade, you can either pull and rebuild yourself (with a working go setup), or use the install script. You'll need to open up a new terminal to ensure that all the `dvm` functions are set properly.

## Usage

To install the 1.8.2 release of docker, do this:

    dvm install 1.8.2

Now in any new shell use the installed version:

    dvm use 1.8.2

If you want to see what versions are installed:

    dvm ls

If you want to see what versions are available to install:

    dvm ls-remote

To restore your PATH, you can deactivate it:

    dvm deactivate

## Bash and zsh completion

There is bash and zsh completion available in `$DVM_DIR/bash_completion`. To invoke it into your shell, run

```bash
[[ -r $DVM_DIR/bash_completion ]] && . $DVM_DIR/bash_completion
```

For zsh, there's a bit of special sauce using `bashcompinit` from the more recent versions of zsh.

### Usage

```
$ dvm [TAB]
alias        install      ls           uninstall    which
current      list         ls-alias     unload
deactivate   list-alias   ls-remote    use
help         list-remote  unalias      version
$ dvm u[TAB]
unalias    uninstall  unload     use
$ dvm us[TAB]
$ dvm use [TAB]
1.8.2         1.9.0         carina        default       experimental
```
## Mirroring docker builds

You may want to use a local mirror for Docker binaries instead of downloading them from the default site (`https://get.docker.com/builds`). There are a few possible reasons for this, most commonly the need to avoid dealing with corporate proxies every time.

The environment variable DVM_MIRROR_URL can be set to a local mirror inside your LAN:

```
export DVM_MIRROR_URL="http://localserver/docker/builds"
dvm install 1.10.3
```
