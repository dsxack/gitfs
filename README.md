# gitfs

FUSE filesystem for browsing contents of git repositories revisions.

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

### Requirements

- Installed fuse library (libfuse-dev on Debian/Ubuntu) or [osxfuse](https://osxfuse.github.io/) on macOS
