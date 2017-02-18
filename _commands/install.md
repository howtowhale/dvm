---
title: install
summary: Installs the specified version of the Docker client
---

```bash
dvm install <version>
```

If `<version>` is omitted, dvm uses the value of the `DOCKER_VERSION` environment variable, if present.

Run `dvm install experimental` to install the latest, nightly build.

### Optional Arguments
* `--mirror-url`

  Specify an alternate URL from which to download the Docker client. Defaults to https://get.docker.com/builds, unless the environment variable `DVM_MIRROR_URL` is set.
* `--nocheck`

  Do not check if version exists (use with caution). Defaults to false, unless the environment variable `DVM_NOCHECK` is set.

### Example

```
$ dvm install 1.12.1
Installing 1.12.1...
Now using Docker 1.12.1
```

### Alias
```
dvm -i <version>
```
