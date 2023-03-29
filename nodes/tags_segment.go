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
	_ fs.InodeEmbedder = (*TagSegmentNode)(nil)
	_ fs.NodeReaddirer = (*TagSegmentNode)(nil)
	_ fs.NodeLookuper  = (*TagSegmentNode)(nil)
)

// TagSegmentNode is a node that represents a segment of a tag name.
// For example, if the tag name is "release/v1.0.0" then there will be two
// TagSegmentNodes, one for "release" and one for "v1.0.0".
type TagSegmentNode struct {
	fs.Inode
	repository *git.Repository
	tagPrefix  string
}

// NewTagSegmentNode creates a new TagSegmentNode.
func NewTagSegmentNode(repository *git.Repository, tagPrefix string) *TagSegmentNode {
	return &TagSegmentNode{repository: repository, tagPrefix: tagPrefix}
}

// Lookup returns the child node with the given name.
// If the name is a tag name, then a new ObjectTreeNode is returned.
// Otherwise, a new TagSegmentNode is returned.
// It returns ENOENT if the name is not found.
func (node *TagSegmentNode) Lookup(ctx context.Context, name string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	revision := revisionTagName(filepath.Join(node.tagPrefix, name))
	tags, err := node.repository.Tags()
	if err != nil {
		return nil, syscall.ENOENT
	}

	ok := referenceiter.Has(tags, revision)
	if ok {
		tagNode, err := NewObjectTreeNodeByRevision(node.repository, revision)
		if err != nil {
			return nil, syscall.ENOENT
		}
		return node.NewInode(ctx, tagNode, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}
	ok = referenceiter.HasPrefix(tags, revision+tagNameSeparator)
	if !ok {
		return nil, syscall.ENOENT
	}
	ops := TagSegmentNode{repository: node.repository, tagPrefix: filepath.Join(node.tagPrefix, name) + tagNameSeparator}
	return node.NewInode(ctx, &ops, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
}

// Readdir returns the child nodes of this node.
// The child nodes are the segments of the tag names.
// For example, if the tag names are "release/v1.0.0" and "release/v1.1.0"
// then will return "release" directory with two children, "v1.0.0" and "v1.1.0".
func (node *TagSegmentNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
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
