package store

import (
	"cosmossdk.io/core/store"
)

type Hash = []byte

var _ store.KVStore = (WritableState)(nil)

// Store defines the underlying storage engine of an app.
type Store interface {
	LatestVerseion() (uint64, error)
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, ReadonlyAccountsState, error)

	// StateAt returns a readonly view over the provided
	// state. Must error when the version does not exist.
	StateAt(version uint64) (ReadonlyAccountsState, error)

	// StateCommit commits the provided changeset and returns
	// the new state root of the state.
	StateCommit(changes []AccountStateChanges) (Hash, error)
}

// ReadonlyAccountsState represents a readonly view over all the accounts state.
type ReadonlyAccountsState interface {
	// GetAccountReadonlyState must return the state for the provided address.
	// Storage implements might treat this as a prefix store over an address.
	// Prefix safety is on the implementer.
	GetAccountReadonlyState(address []byte) (ReadonlyState, error)
}

// WritableAccountsState represents a writable account state.
type WritableAccountsState interface {
	ReadonlyAccountsState
	// GetAccountWritableState must the return a WritableState
	// for the provided account address.
	GetAccountWritableState(address []byte) (WritableState, error)
	// ApplyAccountsStateChanges applies all the state changes
	// of the accounts. Ordering of the returned state changes
	// is an implementation detail and must not be assumed.
	ApplyAccountsStateChanges(stateChanges []AccountStateChanges) error
	// GetAccountsStateChanges returns the list of the state
	// changes so far applied. Order must not be assumed.
	GetAccountsStateChanges() ([]AccountStateChanges, error)
}

type AccountStateChanges struct {
	Account      []byte // Staking/Bank/x/Account
	StateChanges []StateChange
}

// StateChange represents a change in a key and value of state.
// Remove being true signals the key must be removed from state.
type StateChange struct {
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
	ApplyChangeSets(changes []StateChange) error
	ChangeSets() ([]StateChange, error)
}

// ReadonlyState defines a sub-set of the methods exposed by store.KVStore.
// The methods defined work only at read level.
type ReadonlyState interface {
	Has(key []byte) (bool, error)
	Get([]byte) ([]byte, error)
	Iterator(start, end []byte) (store.Iterator, error)        // consider removing iterate?
	ReverseIterator(start, end []byte) (store.Iterator, error) // consider removing reverse iterate
}
