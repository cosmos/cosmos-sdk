package simulation

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"
)

type mockValidator struct {
	val           abci.ValidatorUpdate
	livenessState int
}

func (mv mockValidator) String() string {
	return fmt.Sprintf("mockValidator{%s:%X power:%v state:%v}",
		mv.val.PubKey.Type,
		mv.val.PubKey.Data,
		mv.val.Power,
		mv.livenessState)
}

type mockValidators map[string]mockValidator

// get mockValidators from abci validators
func newMockValidators(r *rand.Rand, abciVals []abci.ValidatorUpdate,
	params Params) mockValidators {

	validators := make(mockValidators)
	for _, validator := range abciVals {
		str := fmt.Sprintf("%v", validator.PubKey)
		liveliness := GetMemberOfInitialState(r,
			params.InitialLivenessWeightings)

		validators[str] = mockValidator{
			val:           validator,
			livenessState: liveliness,
		}
	}

	return validators
}

// TODO describe usage
func (vals mockValidators) getKeys() []string {
	keys := make([]string, len(vals))
	i := 0
	for key := range vals {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

//_________________________________________________________________________________

// randomProposer picks a random proposer from the current validator set
func (vals mockValidators) randomProposer(r *rand.Rand) cmn.HexBytes {
	keys := vals.getKeys()
	if len(keys) == 0 {
		return nil
	}
	key := keys[r.Intn(len(keys))]
	proposer := vals[key].val
	pk, err := tmtypes.PB2TM.PubKey(proposer.PubKey)
	if err != nil {
		panic(err)
	}
	return pk.Address()
}

// updateValidators mimicks Tendermint's update logic
// nolint: unparam
func updateValidators(tb testing.TB, r *rand.Rand, params Params,
	current map[string]mockValidator, updates []abci.ValidatorUpdate,
	event func(string)) map[string]mockValidator {

	for _, update := range updates {
		str := fmt.Sprintf("%v", update.PubKey)

		if update.Power == 0 {
			if _, ok := current[str]; !ok {
				tb.Fatalf("tried to delete a nonexistent validator")
			}
			event("endblock/validatorupdates/kicked")
			delete(current, str)

		} else if mVal, ok := current[str]; ok {
			// validator already exists
			mVal.val = update
			event("endblock/validatorupdates/updated")

		} else {
			// Set this new validator
			current[str] = mockValidator{
				update,
				GetMemberOfInitialState(r, params.InitialLivenessWeightings),
			}
			event("endblock/validatorupdates/added")
		}
	}

	return current
}

// RandomRequestBeginBlock generates a list of signing validators according to
// the provided list of validators, signing fraction, and evidence fraction
func RandomRequestBeginBlock(r *rand.Rand, params Params,
	validators mockValidators, pastTimes []time.Time,
	pastVoteInfos [][]abci.VoteInfo,
	event func(string), header abci.Header) abci.RequestBeginBlock {

	if len(validators) == 0 {
		return abci.RequestBeginBlock{
			Header: header,
		}
	}

	voteInfos := make([]abci.VoteInfo, len(validators))
	for i, key := range validators.getKeys() {
		mVal := validators[key]
		mVal.livenessState = params.LivenessTransitionMatrix.NextState(r, mVal.livenessState)
		signed := true

		if mVal.livenessState == 1 {
			// spotty connection, 50% probability of success
			// See https://github.com/golang/go/issues/23804#issuecomment-365370418
			// for reasoning behind computing like this
			signed = r.Int63()%2 == 0
		} else if mVal.livenessState == 2 {
			// offline
			signed = false
		}

		if signed {
			event("beginblock/signing/signed")
		} else {
			event("beginblock/signing/missed")
		}

		pubkey, err := tmtypes.PB2TM.PubKey(mVal.val.PubKey)
		if err != nil {
			panic(err)
		}
		voteInfos[i] = abci.VoteInfo{
			Validator: abci.Validator{
				Address: pubkey.Address(),
				Power:   mVal.val.Power,
			},
			SignedLastBlock: signed,
		}
	}

	// return if no past times
	if len(pastTimes) <= 0 {
		return abci.RequestBeginBlock{
			Header: header,
			LastCommitInfo: abci.LastCommitInfo{
				Votes: voteInfos,
			},
		}
	}

	// TODO: Determine capacity before allocation
	evidence := make([]abci.Evidence, 0)
	for r.Float64() < params.EvidenceFraction {

		height := header.Height
		time := header.Time
		vals := voteInfos

		if r.Float64() < params.PastEvidenceFraction && header.Height > 1 {
			height = int64(r.Intn(int(header.Height)-1)) + 1 // Tendermint starts at height 1
			// array indices offset by one
			time = pastTimes[height-1]
			vals = pastVoteInfos[height-1]
		}
		validator := vals[r.Intn(len(vals))].Validator

		var totalVotingPower int64
		for _, val := range vals {
			totalVotingPower += val.Validator.Power
		}

		evidence = append(evidence,
			abci.Evidence{
				Type:             tmtypes.ABCIEvidenceTypeDuplicateVote,
				Validator:        validator,
				Height:           height,
				Time:             time,
				TotalVotingPower: totalVotingPower,
			},
		)
		event("beginblock/evidence")
	}

	return abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Votes: voteInfos,
		},
		ByzantineValidators: evidence,
	}
}
