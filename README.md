# GoCrane

GoCrane is a tool to help run and rebuild applications running in a docker composed environment.

This tool is heavily inspired by [go-watcher](https://github.com/canthefason/go-watcher) but makes a few improvements (verbose printing, automatic new directory watching, using cached builds). Normally I would prefer to contribute to the original project but I needed to troubleshoot a problem and found it easier/faster to roll out my own.

**WARNING:** This project is still in early Alpha stage and is subject to breaking changes (e.g. flag names).

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

If you are already in the project folder, you can simply run `gocrane`.

The tool supports the following flags for more custom configuration:

* `--verbose, -v` - Enables verbose logging. Good for troubleshooting issues.
* `--path, -p` - Specifies a path to watch recursively. This flag can be specified multiple times if multiple root directories should be watched.
* `--exclude, -e` - Specifies a path to exclude from watching. This flag can be specified multiple times if multiple directory branches are to be excluded.
* `--glob-exclude, -ge` - Specifies a glob pattern for files and directories that should be excluded. Can be specified multiple times. The pattern works only for a single path segment (e.g. `vendor`, `.DS_Store`, `*_test.go`)
* `--run, -r` - Specifies the folder that includes the `main` package to be run.
* `--cache, -c` - Specifies a pre-build executable to use the first time. This is useful when using `gocrane` with Docker images that have the built executable available to avoid unnecessary initial builds on startup.

By default `gocrane` excludes the following glob patterns:

* `.git`
* `.DS_Store`
* `.vscode`
