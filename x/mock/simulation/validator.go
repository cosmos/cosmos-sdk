package simulation

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"
)

type mockValidator struct {
	val           abci.ValidatorUpdate
	livenessState int
}

// TODO describe usage
func getKeys(validators map[string]mockValidator) []string {
	keys := make([]string, len(validators))
	i := 0
	for key := range validators {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

// randomProposer picks a random proposer from the current validator set
func randomProposer(r *rand.Rand, validators map[string]mockValidator) cmn.HexBytes {
	keys := getKeys(validators)
	if len(keys) == 0 {
		return nil
	}
	key := keys[r.Intn(len(keys))]
	proposer := validators[key].val
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
