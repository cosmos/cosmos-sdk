package simulation

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AppStateFn returns the app state json bytes, the genesis accounts, and the chain identifier
type AppStateFn func(r *rand.Rand, accs []Account, genesisTimestamp time.Time) (appState json.RawMessage, accounts []Account, chainId string)

// initialize the chain for the simulation
func initChain(
	r *rand.Rand, params Params, accounts []Account,
	app *baseapp.BaseApp, appStateFn AppStateFn, genesisTimestamp time.Time,
) (mockValidators, []Account) {

	appState, accounts, chainID := appStateFn(r, accounts, genesisTimestamp)

	req := abci.RequestInitChain{
		AppStateBytes: appState,
		ChainId:       chainID,
	}
	res := app.InitChain(req)
	validators := newMockValidators(r, res.Validators, params)

	return validators, accounts
}

// SimulateFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided seed.
// TODO split this monster function up
func SimulateFromSeed(
	tb testing.TB, w io.Writer, app *baseapp.BaseApp,
	appStateFn AppStateFn, seed int64, ops WeightedOperations,
	invariants sdk.Invariants,
	numBlocks, blockSize int, commit, lean, onOperation bool,
) (stopEarly bool, simError error) {

	// in case we have to end early, don't os.Exit so that we can run cleanup code.
	testingMode, t, b := getTestingMode(tb)
	fmt.Fprintf(w, "Starting SimulateFromSeed with randomness created with seed %d\n", int(seed))

	r := rand.New(rand.NewSource(seed))
	params := RandomParams(r)
	fmt.Fprintf(w, "Randomized simulation params: \n%s\n", mustMarshalJSONIndent(params))

	genesisTimestamp := RandTimestamp(r)
	fmt.Printf(
		"Starting the simulation from time %v, unixtime %v\n",
		genesisTimestamp.UTC().Format(time.UnixDate), genesisTimestamp.Unix(),
	)

	timeDiff := maxTimePerBlock - minTimePerBlock
	accs := RandomAccounts(r, params.NumKeys)
	eventStats := newEventStats()

	// Second variable to keep pending validator set (delayed one block since
	// TM 0.24) Initially this is the same as the initial validator set
	validators, accs := initChain(r, params, accs, app, appStateFn, genesisTimestamp)
	if len(accs) == 0 {
		return true, fmt.Errorf("must have greater than zero genesis accounts")
	}

	nextValidators := validators

	header := abci.Header{
		Height:          1,
		Time:            genesisTimestamp,
		ProposerAddress: validators.randomProposer(r),
	}
	opCount := 0

	// Setup code to catch SIGTERM's
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		receivedSignal := <-c
		fmt.Fprintf(w, "\nExiting early due to %s, on block %d, operation %d\n", receivedSignal, header.Height, opCount)
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

	logWriter := NewLogWriter(testingMode)

	blockSimulator := createBlockSimulator(
		testingMode, tb, t, w, params, eventStats.tally, invariants,
		ops, operationQueue, timeOperationQueue,
		numBlocks, blockSize, logWriter, lean, onOperation)

	if !testingMode {
		b.ResetTimer()
	} else {
		// Recover logs in case of panic
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(w, "panic with err: %v\n", r)
				stackTrace := string(debug.Stack())
				fmt.Println(stackTrace)
				logWriter.PrintLogs()
				simError = fmt.Errorf("Simulation halted due to panic on block %d", header.Height)
			}
		}()
	}

	// TODO split up the contents of this for loop into new functions
	for height := 1; height <= numBlocks && !stopEarly; height++ {

		// Log the header time for future lookup
		pastTimes = append(pastTimes, header.Time)
		pastVoteInfos = append(pastVoteInfos, request.LastCommitInfo.Votes)

		// Run the BeginBlock handler
		logWriter.AddEntry(BeginBlockEntry(int64(height)))
		app.BeginBlock(request)

		if testingMode {
			assertAllInvariants(t, app, invariants, "BeginBlock", logWriter)
		}

		ctx := app.NewContext(false, header)

		// Run queued operations. Ignores blocksize if blocksize is too small
		numQueuedOpsRan := runQueuedOperations(
			operationQueue, int(header.Height),
			tb, r, app, ctx, accs, logWriter, eventStats.tally, lean)

		numQueuedTimeOpsRan := runQueuedTimeOperations(
			timeOperationQueue, int(header.Height), header.Time,
			tb, r, app, ctx, accs, logWriter, eventStats.tally, lean)

		if testingMode && onOperation {
			assertAllInvariants(t, app, invariants, "QueuedOperations", logWriter)
		}

		// run standard operations
		operations := blockSimulator(r, app, ctx, accs, header)
		opCount += operations + numQueuedOpsRan + numQueuedTimeOpsRan
		if testingMode {
			assertAllInvariants(t, app, invariants, "StandardOperations", logWriter)
		}

		res := app.EndBlock(abci.RequestEndBlock{})
		header.Height++
		header.Time = header.Time.Add(
			time.Duration(minTimePerBlock) * time.Second)
		header.Time = header.Time.Add(
			time.Duration(int64(r.Intn(int(timeDiff)))) * time.Second)
		header.ProposerAddress = validators.randomProposer(r)
		logWriter.AddEntry(EndBlockEntry(int64(height)))

		if testingMode {
			assertAllInvariants(t, app, invariants, "EndBlock", logWriter)
		}
		if commit {
			app.Commit()
		}

		if header.ProposerAddress == nil {
			fmt.Fprintf(w, "\nSimulation stopped early as all validators have been unbonded; nobody left to propose a block!\n")
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
		eventStats.Print(w)
		return true, simError
	}

	fmt.Fprintf(
		w,
		"\nSimulation complete; Final height (blocks): %d, final time (seconds): %v, operations ran: %d\n",
		header.Height, header.Time, opCount,
	)

	eventStats.Print(w)
	return false, nil
}

//______________________________________________________________________________

type blockSimFn func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
	accounts []Account, header abci.Header) (opCount int)

// Returns a function to simulate blocks. Written like this to avoid constant
// parameters being passed everytime, to minimize memory overhead.
func createBlockSimulator(testingMode bool, tb testing.TB, t *testing.T, w io.Writer, params Params,
	event func(string), invariants sdk.Invariants, ops WeightedOperations,
	operationQueue OperationQueue, timeOperationQueue []FutureOperation,
	totalNumBlocks, avgBlockSize int, logWriter LogWriter, lean, onOperation bool) blockSimFn {

	lastBlocksizeState := 0 // state for [4 * uniform distribution]
	blocksize := 0
	selectOp := ops.getSelectOpFn()

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accounts []Account, header abci.Header) (opCount int) {

		fmt.Fprintf(w, "\rSimulating... block %d/%d, operation %d/%d. ",
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
			opMsg, futureOps, err := op(r2, app, ctx, accounts)
			opMsg.LogEvent(event)
			if !lean || opMsg.OK {
				logWriter.AddEntry(MsgEntry(header.Height, opMsg, int64(i)))
			}
			if err != nil {
				logWriter.PrintLogs()
				tb.Fatalf("error on operation %d within block %d, %v",
					header.Height, opCount, err)
			}

			queueOperations(operationQueue, timeOperationQueue, futureOps)
			if testingMode {
				if onOperation {
					fmt.Fprintf(w, "\rSimulating... block %d/%d, operation %d/%d. ",
						header.Height, totalNumBlocks, opCount, blocksize)
					eventStr := fmt.Sprintf("operation: %v", opMsg.String())
					assertAllInvariants(t, app, invariants, eventStr, logWriter)
				} else if opCount%50 == 0 {
					fmt.Fprintf(w, "\rSimulating... block %d/%d, operation %d/%d. ",
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
	ctx sdk.Context, accounts []Account, logWriter LogWriter, tallyEvent func(string), lean bool) (numOpsRan int) {

	queuedOp, ok := queueOps[height]
	if !ok {
		return 0
	}

	numOpsRan = len(queuedOp)
	for i := 0; i < numOpsRan; i++ {

		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		opMsg, _, err := queuedOp[i](r, app, ctx, accounts)
		opMsg.LogEvent(tallyEvent)
		if !lean || opMsg.OK {
			logWriter.AddEntry((QueuedMsgEntry(int64(height), opMsg)))
		}
		if err != nil {
			logWriter.PrintLogs()
			tb.FailNow()
		}
	}
	delete(queueOps, height)
	return numOpsRan
}

func runQueuedTimeOperations(queueOps []FutureOperation,
	height int, currentTime time.Time, tb testing.TB, r *rand.Rand,
	app *baseapp.BaseApp, ctx sdk.Context, accounts []Account,
	logWriter LogWriter, tallyEvent func(string), lean bool) (numOpsRan int) {

	numOpsRan = 0
	for len(queueOps) > 0 && currentTime.After(queueOps[0].BlockTime) {

		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		opMsg, _, err := queueOps[0].Op(r, app, ctx, accounts)
		opMsg.LogEvent(tallyEvent)
		if !lean || opMsg.OK {
			logWriter.AddEntry(QueuedMsgEntry(int64(height), opMsg))
		}
		if err != nil {
			logWriter.PrintLogs()
			tb.FailNow()
		}

		queueOps = queueOps[1:]
		numOpsRan++
	}
	return numOpsRan
}
