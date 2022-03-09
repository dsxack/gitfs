package nodes

import (
	"bytes"
	"context"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"io"
	"sync"
	"syscall"
)

var (
	_ fs.InodeEmbedder = (*fileNode)(nil)
	_ fs.NodeOpener    = (*fileNode)(nil)
	_ fs.NodeGetattrer = (*fileNode)(nil)
	_ fs.NodeReader    = (*fileNode)(nil)
	_ fs.NodeReleaser  = (*fileNode)(nil)
)

type fileNode struct {
	fs.Inode
	file   *object.File
	buffer *bytes.Reader
	mu     sync.Mutex
	commit *object.Commit
}

func newFileNode(file *object.File, commit *object.Commit) *fileNode {
	return &fileNode{file: file, commit: commit}
}

func (node *fileNode) Open(_ context.Context, _ uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	reader, err := node.file.Reader()
	if err != nil {
		return nil, 0, syscall.ENOENT
	}
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, 0, syscall.ENOENT
	}
	node.buffer = bytes.NewReader(buf)
	return nil, 0, 0
}

func (node *fileNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Size = uint64(node.file.Size)
	out.Mode = uint32(node.file.Mode)
	out.Mtime = uint64(node.commit.Committer.When.Unix())
	return 0
}

func (node *fileNode) Read(_ context.Context, _ fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	node.mu.Lock()
	defer node.mu.Unlock()

	n, err := node.buffer.ReadAt(dest, off)
	if err != nil {
		return nil, syscall.ENOENT
	}

	return fuse.ReadResultData(dest[:n]), 0
}

func (node *fileNode) Release(_ context.Context, _ fs.FileHandle) syscall.Errno {
	node.mu.Lock()
	defer node.mu.Unlock()
	node.buffer.Reset([]byte{})
	return 0
}
