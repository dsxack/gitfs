package iter

import (
	"io"
)

type SliceIter[T any] struct {
	slice   []T
	current int
}

func NewSliceIter[T any](slice []T) *SliceIter[T] {
	return &SliceIter[T]{
		slice:   slice,
		current: 0,
	}
}

func (s *SliceIter[T]) Next() (T, error) {
	if s.current >= len(s.slice) {
		var empty T
		return empty, io.EOF
	}
	s.current++
	return s.slice[s.current-1], nil
}

func (s *SliceIter[T]) ForEach(f func(T) error) error {
	for _, v := range s.slice {
		if err := f(v); err != nil {
			return err
		}
	}
	return nil
}

func (s *SliceIter[T]) Close() {}
