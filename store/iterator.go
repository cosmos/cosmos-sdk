package store

// Iterator defines an interface for iterating over a domain of key/value pairs.
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

	// Close releases associated resources. It should NOT be idempotent. It must
	// only be called once and any call after may panic.
	Close()
}

// IteratorCreator defines an interface for creating forward and reverse iterators.
type IteratorCreator interface {
	// Iterator creates a new iterator for the given store name and domain, where
	// domain is defined by [start, end). Note, both start and end are optional.
	Iterator(storeKey string, start, end []byte) (Iterator, error)
	// ReverseIterator creates a new reverse iterator for the given store name
	// and domain, where domain is defined by [start, end). Note, both start and
	// end are optional.
	ReverseIterator(storeKey string, start, end []byte) (Iterator, error)
}
