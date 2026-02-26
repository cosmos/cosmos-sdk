package corecompat

import "context"

// KVStoreService represents a unique, non-forgeable handle to a regular merkle-tree
// backed KVStore. It should be provided as a module-scoped dependency by the runtime
// module being used to build the app.
type KVStoreService = interface {
	// OpenKVStore retrieves the KVStore from the context.
	OpenKVStore(context.Context) KVStore
}

// MemoryStoreService represents a unique, non-forgeable handle to a memory-backed
// KVStore. It should be provided as a module-scoped dependency by the runtime
// module being used to build the app.
type MemoryStoreService = interface {
	// OpenMemoryStore retrieves the memory store from the context.
	OpenMemoryStore(context.Context) KVStore
}

// KVStore describes the basic interface for interacting with key-value stores.
type KVStore = interface {
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

// Iterator represents an iterator over a domain of keys. Callers must call
// Close when done. No writes can happen to a domain while there exists an
// iterator over it. Some backends may take out database locks to ensure this
// will not happen.
//
// Callers must make sure the iterator is valid before calling any methods on it,
// otherwise these methods will panic.
type Iterator = interface {
	// Domain returns the start (inclusive) and end (exclusive) limits of the iterator.
	Domain() (start, end []byte)

	// Valid returns whether the current iterator is valid. Once invalid, the Iterator remains
	// invalid forever.
	Valid() bool

	// Next moves the iterator to the next key in the database, as defined by order of iteration.
	// If Valid returns false, this method will panic.
	Next()

	// Key returns the key at the current position. Panics if the iterator is invalid.
	// Note, the key returned should be a copy and thus safe for modification.
	Key() []byte

	// Value returns the value at the current position. Panics if the iterator is
	// invalid.
	// Note, the value returned should be a copy and thus safe for modification.
	Value() []byte

	// Error returns the last error encountered by the iterator, if any.
	Error() error

	// Close closes the iterator, releasing any allocated resources.
	Close() error
}
