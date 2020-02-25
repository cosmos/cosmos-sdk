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

// This function tests UpdateValidator, GetValidator, GetLastValidators, RemoveValidator
func TestValidatorBasics(t *testing.T) {
	app, ctx, _, addrVals := bootstrapValidatorTest(t, 1000, 20)

	//construct the validators
	var validators [3]types.Validator
	powers := []int64{9, 8, 7}
	for i, power := range powers {
		validators[i] = types.NewValidator(addrVals[i], PKs[i], types.Description{})
		validators[i].Status = sdk.Unbonded
		validators[i].Tokens = sdk.ZeroInt()
		tokens := sdk.TokensFromConsensusPower(power)

		validators[i], _ = validators[i].AddTokensFromDel(tokens)
	}
	assert.Equal(t, sdk.TokensFromConsensusPower(9), validators[0].Tokens)
	assert.Equal(t, sdk.TokensFromConsensusPower(8), validators[1].Tokens)
	assert.Equal(t, sdk.TokensFromConsensusPower(7), validators[2].Tokens)

	// check the empty keeper first
	_, found := app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.False(t, found)
	resVals := app.StakingKeeper.GetLastValidators(ctx)
	require.Zero(t, len(resVals))

	resVals = app.StakingKeeper.GetValidators(ctx, 2)
	require.Zero(t, len(resVals))

	// set and retrieve a record
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], true)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, validators[0])
	resVal, found := app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	assert.True(ValEq(t, validators[0], resVal))

	// retrieve from consensus
	resVal, found = app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.ConsAddress(PKs[0].Address()))
	require.True(t, found)
	assert.True(ValEq(t, validators[0], resVal))
	resVal, found = app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(PKs[0]))
	require.True(t, found)
	assert.True(ValEq(t, validators[0], resVal))

	resVals = app.StakingKeeper.GetLastValidators(ctx)
	require.Equal(t, 1, len(resVals))
	assert.True(ValEq(t, validators[0], resVals[0]))
	assert.Equal(t, sdk.Bonded, validators[0].Status)
	assert.True(sdk.IntEq(t, sdk.TokensFromConsensusPower(9), validators[0].BondedTokens()))

	// modify a records, save, and retrieve
	validators[0].Status = sdk.Bonded
	validators[0].Tokens = sdk.TokensFromConsensusPower(10)
	validators[0].DelegatorShares = validators[0].Tokens.ToDec()
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], true)
	resVal, found = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	assert.True(ValEq(t, validators[0], resVal))

	resVals = app.StakingKeeper.GetLastValidators(ctx)
	require.Equal(t, 1, len(resVals))
	assert.True(ValEq(t, validators[0], resVals[0]))

	// add other validators
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], true)
	validators[2] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[2], true)
	resVal, found = app.StakingKeeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	assert.True(ValEq(t, validators[1], resVal))
	resVal, found = app.StakingKeeper.GetValidator(ctx, addrVals[2])
	require.True(t, found)
	assert.True(ValEq(t, validators[2], resVal))

	resVals = app.StakingKeeper.GetLastValidators(ctx)
	require.Equal(t, 3, len(resVals))
	assert.True(ValEq(t, validators[0], resVals[0])) // order doesn't matter here
	assert.True(ValEq(t, validators[1], resVals[1]))
	assert.True(ValEq(t, validators[2], resVals[2]))

	// remove a record

	// shouldn't be able to remove if status is not unbonded
	assert.PanicsWithValue(t,
		"cannot call RemoveValidator on bonded or unbonding validators",
		func() { app.StakingKeeper.RemoveValidator(ctx, validators[1].OperatorAddress) })

	// shouldn't be able to remove if there are still tokens left
	validators[1].Status = sdk.Unbonded
	app.StakingKeeper.SetValidator(ctx, validators[1])
	assert.PanicsWithValue(t,
		"attempting to remove a validator which still contains tokens",
		func() { app.StakingKeeper.RemoveValidator(ctx, validators[1].OperatorAddress) })

	validators[1].Tokens = sdk.ZeroInt()                                  // ...remove all tokens
	app.StakingKeeper.SetValidator(ctx, validators[1])                    // ...set the validator
	app.StakingKeeper.RemoveValidator(ctx, validators[1].OperatorAddress) // Now it can be removed.
	_, found = app.StakingKeeper.GetValidator(ctx, addrVals[1])
	require.False(t, found)
}

// test how the validators are sorted, tests GetBondedValidatorsByPower
func TestGetValidatorSortingUnmixed(t *testing.T) {
	app, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	// initialize some validators into the state
	amts := []int64{
		0,
		100 * sdk.PowerReduction.Int64(),
		1 * sdk.PowerReduction.Int64(),
		400 * sdk.PowerReduction.Int64(),
		200 * sdk.PowerReduction.Int64()}
	n := len(amts)
	var validators [5]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(sdk.ValAddress(addrs[i]), PKs[i], types.Description{})
		validators[i].Status = sdk.Bonded
		validators[i].Tokens = sdk.NewInt(amt)
		validators[i].DelegatorShares = sdk.NewDec(amt)
		keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[i], true)
	}

	// first make sure everything made it in to the gotValidator group
	resValidators := app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, n, len(resValidators))
	assert.Equal(t, sdk.NewInt(400).Mul(sdk.PowerReduction), resValidators[0].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(200).Mul(sdk.PowerReduction), resValidators[1].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(100).Mul(sdk.PowerReduction), resValidators[2].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(1).Mul(sdk.PowerReduction), resValidators[3].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(0), resValidators[4].BondedTokens(), "%v", resValidators)
	assert.Equal(t, validators[3].OperatorAddress, resValidators[0].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[4].OperatorAddress, resValidators[1].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[1].OperatorAddress, resValidators[2].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[2].OperatorAddress, resValidators[3].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[0].OperatorAddress, resValidators[4].OperatorAddress, "%v", resValidators)

	// test a basic increase in voting power
	validators[3].Tokens = sdk.NewInt(500).Mul(sdk.PowerReduction)
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))

	// test a decrease in voting power
	validators[3].Tokens = sdk.NewInt(300).Mul(sdk.PowerReduction)
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))

	// test equal voting power, different age
	validators[3].Tokens = sdk.NewInt(200).Mul(sdk.PowerReduction)
	ctx = ctx.WithBlockHeight(10)
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))

	// no change in voting power - no change in sort
	ctx = ctx.WithBlockHeight(20)
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[4], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))

	// change in voting power of both validators, both still in v-set, no age change
	validators[3].Tokens = sdk.NewInt(300).Mul(sdk.PowerReduction)
	validators[4].Tokens = sdk.NewInt(300).Mul(sdk.PowerReduction)
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	ctx = ctx.WithBlockHeight(30)
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[4], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n, "%v", resValidators)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))
}

func TestGetValidatorSortingMixed(t *testing.T) {
	app, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	app.BankKeeper.SetBalances(ctx, bondedPool.GetAddress(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.TokensFromConsensusPower(501))))
	app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.TokensFromConsensusPower(0))))
	app.SupplyKeeper.SetModuleAccount(ctx, notBondedPool)
	app.SupplyKeeper.SetModuleAccount(ctx, bondedPool)

	// now 2 max resValidators
	params := app.StakingKeeper.GetParams(ctx)
	params.MaxValidators = 2
	app.StakingKeeper.SetParams(ctx, params)

	// initialize some validators into the state
	amts := []int64{
		0,
		100 * sdk.PowerReduction.Int64(),
		1 * sdk.PowerReduction.Int64(),
		400 * sdk.PowerReduction.Int64(),
		200 * sdk.PowerReduction.Int64()}

	var validators [5]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(sdk.ValAddress(addrs[i]), PKs[i], types.Description{})
		validators[i].DelegatorShares = sdk.NewDec(amt)
		validators[i].Status = sdk.Bonded
		validators[i].Tokens = sdk.NewInt(amt)
		keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[i], true)
	}

	val0, found := app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[0]))
	require.True(t, found)
	val1, found := app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[1]))
	require.True(t, found)
	val2, found := app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[2]))
	require.True(t, found)
	val3, found := app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[3]))
	require.True(t, found)
	val4, found := app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[4]))
	require.True(t, found)
	require.Equal(t, sdk.Bonded, val0.Status)
	require.Equal(t, sdk.Unbonding, val1.Status)
	require.Equal(t, sdk.Unbonding, val2.Status)
	require.Equal(t, sdk.Bonded, val3.Status)
	require.Equal(t, sdk.Bonded, val4.Status)

	// first make sure everything made it in to the gotValidator group
	resValidators := app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	// The validators returned should match the max validators
	assert.Equal(t, 2, len(resValidators))
	assert.Equal(t, sdk.NewInt(400).Mul(sdk.PowerReduction), resValidators[0].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(200).Mul(sdk.PowerReduction), resValidators[1].BondedTokens(), "%v", resValidators)
	assert.Equal(t, validators[3].OperatorAddress, resValidators[0].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[4].OperatorAddress, resValidators[1].OperatorAddress, "%v", resValidators)
}

// TODO separate out into multiple tests
func TestGetValidatorsEdgeCases(t *testing.T) {
	app, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	// set max validators to 2
	params := app.StakingKeeper.GetParams(ctx)
	nMax := uint32(2)
	params.MaxValidators = nMax
	app.StakingKeeper.SetParams(ctx, params)

	// initialize some validators into the state
	powers := []int64{0, 100, 400, 400}
	var validators [4]types.Validator
	for i, power := range powers {
		moniker := fmt.Sprintf("val#%d", int64(i))
		validators[i] = types.NewValidator(sdk.ValAddress(addrs[i]), PKs[i], types.Description{Moniker: moniker})

		tokens := sdk.TokensFromConsensusPower(power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

		notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
		balances := app.BankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
		require.NoError(t, app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), balances.Add(sdk.NewCoin(params.BondDenom, tokens))))

		app.SupplyKeeper.SetModuleAccount(ctx, notBondedPool)
		validators[i] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[i], true)
	}

	// ensure that the first two bonded validators are the largest validators
	resValidators := app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint32(len(resValidators)))
	assert.True(ValEq(t, validators[2], resValidators[0]))
	assert.True(ValEq(t, validators[3], resValidators[1]))

	// delegate 500 tokens to validator 0
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	delTokens := sdk.TokensFromConsensusPower(500)
	validators[0], _ = validators[0].AddTokensFromDel(delTokens)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	newTokens := sdk.NewCoins()
	balances := app.BankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
	require.NoError(t, app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), balances.Add(newTokens...)))
	app.SupplyKeeper.SetModuleAccount(ctx, notBondedPool)

	// test that the two largest validators are
	//   a) validator 0 with 500 tokens
	//   b) validator 2 with 400 tokens (delegated before validator 3)
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint32(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))

	// A validator which leaves the bonded validator set due to a decrease in voting power,
	// then increases to the original voting power, does not get its spot back in the
	// case of a tie.
	//
	// Order of operations for this test:
	//  - validator 3 enter validator set with 1 new token
	//  - validator 3 removed validator set by removing 201 tokens (validator 2 enters)
	//  - validator 3 adds 200 tokens (equal to validator 2 now) and does not get its spot back

	// validator 3 enters bonded validator set
	ctx = ctx.WithBlockHeight(40)

	var found bool
	validators[3], found = app.StakingKeeper.GetValidator(ctx, validators[3].OperatorAddress)
	assert.True(t, found)
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	validators[3], _ = validators[3].AddTokensFromDel(sdk.TokensFromConsensusPower(1))

	notBondedPool = app.StakingKeeper.GetNotBondedPool(ctx)
	newTokens = sdk.NewCoins(sdk.NewCoin(params.BondDenom, sdk.TokensFromConsensusPower(1)))
	balances = app.BankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
	require.NoError(t, app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), balances.Add(newTokens...)))
	app.SupplyKeeper.SetModuleAccount(ctx, notBondedPool)

	validators[3] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint32(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[3], resValidators[1]))

	// validator 3 kicked out temporarily
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	rmTokens := validators[3].TokensFromShares(sdk.NewDec(201)).TruncateInt()
	validators[3], _ = validators[3].RemoveDelShares(sdk.NewDec(201))

	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	balances = app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	require.NoError(t, app.BankKeeper.SetBalances(ctx, bondedPool.GetAddress(), balances.Add(sdk.NewCoin(params.BondDenom, rmTokens))))
	app.SupplyKeeper.SetModuleAccount(ctx, bondedPool)

	validators[3] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint32(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))

	// validator 3 does not get spot back
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	validators[3], _ = validators[3].AddTokensFromDel(sdk.NewInt(200))

	notBondedPool = app.StakingKeeper.GetNotBondedPool(ctx)
	balances = app.BankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
	require.NoError(t, app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), balances.Add(sdk.NewCoin(params.BondDenom, sdk.NewInt(200)))))
	app.SupplyKeeper.SetModuleAccount(ctx, notBondedPool)

	validators[3] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint32(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))
	_, exists := app.StakingKeeper.GetValidator(ctx, validators[3].OperatorAddress)
	require.True(t, exists)
}
