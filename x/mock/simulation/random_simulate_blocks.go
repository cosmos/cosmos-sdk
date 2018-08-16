package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
)

// Simulate tests application by sending random messages.
func Simulate(
	t *testing.T, app *baseapp.BaseApp, appStateFn func(r *rand.Rand, keys []crypto.PrivKey, accs []sdk.AccAddress) json.RawMessage, ops []TestAndRunTx, setups []RandSetup,
	invariants []Invariant, numBlocks int, blockSize int, commit bool,
) {
	time := time.Now().UnixNano()
	SimulateFromSeed(t, app, appStateFn, time, ops, setups, invariants, numBlocks, blockSize, commit)
}

// SimulateFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided seed.
func SimulateFromSeed(
	t *testing.T, app *baseapp.BaseApp, appStateFn func(r *rand.Rand, keys []crypto.PrivKey, accs []sdk.AccAddress) json.RawMessage, seed int64, ops []TestAndRunTx, setups []RandSetup,
	invariants []Invariant, numBlocks int, blockSize int, commit bool,
) {
	log := fmt.Sprintf("Starting SimulateFromSeed with randomness created with seed %d", int(seed))
	fmt.Printf("%s\n", log)
	r := rand.New(rand.NewSource(seed))
	keys, accs := mock.GeneratePrivKeyAddressPairsFromRand(r, numKeys)

	// Setup event stats
	events := make(map[string]uint)
	event := func(what string) {
		log += "\nevent - " + what
		events[what]++
	}

	timestamp := time.Unix(0, 0)
	timeDiff := maxTimePerBlock - minTimePerBlock

	res := app.InitChain(abci.RequestInitChain{AppStateBytes: appStateFn(r, keys, accs)})
	validators := make(map[string]mockValidator)
	for _, validator := range res.Validators {
		validators[string(validator.Address)] = mockValidator{validator, GetMemberOfInitialState(r, initialLivenessWeightings)}
	}

	for i := 0; i < len(setups); i++ {
		setups[i](r, keys)
	}

	header := abci.Header{Height: 0, Time: timestamp}
	opCount := 0

	request := abci.RequestBeginBlock{Header: header}

	var pastTimes []time.Time

	for i := 0; i < numBlocks; i++ {

		// Log the header time for future lookup
		pastTimes = append(pastTimes, header.Time)

		// Run the BeginBlock handler
		app.BeginBlock(request)

		log += "\nBeginBlock"

		// Make sure invariants hold at beginning of block
		AssertAllInvariants(t, app, invariants, log)

		ctx := app.NewContext(false, header)

		var thisBlockSize int
		load := r.Float64()
		switch {
		case load < 0.33:
			thisBlockSize = 0
		case load < 0.66:
			thisBlockSize = r.Intn(blockSize * 2)
		default:
			thisBlockSize = r.Intn(blockSize * 4)
		}
		for j := 0; j < thisBlockSize; j++ {
			logUpdate, err := ops[r.Intn(len(ops))](t, r, app, ctx, keys, log, event)
			log += "\n" + logUpdate

			require.Nil(t, err, log)
			if onOperation {
				AssertAllInvariants(t, app, invariants, log)
			}
			if opCount%200 == 0 {
				fmt.Printf("\rSimulating... block %d/%d, operation %d.", header.Height, numBlocks, opCount)
			}
			opCount++
		}

		res := app.EndBlock(abci.RequestEndBlock{})
		header.Height++
		header.Time = header.Time.Add(time.Duration(minTimePerBlock) * time.Second).Add(time.Duration(int64(r.Intn(int(timeDiff)))) * time.Second)

		log += "\nEndBlock"

		// Make sure invariants hold at end of block
		AssertAllInvariants(t, app, invariants, log)

		if commit {
			app.Commit()
		}

		// Generate a random RequestBeginBlock with the current validator set for the next block
		request = RandomRequestBeginBlock(t, r, validators, livenessTransitionMatrix, evidenceFraction, pastTimes, event, header, log)

		// Update the validator set
		validators = updateValidators(t, r, validators, res.ValidatorUpdates, event)
	}

	fmt.Printf("\nSimulation complete. Final height (blocks): %d, final time (seconds): %v\n", header.Height, header.Time)
	DisplayEvents(events)
}

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

// RandomRequestBeginBlock generates a list of signing validators according to the provided list of validators, signing fraction, and evidence fraction
func RandomRequestBeginBlock(t *testing.T, r *rand.Rand, validators map[string]mockValidator, livenessTransitions TransitionMatrix, evidenceFraction float64,
	pastTimes []time.Time, event func(string), header abci.Header, log string) abci.RequestBeginBlock {
	if len(validators) == 0 {
		return abci.RequestBeginBlock{Header: header}
	}
	signingValidators := make([]abci.SigningValidator, len(validators))
	i := 0

	for _, key := range getKeys(validators) {
		mVal := validators[key]
		mVal.livenessState = livenessTransitions.NextState(r, mVal.livenessState)
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
		signingValidators[i] = abci.SigningValidator{
			Validator:       mVal.val,
			SignedLastBlock: signed,
		}
		i++
	}
	evidence := make([]abci.Evidence, 0)
	for r.Float64() < evidenceFraction {
		height := header.Height
		time := header.Time
		if r.Float64() < pastEvidenceFraction {
			height = int64(r.Intn(int(header.Height)))
			time = pastTimes[height]
		}
		validator := signingValidators[r.Intn(len(signingValidators))].Validator
		var currentTotalVotingPower int64
		for _, mVal := range validators {
			currentTotalVotingPower += mVal.val.Power
		}
		evidence = append(evidence, abci.Evidence{
			Type:             tmtypes.ABCIEvidenceTypeDuplicateVote,
			Validator:        validator,
			Height:           height,
			Time:             time,
			TotalVotingPower: currentTotalVotingPower,
		})
		event("beginblock/evidence")
	}
	return abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Validators: signingValidators,
		},
		ByzantineValidators: evidence,
	}
}

// AssertAllInvariants asserts a list of provided invariants against application state
func AssertAllInvariants(t *testing.T, app *baseapp.BaseApp, tests []Invariant, log string) {
	for i := 0; i < len(tests); i++ {
		tests[i](t, app, log)
	}
}

// updateValidators mimicks Tendermint's update logic
func updateValidators(t *testing.T, r *rand.Rand, current map[string]mockValidator, updates []abci.Validator, event func(string)) map[string]mockValidator {
	for _, update := range updates {
		switch {
		case update.Power == 0:
			require.NotNil(t, current[string(update.PubKey.Data)], "tried to delete a nonexistent validator")
			event("endblock/validatorupdates/kicked")
			delete(current, string(update.PubKey.Data))
		default:
			// Does validator already exist?
			if mVal, ok := current[string(update.PubKey.Data)]; ok {
				mVal.val = update
				event("endblock/validatorupdates/updated")
			} else {
				// Set this new validator
				current[string(update.PubKey.Data)] = mockValidator{update, GetMemberOfInitialState(r, initialLivenessWeightings)}
				event("endblock/validatorupdates/added")
			}
		}
	}
	return current
}
