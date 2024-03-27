package store

import (
	"cosmossdk.io/core/store"
)

// Store defines the underlying storage engine of an app.
type Store interface {
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, store.ReaderMap, error)

	// StateAt returns a readonly view over the provided
	// state. Must error when the version does not exist.
	StateAt(version uint64) (store.ReaderMap, error)
}

// Database represents an interface for interacting with a database.
type Database interface {
	// Has checks if a key exists in the database for a given store key and version.
	Has(storeKey []byte, version uint64, key []byte) (bool, error)

	// Get retrieves the value associated with a key from the database for a given store key and version.
	Get(storeKey []byte, version uint64, key []byte) ([]byte, error)

	// GetLatestVersion returns the latest version of the database.
	GetLatestVersion() (uint64, error)

	// SetLatestVersion sets the latest version of the database.
	SetLatestVersion(version uint64) error

	// Iterator returns an iterator for iterating over a range of key-value pairs in the database
	// for a given store key, version, start key, and end key.
	Iterator(storeKey []byte, version uint64, start, end []byte) (store.Iterator, error)

	// ReverseIterator returns a reverse iterator for iterating over a range of key-value pairs in the database
	// for a given store key, version, start key, and end key in reverse order.
	ReverseIterator(storeKey []byte, version uint64, start, end []byte) (store.Iterator, error)
}
