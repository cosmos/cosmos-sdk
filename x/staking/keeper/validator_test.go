package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/cosmos/cosmos-sdk/simapp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func bootstrapValidatorTest(t *testing.T, power int64, numAddrs int) (*simapp.SimApp, sdk.Context, []sdk.AccAddress, []sdk.ValAddress) {
	_, app, ctx := getBaseSimappWithCustomKeeper()

	addrDels, addrVals := generateAddresses(app, ctx, numAddrs)

	amt := sdk.TokensFromConsensusPower(power)
	totalSupply := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	err := app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), totalSupply)
	require.NoError(t, err)
	app.SupplyKeeper.SetModuleAccount(ctx, notBondedPool)

	app.SupplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))

	return app, ctx, addrDels, addrVals
}

func TestSetValidator(t *testing.T) {
	app, ctx, _, _ := bootstrapValidatorTest(t, 10, 100)

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

func TestUpdateValidatorByPowerIndex(t *testing.T) {
	app, ctx, _, _ := bootstrapValidatorTest(t, 0, 100)
	_, addrVals := generateAddresses(app, ctx, 1)

	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	app.BankKeeper.SetBalances(ctx, bondedPool.GetAddress(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.TokensFromConsensusPower(1234))))
	app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.TokensFromConsensusPower(10000))))
	app.SupplyKeeper.SetModuleAccount(ctx, bondedPool)
	app.SupplyKeeper.SetModuleAccount(ctx, notBondedPool)

	// add a validator
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	validator, delSharesCreated := validator.AddTokensFromDel(sdk.TokensFromConsensusPower(100))
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, sdk.TokensFromConsensusPower(100), validator.Tokens)
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
	validator, found := app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, sdk.TokensFromConsensusPower(100), validator.Tokens)

	power := types.GetValidatorsByPowerIndexKey(validator)
	require.True(t, keeper.ValidatorByPowerIndexExists(ctx, app.StakingKeeper, power))

	// burn half the delegator shares
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validator)
	validator, burned := validator.RemoveDelShares(delSharesCreated.Quo(sdk.NewDec(2)))
	require.Equal(t, sdk.TokensFromConsensusPower(50), burned)
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true) // update the validator, possibly kicking it out
	require.False(t, keeper.ValidatorByPowerIndexExists(ctx, app.StakingKeeper, power))

	validator, found = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)

	power = types.GetValidatorsByPowerIndexKey(validator)
	require.True(t, keeper.ValidatorByPowerIndexExists(ctx, app.StakingKeeper, power))
}

func TestUpdateBondedValidatorsDecreaseCliff(t *testing.T) {
	numVals := 10
	maxVals := 5

	// create context, keeper, and pool for tests
	app, ctx, _, valAddrs := bootstrapValidatorTest(t, 0, 100)

	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	// create keeper parameters
	params := app.StakingKeeper.GetParams(ctx)
	params.MaxValidators = uint32(maxVals)
	app.StakingKeeper.SetParams(ctx, params)

	// create a random pool
	app.BankKeeper.SetBalances(ctx, bondedPool.GetAddress(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.TokensFromConsensusPower(1234))))
	app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.TokensFromConsensusPower(10000))))
	app.SupplyKeeper.SetModuleAccount(ctx, bondedPool)
	app.SupplyKeeper.SetModuleAccount(ctx, notBondedPool)

	validators := make([]types.Validator, numVals)
	for i := 0; i < len(validators); i++ {
		moniker := fmt.Sprintf("val#%d", int64(i))
		val := types.NewValidator(valAddrs[i], PKs[i], types.Description{Moniker: moniker})
		delTokens := sdk.TokensFromConsensusPower(int64((i + 1) * 10))
		val, _ = val.AddTokensFromDel(delTokens)

		val = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, val, true)
		validators[i] = val
	}

	nextCliffVal := validators[numVals-maxVals+1]

	// remove enough tokens to kick out the validator below the current cliff
	// validator and next in line cliff validator
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, nextCliffVal)
	shares := sdk.TokensFromConsensusPower(21)
	nextCliffVal, _ = nextCliffVal.RemoveDelShares(shares.ToDec())
	nextCliffVal = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, nextCliffVal, true)

	expectedValStatus := map[int]sdk.BondStatus{
		9: sdk.Bonded, 8: sdk.Bonded, 7: sdk.Bonded, 5: sdk.Bonded, 4: sdk.Bonded,
		0: sdk.Unbonding, 1: sdk.Unbonding, 2: sdk.Unbonding, 3: sdk.Unbonding, 6: sdk.Unbonding,
	}

	// require all the validators have their respective statuses
	for valIdx, status := range expectedValStatus {
		valAddr := validators[valIdx].OperatorAddress
		val, _ := app.StakingKeeper.GetValidator(ctx, valAddr)

		assert.Equal(
			t, status, val.GetStatus(),
			fmt.Sprintf("expected validator at index %v to have status: %s", valIdx, status),
		)
	}
}

func TestSlashToZeroPowerRemoved(t *testing.T) {
	// initialize setup
	app, ctx, _, addrVals := bootstrapValidatorTest(t, 100, 20)

	// add a validator
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	valTokens := sdk.TokensFromConsensusPower(100)

	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	err := app.BankKeeper.SetBalances(ctx, bondedPool.GetAddress(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), valTokens)))
	require.NoError(t, err)
	app.SupplyKeeper.SetModuleAccount(ctx, bondedPool)

	validator, _ = validator.AddTokensFromDel(valTokens)
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, valTokens, validator.Tokens)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, validator)
	validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
	require.Equal(t, valTokens, validator.Tokens, "\nvalidator %v\npool %v", validator, valTokens)

	// slash the validator by 100%
	app.StakingKeeper.Slash(ctx, sdk.ConsAddress(PKs[0].Address()), 0, 100, sdk.OneDec())
	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	// validator should be unbonding
	validator, _ = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.Equal(t, validator.GetStatus(), sdk.Unbonding)
}
