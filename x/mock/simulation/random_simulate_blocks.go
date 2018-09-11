package simulation

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"runtime/debug"
	"sort"
	"strings"
	"syscall"
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
) error {
	time := time.Now().UnixNano()
	return SimulateFromSeed(t, app, appStateFn, time, ops, setups, invariants, numBlocks, blockSize, commit)
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
) (simError error) {
	// in case we have to end early, don't os.Exit so that we can run cleanup code.
	stopEarly := false
	testingMode, t, b := getTestingMode(tb)
	fmt.Printf("Starting SimulateFromSeed with randomness created with seed %d\n", int(seed))
	r := rand.New(rand.NewSource(seed))
	timestamp := randTimestamp(r)
	fmt.Printf("Starting the simulation from time %v, unixtime %v\n", timestamp.UTC().Format(time.UnixDate), timestamp.Unix())
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

	// Setup code to catch SIGTERM's
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		receivedSignal := <-c
		fmt.Printf("Exiting early due to %s, on block %d, operation %d\n", receivedSignal, header.Height, opCount)
		simError = fmt.Errorf("Exited due to %s", receivedSignal)
		stopEarly = true
	}()

	var pastTimes []time.Time
	var pastSigningValidators [][]abci.SigningValidator

	request := RandomRequestBeginBlock(r, validators, livenessTransitionMatrix, evidenceFraction, pastTimes, pastSigningValidators, event, header)
	// These are operations which have been queued by previous operations
	operationQueue := make(map[int][]Operation)
	var blockLogBuilders []*strings.Builder

	if testingMode {
		blockLogBuilders = make([]*strings.Builder, numBlocks)
	}
	displayLogs := logPrinter(testingMode, blockLogBuilders)
	blockSimulator := createBlockSimulator(testingMode, tb, t, event, invariants, ops, operationQueue, numBlocks, displayLogs)
	if !testingMode {
		b.ResetTimer()
	} else {
		// Recover logs in case of panic
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Panic with err\n", r)
				stackTrace := string(debug.Stack())
				fmt.Println(stackTrace)
				displayLogs()
				simError = fmt.Errorf("Simulation halted due to panic on block %d", header.Height)
			}
		}()
	}

	for i := 0; i < numBlocks && !stopEarly; i++ {
		// Log the header time for future lookup
		pastTimes = append(pastTimes, header.Time)
		pastSigningValidators = append(pastSigningValidators, request.LastCommitInfo.Validators)

		// Run the BeginBlock handler
		app.BeginBlock(request)

		if testingMode {
			// Make sure invariants hold at beginning of block
			assertAllInvariants(t, app, invariants, displayLogs)
		}
		logWriter := addLogMessage(testingMode, blockLogBuilders, i)

		ctx := app.NewContext(false, header)
		thisBlockSize := getBlockSize(r, blockSize)

		// Run queued operations. Ignores blocksize if blocksize is too small
		numQueuedOpsRan := runQueuedOperations(operationQueue, int(header.Height), tb, r, app, ctx, keys, logWriter, displayLogs, event)
		thisBlockSize -= numQueuedOpsRan
		operations := blockSimulator(thisBlockSize, r, app, ctx, keys, header, logWriter)
		opCount += operations + numQueuedOpsRan

		res := app.EndBlock(abci.RequestEndBlock{})
		header.Height++
		header.Time = header.Time.Add(time.Duration(minTimePerBlock) * time.Second).Add(time.Duration(int64(r.Intn(int(timeDiff)))) * time.Second)
		logWriter("EndBlock")

		if testingMode {
			// Make sure invariants hold at end of block
			assertAllInvariants(t, app, invariants, displayLogs)
		}
		if commit {
			app.Commit()
		}

		// Generate a random RequestBeginBlock with the current validator set for the next block
		request = RandomRequestBeginBlock(r, validators, livenessTransitionMatrix, evidenceFraction, pastTimes, pastSigningValidators, event, header)

		// Update the validator set
		validators = updateValidators(tb, r, validators, res.ValidatorUpdates, event)
	}
	if stopEarly {
		DisplayEvents(events)
		return
	}
	fmt.Printf("\nSimulation complete. Final height (blocks): %d, final time (seconds), : %v, operations ran %d\n", header.Height, header.Time, opCount)
	DisplayEvents(events)
	return nil
}

// Returns a function to simulate blocks. Written like this to avoid constant parameters being passed everytime, to minimize
// memory overhead
func createBlockSimulator(testingMode bool, tb testing.TB, t *testing.T, event func(string), invariants []Invariant, ops []Operation, operationQueue map[int][]Operation, totalNumBlocks int, displayLogs func()) func(
	blocksize int, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, privKeys []crypto.PrivKey, header abci.Header, logWriter func(string)) (opCount int) {
	return func(blocksize int, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		keys []crypto.PrivKey, header abci.Header, logWriter func(string)) (opCount int) {
		for j := 0; j < blocksize; j++ {
			logUpdate, futureOps, err := ops[r.Intn(len(ops))](r, app, ctx, keys, event)
			if err != nil {
				displayLogs()
				tb.Fatalf("error on operation %d within block %d, %v", header.Height, opCount, err)
			}
			logWriter(logUpdate)

			queueOperations(operationQueue, futureOps)
			if testingMode {
				if onOperation {
					assertAllInvariants(t, app, invariants, displayLogs)
				}
				if opCount%50 == 0 {
					fmt.Printf("\rSimulating... block %d/%d, operation %d/%d.  ", header.Height, totalNumBlocks, opCount, blocksize)
				}
			}
			opCount++
		}
		return opCount
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
	privKeys []crypto.PrivKey, logWriter func(string), displayLogs func(), event func(string)) (numOpsRan int) {
	if queuedOps, ok := queueOperations[height]; ok {
		numOps := len(queuedOps)
		for i := 0; i < numOps; i++ {
			// For now, queued operations cannot queue more operations.
			// If a need arises for us to support queued messages to queue more messages, this can
			// be changed.
			logUpdate, _, err := queuedOps[i](r, app, ctx, privKeys, event)
			logWriter(logUpdate)
			if err != nil {
				displayLogs()
				tb.FailNow()
			}
		}
		delete(queueOperations, height)
		return numOps
	}
	return 0
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
	pastTimes []time.Time, pastSigningValidators [][]abci.SigningValidator, event func(string), header abci.Header) abci.RequestBeginBlock {
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
	// Anything but the first block
	if len(pastTimes) > 0 {
		for r.Float64() < evidenceFraction {
			height := header.Height
			time := header.Time
			vals := signingValidators
			if r.Float64() < pastEvidenceFraction {
				height = int64(r.Intn(int(header.Height)))
				time = pastTimes[height]
				vals = pastSigningValidators[height]
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
			Validators: signingValidators,
		},
		ByzantineValidators: evidence,
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
