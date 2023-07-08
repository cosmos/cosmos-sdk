package v11

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/gaia/v11/app/keepers"
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

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Setting Params in the Staking Module...")
		SetParamsStaking(ctx, keepers.stakingkeeper)

		ctx.Logger().Info("Setting Validator...")
		SetAllValidatorBondAndLiquidSharesToZero(ctx, keepers.stakingkeeper)

		ctx.Logger().Info("Setting Delegation...")
		SetAllDelegationValidatorBondsFalse(ctx, keepers.stakingkeeper)

		// Refesh total liquid staked
		ctx.Logger().Info("Refreshing total liquid staked")
		if err := stakingkeeper.Keeper.RefreshTotalLiquidStaked(ctx, keepers.StakingKeeper); err != nil {
			return nil, sdkerrors.Wrap(err, "unable to refresh total liquid staked")
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
