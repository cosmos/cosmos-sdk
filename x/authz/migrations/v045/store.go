package v045

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	v044 "github.com/cosmos/cosmos-sdk/x/authz/migrations/v044"
)

// MigrateStore performs in-place store migrations from v0.44 to v0.45. The
// migration includes:
//
// - pruning expired authorizations
// -
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	err := addExpiredGrantsIndex(ctx, store, cdc)
	if err != nil {
		return err
	}

	return nil
}

func addExpiredGrantsIndex(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec) error {
	grantsStore := prefix.NewStore(store, v044.GrantKey)

	grantsIter := grantsStore.Iterator(nil, nil)
	defer grantsIter.Close()

	for ; grantsIter.Valid(); grantsIter.Next() {
		var grant authz.Grant
		bz := grantsIter.Value()
		if err := cdc.Unmarshal(bz, &grant); err != nil {
			return err
		}

		// delete expired authorization
		if grant.Expiration.Before(ctx.BlockTime()) {
			grantsStore.Delete(grantsIter.Key())
		} else {
			queueKey := v044.GrantQueueKey(grantsIter.Key(), grant.Expiration)
			store.Set(queueKey, grantsIter.Key())
		}

	}

	return nil
}
