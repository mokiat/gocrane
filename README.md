# GoCrane

![master status badge](https://github.com/mokiat/gocrane/workflows/Master/badge.svg)
![release status badge](https://github.com/mokiat/gocrane/workflows/Release/badge.svg)

GoCrane is a tool to help run and automatically rebuild applications in a `Docker` environment.

This tool is heavily inspired by [go-watcher](https://github.com/canthefason/go-watcher) but has a few improvements:

* You can enable verbose logging to troubleshoot any issues
* It uses faster file traversal thanks to `WalkDir` in Go 1.16
* Folders created within watched folders are automatically watched
    * This includes nested folders
    * It also handles drag-drop situations
* You can configure build and run arguments
* All configurations can also be specified through environment variables
    * This reduces what would normally be one-line clutter
* Is capable of skipping initial builds through a source digest comparison
* It has a configurable batching duration, so that you can avoid excessive builds when fetching files
* You can configure files and folders that should only trigger a restart and not a rebuild
    * This can be useful for configuration files

**WARNING:** This project is still young and is subject to breaking changes. Your best choice is to use the versioned [Docker images](https://hub.docker.com/r/mokiat/gocrane/tags) and not `latest`.

## User's Guide

It is important to understand how GoCrane works, so that you can configure it optimally and avoid unexpected behavior.

GoCrane uses a number of configuration flags / options that relate to file paths. Such configuration options can often take a combination of folder paths, file paths, and glob patterns. You should check the documentation for each individual flag to see which combination of these are allowed.

*Note:* Whenever GoCrane deals with paths it tries to convert those to absolute. Failure to do so would lead to errors and if there are multiple absolute forms of the same path, then the behavior is undefined. For most scenarios this should not be a problem but in general avoid symbolic links.

GoCrane distinguishes between paths (folders and files) and glob patterns through a special prefix (`*/`) that glob patterns need to have. Globs are a means to express path segments through a wildcard pattern. Only individual path segments may be represented. That is, the pattern `*/*data` (representing all folders that have a `data` suffix) is acceptable, whereas `*/hello/world` is not.

Here we will take a look at the most important flags that GoCrane provides. For the rest and their respective aliases and environment variable names, you should check `gocrane --help`.

* `dir` - This flag specifies a folder that GoCrane should watch. It can be specified multiple times in which case GoCrane would watch all the specified folders. In addition, GoCrane watches recursively all sub-folders. You should NOT specify any files. You may use a glob pattern, through its purpose would only be to supersede any watch exceptions imposed by `exclude-dir`. Most other flags are confined to the boundaries of the `dir` flag (i.e. if you were to specify a folder that is not contained by a watched folder, it would be ignored). By default this is set to the `./` (i.e. `PWD`) folder.

* `exclude-dir` - This flag specifies a folder or glob pattern for folders that should be ignored from watching. It is useful when you have sub-folders of watched folders that you don't want to be watched or evaluated (e.g. `.git`). This flag can be specified multiple times in which case if a directory matches any of the specified values it will be ignored. By default GoCrane sets this flag to a collection of reasonable glob patterns (like `.git`, `.vscode`, etc.). If you were to specify this flag, you would need to relist those. Even if a folder is excluded from watching via this flag, if a nested folder is explicitly marked as watched via `dir` it (and its children) will be watched. Same goes for any glob patterns.

* `source` - This flag specifies a folder, file, or glob pattern that indicates what files should be constituted as source code. This helps GoCrane decide whether a file change event should retrigger a rebuild (and a subsequent restart) of the application. It is also used as means to determine which files should be used to calculate the digest. You can specify this flag any number of times and if a path matches any of the specified values, it will be considered as source code. By default GoCrane sets this flag to `*/*.go`. This should be sufficient for most use cases but if, for example, you are using some type of file embedding, then you may want to add non-go files as well, so that a rebuild would be triggered accordingly.

* `exclude-source` - This flag specifies a folder, file, or glob pattern for files that should not be considered as source code, even if they match a `source` flag value. This flag can be specified multiple times. By default GoCrane sets this to `*/*_test.go` so that test files do not trigger a rebuild.

* `resource` - This flag specifies a folder, file, or glob pattern for files that should be considered as resources. A change to such files would make GoCrane restart, but NOT rebuild, your application. This flag can be specified multiple times. It is mostly useful if your application reads data from the filesystem (e.g. configuration files) during startup. By default GoCrane does not have this flag set, hence no file is considred a resource.

* `exclude-resource` - This flag specifies a folder, file, or glob pattern for files that should not be considered as resources, even if they match a `resource` flag value. It can be specified multiple times. By default GoCrane has this file set to a number of common files (e.g. `Dockerfile`, `README.md`, `.gitignore`) that are unlikely to be used by your application. If you set this flag, you would need to list your own defaults.

* `main` - This flag specifies the folder where your application's main package is located. Unlike previous flags, this one can point to a location that is not specified through a `dir` flag, however, this would rarely ever be meaningful, since it is likely that you would like to have GoCrane rebuild and restart your application when a Go file in the main package changes.

* `binary` - This flag specifies an executable that GoCrane should use when starting up, instead of rebuilding your application, as the latter could be a CPU-intensive operation, especially if you have multiple GoCrane-managed applications starting at the same time. You should only specify this flag with the `gocrane run` command if the binary you reference has been built with `gocrane build`, since GoCrane would look for a `<executable>.dig` file to compare digest sums. If the digest sums don't match (which means that the source code you have mounted in the container has changed since `gocrane build` was used), GoCrane would default to triggering a rebuild and will not use the executable.

### Using in Docker-Compose

The main purpose of gocrane is to be used within a `Docker` or `docker-compose` environment. You can check the included [example](https://github.com/mokiat/gocrane/tree/master/example), which showcases how GoCrane can be used to detect changes while you develop a project locally.

**Note:** If you want to make use of the caching behavior, make sure to use `docker-compose build --no-cache example` once you are satisfied with your changes to `example`. Future starts of `docker-compose up` would not trigger a build until you make changes to the source code on the host machine.

### Using locally

To use the tool locally, you can get it as follows (Go 1.19+):

```sh
go install github.com/mokiat/gocrane@latest
```

Normally, it would be sufficient to run `gocrane run` in the root folder of your project. If the `main` package is not located in the root folder (e.g. in `./cmd/executable/`), you would need to use the `main` flag to specify that.

For more information, consult `gocrane --help`.
