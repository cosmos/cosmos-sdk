package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tmlibs/rational"

	"github.com/cosmos/cosmos-sdk/state"
)

func TestCandidatesSort(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	N := 5
	actors := newActors(N)
	candidates := candidatesFromActors(actors, []int64{10, 300, 123, 4, 200})
	expectedOrder := []int{1, 4, 2, 0, 3}

	// test basic sort
	candidates.Sort()

	vals := candidates.Validators()
	require.Equal(N, len(vals))

	for i, val := range vals {
		expectedIdx := expectedOrder[i]
		assert.Equal(val.PubKey, pks[expectedIdx])
	}
}

func TestValidatorsSort(t *testing.T) {
	assert := assert.New(t)

	v1 := (&Candidate{PubKey: pks[0], VotingPower: rational.New(25)}).validator()
	v2 := (&Candidate{PubKey: pks[1], VotingPower: rational.New(1234)}).validator()
	v3 := (&Candidate{PubKey: pks[2], VotingPower: rational.New(122)}).validator()
	v4 := (&Candidate{PubKey: pks[3], VotingPower: rational.New(13)}).validator()
	v5 := (&Candidate{PubKey: pks[4], VotingPower: rational.New(1111)}).validator()

	// test from nothing to something
	vs := Validators{v4, v2, v5, v1, v3}

	// test basic sort
	vs.Sort()

	for i, v := range vs {
		assert.True(v.PubKey.Equals(pks[i]))
	}
}

func TestUpdateVotingPower(t *testing.T) {
	assert := assert.New(t)
	store := state.NewMemKVStore()
	params := loadParams(store)
	gs := loadGlobalState(store)

	N := 5
	actors := newActors(N)
	candidates := candidatesFromActors(actors, []int64{400, 200, 100, 10, 1})

	// test a basic change in voting power
	candidates[0].Assets = rational.New(500)
	candidates.updateVotingPower(store, gs, params)
	assert.Equal(int64(500), candidates[0].VotingPower.Evaluate(), "%v", candidates[0])

	// test a swap in voting power
	candidates[1].Assets = rational.New(600)
	candidates.updateVotingPower(store, gs, params)
	assert.Equal(int64(600), candidates[0].VotingPower.Evaluate(), "%v", candidates[0])
	assert.Equal(int64(500), candidates[1].VotingPower.Evaluate(), "%v", candidates[1])

	// test the max validators term
	params.MaxVals = 4
	saveParams(store, params)
	candidates.updateVotingPower(store, gs, params)
	assert.Equal(int64(0), candidates[4].VotingPower.Evaluate(), "%v", candidates[4])
}

func TestGetValidators(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	N := 5
	actors := newActors(N)
	candidates := candidatesFromActors(actors, []int64{400, 200, 0, 0, 0})

	validators := candidates.Validators()
	require.Equal(2, len(validators))
	assert.Equal(candidates[0].PubKey, validators[0].PubKey)
	assert.Equal(candidates[1].PubKey, validators[1].PubKey)
}

func TestValidatorsChanged(t *testing.T) {
	require := require.New(t)

	v1 := (&Candidate{PubKey: pks[0], VotingPower: rational.New(10)}).validator()
	v2 := (&Candidate{PubKey: pks[1], VotingPower: rational.New(10)}).validator()
	v3 := (&Candidate{PubKey: pks[2], VotingPower: rational.New(10)}).validator()
	v4 := (&Candidate{PubKey: pks[3], VotingPower: rational.New(10)}).validator()
	v5 := (&Candidate{PubKey: pks[4], VotingPower: rational.New(10)}).validator()

	// test from nothing to something
	vs1 := Validators{}
	vs2 := Validators{v1, v2}
	changed := vs1.validatorsUpdated(vs2)
	require.Equal(2, len(changed))
	testChange(t, vs2[0], changed[0])
	testChange(t, vs2[1], changed[1])

	// test from something to nothing
	vs1 = Validators{v1, v2}
	vs2 = Validators{}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(2, len(changed))
	testRemove(t, vs1[0], changed[0])
	testRemove(t, vs1[1], changed[1])

	// test identical
	vs1 = Validators{v1, v2, v4}
	vs2 = Validators{v1, v2, v4}
	changed = vs1.validatorsUpdated(vs2)
	require.Zero(len(changed))

	// test single value change
	vs2[2].VotingPower = rational.One
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(1, len(changed))
	testChange(t, vs2[2], changed[0])

	// test multiple value change
	vs2[0].VotingPower = rational.New(11)
	vs2[2].VotingPower = rational.New(5)
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(2, len(changed))
	testChange(t, vs2[0], changed[0])
	testChange(t, vs2[2], changed[1])

	// test validator added at the beginning
	vs1 = Validators{v2, v4}
	vs2 = Validators{v2, v4, v1}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(1, len(changed))
	testChange(t, vs2[0], changed[0])

	// test validator added in the middle
	vs1 = Validators{v1, v2, v4}
	vs2 = Validators{v3, v1, v4, v2}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(1, len(changed))
	testChange(t, vs2[2], changed[0])

	// test validator added at the end
	vs2 = Validators{v1, v2, v4, v5}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(1, len(changed))
	testChange(t, vs2[3], changed[0])

	// test multiple validators added
	vs2 = Validators{v1, v2, v3, v4, v5}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(2, len(changed))
	testChange(t, vs2[2], changed[0])
	testChange(t, vs2[4], changed[1])

	// test validator removed at the beginning
	vs2 = Validators{v2, v4}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(1, len(changed))
	testRemove(t, vs1[0], changed[0])

	// test validator removed in the middle
	vs2 = Validators{v1, v4}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(1, len(changed))
	testRemove(t, vs1[1], changed[0])

	// test validator removed at the end
	vs2 = Validators{v1, v2}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(1, len(changed))
	testRemove(t, vs1[2], changed[0])

	// test multiple validators removed
	vs2 = Validators{v1}
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(2, len(changed))
	testRemove(t, vs1[1], changed[0])
	testRemove(t, vs1[2], changed[1])

	// test many types of changes
	vs2 = Validators{v1, v3, v4, v5}
	vs2[2].VotingPower = rational.New(11)
	changed = vs1.validatorsUpdated(vs2)
	require.Equal(4, len(changed), "%v", changed) // change 1, remove 1, add 2
	testRemove(t, vs1[1], changed[0])
	testChange(t, vs2[1], changed[1])
	testChange(t, vs2[2], changed[2])
	testChange(t, vs2[3], changed[3])

}

func TestUpdateValidatorSet(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	store := state.NewMemKVStore()
	params := loadParams(store)
	gs := loadGlobalState(store)

	N := 5
	actors := newActors(N)
	candidates := candidatesFromActors(actors, []int64{400, 200, 100, 10, 1})
	for _, c := range candidates {
		saveCandidate(store, c)
	}

	// they should all already be validators
	change, err := UpdateValidatorSet(store, gs, params)
	require.Nil(err)
	require.Equal(0, len(change), "%v", change) // change 1, remove 1, add 2

	// test the max value and test again
	params.MaxVals = 4
	saveParams(store, params)
	change, err = UpdateValidatorSet(store, gs, params)
	require.Nil(err)
	require.Equal(1, len(change), "%v", change)
	testRemove(t, candidates[4].validator(), change[0])
	candidates = loadCandidates(store)
	assert.Equal(int64(0), candidates[4].VotingPower.Evaluate())

	// mess with the power's of the candidates and test
	candidates[0].Assets = rational.New(10)
	candidates[1].Assets = rational.New(600)
	candidates[2].Assets = rational.New(1000)
	candidates[3].Assets = rational.One
	candidates[4].Assets = rational.New(10)
	for _, c := range candidates {
		saveCandidate(store, c)
	}
	change, err = UpdateValidatorSet(store, gs, params)
	require.Nil(err)
	require.Equal(5, len(change), "%v", change) // 3 changed, 1 added, 1 removed
	candidates = loadCandidates(store)
	testChange(t, candidates[0].validator(), change[0])
	testChange(t, candidates[1].validator(), change[1])
	testChange(t, candidates[2].validator(), change[2])
	testRemove(t, candidates[3].validator(), change[3])
	testChange(t, candidates[4].validator(), change[4])
}

// XXX test global state functions, candidate exchange rate functions etc.
