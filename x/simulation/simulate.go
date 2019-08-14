package simulation

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AppStateFn returns the app state json bytes, the genesis accounts, and the chain identifier
type AppStateFn func(
	r *rand.Rand, accs []Account,
) (appState json.RawMessage, accounts []Account, chainId string, genesisTimestamp time.Time)

// initialize the chain for the simulation
func initChain(
	r *rand.Rand, params Params, accounts []Account, app *baseapp.BaseApp, appStateFn AppStateFn,
) (mockValidators, time.Time, []Account) {

	appState, accounts, chainID, genesisTimestamp := appStateFn(r, accounts)

	req := abci.RequestInitChain{
		AppStateBytes: appState,
		ChainId:       chainID,
	}
	res := app.InitChain(req)
	validators := newMockValidators(r, res.Validators, params)

	return validators, genesisTimestamp, accounts
}

// SimulateFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided seed.
// TODO: split this monster function up
func SimulateFromSeed(
	tb testing.TB, w io.Writer, app *baseapp.BaseApp,
	appStateFn AppStateFn, seed int64,
	ops WeightedOperations, invariants sdk.Invariants,
	initialHeight, numBlocks, exportParamsHeight, blockSize int,
	exportStatsPath string,
	exportParams, commit, lean, onOperation, allInvariants bool,
	blackListedAccs map[string]bool,
) (stopEarly bool, exportedParams Params, err error) {

	// in case we have to end early, don't os.Exit so that we can run cleanup code.
	testingMode, t, b := getTestingMode(tb)
	fmt.Fprintf(w, "Starting SimulateFromSeed with randomness created with seed %d\n", int(seed))

	r := rand.New(rand.NewSource(seed))
	params := RandomParams(r)
	fmt.Fprintf(w, "Randomized simulation params: \n%s\n", mustMarshalJSONIndent(params))

	timeDiff := maxTimePerBlock - minTimePerBlock
	accs := RandomAccounts(r, params.NumKeys)
	eventStats := NewEventStats()

	// Second variable to keep pending validator set (delayed one block since
	// TM 0.24) Initially this is the same as the initial validator set
	validators, genesisTimestamp, accs := initChain(r, params, accs, app, appStateFn)
	if len(accs) == 0 {
		return true, params, fmt.Errorf("must have greater than zero genesis accounts")
	}

	fmt.Printf(
		"Starting the simulation from time %v (unixtime %v)\n",
		genesisTimestamp.UTC().Format(time.UnixDate), genesisTimestamp.Unix(),
	)

	// remove module account address if they exist in accs
	var tmpAccs []Account
	for _, acc := range accs {
		if !blackListedAccs[acc.Address.String()] {
			tmpAccs = append(tmpAccs, acc)
		}
	}

	accs = tmpAccs

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
		err = fmt.Errorf("Exited due to %s", receivedSignal)
		stopEarly = true
	}()

	var pastTimes []time.Time
	var pastVoteInfos [][]abci.VoteInfo

	request := RandomRequestBeginBlock(r, params,
		validators, pastTimes, pastVoteInfos, eventStats.Tally, header)

	// These are operations which have been queued by previous operations
	operationQueue := NewOperationQueue()
	timeOperationQueue := []FutureOperation{}

	logWriter := NewLogWriter(testingMode)

	blockSimulator := createBlockSimulator(
		testingMode, tb, t, w, params, eventStats.Tally, invariants,
		ops, operationQueue, timeOperationQueue,
		numBlocks, blockSize, logWriter, lean, onOperation, allInvariants)

	if !testingMode {
		b.ResetTimer()
	} else {
		// recover logs in case of panic
		defer func() {
			if r := recover(); r != nil {
				_, _ = fmt.Fprintf(w, "simulation halted due to panic on block %d; %v\n", header.Height, r)
				logWriter.PrintLogs()
				panic(r)
			}
		}()
	}

	// set exported params to the initial state
	if exportParams && exportParamsHeight == 0 {
		exportedParams = params
	}

	// TODO: split up the contents of this for loop into new functions
	for height := initialHeight; height < numBlocks+initialHeight && !stopEarly; height++ {

		// Log the header time for future lookup
		pastTimes = append(pastTimes, header.Time)
		pastVoteInfos = append(pastVoteInfos, request.LastCommitInfo.Votes)

		// Run the BeginBlock handler
		logWriter.AddEntry(BeginBlockEntry(int64(height)))
		app.BeginBlock(request)

		if testingMode {
			assertAllInvariants(t, app, invariants, "BeginBlock", logWriter, allInvariants)
		}

		ctx := app.NewContext(false, header)

		// Run queued operations. Ignores blocksize if blocksize is too small
		numQueuedOpsRan := runQueuedOperations(
			operationQueue, int(header.Height),
			tb, r, app, ctx, accs, logWriter, eventStats.Tally, lean)

		numQueuedTimeOpsRan := runQueuedTimeOperations(
			timeOperationQueue, int(header.Height), header.Time,
			tb, r, app, ctx, accs, logWriter, eventStats.Tally, lean)

		if testingMode && onOperation {
			assertAllInvariants(t, app, invariants, "QueuedOperations", logWriter, allInvariants)
		}

		// run standard operations
		operations := blockSimulator(r, app, ctx, accs, header)
		opCount += operations + numQueuedOpsRan + numQueuedTimeOpsRan
		if testingMode {
			assertAllInvariants(t, app, invariants, "StandardOperations", logWriter, allInvariants)
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
			assertAllInvariants(t, app, invariants, "EndBlock", logWriter, allInvariants)
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
			pastTimes, pastVoteInfos, eventStats.Tally, header)

		// Update the validator set, which will be reflected in the application
		// on the next block
		validators = nextValidators
		nextValidators = updateValidators(tb, r, params,
			validators, res.ValidatorUpdates, eventStats.Tally)

		// update the exported params
		if exportParams && exportParamsHeight == height {
			exportedParams = params
		}
	}

	if stopEarly {
		if exportStatsPath != "" {
			fmt.Println("Exporting simulation statistics...")
			eventStats.ExportJSON(exportStatsPath)
		} else {
			eventStats.Print(w)
		}

		return true, exportedParams, err
	}

	fmt.Fprintf(
		w,
		"\nSimulation complete; Final height (blocks): %d, final time (seconds): %v, operations ran: %d\n",
		header.Height, header.Time, opCount,
	)

	if exportStatsPath != "" {
		fmt.Println("Exporting simulation statistics...")
		eventStats.ExportJSON(exportStatsPath)
	} else {
		eventStats.Print(w)
	}

	return false, exportedParams, nil
}

//______________________________________________________________________________

type blockSimFn func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
	accounts []Account, header abci.Header) (opCount int)

// Returns a function to simulate blocks. Written like this to avoid constant
// parameters being passed everytime, to minimize memory overhead.
func createBlockSimulator(testingMode bool, tb testing.TB, t *testing.T, w io.Writer, params Params,
	event func(route, op, evResult string), invariants sdk.Invariants, ops WeightedOperations,
	operationQueue OperationQueue, timeOperationQueue []FutureOperation,
	totalNumBlocks, avgBlockSize int, logWriter LogWriter, lean, onOperation, allInvariants bool) blockSimFn {

	lastBlockSizeState := 0 // state for [4 * uniform distribution]
	blocksize := 0
	selectOp := ops.getSelectOpFn()

	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accounts []Account, header abci.Header,
	) (opCount int) {

		_, _ = fmt.Fprintf(
			w, "\rSimulating... block %d/%d, operation %d/%d.",
			header.Height, totalNumBlocks, opCount, blocksize,
		)
		lastBlockSizeState, blocksize = getBlockSize(r, params, lastBlockSizeState, avgBlockSize)

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
				logWriter.AddEntry(MsgEntry(header.Height, int64(i), opMsg))
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
					assertAllInvariants(t, app, invariants, eventStr, logWriter, allInvariants)
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
	ctx sdk.Context, accounts []Account, logWriter LogWriter, event func(route, op, evResult string), lean bool) (numOpsRan int) {

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
		opMsg.LogEvent(event)
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
	logWriter LogWriter, event func(route, op, evResult string), lean bool) (numOpsRan int) {

	numOpsRan = 0
	for len(queueOps) > 0 && currentTime.After(queueOps[0].BlockTime) {

		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		opMsg, _, err := queueOps[0].Op(r, app, ctx, accounts)
		opMsg.LogEvent(event)
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
