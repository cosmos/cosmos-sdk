package keeper

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//_______________________________________________________

func TestSetValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 10)
	pool := keeper.GetPool(ctx)

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

	// test how the validator is set from a purely unbonbed pool
	validator := types.NewValidator(valAddr, valPubKey, types.Description{})
	validator, pool, _ = validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, sdk.Unbonded, validator.Status)
	assert.True(sdk.DecEq(t, sdk.NewDec(10), validator.Tokens))
	assert.True(sdk.DecEq(t, sdk.NewDec(10), validator.DelegatorShares))
	keeper.SetPool(ctx, pool)
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
	assert.True(sdk.DecEq(t, sdk.NewDec(10), validator.Tokens))
	assert.True(sdk.DecEq(t, sdk.NewDec(10), validator.DelegatorShares))

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
	pool := keeper.GetPool(ctx)

	// create a random pool
	pool.LooseTokens = sdk.NewDec(10000)
	pool.BondedTokens = sdk.NewDec(1234)
	keeper.SetPool(ctx, pool)

	// add a validator
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	validator, pool, delSharesCreated := validator.AddTokensFromDel(pool, sdk.NewInt(100))
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.RoundInt64())
	keeper.SetPool(ctx, pool)
	TestingUpdateValidator(keeper, ctx, validator, true)
	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, int64(100), validator.Tokens.RoundInt64(), "\nvalidator %v\npool %v", validator, pool)

	pool = keeper.GetPool(ctx)
	power := GetValidatorsByPowerIndexKey(validator)
	require.True(t, validatorByPowerIndexExists(keeper, ctx, power))

	// burn half the delegator shares
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	validator, pool, burned := validator.RemoveDelShares(pool, delSharesCreated.Quo(sdk.NewDec(2)))
	require.Equal(t, int64(50), burned.RoundInt64())
	keeper.SetPool(ctx, pool)                            // update the pool
	TestingUpdateValidator(keeper, ctx, validator, true) // update the validator, possibly kicking it out
	require.False(t, validatorByPowerIndexExists(keeper, ctx, power))

	pool = keeper.GetPool(ctx)
	validator, found = keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	power = GetValidatorsByPowerIndexKey(validator)
	require.True(t, validatorByPowerIndexExists(keeper, ctx, power))
}

func TestUpdateBondedValidatorsDecreaseCliff(t *testing.T) {
	numVals := 10
	maxVals := 5

	// create context, keeper, and pool for tests
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)

	// create keeper parameters
	params := keeper.GetParams(ctx)
	params.MaxValidators = uint16(maxVals)
	keeper.SetParams(ctx, params)

	// create a random pool
	pool.LooseTokens = sdk.NewDec(10000)
	pool.BondedTokens = sdk.NewDec(1234)
	keeper.SetPool(ctx, pool)

	validators := make([]types.Validator, numVals)
	for i := 0; i < len(validators); i++ {
		moniker := fmt.Sprintf("val#%d", int64(i))
		val := types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{Moniker: moniker})
		val.BondHeight = int64(i)
		val, pool, _ = val.AddTokensFromDel(pool, sdk.NewInt(int64((i+1)*10)))

		keeper.SetPool(ctx, pool)
		val = TestingUpdateValidator(keeper, ctx, val, true)
		validators[i] = val
	}

	nextCliffVal := validators[numVals-maxVals+1]

	// remove enough tokens to kick out the validator below the current cliff
	// validator and next in line cliff validator
	keeper.DeleteValidatorByPowerIndex(ctx, nextCliffVal)
	nextCliffVal, pool, _ = nextCliffVal.RemoveDelShares(pool, sdk.NewDec(21))
	keeper.SetPool(ctx, pool)
	nextCliffVal = TestingUpdateValidator(keeper, ctx, nextCliffVal, true)

	expectedValStatus := map[int]sdk.BondStatus{
		9: sdk.Bonded, 8: sdk.Bonded, 7: sdk.Bonded, 5: sdk.Bonded, 4: sdk.Bonded,
		0: sdk.Unbonding, 1: sdk.Unbonding, 2: sdk.Unbonding, 3: sdk.Unbonding, 6: sdk.Unbonding,
	}

	// require all the validators have their respective statuses
	for valIdx, status := range expectedValStatus {
		valAddr := validators[valIdx].OperatorAddr
		val, _ := keeper.GetValidator(ctx, valAddr)

		assert.Equal(
			t, status, val.GetStatus(),
			fmt.Sprintf("expected validator at index %v to have status: %s",
				valIdx,
				sdk.BondStatusToString(status)))
	}
}

func TestSlashToZeroPowerRemoved(t *testing.T) {
	// initialize setup
	ctx, _, keeper := CreateTestInput(t, false, 100)
	pool := keeper.GetPool(ctx)

	// add a validator
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	validator, pool, _ = validator.AddTokensFromDel(pool, sdk.NewInt(100))
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.RoundInt64())
	keeper.SetPool(ctx, pool)
	keeper.SetValidatorByConsAddr(ctx, validator)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	require.Equal(t, int64(100), validator.Tokens.RoundInt64(), "\nvalidator %v\npool %v", validator, pool)

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
	pool := keeper.GetPool(ctx)

	//construct the validators
	var validators [3]types.Validator
	amts := []int64{9, 8, 7}
	for i, amt := range amts {
		validators[i] = types.NewValidator(addrVals[i], PKs[i], types.Description{})
		validators[i].Status = sdk.Unbonded
		validators[i].Tokens = sdk.ZeroDec()
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))
		keeper.SetPool(ctx, pool)
	}
	assert.True(sdk.DecEq(t, sdk.NewDec(9), validators[0].Tokens))
	assert.True(sdk.DecEq(t, sdk.NewDec(8), validators[1].Tokens))
	assert.True(sdk.DecEq(t, sdk.NewDec(7), validators[2].Tokens))

	// check the empty keeper first
	_, found := keeper.GetValidator(ctx, addrVals[0])
	require.False(t, found)
	resVals := keeper.GetLastValidators(ctx)
	require.Zero(t, len(resVals))

	resVals = keeper.GetValidators(ctx, 2)
	require.Zero(t, len(resVals))

	pool = keeper.GetPool(ctx)
	assert.True(sdk.DecEq(t, sdk.ZeroDec(), pool.BondedTokens))

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
	assert.True(sdk.DecEq(t, sdk.NewDec(9), validators[0].BondedTokens()))

	pool = keeper.GetPool(ctx)
	assert.True(sdk.DecEq(t, pool.BondedTokens, validators[0].BondedTokens()))

	// modify a records, save, and retrieve
	validators[0].Status = sdk.Bonded
	validators[0].Tokens = sdk.NewDec(10)
	validators[0].DelegatorShares = sdk.NewDec(10)
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
	validators[1].Status = sdk.Unbonded                     // First must set to Unbonded.
	keeper.SetValidator(ctx, validators[1])                 // ...
	keeper.RemoveValidator(ctx, validators[1].OperatorAddr) // Now it can be removed.
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
		validators[i].Tokens = sdk.NewDec(amt)
		validators[i].DelegatorShares = sdk.NewDec(amt)
		TestingUpdateValidator(keeper, ctx, validators[i], true)
	}

	// first make sure everything made it in to the gotValidator group
	resValidators := keeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, n, len(resValidators))
	assert.Equal(t, sdk.NewDec(400), resValidators[0].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewDec(200), resValidators[1].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewDec(100), resValidators[2].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewDec(1), resValidators[3].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewDec(0), resValidators[4].BondedTokens(), "%v", resValidators)
	assert.Equal(t, validators[3].OperatorAddr, resValidators[0].OperatorAddr, "%v", resValidators)
	assert.Equal(t, validators[4].OperatorAddr, resValidators[1].OperatorAddr, "%v", resValidators)
	assert.Equal(t, validators[1].OperatorAddr, resValidators[2].OperatorAddr, "%v", resValidators)
	assert.Equal(t, validators[2].OperatorAddr, resValidators[3].OperatorAddr, "%v", resValidators)
	assert.Equal(t, validators[0].OperatorAddr, resValidators[4].OperatorAddr, "%v", resValidators)

	// test a basic increase in voting power
	validators[3].Tokens = sdk.NewDec(500)
	TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))

	// test a decrease in voting power
	validators[3].Tokens = sdk.NewDec(300)
	TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))

	// test equal voting power, different age
	validators[3].Tokens = sdk.NewDec(200)
	ctx = ctx.WithBlockHeight(10)
	TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))
	require.Equal(t, int64(0), resValidators[0].BondHeight, "%v", resValidators)
	require.Equal(t, int64(0), resValidators[1].BondHeight, "%v", resValidators)

	// no change in voting power - no change in sort
	ctx = ctx.WithBlockHeight(20)
	TestingUpdateValidator(keeper, ctx, validators[4], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))

	// change in voting power of both validators, both still in v-set, no age change
	validators[3].Tokens = sdk.NewDec(300)
	validators[4].Tokens = sdk.NewDec(300)
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
	validators[0].Tokens = sdk.NewDec(amts[0])
	validators[1].Tokens = sdk.NewDec(amts[1])
	validators[2].Tokens = sdk.NewDec(amts[2])

	validators[3].Status = sdk.Bonded
	validators[4].Status = sdk.Bonded
	validators[3].Tokens = sdk.NewDec(amts[3])
	validators[4].Tokens = sdk.NewDec(amts[4])

	for i := range amts {
		TestingUpdateValidator(keeper, ctx, validators[i], true)
	}
	val0, found := keeper.GetValidator(ctx, sdk.ValAddress(Addrs[0]))
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
	assert.Equal(t, sdk.NewDec(400), resValidators[0].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewDec(200), resValidators[1].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewDec(100), resValidators[2].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewDec(1), resValidators[3].BondedTokens(), "%v", resValidators)
	assert.Equal(t, sdk.NewDec(0), resValidators[4].BondedTokens(), "%v", resValidators)
	assert.Equal(t, validators[3].OperatorAddr, resValidators[0].OperatorAddr, "%v", resValidators)
	assert.Equal(t, validators[4].OperatorAddr, resValidators[1].OperatorAddr, "%v", resValidators)
	assert.Equal(t, validators[1].OperatorAddr, resValidators[2].OperatorAddr, "%v", resValidators)
	assert.Equal(t, validators[2].OperatorAddr, resValidators[3].OperatorAddr, "%v", resValidators)
	assert.Equal(t, validators[0].OperatorAddr, resValidators[4].OperatorAddr, "%v", resValidators)
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
	amts := []int64{0, 100, 400, 400}
	var validators [4]types.Validator
	for i, amt := range amts {
		pool := keeper.GetPool(ctx)
		moniker := fmt.Sprintf("val#%d", int64(i))
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{Moniker: moniker})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))
		keeper.SetPool(ctx, pool)
		validators[i] = TestingUpdateValidator(keeper, ctx, validators[i], true)
	}

	for i := range amts {
		validators[i], found = keeper.GetValidator(ctx, validators[i].OperatorAddr)
		require.True(t, found)
	}
	resValidators := keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[2], resValidators[0]))
	assert.True(ValEq(t, validators[3], resValidators[1]))

	pool := keeper.GetPool(ctx)
	keeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	validators[0], pool, _ = validators[0].AddTokensFromDel(pool, sdk.NewInt(500))
	keeper.SetPool(ctx, pool)
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

	validators[3], found = keeper.GetValidator(ctx, validators[3].OperatorAddr)
	require.True(t, found)
	keeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	validators[3], pool, _ = validators[3].AddTokensFromDel(pool, sdk.NewInt(1))
	keeper.SetPool(ctx, pool)
	validators[3] = TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[3], resValidators[1]))

	// validator 3 kicked out temporarily
	keeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	validators[3], pool, _ = validators[3].RemoveDelShares(pool, sdk.NewDec(201))
	keeper.SetPool(ctx, pool)
	validators[3] = TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))

	// validator 4 does not get spot back
	keeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	validators[3], pool, _ = validators[3].AddTokensFromDel(pool, sdk.NewInt(200))
	keeper.SetPool(ctx, pool)
	validators[3] = TestingUpdateValidator(keeper, ctx, validators[3], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))
	validator, exists := keeper.GetValidator(ctx, validators[3].OperatorAddr)
	require.Equal(t, exists, true)
	require.Equal(t, int64(40), validator.BondHeight)
}

func TestValidatorBondHeight(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)
	pool := keeper.GetPool(ctx)

	// now 2 max resValidators
	params := keeper.GetParams(ctx)
	params.MaxValidators = 2
	keeper.SetParams(ctx, params)

	// initialize some validators into the state
	var validators [3]types.Validator
	validators[0] = types.NewValidator(sdk.ValAddress(Addrs[0]), PKs[0], types.Description{})
	validators[1] = types.NewValidator(sdk.ValAddress(Addrs[1]), PKs[1], types.Description{})
	validators[2] = types.NewValidator(sdk.ValAddress(Addrs[2]), PKs[2], types.Description{})

	validators[0], pool, _ = validators[0].AddTokensFromDel(pool, sdk.NewInt(200))
	validators[1], pool, _ = validators[1].AddTokensFromDel(pool, sdk.NewInt(100))
	validators[2], pool, _ = validators[2].AddTokensFromDel(pool, sdk.NewInt(100))
	keeper.SetPool(ctx, pool)

	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], true)

	////////////////////////////////////////
	// If two validators both increase to the same voting power in the same block,
	// the one with the first transaction should become bonded
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], true)
	validators[2] = TestingUpdateValidator(keeper, ctx, validators[2], true)

	pool = keeper.GetPool(ctx)

	resValidators := keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, uint16(len(resValidators)), params.MaxValidators)

	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[1], resValidators[1]))
	keeper.DeleteValidatorByPowerIndex(ctx, validators[1])
	keeper.DeleteValidatorByPowerIndex(ctx, validators[2])
	validators[1], pool, _ = validators[1].AddTokensFromDel(pool, sdk.NewInt(50))
	validators[2], pool, _ = validators[2].AddTokensFromDel(pool, sdk.NewInt(50))
	keeper.SetPool(ctx, pool)
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
	amts := []int64{0, 100, 400, 400, 200}
	var validators [5]types.Validator
	for i, amt := range amts {
		pool := keeper.GetPool(ctx)
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))
		keeper.SetPool(ctx, pool)
		TestingUpdateValidator(keeper, ctx, validators[i], true)
	}
	for i := range amts {
		var found bool
		validators[i], found = keeper.GetValidator(ctx, validators[i].OperatorAddr)
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
	pool := keeper.GetPool(ctx)
	validators[0], pool, _ = validators[0].AddTokensFromDel(pool, sdk.NewInt(600))
	keeper.SetPool(ctx, pool)
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, max, len(resValidators))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))
}

func TestApplyAndReturnValidatorSetUpdatesAllNone(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	amts := []int64{10, 20}
	var validators [2]types.Validator
	for i, amt := range amts {
		pool := keeper.GetPool(ctx)

		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = types.NewValidator(valAddr, valPubKey, types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))
		keeper.SetPool(ctx, pool)
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
	validators[0], _ = keeper.GetValidator(ctx, validators[0].OperatorAddr)
	validators[1], _ = keeper.GetValidator(ctx, validators[1].OperatorAddr)
	assert.Equal(t, validators[0].ABCIValidatorUpdate(), updates[1])
	assert.Equal(t, validators[1].ABCIValidatorUpdate(), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesIdentical(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	amts := []int64{10, 20}
	var validators [2]types.Validator
	for i, amt := range amts {
		pool := keeper.GetPool(ctx)
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))
		keeper.SetPool(ctx, pool)
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

	amts := []int64{10, 20}
	var validators [2]types.Validator
	for i, amt := range amts {
		pool := keeper.GetPool(ctx)
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))
		keeper.SetPool(ctx, pool)
	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test single value change
	//  tendermintUpdate set: {} -> {c1'}
	validators[0].Status = sdk.Bonded
	validators[0].Tokens = sdk.NewDec(600)
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)

	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesMultipleValueChange(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	amts := []int64{10, 20}
	var validators [2]types.Validator
	for i, amt := range amts {
		pool := keeper.GetPool(ctx)
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))
		keeper.SetPool(ctx, pool)
	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test multiple value change
	//  tendermintUpdate set: {c1, c3} -> {c1', c3'}
	pool := keeper.GetPool(ctx)
	validators[0], pool, _ = validators[0].AddTokensFromDel(pool, sdk.NewInt(190))
	validators[1], pool, _ = validators[1].AddTokensFromDel(pool, sdk.NewInt(80))
	keeper.SetPool(ctx, pool)
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)

	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[1])
}

func TestApplyAndReturnValidatorSetUpdatesInserted(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	amts := []int64{10, 20, 5, 15, 25}
	var validators [5]types.Validator
	for i, amt := range amts {
		pool := keeper.GetPool(ctx)
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))
		keeper.SetPool(ctx, pool)
	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test validtor added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	keeper.SetValidator(ctx, validators[2])
	keeper.SetValidatorByPowerIndex(ctx, validators[2])
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validators[2], _ = keeper.GetValidator(ctx, validators[2].OperatorAddr)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[2].ABCIValidatorUpdate(), updates[0])

	// test validtor added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	keeper.SetValidator(ctx, validators[3])
	keeper.SetValidatorByPowerIndex(ctx, validators[3])
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validators[3], _ = keeper.GetValidator(ctx, validators[3].OperatorAddr)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[3].ABCIValidatorUpdate(), updates[0])

	// test validtor added at the end
	//  tendermintUpdate set: {} -> {c0}
	keeper.SetValidator(ctx, validators[4])
	keeper.SetValidatorByPowerIndex(ctx, validators[4])
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validators[4], _ = keeper.GetValidator(ctx, validators[4].OperatorAddr)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[4].ABCIValidatorUpdate(), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesWithCliffValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)
	params := types.DefaultParams()
	params.MaxValidators = 2
	keeper.SetParams(ctx, params)

	amts := []int64{10, 20, 5}
	var validators [5]types.Validator
	for i, amt := range amts {
		pool := keeper.GetPool(ctx)
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))
		keeper.SetPool(ctx, pool)
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

	pool := keeper.GetPool(ctx)
	validators[2], pool, _ = validators[2].AddTokensFromDel(pool, sdk.NewInt(10))
	keeper.SetPool(ctx, pool)
	keeper.SetValidator(ctx, validators[2])
	keeper.SetValidatorByPowerIndex(ctx, validators[2])
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validators[2], _ = keeper.GetValidator(ctx, validators[2].OperatorAddr)
	require.Equal(t, 2, len(updates), "%v", updates)
	require.Equal(t, validators[0].ABCIValidatorUpdateZero(), updates[1])
	require.Equal(t, validators[2].ABCIValidatorUpdate(), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesPowerDecrease(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 1000)

	amts := []int64{100, 100}
	var validators [2]types.Validator
	for i, amt := range amts {
		pool := keeper.GetPool(ctx)
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))
		keeper.SetPool(ctx, pool)
	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// check initial power
	require.Equal(t, sdk.NewDec(100).RoundInt64(), validators[0].GetPower().RoundInt64())
	require.Equal(t, sdk.NewDec(100).RoundInt64(), validators[1].GetPower().RoundInt64())

	// test multiple value change
	//  tendermintUpdate set: {c1, c3} -> {c1', c3'}
	pool := keeper.GetPool(ctx)
	validators[0], pool, _ = validators[0].RemoveDelShares(pool, sdk.NewDec(20))
	validators[1], pool, _ = validators[1].RemoveDelShares(pool, sdk.NewDec(30))
	keeper.SetPool(ctx, pool)
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)

	// power has changed
	require.Equal(t, sdk.NewDec(80).RoundInt64(), validators[0].GetPower().RoundInt64())
	require.Equal(t, sdk.NewDec(70).RoundInt64(), validators[1].GetPower().RoundInt64())

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

	amts := []int64{100, 100}
	var validators [2]types.Validator

	// initialize some validators into the state
	for i, amt := range amts {
		pool := keeper.GetPool(ctx)
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = types.NewValidator(valAddr, valPubKey, types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))

		keeper.SetPool(ctx, pool)
		keeper.SetValidator(ctx, validators[i])
		keeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// verify initial Tendermint updates are correct
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, len(validators), len(updates))
	validators[0], _ = keeper.GetValidator(ctx, validators[0].OperatorAddr)
	validators[1], _ = keeper.GetValidator(ctx, validators[1].OperatorAddr)
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[1])

	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// update initial validator set
	for i, amt := range amts {
		pool := keeper.GetPool(ctx)
		keeper.DeleteValidatorByPowerIndex(ctx, validators[i])
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))

		keeper.SetPool(ctx, pool)
		keeper.SetValidator(ctx, validators[i])
		keeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// add a new validator that goes from zero power, to non-zero power, back to
	// zero power
	pool := keeper.GetPool(ctx)
	valPubKey := PKs[len(validators)+1]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	amt := sdk.NewInt(100)

	validator := types.NewValidator(valAddr, valPubKey, types.Description{})
	validator, pool, _ = validator.AddTokensFromDel(pool, amt)

	keeper.SetPool(ctx, pool)
	keeper.SetValidator(ctx, validator)

	validator, pool, _ = validator.RemoveDelShares(pool, sdk.NewDecFromInt(amt))
	keeper.SetValidator(ctx, validator)
	keeper.SetValidatorByPowerIndex(ctx, validator)

	// add a new validator that increases in power
	valPubKey = PKs[len(validators)+2]
	valAddr = sdk.ValAddress(valPubKey.Address().Bytes())

	validator = types.NewValidator(valAddr, valPubKey, types.Description{})
	validator, pool, _ = validator.AddTokensFromDel(pool, sdk.NewInt(500))
	keeper.SetValidator(ctx, validator)
	keeper.SetValidatorByPowerIndex(ctx, validator)
	keeper.SetPool(ctx, pool)

	// verify initial Tendermint updates are correct
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validator, _ = keeper.GetValidator(ctx, validator.OperatorAddr)
	validators[0], _ = keeper.GetValidator(ctx, validators[0].OperatorAddr)
	validators[1], _ = keeper.GetValidator(ctx, validators[1].OperatorAddr)
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

	amts := []int64{100, 200, 300}
	var validators [3]types.Validator

	// initialize some validators into the state
	for i, amt := range amts {
		pool := keeper.GetPool(ctx)
		moniker := fmt.Sprintf("%d", i)
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = types.NewValidator(valAddr, valPubKey, types.Description{Moniker: moniker})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, sdk.NewInt(amt))
		keeper.SetPool(ctx, pool)
		keeper.SetValidator(ctx, validators[i])
		keeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// verify initial Tendermint updates are correct
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))
	validators[2], _ = keeper.GetValidator(ctx, validators[2].OperatorAddr)
	validators[1], _ = keeper.GetValidator(ctx, validators[1].OperatorAddr)
	require.Equal(t, validators[2].ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[1])

	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// delegate to validator with lowest power but not enough to bond
	ctx = ctx.WithBlockHeight(1)
	pool := keeper.GetPool(ctx)

	var found bool
	validators[0], found = keeper.GetValidator(ctx, validators[0].OperatorAddr)
	require.True(t, found)

	keeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	validators[0], pool, _ = validators[0].AddTokensFromDel(pool, sdk.NewInt(1))
	keeper.SetPool(ctx, pool)
	keeper.SetValidator(ctx, validators[0])
	keeper.SetValidatorByPowerIndex(ctx, validators[0])

	// verify initial Tendermint updates are correct
	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// create a series of events that will bond and unbond the validator with
	// lowest power in a single block context (height)
	ctx = ctx.WithBlockHeight(2)
	pool = keeper.GetPool(ctx)

	validators[1], found = keeper.GetValidator(ctx, validators[1].OperatorAddr)
	require.True(t, found)

	keeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	validators[0], pool, _ = validators[0].RemoveDelShares(pool, validators[0].DelegatorShares)
	keeper.SetPool(ctx, pool)
	keeper.SetValidator(ctx, validators[0])
	keeper.SetValidatorByPowerIndex(ctx, validators[0])
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 0, len(updates))

	keeper.DeleteValidatorByPowerIndex(ctx, validators[1])
	validators[1], pool, _ = validators[1].AddTokensFromDel(pool, sdk.NewInt(250))
	keeper.SetPool(ctx, pool)
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
			val, found := keeper.GetValidator(ctx, tc.validator.OperatorAddr)

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
