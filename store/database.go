package store

import (
	"io"

	ics23 "github.com/cosmos/ics23/go"
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
	IteratorCreator
	io.Closer
}

// VersionedDatabase defines an API for a versioned database that allows reads,
// writes, iteration and commitment over a series of versions.
type VersionedDatabase interface {
	Has(storeKey string, version uint64, key []byte) (bool, error)
	Get(storeKey string, version uint64, key []byte) ([]byte, error)
	GetLatestVersion() (uint64, error)
	SetLatestVersion(version uint64) error

	Iterator(storeKey string, version uint64, start, end []byte) (Iterator, error)
	ReverseIterator(storeKey string, version uint64, start, end []byte) (Iterator, error)

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
	WriteBatch(cs *Changeset) error
	WorkingStoreInfos(version uint64) []StoreInfo
	GetLatestVersion() (uint64, error)
	LoadVersion(targetVersion uint64) error
	Commit() ([]StoreInfo, error)
	SetInitialVersion(version uint64) error
	GetProof(storeKey string, version uint64, key []byte) (*ics23.CommitmentProof, error)

	// Prune attempts to prune all versions up to and including the provided
	// version argument. The operation should be idempotent. An error should be
	// returned upon failure.
	Prune(version uint64) error

	// Close releases associated resources. It should NOT be idempotent. It must
	// only be called once and any call after may panic.
	io.Closer
}
