package v2

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
)

func addAllowancesByExpTimeQueue(ctx types.Context, store storetypes.KVStore, cdc codec.BinaryCodec) error {
	prefixStore := prefix.NewStore(store, FeeAllowanceKeyPrefix)
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var grant feegrant.Grant
		bz := iterator.Value()
		if err := cdc.Unmarshal(bz, &grant); err != nil {
			return err
		}

		grantInfo, err := grant.GetGrant()
		if err != nil {
			return err
		}

		exp, err := grantInfo.ExpiresAt()
		if err != nil {
			return err
		}

		if exp != nil {
			// store key is not changed in 0.46
			key := iterator.Key()
			if exp.Before(ctx.BlockTime()) {
				prefixStore.Delete(key)
			} else {
				grantByExpTimeQueueKey := FeeAllowancePrefixQueue(exp, key)
				store.Set(grantByExpTimeQueueKey, []byte{})
			}
		}
	}

	return nil
}

func MigrateStore(ctx types.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	return addAllowancesByExpTimeQueue(ctx, store, cdc)
}
