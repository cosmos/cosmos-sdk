package keeper

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// UnbondingTime
func (k Keeper) UnbondingTime(ctx sdk.Context) time.Duration {
	params := k.GetParams(ctx)
	return params.UnbondingTime
}

// MaxValidators - Maximum number of validators
func (k Keeper) MaxValidators(ctx sdk.Context) uint32 {
	params := k.GetParams(ctx)
	return params.MaxValidators
}

// MaxEntries - Maximum number of simultaneous unbonding
// delegations or redelegations (per pair/trio)
func (k Keeper) MaxEntries(ctx sdk.Context) uint32 {
	params := k.GetParams(ctx)
	return params.MaxEntries
}

// HistoricalEntries = number of historical info entries
// to persist in store
func (k Keeper) HistoricalEntries(ctx sdk.Context) (res uint32) {
	params := k.GetParams(ctx)
	return params.HistoricalEntries
}

// BondDenom - Bondable coin denomination
func (k Keeper) BondDenom(ctx sdk.Context) (res string) {
	params := k.GetParams(ctx)
	return params.BondDenom
}

// PowerReduction - is the amount of staking tokens required for 1 unit of consensus-engine power.
// Currently, this returns a global variable that the app developer can tweak.
// TODO: we might turn this into an on-chain param:
// https://github.com/cosmos/cosmos-sdk/issues/8365
func (k Keeper) PowerReduction(ctx sdk.Context) math.Int {
	return sdk.DefaultPowerReduction
}

// MinCommissionRate - Minimum validator commission rate
func (k Keeper) MinCommissionRate(ctx sdk.Context) sdk.Dec {
	params := k.GetParams(ctx)
	return params.MinCommissionRate
}

//remove-comment
//// Get all parameters as types.Params
//func (k Keeper) GetParams(ctx sdk.Context) types.Params {
//	return types.NewParams(
//		k.UnbondingTime(ctx),
//		k.MaxValidators(ctx),
//		k.MaxEntries(ctx),
//		k.HistoricalEntries(ctx),
//		k.BondDenom(ctx),
//		k.MinCommissionRate(ctx),
//	)
//}

//remove-comment
// set the params
//func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
//	if err := params.Validate(); err != nil {
//		return err
//	}
//
//	k.paramstore.SetParamSet(ctx, &params)
//}

// SetParams sets the x/staking module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)

	return nil
}

// GetParams sets the x/staking module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}
