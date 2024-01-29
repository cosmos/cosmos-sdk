package store

import (
	"cosmossdk.io/core/store"
)

type Hash = []byte

var _ store.KVStore = (Writer)(nil)

// Store defines the underlying storage engine of an app.
type Store interface {
	LatestVerseion() (uint64, error)
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, ReaderMap, error)

	// StateAt returns a readonly view over the provided
	// state. Must error when the version does not exist.
	StateAt(version uint64) (ReaderMap, error)

	// StateCommit commits the provided changeset and returns
	// the new state root of the state.
	StateCommit(changes []StateChanges) (Hash, error)
}

// ReaderMap represents a readonly view over all the accounts state.
type ReaderMap interface {
	// ReaderMap must return the state for the provided actor.
	// Storage implements might treat this as a prefix store over an actor.
	// Prefix safety is on the implementer.
	GetReader(actor []byte) (Reader, error)
}

// WriterMap represents a writable actor state.
type WriterMap interface {
	ReaderMap
	// WriterMap must the return a WritableState
	// for the provided actor namespace.
	GetWriter(actor []byte) (Writer, error)
	// ApplyStateChanges applies all the state changes
	// of the accounts. Ordering of the returned state changes
	// is an implementation detail and must not be assumed.
	ApplyStateChanges(stateChanges []StateChanges) error
	// GetStateChanges returns the list of the state
	// changes so far applied. Order must not be assumed.
	GetStateChanges() ([]StateChanges, error)
}

type StateChanges struct {
	Actor        []byte // actor represents the space in storage where state is stored, previously this was called a "storekey"
	StateChanges []KVPair
}

// KVPair represents a change in a key and value of state.
// Remove being true signals the key must be removed from state.
type KVPair struct {
	// Key defines the key being updated.
	Key []byte
	// Value defines the value associated with the updated key.
	Value []byte
	// Remove is true when the key must be removed from state.
	Remove bool
}

// Writer defines an instance of an actor state at a specific version that can be written to.
type Writer interface {
	Reader
	Set(key, value []byte) error
	Delete(key []byte) error
	ApplyChangeSets(changes []KVPair) error
	ChangeSets() ([]KVPair, error)
}

// Reader defines a sub-set of the methods exposed by store.KVStore.
// The methods defined work only at read level.
type Reader interface {
	Has(key []byte) (bool, error)
	Get([]byte) ([]byte, error)
	Iterator(start, end []byte) (store.Iterator, error)        // consider removing iterate?
	ReverseIterator(start, end []byte) (store.Iterator, error) // consider removing reverse iterate
}
