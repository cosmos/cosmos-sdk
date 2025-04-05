package v4

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrate migrates state to consensus version 4. Specifically, the migration
// deletes the existing validator bitmap entries, lazily, with a provided limit.
// They are replaced with a new key, storing real "chunked" bitmap.
func Migrate(ctx sdk.Context, store storetypes.KVStore, index []byte, limit int32) error {
	prefixStore := prefix.NewStore(store, ValidatorSigningInfoKeyPrefix)

	iter := prefixStore.Iterator(index, nil)
	defer iter.Close()

	iterCounter := 0
	for ; iter.Valid(); iter.Next() {
		if iterCounter >= int(limit) {
			ctx.Logger().Info(fmt.Sprintf("removed %d validator(s) missed blocks bit arrays, next validator key %x", limit, iter.Key()))
			store.Set(NextMigrateValidatorMissedBlocksKey, iter.Key())
			break
		}

		// For each missed blocks entry, of which there should only be one per validator,
		// we clear all the old entries.
		address := ValidatorSigningInfoAddress(iter.Key())
		deleteValidatorMissedBlockBitArray(store, address)

		iterCounter++
	}

	if !iter.Valid() {
		ctx.Logger().Info(fmt.Sprintf("successfully completed migration for validator missed blocks key removal"))
		store.Delete(NextMigrateValidatorMissedBlocksKey)
	}

	return nil
}

func deleteValidatorMissedBlockBitArray(store storetypes.KVStore, addr sdk.ConsAddress) {
	iter := storetypes.KVStorePrefixIterator(store, validatorMissedBlockBitArrayPrefixKey(addr))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}
