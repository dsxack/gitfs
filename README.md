# gitfs

[![golangci-lint](https://github.com/dsxack/gitfs/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/dsxack/gitfs/actions/workflows/golangci-lint.yml)
[![Go package](https://github.com/dsxack/gitfs/actions/workflows/go-test.yml/badge.svg)](https://github.com/dsxack/gitfs/actions/workflows/go-test.yml)
[![codecov](https://codecov.io/gh/dsxack/gitfs/branch/master/graph/badge.svg?token=JG8wPYoqAq)](https://codecov.io/gh/dsxack/gitfs)

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

Mount with local repository clone or 
remote repository url (in this case repository will be cloned into temporary directory)

```sh
gitfs mount <local repository url> <mountpoint>
```

Mount in daemon mode

```sh
gitfs mount -d <repository> <mountpoint>
```

Umount previously mounted in daemon mode filesystem

```sh
gitfs umount <mountpoint>
```