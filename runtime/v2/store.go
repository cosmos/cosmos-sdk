package runtime

import (
	"cosmossdk.io/core/store"
	corestore "cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/stf"
	storetypes "cosmossdk.io/store/types"
	storev2 "cosmossdk.io/store/v2"
)

// NewKVStoreService creates a new KVStoreService.
// This wrapper is kept for backwards compatibility.
func NewKVStoreService(storeKey *storetypes.KVStoreKey) store.KVStoreService {
	return stf.NewKVStoreService([]byte(storeKey.Name()))
}

type Store interface {
	// LatestVersion returns the latest version that consensus has been made on
	LatestVersion() (uint64, error)
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, corestore.ReaderMap, error)

	// StateAt returns a readonly view over the provided
	// state. Must error when the version does not exist.
	StateAt(version uint64) (corestore.ReaderMap, error)

	// StateCommit commits the provided changeset and returns
	// the new state root of the state.
	StateCommit(changes []corestore.StateChanges) (corestore.Hash, error)

	// Query is a key/value query directly to the underlying database. This skips the appmanager
	Query(storeKey string, version uint64, key []byte, prove bool) (storev2.QueryResult, error)

	// GetStateStorage returns the SS backend.
	GetStateStorage() storev2.VersionedDatabase

	// GetStateCommitment returns the SC backend.
	GetStateCommitment() storev2.Committer

	// LoadVersion loads the RootStore to the given version.
	LoadVersion(version uint64) error

	// LoadLatestVersion behaves identically to LoadVersion except it loads the
	// latest version implicitly.
	LoadLatestVersion() error
}
