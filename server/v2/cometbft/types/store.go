package types

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/proof"
)

type Store interface {
	// LatestVersion returns the latest version that consensus has been made on
	LatestVersion() (uint64, error)
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, corestore.ReaderMap, error)

	// StateCommit commits the provided changeset and returns
	// the new state root of the state.
	StateCommit(changes []corestore.StateChanges) (corestore.Hash, error)

	// Query is a key/value query directly to the underlying database. This skips the appmanager
	Query(storeKey string, version uint64, key []byte, prove bool) (store.QueryResult, error)

	// LastCommitID returns a CommitID pertaining to the last commitment.
	LastCommitID() (proof.CommitID, error)

	// GetStateStorage returns the SS backend.
	GetStateStorage() store.VersionedDatabase

	// GetStateCommitment returns the SC backend.
	GetStateCommitment() store.Committer
}
