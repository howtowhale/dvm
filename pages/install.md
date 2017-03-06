---
title: Install
tags:
keywords:
summary: Install the Docker Version Manager on Mac, Windows or Linux
sidebar: home_sidebar
permalink: install.html
folder:
---

To download and install the Docker Version Manager, use the appropriate
instructions for your operating system.
Copy the commands to load `dvm` from the output, and then paste and run them to
finalize the installation.

## Mac OS X with Homebrew

Open a terminal, and then run the following commands:

```bash
brew update
brew install dvm
```

## Linux and Mac OS X without Homebrew

Open a terminal, and then run the following command:

```bash
curl -sL https://howtowhale.github.io/dvm/download/latest/install.sh | sh
```

## Windows

PowerShell performs the initial installation; you can use dvm with PowerShell or
CMD after it is installed.

Open PowerShell, and then run the following command:

```powershell
iwr 'https://howtowhale.github.io/dvm/download/latest/install.ps1' -UseBasicParsing | iex
```
