package v045

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040staking "github.com/cosmos/cosmos-sdk/x/staking/migrations/v040"
)

// MigrateStore performs in-place store migrations from v0.43/v0.44 to v0.45.
// The migration includes:
//
// - Removing delegations that have a zero share or token amount.
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)

	return deleteDelegations(store, cdc)
}

func deleteDelegations(store sdk.KVStore, cdc codec.BinaryCodec) error {
	oldStore := prefix.NewStore(store, v040staking.DelegationKey)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		var delegation v040staking.Delegation
		if err := cdc.Unmarshal(oldStoreIter.Value(), &delegation); err != nil {
			return err
		}

		if validator.TokensFromShares(delegation.Shares).TruncateInt().IsZero() || delegation.Shares.IsZero() {
			oldStore.Delete(oldStoreIter.Key())
		}
	}

	return nil
}
