package v3

import (
	"bytes"
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/store/v2/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

// MigrateStore performs in-place store migrations from version 2 to 3. It sets
// every fee allowance queue entry that has an empty value to the
// collections.BoolValue encoding of true.
//
// The v1 -> v2 migration and the standalone cosmossdk.io/x/feegrant module
// wrote queue entries with an empty value. The keeper reads the queue with
// collections.BoolValue, which cannot decode an empty value, so
// RemoveExpiredAllowances fails in the EndBlocker once such an entry expires.
// Entries that already have a value are left unchanged.
func MigrateStore(ctx context.Context, storeService store.KVStoreService) error {
	kvStore := storeService.OpenKVStore(ctx)
	queueStore := prefix.NewStore(runtime.KVStoreAdapter(kvStore), FeeAllowanceQueueKeyPrefix)

	keys := emptyValueKeys(queueStore)

	value, err := collections.BoolValue.Encode(true)
	if err != nil {
		return err
	}

	for _, key := range keys {
		queueStore.Set(key, value)
	}

	return nil
}

// emptyValueKeys returns the keys of the store that have an empty value. The
// keys are collected before any write so the store is not modified while it is
// being iterated.
func emptyValueKeys(store storetypes.KVStore) [][]byte {
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var keys [][]byte
	for ; iterator.Valid(); iterator.Next() {
		if len(iterator.Value()) == 0 {
			keys = append(keys, bytes.Clone(iterator.Key()))
		}
	}

	return keys
}
