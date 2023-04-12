package v5

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, DelegationKey)

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		del, val, err := types.ParseDelegationKey(key)
		if err != nil {
			return err
		}

		store.Set(types.GetDelegationsByValKey(val, del), []byte{})
	}

	return nil
}
