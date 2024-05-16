package types

import (
	"cosmossdk.io/core/store"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/proof"
)

type Store interface {
	// LatestVersion returns the latest version that consensus has been made on
	GetLatestVersion() (uint64, error)
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, store.ReaderMap, error)

	WorkingHash(*store.Changeset) ([]byte, error)

	// StateCommit commits the provided changeset and returns
	// the new state root of the state.
	Commit(*store.Changeset) (store.Hash, error)

	// Query is a key/value query directly to the underlying database. This skips the appmanager
	Query(storeKey []byte, version uint64, key []byte, prove bool) (storev2.QueryResult, error)

	// LastCommitID returns a CommitID pertaining to the last commitment.
	LastCommitID() (proof.CommitID, error)

	// GetStateStorage returns the SS backend.
	GetStateStorage() storev2.VersionedDatabase

	// GetStateCommitment returns the SC backend.
	GetStateCommitment() storev2.Committer
}
