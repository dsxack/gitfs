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
	_ fs.InodeEmbedder = (*tagsNode)(nil)
	_ fs.NodeReaddirer = (*tagsNode)(nil)
	_ fs.NodeLookuper  = (*tagsNode)(nil)
)

const tagNameSeparator = string(filepath.Separator)

type tagsNode struct {
	fs.Inode
	repository *git.Repository
}

func newTagsNode(repository *git.Repository) *tagsNode {
	return &tagsNode{repository: repository}
}

func (node *tagsNode) Lookup(ctx context.Context, name string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	revision := revisionTagName(name)
	tags, err := node.repository.Tags()
	if err != nil {
		return nil, syscall.ENOENT
	}
	ok := referenceiter.Has(tags, revision)
	if ok {
		branchNode, err := newObjectTreeNodeByRevision(node.repository, revision)
		if err != nil {
			return nil, syscall.ENOENT
		}
		return node.NewInode(ctx, branchNode, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}
	return node.NewInode(
		ctx,
		newTagSegmentNode(node.repository, name+tagNameSeparator),
		fs.StableAttr{Mode: syscall.S_IFDIR},
	), 0
}

func (node *tagsNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	tagRefs, err := node.repository.Tags()
	if err != nil {
		return nil, syscall.ENOENT
	}
	dirEntries := set.NewSet[fuse.DirEntry]()
	_ = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		tagName := bareTagName(tagRef.Name().String())
		segments := strings.Split(tagName, tagNameSeparator)
		dirEntries.Add(fuse.DirEntry{Name: segments[0], Mode: syscall.S_IFDIR})
		return nil
	})
	return fs.NewListDirStream(dirEntries.Values()), 0
}

const revisionTagPrefix = "refs/tags/"

func bareTagName(revision string) string {
	return strings.TrimPrefix(revision, revisionTagPrefix)
}

func revisionTagName(tag string) string {
	return revisionTagPrefix + tag
}
