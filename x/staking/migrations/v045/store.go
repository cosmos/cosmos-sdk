package v045

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040staking "github.com/cosmos/cosmos-sdk/x/staking/migrations/v040"
)

// MigrateStore performs in-place store migrations from v0.43/v0.44 to v0.45.
// The migration includes:
//
// - Removing delegations that have a zero share or token amount.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)

	return purgeDelegations(store, cdc)
}

func purgeDelegations(store sdk.KVStore, cdc codec.BinaryCodec) error {
	prefixDelStore := prefix.NewStore(store, v040staking.DelegationKey)

	delStoreIter := prefixDelStore.Iterator(nil, nil)
	defer delStoreIter.Close()

	valCache := make(map[string]v040staking.Validator)

	for ; delStoreIter.Valid(); delStoreIter.Next() {
		var delegation v040staking.Delegation
		if err := cdc.Unmarshal(delStoreIter.Value(), &delegation); err != nil {
			return err
		}

		validator, ok := valCache[delegation.ValidatorAddress]
		if !ok {
			valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
			if err != nil {
				return err
			}

			if err := cdc.Unmarshal(store.Get(getValidatorKey(valAddr)), &validator); err != nil {
				return err
			}

			valCache[delegation.ValidatorAddress] = validator
		}

		// TODO: On-chain, we call BeforeDelegationRemoved prior to removing the
		// object from state. Do we need to do the same here?
		if validator.TokensFromShares(delegation.Shares).TruncateInt().IsZero() || delegation.Shares.IsZero() {
			prefixDelStore.Delete(delStoreIter.Key())
		}
	}

	return nil
}
