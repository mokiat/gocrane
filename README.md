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

#### The `dir` flag

To start off, you need to tell gocrane which folders it should watch for changes. This where the `dir` flag comes into play.
You can specify it multiple times and while you could specify nested directories, that is suboptimal and unnecessary, since gocrane
explores directories in depth.

By default gocrane sets this to `./`.

#### The `exclude-dir` flag

There might be cases when certain directories are not relevant for the project. You can use the `exclude-dir` flag to specify paths or globs for files or folders that should be ignored from watching. A good example is `.git` and this is why it is set by default (as well as some other ones). However, if you specify this flag, it will be disabled, so you would need to specify all excludes on your own.

#### The `source` flag

The gocrane tool needs some way to know which files to treat as source code. This is not required for the building of the executable, which works well without this settings. Rather, this is required to avoid triggering unnecessary builds when irrelevant files (e.g. `README.md`) are changed. Furthermore, it is used as a means to calculate the source digest.

By default gocrane sets this to `*.go` but you may want to reconfigure it if for example you are using the new `embed` capability of Go and would like to have other non-source-code resources trigger a rebuild.

#### The `exclude-source` flag

While the `source` flag does a fairly good job, there is still room for optimiziation. In most cases files like `_test.go` are irrelevant for the build of the final executable. This is why, the `exclude-source` flag can be used to specify such patterns.

#### The `resource` flag

If you specify this flag and if changed files or folders match it, gocrane will not trigger a build but rather only a restart. This is useful if your execuable is an HTTP server for example and you have a resource folder with HTML content.

#### The `exclude-resource` flag

This flag works in the same way as the `exclude-source` flag, except that it applies to resources.
