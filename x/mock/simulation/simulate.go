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

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AppStateFn returns the app state json bytes
type AppStateFn func(r *rand.Rand, accs []Account) json.RawMessage

// Simulate tests application by sending random messages.
func Simulate(t *testing.T, app *baseapp.BaseApp,
	appStateFn AppStateFn, ops WeightedOperations,
	invariants Invariants, numBlocks int, blockSize int, commit bool) (bool, error) {

	time := time.Now().UnixNano()
	return SimulateFromSeed(t, app, appStateFn, time, ops,
		invariants, numBlocks, blockSize, commit)
}

// initialize the chain for the simulation
func initChain(r *rand.Rand, params Params, accounts []Account,
	app *baseapp.BaseApp,
	appStateFn AppStateFn) mockValidators {

	req := abci.RequestInitChain{
		AppStateBytes: appStateFn(r, accounts),
	}
	res := app.InitChain(req)
	validators := newMockValidators(r, res.Validators, params)

	return validators
}

// SimulateFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided seed.
// TODO split this monster function up
func SimulateFromSeed(tb testing.TB, app *baseapp.BaseApp,
	appStateFn AppStateFn, seed int64, ops WeightedOperations,
	invariants Invariants,
	numBlocks int, blockSize int, commit bool) (stopEarly bool, simError error) {

	// in case we have to end early, don't os.Exit so that we can run cleanup code.
	testingMode, t, b := getTestingMode(tb)
	fmt.Printf("Starting SimulateFromSeed with randomness "+
		"created with seed %d\n", int(seed))

	r := rand.New(rand.NewSource(seed))
	params := RandomParams(r) // := DefaultParams()
	fmt.Printf("Randomized simulation params: %+v\n", params)

	timestamp := RandTimestamp(r)
	fmt.Printf("Starting the simulation from time %v, unixtime %v\n",
		timestamp.UTC().Format(time.UnixDate), timestamp.Unix())

	timeDiff := maxTimePerBlock - minTimePerBlock
	accs := RandomAccounts(r, params.NumKeys)
	eventStats := newEventStats()

	// Second variable to keep pending validator set (delayed one block since
	// TM 0.24) Initially this is the same as the initial validator set
	validators := initChain(r, params, accs, app, appStateFn)
	nextValidators := validators

	header := abci.Header{
		Height:          1,
		Time:            timestamp,
		ProposerAddress: validators.randomProposer(r),
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
		validators, pastTimes, pastVoteInfos, eventStats.tally, header)

	// These are operations which have been queued by previous operations
	operationQueue := newOperationQueue()
	timeOperationQueue := []FutureOperation{}
	var blockLogBuilders []*strings.Builder

	if testingMode {
		blockLogBuilders = make([]*strings.Builder, numBlocks)
	}
	displayLogs := logPrinter(testingMode, blockLogBuilders)
	blockSimulator := createBlockSimulator(
		testingMode, tb, t, params, eventStats.tally, invariants,
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
			invariants.assertAll(t, app, "BeginBlock", displayLogs)
		}

		ctx := app.NewContext(false, header)

		// Run queued operations. Ignores blocksize if blocksize is too small
		logWriter("Queued operations")
		numQueuedOpsRan := runQueuedOperations(
			operationQueue, int(header.Height),
			tb, r, app, ctx, accs, logWriter,
			displayLogs, eventStats.tally)

		numQueuedTimeOpsRan := runQueuedTimeOperations(
			timeOperationQueue, header.Time,
			tb, r, app, ctx, accs,
			logWriter, displayLogs, eventStats.tally)

		if testingMode && onOperation {
			invariants.assertAll(t, app, "QueuedOperations", displayLogs)
		}

		logWriter("Standard operations")
		operations := blockSimulator(r, app, ctx, accs, header, logWriter)
		opCount += operations + numQueuedOpsRan + numQueuedTimeOpsRan
		if testingMode {
			invariants.assertAll(t, app, "StandardOperations", displayLogs)
		}

		res := app.EndBlock(abci.RequestEndBlock{})
		header.Height++
		header.Time = header.Time.Add(
			time.Duration(minTimePerBlock) * time.Second)
		header.Time = header.Time.Add(
			time.Duration(int64(r.Intn(int(timeDiff)))) * time.Second)
		header.ProposerAddress = validators.randomProposer(r)
		logWriter("EndBlock")

		if testingMode {
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

		// Generate a random RequestBeginBlock with the current validator set
		// for the next block
		request = RandomRequestBeginBlock(r, params, validators,
			pastTimes, pastVoteInfos, eventStats.tally, header)

		// Update the validator set, which will be reflected in the application
		// on the next block
		validators = nextValidators
		nextValidators = updateValidators(tb, r, params,
			validators, res.ValidatorUpdates, eventStats.tally)
	}

	if stopEarly {
		eventStats.Print()
		return true, simError
	}
	fmt.Printf("\nSimulation complete. Final height (blocks): %d, "+
		"final time (seconds), : %v, operations ran %d\n",
		header.Height, header.Time, opCount)

	eventStats.Print()
	return false, nil
}

//______________________________________________________________________________

type blockSimFn func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
	accounts []Account, header abci.Header, logWriter func(string)) (opCount int)

// Returns a function to simulate blocks. Written like this to avoid constant
// parameters being passed everytime, to minimize memory overhead.
func createBlockSimulator(testingMode bool, tb testing.TB, t *testing.T, params Params,
	event func(string), invariants Invariants, ops WeightedOperations,
	operationQueue OperationQueue, timeOperationQueue []FutureOperation,
	totalNumBlocks int, avgBlockSize int, displayLogs func()) blockSimFn {

	var lastBlocksizeState = 0 // state for [4 * uniform distribution]
	var blocksize int
	selectOp := ops.getSelectOpFn()

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accounts []Account, header abci.Header, logWriter func(string)) (opCount int) {

		fmt.Printf("\rSimulating... block %d/%d, operation %d/%d. ",
			header.Height, totalNumBlocks, opCount, blocksize)
		lastBlocksizeState, blocksize = getBlockSize(r, params, lastBlocksizeState, avgBlockSize)

		type opAndR struct {
			op   Operation
			rand *rand.Rand
		}
		opAndRz := make([]opAndR, 0, blocksize)
		// Predetermine the blocksize slice so that we can do things like block
		// out certain operations without changing the ops that follow.
		for i := 0; i < blocksize; i++ {
			opAndRz = append(opAndRz, opAndR{
				op:   selectOp(r),
				rand: DeriveRand(r),
			})
		}

		for i := 0; i < blocksize; i++ {
			// NOTE: the Rand 'r' should not be used here.
			opAndR := opAndRz[i]
			op, r2 := opAndR.op, opAndR.rand
			logUpdate, futureOps, err := op(r2, app, ctx, accounts, event)
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
	height int, tb testing.TB, r *rand.Rand, app *baseapp.BaseApp,
	ctx sdk.Context, accounts []Account, logWriter func(string),
	displayLogs func(), tallyEvent func(string)) (numOpsRan int) {

	queuedOp, ok := queueOps[height]
	if !ok {
		return 0
	}

	numOpsRan = len(queuedOp)
	for i := 0; i < numOpsRan; i++ {

		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		logUpdate, _, err := queuedOp[i](r, app, ctx, accounts, tallyEvent)
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
	logWriter func(string), displayLogs func(), tallyEvent func(string)) (numOpsRan int) {

	numOpsRan = 0
	for len(queueOps) > 0 && currentTime.After(queueOps[0].BlockTime) {

		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		logUpdate, _, err := queueOps[0].Op(r, app, ctx, accounts, tallyEvent)
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
