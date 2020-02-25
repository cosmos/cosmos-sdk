package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func bootstrapValidatorTest(t *testing.T, power int64) (*simapp.SimApp, sdk.Context, []sdk.AccAddress, []sdk.ValAddress) {
	_, app, ctx := getBaseSimappWithCustomKeeper()

	addrDels, addrVals := generateAddresses(app, ctx, 100)

	amt := sdk.TokensFromConsensusPower(power)
	totalSupply := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	err := app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), totalSupply)
	require.NoError(t, err)
	app.SupplyKeeper.SetModuleAccount(ctx, notBondedPool)

	return app, ctx, addrDels, addrVals
}

func TestSetValidator(t *testing.T) {
	app, ctx, _, _ := bootstrapValidatorTest(t, 10)

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	valTokens := sdk.TokensFromConsensusPower(10)

	// test how the validator is set from a purely unbonbed pool
	validator := types.NewValidator(valAddr, valPubKey, types.Description{})
	validator, _ = validator.AddTokensFromDel(valTokens)
	require.Equal(t, sdk.Unbonded, validator.Status)
	assert.Equal(t, valTokens, validator.Tokens)
	assert.Equal(t, valTokens, validator.DelegatorShares.RoundInt())
	app.StakingKeeper.SetValidator(ctx, validator)
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, validator)

	// ensure update
	updates := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validator, found := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validator.ABCIValidatorUpdate(), updates[0])

	// after the save the validator should be bonded
	require.Equal(t, sdk.Bonded, validator.Status)
	assert.Equal(t, valTokens, validator.Tokens)
	assert.Equal(t, valTokens, validator.DelegatorShares.RoundInt())

	// Check each store for being saved
	resVal, found := app.StakingKeeper.GetValidator(ctx, valAddr)
	assert.True(ValEq(t, validator, resVal))
	require.True(t, found)

	resVals := app.StakingKeeper.GetLastValidators(ctx)
	require.Equal(t, 1, len(resVals))
	assert.True(ValEq(t, validator, resVals[0]))

	resVals = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, 1, len(resVals))
	require.True(ValEq(t, validator, resVals[0]))

	resVals = app.StakingKeeper.GetValidators(ctx, 1)
	require.Equal(t, 1, len(resVals))
	require.True(ValEq(t, validator, resVals[0]))

	resVals = app.StakingKeeper.GetValidators(ctx, 10)
	require.Equal(t, 1, len(resVals))
	require.True(ValEq(t, validator, resVals[0]))

	allVals := app.StakingKeeper.GetAllValidators(ctx)
	require.Equal(t, 1, len(allVals))
}
