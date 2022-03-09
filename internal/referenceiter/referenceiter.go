package referenceiter

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
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
