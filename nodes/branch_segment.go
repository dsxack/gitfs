package nodes

import (
	"context"
	"github.com/dsxack/gitfs/internal/iter"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"log/slog"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	_ fs.InodeEmbedder = (*BranchSegmentNode)(nil)
	_ fs.NodeReaddirer = (*BranchSegmentNode)(nil)
	_ fs.NodeLookuper  = (*BranchSegmentNode)(nil)
)

// BranchSegmentNode is a node that represents a segment of a branch name.
// For example, if the branch name is "foo/bar/baz", then there will be three
// BranchSegmentNodes, one for "foo", one for "bar", and one for "baz".
type BranchSegmentNode struct {
	fs.Inode
	repository   *git.Repository
	branchPrefix string
}

// NewBranchSegmentNode creates a new BranchSegmentNode.
func NewBranchSegmentNode(repository *git.Repository, branchPrefix string) *BranchSegmentNode {
	return &BranchSegmentNode{repository: repository, branchPrefix: branchPrefix}
}

// Lookup returns the child node with the given name.
// If the name is a branch name, then a new ObjectTreeNode is returned.
// Otherwise, a new BranchSegmentNode is returned.
// It returns ENOENT if the name is not found.
func (node *BranchSegmentNode) Lookup(ctx context.Context, name string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	logger := slog.Default().
		With(slog.String("lookupBranchSegmentName", name)).
		With(slog.String("branchPrefix", node.branchPrefix))
	revision := revisionBranchName(filepath.Join(node.branchPrefix, name))
	branches, err := node.repository.Branches()

	if err != nil {
		logger.Error("Error lookup branch segment", slog.String("error", err.Error()))
		return nil, syscall.ENOENT
	}
	ok, hasPrefix := iter.HasReference(branches, revision)
	if ok {
		branchNode, err := NewObjectTreeNodeByRevision(node.repository, revision)
		if err != nil {
			logger.Error("Error lookup branch object tree", slog.String("error", err.Error()))
			return nil, syscall.ENOENT
		}
		logger.Info("Branch object tree found")
		return node.NewInode(ctx, branchNode, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}
	if !hasPrefix {
		logger.Info("No branch found")
		return nil, syscall.ENOENT
	}
	logger.Info("Branch segment found")

	return node.NewInode(
		ctx,
		NewBranchSegmentNode(node.repository, filepath.Join(node.branchPrefix, name)+branchNameSeparator),
		fs.StableAttr{Mode: syscall.S_IFDIR},
	), 0
}

// Readdir returns the child nodes of this node.
// The child nodes are the branches that start with the branch prefix.
// For example, if branch names are "foo/bar" and "foo/buz", then
// will return "foo" directory with two children, "bar" and "buz".
func (node *BranchSegmentNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	var branches iter.Iter[*plumbing.Reference]
	var err error

	branches, err = node.repository.Branches()
	if err != nil {
		return nil, syscall.ENOENT
	}
	branches = iter.NewFilterIter(branches, func(branchRef *plumbing.Reference) bool {
		branchName := bareBranchName(branchRef.Name().String())
		return strings.HasPrefix(branchName, node.branchPrefix)
	})

	return iter.NewDirStreamAdapter[*plumbing.Reference](
		branches,
		func(branchRef *plumbing.Reference) fuse.DirEntry {
			branchName := bareBranchName(branchRef.Name().String())
			segments := strings.Split(strings.TrimPrefix(branchName, node.branchPrefix), branchNameSeparator)
			return fuse.DirEntry{Name: segments[0], Mode: syscall.S_IFDIR}
		},
	), 0
}
