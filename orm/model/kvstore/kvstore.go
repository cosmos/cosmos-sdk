package kvstore

import dbm "github.com/tendermint/tm-db"

// Reader is an interface for readonly access to a kv-store.
type Reader interface {
	// Get fetches the value of the given key, or nil if it does not exist.
	// CONTRACT: key, value readonly []byte
	Get(key []byte) ([]byte, error)

	// Has checks if a key exists.
	// CONTRACT: key, value readonly []byte
	Has(key []byte) (bool, error)

	// Iterator returns an iterator over a domain of keys, in ascending order. The caller must call
	// Close when done. End is exclusive, and start must be less than end. A nil start iterates
	// from the first key, and a nil end iterates to the last key (inclusive). Empty keys are not
	// valid.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	Iterator(start, end []byte) (Iterator, error)

	// ReverseIterator returns an iterator over a domain of keys, in descending order. The caller
	// must call Close when done. End is exclusive, and start must be less than end. A nil end
	// iterates from the last key (inclusive), and a nil start iterates to the first key (inclusive).
	// Empty keys are not valid.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	ReverseIterator(start, end []byte) (Iterator, error)
}

// Writer is an interface for writing to a kv-store. It shouldn't be assumed
// the writes are automatically readable with this interface, but instead
// that writes are batched and must be committed.
type Writer interface {
	Reader

	// Set sets the value for the given key, replacing it if it already exists.
	// CONTRACT: key, value readonly []byte
	Set(key, value []byte) error

	// Delete deletes the key, or does nothing if the key does not exist.
	// CONTRACT: key readonly []byte
	Delete(key []byte) error
}

// IndexCommitmentReadStore is a read-only version of IndexCommitmentStore.
type IndexCommitmentReadStore interface {
	CommitmentStoreReader() Reader
	IndexStoreReader() Reader
}

// IndexCommitmentStore is a wrapper around two stores - an index store
// which does not need to be back by a merkle-tree and a commitment store
// which should be backed by a merkle-tree if possible. This abstraction allows
// the ORM access the two stores as a single data layer, storing all secondary
// index data in the index layer for efficiency and only storing primary records
// in the commitment store.
type IndexCommitmentStore interface {
	IndexCommitmentReadStore

	// Writer returns a new IndexCommitmentStoreWriter.
	Writer() IndexCommitmentStoreWriter
}

// IndexCommitmentStoreWriter is an interface which allows writing to the
// commitment and index stores and committing them in a unified way.
type IndexCommitmentStoreWriter interface {
	IndexCommitmentReadStore

	// CommitmentStoreWriter returns a writer for the merklized commitment store.
	CommitmentStoreWriter() Writer

	// IndexStoreWriter returns a writer for the index store if a separate one exists,
	// otherwise it returns a writer for commitment store.
	IndexStoreWriter() Writer

	// Commit flushes pending writes and discards the transaction. It should
	// be assumed that writes are not available to read until after Commit
	// has been called although this may not be true of all backends.
	Commit() error

	// Close should be called whenever the caller is done with this writer.
	// If Commit was not called beforehand, the write batch is discarded.
	Close()
}

// Iterator aliases github.com/tendermint/tm-db.Iterator.
type Iterator = dbm.Iterator
