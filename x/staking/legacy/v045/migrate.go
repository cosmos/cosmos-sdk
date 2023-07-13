package v11

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// Set initial param values, based on https://github.com/iqlusioninc/liquidity-staking-module/blob/master/x/staking/spec/08_params.md
func SetParamsStaking(ctx sdk.Context, k stakingkeeper.Keeper) {
	params := k.GetParams(ctx)

	params.ValidatorBondFactor = sdk.Dec(sdk.NewInt(250))
	params.GlobalLiquidStakingCap = sdk.Dec(sdk.NewInt(25).Quo(sdk.NewInt(100)))
	params.ValidatorLiquidStakingCap = sdk.Dec(sdk.NewInt(50).Quo(sdk.NewInt(100)))

	k.SetParams(ctx, params)
}

// Set each validator's TotalValidatorBondShares and TotalLiquidShares to 0
func SetAllValidatorBondAndLiquidSharesToZero(ctx sdk.Context, k stakingkeeper.Keeper) {

	for _, Val := range k.GetAllValidators(ctx) {

		Val.TotalValidatorBondShares = sdk.ZeroDec()
		Val.TotalLiquidShares = sdk.ZeroDec()

		k.SetValidator(ctx, Val)
	}
}

// Set each validator's ValidatorBond to false
func SetAllDelegationValidatorBondsFalse(ctx sdk.Context, k stakingkeeper.Keeper) {
	for _, Del := range k.GetAllDelegations(ctx) {

		Del.ValidatorBond = false

		k.SetDelegation(ctx, Del)
	}
}
