package nodes

import (
	"bytes"
	"context"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"io"
	"log/slog"
	"syscall"
)

var (
	_ fs.InodeEmbedder = (*FileNode)(nil)
	_ fs.NodeOpener    = (*FileNode)(nil)
	_ fs.NodeGetattrer = (*FileNode)(nil)
)

// FileNode is a file node.
type FileNode struct {
	fs.Inode
	file   *object.File
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

	// When remote repository is used, the storage is memory.
	// So, we can use io.ReaderAt directly.
	if readerAt, ok := reader.(io.ReaderAt); ok {
		logger.Info("File opened")
		return NewFileHandler(node.file, readerAt), 0, 0
	}

	buf, err := io.ReadAll(reader)
	if err != nil {
		logger.Error("Error while opening file", slog.String("error", err.Error()))
		return nil, 0, syscall.ENOENT
	}
	logger.Info("File opened")

	return NewFileHandler(node.file, bytes.NewReader(buf)), 0, 0
}

// Getattr gets the file attributes.
func (node *FileNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Size = uint64(node.file.Size)
	out.Mode = uint32(node.file.Mode)
	out.Mtime = uint64(node.commit.Committer.When.Unix())
	slog.Default().Debug("Got file attrs", slog.String("name", node.file.Name))
	return 0
}

var (
	_ fs.FileReader = (*FileHandler)(nil)
)

// FileHandler implements the fs.FileReader interface.
// It is used to read the file.
// It holds bytes.Reader to read the file at the given offset.
type FileHandler struct {
	reader io.ReaderAt
	file   *object.File
}

// NewFileHandler creates a new file handler.
func NewFileHandler(file *object.File, reader io.ReaderAt) *FileHandler {
	return &FileHandler{
		file:   file,
		reader: reader,
	}
}

// Read reads the file.
func (h FileHandler) Read(_ context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	logger := slog.Default().
		With(slog.String("fileName", h.file.Name)).
		With(slog.Int64("fileSize", h.file.Size)).
		With(slog.Int64("offset", off)).
		With(slog.Int("destSize", len(dest)))

	n, err := h.reader.ReadAt(dest, off)
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
