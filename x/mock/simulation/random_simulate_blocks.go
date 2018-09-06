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
	tb testing.TB, app *baseapp.BaseApp, appStateFn func(r *rand.Rand, keys []crypto.PrivKey, accs []sdk.AccAddress) json.RawMessage, seed int64, ops []Operation, setups []RandSetup,
	invariants []Invariant, numBlocks int, blockSize int, commit bool,
) {
	testingMode, t, b := getTestingMode(tb)
	log := fmt.Sprintf("Starting SimulateFromSeed with randomness created with seed %d", int(seed))
	r := rand.New(rand.NewSource(seed))
	timestamp := randTimestamp(r)
	log = updateLog(testingMode, log, "Starting the simulation from time %v, unixtime %v", timestamp.UTC().Format(time.UnixDate), timestamp.Unix())
	fmt.Printf("%s\n", log)
	timeDiff := maxTimePerBlock - minTimePerBlock

	keys, accs := mock.GeneratePrivKeyAddressPairsFromRand(r, numKeys)

	// Setup event stats
	events := make(map[string]uint)
	event := func(what string) {
		log = updateLog(testingMode, log, "event - %s", what)
		events[what]++
	}

	validators := initChain(r, keys, accs, setups, app, appStateFn)

	header := abci.Header{Height: 0, Time: timestamp}
	opCount := 0

	var pastTimes []time.Time
	var pastVoteInfos [][]abci.VoteInfo

	request := RandomRequestBeginBlock(r, validators, livenessTransitionMatrix, evidenceFraction, pastTimes, pastVoteInfos, event, header, log)
	// These are operations which have been queued by previous operations
	operationQueue := make(map[int][]Operation)

	if !testingMode {
		b.ResetTimer()
	}
	blockSimulator := createBlockSimulator(testingMode, tb, t, event, invariants, ops, operationQueue, numBlocks)

	for i := 0; i < numBlocks; i++ {
		// Log the header time for future lookup
		pastTimes = append(pastTimes, header.Time)
		pastVoteInfos = append(pastVoteInfos, request.LastCommitInfo.Validators)

		// Run the BeginBlock handler
		app.BeginBlock(request)
		log = updateLog(testingMode, log, "BeginBlock")

		if testingMode {
			// Make sure invariants hold at beginning of block
			AssertAllInvariants(t, app, invariants, log)
		}

		ctx := app.NewContext(false, header)
		thisBlockSize := getBlockSize(r, blockSize)

		// Run queued operations. Ignores blocksize if blocksize is too small
		log, numQueuedOpsRan := runQueuedOperations(operationQueue, int(header.Height), tb, r, app, ctx, keys, log, event)
		opCount += numQueuedOpsRan
		thisBlockSize -= numQueuedOpsRan
		log, operations := blockSimulator(thisBlockSize, r, app, ctx, keys, log, header)
		opCount += operations

		res := app.EndBlock(abci.RequestEndBlock{})
		header.Height++
		header.Time = header.Time.Add(time.Duration(minTimePerBlock) * time.Second).Add(time.Duration(int64(r.Intn(int(timeDiff)))) * time.Second)
		log = updateLog(testingMode, log, "EndBlock")

		if testingMode {
			// Make sure invariants hold at end of block
			AssertAllInvariants(t, app, invariants, log)
		}
		if commit {
			app.Commit()
		}

		// Generate a random RequestBeginBlock with the current validator set for the next block
		request = RandomRequestBeginBlock(r, validators, livenessTransitionMatrix, evidenceFraction, pastTimes, pastVoteInfos, event, header, log)

		// Update the validator set
		validators = updateValidators(tb, r, validators, res.ValidatorUpdates, event)
	}

	fmt.Printf("\nSimulation complete. Final height (blocks): %d, final time (seconds), : %v, operations ran %d\n", header.Height, header.Time, opCount)
	DisplayEvents(events)
}

// Returns a function to simulate blocks. Written like this to avoid constant parameters being passed everytime, to minimize
// memory overhead
func createBlockSimulator(testingMode bool, tb testing.TB, t *testing.T, event func(string), invariants []Invariant, ops []Operation, operationQueue map[int][]Operation, totalNumBlocks int) func(
	blocksize int, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, privKeys []crypto.PrivKey, log string, header abci.Header) (updatedLog string, opCount int) {
	return func(blocksize int, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		keys []crypto.PrivKey, log string, header abci.Header) (updatedLog string, opCount int) {
		for j := 0; j < blocksize; j++ {
			logUpdate, futureOps, err := ops[r.Intn(len(ops))](tb, r, app, ctx, keys, log, event)
			log = updateLog(testingMode, log, logUpdate)
			if err != nil {
				tb.Fatalf("error on operation %d within block %d, %v, log %s", header.Height, opCount, err, log)
			}

			queueOperations(operationQueue, futureOps)
			if testingMode {
				if onOperation {
					AssertAllInvariants(t, app, invariants, log)
				}
				if opCount%50 == 0 {
					fmt.Printf("\rSimulating... block %d/%d, operation %d/%d.  ", header.Height, totalNumBlocks, opCount, blocksize)
				}
			}
			opCount++
		}
		return log, opCount
	}
}

func getTestingMode(tb testing.TB) (testingMode bool, t *testing.T, b *testing.B) {
	testingMode = false
	if _t, ok := tb.(*testing.T); ok {
		t = _t
		testingMode = true
	} else {
		b = tb.(*testing.B)
	}
	return
}

func updateLog(testingMode bool, log string, update string, args ...interface{}) (updatedLog string) {
	if testingMode {
		update = fmt.Sprintf(update, args...)
		return fmt.Sprintf("%s\n%s", log, update)
	}
	return ""
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

// nolint: errcheck
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
// nolint: unparam
func RandomRequestBeginBlock(r *rand.Rand, validators map[string]mockValidator, livenessTransitions TransitionMatrix, evidenceFraction float64,
	pastTimes []time.Time, pastVoteInfos [][]abci.VoteInfo, event func(string), header abci.Header, log string) abci.RequestBeginBlock {
	if len(validators) == 0 {
		return abci.RequestBeginBlock{Header: header}
	}
	voteInfos := make([]abci.VoteInfo, len(validators))
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
		voteInfos[i] = abci.VoteInfo{
			Validator:       mVal.val,
			SignedLastBlock: signed,
		}
		i++
	}
	// TODO: Determine capacity before allocation
	evidence := make([]abci.Evidence, 0)
	// Anything but the first block
	if len(pastTimes) > 0 {
		for r.Float64() < evidenceFraction {
			height := header.Height
			time := header.Time
			vals := voteInfos
			if r.Float64() < pastEvidenceFraction {
				height = int64(r.Intn(int(header.Height)))
				time = pastTimes[height]
				vals = pastVoteInfos[height]
			}
			validator := vals[r.Intn(len(vals))].Validator
			var totalVotingPower int64
			for _, val := range vals {
				totalVotingPower += val.Validator.Power
			}
			evidence = append(evidence, abci.Evidence{
				Type:             tmtypes.ABCIEvidenceTypeDuplicateVote,
				Validator:        validator,
				Height:           height,
				Time:             time,
				TotalVotingPower: totalVotingPower,
			})
			event("beginblock/evidence")
		}
	}
	return abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Validators: voteInfos,
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
// nolint: unparam
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
