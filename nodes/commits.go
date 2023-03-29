package nodes

import (
	"context"
	"github.com/dsxack/gitfs/internal/set"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"syscall"
)

var (
	_ fs.InodeEmbedder = (*commitsNode)(nil)
	_ fs.NodeReaddirer = (*commitsNode)(nil)
	_ fs.NodeLookuper  = (*commitsNode)(nil)
)

type commitsNode struct {
	fs.Inode
	repository *git.Repository
}

func newCommitsNode(repository *git.Repository) *commitsNode {
	return &commitsNode{repository: repository}
}

func (node *commitsNode) Lookup(ctx context.Context, hash string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	objectNode, err := newObjectTreeNodeByRevision(node.repository, hash)
	if err != nil {
		return nil, syscall.ENOENT
	}
	return node.NewInode(ctx, objectNode, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
}

func (node *commitsNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	commits, err := node.repository.CommitObjects()
	if err != nil {
		return nil, syscall.ENOENT
	}
	dirEntries := set.NewSet[fuse.DirEntry]()
	_ = commits.ForEach(func(commit *object.Commit) error {
		dirEntries.Add(fuse.DirEntry{Name: commit.Hash.String(), Mode: syscall.S_IFDIR})
		return nil
	})
	return fs.NewListDirStream(dirEntries.Values()), 0
}
