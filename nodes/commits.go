package nodes

import (
	"context"
	"github.com/dsxack/gitfs/internal/iter"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"log/slog"
	"syscall"
)

var (
	_ fs.InodeEmbedder = (*CommitsNode)(nil)
	_ fs.NodeReaddirer = (*CommitsNode)(nil)
	_ fs.NodeLookuper  = (*CommitsNode)(nil)
)

// CommitsNode is a filesystem node that represents a list of commits.
// It is a child of the root node.
// It is a directory.
// It contains a list of directories, each directory represents a commit.
type CommitsNode struct {
	fs.Inode
	repository *git.Repository
}

// NewCommitsNode creates a new CommitsNode.
func NewCommitsNode(repository *git.Repository) *CommitsNode {
	return &CommitsNode{repository: repository}
}

// Lookup looks up a commit by its hash.
// It returns a directory that represents the commit.
// It returns ENOENT if the name is not found.
func (node *CommitsNode) Lookup(ctx context.Context, hash string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	logger := slog.Default().With(slog.String("lookupCommitHash", hash))
	objectNode, err := NewObjectTreeNodeByRevision(node.repository, hash)
	if err != nil {
		logger.Warn("Error lookup commit object tree", slog.String("error", err.Error()))
		return nil, syscall.ENOENT
	}
	logger.Info("Commit object tree found")

	return node.NewInode(ctx, objectNode, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
}

// Readdir reads the list of commits.
// It returns a list of directories, each directory represents a commit.
func (node *CommitsNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	commits, err := node.repository.CommitObjects()
	if err != nil {
		return nil, syscall.ENOENT
	}
	slog.Default().Info("Dir of repository commits has been read")
	return iter.NewDirStreamAdapter[*object.Commit](
		commits, func(commit *object.Commit) fuse.DirEntry {
			return fuse.DirEntry{Name: commit.Hash.String(), Mode: syscall.S_IFDIR}
		},
	), 0
}
