package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// subspace contains the method needed for migrations of the
// legacy Params subspace
type subspace interface {
	GetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
	HasKeyTable() bool
	WithKeyTable(paramtypes.KeyTable) paramtypes.Subspace
	Set(ctx sdk.Context, key []byte, value interface{})
}

// Adds the following LSM params:
// - ValidatorBondFactor
// - GlobalLiquidStakingCap
// - ValidatorLiquidStakingCap
func migrateParamsStore(ctx sdk.Context, paramstore subspace) {
	if paramstore.HasKeyTable() {
		paramstore.WithKeyTable(types.ParamKeyTable())
	}
	paramstore.Set(ctx, types.KeyValidatorBondFactor, types.DefaultValidatorBondFactor)
	paramstore.Set(ctx, types.KeyGlobalLiquidStakingCap, types.DefaultGlobalLiquidStakingCap)
	paramstore.Set(ctx, types.KeyValidatorLiquidStakingCap, types.DefaultValidatorLiquidStakingCap)
}

// Set each validator's TotalValidatorBondShares and TotalLiquidShares to 0
func migrateValidators(ctx sdk.Context, k stakingkeeper.Keeper) {
	for _, validator := range k.GetAllValidators(ctx) {
		validator.TotalValidatorBondShares = sdk.ZeroDec()
		validator.TotalLiquidShares = sdk.ZeroDec()
		k.SetValidator(ctx, validator)
	}
}

// Set each delegation's ValidatorBond field to false
func migrateDelegations(ctx sdk.Context, k stakingkeeper.Keeper) {
	for _, delegation := range k.GetAllDelegations(ctx) {
		delegation.ValidatorBond = false
		k.SetDelegation(ctx, delegation)
	}
}

// Peforms the in-place store migration for adding LSM support to v0.45.16-ics, including:
//   - Adding params ValidatorBondFactor, GlobalLiquidStakingCap, ValidatorLiquidStakingCap
//   - Setting each validator's TotalValidatorBondShares and TotalLiquidShares to 0
//   - Setting each delegation's ValidatorBond field to false
//   - Calculating the total liquid staked by summing the delegations from ICA accounts
func MigrateStore(ctx sdk.Context, k stakingkeeper.Keeper, paramstore subspace) error {
	k.Logger(ctx).Info("Migrating param store")
	migrateParamsStore(ctx, paramstore)

	k.Logger(ctx).Info("Migrating validators")
	migrateValidators(ctx, k)

	k.Logger(ctx).Info("Migrating delegations")
	migrateDelegations(ctx, k)

	k.Logger(ctx).Info("Calculating total liquid staked")
	return k.RefreshTotalLiquidStaked(ctx)
}
