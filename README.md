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
Mac OS X and Linux require curl or wget.

## Installation
1. Run the one of the installation scripts below.
2. Copy, paste and run the commands from the installation output to finalize the installation.

**Mac OS X and Linux**

```bash
$ curl -s -L https://download.getcarina.com/dvm/latest/install.sh | sh
```

**Windows**

Open a PowerShell command prompt and execute the following command. We use PowerShell to do the initial
installation but you can use `dvm` with PowerShell or CMD once it's installed.

```powershell
> iex (wget https://download.getcarina.com/dvm/latest/install.ps1)
```

## Upgrading from previous dvm

`dvm` used to *only* be only one shell script, relying on a git backed `~/.dvm`. This worked well for \*nix users and was not workable for Windows users. We've since switched over to small wrapper scripts and a go binary called `dvm-helper` to make cross platform simple and easy. To upgrade, you can either pull and rebuild yourself (with a working go setup), or use the install script. You'll need to open up a new terminal to ensure that all the `dvm` functions are set properly.

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
