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

// AppStateFn returns the app state json bytes and the genesis accounts
type AppStateFn func(r *rand.Rand, accs []Account, config Config) (
	appState json.RawMessage, accounts []Account, chainId string, genesisTimestamp time.Time,
)

// initialize the chain for the simulation
func initChain(
	r *rand.Rand, params Params, accounts []Account, app *baseapp.BaseApp,
	appStateFn AppStateFn, config Config,
) (mockValidators, time.Time, []Account, string) {

	appState, accounts, chainID, genesisTimestamp := appStateFn(r, accounts, config)

	req := abci.RequestInitChain{
		AppStateBytes: appState,
		ChainId:       chainID,
	}
	res := app.InitChain(req)
	validators := newMockValidators(r, res.Validators, params)

	return validators, genesisTimestamp, accounts, chainID
}

// SimulateFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided config.Seed.
// TODO: split this monster function up
func SimulateFromSeed(
	tb testing.TB, w io.Writer, app *baseapp.BaseApp,
	appStateFn AppStateFn, ops WeightedOperations,
	blackListedAccs map[string]bool, config Config,
) (stopEarly bool, exportedParams Params, err error) {

	// in case we have to end early, don't os.Exit so that we can run cleanup code.
	testingMode, t, b := getTestingMode(tb)
	fmt.Fprintf(w, "Starting SimulateFromSeed with randomness created with seed %d\n", int(config.Seed))

	r := rand.New(rand.NewSource(config.Seed))
	params := RandomParams(r)
	fmt.Fprintf(w, "Randomized simulation params: \n%s\n", mustMarshalJSONIndent(params))

	timeDiff := maxTimePerBlock - minTimePerBlock
	accs := RandomAccounts(r, params.NumKeys)
	eventStats := NewEventStats()

	// Second variable to keep pending validator set (delayed one block since
	// TM 0.24) Initially this is the same as the initial validator set
	validators, genesisTimestamp, accs, chainID := initChain(r, params, accs, app, appStateFn, config)
	if len(accs) == 0 {
		return true, params, fmt.Errorf("must have greater than zero genesis accounts")
	}

	config.ChainID = chainID

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
		ChainID:         config.ChainID,
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
		err = fmt.Errorf("exited due to %s", receivedSignal)
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
		testingMode, tb, t, w, params, eventStats.Tally,
		ops, operationQueue, timeOperationQueue, logWriter, config)

	if !testingMode {
		b.ResetTimer()
	} else {
		// recover logs in case of panic
		defer func() {
			if r := recover(); r != nil {
				_, _ = fmt.Fprintf(w, "simulation halted due to panic on block %d\n", header.Height)
				logWriter.PrintLogs()
				panic(r)
			}
		}()
	}

	// set exported params to the initial state
	if config.ExportParamsPath != "" && config.ExportParamsHeight == 0 {
		exportedParams = params
	}

	// TODO: split up the contents of this for loop into new functions
	for height := config.InitialBlockHeight; height < config.NumBlocks+config.InitialBlockHeight && !stopEarly; height++ {

		// Log the header time for future lookup
		pastTimes = append(pastTimes, header.Time)
		pastVoteInfos = append(pastVoteInfos, request.LastCommitInfo.Votes)

		// Run the BeginBlock handler
		logWriter.AddEntry(BeginBlockEntry(int64(height)))
		app.BeginBlock(request)

		ctx := app.NewContext(false, header)

		// Run queued operations. Ignores blocksize if blocksize is too small
		numQueuedOpsRan := runQueuedOperations(
			operationQueue, int(header.Height), tb, r, app, ctx, accs, logWriter,
			eventStats.Tally, config.Lean, config.ChainID,
		)

		numQueuedTimeOpsRan := runQueuedTimeOperations(
			timeOperationQueue, int(header.Height), header.Time,
			tb, r, app, ctx, accs, logWriter, eventStats.Tally,
			config.Lean, config.ChainID,
		)

		// run standard operations
		operations := blockSimulator(r, app, ctx, accs, header)
		opCount += operations + numQueuedOpsRan + numQueuedTimeOpsRan

		res := app.EndBlock(abci.RequestEndBlock{})
		header.Height++
		header.Time = header.Time.Add(
			time.Duration(minTimePerBlock) * time.Second)
		header.Time = header.Time.Add(
			time.Duration(int64(r.Intn(int(timeDiff)))) * time.Second)
		header.ProposerAddress = validators.randomProposer(r)
		logWriter.AddEntry(EndBlockEntry(int64(height)))

		if config.Commit {
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
		if config.ExportParamsPath != "" && config.ExportParamsHeight == height {
			exportedParams = params
		}
	}

	if stopEarly {
		if config.ExportStatsPath != "" {
			fmt.Println("Exporting simulation statistics...")
			eventStats.ExportJSON(config.ExportStatsPath)
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

	if config.ExportStatsPath != "" {
		fmt.Println("Exporting simulation statistics...")
		eventStats.ExportJSON(config.ExportStatsPath)
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
	event func(route, op, evResult string), ops WeightedOperations,
	operationQueue OperationQueue, timeOperationQueue []FutureOperation,
	logWriter LogWriter, config Config) blockSimFn {

	lastBlockSizeState := 0 // state for [4 * uniform distribution]
	blocksize := 0
	selectOp := ops.getSelectOpFn()

	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accounts []Account, header abci.Header,
	) (opCount int) {

		_, _ = fmt.Fprintf(
			w, "\rSimulating... block %d/%d, operation %d/%d.",
			header.Height, config.NumBlocks, opCount, blocksize,
		)
		lastBlockSizeState, blocksize = getBlockSize(r, params, lastBlockSizeState, config.BlockSize)

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
			opMsg, futureOps, err := op(r2, app, ctx, accounts, config.ChainID)
			opMsg.LogEvent(event)

			if !config.Lean || opMsg.OK {
				logWriter.AddEntry(MsgEntry(header.Height, int64(i), opMsg))
			}

			if err != nil {
				logWriter.PrintLogs()
				tb.Fatalf(`error on block  %d/%d, operation (%d/%d) from x/%s:
%v
Comment: %s`,
					header.Height, config.NumBlocks, opCount, blocksize, opMsg.Route, err, opMsg.Comment)
			}

			queueOperations(operationQueue, timeOperationQueue, futureOps)

			if testingMode && opCount%50 == 0 {
				fmt.Fprintf(w, "\rSimulating... block %d/%d, operation %d/%d. ",
					header.Height, config.NumBlocks, opCount, blocksize)
			}

			opCount++
		}
		return opCount
	}
}

// nolint: errcheck
func runQueuedOperations(queueOps map[int][]Operation,
	height int, tb testing.TB, r *rand.Rand, app *baseapp.BaseApp,
	ctx sdk.Context, accounts []Account, logWriter LogWriter,
	event func(route, op, evResult string), lean bool, chainID string) (numOpsRan int) {

	queuedOp, ok := queueOps[height]
	if !ok {
		return 0
	}

	numOpsRan = len(queuedOp)
	for i := 0; i < numOpsRan; i++ {

		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		opMsg, _, err := queuedOp[i](r, app, ctx, accounts, chainID)
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
	logWriter LogWriter, event func(route, op, evResult string),
	lean bool, chainID string) (numOpsRan int) {

	numOpsRan = 0
	for len(queueOps) > 0 && currentTime.After(queueOps[0].BlockTime) {

		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		opMsg, _, err := queueOps[0].Op(r, app, ctx, accounts, chainID)
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
