package types

import (
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/store/v2/proof"
)

type Store interface {
	// LatestVersion returns the latest version that consensus has been made on
	LatestVersion() (uint64, error)
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, store.ReaderMap, error)

	// StateCommit commits the provided changeset and returns
	// the new state root of the state.
	StateCommit(changes []store.StateChanges) (store.Hash, error)

	// Query is a key/value query directly to the underlying database. This skips the appmanager
	Query(storeKey string, version uint64, key []byte, prove bool) (QueryResult, error)

	// LastCommitID returns a CommitID pertaining to the last commitment.
	LastCommitID() (proof.CommitID, error)
}

	// GetStateCommitment returns the SC backend.
	GetStateCommitment() storev2.Committer
}
