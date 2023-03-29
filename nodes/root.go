package nodes

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"syscall"
)

var (
	_ fs.InodeEmbedder = (*RootNode)(nil)
	_ fs.NodeLookuper  = (*RootNode)(nil)
	_ fs.NodeReaddirer = (*RootNode)(nil)
)

// RootNode is the root node of the filesystem.
// It contains the following subdirectories:
// - branches: list of branches
// - commits: list of commits
// - tags: list of tags
type RootNode struct {
	fs.Inode
	repository *git.Repository
}

// NewRootNode creates a new RootNode.
func NewRootNode(repository *git.Repository) *RootNode {
	return &RootNode{repository: repository}
}

// Lookup returns the inode for the given name.
// It returns ENOENT if the name is not found.
func (node *RootNode) Lookup(ctx context.Context, name string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	switch name {
	case "branches":
		ops := NewBranchesNode(node.repository)
		return node.NewInode(ctx, ops, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	case "commits":
		ops := NewCommitsNode(node.repository)
		return node.NewInode(ctx, ops, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	case "tags":
		ops := NewTagsNode(node.repository)
		return node.NewInode(ctx, ops, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}
	return nil, syscall.ENOENT
}

// Readdir returns the list of entries in the directory.
func (node *RootNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	return fs.NewListDirStream([]fuse.DirEntry{
		{Name: "branches", Mode: syscall.S_IFDIR},
		{Name: "commits", Mode: syscall.S_IFDIR},
		{Name: "tags", Mode: syscall.S_IFDIR},
	}), 0
}
