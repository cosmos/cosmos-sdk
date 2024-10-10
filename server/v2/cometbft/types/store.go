package types

import (
	"cosmossdk.io/core/store"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/proof"
)

type Store interface {
	storev2.Backend

	// GetLatestVersion returns the latest version that consensus has been made on
	GetLatestVersion() (uint64, error)
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, store.ReaderMap, error)

	// SetInitialVersion sets the initial version of the store.
	SetInitialVersion(uint64) error

	// WorkingHash writes the provided changeset to the state and returns
	// the working hash of the state.
	WorkingHash(*store.Changeset) (store.Hash, error)

	// Commit commits the provided changeset and returns
	// the new state root of the state.
	Commit(*store.Changeset) (store.Hash, error)

	// Query is a key/value query directly to the underlying database. This skips the appmanager
	Query(storeKey []byte, version uint64, key []byte, prove bool) (storev2.QueryResult, error)

	// LastCommitID returns a CommitID pertaining to the last commitment.
	LastCommitID() (proof.CommitID, error)
}
