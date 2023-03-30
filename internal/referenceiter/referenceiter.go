package referenceiter

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"strings"
)

func Has(iter storer.ReferenceIter, name string) (bool, bool) {
	has := false
	hasPrefix := false
	_ = iter.ForEach(func(reference *plumbing.Reference) error {
		refName := reference.Name().String()
		if refName == name {
			has = true
		}
		if strings.HasPrefix(refName, name) {
			hasPrefix = true
		}
		return nil
	})
	return has, hasPrefix
}
