package db

import (
	"errors"

	corestore "cosmossdk.io/core/store"
)

var (
	// errBatchClosed is returned when a closed or written batch is used.
	errBatchClosed = errors.New("batch has been written or closed")

	// errKeyEmpty is returned when attempting to use an empty or nil key.
	errKeyEmpty = errors.New("key cannot be empty")

	// errValueNil is returned when attempting to set a nil value.
	errValueNil = errors.New("value cannot be nil")
)

// RawDB is the main interface for all key-value database backends. DBs are concurrency-safe.
// Callers must call Close on the database when done.
//
// Keys cannot be nil or empty, while values cannot be nil. Keys and values should be considered
// read-only, both when returned and when given, and must be copied before they are modified.
type RawDB interface {
	// Get fetches the value of the given key, or nil if it does not exist.
	// CONTRACT: key, value readonly []byte
	Get([]byte) ([]byte, error)

	// Has checks if a key exists.
	// CONTRACT: key, value readonly []byte
	Has(key []byte) (bool, error)

	// Iterator returns an iterator over a domain of keys, in ascending order. The caller must call
	// Close when done. End is exclusive, and start must be less than end. A nil start iterates
	// from the first key, and a nil end iterates to the last key (inclusive). Empty keys are not
	// valid.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	Iterator(start, end []byte) (corestore.Iterator, error)

	// ReverseIterator returns an iterator over a domain of keys, in descending order. The caller
	// must call Close when done. End is exclusive, and start must be less than end. A nil end
	// iterates from the last key (inclusive), and a nil start iterates to the first key (inclusive).
	// Empty keys are not valid.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	ReverseIterator(start, end []byte) (corestore.Iterator, error)

	// Close closes the database connection.
	Close() error

	// NewBatch creates a batch for atomic updates. The caller must call Batch.Close.
	NewBatch() RawBatch

	// NewBatchWithSize create a new batch for atomic updates, but with pre-allocated size.
	// This will does the same thing as NewBatch if the batch implementation doesn't support pre-allocation.
	NewBatchWithSize(int) RawBatch
}

// Batch represents a group of writes. They may or may not be written atomically depending on the
// backend. Callers must call Close on the batch when done.
//
// As with RawDB, given keys and values should be considered read-only, and must not be modified after
// passing them to the batch.
type RawBatch interface {
	// Set sets a key/value pair.
	// CONTRACT: key, value readonly []byte
	Set(key, value []byte) error

	// Delete deletes a key/value pair.
	// CONTRACT: key readonly []byte
	Delete(key []byte) error

	// Write writes the batch, possibly without flushing to disk. Only Close() can be called after,
	// other methods will error.
	Write() error

	// WriteSync writes the batch and flushes it to disk. Only Close() can be called after, other
	// methods will error.
	WriteSync() error

	// Close closes the batch. It is idempotent, but calls to other methods afterwards will error.
	Close() error

	// GetByteSize that returns the current size of the batch in bytes. Depending on the implementation,
	// this may return the size of the underlying LSM batch, including the size of additional metadata
	// on top of the expected key and value total byte count.
	GetByteSize() (int, error)
}

type (
	Options interface {
		Get(string) interface{}
	}
)
