package store

// KVStore describes the basic interface for interacting with key-value stores.
type KVStore interface {
	// Get returns nil iff key doesn't exist. Errors on nil key.
	Get(key []byte) ([]byte, error)

	// Has checks if a key exists. Errors on nil key.
	Has(key []byte) (bool, error)

	// Set sets the key. Errors on nil key or value.
	Set(key, value []byte) error

	// Delete deletes the key. Errors on nil key.
	Delete(key []byte) error

	// Iterator iterates over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// To iterate over entire domain, use store.Iterator(nil, nil)
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// Exceptionally allowed for cachekv.Store, safe to write in the modules.
	Iterator(start, end []byte) (Iterator, error)

	// ReverseIterator iterates over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// Exceptionally allowed for cachekv.Store, safe to write in the modules.
	ReverseIterator(start, end []byte) (Iterator, error)
}

type Iterator interface {
	// Domain returns the start (inclusive) and end (exclusive) limits of the iterator.
	// CONTRACT: start, end readonly []byte
	Domain() ([]byte, []byte, error)

	// Valid returns whether the current iterator is valid. Once invalid, the Iterator remains
	// invalid forever.
	Valid() (bool, error)

	// Next moves the iterator to the next key in the database, as defined by order of iteration.
	// If Valid returns false, this method will error.
	Next() error

	// Key returns the key at the current position. Errors if the iterator is invalid.
	// CONTRACT: key readonly []byte
	Key() ([]byte, error)

	// Value returns the value at the current position. Errors if the iterator is invalid.
	// CONTRACT: value readonly []byte
	Value() ([]byte, error)

	// Close closes the iterator, relasing any allocated resources.
	Close() error
}
