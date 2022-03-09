package nodes

import (
	"context"
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
	_ fs.InodeEmbedder = (*branchesNode)(nil)
	_ fs.NodeReaddirer = (*branchesNode)(nil)
	_ fs.NodeLookuper  = (*branchesNode)(nil)
)

const branchNameSeparator = string(filepath.Separator)

type branchesNode struct {
	fs.Inode
	repository *git.Repository
}

func newBranchesNode(repository *git.Repository) *branchesNode {
	return &branchesNode{repository: repository}
}

func (node *branchesNode) Lookup(ctx context.Context, name string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	return node.NewInode(
		ctx,
		newBranchSegmentNode(node.repository, name+branchNameSeparator),
		fs.StableAttr{Mode: syscall.S_IFDIR},
	), 0
}

func (node *branchesNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	branches, err := node.repository.Branches()
	if err != nil {
		return nil, syscall.ENOENT
	}
	dirEntries := set.NewSet[fuse.DirEntry]()
	_ = branches.ForEach(func(branchRef *plumbing.Reference) error {
		segments := strings.Split(branchRef.Name().String(), branchNameSeparator)
		dirEntries.Add(fuse.DirEntry{Name: segments[0], Mode: syscall.S_IFDIR})
		return nil
	})
	return fs.NewListDirStream(dirEntries.Values()), 0
}
