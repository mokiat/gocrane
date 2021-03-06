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

### Understanding File Configuration

There are a number of configurations that are vital to gocrane working fast and correctly. The most important is probably the `include` flag. It should contain files and folders that you would like gocrane to keep watch of. This does not necessary mean that gocrane would trigger a build or a restart if one of those file changes (this is controlled with additional flags) but indicates to gocrane which files it should explore and keep in mind.

The `exclude` flag on the other hand can be used to indicate files, directories, or globs that should be ignored at all cost. This is useful if you would not like gocrane to explore deep folders that are not related to the project (e.g. `.git`). By default gocrane ignores a number of common non-project directories.

The `source` flag can be used to indicate which paths should be considered relevant for building the project. When gocrane detects a change to one of those files, it will trigger a rebuild of the project. Keep in mind that the expressions specified in this flag need to include paths that are part of `include`, otherwise they would not be considered. By default gocrane uses `*.go` but you could chnage that to include additional ones, if for example you are using file embedding.

The `resource` flag on the other hand indicates which paths should be considered relevant for restarting the project. For exmaple, if your project is an HTTP server that serves files from a project folder, you may wont to mark that folder as a resource, so that if there is a change to it, gocrane would trigger a restart, instead of a build (unless the change also matches the `source` flag).
