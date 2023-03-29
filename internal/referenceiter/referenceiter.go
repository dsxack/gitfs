package referenceiter

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"strings"
)

func Has(iter storer.ReferenceIter, name string) bool {
	has := false
	_ = iter.ForEach(func(reference *plumbing.Reference) error {
		refName := reference.Name().String()
		if refName == name {
			has = true
		}
		return nil
	})
	return has
}

func HasPrefix(iter storer.ReferenceIter, prefix string) bool {
	hasPrefix := false
	_ = iter.ForEach(func(branchRef *plumbing.Reference) error {
		branchName := branchRef.Name().String()
		if strings.HasPrefix(branchName, prefix) {
			hasPrefix = true
		}
		return nil
	})
	return hasPrefix
}
