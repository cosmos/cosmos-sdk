package kvstore

import (
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/orm/types/ormhooks"
)

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

// Iterator aliases github.com/tendermint/tm-db.Iterator.
type Iterator = dbm.Iterator

// Writer is an interface for writing to a kv-store.
type Writer interface {
	Reader

	// Set sets the value for the given key, replacing it if it already exists.
	// CONTRACT: key, value readonly []byte
	Set(key, value []byte) error

	// Delete deletes the key, or does nothing if the key does not exist.
	// CONTRACT: key readonly []byte
	Delete(key []byte) error
}

// Store combines reader and writer.
type Store interface {
	Reader
	Writer
}

// ReadBackend is the kv-store backend that the ORM uses for readonly operations.
type ReadBackend interface {
	CommitmentStoreReader() Reader
	IndexStoreReader() Reader
}

// Backend is the kv-store backend that the ORM uses for state mutations.
//
// It is primarily a wrapper around two stores - an index store
// which does not need to be back by a merkle-tree and a commitment store
// which should be backed by a merkle-tree if possible. This abstraction allows
// the ORM access the two stores as a single data layer, storing all secondary
// index data in the index layer for efficiency and only storing primary records
// in the commitment store.
//
// Backend can optionally contain hooks to listen to ORM operations directly.
type Backend interface {
	ReadBackend

	// CommitmentStore returns the merklized commitment store.
	CommitmentStore() Store

	// IndexStore returns the index store if a separate one exists,
	// otherwise it the commitment store.
	IndexStore() Store

	// ORMHooks returns a Hooks instance or nil.
	ORMHooks() ormhooks.Hooks
}
