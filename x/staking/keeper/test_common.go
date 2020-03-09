package keeper // noalias

import (
	"bytes"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// does a certain by-power index record exist
func ValidatorByPowerIndexExists(ctx sdk.Context, keeper Keeper, power []byte) bool {
	store := ctx.KVStore(keeper.storeKey)
	return store.Has(power)
}

// update validator for testing
func TestingUpdateValidator(keeper Keeper, ctx sdk.Context, validator types.Validator, apply bool) types.Validator {
	keeper.SetValidator(ctx, validator)

	// Remove any existing power key for validator.
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ValidatorsByPowerIndexKey)
	defer iterator.Close()
	deleted := false
	for ; iterator.Valid(); iterator.Next() {
		valAddr := types.ParseValidatorPowerRankKey(iterator.Key())
		if bytes.Equal(valAddr, validator.OperatorAddress) {
			if deleted {
				panic("found duplicate power index key")
			} else {
				deleted = true
			}
			store.Delete(iterator.Key())
		}
	}

	keeper.SetValidatorByPowerIndex(ctx, validator)
	if apply {
		keeper.ApplyAndReturnValidatorSetUpdates(ctx)
		validator, found := keeper.GetValidator(ctx, validator.OperatorAddress)
		if !found {
			panic("validator expected but not found")
		}
		return validator
	}

	cachectx, _ := ctx.CacheContext()
	keeper.ApplyAndReturnValidatorSetUpdates(cachectx)

	validator, found := keeper.GetValidator(cachectx, validator.OperatorAddress)
	if !found {
		panic("validator expected but not found")
	}

	return validator
}

// RandomValidator returns a random validator given access to the keeper and ctx
func RandomValidator(r *rand.Rand, keeper Keeper, ctx sdk.Context) (val types.Validator, ok bool) {
	vals := keeper.GetAllValidators(ctx)
	if len(vals) == 0 {
		return types.Validator{}, false
	}

	i := r.Intn(len(vals))
	return vals[i], true
}
