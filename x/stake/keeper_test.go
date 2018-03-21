package stake

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// XXX XXX XXX
// XXX revive these tests but for the store update proceedure
// XXX XXX XXX

//func TestUpdateVotingPower(t *testing.T) {
//assert := assert.New(t)
//store := initTestStore(t)
//params := getParams(store)
//gs := getGlobalState(store)

//N := 5
//actors := newAddrs(N)
//candidates := candidatesFromActors(actors, []int64{400, 200, 100, 10, 1})

//// test a basic change in voting power
//candidates[0].Assets = sdk.NewRat(500)
//candidates.updateVotingPower(store, gs, params)
//assert.Equal(int64(500), candidates[0].VotingPower.Evaluate(), "%v", candidates[0])

//// test a swap in voting power
//candidates[1].Assets = sdk.NewRat(600)
//candidates.updateVotingPower(store, gs, params)
//assert.Equal(int64(600), candidates[0].VotingPower.Evaluate(), "%v", candidates[0])
//assert.Equal(int64(500), candidates[1].VotingPower.Evaluate(), "%v", candidates[1])

//// test the max validators term
//params.MaxValidators = 4
//setParams(store, params)
//candidates.updateVotingPower(store, gs, params)
//assert.Equal(int64(0), candidates[4].VotingPower.Evaluate(), "%v", candidates[4])
//}

//func TestValidatorsChanged(t *testing.T) {
//require := require.New(t)

//v1 := (&Candidate{PubKey: pks[0], VotingPower: sdk.NewRat(10)}).validator()
//v2 := (&Candidate{PubKey: pks[1], VotingPower: sdk.NewRat(10)}).validator()
//v3 := (&Candidate{PubKey: pks[2], VotingPower: sdk.NewRat(10)}).validator()
//v4 := (&Candidate{PubKey: pks[3], VotingPower: sdk.NewRat(10)}).validator()
//v5 := (&Candidate{PubKey: pks[4], VotingPower: sdk.NewRat(10)}).validator()

//// test from nothing to something
//vs1 := []Validator{}
//vs2 := []Validator{v1, v2}
//changed := vs1.validatorsUpdated(vs2)
//require.Equal(2, len(changed))
//testChange(t, vs2[0], changed[0])
//testChange(t, vs2[1], changed[1])

//// test from something to nothing
//vs1 = []Validator{v1, v2}
//vs2 = []Validator{}
//changed = vs1.validatorsUpdated(vs2)
//require.Equal(2, len(changed))
//testRemove(t, vs1[0], changed[0])
//testRemove(t, vs1[1], changed[1])

//// test identical
//vs1 = []Validator{v1, v2, v4}
//vs2 = []Validator{v1, v2, v4}
//changed = vs1.validatorsUpdated(vs2)
//require.ZeroRat(len(changed))

//// test single value change
//vs2[2].VotingPower = sdk.OneRat
//changed = vs1.validatorsUpdated(vs2)
//require.Equal(1, len(changed))
//testChange(t, vs2[2], changed[0])

//// test multiple value change
//vs2[0].VotingPower = sdk.NewRat(11)
//vs2[2].VotingPower = sdk.NewRat(5)
//changed = vs1.validatorsUpdated(vs2)
//require.Equal(2, len(changed))
//testChange(t, vs2[0], changed[0])
//testChange(t, vs2[2], changed[1])

//// test validator added at the beginning
//vs1 = []Validator{v2, v4}
//vs2 = []Validator{v2, v4, v1}
//changed = vs1.validatorsUpdated(vs2)
//require.Equal(1, len(changed))
//testChange(t, vs2[0], changed[0])

//// test validator added in the middle
//vs1 = []Validator{v1, v2, v4}
//vs2 = []Validator{v3, v1, v4, v2}
//changed = vs1.validatorsUpdated(vs2)
//require.Equal(1, len(changed))
//testChange(t, vs2[2], changed[0])

//// test validator added at the end
//vs2 = []Validator{v1, v2, v4, v5}
//changed = vs1.validatorsUpdated(vs2)
//require.Equal(1, len(changed)) //testChange(t, vs2[3], changed[0]) //// test multiple validators added //vs2 = []Validator{v1, v2, v3, v4, v5} //changed = vs1.validatorsUpdated(vs2) //require.Equal(2, len(changed)) //testChange(t, vs2[2], changed[0]) //testChange(t, vs2[4], changed[1]) //// test validator removed at the beginning //vs2 = []Validator{v2, v4} //changed = vs1.validatorsUpdated(vs2) //require.Equal(1, len(changed)) //testRemove(t, vs1[0], changed[0]) //// test validator removed in the middle //vs2 = []Validator{v1, v4} //changed = vs1.validatorsUpdated(vs2) //require.Equal(1, len(changed)) //testRemove(t, vs1[1], changed[0]) //// test validator removed at the end
//vs2 = []Validator{v1, v2}
//changed = vs1.validatorsUpdated(vs2)
//require.Equal(1, len(changed))
//testRemove(t, vs1[2], changed[0])

//// test multiple validators removed
//vs2 = []Validator{v1}
//changed = vs1.validatorsUpdated(vs2)
//require.Equal(2, len(changed))
//testRemove(t, vs1[1], changed[0])
//testRemove(t, vs1[2], changed[1])

//// test many sdk of changes //vs2 = []Validator{v1, v3, v4, v5} //vs2[2].VotingPower = sdk.NewRat(11) //changed = vs1.validatorsUpdated(vs2) //require.Equal(4, len(changed), "%v", changed) // change 1, remove 1, add 2 //testRemove(t, vs1[1], changed[0]) //testChange(t, vs2[1], changed[1]) //testChange(t, vs2[2], changed[2]) //testChange(t, vs2[3], changed[3]) //} //func TestUpdateValidatorSet(t *testing.T) { //assert, require := assert.New(t), require.New(t) //store := initTestStore(t) //params := getParams(store) //gs := getGlobalState(store) //N := 5
//actors := newAddrs(N)
//candidates := candidatesFromActors(actors, []int64{400, 200, 100, 10, 1})
//for _, c := range candidates {
//setCandidate(store, c)
//}

//// they should all already be validators
//change, err := UpdateValidatorSet(store, gs, params)
//require.Nil(err)
//require.Equal(0, len(change), "%v", change) // change 1, remove 1, add 2

//// test the max value and test again
//params.MaxValidators = 4
//setParams(store, params)
//change, err = UpdateValidatorSet(store, gs, params)
//require.Nil(err)
//require.Equal(1, len(change), "%v", change)
//testRemove(t, candidates[4].validator(), change[0])
//candidates = getCandidates(store)
//assert.Equal(int64(0), candidates[4].VotingPower.Evaluate())

//// mess with the power's of the candidates and test
//candidates[0].Assets = sdk.NewRat(10)
//candidates[1].Assets = sdk.NewRat(600)
//candidates[2].Assets = sdk.NewRat(1000)
//candidates[3].Assets = sdk.OneRat
//candidates[4].Assets = sdk.NewRat(10)
//for _, c := range candidates {
//setCandidate(store, c)
//}
//change, err = UpdateValidatorSet(store, gs, params)
//require.Nil(err)
//require.Equal(5, len(change), "%v", change) // 3 changed, 1 added, 1 removed
//candidates = getCandidates(store)
//testChange(t, candidates[0].validator(), change[0])
//testChange(t, candidates[1].validator(), change[1])
//testChange(t, candidates[2].validator(), change[2])
//testRemove(t, candidates[3].validator(), change[3])
//testChange(t, candidates[4].validator(), change[4])
//}

func TestState(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	addrDel := sdk.Address([]byte("addressdelegator"))
	addrVal := sdk.Address([]byte("addressvalidator"))
	//pk := newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB57")
	pk := crypto.GenPrivKeyEd25519().PubKey()

	//----------------------------------------------------------------------
	// Candidate checks

	// XXX expand to include both liabilities and assets use/test all candidate fields
	candidate := Candidate{
		Address:     addrVal,
		PubKey:      pk,
		Assets:      sdk.NewRat(9),
		Liabilities: sdk.NewRat(9),
		VotingPower: sdk.ZeroRat,
	}

	candidatesEqual := func(c1, c2 Candidate) bool {
		return c1.Status == c2.Status &&
			c1.PubKey.Equals(c2.PubKey) &&
			bytes.Equal(c1.Address, c2.Address) &&
			c1.Assets.Equal(c2.Assets) &&
			c1.Liabilities.Equal(c2.Liabilities) &&
			c1.VotingPower.Equal(c2.VotingPower) &&
			c1.Description == c2.Description
	}

	// check the empty keeper first
	_, found := keeper.getCandidate(ctx, addrVal)
	assert.False(t, found)
	resPks := keeper.getCandidates(ctx)
	assert.Zero(t, len(resPks))

	// set and retrieve a record
	keeper.setCandidate(ctx, candidate)
	resCand, found := keeper.getCandidate(ctx, addrVal)
	assert.True(t, found)
	assert.True(t, candidatesEqual(candidate, resCand), "%v \n %v", resCand, candidate)

	// modify a records, save, and retrieve
	candidate.Liabilities = sdk.NewRat(99)
	keeper.setCandidate(ctx, candidate)
	resCand, found = keeper.getCandidate(ctx, addrVal)
	assert.True(t, found)
	assert.True(t, candidatesEqual(candidate, resCand))

	// also test that the pubkey has been added to pubkey list
	resPks = keeper.getCandidates(ctx)
	require.Equal(t, 1, len(resPks))
	assert.Equal(t, addrVal, resPks[0].PubKey)

	//----------------------------------------------------------------------
	// Bond checks

	bond := DelegatorBond{
		DelegatorAddr: addrDel,
		CandidateAddr: addrVal,
		Shares:        sdk.NewRat(9),
	}

	bondsEqual := func(b1, b2 DelegatorBond) bool {
		return bytes.Equal(b1.DelegatorAddr, b2.DelegatorAddr) &&
			bytes.Equal(b1.CandidateAddr, b2.CandidateAddr) &&
			b1.Shares == b2.Shares
	}

	//check the empty keeper first
	_, found = keeper.getDelegatorBond(ctx, addrDel, addrVal)
	assert.False(t, found)

	//Set and retrieve a record
	keeper.setDelegatorBond(ctx, bond)
	resBond, found := keeper.getDelegatorBond(ctx, addrDel, addrVal)
	assert.True(t, found)
	assert.True(t, bondsEqual(bond, resBond))

	//modify a records, save, and retrieve
	bond.Shares = sdk.NewRat(99)
	keeper.setDelegatorBond(ctx, bond)
	resBond, found = keeper.getDelegatorBond(ctx, addrDel, addrVal)
	assert.True(t, found)
	assert.True(t, bondsEqual(bond, resBond))

	//----------------------------------------------------------------------
	// Param checks

	params := defaultParams()

	//check that the empty keeper loads the default
	resParams := keeper.getParams(ctx)
	assert.Equal(t, params, resParams)

	//modify a params, save, and retrieve
	params.MaxValidators = 777
	keeper.setParams(ctx, params)
	resParams = keeper.getParams(ctx)
	assert.Equal(t, params, resParams)
}

func TestGetValidators(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)
	candidatesFromAddrs(ctx, keeper, addrs, []int64{400, 200, 0, 0, 0})

	validators := keeper.getValidators(ctx, 5)
	require.Equal(t, 2, len(validators))
	assert.Equal(t, addrs[0], validators[0].Address)
	assert.Equal(t, addrs[1], validators[1].Address)
}
