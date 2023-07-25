package store

import (
	"io"
)

// Reader wraps the Has and Get method of a backing data store.
type Reader interface {
	// Has retrieves if a key is present in the key-value data store.
	//
	// Note: <key> is safe to modify and read after calling Has.
	Has(key []byte) (bool, error)

	// Get retrieves the given key if it's present in the key-value data store.
	//
	// Note: <key> is safe to modify and read after calling Get.
	// The returned byte slice is safe to read, but cannot be modified.
	Get(key []byte) ([]byte, error)
}

// Writer wraps the Set method of a backing data store.
type Writer interface {
	// Set inserts the given value into the key-value data store.
	//
	// Note: <key, value> are safe to modify and read after calling Set.
	Set(key, value []byte) error

	// Delete removes the key from the backing key-value data store.
	//
	// Note: <key> is safe to modify and read after calling Delete.
	Delete(key []byte) error
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
	io.Closer
}
