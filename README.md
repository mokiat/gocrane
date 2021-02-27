# GoCrane

![master status badge](https://github.com/mokiat/gocrane/workflows/Master/badge.svg)
![release status badge](https://github.com/mokiat/gocrane/workflows/Release/badge.svg)

GoCrane is a tool to help run and rebuild applications running in a docker composed environment.

This tool is heavily inspired by [go-watcher](https://github.com/canthefason/go-watcher) but has a few improvements:
* Has option for verbose logging which can help troubleshoot issues
* Uses faster file traversal thanks to Go 1.16 `WalkDir`
* Automatically watches new directories that are created within watched directories
* Allows the configuration of build and run arguments
* Can be configured through environment variables as well as flags
* Is capable of skipping initial builds through a digest comparison

**WARNING:** This project is still in early Alpha stage and is subject to breaking changes! Your best choice is to use the versioned Docker images and not `latest`.

## User's Guide

You can install the tool with the following command:

```sh
GO111MODULE=on go get github.com/mokiat/gocrane
```

Alternatively, you can include the executable in your Docker image as follows:

```dockerfile
FROM mokiat/gocrane:latest AS gocrane

FROM ... # your base image here
# your Dockerfile stuff here
COPY --from=gocrane /bin/gocrane /bin/gocrane
```

If you are already in the project folder, you can simply run `gocrane`. Use `gocrane --help` to get detailed information on the supported commands and flags.

By default `gocrane` ignores the following files and folders:

* `.git`
* `.github`
* `.gitignore`
* `.DS_Store`
* `.vscode`
