package store

import (
	"cosmossdk.io/core/store"
)

type Hash = []byte

var _ store.KVStore = (WritableState)(nil)

// Store defines the underlying storage engine of an app.
type Store interface {
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, ReadonlyState, error)

	// StateAt returns a readonly view over the provided
	// state. Must error when the version does not exist.
	StateAt(version uint64) (ReadonlyState, error)

	// StateCommit commits the provided changeset and returns
	// the new state root of the state.
	StateCommit(changes []ChangeSet) (Hash, error)
}

// ChangeSet represents a change in a key and value of state.
// Remove being true signals the key must be removed from state.
type ChangeSet struct {
	// Key defines the key being updated.
	Key []byte
	// Value defines the value associated with the updated key.
	Value []byte
	// Remove is true when the key must be removed from state.
	Remove bool
}

// WritableState defines some instance of state at a specific version that can be written to.
type WritableState interface {
	ReadonlyState
	Set(key, value []byte) error
	Delete(key []byte) error
	ApplyChangeSets(changes []ChangeSet) error
	ChangeSets() ([]ChangeSet, error)
}

// ReadonlyState defines a sub-set of the methods exposed by store.KVStore.
// The methods defined work only at read level.
type ReadonlyState interface {
	Has(key []byte) (bool, error)
	Get([]byte) ([]byte, error)
	Iterator(start, end []byte) (store.Iterator, error)        // consider removing iterate?
	ReverseIterator(start, end []byte) (store.Iterator, error) // consider removing reverse iterate
}
