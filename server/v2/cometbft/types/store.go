package types

import (
	"cosmossdk.io/core/store"
	"cosmossdk.io/store/v2/proof"
	ics23 "github.com/cosmos/ics23/go"
	storev2 "cosmossdk.io/store/v2"
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
	Query(storeKey string, version uint64, key []byte, prove bool) (storev2.QueryResult, error)

	// LastCommitID returns a CommitID pertaining to the last commitment.
	LastCommitID() (proof.CommitID, error)

	// GetStateStorage returns the SS backend.
	GetStateStorage() storev2.VersionedDatabase

	// GetStateCommitment returns the SC backend.
	GetStateCommitment() storev2.Committer
}

type QueryResult interface {
	Key() []byte
	Value() []byte
	Version() uint64
	Proof() *ics23.CommitmentProof
	ProofType() string
}
