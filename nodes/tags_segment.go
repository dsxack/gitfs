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
	_ fs.InodeEmbedder = (*tagSegmentNode)(nil)
	_ fs.NodeReaddirer = (*tagSegmentNode)(nil)
	_ fs.NodeLookuper  = (*tagSegmentNode)(nil)
)

type tagSegmentNode struct {
	fs.Inode
	repository *git.Repository
	tagPrefix  string
}

func newTagSegmentNode(repository *git.Repository, tagPrefix string) *tagSegmentNode {
	return &tagSegmentNode{repository: repository, tagPrefix: tagPrefix}
}

func (node *tagSegmentNode) Lookup(ctx context.Context, name string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	revision := revisionTagName(filepath.Join(node.tagPrefix, name))
	tags, err := node.repository.Tags()
	if err != nil {
		return nil, syscall.ENOENT
	}

	ok := referenceiter.Has(tags, revision)
	if ok {
		tagNode, err := newObjectTreeNodeByRevision(node.repository, revision)
		if err != nil {
			return nil, syscall.ENOENT
		}
		return node.NewInode(ctx, tagNode, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}

	ops := tagSegmentNode{repository: node.repository, tagPrefix: filepath.Join(node.tagPrefix, name) + tagNameSeparator}
	return node.NewInode(ctx, &ops, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
}

func (node *tagSegmentNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	tags, err := node.repository.Tags()
	if err != nil {
		return nil, syscall.ENOENT
	}
	dirEntries := set.NewSet[fuse.DirEntry]()
	_ = tags.ForEach(func(tagRef *plumbing.Reference) error {
		tagName := bareTagName(tagRef.Name().String())
		if !strings.HasPrefix(tagName, node.tagPrefix) {
			return nil
		}
		segments := strings.Split(strings.TrimPrefix(tagName, node.tagPrefix), tagNameSeparator)
		dirEntries.Add(fuse.DirEntry{Name: segments[0], Mode: syscall.S_IFDIR})
		return nil
	})
	return fs.NewListDirStream(dirEntries.Values()), 0
}
