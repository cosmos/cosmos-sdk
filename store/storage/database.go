package storage

import (
	"io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
)

// Database is an interface that wraps the storage database methods. A wrapper
// is useful for instances where you want to perform logic that is identical for all SS
// backends, such as restoring snapshots.
type Database interface {
	NewBatch(version uint64) (store.Batch, error)
	Has(storeKey []byte, version uint64, key []byte) (bool, error)
	Get(storeKey []byte, version uint64, key []byte) ([]byte, error)
	GetLatestVersion() (uint64, error)
	SetLatestVersion(version uint64) error

	Iterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error)
	ReverseIterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error)

	Prune(version uint64) error

	io.Closer
}
