# gitfs

[![golangci-lint](https://github.com/dsxack/gitfs/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/dsxack/gitfs/actions/workflows/golangci-lint.yml)
[![Go package](https://github.com/dsxack/gitfs/actions/workflows/go-test.yml/badge.svg)](https://github.com/dsxack/gitfs/actions/workflows/go-test.yml)

FUSE filesystem for browsing contents of git repositories revisions.

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

Example of directory structure of mounted repository

```shell
tree ./gitfs
./gitfs
├── branches
│         └── master
│             ├── README.md
│             ├── cmd
│             │         └── gitfs
│             │             ├── main.go
│             │             ├── mount.go
│             │             ├── mount_darwin.go
│             │             ├── mount_linux.go
│             │             └── root.go
│             ├── go.mod
│             ├── go.sum
│             ├── internal
│             │         ├── referenceiter
│             │         │         └── referenceiter.go
│             │         ├── set
│             │         │         └── set.go
│             │         └── testdata
│             │             ├── testmain.go
│             │             └── testrepo.zip
│             └── nodes
│                 ├── branch_segment.go
│                 ├── branch_segment_test.go
│                 ├── branches.go
│                 ├── branches_test.go
│                 ├── commits.go
│                 ├── commits_test.go
│                 ├── file.go
│                 ├── object.go
│                 ├── root.go
│                 ├── tags.go
│                 ├── tags_segment.go
│                 ├── tags_segment_test.go
│                 └── tags_test.go
├── commits
│         ├── 0876607f6575d31fc97ebbbc2127313c32ed1fe1
│         │         ├── README.md
│         │         ├── ...
│         ├── 1bd32494e702af332e5e6616f7a57943863eaa76
│         │         ├── README.md
│         │         ├── ...
│         ├── 1fb3a9c1773f0b616db1752551635067c8b35fb7
│         │         ├── README.md
│         │         ├── ...
│         ├── 6f176b46ba8e3dcf0868a49e01745117d58930b4
│         │         ├── README.md
│         │         ├── ...
│         ├── ...
└── tags
    └── v1.0.0
        ├── README.md
        ├── ...

149 directories, 342 files
```