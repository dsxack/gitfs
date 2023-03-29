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
	revision := revisionTagName(name)
	tags, err := node.repository.Tags()
	if err != nil {
		return nil, syscall.ENOENT
	}
	ok := referenceiter.Has(tags, revision)
	if ok {
		branchNode, err := NewObjectTreeNodeByRevision(node.repository, revision)
		if err != nil {
			return nil, syscall.ENOENT
		}
		return node.NewInode(ctx, branchNode, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}
	ok = referenceiter.HasPrefix(tags, revision+tagNameSeparator)
	if !ok {
		return nil, syscall.ENOENT
	}
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
