package runtime

import (
	"cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/stf"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/proof"
)

// NewKVStoreService creates a new KVStoreService.
// This wrapper is kept for backwards compatibility.
// When migrating from runtime to runtime/v2, use runtimev2.NewKVStoreService(storeKey.Name()) instead of runtime.NewKVStoreService(storeKey).
func NewKVStoreService(storeKey string) store.KVStoreService {
	return stf.NewKVStoreService([]byte(storeKey))
}

type Store interface {
	// GetLatestVersion returns the latest version that consensus has been made on
	GetLatestVersion() (uint64, error)
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, store.ReaderMap, error)

	// StateAt returns a readonly view over the provided
	// version. Must error when the version does not exist.
	StateAt(version uint64) (store.ReaderMap, error)

	// SetInitialVersion sets the initial version of the store.
	SetInitialVersion(uint64) error

	// WorkingHash writes the provided changeset to the state and returns
	// the working hash of the state.
	WorkingHash(changeset *store.Changeset) (store.Hash, error)

	// Commit commits the provided changeset and returns the new state root of the state.
	Commit(changeset *store.Changeset) (store.Hash, error)

	// Query is a key/value query directly to the underlying database. This skips the appmanager
	Query(storeKey []byte, version uint64, key []byte, prove bool) (storev2.QueryResult, error)

	// GetStateStorage returns the SS backend.
	GetStateStorage() storev2.VersionedDatabase

	// GetStateCommitment returns the SC backend.
	GetStateCommitment() storev2.Committer

	// LoadVersion loads the RootStore to the given version.
	LoadVersion(version uint64) error

	// LoadLatestVersion behaves identically to LoadVersion except it loads the
	// latest version implicitly.
	LoadLatestVersion() error

	// LastCommitID returns the latest commit ID
	LastCommitID() (proof.CommitID, error)
}
