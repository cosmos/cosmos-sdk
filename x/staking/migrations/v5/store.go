package v5

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) (storetypes.KVStore, error) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, DelegationKey)

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		del, val, err := types.ParseDelegationKey(key)
		if err != nil {
			return store, err
		}

		store.Set(types.GetDelegationsByValKey(val, del), []byte{})
		fmt.Println(types.GetDelegationsByValKey(val, del), val.String(), del.String())
	}

	return store, nil
}
