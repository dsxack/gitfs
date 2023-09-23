package nodes

import (
	"bytes"
	"context"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"io"
	"log/slog"
	"sync"
	"syscall"
)

var (
	_ fs.InodeEmbedder = (*FileNode)(nil)
	_ fs.NodeOpener    = (*FileNode)(nil)
	_ fs.NodeGetattrer = (*FileNode)(nil)
	_ fs.NodeReader    = (*FileNode)(nil)
	_ fs.NodeFlusher   = (*FileNode)(nil)
)

// FileNode is a file node.
type FileNode struct {
	fs.Inode
	file   *object.File
	buffer *bytes.Reader
	mu     sync.Mutex
	commit *object.Commit
}

// NewFileNode creates a new file node.
func NewFileNode(file *object.File, commit *object.Commit) *FileNode {
	return &FileNode{file: file, commit: commit}
}

// Open opens the file.
func (node *FileNode) Open(_ context.Context, _ uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	logger := slog.Default().
		With(slog.String("fileName", node.file.Name)).
		With(slog.Int64("fileSize", node.file.Size))
	reader, err := node.file.Reader()
	if err != nil {
		logger.Error("Error while opening file", slog.String("error", err.Error()))
		return nil, 0, syscall.ENOENT
	}
	buf, err := io.ReadAll(reader)
	if err != nil {
		logger.Error("Error while opening file", slog.String("error", err.Error()))
		return nil, 0, syscall.ENOENT
	}
	node.buffer = bytes.NewReader(buf)
	logger.Info("File opened")

	return nil, 0, 0
}

// Getattr gets the file attributes.
func (node *FileNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Size = uint64(node.file.Size)
	out.Mode = uint32(node.file.Mode)
	out.Mtime = uint64(node.commit.Committer.When.Unix())
	slog.Default().Debug("Got file attrs", slog.String("name", node.file.Name))
	return 0
}

// Read reads the file.
func (node *FileNode) Read(_ context.Context, _ fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	node.mu.Lock()
	defer node.mu.Unlock()

	logger := slog.Default().
		With(slog.String("fileName", node.file.Name)).
		With(slog.Int64("fileSize", node.file.Size)).
		With(slog.Int64("offset", off)).
		With(slog.Int("destSize", len(dest))).
		With(slog.Int("bufferSize", node.buffer.Len()))

	n, err := node.buffer.ReadAt(dest, off)
	if err == io.EOF {
		logger.Info("File read")
		return fuse.ReadResultData(dest[off : off+int64(n)]), 0
	}
	if err != nil {
		logger.Error("Error while reading file", slog.String("error", err.Error()))
		return nil, syscall.ENOENT
	}
	logger.Info("File read")

	return fuse.ReadResultData(dest[:n]), 0
}

func (node *FileNode) Flush(_ context.Context, _ fs.FileHandle) syscall.Errno {
	node.mu.Lock()
	defer node.mu.Unlock()
	node.buffer.Reset([]byte{})
	slog.Default().Info("File flushed", slog.String("fileName", node.file.Name))
	return 0
}
