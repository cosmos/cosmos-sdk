package store

// Iterator ...
type Iterator interface {
	// Domain returns the start (inclusive) and end (exclusive) limits of the iterator.
	Domain() ([]byte, []byte)

	// Valid returns if the iterator is currently valid.
	Valid() bool

	// Next moves the iterator to the next key/value pair.
	Next() bool

	// Error returns any accumulated error. Error() should be called after all
	// key/value pairs have been exhausted, i.e. after Next() has returned false.
	Error() error

	// Key returns the key of the current key/value pair, or nil if done.
	Key() []byte

	// Value returns the value of the current key/value pair, or nil if done.
	Value() []byte

	// Close releases associated resources. Release should always succeed and can
	// be called multiple times without causing error.
	Close()
}

// IteratorCreator ...
type IteratorCreator interface {
	// NewIterator creates a new iterator for the given store name and domain, where
	// domain is defined by [start, end). Note, both start and end are optional.
	NewIterator(storeKey string, start, end []byte) (Iterator, error)
	// NewReverseIterator creates a new reverse iterator for the given store name
	// and domain, where domain is defined by [start, end). Note, both start and
	// end are optional.
	NewReverseIterator(storeKey string, start, end []byte) (Iterator, error)
}

type VersionedIteratorCreator interface {
	NewIterator(storeKey string, version uint64, start, end []byte) (Iterator, error)

	// TODO(@bez): Consider removing this API and forcing all query handlers to use
	// a forward (normal) iterator. A reverse iterator requires additional implementation
	// burden on the SS backend and in some cases, may not even be possible to
	// implement efficiently.
	NewReverseIterator(storeKey string, version uint64, start, end []byte) (Iterator, error)
}
