package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// Default parameter namespace
const (
	DefaultParamspace = "stake"
)

// ParamTable for stake module
func ParamTypeTable() params.TypeTable {
	return params.NewTypeTable().RegisterParamSet(&types.Params{})
}

// InflationRateChange - Maximum annual change in inflation rate
func (k Keeper) InflationRateChange(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyInflationRateChange, &res)
	return
}

// InflationMax - Maximum inflation rate
func (k Keeper) InflationMax(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyInflationMax, &res)
	return
}

// InflationMin - Minimum inflation rate
func (k Keeper) InflationMin(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyInflationMin, &res)
	return
}

// GoalBonded - Goal of percent bonded atoms
func (k Keeper) GoalBonded(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyGoalBonded, &res)
	return
}

// UnbondingTime
func (k Keeper) UnbondingTime(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyUnbondingTime, &res)
	return
}

// MaxValidators - Maximum number of validators
func (k Keeper) MaxValidators(ctx sdk.Context) (res uint16) {
	k.paramstore.Get(ctx, types.KeyMaxValidators, &res)
	return
}

// BondDenom - Bondable coin denomination
func (k Keeper) BondDenom(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyBondDenom, &res)
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

// set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
