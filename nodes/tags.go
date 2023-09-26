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
	_ fs.InodeEmbedder = (*TagsNode)(nil)
	_ fs.NodeReaddirer = (*TagsNode)(nil)
	_ fs.NodeLookuper  = (*TagsNode)(nil)
)

const tagNameSeparator = string(filepath.Separator)

// TagsNode is a node that represents a git repository's tags.
// It is a directory that contains a directory for each tag.
// If tag contains directory separator, it will be split into segments and each segment will be a nested directory.
// For example, if tag name is "foo/bar", it will be represented as "foo" directory with "bar" directory inside.
type TagsNode struct {
	fs.Inode
	repository *git.Repository
}

// NewTagsNode creates a new TagsNode.
func NewTagsNode(repository *git.Repository) *TagsNode {
	return &TagsNode{repository: repository}
}

// Lookup returns a tag commit tree node or a tag segment node.
// If tag name is "foo", it will return a tag commit tree node.
// If tag name is "foo/bar", it will return a tag segment node with name "bar".
// It returns ENOENT if the name is not found.
func (node *TagsNode) Lookup(ctx context.Context, name string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	logger := slog.Default().With(slog.String("lookupTagName", name))
	revision := revisionTagName(name)
	tags, err := node.repository.Tags()
	if err != nil {
		logger.Error("Error lookup tag", slog.String("error", err.Error()))
		return nil, syscall.ENOENT
	}
	ok, hasPrefix := iter.HasReference(tags, revision)
	if ok {
		branchNode, err := NewObjectTreeNodeByRevision(node.repository, revision)
		if err != nil {
			logger.Error("Error lookup tag object tree", slog.String("error", err.Error()))
			return nil, syscall.ENOENT
		}
		logger.Info("Tag object tree found")
		return node.NewInode(ctx, branchNode, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}
	if !hasPrefix {
		logger.Warn("Tag not found")
		return nil, syscall.ENOENT
	}
	logger.Info("Tag segment found")

	return node.NewInode(
		ctx,
		NewTagSegmentNode(node.repository, name+tagNameSeparator),
		fs.StableAttr{Mode: syscall.S_IFDIR},
	), 0
}

// Readdir returns a list of tag names.
// If tag name is "foo/bar", it will return "foo" directory with "bar" directory inside.
func (node *TagsNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	tagRefs, err := node.repository.Tags()
	if err != nil {
		return nil, syscall.ENOENT
	}
	slog.Default().Info("Dir of repository tags has been read")
	return iter.NewDirStreamAdapter[*plumbing.Reference](
		tagRefs,
		func(tagRef *plumbing.Reference) fuse.DirEntry {
			tagName := bareTagName(tagRef.Name().String())
			segments := strings.Split(tagName, tagNameSeparator)
			return fuse.DirEntry{Name: segments[0], Mode: syscall.S_IFDIR}
		},
	), 0
}

const revisionTagPrefix = "refs/tags/"

func bareTagName(revision string) string {
	return strings.TrimPrefix(revision, revisionTagPrefix)
}

func revisionTagName(tag string) string {
	return revisionTagPrefix + tag
}
