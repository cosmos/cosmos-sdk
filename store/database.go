package store

import (
	"io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2/proof"
)

// Reader wraps the Has and Get method of a backing data store.
type Reader interface {
	// Has retrieves if a key is present in the key-value data store.
	//
	// Note: <key> is safe to modify and read after calling Has.
	Has(storeKey, key []byte) (bool, error)

	// Get retrieves the given key if it's present in the key-value data store.
	//
	// Note: <key> is safe to modify and read after calling Get.
	// The returned byte slice is safe to read, but cannot be modified.
	Get(storeKey, key []byte) ([]byte, error)
}

// Writer wraps the Set method of a backing data store.
type Writer interface {
	// Set inserts the given value into the key-value data store.
	//
	// Note: <key, value> are safe to modify and read after calling Set.
	Set(storeKey, key, value []byte) error

	// Delete removes the key from the backing key-value data store.
	//
	// Note: <key> is safe to modify and read after calling Delete.
	Delete(storeKey, key []byte) error
}

// Database contains all the methods required to allow handling different
// key-value data stores backing the database.
type Database interface {
	Reader
	Writer
	corestore.IteratorCreator
	io.Closer
}

// VersionedDatabase defines an API for a versioned database that allows reads,
// writes, iteration and commitment over a series of versions.
type VersionedDatabase interface {
	Has(storeKey []byte, version uint64, key []byte) (bool, error)
	Get(storeKey []byte, version uint64, key []byte) ([]byte, error)
	GetLatestVersion() (uint64, error)
	SetLatestVersion(version uint64) error

	Iterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error)
	ReverseIterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error)

	ApplyChangeset(version uint64, cs *corestore.Changeset) error

	// Prune attempts to prune all versions up to and including the provided
	// version argument. The operation should be idempotent. An error should be
	// returned upon failure.
	Prune(version uint64) error

	// Close releases associated resources. It should NOT be idempotent. It must
	// only be called once and any call after may panic.
	io.Closer
}

// Committer defines an API for committing state.
type Committer interface {
	// WriteBatch writes a batch of key-value pairs to the tree.
	WriteBatch(cs *corestore.Changeset) error

	// WorkingCommitInfo returns the CommitInfo for the working tree.
	WorkingCommitInfo(version uint64) *proof.CommitInfo

	// GetLatestVersion returns the latest version.
	GetLatestVersion() (uint64, error)

	// LoadVersion loads the tree at the given version.
	LoadVersion(targetVersion uint64) error

	// Commit commits the working tree to the database.
	Commit(version uint64) (*proof.CommitInfo, error)

	// GetProof returns the proof of existence or non-existence for the given key.
	GetProof(storeKey []byte, version uint64, key []byte) ([]proof.CommitmentOp, error)

	// Get returns the value for the given key at the given version.
	//
	// NOTE: This method only exists to support migration from IAVL v0/v1 to v2.
	// Once migration is complete, this method should be removed and/or not used.
	Get(storeKey []byte, version uint64, key []byte) ([]byte, error)

	// SetInitialVersion sets the initial version of the tree.
	SetInitialVersion(version uint64) error

	// GetCommitInfo returns the CommitInfo for the given version.
	GetCommitInfo(version uint64) (*proof.CommitInfo, error)

	// Prune attempts to prune all versions up to and including the provided
	// version argument. The operation should be idempotent. An error should be
	// returned upon failure.
	Prune(version uint64) error

	// Close releases associated resources. It should NOT be idempotent. It must
	// only be called once and any call after may panic.
	io.Closer
}

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
