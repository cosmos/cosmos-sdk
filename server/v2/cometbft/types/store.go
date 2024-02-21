package types

import (
	ics23 "github.com/cosmos/ics23/go"

	"cosmossdk.io/server/v2/core/store"
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
}

type QueryResult interface {
	Key() []byte
	Value() []byte
	Version() uint64
	Proof() *ics23.CommitmentProof
	ProofType() string
}
