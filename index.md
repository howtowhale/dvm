---
title: Docker Version Manager
keywords:
- docker
- dvm
tags:
permalink: index.html
summary: The Docker Version Manager (dvm) is a cross-platform command-line tool that helps you install and switch between Docker clients
---

The Docker Version Manager (dvm) is a cross-platform command-line tool that helps
you install and switch between Docker clients. It also helps both avoid and address the following Docker client/server API mismatch error message:

```bash
Error response from daemon: client is newer than server (client API version: 1.21, server API version: 1.20)
```

**Note:** dvm manipulates the PATH variable of the current shell
session, and so the changes that dvm makes are temporary.


### Quick Start
After you [install]({{site.baseurl}}{% link pages/install.md %}) dvm, run 
the [detect]({{site.baseurl}}{% link _commands/detect.md %}) command and you 
are all ready to start using Docker.
 
```
$ dvm detect
1.13.1 is not installed. Installing now...
Installing 1.13.1...
Now using Docker 1.13.1

$ docker --version
Docker version 1.13.1, build 092cba3
```

### Global Flags
You can use the following global flags with any
of the commands. Global flags should be specified before the command,
for example: `dvm --silent install 1.9.0`.

* `--silent`

  Suppresses a command's normal output. Errors are still displayed.
* `--debug`

  Prints additional debug information.
