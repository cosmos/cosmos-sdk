package v2

import (
	"context"

	"cosmossdk.io/core/store"
	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

func addAllowancesByExpTimeQueue(ctx context.Context, store store.KVStore, cdc codec.BinaryCodec) error {
	prefixStore := prefix.NewStore(runtime.KVStoreAdapter(store), FeeAllowanceKeyPrefix)
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
			if exp.Before(types.UnwrapSDKContext(ctx).BlockTime()) {
				prefixStore.Delete(key)
			} else {
				grantByExpTimeQueueKey := FeeAllowancePrefixQueue(exp, key)
				err = store.Set(grantByExpTimeQueueKey, []byte{})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func MigrateStore(ctx context.Context, storeService store.KVStoreService, cdc codec.BinaryCodec) error {
	kvStore := storeService.OpenKVStore(ctx)
	return addAllowancesByExpTimeQueue(ctx, kvStore, cdc)
}
