# GoCrane

![master status badge](https://github.com/mokiat/gocrane/workflows/Master/badge.svg)
![release status badge](https://github.com/mokiat/gocrane/workflows/Release/badge.svg)

GoCrane is a tool to help run and rebuild applications running in a docker-compose environment.

This tool is heavily inspired by [go-watcher](https://github.com/canthefason/go-watcher) but has a few improvements:
* Has option for verbose logging which can help troubleshoot issues
* Uses faster file traversal thanks to Go 1.16 `WalkDir`
* Automatically watches new directories that are created within watched directories
* Allows the configuration of build and run arguments
* Can be configured through environment variables as well as flags
* Is capable of skipping initial builds through a digest comparison

**WARNING:** This project is still in early Alpha stage and is subject to breaking changes! Your best choice is to use the versioned Docker images and not `latest`.

## User's Guide

The main purpose of gocrane is to be used within a `docker-compose` environment. You can check the included [example](https://github.com/mokiat/gocrane/tree/master/example), which showcases how gocrane can be used to detect changes while you develop a project locally.

**Note:** If you want to make use of the caching behavior, make sure to use `docker-compose build example` once you are satisfied with your changes to example. Future starts of `docker-compose up` would not trigger a build until you make changes to the source code on the host machine.

To use the tool locally, you can get it as follows:

```sh
GO111MODULE=on go get github.com/mokiat/gocrane
```

Use can use `gocrane --help` to get detailed information on the supported commands and flags.

By default `gocrane` ignores the following files and folders:
* `.git`
* `.github`
* `.gitignore`
* `.DS_Store`
* `.vscode`
