package v5

import (
	"fmt"
	"strconv"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func migrateDelegationsByValidatorIndex(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec) error {
	iterator := storetypes.KVStorePrefixIterator(store, DelegationKey)
	iterationLimit := 1000
	iterationCounter := 0

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		del, val, err := ParseDelegationKey(key)
		if err != nil {
			return err
		}

		store.Set(GetDelegationsByValKey(val, del), []byte{})

		iterationCounter++
		if iterationCounter >= iterationLimit {
			ctx.Logger().Info(fmt.Sprintf("Migrated 1000 delegations, next key %x", key[1:]))
			// we set the store to the key sans the DelgationKey as we create a prefix store to iterate per block
			store.Set(NextMigrateDelegationsByValidatorIndexKey, key[1:])
			break
		}
	}

	return nil
}

// MigrateStore performs in-place store migrations from v4 to v5.
func MigrateStore(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec) error {
	if err := migrateDelegationsByValidatorIndex(ctx, store, cdc); err != nil {
		return err
	}
	return MigrateHistoricalInfoKeys(ctx, store, nil, 1000)
}

// MigrateHistoricalInfoKeys migrate HistoricalInfo keys to binary format
// old key is of format:
// prefix (0x50) || heightBytes (string representation of height in 10 base)
// new key is of format:
// prefix (0x50) || heightBytes (byte array representation using big-endian byte order)
//
// The migration function accepts a starting index and a limit. A nil index will start iterator from the beginning
// of the historical info keyspace.
func MigrateHistoricalInfoKeys(ctx sdk.Context, store storetypes.KVStore, index []byte, limit uint32) error {
	prefixStore := prefix.NewStore(store, HistoricalInfoKey)

	iter := prefixStore.Iterator(index, nil)
	defer sdk.LogDeferred(ctx.Logger(), func() error { return iter.Close() })

	iterationCounter := 0
	for ; iter.Valid(); iter.Next() {
		if iterationCounter >= int(limit) {
			ctx.Logger().Info(fmt.Sprintf("migrated %d historical info entries, next key %s", limit, iter.Key()))
			store.Set(NextMigrateHistoricalInfoKey, iter.Key())
			break
		}

		strHeight := iter.Key()

		intHeight, err := strconv.ParseInt(string(strHeight), 10, 64)
		if err != nil {
			return fmt.Errorf("can't parse height from key %q to int64: %v", strHeight, err)
		}

		newStoreKey := GetHistoricalInfoKey(intHeight)

		// Set new key on store. Values don't change.
		store.Set(newStoreKey, iter.Value())
		prefixStore.Delete(iter.Key())

		iterationCounter++
	}

	if !iter.Valid() {
		ctx.Logger().Info(fmt.Sprintf("successfully completed migration for historical info to binary key format"))
		store.Delete(NextMigrateHistoricalInfoKey)
	}

	return nil
}
