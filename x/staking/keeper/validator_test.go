package keeper

import (
	"fmt"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//_______________________________________________________

func TestSetValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 10)

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	valTokens := sdk.TokensFromTendermintPower(10)

	// test how the validator is set from a purely unbonbed pool
	validator := types.NewValidator(valAddr, valPubKey, types.Description{})
	validator, _ = keeper.AddTokensFromDel(ctx, validator, valTokens)
	require.Equal(t, sdk.Unbonded, validator.Status)
	assert.Equal(t, valTokens, validator.Tokens)
	assert.Equal(t, valTokens, validator.DelegatorShares.RoundInt())
	keeper.SetValidator(ctx, validator)
	keeper.SetValidatorByPowerIndex(ctx, validator)

	// ensure update
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validator, found := keeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validator.ABCIValidatorUpdate(), updates[0])

	// after the save the validator should be bonded
	require.Equal(t, sdk.Bonded, validator.Status)
	assert.Equal(t, valTokens, validator.Tokens)
	assert.Equal(t, valTokens, validator.DelegatorShares.RoundInt())

	// Check each store for being saved
	resVal, found := keeper.GetValidator(ctx, valAddr)
	assert.True(ValEq(t, validator, resVal))
	require.True(t, found)

	resVals := keeper.GetLastValidators(ctx)
	require.Equal(t, 1, len(resVals))
	assert.True(ValEq(t, validator, resVals[0]))

	resVals = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, 1, len(resVals))
	require.True(ValEq(t, validator, resVals[0]))

	resVals = keeper.GetValidators(ctx, 1)
	require.Equal(t, 1, len(resVals))
	require.True(ValEq(t, validator, resVals[0]))

	resVals = keeper.GetValidators(ctx, 10)
	require.Equal(t, 1, len(resVals))
	require.True(ValEq(t, validator, resVals[0]))

	allVals := keeper.GetAllValidators(ctx)
	require.Equal(t, 1, len(allVals))
}

func TestUpdateValidatorByPowerIndex(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	bondedPool, notBondedPool := keeper.GetPools(ctx)

	// create a random pool
	bondedPool.SetCoins(sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), sdk.TokensFromTendermintPower(1234))))
	notBondedPool.SetCoins(sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), sdk.TokensFromTendermintPower(10000))))
	keeper.SetBondedPool(ctx, bondedPool)
	keeper.SetNotBondedPool(ctx, notBondedPool)

	// add a validator
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	validator, delSharesCreated := keeper.AddTokensFromDel(ctx, validator, sdk.NewInt(100))
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.Int64())
	TestingUpdateValidator(keeper, ctx, validator, true)
	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, int64(100), validator.Tokens.Int64())

	power := types.GetValidatorsByPowerIndexKey(validator)
	require.True(t, validatorByPowerIndexExists(keeper, ctx, power))

	// burn half the delegator shares
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	validator, burned := keeper.removeDelShares(ctx, validator, delSharesCreated.Quo(sdk.NewDec(2)))
	require.Equal(t, int64(50), burned.Int64())
	TestingUpdateValidator(keeper, ctx, validator, true) // update the validator, possibly kicking it out
	require.False(t, validatorByPowerIndexExists(keeper, ctx, power))

	validator, found = keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	power = types.GetValidatorsByPowerIndexKey(validator)
	require.True(t, validatorByPowerIndexExists(keeper, ctx, power))
}

func TestUpdateBondedValidatorsDecreaseCliff(t *testing.T) {
	numVals := 10
	maxVals := 5

	// create context, keeper, and pool for tests
	ctx, _, keeper := CreateTestInput(t, false, 0)
	bondedPool, notBondedPool := keeper.GetPools(ctx)

	// create keeper parameters
	params := keeper.GetParams(ctx)
	params.MaxValidators = uint16(maxVals)
	keeper.SetParams(ctx, params)

	// create a random pool
	bondedPool.SetCoins(sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), sdk.TokensFromTendermintPower(1234))))
	notBondedPool.SetCoins(sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), sdk.TokensFromTendermintPower(10000))))
	keeper.SetBondedPool(ctx, bondedPool)
	keeper.SetNotBondedPool(ctx, notBondedPool)

	validators := make([]types.Validator, numVals)
	for i := 0; i < len(validators); i++ {
		moniker := fmt.Sprintf("val#%d", int64(i))
		val := types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{Moniker: moniker})
		delTokens := sdk.TokensFromTendermintPower(int64((i + 1) * 10))
		val, _ = keeper.AddTokensFromDel(ctx, val, delTokens)

		val = TestingUpdateValidator(keeper, ctx, val, true)
		validators[i] = val
	}

	nextCliffVal := validators[numVals-maxVals+1]

	// remove enough tokens to kick out the validator below the current cliff
	// validator and next in line cliff validator
	keeper.DeleteValidatorByPowerIndex(ctx, nextCliffVal)
	shares := sdk.TokensFromTendermintPower(21)
	nextCliffVal, _ = keeper.removeDelShares(ctx, nextCliffVal, shares.ToDec())
	nextCliffVal = TestingUpdateValidator(keeper, ctx, nextCliffVal, true)

	expectedValStatus := map[int]sdk.BondStatus{
		9: sdk.Bonded, 8: sdk.Bonded, 7: sdk.Bonded, 5: sdk.Bonded, 4: sdk.Bonded,
		0: sdk.Unbonding, 1: sdk.Unbonding, 2: sdk.Unbonding, 3: sdk.Unbonding, 6: sdk.Unbonding,
	}

	// require all the validators have their respective statuses
	for valIdx, status := range expectedValStatus {
		valAddr := validators[valIdx].OperatorAddress
		val, _ := keeper.GetValidator(ctx, valAddr)

		assert.Equal(
			t, status, val.GetStatus(),
			fmt.Sprintf("expected validator at index %v to have status: %s", valIdx, status),
		)
	}
}

func TestSlashToZeroPowerRemoved(t *testing.T) {
	// initialize setup
	ctx, _, keeper := CreateTestInput(t, false, 100)

	// add a validator
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	valTokens := sdk.TokensFromTendermintPower(100)
	validator, _ = keeper.AddTokensFromDel(ctx, validator, valTokens)
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, valTokens, validator.Tokens)
	keeper.SetValidatorByConsAddr(ctx, validator)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	require.Equal(t, valTokens, validator.Tokens, "\nvalidator %v\npool %v", validator, valTokens)

	// slash the validator by 100%
	consAddr0 := sdk.ConsAddress(PKs[0].Address())
	keeper.Slash(ctx, consAddr0, 0, 100, sdk.OneDec())
	// apply TM updates
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	// validator should be unbonding
	validator, _ = keeper.GetValidator(ctx, addrVals[0])
	require.Equal(t, validator.GetStatus(), sdk.Unbonding)
}

// This function tests UpdateValidator, GetValidator, GetLastValidators, RemoveValidator
func TestValidatorBasics(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	//construct the validators
	var validators [3]types.Validator
	powers := []int64{9, 8, 7}
	for i, power := range powers {
		validators[i] = types.NewValidator(addrVals[i], PKs[i], types.Description{})
		validators[i].Status = sdk.Unbonded
		validators[i].Tokens = sdk.ZeroInt()
		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)
	}
	assert.Equal(t, sdk.TokensFromTendermintPower(9), validators[0].Tokens)
	assert.Equal(t, sdk.TokensFromTendermintPower(8), validators[1].Tokens)
	assert.Equal(t, sdk.TokensFromTendermintPower(7), validators[2].Tokens)

	// check the empty keeper first
	_, found := keeper.GetValidator(ctx, addrVals[0])
	require.False(t, found)
	resVals := keeper.GetLastValidators(ctx)
	require.Zero(t, len(resVals))

	resVals = keeper.GetValidators(ctx, 2)
	require.Zero(t, len(resVals))

	bondedPool, _ := keeper.GetPools(ctx)
	assert.True(sdk.IntEq(t, sdk.ZeroInt(), bondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx))))

	// set and retrieve a record
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], true)
	keeper.SetValidatorByConsAddr(ctx, validators[0])
	resVal, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	assert.True(ValEq(t, validators[0], resVal))

	// retrieve from consensus
	resVal, found = keeper.GetValidatorByConsAddr(ctx, sdk.ConsAddress(PKs[0].Address()))
	require.True(t, found)
	assert.True(ValEq(t, validators[0], resVal))
	resVal, found = keeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(PKs[0]))
	require.True(t, found)
	assert.True(ValEq(t, validators[0], resVal))

	resVals = keeper.GetLastValidators(ctx)
	require.Equal(t, 1, len(resVals))
	assert.True(ValEq(t, validators[0], resVals[0]))
	assert.Equal(t, sdk.Bonded, validators[0].Status)
	assert.True(sdk.IntEq(t, sdk.TokensFromTendermintPower(9), validators[0].BondedTokens()))

	bondedPool, _ = keeper.GetPools(ctx)
	assert.True(sdk.IntEq(t, bondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)), validators[0].BondedTokens()))

	// modify a records, save, and retrieve
	validators[0].Status = sdk.Bonded
	validators[0].Tokens = sdk.TokensFromTendermintPower(10)
	validators[0].DelegatorShares = validators[0].Tokens.ToDec()
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], true)
	resVal, found = keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	assert.True(ValEq(t, validators[0], resVal))

	resVals = keeper.GetLastValidators(ctx)
	require.Equal(t, 1, len(resVals))
	assert.True(ValEq(t, validators[0], resVals[0]))

	// add other validators
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], true)
	validators[2] = TestingUpdateValidator(keeper, ctx, validators[2], true)
	resVal, found = keeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	assert.True(ValEq(t, validators[1], resVal))
	resVal, found = keeper.GetValidator(ctx, addrVals[2])
	require.True(t, found)
	assert.True(ValEq(t, validators[2], resVal))

	resVals = keeper.GetLastValidators(ctx)
	require.Equal(t, 3, len(resVals))
	assert.True(ValEq(t, validators[0], resVals[0])) // order doesn't matter here
	assert.True(ValEq(t, validators[1], resVals[1]))
	assert.True(ValEq(t, validators[2], resVals[2]))

	// remove a record

	// shouldn't be able to remove if status is not unbonded
	assert.PanicsWithValue(t,
		"cannot call RemoveValidator on bonded or unbonding validators",
		func() { keeper.RemoveValidator(ctx, validators[1].OperatorAddress) })

	// shouldn't be able to remove if there are still tokens left
	validators[1].Status = sdk.Unbonded
	keeper.SetValidator(ctx, validators[1])
	assert.PanicsWithValue(t,
		"attempting to remove a validator which still contains tokens",
		func() { keeper.RemoveValidator(ctx, validators[1].OperatorAddress) })

	validators[1].Tokens = sdk.ZeroInt()                       // ...remove all tokens
	keeper.SetValidator(ctx, validators[1])                    // ...set the validator
	keeper.RemoveValidator(ctx, validators[1].OperatorAddress) // Now it can be removed.
	_, found = keeper.GetValidator(ctx, addrVals[1])
	require.False(t, found)
}

// test how the validators are sorted, tests GetBondedValidatorsByPower
func GetValidatorSortingUnmixed(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	// initialize some validators into the state
	amts := []int64{0, 100, 1, 400, 200}
	n := len(amts)
	var validators [5]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		validators[i].Status = sdk.Bonded
		validators[i].Tokens = sdk.NewInt(amt)
		validators[i].DelegatorShares = sdk.NewDec(amt)
		TestingUpdateValidator(keeper, ctx, validators[i], true)
	}

	// first make sure everything made it in to the gotValidator group
	resValidators := keeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, n, len(resValidators))
	assert.Equal(t, sdk.NewInt(400), resValidators[0].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(200), resValidators[1].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(100), resValidators[2].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(1), resValidators[3].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(0), resValidators[4].BondedTokens(), "%v", resValidators)
	assert.Equal(t, validators[3].OperatorAddress, resValidators[0].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[4].OperatorAddress, resValidators[1].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[1].OperatorAddress, resValidators[2].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[2].OperatorAddress, resValidators[3].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[0].OperatorAddress, resValidators[4].OperatorAddress, "%v", resValidators)

	// test a basic increase in voting power
	validators[3].Tokens = sdk.NewInt(500)
	TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))

	// test a decrease in voting power
	validators[3].Tokens = sdk.NewInt(300)
	TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))

	// test equal voting power, different age
	validators[3].Tokens = sdk.NewInt(200)
	ctx = ctx.WithBlockHeight(10)
	TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))

	// no change in voting power - no change in sort
	ctx = ctx.WithBlockHeight(20)
	TestingUpdateValidator(keeper, ctx, validators[4], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))

	// change in voting power of both validators, both still in v-set, no age change
	validators[3].Tokens = sdk.NewInt(300)
	validators[4].Tokens = sdk.NewInt(300)
	TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	ctx = ctx.WithBlockHeight(30)
	TestingUpdateValidator(keeper, ctx, validators[4], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n, "%v", resValidators)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))
}

func GetValidatorSortingMixed(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	// now 2 max resValidators
	params := keeper.GetParams(ctx)
	params.MaxValidators = 2
	keeper.SetParams(ctx, params)

	// initialize some validators into the state
	amts := []int64{0, 100, 1, 400, 200}

	n := len(amts)
	var validators [5]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		validators[i].DelegatorShares = sdk.NewDec(amt)
	}

	validators[0].Status = sdk.Bonded
	validators[1].Status = sdk.Bonded
	validators[2].Status = sdk.Bonded
	validators[0].Tokens = sdk.NewInt(amts[0])
	validators[1].Tokens = sdk.NewInt(amts[1])
	validators[2].Tokens = sdk.NewInt(amts[2])

	validators[3].Status = sdk.Bonded
	validators[4].Status = sdk.Bonded
	validators[3].Tokens = sdk.NewInt(amts[3])
	validators[4].Tokens = sdk.NewInt(amts[4])

	for i := range amts {
		TestingUpdateValidator(keeper, ctx, validators[i], true)
	}
	val0, found := keeper.GetValidator(ctx, sdk.ValAddress(sdk.ValAddress(PKs[0].Address().Bytes())))
	require.True(t, found)
	val1, found := keeper.GetValidator(ctx, sdk.ValAddress(Addrs[1]))
	require.True(t, found)
	val2, found := keeper.GetValidator(ctx, sdk.ValAddress(Addrs[2]))
	require.True(t, found)
	val3, found := keeper.GetValidator(ctx, sdk.ValAddress(Addrs[3]))
	require.True(t, found)
	val4, found := keeper.GetValidator(ctx, sdk.ValAddress(Addrs[4]))
	require.True(t, found)
	require.Equal(t, sdk.Unbonded, val0.Status)
	require.Equal(t, sdk.Unbonded, val1.Status)
	require.Equal(t, sdk.Unbonded, val2.Status)
	require.Equal(t, sdk.Bonded, val3.Status)
	require.Equal(t, sdk.Bonded, val4.Status)

	// first make sure everything made it in to the gotValidator group
	resValidators := keeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, n, len(resValidators))
	assert.Equal(t, sdk.NewInt(400), resValidators[0].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(200), resValidators[1].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(100), resValidators[2].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(1), resValidators[3].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewInt(0), resValidators[4].BondedTokens(), "%v", resValidators)
	assert.Equal(t, validators[3].OperatorAddress, resValidators[0].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[4].OperatorAddress, resValidators[1].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[1].OperatorAddress, resValidators[2].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[2].OperatorAddress, resValidators[3].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[0].OperatorAddress, resValidators[4].OperatorAddress, "%v", resValidators)
}

// TODO separate out into multiple tests
func TestGetValidatorsEdgeCases(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)
	var found bool

	// now 2 max resValidators
	params := keeper.GetParams(ctx)
	nMax := uint16(2)
	params.MaxValidators = nMax
	keeper.SetParams(ctx, params)

	// initialize some validators into the state
	powers := []int64{0, 100, 400, 400}
	var validators [4]types.Validator
	for i, power := range powers {
		moniker := fmt.Sprintf("val#%d", int64(i))
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{Moniker: moniker})
		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)
		validators[i] = TestingUpdateValidator(keeper, ctx, validators[i], true)
	}

	for i := range powers {
		validators[i], found = keeper.GetValidator(ctx, validators[i].OperatorAddress)
		require.True(t, found)
	}
	resValidators := keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[2], resValidators[0]))
	assert.True(ValEq(t, validators[3], resValidators[1]))

	keeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	delTokens := sdk.TokensFromTendermintPower(500)
	validators[0], _ = keeper.AddTokensFromDel(ctx, validators[0], delTokens)
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))

	// A validator which leaves the gotValidator set due to a decrease in voting power,
	// then increases to the original voting power, does not get its spot back in the
	// case of a tie.

	// validator 3 enters bonded validator set
	ctx = ctx.WithBlockHeight(40)

	validators[3], found = keeper.GetValidator(ctx, validators[3].OperatorAddress)
	require.True(t, found)
	keeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	validators[3], _ = keeper.AddTokensFromDel(ctx, validators[3], sdk.NewInt(1))
	validators[3] = TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[3], resValidators[1]))

	// validator 3 kicked out temporarily
	keeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	validators[3], _ = keeper.removeDelShares(ctx, validators[3], sdk.NewDec(201))
	validators[3] = TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))

	// validator 4 does not get spot back
	keeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	validators[3], _ = keeper.AddTokensFromDel(ctx, validators[3], sdk.NewInt(200))
	validators[3] = TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))
	_, exists := keeper.GetValidator(ctx, validators[3].OperatorAddress)
	require.True(t, exists)
}

func TestValidatorBondHeight(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	// now 2 max resValidators
	params := keeper.GetParams(ctx)
	params.MaxValidators = 2
	keeper.SetParams(ctx, params)

	// initialize some validators into the state
	var validators [3]types.Validator
	validators[0] = types.NewValidator(sdk.ValAddress(sdk.ValAddress(PKs[0].Address().Bytes())), PKs[0], types.Description{})
	validators[1] = types.NewValidator(sdk.ValAddress(Addrs[1]), PKs[1], types.Description{})
	validators[2] = types.NewValidator(sdk.ValAddress(Addrs[2]), PKs[2], types.Description{})

	tokens0 := sdk.TokensFromTendermintPower(200)
	tokens1 := sdk.TokensFromTendermintPower(100)
	tokens2 := sdk.TokensFromTendermintPower(100)
	validators[0], _ = keeper.AddTokensFromDel(ctx, validators[0], tokens0)
	validators[1], _ = keeper.AddTokensFromDel(ctx, validators[1], tokens1)
	validators[2], _ = keeper.AddTokensFromDel(ctx, validators[2], tokens2)

	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], true)

	////////////////////////////////////////
	// If two validators both increase to the same voting power in the same block,
	// the one with the first transaction should become bonded
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], true)
	validators[2] = TestingUpdateValidator(keeper, ctx, validators[2], true)

	resValidators := keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, uint16(len(resValidators)), params.MaxValidators)

	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[1], resValidators[1]))
	keeper.DeleteValidatorByPowerIndex(ctx, validators[1])
	keeper.DeleteValidatorByPowerIndex(ctx, validators[2])
	delTokens := sdk.TokensFromTendermintPower(50)
	validators[1], _ = keeper.AddTokensFromDel(ctx, validators[1], delTokens)
	validators[2], _ = keeper.AddTokensFromDel(ctx, validators[2], delTokens)
	validators[2] = TestingUpdateValidator(keeper, ctx, validators[2], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, params.MaxValidators, uint16(len(resValidators)))
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], true)
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))
}

func TestFullValidatorSetPowerChange(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)
	params := keeper.GetParams(ctx)
	max := 2
	params.MaxValidators = uint16(2)
	keeper.SetParams(ctx, params)

	// initialize some validators into the state
	powers := []int64{0, 100, 400, 400, 200}
	var validators [5]types.Validator
	for i, power := range powers {
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)
		TestingUpdateValidator(keeper, ctx, validators[i], true)
	}
	for i := range powers {
		var found bool
		validators[i], found = keeper.GetValidator(ctx, validators[i].OperatorAddress)
		require.True(t, found)
	}
	assert.Equal(t, sdk.Unbonded, validators[0].Status)
	assert.Equal(t, sdk.Unbonding, validators[1].Status)
	assert.Equal(t, sdk.Bonded, validators[2].Status)
	assert.Equal(t, sdk.Bonded, validators[3].Status)
	assert.Equal(t, sdk.Unbonded, validators[4].Status)
	resValidators := keeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, max, len(resValidators))
	assert.True(ValEq(t, validators[2], resValidators[0])) // in the order of txs
	assert.True(ValEq(t, validators[3], resValidators[1]))

	// test a swap in voting power

	tokens := sdk.TokensFromTendermintPower(600)
	validators[0], _ = keeper.AddTokensFromDel(ctx, validators[0], tokens)
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, max, len(resValidators))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))
}

func TestApplyAndReturnValidatorSetUpdatesAllNone(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = types.NewValidator(valAddr, valPubKey, types.Description{})
		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)
	}

	// test from nothing to something
	//  tendermintUpdate set: {} -> {c1, c3}
	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))
	keeper.SetValidator(ctx, validators[0])
	keeper.SetValidatorByPowerIndex(ctx, validators[0])
	keeper.SetValidator(ctx, validators[1])
	keeper.SetValidatorByPowerIndex(ctx, validators[1])

	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	assert.Equal(t, 2, len(updates))
	validators[0], _ = keeper.GetValidator(ctx, validators[0].OperatorAddress)
	validators[1], _ = keeper.GetValidator(ctx, validators[1].OperatorAddress)
	assert.Equal(t, validators[0].ABCIValidatorUpdate(), updates[1])
	assert.Equal(t, validators[1].ABCIValidatorUpdate(), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesIdentical(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})

		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)

	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test identical,
	//  tendermintUpdate set: {} -> {}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))
}

func TestApplyAndReturnValidatorSetUpdatesSingleValueChange(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {

		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})

		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)

	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test single value change
	//  tendermintUpdate set: {} -> {c1'}
	validators[0].Status = sdk.Bonded
	validators[0].Tokens = sdk.TokensFromTendermintPower(600)
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)

	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesMultipleValueChange(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {

		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})

		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)

	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test multiple value change
	//  tendermintUpdate set: {c1, c3} -> {c1', c3'}
	delTokens1 := sdk.TokensFromTendermintPower(190)
	delTokens2 := sdk.TokensFromTendermintPower(80)
	validators[0], _ = keeper.AddTokensFromDel(ctx, validators[0], delTokens1)
	validators[1], _ = keeper.AddTokensFromDel(ctx, validators[1], delTokens2)
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)

	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[1])
}

func TestApplyAndReturnValidatorSetUpdatesInserted(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	powers := []int64{10, 20, 5, 15, 25}
	var validators [5]types.Validator
	for i, power := range powers {

		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})

		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)

	}

	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test validtor added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	keeper.SetValidator(ctx, validators[2])
	keeper.SetValidatorByPowerIndex(ctx, validators[2])
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validators[2], _ = keeper.GetValidator(ctx, validators[2].OperatorAddress)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[2].ABCIValidatorUpdate(), updates[0])

	// test validtor added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	keeper.SetValidator(ctx, validators[3])
	keeper.SetValidatorByPowerIndex(ctx, validators[3])
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validators[3], _ = keeper.GetValidator(ctx, validators[3].OperatorAddress)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[3].ABCIValidatorUpdate(), updates[0])

	// test validtor added at the end
	//  tendermintUpdate set: {} -> {c0}
	keeper.SetValidator(ctx, validators[4])
	keeper.SetValidatorByPowerIndex(ctx, validators[4])
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validators[4], _ = keeper.GetValidator(ctx, validators[4].OperatorAddress)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[4].ABCIValidatorUpdate(), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesWithCliffValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)
	params := types.DefaultParams()
	params.MaxValidators = 2
	keeper.SetParams(ctx, params)

	powers := []int64{10, 20, 5}
	var validators [5]types.Validator
	for i, power := range powers {

		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})

		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)

	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test validator added at the end but not inserted in the valset
	//  tendermintUpdate set: {} -> {}
	TestingUpdateValidator(keeper, ctx, validators[2], false)
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 0, len(updates))

	// test validator change its power and become a gotValidator (pushing out an existing)
	//  tendermintUpdate set: {}     -> {c0, c4}
	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	tokens := sdk.TokensFromTendermintPower(10)
	validators[2], _ = keeper.AddTokensFromDel(ctx, validators[2], tokens)
	keeper.SetValidator(ctx, validators[2])
	keeper.SetValidatorByPowerIndex(ctx, validators[2])
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validators[2], _ = keeper.GetValidator(ctx, validators[2].OperatorAddress)
	require.Equal(t, 2, len(updates), "%v", updates)
	require.Equal(t, validators[0].ABCIValidatorUpdateZero(), updates[1])
	require.Equal(t, validators[2].ABCIValidatorUpdate(), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesPowerDecrease(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	powers := []int64{100, 100}
	var validators [2]types.Validator
	for i, power := range powers {

		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})

		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)

	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// check initial power
	require.Equal(t, int64(100), validators[0].GetTendermintPower())
	require.Equal(t, int64(100), validators[1].GetTendermintPower())

	// test multiple value change
	//  tendermintUpdate set: {c1, c3} -> {c1', c3'}
	delTokens1 := sdk.TokensFromTendermintPower(20)
	delTokens2 := sdk.TokensFromTendermintPower(30)
	validators[0], _ = keeper.removeDelShares(ctx, validators[0], delTokens1.ToDec())
	validators[1], _ = keeper.removeDelShares(ctx, validators[1], delTokens2.ToDec())
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)

	// power has changed
	require.Equal(t, int64(80), validators[0].GetTendermintPower())
	require.Equal(t, int64(70), validators[1].GetTendermintPower())

	// Tendermint updates should reflect power change
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[1])
}

func TestApplyAndReturnValidatorSetUpdatesNewValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)
	params := keeper.GetParams(ctx)
	params.MaxValidators = uint16(3)

	keeper.SetParams(ctx, params)

	powers := []int64{100, 100}
	var validators [2]types.Validator

	// initialize some validators into the state
	for i, power := range powers {

		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = types.NewValidator(valAddr, valPubKey, types.Description{})
		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)

		keeper.SetValidator(ctx, validators[i])
		keeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// verify initial Tendermint updates are correct
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, len(validators), len(updates))
	validators[0], _ = keeper.GetValidator(ctx, validators[0].OperatorAddress)
	validators[1], _ = keeper.GetValidator(ctx, validators[1].OperatorAddress)
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[1])

	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// update initial validator set
	for i, power := range powers {

		keeper.DeleteValidatorByPowerIndex(ctx, validators[i])
		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)

		keeper.SetValidator(ctx, validators[i])
		keeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// add a new validator that goes from zero power, to non-zero power, back to
	// zero power
	valPubKey := PKs[len(validators)+1]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	amt := sdk.NewInt(100)

	validator := types.NewValidator(valAddr, valPubKey, types.Description{})
	validator, _ = keeper.AddTokensFromDel(ctx, validator, amt)

	keeper.SetValidator(ctx, validator)

	validator, _ = keeper.removeDelShares(ctx, validator, amt.ToDec())
	keeper.SetValidator(ctx, validator)
	keeper.SetValidatorByPowerIndex(ctx, validator)

	// add a new validator that increases in power
	valPubKey = PKs[len(validators)+2]
	valAddr = sdk.ValAddress(valPubKey.Address().Bytes())

	validator = types.NewValidator(valAddr, valPubKey, types.Description{})
	tokens := sdk.TokensFromTendermintPower(500)
	validator, _ = keeper.AddTokensFromDel(ctx, validator, tokens)
	keeper.SetValidator(ctx, validator)
	keeper.SetValidatorByPowerIndex(ctx, validator)

	// verify initial Tendermint updates are correct
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validator, _ = keeper.GetValidator(ctx, validator.OperatorAddress)
	validators[0], _ = keeper.GetValidator(ctx, validators[0].OperatorAddress)
	validators[1], _ = keeper.GetValidator(ctx, validators[1].OperatorAddress)
	require.Equal(t, len(validators)+1, len(updates))
	require.Equal(t, validator.ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[1])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[2])
}

func TestApplyAndReturnValidatorSetUpdatesBondTransition(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)
	params := keeper.GetParams(ctx)
	params.MaxValidators = uint16(2)

	keeper.SetParams(ctx, params)

	powers := []int64{100, 200, 300}
	var validators [3]types.Validator

	// initialize some validators into the state
	for i, power := range powers {
		moniker := fmt.Sprintf("%d", i)
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = types.NewValidator(valAddr, valPubKey, types.Description{Moniker: moniker})
		tokens := sdk.TokensFromTendermintPower(power)
		validators[i], _ = keeper.AddTokensFromDel(ctx, validators[i], tokens)
		keeper.SetValidator(ctx, validators[i])
		keeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// verify initial Tendermint updates are correct
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))
	validators[2], _ = keeper.GetValidator(ctx, validators[2].OperatorAddress)
	validators[1], _ = keeper.GetValidator(ctx, validators[1].OperatorAddress)
	require.Equal(t, validators[2].ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[1])

	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// delegate to validator with lowest power but not enough to bond
	ctx = ctx.WithBlockHeight(1)

	var found bool
	validators[0], found = keeper.GetValidator(ctx, validators[0].OperatorAddress)
	require.True(t, found)

	keeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	tokens := sdk.TokensFromTendermintPower(1)
	validators[0], _ = keeper.AddTokensFromDel(ctx, validators[0], tokens)
	keeper.SetValidator(ctx, validators[0])
	keeper.SetValidatorByPowerIndex(ctx, validators[0])

	// verify initial Tendermint updates are correct
	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// create a series of events that will bond and unbond the validator with
	// lowest power in a single block context (height)
	ctx = ctx.WithBlockHeight(2)

	validators[1], found = keeper.GetValidator(ctx, validators[1].OperatorAddress)
	require.True(t, found)

	keeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	validators[0], _ = keeper.removeDelShares(ctx, validators[0], validators[0].DelegatorShares)
	keeper.SetValidator(ctx, validators[0])
	keeper.SetValidatorByPowerIndex(ctx, validators[0])
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 0, len(updates))

	keeper.DeleteValidatorByPowerIndex(ctx, validators[1])
	tokens = sdk.TokensFromTendermintPower(250)
	validators[1], _ = keeper.AddTokensFromDel(ctx, validators[1], tokens)
	keeper.SetValidator(ctx, validators[1])
	keeper.SetValidatorByPowerIndex(ctx, validators[1])

	// verify initial Tendermint updates are correct
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[0])

	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))
}

func TestUpdateValidatorCommission(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Now().UTC()})

	commission1 := types.NewCommissionWithTime(
		sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(3, 1),
		sdk.NewDecWithPrec(1, 1), time.Now().UTC().Add(time.Duration(-1)*time.Hour),
	)
	commission2 := types.NewCommission(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(3, 1), sdk.NewDecWithPrec(1, 1))

	val1 := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	val2 := types.NewValidator(addrVals[1], PKs[1], types.Description{})

	val1, _ = val1.SetInitialCommission(commission1)
	val2, _ = val2.SetInitialCommission(commission2)

	keeper.SetValidator(ctx, val1)
	keeper.SetValidator(ctx, val2)

	testCases := []struct {
		validator   types.Validator
		newRate     sdk.Dec
		expectedErr bool
	}{
		{val1, sdk.ZeroDec(), true},
		{val2, sdk.NewDecWithPrec(-1, 1), true},
		{val2, sdk.NewDecWithPrec(4, 1), true},
		{val2, sdk.NewDecWithPrec(3, 1), true},
		{val2, sdk.NewDecWithPrec(2, 1), false},
	}

	for i, tc := range testCases {
		commission, err := keeper.UpdateValidatorCommission(ctx, tc.validator, tc.newRate)

		if tc.expectedErr {
			require.Error(t, err, "expected error for test case #%d with rate: %s", i, tc.newRate)
		} else {
			tc.validator.Commission = commission
			keeper.SetValidator(ctx, tc.validator)
			val, found := keeper.GetValidator(ctx, tc.validator.OperatorAddress)

			require.True(t, found,
				"expected to find validator for test case #%d with rate: %s", i, tc.newRate,
			)
			require.NoError(t, err,
				"unexpected error for test case #%d with rate: %s", i, tc.newRate,
			)
			require.Equal(t, tc.newRate, val.Commission.Rate,
				"expected new validator commission rate for test case #%d with rate: %s", i, tc.newRate,
			)
			require.Equal(t, ctx.BlockHeader().Time, val.Commission.UpdateTime,
				"expected new validator commission update time for test case #%d with rate: %s", i, tc.newRate,
			)
		}
	}
}

func TestRemoveTokens(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

	validator := types.Validator{
		OperatorAddress: valAddr,
		ConsPubKey:      valPubKey,
		Status:          sdk.Bonded,
		Tokens:          sdk.NewInt(100),
		DelegatorShares: sdk.NewDec(100),
	}

	bondedPool, notBondedPool := keeper.GetPools(ctx)
	err := bondedPool.SetCoins(sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), validator.BondedTokens())))
	require.NoError(t, err)
	err = notBondedPool.SetCoins(sdk.NewCoins())
	require.NoError(t, err)
	keeper.SetBondedPool(ctx, bondedPool)
	keeper.SetNotBondedPool(ctx, notBondedPool)

	// remove tokens and test check everything
	validator = keeper.removeTokens(ctx, validator, sdk.NewInt(10))
	bondedPool, notBondedPool = keeper.GetPools(ctx)

	require.Equal(t, int64(90), validator.Tokens.Int64())
	require.Equal(t, int64(90), bondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())
	require.Equal(t, int64(0), notBondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())

	// update validator to from bonded -> unbonded and remove some more tokens
	validator = keeper.UpdateStatus(ctx, validator, sdk.Unbonded)
	bondedPool, notBondedPool = keeper.GetPools(ctx)
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, int64(0), bondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())
	require.Equal(t, int64(90), notBondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())

	validator = keeper.removeTokens(ctx, validator, sdk.NewInt(10))
	bondedPool, notBondedPool = keeper.GetPools(ctx)
	require.Equal(t, int64(80), validator.Tokens.Int64())
	require.Equal(t, int64(0), bondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())
	require.Equal(t, int64(80), notBondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())

	require.Panics(t, func() { keeper.removeTokens(ctx, validator, sdk.NewInt(-1)) })
	require.Panics(t, func() { keeper.removeTokens(ctx, validator, sdk.NewInt(100)) })
}

func TestAddTokensValidatorBonded(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	validator := types.NewValidator(sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0], types.Description{})
	validator = keeper.UpdateStatus(ctx, validator, sdk.Bonded)
	validator, delShares := keeper.AddTokensFromDel(ctx, validator, sdk.NewInt(10))

	assert.True(sdk.DecEq(t, sdk.NewDec(10), delShares))
	assert.True(sdk.IntEq(t, sdk.NewInt(10), validator.BondedTokens()))
	assert.True(sdk.DecEq(t, sdk.NewDec(10), validator.DelegatorShares))
}

func TestAddTokensValidatorUnbonding(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)
	validator := types.NewValidator(sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0], types.Description{})
	validator = keeper.UpdateStatus(ctx, validator, sdk.Unbonding)
	validator, delShares := keeper.AddTokensFromDel(ctx, validator, sdk.NewInt(10))

	assert.True(sdk.DecEq(t, sdk.NewDec(10), delShares))
	assert.Equal(t, sdk.Unbonding, validator.Status)
	assert.True(sdk.IntEq(t, sdk.NewInt(10), validator.Tokens))
	assert.True(sdk.DecEq(t, sdk.NewDec(10), validator.DelegatorShares))
}

func TestAddTokensValidatorUnbonded(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	validator := types.NewValidator(sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0], types.Description{})
	validator = keeper.UpdateStatus(ctx, validator, sdk.Unbonded)
	validator, delShares := keeper.AddTokensFromDel(ctx, validator, sdk.NewInt(10))

	assert.True(sdk.DecEq(t, sdk.NewDec(10), delShares))
	assert.Equal(t, sdk.Unbonded, validator.Status)
	assert.True(sdk.IntEq(t, sdk.NewInt(10), validator.Tokens))
	assert.True(sdk.DecEq(t, sdk.NewDec(10), validator.DelegatorShares))
}

// TODO refactor to make simpler like the AddToken tests above
func TestRemoveDelShares(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	valA := types.Validator{
		OperatorAddress: sdk.ValAddress(PKs[0].Address().Bytes()),
		ConsPubKey:      PKs[0],
		Status:          sdk.Bonded,
		Tokens:          sdk.NewInt(100),
		DelegatorShares: sdk.NewDec(100),
	}

	bondedPool, notBondedPool := keeper.GetPools(ctx)
	notBondedPool.SetCoins(sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), sdk.NewInt(10))))
	bondedPool.SetCoins(sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), valA.BondedTokens())))
	keeper.SetBondedPool(ctx, bondedPool)
	keeper.SetNotBondedPool(ctx, notBondedPool)

	expectedTotal := bondedPool.GetCoins().Add(notBondedPool.GetCoins())

	// Remove delegator shares
	valB, coinsB := keeper.removeDelShares(ctx, valA, sdk.NewDec(10))
	require.Equal(t, int64(10), coinsB.Int64())
	require.Equal(t, int64(90), valB.DelegatorShares.RoundInt64())
	require.Equal(t, int64(90), valB.BondedTokens().Int64())

	bondedPool, notBondedPool = keeper.GetPools(ctx)
	require.Equal(t, int64(90), bondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())
	require.Equal(t, int64(20), notBondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())

	newTotal := bondedPool.GetCoins().Add(notBondedPool.GetCoins())
	// conservation of tokens
	require.True(t, expectedTotal.IsEqual(newTotal))

	// specific case from random tests
	poolTokens := sdk.NewInt(5102)
	delShares := sdk.NewDec(115)
	validator := types.Validator{
		OperatorAddress: sdk.ValAddress(PKs[0].Address().Bytes()),
		ConsPubKey:      PKs[0],
		Status:          sdk.Bonded,
		Tokens:          poolTokens,
		DelegatorShares: delShares,
	}

	bondedPool.SetCoins(sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), sdk.NewInt(248305))))
	notBondedPool.SetCoins(sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), sdk.NewInt(232147))))
	keeper.SetBondedPool(ctx, bondedPool)
	keeper.SetNotBondedPool(ctx, notBondedPool)

	shares := sdk.NewDec(29)
	_, tokens := keeper.removeDelShares(ctx, validator, shares)

	require.True(sdk.IntEq(t, sdk.NewInt(1286), tokens))

	newBondedPool, newUnbondedPool := keeper.GetPools(ctx)

	require.True(t,
		bondedPool.GetCoins().Add(notBondedPool.GetCoins()).IsEqual(
			newBondedPool.GetCoins().Add(newUnbondedPool.GetCoins()),
		))
}

func TestAddTokensFromDel(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	val := types.NewValidator(sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0], types.Description{})

	bondedPool, notBondedPool := keeper.GetPools(ctx)
	notBondedPool.SetCoins(sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), sdk.NewInt(10))))
	keeper.SetNotBondedPool(ctx, notBondedPool)

	val, shares := keeper.AddTokensFromDel(ctx, val, sdk.NewInt(6))
	require.True(sdk.DecEq(t, sdk.NewDec(6), shares))
	require.True(sdk.DecEq(t, sdk.NewDec(6), val.DelegatorShares))
	require.True(sdk.IntEq(t, sdk.NewInt(6), val.Tokens))

	bondedPool, notBondedPool = keeper.GetPools(ctx)
	require.True(sdk.IntEq(t, sdk.NewInt(0), bondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx))))
	require.True(sdk.IntEq(t, sdk.NewInt(10), notBondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx))))

	val, shares = keeper.AddTokensFromDel(ctx, val, sdk.NewInt(3))
	require.True(sdk.DecEq(t, sdk.NewDec(3), shares))
	require.True(sdk.DecEq(t, sdk.NewDec(9), val.DelegatorShares))
	require.True(sdk.IntEq(t, sdk.NewInt(9), val.Tokens))

	bondedPool, notBondedPool = keeper.GetPools(ctx)
	require.True(sdk.IntEq(t, sdk.NewInt(0), bondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx))))
	require.True(sdk.IntEq(t, sdk.NewInt(10), notBondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx))))
}

func TestUpdateStatus(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	bondedPool, notBondedPool := keeper.GetPools(ctx)
	notBondedPool.SetCoins(sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), sdk.NewInt(100))))
	bondedPool.SetCoins(sdk.Coins{})
	keeper.SetNotBondedPool(ctx, notBondedPool)
	keeper.SetBondedPool(ctx, bondedPool)

	validator := types.NewValidator(sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0], types.Description{})
	validator, _ = keeper.AddTokensFromDel(ctx, validator, sdk.NewInt(100))
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.Int64())

	// Unbonded to Bonded
	validator = keeper.UpdateStatus(ctx, validator, sdk.Bonded)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.Int64())

	bondedPool, notBondedPool = keeper.GetPools(ctx)
	require.Equal(t, int64(100), bondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())
	require.Equal(t, int64(0), notBondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())

	// Bonded to Unbonding
	validator = keeper.UpdateStatus(ctx, validator, sdk.Unbonding)
	require.Equal(t, sdk.Unbonding, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.Int64())

	bondedPool, notBondedPool = keeper.GetPools(ctx)
	require.Equal(t, int64(0), bondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())
	require.Equal(t, int64(100), notBondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())

	// Unbonding to Bonded
	validator = keeper.UpdateStatus(ctx, validator, sdk.Bonded)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.Int64())

	bondedPool, notBondedPool = keeper.GetPools(ctx)
	require.Equal(t, int64(100), bondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())
	require.Equal(t, int64(0), notBondedPool.GetCoins().AmountOf(keeper.BondDenom(ctx)).Int64())

}

func TestPossibleOverflow(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	delShares := sdk.NewDec(391432570689183511).Quo(sdk.NewDec(40113011844664))
	validator := types.Validator{
		OperatorAddress: sdk.ValAddress(PKs[0].Address().Bytes()),
		ConsPubKey:      PKs[0],
		Status:          sdk.Bonded,
		Tokens:          sdk.NewInt(2159),
		DelegatorShares: delShares,
	}

	newValidator, _ := keeper.AddTokensFromDel(ctx, validator, sdk.NewInt(71))

	require.False(t, newValidator.DelegatorShares.IsNegative())
	require.False(t, newValidator.Tokens.IsNegative())
}
