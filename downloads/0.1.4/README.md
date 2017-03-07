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
$ curl -s https://howtowhale.github.io/dvm/downloads/latest/install.sh | sh
```

**Windows**

Open a PowerShell command prompt and execute the following command. We use PowerShell to do the initial
installation but you can use `dvm` with PowerShell or CMD once it's installed.

```powershell
> iex (wget https://howtowhale.github.io/dvm/downloads/latest/install.ps1)
```

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
