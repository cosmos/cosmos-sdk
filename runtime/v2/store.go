package runtime

import (
	"cosmossdk.io/core/store"
	corestore "cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/stf"
	storetypes "cosmossdk.io/store/types"
)

// NewKVStoreService creates a new KVStoreService.
// This wrapper is kept for backwards compatibility.
func NewKVStoreService(storeKey *storetypes.KVStoreKey) store.KVStoreService {
	return stf.NewKVStoreService([]byte(storeKey.Name()))
}

// Store defines the underlying storage engine of an app.
type Store interface {
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, corestore.ReaderMap, error)

	// StateAt returns a readonly view over the provided
	// state. Must error when the version does not exist.
	StateAt(version uint64) (corestore.ReaderMap, error)

	// LoadVersion loads the RootStore to the given version.
	LoadVersion(version uint64) error

	// LoadLatestVersion behaves identically to LoadVersion except it loads the
	// latest version implicitly.
	LoadLatestVersion() error
}
