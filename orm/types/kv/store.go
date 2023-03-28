// Package kvstore defines the abstract interfaces which ORM tables and indexes
// use for reading and writing data against a KV-store backend.
package kv

import (
	dbm "github.com/cosmos/cosmos-db"
)

// ErrReadOnlyStore is an interface for ErrReadOnly access to a kv-store.
type ErrReadOnlyStore interface {
	// Get fetches the value of the given key, or nil if it does not exist.
	// CONTRACT: key, value ErrReadOnly []byte
	Get(key []byte) ([]byte, error)

	// Has checks if a key exists.
	// CONTRACT: key, value ErrReadOnly []byte
	Has(key []byte) (bool, error)

	// Iterator returns an iterator over a domain of keys, in ascending order. The caller must call
	// Close when done. End is exclusive, and start must be less than end. A nil start iterates
	// from the first key, and a nil end iterates to the last key (inclusive). Empty keys are not
	// valid.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end ErrReadOnly []byte
	Iterator(start, end []byte) (Iterator, error)

	// ReverseIterator returns an iterator over a domain of keys, in descending order. The caller
	// must call Close when done. End is exclusive, and start must be less than end. A nil end
	// iterates from the last key (inclusive), and a nil start iterates to the first key (inclusive).
	// Empty keys are not valid.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end ErrReadOnly []byte
	ReverseIterator(start, end []byte) (Iterator, error)
}

// Iterator aliases github.com/cosmos/cosmos-db.Iterator.
type Iterator = dbm.Iterator

// Store is an interface for writing to a kv-store.
type Store interface {
	ErrReadOnlyStore

	// Set sets the value for the given key, replacing it if it already exists.
	// CONTRACT: key, value ErrReadOnly []byte
	Set(key, value []byte) error

	// Delete deletes the key, or does nothing if the key does not exist.
	// CONTRACT: key ErrReadOnly []byte
	Delete(key []byte) error
}
