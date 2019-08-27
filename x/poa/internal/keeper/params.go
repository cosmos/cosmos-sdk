package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/poa/internal/types"
)

// Default parameter namespace
const (
	DefaultParamspace = types.ModuleName
)

// ParamTable for POA module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&types.Params{})
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

// MaxEntries - Maximum number of simultaneous unbonding
// delegations or redelegations (per pair/trio)
func (k Keeper) MaxEntries(ctx sdk.Context) (res uint16) {
	k.paramstore.Get(ctx, types.KeyMaxEntries, &res)
	return
}

// Increase - Bondable coin denomination
func (k Keeper) AcceptAllValidators(ctx sdk.Context) (res bool) {
	k.paramstore.Get(ctx, types.KeyAcceptAllValidators, &res)
	return
}

// Increase - Bondable coin denomination
func (k Keeper) IncreaseWeight(ctx sdk.Context) (res bool) {
	k.paramstore.Get(ctx, types.KeyIncreaseWeight, &res)
	return
}

// Get all parameteras as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.UnbondingTime(ctx),
		k.MaxValidators(ctx),
		k.MaxEntries(ctx),
		k.AcceptAllValidators(ctx),
		k.IncreaseWeight(ctx),
	)
}

// set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
