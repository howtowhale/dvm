---
title: Upgrade
tags:
keywords:
summary: Upgrade to the latest version of the Docker Version Manager
sidebar: home_sidebar
permalink: upgrade.html
folder:
---

{% include warning.html content="If your current version is less than 0.8.0, you must first upgrade to 0.8.0, then upgrade to the latest version.

<br/><br/>

**dvm upgrade --version 0.8.0 && dvm upgrade**

<br/><br/>

_This only applies to installations that weren't performed with a package manager, like Homebrew or Chocolately._
" %}

## Mac OS X with Homebrew
Open a terminal, and then run the following commands:

```
brew update
brew upgrade dvm
```

## Linux and Mac OS X without Homebrew
Open a terminal, and then run the following command:

```
dvm upgrade
```

## Windows
Open a terminal, and then run the following command:

```
dvm upgrade
```
