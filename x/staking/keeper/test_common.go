package keeper // noalias

import (
	"bytes"
	"context"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// does a certain by-power index record exist
func ValidatorByPowerIndexExists(ctx context.Context, keeper *Keeper, power []byte) bool {
	store := keeper.storeService.OpenKVStore(ctx)
	has, err := store.Has(power)
	if err != nil {
		panic(err)
	}
	return has
}

// update validator for testing
func TestingUpdateValidator(keeper *Keeper, ctx sdk.Context, validator types.Validator, apply bool) types.Validator {
	err := keeper.SetValidator(ctx, validator)
	if err != nil {
		panic(err)
	}

	// Remove any existing power key for validator.
	store := keeper.storeService.OpenKVStore(ctx)
	deleted := false

	iterator, err := store.Iterator(types.ValidatorsByPowerIndexKey, storetypes.PrefixEndBytes(types.ValidatorsByPowerIndexKey))
	if err != nil {
		panic(err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		valAddr := types.ParseValidatorPowerRankKey(iterator.Key())
		if bytes.Equal(valAddr, validator.GetOperator()) {
			if deleted {
				panic("found duplicate power index key")
			} else {
				deleted = true
			}

			if err = store.Delete(iterator.Key()); err != nil {
				panic(err)
			}
		}
	}

	if err = keeper.SetValidatorByPowerIndex(ctx, validator); err != nil {
		panic(err)
	}

	if !apply {
		ctx, _ = ctx.CacheContext()
	}
	_, err = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	if err != nil {
		panic(err)
	}

	validator, err = keeper.GetValidator(ctx, validator.GetOperator())
	if err != nil {
		panic(err)
	}

	return validator
}
