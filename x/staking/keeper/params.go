package keeper

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// UnbondingTime - The time duration for unbonding
func (k Keeper) UnbondingTime(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).UnbondingTime
}

// MaxValidators - Maximum number of validators
func (k Keeper) MaxValidators(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).MaxValidators
}

// MaxEntries - Maximum number of simultaneous unbonding
// delegations or redelegations (per pair/trio)
func (k Keeper) MaxEntries(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).MaxEntries
}

// HistoricalEntries = number of historical info entries
// to persist in store
func (k Keeper) HistoricalEntries(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).HistoricalEntries
}

// BondDenom - Bondable coin denomination
func (k Keeper) BondDenom(ctx sdk.Context) string {
	return k.GetParams(ctx).BondDenom
}

// PowerReduction - is the amount of staking tokens required for 1 unit of consensus-engine power.
// Currently, this returns a global variable that the app developer can tweak.
// TODO: we might turn this into an on-chain param:
// https://github.com/cosmos/cosmos-sdk/issues/8365
func (k Keeper) PowerReduction(ctx sdk.Context) math.Int {
	return sdk.DefaultPowerReduction
}

// MinCommissionRate - Minimum validator commission rate
func (k Keeper) MinCommissionRate(ctx sdk.Context) math.LegacyDec {
	return k.GetParams(ctx).MinCommissionRate
}

// ValidatorBondFactor - validator bond factor for all validators
//
// Since: cosmos-sdk 0.47-lsm
func (k Keeper) ValidatorBondFactor(ctx sdk.Context) (res sdk.Dec) {
	return k.GetParams(ctx).ValidatorBondFactor
}

// Global liquid staking cap across all liquid staking providers
//
// Since: cosmos-sdk 0.47-lsm
func (k Keeper) GlobalLiquidStakingCap(ctx sdk.Context) (res sdk.Dec) {
	return k.GetParams(ctx).GlobalLiquidStakingCap
}

// Liquid staking cap for each validator
//
// Since: cosmos-sdk 0.47-lsm
func (k Keeper) ValidatorLiquidStakingCap(ctx sdk.Context) (res sdk.Dec) {
	return k.GetParams(ctx).ValidatorLiquidStakingCap
}

// SetParams sets the x/staking module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
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
