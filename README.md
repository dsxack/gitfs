# gitfs

FUSE filesystem for browsing contents of git repositories revisions.

### Install

```sh
go install github.com/dsxack/gitfs/cmd/gitfs@latest
```

### Usage example

```sh
gitfs mount path/to/git/repository ./path/to/mount/directory
```
