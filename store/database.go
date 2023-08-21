package store

import (
	"io"
)

// VersionedReader extends the Reader interface by making all reads versioned.
type VersionedReader interface {
	Has(storeKey string, version uint64, key []byte) (bool, error)
	Get(storeKey string, version uint64, key []byte) ([]byte, error)
	GetLatestVersion() (uint64, error)
}

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

// VersionedWriter extends the Writer interface by making all writes versioned.
type VersionedWriter interface {
	Set(storeKey string, version uint64, key, value []byte) error
	Delete(storeKey string, version uint64, key []byte) error
	SetLatestVersion(version uint64) error
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

// VersionedReaderWriter combines the VersionedReader and VersionedWriter interfaces.
type VersionedReaderWriter interface {
	VersionedReader
	VersionedWriter
}

// ReaderWriter combines the Reader and Writer interfaces.
type ReaderWriter interface {
	Reader
	Writer
}

// Database contains all the methods required to allow handling different
// key-value data stores backing the database.
type Database interface {
	ReaderWriter
	IteratorCreator
	io.Closer
}

// VersionedDatabase extends the Database interface by making all reads and writes
// versioned.
type VersionedDatabase interface {
	VersionedReaderWriter
	VersionedIteratorCreator
	VersionedBatcher
	io.Closer
}

type Committer interface {
	Commit() error
}
