package stake

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

// This function tests GetCandidate, GetCandidates, setCandidate, removeCandidate
func TestCandidate(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	//construct the candidates
	var candidates [3]Candidate
	amts := []int64{9, 8, 7}
	for i, amt := range amts {
		candidates[i] = NewCandidate(addrVals[i], pks[i], Description{})
		candidates[i].BondedShares = sdk.NewRat(amt)
		candidates[i].DelegatorShares = sdk.NewRat(amt)
	}

	// check the empty keeper first
	_, found := keeper.GetCandidate(ctx, addrVals[0])
	assert.False(t, found)
	resCands := keeper.GetCandidates(ctx, 100)
	assert.Zero(t, len(resCands))

	// set and retrieve a record
	keeper.setCandidate(ctx, candidates[0])
	resCand, found := keeper.GetCandidate(ctx, addrVals[0])
	require.True(t, found)
	assert.True(t, candidates[0].equal(resCand), "%v \n %v", resCand, candidates[0])

	// modify a records, save, and retrieve
	candidates[0].DelegatorShares = sdk.NewRat(99)
	keeper.setCandidate(ctx, candidates[0])
	resCand, found = keeper.GetCandidate(ctx, addrVals[0])
	require.True(t, found)
	assert.True(t, candidates[0].equal(resCand))

	// also test that the address has been added to address list
	resCands = keeper.GetCandidates(ctx, 100)
	require.Equal(t, 1, len(resCands))
	assert.Equal(t, addrVals[0], resCands[0].Address)

	// add other candidates
	keeper.setCandidate(ctx, candidates[1])
	keeper.setCandidate(ctx, candidates[2])
	resCand, found = keeper.GetCandidate(ctx, addrVals[1])
	require.True(t, found)
	assert.True(t, candidates[1].equal(resCand), "%v \n %v", resCand, candidates[1])
	resCand, found = keeper.GetCandidate(ctx, addrVals[2])
	require.True(t, found)
	assert.True(t, candidates[2].equal(resCand), "%v \n %v", resCand, candidates[2])
	resCands = keeper.GetCandidates(ctx, 100)
	require.Equal(t, 3, len(resCands))
	assert.True(t, candidates[0].equal(resCands[0]), "%v \n %v", resCands[0], candidates[0])
	assert.True(t, candidates[1].equal(resCands[1]), "%v \n %v", resCands[1], candidates[1])
	assert.True(t, candidates[2].equal(resCands[2]), "%v \n %v", resCands[2], candidates[2])

	// remove a record
	keeper.removeCandidate(ctx, candidates[1].Address)
	_, found = keeper.GetCandidate(ctx, addrVals[1])
	assert.False(t, found)
}

// tests GetDelegatorBond, GetDelegatorBonds, SetDelegatorBond, removeDelegatorBond, GetBonds
func TestBond(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	//construct the candidates
	amts := []int64{9, 8, 7}
	var candidates [3]Candidate
	for i, amt := range amts {
		candidates[i] = NewCandidate(addrVals[i], pks[i], Description{})
		candidates[i].BondedShares = sdk.NewRat(amt)
		candidates[i].DelegatorShares = sdk.NewRat(amt)
	}

	// first add a candidates[0] to delegate too
	keeper.setCandidate(ctx, candidates[0])

	bond1to1 := DelegatorBond{
		DelegatorAddr: addrDels[0],
		CandidateAddr: addrVals[0],
		Shares:        sdk.NewRat(9),
	}

	// check the empty keeper first
	_, found := keeper.GetDelegatorBond(ctx, addrDels[0], addrVals[0])
	assert.False(t, found)

	// set and retrieve a record
	keeper.setDelegatorBond(ctx, bond1to1)
	resBond, found := keeper.GetDelegatorBond(ctx, addrDels[0], addrVals[0])
	assert.True(t, found)
	assert.True(t, bond1to1.equal(resBond))

	// modify a records, save, and retrieve
	bond1to1.Shares = sdk.NewRat(99)
	keeper.setDelegatorBond(ctx, bond1to1)
	resBond, found = keeper.GetDelegatorBond(ctx, addrDels[0], addrVals[0])
	assert.True(t, found)
	assert.True(t, bond1to1.equal(resBond))

	// add some more records
	keeper.setCandidate(ctx, candidates[1])
	keeper.setCandidate(ctx, candidates[2])
	bond1to2 := DelegatorBond{addrDels[0], addrVals[1], sdk.NewRat(9), 0}
	bond1to3 := DelegatorBond{addrDels[0], addrVals[2], sdk.NewRat(9), 1}
	bond2to1 := DelegatorBond{addrDels[1], addrVals[0], sdk.NewRat(9), 2}
	bond2to2 := DelegatorBond{addrDels[1], addrVals[1], sdk.NewRat(9), 3}
	bond2to3 := DelegatorBond{addrDels[1], addrVals[2], sdk.NewRat(9), 4}
	keeper.setDelegatorBond(ctx, bond1to2)
	keeper.setDelegatorBond(ctx, bond1to3)
	keeper.setDelegatorBond(ctx, bond2to1)
	keeper.setDelegatorBond(ctx, bond2to2)
	keeper.setDelegatorBond(ctx, bond2to3)

	// test all bond retrieve capabilities
	resBonds := keeper.GetDelegatorBonds(ctx, addrDels[0], 5)
	require.Equal(t, 3, len(resBonds))
	assert.True(t, bond1to1.equal(resBonds[0]))
	assert.True(t, bond1to2.equal(resBonds[1]))
	assert.True(t, bond1to3.equal(resBonds[2]))
	resBonds = keeper.GetDelegatorBonds(ctx, addrDels[0], 3)
	require.Equal(t, 3, len(resBonds))
	resBonds = keeper.GetDelegatorBonds(ctx, addrDels[0], 2)
	require.Equal(t, 2, len(resBonds))
	resBonds = keeper.GetDelegatorBonds(ctx, addrDels[1], 5)
	require.Equal(t, 3, len(resBonds))
	assert.True(t, bond2to1.equal(resBonds[0]))
	assert.True(t, bond2to2.equal(resBonds[1]))
	assert.True(t, bond2to3.equal(resBonds[2]))
	allBonds := keeper.getBonds(ctx, 1000)
	require.Equal(t, 6, len(allBonds))
	assert.True(t, bond1to1.equal(allBonds[0]))
	assert.True(t, bond1to2.equal(allBonds[1]))
	assert.True(t, bond1to3.equal(allBonds[2]))
	assert.True(t, bond2to1.equal(allBonds[3]))
	assert.True(t, bond2to2.equal(allBonds[4]))
	assert.True(t, bond2to3.equal(allBonds[5]))

	// delete a record
	keeper.removeDelegatorBond(ctx, bond2to3)
	_, found = keeper.GetDelegatorBond(ctx, addrDels[1], addrVals[2])
	assert.False(t, found)
	resBonds = keeper.GetDelegatorBonds(ctx, addrDels[1], 5)
	require.Equal(t, 2, len(resBonds))
	assert.True(t, bond2to1.equal(resBonds[0]))
	assert.True(t, bond2to2.equal(resBonds[1]))

	// delete all the records from delegator 2
	keeper.removeDelegatorBond(ctx, bond2to1)
	keeper.removeDelegatorBond(ctx, bond2to2)
	_, found = keeper.GetDelegatorBond(ctx, addrDels[1], addrVals[0])
	assert.False(t, found)
	_, found = keeper.GetDelegatorBond(ctx, addrDels[1], addrVals[1])
	assert.False(t, found)
	resBonds = keeper.GetDelegatorBonds(ctx, addrDels[1], 5)
	require.Equal(t, 0, len(resBonds))
}

// TODO seperate out into multiple tests
func TestGetValidators(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	// initialize some candidates into the state
	amts := []int64{0, 100, 1, 400, 200}
	n := len(amts)
	var candidates [5]Candidate
	for i, amt := range amts {
		candidates[i] = NewCandidate(addrs[i], pks[i], Description{})
		candidates[i].BondedShares = sdk.NewRat(amt)
		candidates[i].DelegatorShares = sdk.NewRat(amt)
		keeper.setCandidate(ctx, candidates[i])
	}

	// first make sure everything made it in to the validator group
	validators := keeper.getValidatorsOrdered(ctx)
	require.Equal(t, len(validators), n)
	assert.Equal(t, sdk.NewRat(400), validators[0].Power, "%v", validators)
	assert.Equal(t, sdk.NewRat(200), validators[1].Power, "%v", validators)
	assert.Equal(t, sdk.NewRat(100), validators[2].Power, "%v", validators)
	assert.Equal(t, sdk.NewRat(1), validators[3].Power, "%v", validators)
	assert.Equal(t, sdk.NewRat(0), validators[4].Power, "%v", validators)
	assert.Equal(t, candidates[3].Address, validators[0].Address, "%v", validators)
	assert.Equal(t, candidates[4].Address, validators[1].Address, "%v", validators)
	assert.Equal(t, candidates[1].Address, validators[2].Address, "%v", validators)
	assert.Equal(t, candidates[2].Address, validators[3].Address, "%v", validators)
	assert.Equal(t, candidates[0].Address, validators[4].Address, "%v", validators)

	// test a basic increase in voting power
	candidates[3].BondedShares = sdk.NewRat(500)
	keeper.setCandidate(ctx, candidates[3])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, len(validators), n)
	assert.Equal(t, sdk.NewRat(500), validators[0].Power, "%v", validators)
	assert.Equal(t, candidates[3].Address, validators[0].Address, "%v", validators)

	// test a decrease in voting power
	candidates[3].BondedShares = sdk.NewRat(300)
	keeper.setCandidate(ctx, candidates[3])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, len(validators), n)
	assert.Equal(t, sdk.NewRat(300), validators[0].Power, "%v", validators)
	assert.Equal(t, candidates[3].Address, validators[0].Address, "%v", validators)
	assert.Equal(t, candidates[4].Address, validators[1].Address, "%v", validators)

	// test equal voting power, different age
	candidates[3].BondedShares = sdk.NewRat(200)
	ctx = ctx.WithBlockHeight(10)
	keeper.setCandidate(ctx, candidates[3])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, len(validators), n)
	assert.Equal(t, sdk.NewRat(200), validators[0].Power, "%v", validators)
	assert.Equal(t, sdk.NewRat(200), validators[1].Power, "%v", validators)
	assert.Equal(t, candidates[3].Address, validators[0].Address, "%v", validators)
	assert.Equal(t, candidates[4].Address, validators[1].Address, "%v", validators)
	assert.Equal(t, int64(0), validators[0].Height, "%v", validators)
	assert.Equal(t, int64(0), validators[1].Height, "%v", validators)

	// no change in voting power - no change in sort
	ctx = ctx.WithBlockHeight(20)
	keeper.setCandidate(ctx, candidates[4])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, len(validators), n)
	assert.Equal(t, candidates[3].Address, validators[0].Address, "%v", validators)
	assert.Equal(t, candidates[4].Address, validators[1].Address, "%v", validators)

	// change in voting power of both candidates, both still in v-set, no age change
	candidates[3].BondedShares = sdk.NewRat(300)
	candidates[4].BondedShares = sdk.NewRat(300)
	keeper.setCandidate(ctx, candidates[3])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, len(validators), n)
	ctx = ctx.WithBlockHeight(30)
	keeper.setCandidate(ctx, candidates[4])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, len(validators), n, "%v", validators)
	assert.Equal(t, candidates[3].Address, validators[0].Address, "%v", validators)
	assert.Equal(t, candidates[4].Address, validators[1].Address, "%v", validators)
}

// TODO seperate out into multiple tests
func TestGetValidatorsEdgeCases(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	// now 2 max validators
	params := keeper.GetParams(ctx)
	params.MaxValidators = 2
	keeper.setParams(ctx, params)

	// initialize some candidates into the state
	amts := []int64{0, 100, 400, 400, 200}
	n := len(amts)
	var candidates [5]Candidate
	for i, amt := range amts {
		candidates[i] = NewCandidate(addrs[i], pks[i], Description{})
		candidates[i].BondedShares = sdk.NewRat(amt)
		candidates[i].DelegatorShares = sdk.NewRat(amt)
		keeper.setCandidate(ctx, candidates[i])
	}

	candidates[0].BondedShares = sdk.NewRat(500)
	keeper.setCandidate(ctx, candidates[0])
	validators := keeper.getValidatorsOrdered(ctx)
	require.Equal(t, uint16(len(validators)), params.MaxValidators)
	require.Equal(t, candidates[0].Address, validators[0].Address, "%v", validators)
	// candidate 3 was set before candidate 4
	require.Equal(t, candidates[2].Address, validators[1].Address, "%v", validators)

	// A candidate which leaves the validator set due to a decrease in voting power,
	// then increases to the original voting power, does not get its spot back in the
	// case of a tie.
	// ref https://github.com/cosmos/cosmos-sdk/issues/582#issuecomment-380757108
	candidates[3].BondedShares = sdk.NewRat(401)
	keeper.setCandidate(ctx, candidates[3])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, uint16(len(validators)), params.MaxValidators)
	require.Equal(t, candidates[0].Address, validators[0].Address, "%v", validators)
	require.Equal(t, candidates[3].Address, validators[1].Address, "%v", validators)
	ctx = ctx.WithBlockHeight(40)
	// candidate 3 kicked out temporarily
	candidates[3].BondedShares = sdk.NewRat(200)
	keeper.setCandidate(ctx, candidates[3])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, uint16(len(validators)), params.MaxValidators)
	require.Equal(t, candidates[0].Address, validators[0].Address, "%v", validators)
	require.Equal(t, candidates[2].Address, validators[1].Address, "%v", validators)
	// candidate 4 does not get spot back
	candidates[3].BondedShares = sdk.NewRat(400)
	keeper.setCandidate(ctx, candidates[3])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, uint16(len(validators)), params.MaxValidators)
	require.Equal(t, candidates[0].Address, validators[0].Address, "%v", validators)
	require.Equal(t, candidates[2].Address, validators[1].Address, "%v", validators)
	candidate, exists := keeper.GetCandidate(ctx, candidates[3].Address)
	require.Equal(t, exists, true)
	require.Equal(t, candidate.ValidatorBondHeight, int64(40))

	// If two candidates both increase to the same voting power in the same block,
	// the one with the first transaction should take precedence (become a validator).
	// ref https://github.com/cosmos/cosmos-sdk/issues/582#issuecomment-381250392
	candidates[0].BondedShares = sdk.NewRat(2000)
	keeper.setCandidate(ctx, candidates[0])
	candidates[1].BondedShares = sdk.NewRat(1000)
	candidates[2].BondedShares = sdk.NewRat(1000)
	keeper.setCandidate(ctx, candidates[1])
	keeper.setCandidate(ctx, candidates[2])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, uint16(len(validators)), params.MaxValidators)
	require.Equal(t, candidates[0].Address, validators[0].Address, "%v", validators)
	require.Equal(t, candidates[1].Address, validators[1].Address, "%v", validators)
	candidates[1].BondedShares = sdk.NewRat(1100)
	candidates[2].BondedShares = sdk.NewRat(1100)
	keeper.setCandidate(ctx, candidates[2])
	keeper.setCandidate(ctx, candidates[1])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, uint16(len(validators)), params.MaxValidators)
	require.Equal(t, candidates[0].Address, validators[0].Address, "%v", validators)
	require.Equal(t, candidates[2].Address, validators[1].Address, "%v", validators)

	// reset assets / heights
	params.MaxValidators = 100
	keeper.setParams(ctx, params)
	candidates[0].BondedShares = sdk.NewRat(0)
	candidates[1].BondedShares = sdk.NewRat(100)
	candidates[2].BondedShares = sdk.NewRat(1)
	candidates[3].BondedShares = sdk.NewRat(300)
	candidates[4].BondedShares = sdk.NewRat(200)
	ctx = ctx.WithBlockHeight(0)
	keeper.setCandidate(ctx, candidates[0])
	keeper.setCandidate(ctx, candidates[1])
	keeper.setCandidate(ctx, candidates[2])
	keeper.setCandidate(ctx, candidates[3])
	keeper.setCandidate(ctx, candidates[4])

	// test a swap in voting power
	candidates[0].BondedShares = sdk.NewRat(600)
	keeper.setCandidate(ctx, candidates[0])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, len(validators), n)
	assert.Equal(t, sdk.NewRat(600), validators[0].Power, "%v", validators)
	assert.Equal(t, candidates[0].Address, validators[0].Address, "%v", validators)
	assert.Equal(t, sdk.NewRat(300), validators[1].Power, "%v", validators)
	assert.Equal(t, candidates[3].Address, validators[1].Address, "%v", validators)

	// test the max validators term
	params = keeper.GetParams(ctx)
	n = 2
	params.MaxValidators = uint16(n)
	keeper.setParams(ctx, params)
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, len(validators), n)
	assert.Equal(t, sdk.NewRat(600), validators[0].Power, "%v", validators)
	assert.Equal(t, candidates[0].Address, validators[0].Address, "%v", validators)
	assert.Equal(t, sdk.NewRat(300), validators[1].Power, "%v", validators)
	assert.Equal(t, candidates[3].Address, validators[1].Address, "%v", validators)
}

// clear the tracked changes to the validator set
func TestClearAccUpdateValidators(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	amts := []int64{100, 400, 200}
	candidates := make([]Candidate, len(amts))
	for i, amt := range amts {
		candidates[i] = NewCandidate(addrs[i], pks[i], Description{})
		candidates[i].BondedShares = sdk.NewRat(amt)
		candidates[i].DelegatorShares = sdk.NewRat(amt)
		keeper.setCandidate(ctx, candidates[i])
	}

	acc := keeper.getAccUpdateValidators(ctx)
	assert.Equal(t, len(amts), len(acc))
	keeper.clearAccUpdateValidators(ctx)
	acc = keeper.getAccUpdateValidators(ctx)
	assert.Equal(t, 0, len(acc))
}

// test the mechanism which keeps track of a validator set change
func TestGetAccUpdateValidators(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := defaultParams()
	params.MaxValidators = 4
	keeper.setParams(ctx, params)

	// TODO eliminate use of candidatesIn here
	// tests could be clearer if they just
	// created the candidate at time of use
	// and were labelled by power in the comments
	// outlining in each test
	amts := []int64{10, 11, 12, 13, 1}
	var candidatesIn [5]Candidate
	for i, amt := range amts {
		candidatesIn[i] = NewCandidate(addrs[i], pks[i], Description{})
		candidatesIn[i].BondedShares = sdk.NewRat(amt)
		candidatesIn[i].DelegatorShares = sdk.NewRat(amt)
	}

	// test from nothing to something
	//  candidate set: {} -> {c1, c3}
	//  validator set: {} -> {c1, c3}
	//  accUpdate set: {} -> {c1, c3}
	assert.Equal(t, 0, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 0, len(keeper.GetValidators(ctx)))
	assert.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx)))

	keeper.setCandidate(ctx, candidatesIn[1])
	keeper.setCandidate(ctx, candidatesIn[3])

	vals := keeper.getValidatorsOrdered(ctx) // to init recent validator set
	require.Equal(t, 2, len(vals))
	acc := keeper.getAccUpdateValidators(ctx)
	require.Equal(t, 2, len(acc))
	candidates := keeper.GetCandidates(ctx, 5)
	require.Equal(t, 2, len(candidates))
	assert.Equal(t, candidates[0].validator().abciValidator(keeper.cdc), acc[0])
	assert.Equal(t, candidates[1].validator().abciValidator(keeper.cdc), acc[1])
	assert.True(t, candidates[0].validator().equal(vals[1]))
	assert.True(t, candidates[1].validator().equal(vals[0]))

	// test identical,
	//  candidate set: {c1, c3} -> {c1, c3}
	//  accUpdate set: {} -> {}
	keeper.clearAccUpdateValidators(ctx)
	assert.Equal(t, 2, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx)))

	keeper.setCandidate(ctx, candidates[0])
	keeper.setCandidate(ctx, candidates[1])

	require.Equal(t, 2, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx)))

	// test single value change
	//  candidate set: {c1, c3} -> {c1', c3}
	//  accUpdate set: {} -> {c1'}
	keeper.clearAccUpdateValidators(ctx)
	assert.Equal(t, 2, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx)))

	candidates[0].BondedShares = sdk.NewRat(600)
	keeper.setCandidate(ctx, candidates[0])

	candidates = keeper.GetCandidates(ctx, 5)
	require.Equal(t, 2, len(candidates))
	assert.True(t, candidates[0].BondedShares.Equal(sdk.NewRat(600)))
	acc = keeper.getAccUpdateValidators(ctx)
	require.Equal(t, 1, len(acc))
	assert.Equal(t, candidates[0].validator().abciValidator(keeper.cdc), acc[0])

	// test multiple value change
	//  candidate set: {c1, c3} -> {c1', c3'}
	//  accUpdate set: {c1, c3} -> {c1', c3'}
	keeper.clearAccUpdateValidators(ctx)
	assert.Equal(t, 2, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx)))

	candidates[0].BondedShares = sdk.NewRat(200)
	candidates[1].BondedShares = sdk.NewRat(100)
	keeper.setCandidate(ctx, candidates[0])
	keeper.setCandidate(ctx, candidates[1])

	acc = keeper.getAccUpdateValidators(ctx)
	require.Equal(t, 2, len(acc))
	candidates = keeper.GetCandidates(ctx, 5)
	require.Equal(t, 2, len(candidates))
	require.Equal(t, candidates[0].validator().abciValidator(keeper.cdc), acc[0])
	require.Equal(t, candidates[1].validator().abciValidator(keeper.cdc), acc[1])

	// test validtor added at the beginning
	//  candidate set: {c1, c3} -> {c0, c1, c3}
	//  accUpdate set: {} -> {c0}
	keeper.clearAccUpdateValidators(ctx)
	assert.Equal(t, 2, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx)))

	keeper.setCandidate(ctx, candidatesIn[0])
	acc = keeper.getAccUpdateValidators(ctx)
	require.Equal(t, 1, len(acc))
	candidates = keeper.GetCandidates(ctx, 5)
	require.Equal(t, 3, len(candidates))
	assert.Equal(t, candidates[0].validator().abciValidator(keeper.cdc), acc[0])

	// test validator added at the middle
	//  candidate set: {c0, c1, c3} -> {c0, c1, c2, c3]
	//  accUpdate set: {} -> {c2}
	keeper.clearAccUpdateValidators(ctx)
	assert.Equal(t, 3, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx)))

	keeper.setCandidate(ctx, candidatesIn[2])
	acc = keeper.getAccUpdateValidators(ctx)
	require.Equal(t, 1, len(acc))
	candidates = keeper.GetCandidates(ctx, 5)
	require.Equal(t, 4, len(candidates))
	assert.Equal(t, candidates[2].validator().abciValidator(keeper.cdc), acc[0])

	// test candidate added at the end but not inserted in the valset
	//  candidate set: {c0, c1, c2, c3} -> {c0, c1, c2, c3, c4}
	//  validator set: {c0, c1, c2, c3} -> {c0, c1, c2, c3}
	//  accUpdate set: {} -> {}
	keeper.clearAccUpdateValidators(ctx)
	assert.Equal(t, 4, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 4, len(keeper.GetValidators(ctx)))
	assert.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx)))

	keeper.setCandidate(ctx, candidatesIn[4])

	assert.Equal(t, 5, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 4, len(keeper.GetValidators(ctx)))
	require.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx))) // max validator number is 4

	// test candidate change its power but still not in the valset
	//  candidate set: {c0, c1, c2, c3, c4} -> {c0, c1, c2, c3, c4}
	//  validator set: {c0, c1, c2, c3}     -> {c0, c1, c2, c3}
	//  accUpdate set: {}     -> {}
	keeper.clearAccUpdateValidators(ctx)
	assert.Equal(t, 5, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 4, len(keeper.GetValidators(ctx)))
	assert.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx)))

	candidatesIn[4].BondedShares = sdk.NewRat(1)
	keeper.setCandidate(ctx, candidatesIn[4])

	assert.Equal(t, 5, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 4, len(keeper.GetValidators(ctx)))
	require.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx))) // max validator number is 4

	// test candidate change its power and become a validator (pushing out an existing)
	//  candidate set: {c0, c1, c2, c3, c4} -> {c0, c1, c2, c3, c4}
	//  validator set: {c0, c1, c2, c3}     -> {c1, c2, c3, c4}
	//  accUpdate set: {}     -> {c0, c4}
	keeper.clearAccUpdateValidators(ctx)
	assert.Equal(t, 5, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 4, len(keeper.GetValidators(ctx)))
	assert.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx)))

	candidatesIn[4].BondedShares = sdk.NewRat(1000)
	keeper.setCandidate(ctx, candidatesIn[4])

	candidates = keeper.GetCandidates(ctx, 5)
	require.Equal(t, 5, len(candidates))
	vals = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, 4, len(vals))
	assert.Equal(t, candidatesIn[1].Address, vals[1].Address)
	assert.Equal(t, candidatesIn[2].Address, vals[3].Address)
	assert.Equal(t, candidatesIn[3].Address, vals[2].Address)
	assert.Equal(t, candidatesIn[4].Address, vals[0].Address)

	acc = keeper.getAccUpdateValidators(ctx)
	require.Equal(t, 2, len(acc), "%v", acc)

	assert.Equal(t, candidatesIn[0].PubKey.Bytes(), acc[0].PubKey)
	assert.Equal(t, int64(0), acc[0].Power)
	assert.Equal(t, vals[0].abciValidator(keeper.cdc), acc[1])

	// test from something to nothing
	//  candidate set: {c0, c1, c2, c3, c4} -> {}
	//  validator set: {c1, c2, c3, c4}  -> {}
	//  accUpdate set: {} -> {c1, c2, c3, c4}
	keeper.clearAccUpdateValidators(ctx)
	assert.Equal(t, 5, len(keeper.GetCandidates(ctx, 5)))
	assert.Equal(t, 4, len(keeper.GetValidators(ctx)))
	assert.Equal(t, 0, len(keeper.getAccUpdateValidators(ctx)))

	keeper.removeCandidate(ctx, candidatesIn[0].Address)
	keeper.removeCandidate(ctx, candidatesIn[1].Address)
	keeper.removeCandidate(ctx, candidatesIn[2].Address)
	keeper.removeCandidate(ctx, candidatesIn[3].Address)
	keeper.removeCandidate(ctx, candidatesIn[4].Address)

	vals = keeper.getValidatorsOrdered(ctx)
	assert.Equal(t, 0, len(vals), "%v", vals)
	candidates = keeper.GetCandidates(ctx, 5)
	require.Equal(t, 0, len(candidates))
	acc = keeper.getAccUpdateValidators(ctx)
	require.Equal(t, 4, len(acc))
	assert.Equal(t, candidatesIn[1].PubKey.Bytes(), acc[0].PubKey)
	assert.Equal(t, candidatesIn[2].PubKey.Bytes(), acc[1].PubKey)
	assert.Equal(t, candidatesIn[3].PubKey.Bytes(), acc[2].PubKey)
	assert.Equal(t, candidatesIn[4].PubKey.Bytes(), acc[3].PubKey)
	assert.Equal(t, int64(0), acc[0].Power)
	assert.Equal(t, int64(0), acc[1].Power)
	assert.Equal(t, int64(0), acc[2].Power)
	assert.Equal(t, int64(0), acc[3].Power)
}

// test if is a validator from the last update
func TestIsValidator(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	amts := []int64{9, 8, 7, 10, 6}
	var candidatesIn [5]Candidate
	for i, amt := range amts {
		candidatesIn[i] = NewCandidate(addrVals[i], pks[i], Description{})
		candidatesIn[i].BondedShares = sdk.NewRat(amt)
		candidatesIn[i].DelegatorShares = sdk.NewRat(amt)
	}

	// test that an empty validator set doesn't have any validators
	validators := keeper.getValidatorsOrdered(ctx)
	assert.Equal(t, 0, len(validators))

	// get the validators for the first time
	keeper.setCandidate(ctx, candidatesIn[0])
	keeper.setCandidate(ctx, candidatesIn[1])
	validators = keeper.getValidatorsOrdered(ctx)
	require.Equal(t, 2, len(validators))
	assert.True(t, candidatesIn[0].validator().equal(validators[0]))
	c1ValWithCounter := candidatesIn[1].validator()
	c1ValWithCounter.Counter = int16(1)
	assert.True(t, c1ValWithCounter.equal(validators[1]))

	// test a basic retrieve of something that should be a recent validator
	assert.True(t, keeper.IsValidator(ctx, candidatesIn[0].PubKey))
	assert.True(t, keeper.IsValidator(ctx, candidatesIn[1].PubKey))

	// test a basic retrieve of something that should not be a recent validator
	assert.False(t, keeper.IsValidator(ctx, candidatesIn[2].PubKey))

	// remove that validator, but don't retrieve the recent validator group
	keeper.removeCandidate(ctx, candidatesIn[0].Address)

	// test that removed validator is not considered a recent validator
	assert.False(t, keeper.IsValidator(ctx, candidatesIn[0].PubKey))
}

// test if is a validator from the last update
func TestGetTotalPrecommitVotingPower(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	amts := []int64{10000, 1000, 100, 10, 1}
	var candidatesIn [5]Candidate
	for i, amt := range amts {
		candidatesIn[i] = NewCandidate(addrVals[i], pks[i], Description{})
		candidatesIn[i].BondedShares = sdk.NewRat(amt)
		candidatesIn[i].DelegatorShares = sdk.NewRat(amt)
		keeper.setCandidate(ctx, candidatesIn[i])
	}

	// test that an empty validator set doesn't have any validators
	validators := keeper.GetValidators(ctx)
	assert.Equal(t, 5, len(validators))

	totPow := keeper.GetTotalPrecommitVotingPower(ctx)
	exp := sdk.NewRat(11111)
	assert.True(t, exp.Equal(totPow), "exp %v, got %v", exp, totPow)

	// set absent validators to be the 1st and 3rd record sorted by pubKey address
	ctx = ctx.WithAbsentValidators([]int32{1, 3})
	totPow = keeper.GetTotalPrecommitVotingPower(ctx)

	// XXX verify that this order should infact exclude these two records
	exp = sdk.NewRat(11100)
	assert.True(t, exp.Equal(totPow), "exp %v, got %v", exp, totPow)
}

func TestParams(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	expParams := defaultParams()

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
	expPool := initialPool()

	//check that the empty keeper loads the default
	resPool := keeper.GetPool(ctx)
	assert.True(t, expPool.equal(resPool))

	//modify a params, save, and retrieve
	expPool.TotalSupply = 777
	keeper.setPool(ctx, expPool)
	resPool = keeper.GetPool(ctx)
	assert.True(t, expPool.equal(resPool))
}

func TestValidatorsetKeeper(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	total := int64(0)
	amts := []int64{9, 8, 7}
	var validators [3]Validator
	for i, amt := range amts {
		candidates[i] = Candidate{
			Address:     addrVals[i],
			PubKey:      pks[i],
			Assets:      sdk.NewRat(amt),
			Liabilities: sdk.NewRat(amt),
		}

		keeper.setValidator(ctx, validators[i])

		total += amt
	}

	assert.Equal(t, 3, keeper.Size(ctx))

	for _, addr := range addrVals[:3] {
		assert.True(t, keeper.IsValidator(ctx, addr))
	}
	for _, addr := range addrVals[3:] {
		assert.False(t, keeper.IsValidator(ctx, addr))
	}

	for i, addr := range addrVals[:3] {
		index, val := keeper.GetByAddress(ctx, addr)
		assert.Equal(t, i, index)
		assert.Equal(t, candidates[i].validator().abciValidator(keeper.cdc), *val)
	}

	for _, addr := range addrVals[3:] {
		index, val := keeper.GetByAddress(ctx, addr)
		assert.Equal(t, -1, index)
		assert.Nil(t, val)
	}

	for i, can := range candidates {
		assert.Equal(t, can.validator().abciValidator(keeper.cdc), *keeper.GetByIndex(ctx, i))
	}

	assert.Equal(t, total, keeper.TotalPower(ctx).Evaluate())
}
