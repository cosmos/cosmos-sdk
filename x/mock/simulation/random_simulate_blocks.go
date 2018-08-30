package simulation

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
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
	t *testing.T, app *baseapp.BaseApp, appStateFn func(r *rand.Rand, keys []crypto.PrivKey, accs []sdk.AccAddress) json.RawMessage, ops []Operation, setups []RandSetup,
	invariants []Invariant, numBlocks int, blockSize int, commit bool,
) {
	time := time.Now().UnixNano()
	SimulateFromSeed(t, app, appStateFn, time, ops, setups, invariants, numBlocks, blockSize, commit)
}

func initChain(r *rand.Rand, keys []crypto.PrivKey, accs []sdk.AccAddress, setups []RandSetup, app *baseapp.BaseApp,
	appStateFn func(r *rand.Rand, keys []crypto.PrivKey, accs []sdk.AccAddress) json.RawMessage) (validators map[string]mockValidator) {
	res := app.InitChain(abci.RequestInitChain{AppStateBytes: appStateFn(r, keys, accs)})
	validators = make(map[string]mockValidator)
	for _, validator := range res.Validators {
		validators[string(validator.Address)] = mockValidator{validator, GetMemberOfInitialState(r, initialLivenessWeightings)}
	}

	for i := 0; i < len(setups); i++ {
		setups[i](r, keys)
	}

	return
}

func randTimestamp(r *rand.Rand) time.Time {
	unixTime := r.Int63n(int64(math.Pow(2, 40)))
	return time.Unix(unixTime, 0)
}

// SimulateFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided seed.
func SimulateFromSeed(
	t *testing.T, app *baseapp.BaseApp, appStateFn func(r *rand.Rand, keys []crypto.PrivKey, accs []sdk.AccAddress) json.RawMessage, seed int64, ops []Operation, setups []RandSetup,
	invariants []Invariant, numBlocks int, blockSize int, commit bool,
) {
	log := fmt.Sprintf("Starting SimulateFromSeed with randomness created with seed %d", int(seed))
	r := rand.New(rand.NewSource(seed))
	timestamp := randTimestamp(r)
	log = fmt.Sprintf("%s\nStarting the simulation from time %v, unixtime %v", log, timestamp.UTC().Format(time.UnixDate), timestamp.Unix())
	fmt.Printf("%s\n", log)
	timeDiff := maxTimePerBlock - minTimePerBlock

	keys, accs := mock.GeneratePrivKeyAddressPairsFromRand(r, numKeys)

	// Setup event stats
	events := make(map[string]uint)
	event := func(what string) {
		log += "\nevent - " + what
		events[what]++
	}

	validators := initChain(r, keys, accs, setups, app, appStateFn)

	header := abci.Header{Height: 0, Time: timestamp}
	opCount := 0

	request := abci.RequestBeginBlock{Header: header}

	var lastHeaderTime time.Time
	// These are operations which have been queued by previous operations
	operationQueue := make(map[int][]Operation)

	for i := 0; i < numBlocks; i++ {
		// Run the BeginBlock handler
		app.BeginBlock(request)

		log += "\nBeginBlock"

		// Make sure invariants hold at beginning of block
		AssertAllInvariants(t, app, invariants, log)

		ctx := app.NewContext(false, header)

		thisBlockSize := getBlockSize(r, blockSize)

		// Run queued operations. Ignores blocksize if blocksize is too small
		log, numQueuedOpsRan := runQueuedOperations(operationQueue, int(header.Height), t, r, app, ctx, keys, log, event)
		opCount += numQueuedOpsRan
		thisBlockSize -= numQueuedOpsRan
		for j := 0; j < thisBlockSize; j++ {
			logUpdate, futureOps, err := ops[r.Intn(len(ops))](t, r, app, ctx, keys, log, event)
			log += "\n" + logUpdate
			queueOperations(operationQueue, futureOps)

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
		lastHeaderTime = header.Time
		header.Time = header.Time.Add(time.Duration(minTimePerBlock) * time.Second).Add(time.Duration(int64(r.Intn(int(timeDiff)))) * time.Second)

		log += "\nEndBlock"

		// Make sure invariants hold at end of block
		AssertAllInvariants(t, app, invariants, log)

		if commit {
			app.Commit()
		}

		// Generate a random RequestBeginBlock with the current validator set for the next block
		request = RandomRequestBeginBlock(r, validators, livenessTransitionMatrix, evidenceFraction, lastHeaderTime, event, header, log)

		// Update the validator set
		validators = updateValidators(t, r, validators, res.ValidatorUpdates, event)
	}

	fmt.Printf("\nSimulation complete. Final height (blocks): %d, final time (seconds): %v\n", header.Height, header.Time)
	DisplayEvents(events)
}

func getBlockSize(r *rand.Rand, blockSize int) int {
	load := r.Float64()
	switch {
	case load < 0.33:
		return 0
	case load < 0.66:
		return r.Intn(blockSize * 2)
	default:
		return r.Intn(blockSize * 4)
	}
}

// Simulate from seed, benchmarks
func BenchmarkSimulationFromSeed(b *testing.B, app *baseapp.BaseApp, appStateFn func(r *rand.Rand, keys []crypto.PrivKey, accs []sdk.AccAddress) json.RawMessage,
	seed int64, ops []Operation, setups []RandSetup, numBlocks int, blockSize int, commit bool) {
	r := rand.New(rand.NewSource(seed))
	timestamp := randTimestamp(r)
	timeDiff := maxTimePerBlock - minTimePerBlock
	keys, accs := mock.GeneratePrivKeyAddressPairsFromRand(r, numKeys)

	// Setup event stats
	events := make(map[string]uint)
	event := func(what string) {
		events[what]++
	}

	validators := initChain(r, keys, accs, setups, app, appStateFn)

	header := abci.Header{Height: 0, Time: timestamp}
	opCount := 0

	request := abci.RequestBeginBlock{Header: header}

	var lastHeaderTime time.Time
	// These are operations which have been queued by previous operations
	operationQueue := make(map[int][]Operation)
	b.ResetTimer()

	for i := 0; i < numBlocks; i++ {

		// Run the BeginBlock handler
		app.BeginBlock(request)

		ctx := app.NewContext(false, header)

		thisBlockSize := getBlockSize(r, blockSize)

		// Run queued operations. Ignores blocksize if blocksize is too small
		log, numQueuedOpsRan := runQueuedOperations(operationQueue, int(header.Height), b, r, app, ctx, keys, "", event)
		opCount += numQueuedOpsRan
		thisBlockSize -= numQueuedOpsRan
		for j := 0; j < thisBlockSize; j++ {
			_, futureOps, err := ops[r.Intn(len(ops))](b, r, app, ctx, keys, "", event)
			queueOperations(operationQueue, futureOps)
			if err != nil {
				b.Fatalf("error on operation %d, %v", opCount, err)
			}
			opCount++
		}

		res := app.EndBlock(abci.RequestEndBlock{})
		header.Height++
		lastHeaderTime = header.Time
		header.Time = header.Time.Add(time.Duration(minTimePerBlock) * time.Second).Add(time.Duration(int64(r.Intn(int(timeDiff)))) * time.Second)
		if commit {
			app.Commit()
		}

		// Generate a random RequestBeginBlock with the current validator set for the next block
		request = RandomRequestBeginBlock(r, validators, livenessTransitionMatrix, evidenceFraction, lastHeaderTime, event, header, log)

		// Update the validator set
		validators = updateValidators(b, r, validators, res.ValidatorUpdates, event)
	}
	DisplayEvents(events)
	fmt.Printf("Benchmark simulation ran %d operations\n", opCount)
}

// adds all future operations into the operation queue.
func queueOperations(queuedOperations map[int][]Operation, futureOperations []FutureOperation) {
	if futureOperations == nil {
		return
	}
	for _, futureOp := range futureOperations {
		if val, ok := queuedOperations[futureOp.BlockHeight]; ok {
			queuedOperations[futureOp.BlockHeight] = append(val, futureOp.Op)
		} else {
			queuedOperations[futureOp.BlockHeight] = []Operation{futureOp.Op}
		}
	}
}

func runQueuedOperations(queueOperations map[int][]Operation, height int, tb testing.TB, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
	privKeys []crypto.PrivKey, log string, event func(string)) (updatedLog string, numOpsRan int) {
	updatedLog = log
	if queuedOps, ok := queueOperations[height]; ok {
		numOps := len(queuedOps)
		for i := 0; i < numOps; i++ {
			// For now, queued operations cannot queue more operations.
			// If a need arises for us to support queued messages to queue more messages, this can
			// be changed.
			logUpdate, _, err := queuedOps[i](tb, r, app, ctx, privKeys, updatedLog, event)
			updatedLog = fmt.Sprintf("%s\n%s", updatedLog, logUpdate)
			if err != nil {
				fmt.Fprint(os.Stderr, updatedLog)
				tb.FailNow()
			}
		}
		delete(queueOperations, height)
		return updatedLog, numOps
	}
	return log, 0
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
func RandomRequestBeginBlock(r *rand.Rand, validators map[string]mockValidator, livenessTransitions TransitionMatrix, evidenceFraction float64,
	lastHeaderTime time.Time, event func(string), header abci.Header, log string) abci.RequestBeginBlock {
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
	// TODO: Determine capacity before allocation
	evidence := make([]abci.Evidence, 0)
	for r.Float64() < evidenceFraction {
		height := header.Height
		time := header.Time
		if r.Float64() < pastEvidenceFraction {
			height = int64(r.Intn(int(header.Height)))
			time = lastHeaderTime
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
func updateValidators(tb testing.TB, r *rand.Rand, current map[string]mockValidator, updates []abci.Validator, event func(string)) map[string]mockValidator {
	for _, update := range updates {
		switch {
		case update.Power == 0:
			// // TEMPORARY DEBUG CODE TO PROVE THAT THE OLD METHOD WAS BROKEN
			// // (i.e. didn't catch in the event of problem)
			// if val, ok := tb.(*testing.T); ok {
			// 	require.NotNil(val, current[string(update.PubKey.Data)])
			// }
			// // CORRECT CHECK
			// if _, ok := current[string(update.PubKey.Data)]; !ok {
			// 	tb.Fatalf("tried to delete a nonexistent validator")
			// }
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
