package store

import (
	"cosmossdk.io/core/store"
)

type Hash = []byte

var _ store.KVStore = (Writer)(nil)

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
