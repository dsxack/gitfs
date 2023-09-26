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
	logger := slog.Default().
		With(slog.String("lookupTagSegmentName", name)).
		With(slog.String("tagPrefix", node.tagPrefix))
	revision := revisionTagName(filepath.Join(node.tagPrefix, name))
	tags, err := node.repository.Tags()
	if err != nil {
		logger.Error("Error lookup tag segment", slog.String("error", err.Error()))
		return nil, syscall.ENOENT
	}

	ok, hasPrefix := iter.HasReference(tags, revision)
	if ok {
		tagNode, err := NewObjectTreeNodeByRevision(node.repository, revision)
		if err != nil {
			logger.Error("Error lookup tag object tree", slog.String("error", err.Error()))
			return nil, syscall.ENOENT
		}
		logger.Info("Tag object tree found")
		return node.NewInode(ctx, tagNode, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}
	if !hasPrefix {
		logger.Warn("Tag segment not found")
		return nil, syscall.ENOENT
	}
	logger.Info("Tag segment found")

	return node.NewInode(
		ctx,
		NewTagSegmentNode(node.repository, filepath.Join(node.tagPrefix, name)+tagNameSeparator),
		fs.StableAttr{Mode: syscall.S_IFDIR},
	), 0
}

// Readdir returns the child nodes of this node.
// The child nodes are the segments of the tag names.
// For example, if the tag names are "release/v1.0.0" and "release/v1.1.0"
// then will return "release" directory with two children, "v1.0.0" and "v1.1.0".
func (node *TagSegmentNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	var tags iter.Iter[*plumbing.Reference]
	var err error
	tags, err = node.repository.Tags()
	if err != nil {
		return nil, syscall.ENOENT
	}
	tags = iter.NewFilterIter(tags, func(tagRef *plumbing.Reference) bool {
		tagName := bareTagName(tagRef.Name().String())
		return strings.HasPrefix(tagName, node.tagPrefix)
	})
	slog.Default().Info("Dir of repository tag segment has been read", slog.String("tagPrefix", node.tagPrefix))
	return iter.NewDirStreamAdapter[*plumbing.Reference](
		tags,
		func(tagRef *plumbing.Reference) fuse.DirEntry {
			tagName := bareTagName(tagRef.Name().String())
			segments := strings.Split(strings.TrimPrefix(tagName, node.tagPrefix), tagNameSeparator)
			return fuse.DirEntry{Name: segments[0], Mode: syscall.S_IFDIR}
		},
	), 0
}
