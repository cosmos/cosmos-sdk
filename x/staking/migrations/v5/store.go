package v5

import (
	"context"
	"fmt"
	"strconv"

	"cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
)

func migrateDelegationsByValidatorIndex(store corestore.KVStore) error {
	itStore := runtime.KVStoreAdapter(store)
	iterator := storetypes.KVStorePrefixIterator(itStore, DelegationKey)

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		del, val, err := ParseDelegationKey(key)
		if err != nil {
			return err
		}

		if err := store.Set(GetDelegationsByValKey(val, del), []byte{}); err != nil {
			return err
		}
	}

	return nil
}

// MigrateStore performs in-place store migrations from v4 to v5.
func MigrateStore(ctx context.Context, store corestore.KVStore, cdc codec.BinaryCodec, logger log.Logger) error {
	if err := migrateDelegationsByValidatorIndex(store); err != nil {
		return err
	}
	return migrateHistoricalInfoKeys(store, logger)
}

// migrateHistoricalInfoKeys migrate HistoricalInfo keys to binary format
func migrateHistoricalInfoKeys(store corestore.KVStore, logger log.Logger) error {
	// old key is of format:
	// prefix (0x50) || heightBytes (string representation of height in 10 base)
	// new key is of format:
	// prefix (0x50) || heightBytes (byte array representation using big-endian byte order)
	itStore := runtime.KVStoreAdapter(store)
	oldStore := prefix.NewStore(itStore, HistoricalInfoKey)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer logDeferred(logger, func() error { return oldStoreIter.Close() })

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		strHeight := oldStoreIter.Key()

		intHeight, err := strconv.ParseInt(string(strHeight), 10, 64)
		if err != nil {
			return fmt.Errorf("can't parse height from key %q to int64: %w", strHeight, err)
		}

		newStoreKey := GetHistoricalInfoKey(intHeight)

		// Set new key on store. Values don't change.
		if err := store.Set(newStoreKey, oldStoreIter.Value()); err != nil {
			return err
		}
		oldStore.Delete(oldStoreIter.Key())
	}

	return nil
}

// logDeferred logs an error in a deferred function call if the returned error is non-nil.
func logDeferred(logger log.Logger, f func() error) {
	if err := f(); err != nil {
		logger.Error(err.Error())
	}
}
