package storage

import (
	"io"

	"cosmossdk.io/store/v2"
)

// Database is an interface that wraps the storage database methods.
type Database interface {
	NewBatch(version uint64) (store.Batch, error)
	Has(storeKey string, version uint64, key []byte) (bool, error)
	Get(storeKey string, version uint64, key []byte) ([]byte, error)
	GetLatestVersion() (uint64, error)
	SetLatestVersion(version uint64) error

	Iterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error)
	ReverseIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error)

	Prune(version uint64) error

	io.Closer
}
