# Docker Version Manager

[![Build Status](https://travis-ci.org/getcarina/dvm.svg?branch=master)](https://travis-ci.org/getcarina/dvm)

Version management for your Docker clients. Heavily influenced by [nvm](https://github.com/creationix/nvm) (ok, "borrowed" from).
This tool modifies your current PATH to switch between different Docker clients.

![Sample dvm run](https://cloud.githubusercontent.com/assets/836375/10118445/d4b821dc-643d-11e5-803c-91c79b93aa6c.png)

Escape from this error for a little bit longer:

```
Error response from daemon: client and server don't have same version (client : 1.18, server: 1.16)
```

## Prerequisites
Mac OS X and Linux require curl.

## Installation
After running the install script, follow the outputted instructions on how to permanently add the `dvm` command
to your shell sessions.

**Mac OS X and Linux**

```bash
$ curl -s https://raw.githubusercontent.com/getcarina/dvm/master/install.sh | sh

Downloading dvm.sh...
Docker Version Manager (dvm) has been installed to /root/.dvm
Add the following command to your bash profile (e.g. ~/.bashrc or ~/.bash_profile) complete the installation:

	source /root/.dvm/dvm.sh
```

**Windows**

Open a PowerShell command prompt and execute the following command. We use PowerShell to do the initial
installation but you can use `dvm` with PowerShell or CMD once it's installed.

```powershell
> iex (wget https://raw.githubusercontent.com/getcarina/dvm/master/install.ps1)

Downloading dvm.ps1...
Downloading dvm.cmd...

Docker Version Manager (dvm) has been installed to C:\Users\caro8994\.dvm

PowerShell Users: Add the following command to your PowerShell profile:
        . C:\Users\caro8994\.dvm\dvm.ps1

CMD Users: Run the following commands to add dvm.cmd to your PATH:
        PATH=%PATH%;C:\Users\caro8994\.dvm
        setx PATH "%PATH%;C:\Users\caro8994\.dvm"
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
