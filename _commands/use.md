---
title: use
summary: Switches to the specified version of the Docker client. If the version is not installed, dvm installs it automatically.
---

```
$ dvm use <version>
```

### Examples

```
$ dvm use 1.12.5
1.12.5 is not installed. Installing now...
Installing 1.12.5...
Now using Docker 1.12.5
```

If `<version>` is omitted, dvm uses the value of the `DOCKER_VERSION` environment variable, if present.
```
$ export DOCKER_VERSION=1.10.1
$ dvm use
Now using Docker 1.10.3
```

You can also use an alias in place of the version, either a built-in alias such
as `system` or `experimental`, or a
[user-defined alias]({{site.baseurl}}{% link _commands/alias.md %})).

```
$ dvm use experimental
Now using Docker experimental (1.12.6+78d1802)
```
