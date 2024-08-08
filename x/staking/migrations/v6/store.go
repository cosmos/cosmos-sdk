package v6

import (
	"context"

	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
)

// MigrateStore performs in-place store migrations from v5 to v6.
// It deletes the ValidatorUpdatesKey from the store.
func MigrateStore(_ context.Context, store storetypes.KVStore, cdc codec.BinaryCodec) error {
	if err := store.Delete(ValidatorUpdatesKey); err != nil {
		return err
	}
	if err := store.Delete(HistoricalInfoKey); err != nil {
		return err
	}
	return nil
}
