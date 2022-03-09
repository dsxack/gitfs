package set

type Set[T comparable] map[T]struct{}

func NewSet[T comparable]() Set[T] {
	return Set[T]{}
}

func (set Set[T]) Add(value T) {
	set[value] = struct{}{}
}

func (set Set[T]) Values() []T {
	values := make([]T, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	return values
}
