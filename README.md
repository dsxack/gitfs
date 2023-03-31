# gitfs

[![golangci-lint](https://github.com/dsxack/gitfs/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/dsxack/gitfs/actions/workflows/golangci-lint.yml)
[![Go package](https://github.com/dsxack/gitfs/actions/workflows/go-test.yml/badge.svg)](https://github.com/dsxack/gitfs/actions/workflows/go-test.yml)

FUSE filesystem for browsing contents of git repositories revisions.

[![asciicast](https://asciinema.org/a/fWB94T6kTTGal1fum79rWYFfH.svg)](https://asciinema.org/a/fWB94T6kTTGal1fum79rWYFfH)

### Requirements

- Linux or macOS
- Installed fuse library (libfuse-dev on Debian/Ubuntu) or [osxfuse](https://osxfuse.github.io/) on macOS

### Install

```sh
go install github.com/dsxack/gitfs/cmd/gitfs@latest
```

### Usage example

with local clone of repository

```sh
gitfs mount ./path/to/local/clone/git/repository ./path/to/mount/directory
```

or with remote source (repository will be cloned into temporary directory)

```sh
gitfs mount https://github.com/dsxack/gitfs.git ./gitfs
```
