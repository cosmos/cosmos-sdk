package v11

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// Test setting params in the staking module
func TestSetParamsStaking(t *testing.T) {

	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// set the params in the staking store; exclude the new v11 params
	params := app.StakingKeeper.GetParams(ctx)
	preUpgradeParams := stakingtypes.Params{
		UnbondingTime:     params.UnbondingTime,
		MaxValidators:     params.MaxValidators,
		MaxEntries:        params.MaxEntries,
		HistoricalEntries: params.HistoricalEntries,
		BondDenom:         params.BondDenom,
	}
	app.StakingKeeper.SetParams(ctx, preUpgradeParams)

	SetParamsStaking(ctx, app.StakingKeeper)

	// check that the params were set correctly
	params = app.StakingKeeper.GetParams(ctx)
	require.Equal(t, sdk.NewDec(250), params.ValidatorBondFactor)
	require.Equal(t, sdk.NewDec(25).Quo(sdk.NewDec(100)), params.GlobalLiquidStakingCap)
	require.Equal(t, sdk.NewDec(50).Quo(sdk.NewDec(100)), params.ValidatorLiquidStakingCap)
}

// Test setting each validator's TotalValidatorBondShares and TotalLiquidShares to 0
func TestSetAllValidatorBondAndLiquidSharesToZero(t *testing.T) {

	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	for _, Val := range app.StakingKeeper.GetAllValidators(ctx) {

		// set the validator attributes on each val; exclude the new v11 attributes
		preUpgradeValidator := stakingtypes.Validator{
			OperatorAddress:   Val.OperatorAddress,
			ConsensusPubkey:   Val.ConsensusPubkey,
			Jailed:            Val.Jailed,
			Status:            Val.Status,
			Tokens:            Val.Tokens,
			DelegatorShares:   Val.DelegatorShares,
			Description:       Val.Description,
			UnbondingHeight:   Val.UnbondingHeight,
			MinSelfDelegation: Val.MinSelfDelegation,
			Commission:        Val.Commission,
			UnbondingTime:     Val.UnbondingTime,
		}
		app.StakingKeeper.SetValidator(ctx, preUpgradeValidator)
	}

	SetAllValidatorBondAndLiquidSharesToZero(ctx, app.StakingKeeper)

	// check that the validator TotalValidatorBondShares and TotalLiquidShares were correctly set to 0
	for _, Val := range app.StakingKeeper.GetAllValidators(ctx) {
		require.Equal(t, sdk.ZeroDec(), Val.TotalValidatorBondShares)
		require.Equal(t, sdk.ZeroDec(), Val.TotalLiquidShares)
	}
}

// Test setting each validator's TotalDelegatorBondShares to 0
func TestSetAllDelegatorBondSharesToZero(t *testing.T) {

	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	for _, Del := range app.StakingKeeper.GetAllDelegations(ctx) {

		// set the delegation attributes on each del; exclude the new v11 attributes
		preUpgradeDelegation := stakingtypes.Delegation{
			DelegatorAddress: Del.DelegatorAddress,
			ValidatorAddress: Del.ValidatorAddress,
			Shares:           Del.Shares,
		}
		app.StakingKeeper.SetDelegation(ctx, preUpgradeDelegation)
	}

	SetAllDelegationValidatorBondsFalse(ctx, app.StakingKeeper)

	// check that the delegation ValidatorBond was correctly set to false
	for _, Del := range app.StakingKeeper.GetAllDelegations(ctx) {
		require.Equal(t, false, Del.ValidatorBond)
	}

}
