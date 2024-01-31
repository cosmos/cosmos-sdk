package runtime

import (
	"cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/stf"
	storetypes "cosmossdk.io/store/types"
)

// NewKVStoreService creates a new KVStoreService.
// This wrapper is kept for backwards compatibility.
func NewKVStoreService(storeKey *storetypes.KVStoreKey) store.KVStoreService {
	return stf.NewKVStoreService([]byte(storeKey.Name()))
}
