package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// Default parameter namespace
const (
	DefaultParamSpace = "Stake"
)

// nolint - Key generators for parameter access
func KeyInflationRateChange() params.Key { return params.NewKey("InflationRateChange") }
func KeyInflationMax() params.Key        { return params.NewKey("InflationMax") }
func KeyInflationMin() params.Key        { return params.NewKey("InflationMin") }
func KeyGoalBonded() params.Key          { return params.NewKey("GoalBonded") }
func KeyUnbondingTime() params.Key       { return params.NewKey("UnbondingTime") }
func KeyMaxValidators() params.Key       { return params.NewKey("MaxValidators") }
func KeyBondDenom() params.Key           { return params.NewKey("BondDenom") }

// Cached parameter keys
var (
	keyInflationRateChange = KeyInflationRateChange()
	keyInflationMax        = KeyInflationMax()
	keyInflationMin        = KeyInflationMin()
	keyGoalBonded          = KeyGoalBonded()
	keyUnbondingTime       = KeyUnbondingTime()
	keyMaxValidators       = KeyMaxValidators()
	keyBondDenom           = KeyBondDenom()
)

// InflationRateChange - Maximum annual change in inflation rate
func (k Keeper) InflationRateChange(ctx sdk.Context) (res sdk.Rat) {
	k.paramstore.Get(ctx, keyInflationRateChange, &res)
	return
}

// InflationMax - Maximum inflation rate
func (k Keeper) InflationMax(ctx sdk.Context) (res sdk.Rat) {
	k.paramstore.Get(ctx, keyInflationMax, &res)
	return
}

// InflationMin - Minimum inflation rate
func (k Keeper) InflationMin(ctx sdk.Context) (res sdk.Rat) {
	k.paramstore.Get(ctx, keyInflationMin, &res)
	return
}

// GoalBonded - Goal of percent bonded atoms
func (k Keeper) GoalBonded(ctx sdk.Context) (res sdk.Rat) {
	k.paramstore.Get(ctx, keyGoalBonded, &res)
	return
}

// UnbondingTime
func (k Keeper) UnbondingTime(ctx sdk.Context) time.Duration {
	var t int64
	k.paramstore.Get(ctx, keyUnbondingTime, &t)
	return time.Duration(t) * time.Second
}

// MaxValidators - Maximum number of validators
func (k Keeper) MaxValidators(ctx sdk.Context) (res uint16) {
	k.paramstore.Get(ctx, keyMaxValidators, &res)
	return
}

// BondDenom - Bondable coin denomination
func (k Keeper) BondDenom(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, keyBondDenom, &res)
	return
}

// Get all parameteras as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (res types.Params) {
	res.InflationRateChange = k.InflationRateChange(ctx)
	res.InflationMax = k.InflationMax(ctx)
	res.InflationMin = k.InflationMin(ctx)
	res.GoalBonded = k.GoalBonded(ctx)
	res.UnbondingTime = k.UnbondingTime(ctx)
	res.MaxValidators = k.MaxValidators(ctx)
	res.BondDenom = k.BondDenom(ctx)
	return
}

// Need a distinct function because setParams depends on an existing previous
// record of params to exist (to check if maxValidators has changed) - and we
// panic on retrieval if it doesn't exist - hence if we use setParams for the very
// first params set it will panic.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if k.MaxValidators(ctx) != params.MaxValidators {
		k.UpdateBondedValidatorsFull(ctx)
	}
	k.SetNewParams(ctx, params)
}

// set the params without updating validator set
func (k Keeper) SetNewParams(ctx sdk.Context, params types.Params) {
	k.paramstore.Set(ctx, keyInflationRateChange, params.InflationRateChange)
	k.paramstore.Set(ctx, keyInflationMax, params.InflationMax)
	k.paramstore.Set(ctx, keyInflationMin, params.InflationMin)
	k.paramstore.Set(ctx, keyGoalBonded, params.GoalBonded)
	k.paramstore.Set(ctx, keyUnbondingTime, params.UnbondingTime)
	k.paramstore.Set(ctx, keyMaxValidators, params.MaxValidators)
	k.paramstore.Set(ctx, keyBondDenom, params.BondDenom)
}
