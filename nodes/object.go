package nodes

import (
	"context"
	"fmt"
	"github.com/dsxack/gitfs/internal/set"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"syscall"
)

var (
	_ fs.InodeEmbedder = (*objectTreeNode)(nil)
	_ fs.NodeReaddirer = (*objectTreeNode)(nil)
	_ fs.NodeGetattrer = (*objectTreeNode)(nil)
	_ fs.NodeLookuper  = (*objectTreeNode)(nil)
)

type objectTreeNode struct {
	fs.Inode
	repository *git.Repository
	revision   string
	commit     *object.Commit
	tree       *object.Tree
}

func newObjectTreeNodeByRevision(repository *git.Repository, revision string) (*objectTreeNode, error) {
	h, err := repository.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		return nil, fmt.Errorf("repository: resolve revision: %v", err)
	}

	commit, err := repository.CommitObject(*h)
	if err != nil {
		return nil, fmt.Errorf("repository: commit object: %v", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("commit tree: %v", err)
	}

	return &objectTreeNode{
		commit:     commit,
		repository: repository,
		revision:   revision,
		tree:       tree,
	}, nil
}

func newObjectTreeNode(
	repository *git.Repository,
	revision string,
	commit *object.Commit,
	tree *object.Tree,
) *objectTreeNode {
	return &objectTreeNode{
		repository: repository,
		revision:   revision,
		commit:     commit,
		tree:       tree,
	}
}

func (node *objectTreeNode) Lookup(ctx context.Context, name string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	entry, err := node.tree.FindEntry(name)
	if err != nil {
		return nil, syscall.ENOENT
	}
	if entry.Mode.IsFile() {
		file, err := node.tree.File(name)
		if err != nil {
			return nil, syscall.ENOENT
		}
		return node.NewInode(ctx, newFileNode(file, node.commit), fs.StableAttr{Mode: syscall.S_IFREG}), 0
	}

	tree, err := node.tree.Tree(name)
	if err != nil {
		return nil, syscall.ENOENT
	}

	return node.NewInode(
		ctx,
		newObjectTreeNode(node.repository, node.revision, node.commit, tree),
		fs.StableAttr{Mode: syscall.S_IFDIR},
	), 0
}

func (node *objectTreeNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	dirEntries := set.NewSet[fuse.DirEntry]()
	for _, entry := range node.tree.Entries {
		var mode uint32 = fuse.S_IFREG
		if !entry.Mode.IsFile() {
			mode = fuse.S_IFDIR
		}

		dirEntries.Add(fuse.DirEntry{
			Name: entry.Name,
			Mode: mode,
		})
	}
	return fs.NewListDirStream(dirEntries.Values()), 0
}

func (node *objectTreeNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mtime = uint64(node.commit.Committer.When.Unix())
	return 0
}
