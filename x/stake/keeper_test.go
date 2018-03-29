package stake

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	addrDel1 = addrs[0]
	addrDel2 = addrs[1]
	addrVal1 = addrs[2]
	addrVal2 = addrs[3]
	addrVal3 = addrs[4]
	pk1      = crypto.GenPrivKeyEd25519().PubKey()
	pk2      = crypto.GenPrivKeyEd25519().PubKey()
	pk3      = crypto.GenPrivKeyEd25519().PubKey()

	candidate1 = Candidate{
		Address:     addrVal1,
		PubKey:      pk1,
		Assets:      sdk.NewRat(9),
		Liabilities: sdk.NewRat(9),
	}
	candidate2 = Candidate{
		Address:     addrVal2,
		PubKey:      pk2,
		Assets:      sdk.NewRat(8),
		Liabilities: sdk.NewRat(8),
	}
	candidate3 = Candidate{
		Address:     addrVal3,
		PubKey:      pk3,
		Assets:      sdk.NewRat(7),
		Liabilities: sdk.NewRat(7),
	}
)

// This function tests GetCandidate, GetCandidates, setCandidate, removeCandidate
func TestCandidate(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	candidatesEqual := func(c1, c2 Candidate) bool {
		return c1.Status == c2.Status &&
			c1.PubKey.Equals(c2.PubKey) &&
			bytes.Equal(c1.Address, c2.Address) &&
			c1.Assets.Equal(c2.Assets) &&
			c1.Liabilities.Equal(c2.Liabilities) &&
			c1.Description == c2.Description
	}

	// check the empty keeper first
	_, found := keeper.GetCandidate(ctx, addrVal1)
	assert.False(t, found)
	resCands := keeper.GetCandidates(ctx, 100)
	assert.Zero(t, len(resCands))

	// set and retrieve a record
	keeper.setCandidate(ctx, candidate1)
	resCand, found := keeper.GetCandidate(ctx, addrVal1)
	require.True(t, found)
	assert.True(t, candidatesEqual(candidate1, resCand), "%v \n %v", resCand, candidate1)

	// modify a records, save, and retrieve
	candidate1.Liabilities = sdk.NewRat(99)
	keeper.setCandidate(ctx, candidate1)
	resCand, found = keeper.GetCandidate(ctx, addrVal1)
	require.True(t, found)
	assert.True(t, candidatesEqual(candidate1, resCand))

	// also test that the address has been added to address list
	resCands = keeper.GetCandidates(ctx, 100)
	require.Equal(t, 1, len(resCands))
	assert.Equal(t, addrVal1, resCands[0].Address)

	// add other candidates
	keeper.setCandidate(ctx, candidate2)
	keeper.setCandidate(ctx, candidate3)
	resCand, found = keeper.GetCandidate(ctx, addrVal2)
	require.True(t, found)
	assert.True(t, candidatesEqual(candidate2, resCand), "%v \n %v", resCand, candidate2)
	resCand, found = keeper.GetCandidate(ctx, addrVal3)
	require.True(t, found)
	assert.True(t, candidatesEqual(candidate3, resCand), "%v \n %v", resCand, candidate3)
	resCands = keeper.GetCandidates(ctx, 100)
	require.Equal(t, 3, len(resCands))
	assert.True(t, candidatesEqual(candidate1, resCands[0]), "%v \n %v", resCands[0], candidate1)
	assert.True(t, candidatesEqual(candidate2, resCands[1]), "%v \n %v", resCands[1], candidate2)
	assert.True(t, candidatesEqual(candidate3, resCands[2]), "%v \n %v", resCands[2], candidate3)

	// remove a record
	keeper.removeCandidate(ctx, candidate2.Address)
	_, found = keeper.GetCandidate(ctx, addrVal2)
	assert.False(t, found)
}

// tests GetDelegatorBond, GetDelegatorBonds, SetDelegatorBond, removeDelegatorBond
func TestBond(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	// first add a candidate1 to delegate too
	keeper.setCandidate(ctx, candidate1)

	bond1to1 := DelegatorBond{
		DelegatorAddr: addrDel1,
		CandidateAddr: addrVal1,
		Shares:        sdk.NewRat(9),
	}

	bondsEqual := func(b1, b2 DelegatorBond) bool {
		return bytes.Equal(b1.DelegatorAddr, b2.DelegatorAddr) &&
			bytes.Equal(b1.CandidateAddr, b2.CandidateAddr) &&
			b1.Shares == b2.Shares
	}

	// check the empty keeper first
	_, found := keeper.getDelegatorBond(ctx, addrDel1, addrVal1)
	assert.False(t, found)

	// set and retrieve a record
	keeper.setDelegatorBond(ctx, bond1to1)
	resBond, found := keeper.getDelegatorBond(ctx, addrDel1, addrVal1)
	assert.True(t, found)
	assert.True(t, bondsEqual(bond1to1, resBond))

	// modify a records, save, and retrieve
	bond1to1.Shares = sdk.NewRat(99)
	keeper.setDelegatorBond(ctx, bond1to1)
	resBond, found = keeper.getDelegatorBond(ctx, addrDel1, addrVal1)
	assert.True(t, found)
	assert.True(t, bondsEqual(bond1to1, resBond))

	// add some more records
	keeper.setCandidate(ctx, candidate2)
	keeper.setCandidate(ctx, candidate3)
	bond1to2 := DelegatorBond{addrDel1, addrVal2, sdk.NewRat(9)}
	bond1to3 := DelegatorBond{addrDel1, addrVal3, sdk.NewRat(9)}
	bond2to1 := DelegatorBond{addrDel2, addrVal1, sdk.NewRat(9)}
	bond2to2 := DelegatorBond{addrDel2, addrVal2, sdk.NewRat(9)}
	bond2to3 := DelegatorBond{addrDel2, addrVal3, sdk.NewRat(9)}
	keeper.setDelegatorBond(ctx, bond1to2)
	keeper.setDelegatorBond(ctx, bond1to3)
	keeper.setDelegatorBond(ctx, bond2to1)
	keeper.setDelegatorBond(ctx, bond2to2)
	keeper.setDelegatorBond(ctx, bond2to3)

	// test all bond retrieve capabilities
	resBonds := keeper.getDelegatorBonds(ctx, addrDel1, 5)
	require.Equal(t, 3, len(resBonds))
	assert.True(t, bondsEqual(bond1to1, resBonds[0]))
	assert.True(t, bondsEqual(bond1to2, resBonds[1]))
	assert.True(t, bondsEqual(bond1to3, resBonds[2]))
	resBonds = keeper.getDelegatorBonds(ctx, addrDel1, 3)
	require.Equal(t, 3, len(resBonds))
	resBonds = keeper.getDelegatorBonds(ctx, addrDel1, 2)
	require.Equal(t, 2, len(resBonds))
	resBonds = keeper.getDelegatorBonds(ctx, addrDel2, 5)
	require.Equal(t, 3, len(resBonds))
	assert.True(t, bondsEqual(bond2to1, resBonds[0]))
	assert.True(t, bondsEqual(bond2to2, resBonds[1]))
	assert.True(t, bondsEqual(bond2to3, resBonds[2]))

	// delete a record
	keeper.removeDelegatorBond(ctx, bond2to3)
	_, found = keeper.getDelegatorBond(ctx, addrDel2, addrVal3)
	assert.False(t, found)
	resBonds = keeper.getDelegatorBonds(ctx, addrDel2, 5)
	require.Equal(t, 2, len(resBonds))
	assert.True(t, bondsEqual(bond2to1, resBonds[0]))
	assert.True(t, bondsEqual(bond2to2, resBonds[1]))

	// delete all the records from delegator 2
	keeper.removeDelegatorBond(ctx, bond2to1)
	keeper.removeDelegatorBond(ctx, bond2to2)
	_, found = keeper.getDelegatorBond(ctx, addrDel2, addrVal1)
	assert.False(t, found)
	_, found = keeper.getDelegatorBond(ctx, addrDel2, addrVal2)
	assert.False(t, found)
	resBonds = keeper.getDelegatorBonds(ctx, addrDel2, 5)
	require.Equal(t, 0, len(resBonds))
}

// TODO integrate in testing for equal validators, whichever one was a validator
// first remains the validator https://github.com/cosmos/cosmos-sdk/issues/582
func TestGetValidators(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	// initialize some candidates into the state
	amts := []int64{0, 100, 1, 400, 200}
	n := len(amts)
	candidates := make([]Candidate, n)
	for i := 0; i < n; i++ {
		c := Candidate{
			Status:      Unbonded,
			PubKey:      pks[i],
			Address:     addrs[i],
			Assets:      sdk.NewRat(amts[i]),
			Liabilities: sdk.NewRat(amts[i]),
		}
		keeper.setCandidate(ctx, c)
		candidates[i] = c
	}

	// first make sure everything as normal is ordered
	validators := keeper.GetValidators(ctx)
	require.Equal(t, len(validators), n)
	assert.Equal(t, sdk.NewRat(400), validators[0].VotingPower, "%v", validators)
	assert.Equal(t, sdk.NewRat(200), validators[1].VotingPower, "%v", validators)
	assert.Equal(t, sdk.NewRat(100), validators[2].VotingPower, "%v", validators)
	assert.Equal(t, sdk.NewRat(1), validators[3].VotingPower, "%v", validators)
	assert.Equal(t, sdk.NewRat(0), validators[4].VotingPower, "%v", validators)
	assert.Equal(t, candidates[3].Address, validators[0].Address, "%v", validators)
	assert.Equal(t, candidates[4].Address, validators[1].Address, "%v", validators)
	assert.Equal(t, candidates[1].Address, validators[2].Address, "%v", validators)
	assert.Equal(t, candidates[2].Address, validators[3].Address, "%v", validators)
	assert.Equal(t, candidates[0].Address, validators[4].Address, "%v", validators)

	// test a basic increase in voting power
	candidates[3].Assets = sdk.NewRat(500)
	keeper.setCandidate(ctx, candidates[3])
	validators = keeper.GetValidators(ctx)
	require.Equal(t, len(validators), n)
	assert.Equal(t, sdk.NewRat(500), validators[0].VotingPower, "%v", validators)
	assert.Equal(t, candidates[3].Address, validators[0].Address, "%v", validators)

	// test a decrease in voting power
	candidates[3].Assets = sdk.NewRat(300)
	keeper.setCandidate(ctx, candidates[3])
	validators = keeper.GetValidators(ctx)
	require.Equal(t, len(validators), n)
	assert.Equal(t, sdk.NewRat(300), validators[0].VotingPower, "%v", validators)
	assert.Equal(t, candidates[3].Address, validators[0].Address, "%v", validators)

	// test a swap in voting power
	candidates[0].Assets = sdk.NewRat(600)
	keeper.setCandidate(ctx, candidates[0])
	validators = keeper.GetValidators(ctx)
	require.Equal(t, len(validators), n)
	assert.Equal(t, sdk.NewRat(600), validators[0].VotingPower, "%v", validators)
	assert.Equal(t, candidates[0].Address, validators[0].Address, "%v", validators)
	assert.Equal(t, sdk.NewRat(300), validators[1].VotingPower, "%v", validators)
	assert.Equal(t, candidates[3].Address, validators[1].Address, "%v", validators)

	// test the max validators term
	params := keeper.GetParams(ctx)
	n = 2
	params.MaxValidators = uint16(n)
	keeper.setParams(ctx, params)
	validators = keeper.GetValidators(ctx)
	require.Equal(t, len(validators), n)
	assert.Equal(t, sdk.NewRat(600), validators[0].VotingPower, "%v", validators)
	assert.Equal(t, candidates[0].Address, validators[0].Address, "%v", validators)
	assert.Equal(t, sdk.NewRat(300), validators[1].VotingPower, "%v", validators)
	assert.Equal(t, candidates[3].Address, validators[1].Address, "%v", validators)
}

// TODO
// test the mechanism which keeps track of a validator set change
func TestGetAccUpdateValidators(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	validatorsEqual := func(t *testing.T, expected []Validator, actual []Validator) {
		require.Equal(t, len(expected), len(actual))
		for i := 0; i < len(expected); i++ {
			assert.Equal(t, expected[i], actual[i])
		}
	}

	amts := []int64{100, 300}
	genCandidates := func(amts []int64) ([]Candidate, []Validator) {
		candidates := make([]Candidate, len(amts))
		validators := make([]Validator, len(amts))
		for i := 0; i < len(amts); i++ {
			c := Candidate{
				Status:      Unbonded,
				PubKey:      pks[i],
				Address:     addrs[i],
				Assets:      sdk.NewRat(amts[i]),
				Liabilities: sdk.NewRat(amts[i]),
			}
			candidates[i] = c
			validators[i] = c.validator()
		}
		return candidates, validators
	}

	candidates, validators := genCandidates(amts)

	//TODO
	// test from nothing to something
	acc := keeper.getAccUpdateValidators(ctx)
	assert.Equal(t, 0, len(acc))
	keeper.setCandidate(ctx, candidates[0])
	keeper.setCandidate(ctx, candidates[1])
	_ = keeper.GetValidators(ctx) // to init recent validator set
	acc = keeper.getAccUpdateValidators(ctx)
	validatorsEqual(t, validators, acc)

	// test identical
	keeper.setCandidate(ctx, candidates[0])
	keeper.setCandidate(ctx, candidates[1])
	acc = keeper.getAccUpdateValidators(ctx)
	validatorsEqual(t, validators, acc)

	acc = keeper.getAccUpdateValidators(ctx)

	// test from something to nothing
	keeper.removeCandidate(ctx, candidates[0].Address)
	keeper.removeCandidate(ctx, candidates[1].Address)
	acc = keeper.getAccUpdateValidators(ctx)
	assert.Equal(t, 2, len(acc))
	assert.Equal(t, validators[0].Address, acc[0].Address)
	assert.Equal(t, int64(0), acc[0].VotingPower.Evaluate())
	assert.Equal(t, validators[1].Address, acc[1].Address)
	assert.Equal(t, int64(0), acc[1].VotingPower.Evaluate())

	// test single value change
	amts[0] = 600
	candidates, validators = genCandidates(amts)
	keeper.setCandidate(ctx, candidates[0])
	keeper.setCandidate(ctx, candidates[1])
	acc = keeper.getAccUpdateValidators(ctx)
	validatorsEqual(t, validators, acc)

	// test multiple value change
	amts[0] = 200
	amts[1] = 0
	candidates, validators = genCandidates(amts)
	keeper.setCandidate(ctx, candidates[0])
	keeper.setCandidate(ctx, candidates[1])
	acc = keeper.getAccUpdateValidators(ctx)
	validatorsEqual(t, validators, acc)

	// test validator added at the beginning
	// test validator added in the middle
	// test validator added at the end
	amts = append(amts, 100)
	candidates, validators = genCandidates(amts)
	keeper.setCandidate(ctx, candidates[0])
	keeper.setCandidate(ctx, candidates[1])
	keeper.setCandidate(ctx, candidates[2])
	acc = keeper.getAccUpdateValidators(ctx)
	validatorsEqual(t, validators, acc)

	// test multiple validators removed
}

// clear the tracked changes to the validator set
func TestClearAccUpdateValidators(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	amts := []int64{0, 400}
	candidates := make([]Candidate, len(amts))
	for i, amt := range amts {
		c := Candidate{
			Status:      Unbonded,
			PubKey:      pks[i],
			Address:     addrs[i],
			Assets:      sdk.NewRat(amt),
			Liabilities: sdk.NewRat(amt),
		}
		candidates[i] = c
		keeper.setCandidate(ctx, c)
	}

	acc := keeper.getAccUpdateValidators(ctx)
	assert.Equal(t, len(amts), len(acc))
	keeper.clearAccUpdateValidators(ctx)
	acc = keeper.getAccUpdateValidators(ctx)
	assert.Equal(t, 0, len(acc))
}

// test if is a validator from the last update
func TestIsRecentValidator(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	// test that an empty validator set doesn't have any validators
	validators := keeper.GetValidators(ctx)
	assert.Equal(t, 0, len(validators))

	// get the validators for the first time
	keeper.setCandidate(ctx, candidate1)
	keeper.setCandidate(ctx, candidate2)
	validators = keeper.GetValidators(ctx)
	require.Equal(t, 2, len(validators))
	assert.Equal(t, candidate1.validator(), validators[0])
	assert.Equal(t, candidate2.validator(), validators[1])

	// test a basic retrieve of something that should be a recent validator
	assert.True(t, keeper.IsRecentValidator(ctx, candidate1.Address))
	assert.True(t, keeper.IsRecentValidator(ctx, candidate2.Address))

	// test a basic retrieve of something that should not be a recent validator
	assert.False(t, keeper.IsRecentValidator(ctx, candidate3.Address))

	// remove that validator, but don't retrieve the recent validator group
	keeper.removeCandidate(ctx, candidate1.Address)

	// test that removed validator is not considered a recent validator
	assert.False(t, keeper.IsRecentValidator(ctx, candidate1.Address))
}

func TestParams(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)
	expParams := defaultParams()

	//check that the empty keeper loads the default
	resParams := keeper.GetParams(ctx)
	assert.Equal(t, expParams, resParams)

	//modify a params, save, and retrieve
	expParams.MaxValidators = 777
	keeper.setParams(ctx, expParams)
	resParams = keeper.GetParams(ctx)
	assert.Equal(t, expParams, resParams)
}
