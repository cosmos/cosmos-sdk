package stake

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	addrDels = []sdk.Address{
		addrs[0],
		addrs[1],
	}
	addrVals = []sdk.Address{
		addrs[2],
		addrs[3],
		addrs[4],
		addrs[5],
		addrs[6],
	}
)

func TestSetValidator(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)

	// test how the validator is set from a purely unbonbed pool
	validator := NewValidator(addrVals[0], pks[0], Description{})
	validator, pool, _ = validator.addTokensFromDel(pool, 10)
	require.Equal(t, sdk.Unbonded, validator.Status())
	assert.True(sdk.RatEq(t, sdk.NewRat(10), validator.PoolShares.Unbonded()))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), validator.DelegatorShares))
	keeper.setPool(ctx, pool)
	keeper.updateValidator(ctx, validator)

	// after the save the validator should be bonded
	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, sdk.Bonded, validator.Status())
	assert.True(sdk.RatEq(t, sdk.NewRat(10), validator.PoolShares.Bonded()))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), validator.DelegatorShares))

	// Check each store for being saved
	resVal, found := keeper.GetValidator(ctx, addrVals[0])
	assert.True(ValEq(t, validator, resVal))

	resVals := keeper.GetValidatorsBonded(ctx)
	require.Equal(t, 1, len(resVals))
	assert.True(ValEq(t, validator, resVals[0]))

	resVals = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, 1, len(resVals))
	assert.True(ValEq(t, validator, resVals[0]))

	updates := keeper.getTendermintUpdates(ctx)
	require.Equal(t, 1, len(updates))
	assert.Equal(t, validator.abciValidator(keeper.cdc), updates[0])

}

// This function tests updateValidator, GetValidator, GetValidatorsBonded, removeValidator
func TestValidatorBasics(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)

	//construct the validators
	var validators [3]Validator
	amts := []int64{9, 8, 7}
	for i, amt := range amts {
		validators[i] = NewValidator(addrVals[i], pks[i], Description{})
		validators[i].PoolShares = NewUnbondedShares(sdk.ZeroRat())
		validators[i].addTokensFromDel(pool, amt)
	}

	// check the empty keeper first
	_, found := keeper.GetValidator(ctx, addrVals[0])
	assert.False(t, found)
	resVals := keeper.GetValidatorsBonded(ctx)
	assert.Zero(t, len(resVals))

	// set and retrieve a record
	validators[0] = keeper.updateValidator(ctx, validators[0])
	resVal, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	assert.True(ValEq(t, validators[0], resVal))

	resVals = keeper.GetValidatorsBonded(ctx)
	require.Equal(t, 1, len(resVals))
	assert.True(ValEq(t, validators[0], resVals[0]))

	// modify a records, save, and retrieve
	validators[0].PoolShares = NewBondedShares(sdk.NewRat(10))
	validators[0].DelegatorShares = sdk.NewRat(10)
	validators[0] = keeper.updateValidator(ctx, validators[0])
	resVal, found = keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	assert.True(ValEq(t, validators[0], resVal))

	resVals = keeper.GetValidatorsBonded(ctx)
	require.Equal(t, 1, len(resVals))
	assert.True(ValEq(t, validators[0], resVals[0]))

	// add other validators
	validators[1] = keeper.updateValidator(ctx, validators[1])
	validators[2] = keeper.updateValidator(ctx, validators[2])
	resVal, found = keeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	assert.True(ValEq(t, validators[1], resVal))
	resVal, found = keeper.GetValidator(ctx, addrVals[2])
	require.True(t, found)
	assert.True(ValEq(t, validators[2], resVal))

	resVals = keeper.GetValidatorsBonded(ctx)
	require.Equal(t, 3, len(resVals))
	assert.True(ValEq(t, validators[0], resVals[2])) // order doesn't matter here
	assert.True(ValEq(t, validators[1], resVals[1]))
	assert.True(ValEq(t, validators[2], resVals[0]))

	// remove a record
	keeper.removeValidator(ctx, validators[1].Owner)
	_, found = keeper.GetValidator(ctx, addrVals[1])
	assert.False(t, found)
}

// test how the validators are sorted, tests GetValidatorsByPower
func GetValidatorSortingUnmixed(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	// initialize some validators into the state
	amts := []int64{0, 100, 1, 400, 200}
	n := len(amts)
	var validators [5]Validator
	for i, amt := range amts {
		validators[i] = NewValidator(addrs[i], pks[i], Description{})
		validators[i].PoolShares = NewBondedShares(sdk.NewRat(amt))
		validators[i].DelegatorShares = sdk.NewRat(amt)
		keeper.updateValidator(ctx, validators[i])
	}

	// first make sure everything made it in to the gotValidator group
	resValidators := keeper.GetValidatorsByPower(ctx)
	require.Equal(t, n, len(resValidators))
	assert.Equal(t, sdk.NewRat(400), resValidators[0].PoolShares.Bonded(), "%v", resValidators)
	assert.Equal(t, sdk.NewRat(200), resValidators[1].PoolShares.Bonded(), "%v", resValidators)
	assert.Equal(t, sdk.NewRat(100), resValidators[2].PoolShares.Bonded(), "%v", resValidators)
	assert.Equal(t, sdk.NewRat(1), resValidators[3].PoolShares.Bonded(), "%v", resValidators)
	assert.Equal(t, sdk.NewRat(0), resValidators[4].PoolShares.Bonded(), "%v", resValidators)
	assert.Equal(t, validators[3].Owner, resValidators[0].Owner, "%v", resValidators)
	assert.Equal(t, validators[4].Owner, resValidators[1].Owner, "%v", resValidators)
	assert.Equal(t, validators[1].Owner, resValidators[2].Owner, "%v", resValidators)
	assert.Equal(t, validators[2].Owner, resValidators[3].Owner, "%v", resValidators)
	assert.Equal(t, validators[0].Owner, resValidators[4].Owner, "%v", resValidators)

	// test a basic increase in voting power
	validators[3].PoolShares = NewBondedShares(sdk.NewRat(500))
	keeper.updateValidator(ctx, validators[3])
	resValidators = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))

	// test a decrease in voting power
	validators[3].PoolShares = NewBondedShares(sdk.NewRat(300))
	keeper.updateValidator(ctx, validators[3])
	resValidators = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))

	// test equal voting power, different age
	validators[3].PoolShares = NewBondedShares(sdk.NewRat(200))
	ctx = ctx.WithBlockHeight(10)
	keeper.updateValidator(ctx, validators[3])
	resValidators = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))
	assert.Equal(t, int64(0), resValidators[0].BondHeight, "%v", resValidators)
	assert.Equal(t, int64(0), resValidators[1].BondHeight, "%v", resValidators)

	// no change in voting power - no change in sort
	ctx = ctx.WithBlockHeight(20)
	keeper.updateValidator(ctx, validators[4])
	resValidators = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))

	// change in voting power of both validators, both still in v-set, no age change
	validators[3].PoolShares = NewBondedShares(sdk.NewRat(300))
	validators[4].PoolShares = NewBondedShares(sdk.NewRat(300))
	keeper.updateValidator(ctx, validators[3])
	resValidators = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	ctx = ctx.WithBlockHeight(30)
	keeper.updateValidator(ctx, validators[4])
	resValidators = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n, "%v", resValidators)
	assert.True(ValEq(t, validators[3], resValidators[0]))
	assert.True(ValEq(t, validators[4], resValidators[1]))
}

func GetValidatorSortingMixed(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	// now 2 max resValidators
	params := keeper.GetParams(ctx)
	params.MaxValidators = 2
	keeper.setParams(ctx, params)

	// initialize some validators into the state
	amts := []int64{0, 100, 1, 400, 200}

	n := len(amts)
	var validators [5]Validator
	for i, amt := range amts {
		validators[i] = NewValidator(addrs[i], pks[i], Description{})
		validators[i].DelegatorShares = sdk.NewRat(amt)
	}
	validators[0].PoolShares = NewUnbondedShares(sdk.NewRat(amts[0]))
	validators[1].PoolShares = NewUnbondedShares(sdk.NewRat(amts[1]))
	validators[2].PoolShares = NewUnbondedShares(sdk.NewRat(amts[2]))
	validators[3].PoolShares = NewBondedShares(sdk.NewRat(amts[3]))
	validators[4].PoolShares = NewBondedShares(sdk.NewRat(amts[4]))
	for i := range amts {
		keeper.updateValidator(ctx, validators[i])
	}
	val0, found := keeper.GetValidator(ctx, addrs[0])
	require.True(t, found)
	val1, found := keeper.GetValidator(ctx, addrs[1])
	require.True(t, found)
	val2, found := keeper.GetValidator(ctx, addrs[2])
	require.True(t, found)
	val3, found := keeper.GetValidator(ctx, addrs[3])
	require.True(t, found)
	val4, found := keeper.GetValidator(ctx, addrs[4])
	require.True(t, found)
	assert.Equal(t, sdk.Unbonded, val0.Status())
	assert.Equal(t, sdk.Unbonded, val1.Status())
	assert.Equal(t, sdk.Unbonded, val2.Status())
	assert.Equal(t, sdk.Bonded, val3.Status())
	assert.Equal(t, sdk.Bonded, val4.Status())

	// first make sure everything made it in to the gotValidator group
	resValidators := keeper.GetValidatorsByPower(ctx)
	require.Equal(t, n, len(resValidators))
	assert.Equal(t, sdk.NewRat(400), resValidators[0].PoolShares.Bonded(), "%v", resValidators)
	assert.Equal(t, sdk.NewRat(200), resValidators[1].PoolShares.Bonded(), "%v", resValidators)
	assert.Equal(t, sdk.NewRat(100), resValidators[2].PoolShares.Bonded(), "%v", resValidators)
	assert.Equal(t, sdk.NewRat(1), resValidators[3].PoolShares.Bonded(), "%v", resValidators)
	assert.Equal(t, sdk.NewRat(0), resValidators[4].PoolShares.Bonded(), "%v", resValidators)
	assert.Equal(t, validators[3].Owner, resValidators[0].Owner, "%v", resValidators)
	assert.Equal(t, validators[4].Owner, resValidators[1].Owner, "%v", resValidators)
	assert.Equal(t, validators[1].Owner, resValidators[2].Owner, "%v", resValidators)
	assert.Equal(t, validators[2].Owner, resValidators[3].Owner, "%v", resValidators)
	assert.Equal(t, validators[0].Owner, resValidators[4].Owner, "%v", resValidators)
}

// TODO seperate out into multiple tests
func TestGetValidatorsEdgeCases(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	var found bool

	// now 2 max resValidators
	params := keeper.GetParams(ctx)
	nMax := uint16(2)
	params.MaxValidators = nMax
	keeper.setParams(ctx, params)

	// initialize some validators into the state
	amts := []int64{0, 100, 400, 400}
	var validators [4]Validator
	for i, amt := range amts {
		validators[i] = NewValidator(addrs[i], pks[i], Description{})
		validators[i].PoolShares = NewUnbondedShares(sdk.NewRat(amt))
		validators[i].DelegatorShares = sdk.NewRat(amt)
		validators[i] = keeper.updateValidator(ctx, validators[i])
	}
	for i := range amts {
		validators[i], found = keeper.GetValidator(ctx, validators[i].Owner)
		require.True(t, found)
	}
	resValidators := keeper.GetValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[2], resValidators[0]))
	assert.True(ValEq(t, validators[3], resValidators[1]))

	validators[0].PoolShares = NewUnbondedShares(sdk.NewRat(500))
	validators[0] = keeper.updateValidator(ctx, validators[0])
	resValidators = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))

	// A validator which leaves the gotValidator set due to a decrease in voting power,
	// then increases to the original voting power, does not get its spot back in the
	// case of a tie.

	// validator 3 enters bonded validator set
	ctx = ctx.WithBlockHeight(40)

	validators[3], found = keeper.GetValidator(ctx, validators[3].Owner)
	require.True(t, found)
	validators[3].PoolShares = NewUnbondedShares(sdk.NewRat(401))
	validators[3] = keeper.updateValidator(ctx, validators[3])
	resValidators = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[3], resValidators[1]))

	// validator 3 kicked out temporarily
	validators[3].PoolShares = NewBondedShares(sdk.NewRat(200))
	validators[3] = keeper.updateValidator(ctx, validators[3])
	resValidators = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))

	// validator 4 does not get spot back
	validators[3].PoolShares = NewBondedShares(sdk.NewRat(400))
	validators[3] = keeper.updateValidator(ctx, validators[3])
	resValidators = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, nMax, uint16(len(resValidators)))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))
	validator, exists := keeper.GetValidator(ctx, validators[3].Owner)
	require.Equal(t, exists, true)
	require.Equal(t, int64(40), validator.BondHeight)
}

func TestValidatorBondHeight(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	// now 2 max resValidators
	params := keeper.GetParams(ctx)
	params.MaxValidators = 2
	keeper.setParams(ctx, params)

	// initialize some validators into the state
	var validators [3]Validator
	validators[0] = NewValidator(addrs[0], pks[0], Description{})
	validators[0].PoolShares = NewUnbondedShares(sdk.NewRat(200))
	validators[0].DelegatorShares = sdk.NewRat(200)
	validators[1] = NewValidator(addrs[1], pks[1], Description{})
	validators[1].PoolShares = NewUnbondedShares(sdk.NewRat(100))
	validators[1].DelegatorShares = sdk.NewRat(100)
	validators[2] = NewValidator(addrs[2], pks[2], Description{})
	validators[2].PoolShares = NewUnbondedShares(sdk.NewRat(100))
	validators[2].DelegatorShares = sdk.NewRat(100)

	validators[0] = keeper.updateValidator(ctx, validators[0])
	////////////////////////////////////////
	// If two validators both increase to the same voting power in the same block,
	// the one with the first transaction should become bonded
	validators[1] = keeper.updateValidator(ctx, validators[1])
	validators[2] = keeper.updateValidator(ctx, validators[2])
	resValidators := keeper.GetValidatorsByPower(ctx)
	require.Equal(t, uint16(len(resValidators)), params.MaxValidators)

	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[1], resValidators[1]))
	validators[1].PoolShares = NewUnbondedShares(sdk.NewRat(150))
	validators[2].PoolShares = NewUnbondedShares(sdk.NewRat(150))
	validators[2] = keeper.updateValidator(ctx, validators[2])
	resValidators = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, params.MaxValidators, uint16(len(resValidators)))
	validators[1] = keeper.updateValidator(ctx, validators[1])
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))
}

func TestFullValidatorSetPowerChange(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := keeper.GetParams(ctx)
	max := 2
	params.MaxValidators = uint16(2)
	keeper.setParams(ctx, params)

	// initialize some validators into the state
	amts := []int64{0, 100, 400, 400, 200}
	var validators [5]Validator
	for i, amt := range amts {
		validators[i] = NewValidator(addrs[i], pks[i], Description{})
		validators[i].PoolShares = NewUnbondedShares(sdk.NewRat(amt))
		validators[i].DelegatorShares = sdk.NewRat(amt)
		keeper.updateValidator(ctx, validators[i])
	}
	for i := range amts {
		var found bool
		validators[i], found = keeper.GetValidator(ctx, validators[i].Owner)
		require.True(t, found)
	}
	assert.Equal(t, sdk.Unbonded, validators[0].Status())
	assert.Equal(t, sdk.Unbonded, validators[1].Status())
	assert.Equal(t, sdk.Bonded, validators[2].Status())
	assert.Equal(t, sdk.Bonded, validators[3].Status())
	assert.Equal(t, sdk.Unbonded, validators[4].Status())
	resValidators := keeper.GetValidatorsByPower(ctx)
	require.Equal(t, max, len(resValidators))
	assert.True(ValEq(t, validators[2], resValidators[0])) // in the order of txs
	assert.True(ValEq(t, validators[3], resValidators[1]))

	// test a swap in voting power
	validators[0].PoolShares = NewUnbondedShares(sdk.NewRat(600))
	validators[0] = keeper.updateValidator(ctx, validators[0])
	resValidators = keeper.GetValidatorsByPower(ctx)
	require.Equal(t, max, len(resValidators))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))
}

// clear the tracked changes to the gotValidator set
func TestClearTendermintUpdates(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	amts := []int64{100, 400, 200}
	validators := make([]Validator, len(amts))
	for i, amt := range amts {
		validators[i] = NewValidator(addrs[i], pks[i], Description{})
		validators[i].PoolShares = NewUnbondedShares(sdk.NewRat(amt))
		validators[i].DelegatorShares = sdk.NewRat(amt)
		keeper.updateValidator(ctx, validators[i])
	}

	updates := keeper.getTendermintUpdates(ctx)
	assert.Equal(t, len(amts), len(updates))
	keeper.clearTendermintUpdates(ctx)
	updates = keeper.getTendermintUpdates(ctx)
	assert.Equal(t, 0, len(updates))
}

func TestGetTendermintUpdatesAllNone(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	amts := []int64{10, 20}
	var validators [2]Validator
	for i, amt := range amts {
		validators[i] = NewValidator(addrs[i], pks[i], Description{})
		validators[i].PoolShares = NewUnbondedShares(sdk.NewRat(amt))
		validators[i].DelegatorShares = sdk.NewRat(amt)
	}

	// test from nothing to something
	//  tendermintUpdate set: {} -> {c1, c3}
	assert.Equal(t, 0, len(keeper.getTendermintUpdates(ctx)))
	validators[0] = keeper.updateValidator(ctx, validators[0])
	validators[1] = keeper.updateValidator(ctx, validators[1])

	updates := keeper.getTendermintUpdates(ctx)
	require.Equal(t, 2, len(updates))
	assert.Equal(t, validators[0].abciValidator(keeper.cdc), updates[0])
	assert.Equal(t, validators[1].abciValidator(keeper.cdc), updates[1])

	// test from something to nothing
	//  tendermintUpdate set: {} -> {c1, c2, c3, c4}
	keeper.clearTendermintUpdates(ctx)
	assert.Equal(t, 0, len(keeper.getTendermintUpdates(ctx)))

	keeper.removeValidator(ctx, validators[0].Owner)
	keeper.removeValidator(ctx, validators[1].Owner)

	updates = keeper.getTendermintUpdates(ctx)
	require.Equal(t, 2, len(updates))
	assert.Equal(t, tmtypes.TM2PB.PubKey(validators[0].PubKey), updates[0].PubKey)
	assert.Equal(t, tmtypes.TM2PB.PubKey(validators[1].PubKey), updates[1].PubKey)
	assert.Equal(t, int64(0), updates[0].Power)
	assert.Equal(t, int64(0), updates[1].Power)
}

func TestGetTendermintUpdatesIdentical(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	amts := []int64{10, 20}
	var validators [2]Validator
	for i, amt := range amts {
		validators[i] = NewValidator(addrs[i], pks[i], Description{})
		validators[i].PoolShares = NewUnbondedShares(sdk.NewRat(amt))
		validators[i].DelegatorShares = sdk.NewRat(amt)
	}
	validators[0] = keeper.updateValidator(ctx, validators[0])
	validators[1] = keeper.updateValidator(ctx, validators[1])
	keeper.clearTendermintUpdates(ctx)
	assert.Equal(t, 0, len(keeper.getTendermintUpdates(ctx)))

	// test identical,
	//  tendermintUpdate set: {} -> {}
	validators[0] = keeper.updateValidator(ctx, validators[0])
	validators[1] = keeper.updateValidator(ctx, validators[1])
	assert.Equal(t, 0, len(keeper.getTendermintUpdates(ctx)))
}

func TestGetTendermintUpdatesSingleValueChange(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	amts := []int64{10, 20}
	var validators [2]Validator
	for i, amt := range amts {
		validators[i] = NewValidator(addrs[i], pks[i], Description{})
		validators[i].PoolShares = NewUnbondedShares(sdk.NewRat(amt))
		validators[i].DelegatorShares = sdk.NewRat(amt)
	}
	validators[0] = keeper.updateValidator(ctx, validators[0])
	validators[1] = keeper.updateValidator(ctx, validators[1])
	keeper.clearTendermintUpdates(ctx)
	assert.Equal(t, 0, len(keeper.getTendermintUpdates(ctx)))

	// test single value change
	//  tendermintUpdate set: {} -> {c1'}
	validators[0].PoolShares = NewBondedShares(sdk.NewRat(600))
	validators[0] = keeper.updateValidator(ctx, validators[0])

	updates := keeper.getTendermintUpdates(ctx)

	require.Equal(t, 1, len(updates))
	assert.Equal(t, validators[0].abciValidator(keeper.cdc), updates[0])
}

func TestGetTendermintUpdatesMultipleValueChange(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	amts := []int64{10, 20}
	var validators [2]Validator
	for i, amt := range amts {
		validators[i] = NewValidator(addrs[i], pks[i], Description{})
		validators[i].PoolShares = NewUnbondedShares(sdk.NewRat(amt))
		validators[i].DelegatorShares = sdk.NewRat(amt)
	}
	validators[0] = keeper.updateValidator(ctx, validators[0])
	validators[1] = keeper.updateValidator(ctx, validators[1])
	keeper.clearTendermintUpdates(ctx)
	assert.Equal(t, 0, len(keeper.getTendermintUpdates(ctx)))

	// test multiple value change
	//  tendermintUpdate set: {c1, c3} -> {c1', c3'}
	validators[0].PoolShares = NewBondedShares(sdk.NewRat(200))
	validators[1].PoolShares = NewBondedShares(sdk.NewRat(100))
	validators[0] = keeper.updateValidator(ctx, validators[0])
	validators[1] = keeper.updateValidator(ctx, validators[1])

	updates := keeper.getTendermintUpdates(ctx)
	require.Equal(t, 2, len(updates))
	require.Equal(t, validators[0].abciValidator(keeper.cdc), updates[0])
	require.Equal(t, validators[1].abciValidator(keeper.cdc), updates[1])
}

func TestGetTendermintUpdatesInserted(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	amts := []int64{10, 20, 5, 15, 25}
	var validators [5]Validator
	for i, amt := range amts {
		validators[i] = NewValidator(addrs[i], pks[i], Description{})
		validators[i].PoolShares = NewUnbondedShares(sdk.NewRat(amt))
		validators[i].DelegatorShares = sdk.NewRat(amt)
	}
	validators[0] = keeper.updateValidator(ctx, validators[0])
	validators[1] = keeper.updateValidator(ctx, validators[1])
	keeper.clearTendermintUpdates(ctx)
	assert.Equal(t, 0, len(keeper.getTendermintUpdates(ctx)))

	// test validtor added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	validators[2] = keeper.updateValidator(ctx, validators[2])
	updates := keeper.getTendermintUpdates(ctx)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[2].abciValidator(keeper.cdc), updates[0])

	// test validtor added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	keeper.clearTendermintUpdates(ctx)
	validators[3] = keeper.updateValidator(ctx, validators[3])
	updates = keeper.getTendermintUpdates(ctx)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[3].abciValidator(keeper.cdc), updates[0])

	// test validtor added at the end
	//  tendermintUpdate set: {} -> {c0}
	keeper.clearTendermintUpdates(ctx)
	validators[4] = keeper.updateValidator(ctx, validators[4])
	updates = keeper.getTendermintUpdates(ctx)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[4].abciValidator(keeper.cdc), updates[0])
}

func TestGetTendermintUpdatesNotValidatorCliff(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := DefaultParams()
	params.MaxValidators = 2
	keeper.setParams(ctx, params)

	amts := []int64{10, 20, 5}
	var validators [5]Validator
	for i, amt := range amts {
		validators[i] = NewValidator(addrs[i], pks[i], Description{})
		validators[i].PoolShares = NewUnbondedShares(sdk.NewRat(amt))
		validators[i].DelegatorShares = sdk.NewRat(amt)
	}
	validators[0] = keeper.updateValidator(ctx, validators[0])
	validators[1] = keeper.updateValidator(ctx, validators[1])
	keeper.clearTendermintUpdates(ctx)
	assert.Equal(t, 0, len(keeper.getTendermintUpdates(ctx)))

	// test validator added at the end but not inserted in the valset
	//  tendermintUpdate set: {} -> {}
	keeper.updateValidator(ctx, validators[2])
	updates := keeper.getTendermintUpdates(ctx)
	require.Equal(t, 0, len(updates))

	// test validator change its power and become a gotValidator (pushing out an existing)
	//  tendermintUpdate set: {}     -> {c0, c4}
	keeper.clearTendermintUpdates(ctx)
	assert.Equal(t, 0, len(keeper.getTendermintUpdates(ctx)))

	validators[2].PoolShares = NewUnbondedShares(sdk.NewRat(15))
	validators[2] = keeper.updateValidator(ctx, validators[2])

	updates = keeper.getTendermintUpdates(ctx)
	require.Equal(t, 2, len(updates), "%v", updates)
	require.Equal(t, validators[0].abciValidatorZero(keeper.cdc), updates[0])
	require.Equal(t, validators[2].abciValidator(keeper.cdc), updates[1])
}

// tests GetDelegation, GetDelegations, SetDelegation, removeDelegation, GetBonds
func TestBond(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	//construct the validators
	amts := []int64{9, 8, 7}
	var validators [3]Validator
	for i, amt := range amts {
		validators[i] = NewValidator(addrVals[i], pks[i], Description{})
		validators[i].PoolShares = NewUnbondedShares(sdk.NewRat(amt))
		validators[i].DelegatorShares = sdk.NewRat(amt)
	}

	// first add a validators[0] to delegate too
	validators[0] = keeper.updateValidator(ctx, validators[0])

	bond1to1 := Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        sdk.NewRat(9),
	}

	// check the empty keeper first
	_, found := keeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	assert.False(t, found)

	// set and retrieve a record
	keeper.setDelegation(ctx, bond1to1)
	resBond, found := keeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	assert.True(t, found)
	assert.True(t, bond1to1.equal(resBond))

	// modify a records, save, and retrieve
	bond1to1.Shares = sdk.NewRat(99)
	keeper.setDelegation(ctx, bond1to1)
	resBond, found = keeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	assert.True(t, found)
	assert.True(t, bond1to1.equal(resBond))

	// add some more records
	validators[1] = keeper.updateValidator(ctx, validators[1])
	validators[2] = keeper.updateValidator(ctx, validators[2])
	bond1to2 := Delegation{addrDels[0], addrVals[1], sdk.NewRat(9), 0}
	bond1to3 := Delegation{addrDels[0], addrVals[2], sdk.NewRat(9), 1}
	bond2to1 := Delegation{addrDels[1], addrVals[0], sdk.NewRat(9), 2}
	bond2to2 := Delegation{addrDels[1], addrVals[1], sdk.NewRat(9), 3}
	bond2to3 := Delegation{addrDels[1], addrVals[2], sdk.NewRat(9), 4}
	keeper.setDelegation(ctx, bond1to2)
	keeper.setDelegation(ctx, bond1to3)
	keeper.setDelegation(ctx, bond2to1)
	keeper.setDelegation(ctx, bond2to2)
	keeper.setDelegation(ctx, bond2to3)

	// test all bond retrieve capabilities
	resBonds := keeper.GetDelegations(ctx, addrDels[0], 5)
	require.Equal(t, 3, len(resBonds))
	assert.True(t, bond1to1.equal(resBonds[0]))
	assert.True(t, bond1to2.equal(resBonds[1]))
	assert.True(t, bond1to3.equal(resBonds[2]))
	resBonds = keeper.GetDelegations(ctx, addrDels[0], 3)
	require.Equal(t, 3, len(resBonds))
	resBonds = keeper.GetDelegations(ctx, addrDels[0], 2)
	require.Equal(t, 2, len(resBonds))
	resBonds = keeper.GetDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 3, len(resBonds))
	assert.True(t, bond2to1.equal(resBonds[0]))
	assert.True(t, bond2to2.equal(resBonds[1]))
	assert.True(t, bond2to3.equal(resBonds[2]))
	allBonds := keeper.getAllDelegations(ctx)
	require.Equal(t, 6, len(allBonds))
	assert.True(t, bond1to1.equal(allBonds[0]))
	assert.True(t, bond1to2.equal(allBonds[1]))
	assert.True(t, bond1to3.equal(allBonds[2]))
	assert.True(t, bond2to1.equal(allBonds[3]))
	assert.True(t, bond2to2.equal(allBonds[4]))
	assert.True(t, bond2to3.equal(allBonds[5]))

	// delete a record
	keeper.removeDelegation(ctx, bond2to3)
	_, found = keeper.GetDelegation(ctx, addrDels[1], addrVals[2])
	assert.False(t, found)
	resBonds = keeper.GetDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 2, len(resBonds))
	assert.True(t, bond2to1.equal(resBonds[0]))
	assert.True(t, bond2to2.equal(resBonds[1]))

	// delete all the records from delegator 2
	keeper.removeDelegation(ctx, bond2to1)
	keeper.removeDelegation(ctx, bond2to2)
	_, found = keeper.GetDelegation(ctx, addrDels[1], addrVals[0])
	assert.False(t, found)
	_, found = keeper.GetDelegation(ctx, addrDels[1], addrVals[1])
	assert.False(t, found)
	resBonds = keeper.GetDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 0, len(resBonds))
}

func TestParams(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	expParams := DefaultParams()

	//check that the empty keeper loads the default
	resParams := keeper.GetParams(ctx)
	assert.True(t, expParams.equal(resParams))

	//modify a params, save, and retrieve
	expParams.MaxValidators = 777
	keeper.setParams(ctx, expParams)
	resParams = keeper.GetParams(ctx)
	assert.True(t, expParams.equal(resParams))
}

func TestPool(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	expPool := InitialPool()

	//check that the empty keeper loads the default
	resPool := keeper.GetPool(ctx)
	assert.True(t, expPool.equal(resPool))

	//modify a params, save, and retrieve
	expPool.BondedTokens = 777
	keeper.setPool(ctx, expPool)
	resPool = keeper.GetPool(ctx)
	assert.True(t, expPool.equal(resPool))
}
