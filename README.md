# gitfs

[![golangci-lint](https://github.com/dsxack/gitfs/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/dsxack/gitfs/actions/workflows/golangci-lint.yml)
[![Go package](https://github.com/dsxack/gitfs/actions/workflows/go-test.yml/badge.svg)](https://github.com/dsxack/gitfs/actions/workflows/go-test.yml)
[![codecov](https://codecov.io/gh/dsxack/gitfs/branch/master/graph/badge.svg?token=JG8wPYoqAq)](https://codecov.io/gh/dsxack/gitfs)

FUSE filesystem for browsing contents of git repositories revisions.

[![asciicast](https://asciinema.org/a/574704.svg)](https://asciinema.org/a/574704)

### Requirements

- Linux or macOS
- Installed fuse library (libfuse-dev on Debian/Ubuntu) or [macFUSE](https://osxfuse.github.io/) on macOS

### Install

with Homebrew
```sh
brew install dsxack/tap/gitfs
```

or with Go
```sh
go install github.com/dsxack/gitfs/cmd/gitfs@latest
```

### Usage

Mount
```sh
gitfs mount <repository url> <mountpoint>
```

Mounting local repository
```sh
gitfs mount /home/dsxack/work/project /mnt/project
```

Mounting remote repository (repository will be cloned into memory)
```sh
gitfs mount https://github.com/dsxack/go /mnt/go
```

Mount in daemon mode
```sh
gitfs mount -d <repository> <mountpoint>
```

Umount previously mounted in daemon mode filesystem
```sh
gitfs umount <mountpoint>
```

Mount with verbose logging for debugging reasons
```sh
# Info
gitfs mount https://github.com/dsxack/go /mnt/go -v

# Debug
gitfs mount https://github.com/dsxack/go /mnt/go -vv

# Trace
gitfs mount https://github.com/dsxack/go /mnt/go -vvv
```

### License

[MIT](LICENSE)