package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// InitGenesis - store genesis parameters
func (k PrivlegedKeeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	store := ctx.KVStore(k.storeKey)
	k.SetPool(ctx, data.Pool)
	k.SetNewParams(ctx, data.Params)
	for _, validator := range data.Validators {

		// set validator
		k.SetValidator(ctx, validator)

		// manually set indexes for the first time
		k.SetValidatorByPubKeyIndex(ctx, validator)
		k.SetValidatorByPowerIndex(ctx, validator, data.Pool)
		if validator.Status() == sdk.Bonded {
			store.Set(GetValidatorsBondedIndexKey(validator.PubKey), validator.Owner)
		}
	}
	for _, bond := range data.Bonds {
		k.SetDelegation(ctx, bond)
	}
	k.UpdateBondedValidatorsFull(ctx)
}

// WriteGenesis - output genesis parameters
func (k PrivlegedKeeper) WriteGenesis(ctx sdk.Context) types.GenesisState {
	pool := k.GetPool(ctx)
	params := k.GetParams(ctx)
	validators := k.GetAllValidators(ctx)
	bonds := k.GetAllDelegations(ctx)
	return types.GenesisState{
		pool,
		params,
		validators,
		bonds,
	}
}
