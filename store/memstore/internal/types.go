package internal

type TypedMemIterator[T any] interface {
	Next()
	Key() []byte
	Value() T
	Valid() bool
	Domain() (start, end []byte)
	Close() error
}
