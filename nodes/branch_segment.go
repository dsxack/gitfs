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
	_ fs.InodeEmbedder = (*branchSegmentNode)(nil)
	_ fs.NodeReaddirer = (*branchSegmentNode)(nil)
	_ fs.NodeLookuper  = (*branchSegmentNode)(nil)
)

type branchSegmentNode struct {
	fs.Inode
	repository   *git.Repository
	branchPrefix string
}

func newBranchSegmentNode(repository *git.Repository, branchPrefix string) *branchSegmentNode {
	return &branchSegmentNode{repository: repository, branchPrefix: branchPrefix}
}

func (node *branchSegmentNode) Lookup(ctx context.Context, name string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	revision := filepath.Join(node.branchPrefix, name)
	branches, err := node.repository.Branches()
	if err != nil {
		return nil, syscall.ENOENT
	}

	ok := referenceiter.Has(branches, revision)
	if ok {
		branchNode, err := newBranchTreeNodeByRevision(node.repository, revision)
		if err != nil {
			return nil, syscall.ENOENT
		}
		return node.NewInode(ctx, branchNode, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}

	ops := branchSegmentNode{repository: node.repository, branchPrefix: filepath.Join(node.branchPrefix, name) + branchNameSeparator}
	return node.NewInode(ctx, &ops, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
}

func (node *branchSegmentNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	branches, err := node.repository.Branches()
	if err != nil {
		return nil, syscall.ENOENT
	}
	dirEntries := set.NewSet[fuse.DirEntry]()
	_ = branches.ForEach(func(branchRef *plumbing.Reference) error {
		branchName := branchRef.Name().String()
		if !strings.HasPrefix(branchName, node.branchPrefix) {
			return nil
		}
		segments := strings.Split(strings.TrimPrefix(branchName, node.branchPrefix), branchNameSeparator)
		dirEntries.Add(fuse.DirEntry{Name: segments[0], Mode: syscall.S_IFDIR})
		return nil
	})
	return fs.NewListDirStream(dirEntries.Values()), 0
}
