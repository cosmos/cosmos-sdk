package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RandSetup performs the random setup the mock module needs.
type RandSetup func(r *rand.Rand, accounts []Account)

// Simulate tests application by sending random messages.
func Simulate(t *testing.T, app *baseapp.BaseApp,
	appStateFn func(r *rand.Rand, accs []Account) json.RawMessage,
	ops []WeightedOperation, setups []RandSetup,
	invariants Invariants, numBlocks int, blockSize int, commit bool) error {

	time := time.Now().UnixNano()
	return SimulateFromSeed(t, app, appStateFn, time, ops,
		setups, invariants, numBlocks, blockSize, commit)
}

// initialize the chain for the simulation
func initChain(r *rand.Rand, params Params,
	accounts []Account, setups []RandSetup, app *baseapp.BaseApp,
	appStateFn func(r *rand.Rand, accounts []Account) json.RawMessage) mockValidators {

	req := abci.RequestInitChain{
		AppStateBytes: appStateFn(r, accounts),
	}
	res := app.InitChain(req)
	validators = newMockValidators(res.Validators)

	for i := 0; i < len(setups); i++ {
		setups[i](r, accounts)
	}
	return validators
}

// SimulateFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided seed.
func SimulateFromSeed(tb testing.TB, app *baseapp.BaseApp,
	appStateFn func(r *rand.Rand, accs []Account) json.RawMessage,
	seed int64, ops WeightedOperations, setups []RandSetup, invariants Invariants,
	numBlocks int, blockSize int, commit bool) (simError error) {

	// in case we have to end early, don't os.Exit so that we can run cleanup code.
	stopEarly := false
	testingMode, t, b := getTestingMode(tb)
	fmt.Printf("Starting SimulateFromSeed with randomness created with seed %d\n", int(seed))
	r := rand.New(rand.NewSource(seed))
	params := RandomParams(r) // := DefaultParams()
	fmt.Printf("Randomized simulation params: %+v\n", params)
	timestamp := RandTimestamp(r)
	fmt.Printf("Starting the simulation from time %v, unixtime %v\n",
		timestamp.UTC().Format(time.UnixDate), timestamp.Unix())
	timeDiff := maxTimePerBlock - minTimePerBlock

	accs := RandomAccounts(r, params.NumKeys)

	// Setup event stats
	events := make(map[string]uint)
	event := func(what string) {
		events[what]++
	}

	validators := initChain(r, params, accs, setups, app, appStateFn)

	// Second variable to keep pending validator set (delayed one block since TM 0.24)
	// Initially this is the same as the initial validator set
	nextValidators := validators

	header := abci.Header{
		Height:          1,
		Time:            timestamp,
		ProposerAddress: randomProposer(r, validators),
	}
	opCount := 0

	// Setup code to catch SIGTERM's
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		receivedSignal := <-c
		fmt.Printf("\nExiting early due to %s, on block %d, operation %d\n",
			receivedSignal, header.Height, opCount)
		simError = fmt.Errorf("Exited due to %s", receivedSignal)
		stopEarly = true
	}()

	var pastTimes []time.Time
	var pastVoteInfos [][]abci.VoteInfo

	request := RandomRequestBeginBlock(r, params,
		validators, pastTimes, pastVoteInfos, event, header)

	// These are operations which have been queued by previous operations
	operationQueue := make(map[int][]Operation)
	timeOperationQueue := []FutureOperation{}
	var blockLogBuilders []*strings.Builder

	if testingMode {
		blockLogBuilders = make([]*strings.Builder, numBlocks)
	}
	displayLogs := logPrinter(testingMode, blockLogBuilders)
	blockSimulator := createBlockSimulator(
		testingMode, tb, t, params, event, invariants,
		ops, operationQueue, timeOperationQueue,
		numBlocks, blockSize, displayLogs)

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
				simError = fmt.Errorf(
					"Simulation halted due to panic on block %d",
					header.Height)
			}
		}()
	}

	// TODO split up the contents of this for loop into new functions
	for i := 0; i < numBlocks && !stopEarly; i++ {
		// Log the header time for future lookup
		pastTimes = append(pastTimes, header.Time)
		pastVoteInfos = append(pastVoteInfos, request.LastCommitInfo.Votes)

		// Construct log writer
		logWriter := addLogMessage(testingMode, blockLogBuilders, i)

		// Run the BeginBlock handler
		logWriter("BeginBlock")
		app.BeginBlock(request)

		if testingMode {
			// Make sure invariants hold at beginning of block
			invariants.assertAll(t, app, "BeginBlock", displayLogs)
		}

		ctx := app.NewContext(false, header)

		// Run queued operations. Ignores blocksize if blocksize is too small
		logWriter("Queued operations")
		numQueuedOpsRan := runQueuedOperations(
			operationQueue, int(header.Height),
			tb, r, app, ctx, accs, logWriter,
			displayLogs, event)
		numQueuedTimeOpsRan := runQueuedTimeOperations(
			timeOperationQueue, header.Time,
			tb, r, app, ctx, accs,
			logWriter, displayLogs, event)
		if testingMode && onOperation {
			// Make sure invariants hold at end of queued operations
			invariants.assertAll(t, app, "QueuedOperations", displayLogs)
		}

		logWriter("Standard operations")
		operations := blockSimulator(r, app, ctx, accs, header, logWriter)
		opCount += operations + numQueuedOpsRan + numQueuedTimeOpsRan
		if testingMode {
			// Make sure invariants hold at the operation
			invariants.assertAll(t, app, "StandardOperations", displayLogs)
		}

		res := app.EndBlock(abci.RequestEndBlock{})
		header.Height++
		header.Time = header.Time.Add(
			time.Duration(minTimePerBlock) * time.Second)
		header.Time = header.Time.Add(
			time.Duration(int64(r.Intn(int(timeDiff)))) * time.Second)
		header.ProposerAddress = randomProposer(r, validators)
		logWriter("EndBlock")

		if testingMode {
			// Make sure invariants hold at end of block
			invariants.assertAll(t, app, "EndBlock", displayLogs)
		}
		if commit {
			app.Commit()
		}

		if header.ProposerAddress == nil {
			fmt.Printf("\nSimulation stopped early as all validators " +
				"have been unbonded, there is nobody left propose a block!\n")
			stopEarly = true
			break
		}

		// Generate a random RequestBeginBlock with the current validator set for the next block
		request = RandomRequestBeginBlock(r, params, validators,
			pastTimes, pastVoteInfos, event, header)

		// Update the validator set, which will be reflected in the application on the next block
		validators = nextValidators
		nextValidators = updateValidators(tb, r, params,
			validators, res.ValidatorUpdates, event)
	}
	if stopEarly {
		DisplayEvents(events)
		return
	}
	fmt.Printf("\nSimulation complete. Final height (blocks): %d, "+
		"final time (seconds), : %v, operations ran %d\n",
		header.Height, header.Time, opCount)

	DisplayEvents(events)
	return nil
}

//______________________________________________________________________________

type blockSimFn func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
	accounts []Account, header abci.Header, logWriter func(string)) (opCount int)

// Returns a function to simulate blocks. Written like this to avoid constant
// parameters being passed everytime, to minimize memory overhead
func createBlockSimulator(testingMode bool, tb testing.TB, t *testing.T, params Params,
	event func(string), invariants Invariants, ops WeightedOperations,
	operationQueue map[int][]Operation, timeOperationQueue []FutureOperation,
	totalNumBlocks int, avgBlockSize int, displayLogs func()) blockSimFn {

	var lastBlocksizeState = 0 // state for [4 * uniform distribution]
	var blocksize int
	selectOp := ops.getSelectOpFn()

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accounts []Account, header abci.Header, logWriter func(string)) (opCount int) {

		fmt.Printf("\rSimulating... block %d/%d, operation %d/%d. ",
			header.Height, totalNumBlocks, opCount, blocksize)
		lastBlocksizeState, blocksize = getBlockSize(r, params, lastBlocksizeState, avgBlockSize)

		for j := 0; j < blocksize; j++ {
			logUpdate, futureOps, err := selectOp(r)(r, app, ctx, accounts, event)
			logWriter(logUpdate)
			if err != nil {
				displayLogs()
				tb.Fatalf("error on operation %d within block %d, %v",
					header.Height, opCount, err)
			}

			queueOperations(operationQueue, timeOperationQueue, futureOps)
			if testingMode {
				if onOperation {
					eventStr := fmt.Sprintf("operation: %v", logUpdate)
					invariants.assertAll(t, app, eventStr, displayLogs)
				}
				if opCount%50 == 0 {
					fmt.Printf("\rSimulating... block %d/%d, operation %d/%d. ",
						header.Height, totalNumBlocks, opCount, blocksize)
				}
			}
			opCount++
		}
		return opCount
	}
}

// nolint: errcheck
func runQueuedOperations(queueOps map[int][]Operation,
	height int, tb testing.TB, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
	accounts []Account, logWriter func(string),
	displayLogs func(), event func(string)) (numOpsRan int) {

	queuedOp, ok := queueOps[height]
	if !ok {
		return 0
	}

	numOpsRan = len(queuedOp)
	for i := 0; i < numOpsRan; i++ {
		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		logUpdate, _, err := queuedOp[i](r, app, ctx, accounts, event)
		logWriter(logUpdate)
		if err != nil {
			displayLogs()
			tb.FailNow()
		}
	}
	delete(queueOps, height)
	return numOpsRan
}

func runQueuedTimeOperations(queueOps []FutureOperation,
	currentTime time.Time, tb testing.TB, r *rand.Rand,
	app *baseapp.BaseApp, ctx sdk.Context, accounts []Account,
	logWriter func(string), displayLogs func(), event func(string)) (numOpsRan int) {

	numOpsRan = 0
	for len(queueOps) > 0 && currentTime.After(queueOps[0].BlockTime) {
		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		logUpdate, _, err := queueOps[0].Op(r, app, ctx, accounts, event)
		logWriter(logUpdate)
		if err != nil {
			displayLogs()
			tb.FailNow()
		}
		queueOps = queueOps[1:]
		numOpsRan++
	}
	return numOpsRan
}

// RandomRequestBeginBlock generates a list of signing validators according to
// the provided list of validators, signing fraction, and evidence fraction
func RandomRequestBeginBlock(r *rand.Rand, params Params,
	validators map[string]mockValidator, pastTimes []time.Time,
	pastVoteInfos [][]abci.VoteInfo,
	event func(string), header abci.Header) abci.RequestBeginBlock {

	if len(validators) == 0 {
		return abci.RequestBeginBlock{
			Header: header,
		}
	}
	voteInfos := make([]abci.VoteInfo, len(validators))

	for i, key := range getKeys(validators) {
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
		if r.Float64() < params.PastEvidenceFraction {
			height = int64(r.Intn(int(header.Height) - 1))
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

	return abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Votes: voteInfos,
		},
		ByzantineValidators: evidence,
	}
}
