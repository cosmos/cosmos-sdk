package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

// initConfig creates 3 validators and bootstrap the app.
func initConfig(t *testing.T) (*simapp.SimApp, sdk.Context, []sdk.ValAddress) {
	_, app, ctx := getBaseSimappWithCustomKeeper()

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 3, sdk.NewInt(10000))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	numVals := int64(3)
	amt := sdk.TokensFromConsensusPower(10)
	bondedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(numVals)))

	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	err := app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), bondedCoins)
	require.NoError(t, err)
	app.SupplyKeeper.SetModuleAccount(ctx, notBondedPool)

	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	err = app.BankKeeper.SetBalances(ctx, bondedPool.GetAddress(), bondedCoins)
	require.NoError(t, err)
	app.SupplyKeeper.SetModuleAccount(ctx, bondedPool)

	for i := int64(0); i < numVals; i++ {
		validator := types.NewValidator(addrVals[i], PKs[i], types.Description{})
		validator, _ = validator.AddTokensFromDel(amt)
		validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
		app.StakingKeeper.SetValidatorByConsAddr(ctx, validator)
	}

	return app, ctx, addrVals
}

// tests Jail, Unjail
func TestRevocation(t *testing.T) {
	app, ctx, addrVals := initConfig(t)

	consAddr := sdk.ConsAddress(PKs[0].Address())

	// initial state
	val, found := app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.False(t, val.IsJailed())

	// test jail
	app.StakingKeeper.Jail(ctx, consAddr)
	val, found = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.True(t, val.IsJailed())

	// test unjail
	app.StakingKeeper.Unjail(ctx, consAddr)
	val, found = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.False(t, val.IsJailed())
}
