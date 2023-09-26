package iter

import (
	"github.com/hanwen/go-fuse/v2/fuse"
	"syscall"
)

type DirStreamAdapter[T any] struct {
	iter          Iter[T]
	next          T
	err           error
	entryCallback func(T) fuse.DirEntry
}

func NewDirStreamAdapter[T any](iter Iter[T], entryCallback func(T) fuse.DirEntry) *DirStreamAdapter[T] {
	return &DirStreamAdapter[T]{
		iter:          iter,
		entryCallback: entryCallback,
	}
}

func (adapter *DirStreamAdapter[T]) HasNext() bool {
	var err error
	adapter.next, err = adapter.iter.Next()
	adapter.err = err
	return err == nil
}

func (adapter *DirStreamAdapter[T]) Next() (fuse.DirEntry, syscall.Errno) {
	if adapter.err != nil {
		return fuse.DirEntry{}, syscall.ENOENT
	}
	return adapter.entryCallback(adapter.next), 0
}

func (adapter *DirStreamAdapter[T]) Close() { adapter.iter.Close() }
