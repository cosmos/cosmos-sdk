package store

// Iterator ...
type Iterator interface {
	// Next moves the iterator to the next key/value pair. It returns whether
	// the iterator successfully moved to a new key/value pair. The iterator may
	// return false if the underlying database has been closed before the iteration
	// has completed, in which case future calls to Error() must return ErrClosed.
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
	// NewIterator creates an iterator over the entire key space contained within
	// the backing key-value database.
	NewIterator() Iterator

	// NewStartIterator creates an iterator over a subset of a database key space
	// starting at a particular key.
	NewStartIterator(start []byte) Iterator

	// NewEndIterator creates an iterator over a subset of a database key space
	// ending at a particular key.
	NewEndIterator(start []byte) Iterator

	// NewPrefixIterator creates an iterator over a subset of a database key space
	// with a particular key prefix.
	NewPrefixIterator(prefix []byte) Iterator
}
