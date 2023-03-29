package nodes

import (
	"context"
	"github.com/dsxack/gitfs/internal/referenceiter"
	"github.com/dsxack/gitfs/internal/set"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
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
	revision := revisionBranchName(filepath.Join(node.branchPrefix, name))
	branches, err := node.repository.Branches()
	if err != nil {
		return nil, syscall.ENOENT
	}
	ok := referenceiter.Has(branches, revision)
	if ok {
		branchNode, err := NewObjectTreeNodeByRevision(node.repository, revision)
		if err != nil {
			return nil, syscall.ENOENT
		}
		return node.NewInode(ctx, branchNode, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}
	ok = referenceiter.HasPrefix(branches, revision+branchNameSeparator)
	if !ok {
		return nil, syscall.ENOENT
	}
	ops := BranchSegmentNode{repository: node.repository, branchPrefix: filepath.Join(node.branchPrefix, name) + branchNameSeparator}
	return node.NewInode(ctx, &ops, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
}

// Readdir returns the child nodes of this node.
// The child nodes are the branches that start with the branch prefix.
// For example, if branch names are "foo/bar" and "foo/buz", then
// will return "foo" directory with two children, "bar" and "buz".
func (node *BranchSegmentNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	branches, err := node.repository.Branches()
	if err != nil {
		return nil, syscall.ENOENT
	}
	dirEntries := set.NewSet[fuse.DirEntry]()
	_ = branches.ForEach(func(branchRef *plumbing.Reference) error {
		branchName := bareBranchName(branchRef.Name().String())
		if !strings.HasPrefix(branchName, node.branchPrefix) {
			return nil
		}
		segments := strings.Split(strings.TrimPrefix(branchName, node.branchPrefix), branchNameSeparator)
		dirEntries.Add(fuse.DirEntry{Name: segments[0], Mode: syscall.S_IFDIR})
		return nil
	})
	return fs.NewListDirStream(dirEntries.Values()), 0
}
