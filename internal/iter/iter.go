package iter

type Iter[T any] interface {
	Next() (T, error)
	ForEach(func(T) error) error
	Close()
}
