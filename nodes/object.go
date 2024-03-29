package nodes

import (
	"context"
	"fmt"
	"github.com/dsxack/gitfs/internal/iter"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"log/slog"
	"syscall"
)

var (
	_ fs.InodeEmbedder = (*ObjectTreeNode)(nil)
	_ fs.NodeReaddirer = (*ObjectTreeNode)(nil)
	_ fs.NodeGetattrer = (*ObjectTreeNode)(nil)
	_ fs.NodeLookuper  = (*ObjectTreeNode)(nil)
)

// ObjectTreeNode is a node that represents a tree object in a git repository.
// It is used to represent the content of commit.
type ObjectTreeNode struct {
	fs.Inode
	repository *git.Repository
	revision   string
	commit     *object.Commit
	tree       *object.Tree
}

// NewObjectTreeNodeByRevision creates a new ObjectTreeNode by a revision name.
// The revision name can be a branch name, a tag name or a commit hash.
func NewObjectTreeNodeByRevision(repository *git.Repository, revision string) (*ObjectTreeNode, error) {
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

	return &ObjectTreeNode{
		commit:     commit,
		repository: repository,
		revision:   revision,
		tree:       tree,
	}, nil
}

// NewObjectTreeNode creates a new ObjectTreeNode.
func NewObjectTreeNode(
	repository *git.Repository,
	revision string,
	commit *object.Commit,
	tree *object.Tree,
) *ObjectTreeNode {
	return &ObjectTreeNode{
		repository: repository,
		revision:   revision,
		commit:     commit,
		tree:       tree,
	}
}

func (node *ObjectTreeNode) Lookup(ctx context.Context, name string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	logger := slog.Default().With(slog.String("lookupEntryName", name))
	entry, err := node.tree.FindEntry(name)
	if err != nil {
		logger.Warn("Error lookup entry of object tree", slog.String("error", err.Error()))
		return nil, syscall.ENOENT
	}
	if entry.Mode.IsFile() {
		file, err := node.tree.File(name)
		if err != nil {
			logger.Error("Error lookup entry of object tree", slog.String("error", err.Error()))
			return nil, syscall.ENOENT
		}
		logger.Info("File object found")
		return node.NewInode(ctx, NewFileNode(file, node.commit), fs.StableAttr{Mode: syscall.S_IFREG}), 0
	}

	tree, err := object.GetTree(node.repository.Storer, entry.Hash)
	if err != nil {
		logger.Error("Error lookup object tree", slog.String("error", err.Error()))
		return nil, syscall.ENOENT
	}
	logger.Info("Directory object tree found")

	return node.NewInode(
		ctx,
		NewObjectTreeNode(node.repository, node.revision, node.commit, tree),
		fs.StableAttr{Mode: syscall.S_IFDIR},
	), 0
}

func (node *ObjectTreeNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	slog.Default().Info("Dir of object tree has been read")
	return iter.NewDirStreamAdapter[object.TreeEntry](
		iter.NewSliceIter(node.tree.Entries),
		func(entry object.TreeEntry) fuse.DirEntry {
			var mode uint32 = fuse.S_IFREG
			if !entry.Mode.IsFile() {
				mode = fuse.S_IFDIR
			}
			return fuse.DirEntry{
				Name: entry.Name,
				Mode: mode,
			}
		},
	), 0
}

func (node *ObjectTreeNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	slog.Default().Debug("Got object tree attrs")
	out.Mtime = uint64(node.commit.Committer.When.Unix())
	return 0
}
