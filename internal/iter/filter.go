package iter

type FilterIter[T any] struct {
	iter       Iter[T]
	filterFunc func(T) bool
}

func NewFilterIter[T any](iter Iter[T], filterFunc func(T) bool) *FilterIter[T] {
	return &FilterIter[T]{iter: iter, filterFunc: filterFunc}
}

func (f FilterIter[T]) Next() (T, error) {
	for {
		item, err := f.iter.Next()
		if err != nil {
			return item, err
		}
		if f.filterFunc(item) {
			return item, nil
		}
	}
}

func (f FilterIter[T]) ForEach(callback func(T) error) error {
	for {
		item, err := f.Next()
		if err != nil {
			return err
		}
		if err := callback(item); err != nil {
			return err
		}
	}
}

func (f FilterIter[T]) Close() { f.iter.Close() }
