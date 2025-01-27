package v6

import (
	"context"

	"cosmossdk.io/core/codec"
	storetypes "cosmossdk.io/store/types"
)

// MigrateStore performs in-place store migrations from v5 to v6.
// It deletes the ValidatorUpdatesKey from the store.
func MigrateStore(ctx context.Context, store storetypes.KVStore, cdc codec.BinaryCodec) error {
	store.Delete(ValidatorUpdatesKey)
	store.Delete(HistoricalInfoKey)
	store.Delete(UnbondingIDKey)
	store.Delete(UnbondingIndexKey)
	store.Delete(UnbondingTypeKey)
	return nil
}
