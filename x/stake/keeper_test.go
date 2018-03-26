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
		Assets:      sdk.NewRat(9),
		Liabilities: sdk.NewRat(9),
	}
	candidate3 = Candidate{
		Address:     addrVal3,
		PubKey:      pk3,
		Assets:      sdk.NewRat(9),
		Liabilities: sdk.NewRat(9),
	}
)

/*
func TestUpdateVotingPower(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	// initialize some candidates into the state
	amts := []int64{400, 200, 100, 10, 1}
	candidates := make([]Candidate, 5)
	for i := 0; i < 5; i++ {
		c := Candidate{
			Status:      Unbonded,
			PubKey:      pks[i],
			Address:     addrs[i],
			Assets:      sdk.NewRat(amts[i]),
			Liabilities: sdk.NewRat(amts[i]),
		}
		keeper.setCandidate(ctx, c)
		candidate[i] = c
	}

	// test a basic change in voting power
	candidates[0].Assets = sdk.NewRat(500)
	keeper.setCandidate(ctx, candidate[0])
	validators

	assert.Equal(int64(500), candidates[0].VotingPower.Evaluate(), "%v", candidates[0])

	// test a swap in voting power
	candidates[1].Assets = sdk.NewRat(600)
	candidates.updateVotingPower(store, p, params)
	assert.Equal(int64(600), candidates[0].VotingPower.Evaluate(), "%v", candidates[0])
	assert.Equal(int64(500), candidates[1].VotingPower.Evaluate(), "%v", candidates[1])

	// test the max validators term
	params.MaxValidators = 4
	setParams(store, params)
	candidates.updateVotingPower(store, p, params)
	assert.Equal(int64(0), candidates[4].VotingPower.Evaluate(), "%v", candidates[4])
}

func TestValidatorsChanged(t *testing.T) {
	require := require.New(t)

	v1 := (&Candidate{PubKey: pks[0], VotingPower: sdk.NewRat(10)}).validator()
	v2 := (&Candidate{PubKey: pks[1], VotingPower: sdk.NewRat(10)}).validator()
	v3 := (&Candidate{PubKey: pks[2], VotingPower: sdk.NewRat(10)}).validator()
	v4 := (&Candidate{PubKey: pks[3], VotingPower: sdk.NewRat(10)}).validator()
	v5 := (&Candidate{PubKey: pks[4], VotingPower: sdk.NewRat(10)}).validator()

	// test from nothing to something
	vs1 := []Validator{}
	vs2 := []Validator{v1, v2}
	changed := vs1.validatorsUpdated(vs2)
	require.Equal(2, len(changed))
	testChange(t, vs2[0], changed[0])
	testChange(t, vs2[1], changed[1])

	// test from something to nothing
	vs1 = []Validator{v1, v2}
	vs2 = []Validator{}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(2, len(changed))
	testRemove(t, vs1[0], changed[0])
	testRemove(t, vs1[1], changed[1])

	// test identical
	vs1 = []Validator{v1, v2, v4}
	vs2 = []Validator{v1, v2, v4}
	changed = vs1.validatorsUpdated(vs2)
	require.ZeroRat(len(changed))

	// test single value change
	vs2[2].VotingPower = sdk.OneRat
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(1, len(changed))
	testChange(t, vs2[2], changed[0])

	// test multiple value change
	vs2[0].VotingPower = sdk.NewRat(11)
	vs2[2].VotingPower = sdk.NewRat(5)
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(2, len(changed))
	testChange(t, vs2[0], changed[0])
	testChange(t, vs2[2], changed[1])

	// test validator added at the beginning
	vs1 = []Validator{v2, v4}
	vs2 = []Validator{v2, v4, v1}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(1, len(changed))
	testChange(t, vs2[0], changed[0])

	// test validator added in the middle
	vs1 = []Validator{v1, v2, v4}
	vs2 = []Validator{v3, v1, v4, v2}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(1, len(changed))
	testChange(t, vs2[2], changed[0])

	// test validator added at the end
	vs2 = []Validator{v1, v2, v4, v5}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(1, len(changed)) //testChange(t, vs2[3], changed[0]) //// test multiple validators added //vs2 = []Validator{v1, v2, v3, v4, v5} //changed = vs1.validatorsUpdated(vs2) //require.Equal(2, len(changed)) //testChange(t, vs2[2], changed[0]) //testChange(t, vs2[4], changed[1]) //// test validator removed at the beginning //vs2 = []Validator{v2, v4} //changed = vs1.validatorsUpdated(vs2) //require.Equal(1, len(changed)) //testRemove(t, vs1[0], changed[0]) //// test validator removed in the middle //vs2 = []Validator{v1, v4} //changed = vs1.validatorsUpdated(vs2) //require.Equal(1, len(changed)) //testRemove(t, vs1[1], changed[0]) //// test validator removed at the end
	vs2 = []Validator{v1, v2}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(1, len(changed))
	testRemove(t, vs1[2], changed[0])

	// test multiple validators removed
	vs2 = []Validator{v1}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(2, len(changed))
	testRemove(t, vs1[1], changed[0])
	testRemove(t, vs1[2], changed[1])

	// test many sdk of changes //vs2 = []Validator{v1, v3, v4, v5} //vs2[2].VotingPower = sdk.NewRat(11) //changed = vs1.validatorsUpdated(vs2) //require.Equal(4, len(changed), "%v", changed) // change 1, remove 1, add 2 //testRemove(t, vs1[1], changed[0]) //testChange(t, vs2[1], changed[1]) //testChange(t, vs2[2], changed[2]) //testChange(t, vs2[3], changed[3]) //} //func TestUpdateValidatorSet(t *testing.T) { //assert, require := assert.New(t), require.New(t) //store := initTestStore(t) //params := GetParams(store) //gs := GetPool(store) //N := 5
	actors := newAddrs(N)
	candidates := candidatesFromActors(actors, []int64{400, 200, 100, 10, 1})
	for _, c := range candidates {
		setCandidate(store, c)
	}

	// they should all already be validators
	change, err := UpdateValidatorSet(store, p, params)
	require.Nil(err)
	require.Equal(0, len(change), "%v", change) // change 1, remove 1, add 2

	// test the max value and test again
	params.MaxValidators = 4
	setParams(store, params)
	change, err = UpdateValidatorSet(store, p, params)
	require.Nil(err)
	require.Equal(1, len(change), "%v", change)
	testRemove(t, candidates[4].validator(), change[0])
	candidates = GetCandidates(store)
	assert.Equal(int64(0), candidates[4].VotingPower.Evaluate())

	// mess with the power's of the candidates and test
	candidates[0].Assets = sdk.NewRat(10)
	candidates[1].Assets = sdk.NewRat(600)
	candidates[2].Assets = sdk.NewRat(1000)
	candidates[3].Assets = sdk.OneRat
	candidates[4].Assets = sdk.NewRat(10)
	for _, c := range candidates {
		setCandidate(store, c)
	}
	change, err = UpdateValidatorSet(store, p, params)
	require.Nil(err)
	require.Equal(5, len(change), "%v", change) // 3 changed, 1 added, 1 removed
	candidates = GetCandidates(store)
	testChange(t, candidates[0].validator(), change[0])
	testChange(t, candidates[1].validator(), change[1])
	testChange(t, candidates[2].validator(), change[2])
	testRemove(t, candidates[3].validator(), change[3])
	testChange(t, candidates[4].validator(), change[4])
}
*/

// XXX BROKEN TEST
func TestGetValidators(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)
	params := keeper.GetParams(ctx)
	params.MaxValidators = 2
	keeper.setParams(ctx, params)
	candidatesFromAddrs(ctx, keeper, addrs, []int64{0, 0, 0, 400, 200, 0}) // XXX rearrange these something messed is happenning!

	validators := keeper.GetValidators(ctx)
	require.Equal(t, 2, len(validators))
	assert.Equal(t, addrs[0], validators[0].Address, "%v", validators)
	assert.Equal(t, addrs[1], validators[1].Address, "%v", validators)
}

// XXX expand to include both liabilities and assets use/test all candidate1 fields
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
	resAddrs := keeper.GetCandidates(ctx, 100)
	assert.Zero(t, len(resAddrs))

	// set and retrieve a record
	keeper.setCandidate(ctx, candidate1)
	resCand, found := keeper.GetCandidate(ctx, addrVal1)
	assert.True(t, found)
	assert.True(t, candidatesEqual(candidate1, resCand), "%v \n %v", resCand, candidate1)

	// modify a records, save, and retrieve
	candidate1.Liabilities = sdk.NewRat(99)
	keeper.setCandidate(ctx, candidate1)
	resCand, found = keeper.GetCandidate(ctx, addrVal1)
	assert.True(t, found)
	assert.True(t, candidatesEqual(candidate1, resCand))

	// also test that the address has been added to address list
	resAddrs = keeper.GetCandidates(ctx, 100)
	require.Equal(t, 1, len(resAddrs))
	assert.Equal(t, addrVal1, resAddrs[0].Address)

}

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
