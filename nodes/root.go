package nodes

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"syscall"
)

var (
	_ fs.InodeEmbedder = (*rootNode)(nil)
	_ fs.NodeLookuper  = (*rootNode)(nil)
	_ fs.NodeReaddirer = (*rootNode)(nil)
)

type rootNode struct {
	fs.Inode
	repository *git.Repository
}

func NewRootNode(repository *git.Repository) *rootNode {
	return &rootNode{repository: repository}
}

func (node *rootNode) Lookup(ctx context.Context, name string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	switch name {
	case "branches":
		ops := newBranchesNode(node.repository)
		return node.NewInode(ctx, ops, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	case "commits":
		ops := newCommitsNode(node.repository)
		return node.NewInode(ctx, ops, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	case "tags":
		ops := newTagsNode(node.repository)
		return node.NewInode(ctx, ops, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}
	return nil, syscall.ENOENT
}

func (node *rootNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	return fs.NewListDirStream([]fuse.DirEntry{
		{Name: "branches", Mode: syscall.S_IFDIR},
		{Name: "commits", Mode: syscall.S_IFDIR},
		{Name: "tags", Mode: syscall.S_IFDIR},
	}), 0
}
