package store

import (
	"io"

	corestore "cosmossdk.io/core/store"
)

// Reader wraps the Has and Get method of a backing data store.
type Reader interface {
	// Has retrieves if a key is present in the key-value data store.
	//
	// Note: <key> is safe to modify and read after calling Has.
	Has(storeKey string, key []byte) (bool, error)

	// Get retrieves the given key if it's present in the key-value data store.
	//
	// Note: <key> is safe to modify and read after calling Get.
	// The returned byte slice is safe to read, but cannot be modified.
	Get(storeKey string, key []byte) ([]byte, error)
}

// Writer wraps the Set method of a backing data store.
type Writer interface {
	// Set inserts the given value into the key-value data store.
	//
	// Note: <key, value> are safe to modify and read after calling Set.
	Set(storeKey string, key, value []byte) error

	// Delete removes the key from the backing key-value data store.
	//
	// Note: <key> is safe to modify and read after calling Delete.
	Delete(storeKey string, key []byte) error
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
	Has(storeKey string, version uint64, key []byte) (bool, error)
	Get(storeKey string, version uint64, key []byte) ([]byte, error)
	GetLatestVersion() (uint64, error)
	SetLatestVersion(version uint64) error

	Iterator(storeKey string, version uint64, start, end []byte) (corestore.Iterator, error)
	ReverseIterator(storeKey string, version uint64, start, end []byte) (corestore.Iterator, error)

	ApplyChangeset(version uint64, cs *Changeset) error

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
	WriteBatch(cs *Changeset) error

	// WorkingCommitInfo returns the CommitInfo for the working tree.
	WorkingCommitInfo(version uint64) *CommitInfo

	// GetLatestVersion returns the latest version.
	GetLatestVersion() (uint64, error)

	// LoadVersion loads the tree at the given version.
	LoadVersion(targetVersion uint64) error

	// Commit commits the working tree to the database.
	Commit(version uint64) (*CommitInfo, error)

	// GetProof returns the proof of existence or non-existence for the given key.
	GetProof(storeKey string, version uint64, key []byte) ([]CommitmentOp, error)

	// Get returns the value for the given key at the given version.
	//
	// NOTE: This method only exists to support migration from IAVL v0/v1 to v2.
	// Once migration is complete, this method should be removed and/or not used.
	Get(storeKey string, version uint64, key []byte) ([]byte, error)

	// SetInitialVersion sets the initial version of the tree.
	SetInitialVersion(version uint64) error

	// GetCommitInfo returns the CommitInfo for the given version.
	GetCommitInfo(version uint64) (*CommitInfo, error)

	// Prune attempts to prune all versions up to and including the provided
	// version argument. The operation should be idempotent. An error should be
	// returned upon failure.
	Prune(version uint64) error

	// Close releases associated resources. It should NOT be idempotent. It must
	// only be called once and any call after may panic.
	io.Closer
}
