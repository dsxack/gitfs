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
	_ fs.InodeEmbedder = (*BranchesNode)(nil)
	_ fs.NodeReaddirer = (*BranchesNode)(nil)
	_ fs.NodeLookuper  = (*BranchesNode)(nil)
)

const branchNameSeparator = string(filepath.Separator)

// BranchesNode is a filesystem node that represents a list of branches.
// It is a directory that contains a list of branches.
// Each branch is a directory.
// If branch contains directory separator, it will be split into segments and each segment will be a nested directory.
// For example, if branch name is "foo/bar", it will be represented as "foo" directory with "bar" directory inside.
type BranchesNode struct {
	fs.Inode
	repository *git.Repository
}

// NewBranchesNode creates a new BranchesNode.
func NewBranchesNode(repository *git.Repository) *BranchesNode {
	return &BranchesNode{repository: repository}
}

// Lookup returns a branch commit three node or a branch segment node.
// If branch name is "foo", it will return a branch commit tree node.
// If branch name is "foo/bar", it will return a branch segment node with name "bar".
// It returns ENOENT if the name is not found.
func (node *BranchesNode) Lookup(ctx context.Context, name string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	logger := slog.Default().With(slog.String("lookupBranchName", name))
	revision := revisionBranchName(name)
	branches, err := node.repository.Branches()
	if err != nil {
		logger.Error("Error lookup branch", slog.String("error", err.Error()))
		return nil, syscall.ENOENT
	}
	ok, hasPrefix := iter.HasReference(branches, revision)
	if ok {
		branchNode, err := NewObjectTreeNodeByRevision(node.repository, revision)
		if err != nil {
			logger.Error("Error lookup branch object tree", slog.String("error", err.Error()))
			return nil, syscall.ENOENT
		}
		logger.Info("Branch object tree found")
		return node.NewInode(ctx, branchNode, fs.StableAttr{Mode: syscall.S_IFDIR}), 0
	}
	if !hasPrefix {
		logger.Info("No branch found")
		return nil, syscall.ENOENT
	}
	logger.Info("Branch segment found")

	return node.NewInode(
		ctx,
		NewBranchSegmentNode(node.repository, name+branchNameSeparator),
		fs.StableAttr{Mode: syscall.S_IFDIR},
	), 0
}

// Readdir returns a list of branches.
// If branch contains directory separator, it will be split into segments and each segment will be a nested directory.
func (node *BranchesNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	branches, err := node.repository.Branches()
	if err != nil {
		return nil, syscall.ENOENT
	}
	slog.Default().Info("Dir of repository branches has been read")
	return iter.NewDirStreamAdapter[*plumbing.Reference](
		branches,
		func(branchRef *plumbing.Reference) fuse.DirEntry {
			name := bareBranchName(branchRef.Name().String())
			segments := strings.Split(name, branchNameSeparator)
			return fuse.DirEntry{Name: segments[0], Mode: syscall.S_IFDIR}
		},
	), 0
}

const revisionBranchPrefix = "refs/heads/"

func bareBranchName(revision string) string {
	return strings.TrimPrefix(revision, revisionBranchPrefix)
}

func revisionBranchName(branch string) string {
	return revisionBranchPrefix + branch
}
