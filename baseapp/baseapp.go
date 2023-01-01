package baseapp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/mempool"
)

type (
	execMode uint8

	// StoreLoader defines a customizable function to control how we load the
	// CommitMultiStore from disk. This is useful for state migration, when
	// loading a datastore written with an older version of the software. In
	// particular, if a module changed the substore key name (or removed a substore)
	// between two versions of the software.
	StoreLoader func(ms storetypes.CommitMultiStore) error
)

const (
	TxHashContextKey   = sdk.ContextKey("tx-hash")
	TxMsgIdxContextKey = sdk.ContextKey("tx-msg-idx")
)

var _ abci.Application = (*BaseApp)(nil)

var _ servertypes.ABCI = (*BaseApp)(nil)

// BaseApp reflects the ABCI application implementation.
type BaseApp struct { //nolint: maligned
	// initialized on creation
	mu                sync.Mutex // mu protects the fields below.
	logger            log.Logger
	name              string // application name from abci.Info
	interfaceRegistry codectypes.InterfaceRegistry
	txDecoder         sdk.TxDecoder // unmarshal []byte into sdk.Tx

	anteHandler sdk.AnteHandler // ante handler for fee and auth
	postHandler sdk.AnteHandler // post handler, optional, e.g. for tips

	appStore
	baseappVersions
	peerFilters
	snapshotData
	abciData
	moduleRouter

	// volatile states:
	//
	// checkState is set on InitChain and reset on Commit
	// deliverState is set on InitChain and BeginBlock and set to nil on Commit
	checkState   *state // for CheckTx
	deliverState *state // for DeliverTx

	// paramStore is used to query for ABCI consensus parameters from an
	// application parameter store.
	paramStore ParamStore

	// queryGasLimit defines the maximum gas for queries; unbounded if 0.
	queryGasLimit uint64

	// The minimum gas prices a validator is willing to accept for processing a
	// transaction. This is mainly used for DoS and spam prevention.
	minGasPrices sdk.DecCoins

	// initialHeight is the initial height at which we start the BaseApp
	initialHeight int64

	// flag for sealing options and parameters to a BaseApp
	sealed bool

	// block height at which to halt the chain and gracefully shutdown
	haltHeight uint64

	// minimum block time (in Unix seconds) at which to halt the chain and gracefully shutdown
	haltTime uint64

	// minRetainBlocks defines the minimum block height offset from the current
	// block being committed, such that all blocks past this offset are pruned
	// from CometBFT. It is used as part of the process of determining the
	// ResponseCommit.RetainHeight value during ABCI Commit. A value of 0 indicates
	// that no blocks should be pruned.
	//
	// Note: CometBFT block pruning is dependent on this parameter in conjunction
	// with the unbonding (safety threshold) period, state pruning and state sync
	// snapshot parameters to determine the correct minimum value of
	// ResponseCommit.RetainHeight.
	minRetainBlocks uint64

	// recovery handler for app.runTx method
	runTxRecoveryMiddleware recoveryMiddleware

	// trace set will return full stack traces for errors in ABCI Log field
	trace bool

	// indexEvents defines the set of events in the form {eventType}.{attributeKey},
	// which informs CometBFT what to index. If empty, all events will be indexed.
	indexEvents map[string]struct{}

	// abciListeners for hooking into the ABCI message processing of the BaseApp
	// and exposing the requests and responses to external consumers
	abciListeners []ABCIListener
}

type appStore struct {
	db          dbm.DB               // common DB backend
	cms         sdk.CommitMultiStore // Main (uncached) state
	qms         sdk.MultiStore       // Optional alternative state provider for query service
	storeLoader StoreLoader          // function to handle store loading, may be overridden with SetStoreLoader()

	// an inter-block write-through cache provided to the context during deliverState
	interBlockCache sdk.MultiStorePersistentCache

	fauxMerkleMode bool // if true, IAVL MountStores uses MountStoresDB for simulation speed.
}

type moduleRouter struct {
	router           sdk.Router        // handle any kind of message
	queryRouter      sdk.QueryRouter   // router for redirecting query calls
	grpcQueryRouter  *GRPCQueryRouter  // router for redirecting gRPC query calls
	msgServiceRouter *MsgServiceRouter // router for redirecting Msg service messages
}

type abciData struct {
	initChainer  sdk.InitChainer  // initialize state with validators and state blob
	beginBlocker sdk.BeginBlocker // logic to run before any txs
	endBlocker   sdk.EndBlocker   // logic to run after all txs, and to determine valset changes

	// absent validators from begin block
	voteInfos []abci.VoteInfo
}

type baseappVersions struct {
	// application's version string
	version string

	// application's protocol version that increments on every upgrade
	// if BaseApp is passed to the upgrade keeper's NewKeeper method.
	appVersion uint64
}

// should really get handled in some db struct
// which then has a sub-item, persistence fields
type snapshotData struct {
	// manages snapshots, i.e. dumps of app state at certain intervals
	snapshotManager *snapshots.Manager
}

// NewBaseApp returns a reference to an initialized BaseApp. It accepts a
// variadic number of option functions, which act on the BaseApp to set
// configuration choices.
func NewBaseApp(
	name string, logger log.Logger, db corestore.KVStoreWithBatch, txDecoder sdk.TxDecoder, options ...func(*BaseApp),
) *BaseApp {
	app := &BaseApp{
		logger: logger,
		name:   name,
		appStore: appStore{
			db:             db,
			cms:            store.NewCommitMultiStore(db),
			storeLoader:    DefaultStoreLoader,
			fauxMerkleMode: false,
		},
		moduleRouter: moduleRouter{
			router:           NewRouter(),
			queryRouter:      NewQueryRouter(),
			grpcQueryRouter:  NewGRPCQueryRouter(),
			msgServiceRouter: NewMsgServiceRouter(),
		},
		txDecoder: txDecoder,
	}

	for _, option := range options {
		option(app)
	}

	if app.mempool == nil {
		app.SetMempool(mempool.NoOpMempool{})
	}

	abciProposalHandler := NewDefaultProposalHandler(app.mempool, app)

	if app.prepareProposal == nil {
		app.SetPrepareProposal(abciProposalHandler.PrepareProposalHandler())
	}
	if app.processProposal == nil {
		app.SetProcessProposal(abciProposalHandler.ProcessProposalHandler())
	}
	if app.extendVote == nil {
		app.SetExtendVoteHandler(NoOpExtendVote())
	}
	if app.verifyVoteExt == nil {
		app.SetVerifyVoteExtensionHandler(NoOpVerifyVoteExtensionHandler())
	}
	if app.interBlockCache != nil {
		app.cms.SetInterBlockCache(app.interBlockCache)
	}
	if app.includeNestedMsgsGas == nil {
		app.includeNestedMsgsGas = make(map[string]struct{})
	}
	app.runTxRecoveryMiddleware = newDefaultRecoveryMiddleware()

	// Initialize with an empty interface registry to avoid nil pointer dereference.
	// Unless SetInterfaceRegistry is called with an interface registry with proper address codecs baseapp will panic.
	app.cdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	return app
}

// Name returns the name of the BaseApp.
func (app *BaseApp) Name() string {
	return app.name
}

// AppVersion returns the application's protocol version.
func (app *BaseApp) AppVersion(ctx context.Context) (uint64, error) {
	if app.versionModifier == nil {
		return 0, errors.New("app.versionModifier is nil")
	}

	return app.versionModifier.AppVersion(ctx)
}

// Version returns the application's version string.
func (app *BaseApp) Version() string {
	return app.version
}

// Logger returns the logger of the BaseApp.
func (app *BaseApp) Logger() log.Logger {
	return app.logger
}

// Trace returns the boolean value for logging error stack traces.
func (app *BaseApp) Trace() bool {
	return app.trace
}

// MsgServiceRouter returns the MsgServiceRouter of a BaseApp.
func (app *BaseApp) MsgServiceRouter() *MsgServiceRouter { return app.msgServiceRouter }

// SetMsgServiceRouter sets the MsgServiceRouter of a BaseApp.
func (app *BaseApp) SetMsgServiceRouter(msgServiceRouter *MsgServiceRouter) {
	app.msgServiceRouter = msgServiceRouter
}

// MountStores mounts all IAVL or DB stores to the provided keys in the BaseApp
// multistore.
func (app *BaseApp) MountStores(keys ...storetypes.StoreKey) {
	for _, key := range keys {
		switch key.(type) {
		case *storetypes.KVStoreKey:
			if !app.fauxMerkleMode {
				app.MountStore(key, storetypes.StoreTypeIAVL)
			} else {
				// StoreTypeDB doesn't do anything upon commit, and it doesn't
				// retain history, but it's useful for faster simulation.
				app.MountStore(key, storetypes.StoreTypeDB)
			}

		case *storetypes.TransientStoreKey:
			app.MountStore(key, storetypes.StoreTypeTransient)

		case *storetypes.MemoryStoreKey:
			app.MountStore(key, storetypes.StoreTypeMemory)

		default:
			panic(fmt.Sprintf("Unrecognized store key type :%T", key))
		}
	}
}

// MountKVStores mounts all IAVL or DB stores to the provided keys in the
// BaseApp multistore.
func (app *BaseApp) MountKVStores(keys map[string]*storetypes.KVStoreKey) {
	for _, key := range keys {
		if !app.fauxMerkleMode {
			app.MountStore(key, storetypes.StoreTypeIAVL)
		} else {
			// StoreTypeDB doesn't do anything upon commit, and it doesn't
			// retain history, but it's useful for faster simulation.
			app.MountStore(key, storetypes.StoreTypeDB)
		}
	}
}

// MountTransientStores mounts all transient stores to the provided keys in
// the BaseApp multistore.
func (app *BaseApp) MountTransientStores(keys map[string]*storetypes.TransientStoreKey) {
	for _, key := range keys {
		app.MountStore(key, storetypes.StoreTypeTransient)
	}
}

// MountMemoryStores mounts all in-memory KVStores with the BaseApp's internal
// commit multi-store.
func (app *BaseApp) MountMemoryStores(keys map[string]*storetypes.MemoryStoreKey) {
	skeys := slices.Sorted(maps.Keys(keys))
	for _, key := range skeys {
		memKey := keys[key]
		app.MountStore(memKey, storetypes.StoreTypeMemory)
	}
}

// MountStore mounts a store to the provided key in the BaseApp multistore,
// using the default DB.
func (app *BaseApp) MountStore(key storetypes.StoreKey, typ storetypes.StoreType) {
	app.cms.MountStoreWithDB(key, typ, nil)
}

// LoadLatestVersion loads the latest application version. It will panic if
// called more than once on a running BaseApp.
func (app *BaseApp) LoadLatestVersion() error {
	err := app.storeLoader(app.cms)
	if err != nil {
		return fmt.Errorf("failed to load latest version: %w", err)
	}

	return app.Init()
}

// DefaultStoreLoader will be used by default and loads the latest version
func DefaultStoreLoader(ms storetypes.CommitMultiStore) error {
	return ms.LoadLatestVersion()
}

// CommitMultiStore returns the root multi-store.
// App constructor can use this to access the `cms`.
// UNSAFE: must not be used during the abci life cycle.
func (app *BaseApp) CommitMultiStore() sdk.CommitMultiStore {
	return app.cms
}

// SnapshotManager returns the snapshot manager.
// application use this to register extra extension snapshotters.
func (app *BaseApp) SnapshotManager() *snapshots.Manager {
	return app.snapshotManager
}

// LoadVersion loads the BaseApp application version. It will panic if called
// more than once on a running baseapp.
func (app *BaseApp) LoadVersion(version int64) error {
	app.logger.Info("NOTICE: this could take a long time to migrate IAVL store to fastnode if you enable Fast Node.\n")
	err := app.cms.LoadVersion(version)
	if err != nil {
		return fmt.Errorf("failed to load version %d: %w", version, err)
	}

	return app.Init()
}

// LastCommitID returns the last CommitID of the multistore.
func (app *BaseApp) LastCommitID() storetypes.CommitID {
	return app.cms.LastCommitID()
}

// LastBlockHeight returns the last committed block height.
func (app *BaseApp) LastBlockHeight() int64 {
	return app.cms.LastCommitID().Version
}

// ChainID returns the chainID of the app.
func (app *BaseApp) ChainID() string {
	return app.chainID
}

// AnteHandler returns the AnteHandler of the app.
func (app *BaseApp) AnteHandler() sdk.AnteHandler {
	return app.anteHandler
}

// Mempool returns the Mempool of the app.
func (app *BaseApp) Mempool() mempool.Mempool {
	return app.mempool
}

// Init initializes the app. It seals the app, preventing any
// further modifications. In addition, it validates the app against
// the earlier provided settings. Returns an error if validation fails.
// nil otherwise. Panics if the app is already sealed.
func (app *BaseApp) Init() error {
	if app.sealed {
		panic("cannot call initFromMainStore: baseapp already sealed")
	}

	if app.cms == nil {
		return errors.New("commit multi-store must not be nil")
	}

	emptyHeader := cmtproto.Header{ChainID: app.chainID}

	// needed for the export command which inits from store but never calls initchain
	app.setState(execModeCheck, emptyHeader)
	app.Seal()

	return app.cms.GetPruning().Validate()
}

func (app *BaseApp) setMinGasPrices(gasPrices sdk.DecCoins) {
	app.minGasPrices = gasPrices
}

func (app *BaseApp) setHaltHeight(haltHeight uint64) {
	app.haltHeight = haltHeight
}

func (app *BaseApp) setHaltTime(haltTime uint64) {
	app.haltTime = haltTime
}

func (app *BaseApp) setMinRetainBlocks(minRetainBlocks uint64) {
	app.minRetainBlocks = minRetainBlocks
}

func (app *BaseApp) setInterBlockCache(cache storetypes.MultiStorePersistentCache) {
	app.interBlockCache = cache
}

func (app *BaseApp) setTrace(trace bool) {
	app.trace = trace
}

func (app *BaseApp) setIndexEvents(ie []string) {
	app.indexEvents = make(map[string]struct{}, len(ie))

	for _, e := range ie {
		app.indexEvents[e] = struct{}{}
	}
}

// Seal seals a BaseApp. It prohibits any further modifications to a BaseApp.
func (app *BaseApp) Seal() { app.sealed = true }

// IsSealed returns true if the BaseApp is sealed and false otherwise.
func (app *BaseApp) IsSealed() bool { return app.sealed }

// setState sets the BaseApp's state for the corresponding mode with a branched
// multi-store (i.e. a CacheMultiStore) and a new Context with the same
// multi-store branch, and provided header.
func (app *BaseApp) setState(mode execMode, h cmtproto.Header) {
	ms := app.cms.CacheMultiStore()
	app.checkState = &state{
		ms:           ms,
		ctx:          sdk.NewContext(ms, header, true, app.logger).WithMinGasPrices(app.minGasPrices),
		eventHistory: []abci.Event{},
	}
}

// setDeliverState sets the BaseApp's deliverState with a branched multi-store
// (i.e. a CacheMultiStore) and a new Context with the same multi-store branch,
// and provided header. It is set on InitChain and BeginBlock and set to nil on
// Commit.
func (app *BaseApp) setDeliverState(header tmproto.Header) {
	ms := app.cms.CacheMultiStore()
	app.deliverState = &state{
		ms:           ms,
		ctx:          sdk.NewContext(ms, header, false, app.logger),
		eventHistory: []abci.Event{},
	}
	app.msgServiceRouter.SetCircuit(cb)
}

// GetConsensusParams returns the current consensus parameters from the BaseApp's
// ParamStore. If the BaseApp has no ParamStore defined, nil is returned.
func (app *BaseApp) GetConsensusParams(ctx sdk.Context) *abci.ConsensusParams {
	if app.paramStore == nil {
		return nil
	}

	cp := new(abci.ConsensusParams)

	if app.paramStore.Has(ctx, ParamStoreKeyBlockParams) {
		var bp abci.BlockParams

		app.paramStore.Get(ctx, ParamStoreKeyBlockParams, &bp)
		cp.Block = &bp
	}

	cp, err := app.paramStore.Get(ctx)
	if err != nil {
		// This could happen while migrating from v0.45/v0.46 to v0.50, we should
		// allow it to happen so during preblock the upgrade plan can be executed
		// and the consensus params set for the first time in the new format.
		app.logger.Error("failed to get consensus params", "err", err)
		return cmtproto.ConsensusParams{}
	}

	return cp
}

// StoreConsensusParams sets the consensus parameters to the BaseApp's param
// store.
func (app *BaseApp) StoreConsensusParams(ctx context.Context, cp cmtproto.ConsensusParams) error {
	if app.paramStore == nil {
		return errors.New("cannot store consensus params with no params store set")
	}

	return app.paramStore.Set(ctx, cp)
}

// AddRunTxRecoveryHandler adds custom app.runTx method panic handlers.
func (app *BaseApp) AddRunTxRecoveryHandler(handlers ...RecoveryHandler) {
	for _, h := range handlers {
		app.runTxRecoveryMiddleware = newRecoveryMiddleware(h, app.runTxRecoveryMiddleware)
	}
}

// StoreConsensusParams sets the consensus parameters to the baseapp's param store.
func (app *BaseApp) StoreConsensusParams(ctx sdk.Context, cp *abci.ConsensusParams) {
	if app.paramStore == nil {
		panic("cannot store consensus params with no params store set")
	}

	if cp == nil {
		return
	}

	app.paramStore.Set(ctx, ParamStoreKeyBlockParams, cp.Block)
	app.paramStore.Set(ctx, ParamStoreKeyEvidenceParams, cp.Evidence)
	app.paramStore.Set(ctx, ParamStoreKeyValidatorParams, cp.Validator)
	// We're explicitly not storing the Tendermint app_version in the param store. It's
	// stored instead in the x/upgrade store, with its own bump logic.
}

// getMaximumBlockGas gets the maximum gas from the consensus params. It panics
// if maximum block gas is less than negative one and returns zero if negative
// one.
func (app *BaseApp) GetMaximumBlockGas(ctx sdk.Context) uint64 {
	cp := app.GetConsensusParams(ctx)
	if cp.Block == nil {
		return 0
	}

	maxGas := cp.Block.MaxGas

	switch {
	case maxGas < -1:
		panic(fmt.Sprintf("invalid maximum block gas: %d", maxGas))

	case maxGas == -1:
		return 0

	default:
		return uint64(maxGas)
	}
}

func (app *BaseApp) validateFinalizeBlockHeight(req *abci.FinalizeBlockRequest) error {
	if req.Height < 1 {
		return fmt.Errorf("invalid height: %d", req.Height)
	}

	lastBlockHeight := app.LastBlockHeight()

	// expectedHeight holds the expected height to validate
	var expectedHeight int64
	if lastBlockHeight == 0 && app.initialHeight > 1 {
		// In this case, we're validating the first block of the chain, i.e no
		// previous commit. The height we're expecting is the initial height.
		expectedHeight = app.initialHeight
	} else {
		// This case can mean two things:
		//
		// - Either there was already a previous commit in the store, in which
		// case we increment the version from there.
		// - Or there was no previous commit, in which case we start at version 1.
		expectedHeight = lastBlockHeight + 1
	}

	if req.Height != expectedHeight {
		return fmt.Errorf("invalid height: %d; expected: %d", req.Height, expectedHeight)
	}

	return nil
}

// validateBasicTxMsgs executes basic validator calls for messages firstly by invoking
// .ValidateBasic if possible, then checking if the message has a known handler.
func validateBasicTxMsgs(router *MsgServiceRouter, msgs []sdk.Msg) error {
	if len(msgs) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "must contain at least one message")
	}

	for _, msg := range msgs {
		if m, ok := msg.(sdk.HasValidateBasic); ok {
			if err := m.ValidateBasic(); err != nil {
				return err
			}
		}

		if router != nil && router.Handler(msg) == nil {
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "no message handler found for %T", msg)
		}
	}

	return nil
}

func (app *BaseApp) getState(mode execMode) *state {
	switch mode {
	case execModeFinalize:
		return app.finalizeBlockState

	case execModePrepareProposal:
		return app.prepareProposalState

	case execModeProcessProposal:
		return app.processProposalState

	default:
		return app.checkState
	}
}

func (app *BaseApp) getBlockGasMeter(ctx sdk.Context) storetypes.GasMeter {
	if maxGas := app.GetMaximumBlockGas(ctx); maxGas > 0 {
		return storetypes.NewGasMeter(maxGas)
	}

	return storetypes.NewInfiniteGasMeter()
}

// retrieve the context for the tx w/ txBytes and other memoized values.
func (app *BaseApp) getContextForTx(mode execMode, txBytes []byte) sdk.Context {
	app.mu.Lock()
	defer app.mu.Unlock()

	modeState := app.getState(mode)
	if modeState == nil {
		panic(fmt.Sprintf("state is nil for mode %v", mode))
	}
	ctx := modeState.Context().
		WithTxBytes(txBytes).
		WithGasMeter(storetypes.NewInfiniteGasMeter())

	ctx = ctx.WithIsSigverifyTx(app.sigverifyTx)

	ctx = ctx.WithConsensusParams(app.GetConsensusParams(ctx))

	if mode == execModeReCheck {
		ctx = ctx.WithIsReCheckTx(true)
	}

	if mode == execModeSimulate {
		ctx, _ = ctx.CacheContext()
		ctx = ctx.WithExecMode(sdk.ExecMode(execModeSimulate))
	}

	return ctx
}

// cacheTxContext returns a new context based off of the provided context with
// a branched multi-store.
func (app *BaseApp) cacheTxContext(ctx sdk.Context, txBytes []byte) (sdk.Context, sdk.CacheMultiStore) {
	txHash, ok := ctx.Context().Value(TxHashContextKey).(string)
	if !ok {
		txHash = fmt.Sprintf("%X", tmhash.Sum(txBytes))
		ctx = ctx.WithContext(context.WithValue(ctx.Context(), TxHashContextKey, txHash))
	}
	ms := ctx.MultiStore()
	// TODO: https://github.com/cosmos/cosmos-sdk/issues/2824
	msCache := ms.CacheMultiStore()
	if msCache.TracingEnabled() {
		msCache = msCache.SetTracingContext(
			storetypes.TraceContext(
				map[string]interface{}{
					"txHash": txHash,
				},
			),
		).(storetypes.CacheMultiStore)
	}

	return ctx.WithMultiStore(msCache), msCache
}

func (app *BaseApp) preBlock(req *abci.FinalizeBlockRequest) ([]abci.Event, error) {
	var events []abci.Event
	if app.preBlocker != nil {
		ctx := app.finalizeBlockState.Context().WithEventManager(sdk.NewEventManager())
		if err := app.preBlocker(ctx, req); err != nil {
			return nil, err
		}
		// ConsensusParams can change in preblocker, so we need to
		// write the consensus parameters in store to context
		ctx = ctx.WithConsensusParams(app.GetConsensusParams(ctx))
		// GasMeter must be set after we get a context with updated consensus params.
		gasMeter := app.getBlockGasMeter(ctx)
		ctx = ctx.WithBlockGasMeter(gasMeter)
		app.finalizeBlockState.SetContext(ctx)
		events = ctx.EventManager().ABCIEvents()
	}
	return events, nil
}

func (app *BaseApp) beginBlock(_ *abci.FinalizeBlockRequest) (sdk.BeginBlock, error) {
	var (
		resp sdk.BeginBlock
		err  error
	)

	if app.beginBlocker != nil {
		resp, err = app.beginBlocker(app.finalizeBlockState.Context())
		if err != nil {
			return resp, err
		}

		// append BeginBlock attributes to all events in the BeginBlock response
		for i, event := range resp.Events {
			resp.Events[i].Attributes = append(
				event.Attributes,
				abci.EventAttribute{Key: "mode", Value: "BeginBlock"},
			)
		}

		resp.Events = sdk.MarkEventsToIndex(resp.Events, app.indexEvents)
	}

	return resp, nil
}

func (app *BaseApp) deliverTx(tx []byte) *abci.ExecTxResult {
	gInfo := sdk.GasInfo{}
	resultStr := "successful"

	var resp *abci.ExecTxResult

	defer func() {
		telemetry.IncrCounter(1, "tx", "count")
		telemetry.IncrCounter(1, "tx", resultStr)
		telemetry.SetGauge(float32(gInfo.GasUsed), "tx", "gas", "used")
		telemetry.SetGauge(float32(gInfo.GasWanted), "tx", "gas", "wanted")
	}()

	gInfo, result, anteEvents, err := app.runTx(execModeFinalize, tx)
	if err != nil {
		resultStr = "failed"
		resp = responseExecTxResultWithEvents(
			err,
			gInfo.GasWanted,
			gInfo.GasUsed,
			sdk.MarkEventsToIndex(anteEvents, app.indexEvents),
			app.trace,
		)
		return resp
	}

	resp = &abci.ExecTxResult{
		GasWanted: int64(gInfo.GasWanted),
		GasUsed:   int64(gInfo.GasUsed),
		Log:       result.Log,
		Data:      result.Data,
		Events:    sdk.MarkEventsToIndex(result.Events, app.indexEvents),
	}

	return resp
}

// endBlock is an application-defined function that is called after transactions
// have been processed in FinalizeBlock.
func (app *BaseApp) endBlock(_ context.Context) (sdk.EndBlock, error) {
	var endblock sdk.EndBlock

	if app.endBlocker != nil {
		eb, err := app.endBlocker(app.finalizeBlockState.Context())
		if err != nil {
			return endblock, err
		}

		// append EndBlock attributes to all events in the EndBlock response
		for i, event := range eb.Events {
			eb.Events[i].Attributes = append(
				event.Attributes,
				abci.EventAttribute{Key: "mode", Value: "EndBlock"},
			)
		}

		eb.Events = sdk.MarkEventsToIndex(eb.Events, app.indexEvents)
		endblock = eb
	}

	return endblock, nil
}

type HasNestedMsgs interface {
	GetMsgs() ([]sdk.Msg, error)
}

// runTx processes a transaction within a given execution mode, encoded transaction
// bytes, and the decoded transaction itself. All state transitions occur through
// a cached Context depending on the mode provided. State only gets persisted
// if all messages get executed successfully and the execution mode is DeliverTx.
// Note, gas execution info is always returned. A reference to a Result is
// returned if the tx does not run out of gas and if all the messages are valid
// and execute successfully. An error is returned otherwise.
func (app *BaseApp) runTx(mode runTxMode, txBytes []byte) (gInfo sdk.GasInfo, result *sdk.Result, anteEvents []abci.Event, err error) {
	// NOTE: GasWanted should be returned by the AnteHandler. GasUsed is
	// determined by the GasMeter. We need access to the context to get the gas
	// meter, so we initialize upfront.
	var gasWanted uint64

	ctx := app.getContextForTx(mode, txBytes)
	ms := ctx.MultiStore()

	// only run the tx if there is block gas remaining
	if mode == runTxModeDeliver && ctx.BlockGasMeter().IsOutOfGas() {
		return gInfo, nil, nil, sdkerrors.Wrap(sdkerrors.ErrOutOfGas, "no block gas left to run tx")
	}

	defer func() {
		if r := recover(); r != nil {
			recoveryMW := newOutOfGasRecoveryMiddleware(gasWanted, ctx, app.runTxRecoveryMiddleware)
			err, result = processRecovery(r, recoveryMW), nil
			ctx.Logger().Error("panic recovered in runTx", "err", err)
		}

		gInfo = sdk.GasInfo{GasWanted: gasWanted, GasUsed: ctx.GasMeter().GasConsumed()}
	}()

	blockGasConsumed := false
	// consumeBlockGas makes sure block gas is consumed at most once. It must happen after
	// tx processing, and must be execute even if tx processing fails. Hence we use trick with `defer`
	consumeBlockGas := func() {
		if !blockGasConsumed {
			blockGasConsumed = true
			ctx.BlockGasMeter().ConsumeGas(
				ctx.GasMeter().GasConsumedToLimit(), "block gas meter",
			)
		}
	}

	// If BlockGasMeter() panics it will be caught by the above recover and will
	// return an error - in any case BlockGasMeter will consume gas past the limit.
	//
	// NOTE: This must exist in a separate defer function for the above recovery
	// to recover from this one.
	if mode == runTxModeDeliver {
		defer consumeBlockGas()
	}

	tx, err := app.txDecoder(txBytes)
	if err != nil {
		return sdk.GasInfo{}, nil, nil, err
	}

	msgs := tx.GetMsgs()
	if err := validateBasicTxMsgs(msgs); err != nil {
		return sdk.GasInfo{}, nil, nil, err
	}

	if app.anteHandler != nil {
		var (
			anteCtx sdk.Context
			msCache storetypes.CacheMultiStore
		)

		// Branch context before AnteHandler call in case it aborts.
		// This is required for both CheckTx and DeliverTx.
		// Ref: https://github.com/cosmos/cosmos-sdk/issues/2772
		//
		// NOTE: Alternatively, we could require that AnteHandler ensures that
		// writes do not happen if aborted/failed.  This may have some
		// performance benefits, but it'll be more difficult to get right.
		anteCtx, msCache = app.cacheTxContext(ctx, txBytes)
		anteCtx = anteCtx.WithEventManager(sdk.NewEventManager())
		if mode == execModeSimulate {
			anteCtx = anteCtx.WithExecMode(sdk.ExecMode(execModeSimulate))
		}
		newCtx, err := app.anteHandler(anteCtx, tx, mode == execModeSimulate)

		if !newCtx.IsZero() {
			// At this point, newCtx.MultiStore() is a store branch, or something else
			// replaced by the AnteHandler. We want the original multistore.
			//
			// Also, in the case of the tx aborting, we need to track gas consumed via
			// the instantiated gas meter in the AnteHandler, so we update the context
			// prior to returning.
			ctx = newCtx.WithMultiStore(ms)
		}

		events := ctx.EventManager().Events()

		// GasMeter expected to be set in AnteHandler
		gasWanted = ctx.GasMeter().Limit()

		if err != nil {
			return gInfo, nil, nil, err
		}

		msCache.Write()
		anteEvents = events.ToABCIEvents()
	}

	// Create a new Context based off of the existing Context with a MultiStore branch
	// in case message processing fails. At this point, the MultiStore
	// is a branch of a branch.
	runMsgCtx, msCache := app.cacheTxContext(ctx, txBytes)

	// Attempt to execute all messages and only update state if all messages pass
	// and we're in DeliverTx. Note, runMsgs will never return a reference to a
	// Result if any single message fails or does not have a registered Handler.
	result, err = app.runMsgs(runMsgCtx, msgs, mode)
	if err == nil {
		// Run optional postHandlers.
		//
		// Note: If the postHandler fails, we also revert the runMsgs state.
		if app.postHandler != nil {
			// The runMsgCtx context currently contains events emitted by the ante handler.
			// We clear this to correctly order events without duplicates.
			// Note that the state is still preserved.
			postCtx := runMsgCtx.WithEventManager(sdk.NewEventManager())

			newCtx, err := app.postHandler(postCtx, tx, mode == runTxModeSimulate)
			if err != nil {
				return gInfo, nil, anteEvents, priority, err
			}

			result.Events = append(result.Events, newCtx.EventManager().ABCIEvents()...)
		}

		if mode == runTxModeDeliver {
			// When block gas exceeds, it'll panic and won't commit the cached store.
			consumeBlockGas()

		msCache.Write()

		if len(anteEvents) > 0 {
			// append the events in the order of occurrence
			result.Events = append(anteEvents, result.Events...)
		}
	}

	return gInfo, result, anteEvents, err
}

// runMsgs iterates through a list of messages and executes them with the provided
// Context and execution mode. Messages will only be executed during simulation
// and DeliverTx. An error is returned if any single message fails or if a
// Handler does not exist for a given message route. Otherwise, a reference to a
// Result is returned. The caller must not commit state if an error is returned.
func (app *BaseApp) runMsgs(ctx sdk.Context, msgs []sdk.Msg, reflectMsgs []protoreflect.Message, mode execMode) (*sdk.Result, error) {
	events := sdk.EmptyEvents()
	historicalEvents := ctx.EventManager().GetABCIEventHistory()
	txMsgData := &sdk.TxMsgData{
		Data: make([]*sdk.MsgData, 0, len(msgs)),
	}

	// NOTE: GasWanted is determined by the AnteHandler and GasUsed by the GasMeter.
	for i, msg := range msgs {
		if mode != execModeFinalize && mode != execModeSimulate {
			break
		}

		var (
			msgResult    *sdk.Result
			eventMsgName string // name to use as value in event `message.action`
			err          error
		)

		msgCtx := ctx.WithEventManager(sdk.NewEventManagerWithHistory(historicalEvents))
		msgCtx = msgCtx.WithContext(context.WithValue(msgCtx.Context(), TxMsgIdxContextKey, i))
		if handler := app.msgServiceRouter.Handler(msg); handler != nil {
			// ADR 031 request type routing
			msgResult, err = handler(msgCtx, msg)
			eventMsgName = sdk.MsgTypeURL(msg)
		} else if legacyMsg, ok := msg.(legacytx.LegacyMsg); ok {
			// legacy sdk.Msg routing
			// Assuming that the app developer has migrated all their Msgs to
			// proto messages and has registered all `Msg services`, then this
			// path should never be called, because all those Msgs should be
			// registered within the `msgServiceRouter` already.
			msgRoute := legacyMsg.Route()
			eventMsgName = legacyMsg.Type()
			handler := app.router.Route(msgCtx, msgRoute)
			if handler == nil {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message route: %s; message index: %d", msgRoute, i)
			}

			msgResult, err = handler(msgCtx, msg)
		} else {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "can't route message %+v", msg)
		}

		// ADR 031 request type routing
		msgResult, err := handler(ctx, msg)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "failed to execute message; message index: %d", i)
		}

		// create message events
		msgEvents, err := createEvents(app.cdc, msgResult.GetEvents(), msg, reflectMsgs[i])
		if err != nil {
			return nil, errorsmod.Wrapf(err, "failed to create message events; message index: %d", i)
		}

		// append message events and data
		//
		// Note: Each message result's data must be length-prefixed in order to
		// separate each result.
		for j, event := range msgEvents {
			// append message index to all events
			msgEvents[j] = event.AppendAttributes(sdk.NewAttribute("msg_index", strconv.Itoa(i)))
		}

		events = events.AppendEvents(msgEvents)
		historicalEvents = append(historicalEvents, msgEvents.ToABCIEvents()...)

		// Each individual sdk.Result that went through the MsgServiceRouter
		// (which should represent 99% of the Msgs now, since everyone should
		// be using protobuf Msgs) has exactly one Msg response.
		// We take that Msg response, and aggregate it into an array.
		if len(msgResult.MsgResponses) > 0 {
			msgResponse := msgResult.MsgResponses[0]
			if msgResponse == nil {
				return nil, sdkerrors.ErrLogic.Wrapf("got nil Msg response at index %d for msg %s", i, sdk.MsgTypeURL(msg))
			}
			msgResponses = append(msgResponses, msgResponse)
		}
	}

	data, err := makeABCIData(msgResponses)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to marshal tx data")
	}

	return &sdk.Result{
		Data:         data,
		Events:       events.ToABCIEvents(),
		MsgResponses: msgResponses,
	}, nil
}

// simulateNestedMessages simulates a message nested messages.
func (app *BaseApp) simulateNestedMessages(ctx sdk.Context, msg sdk.Msg) error {
	nestedMsgs, ok := msg.(HasNestedMsgs)
	if !ok {
		return nil
	}

	msgs, err := nestedMsgs.GetMsgs()
	if err != nil {
		return err
	}

	if err := validateBasicTxMsgs(app.msgServiceRouter, msgs); err != nil {
		return err
	}

	for _, msg := range msgs {
		err = app.simulateNestedMessages(ctx, msg)
		if err != nil {
			return err
		}
	}

	protoMessages := make([]protoreflect.Message, len(msgs))
	for i, msg := range msgs {
		_, protoMsg, err := app.cdc.GetMsgSigners(msg)
		if err != nil {
			return err
		}
		protoMessages[i] = protoMsg
	}

	initialGas := ctx.GasMeter().GasConsumed()
	_, err = app.runMsgs(ctx, msgs, protoMessages, execModeSimulate)
	if err == nil {
		if _, includeGas := app.includeNestedMsgsGas[sdk.MsgTypeURL(msg)]; !includeGas {
			consumedGas := ctx.GasMeter().GasConsumed() - initialGas
			ctx.GasMeter().RefundGas(consumedGas, "simulation of nested messages")
		}
	}
	return err
}

// makeABCIData generates the Data field to be sent to ABCI Check/DeliverTx.
func makeABCIData(msgResponses []*codectypes.Any) ([]byte, error) {
	return proto.Marshal(&sdk.TxMsgData{MsgResponses: msgResponses})
}

func createEvents(cdc codec.Codec, events sdk.Events, msg sdk.Msg, reflectMsg protoreflect.Message) (sdk.Events, error) {
	eventMsgName := sdk.MsgTypeURL(msg)
	msgEvent := sdk.NewEvent(sdk.EventTypeMessage, sdk.NewAttribute(sdk.AttributeKeyAction, eventMsgName))

	// we set the signer attribute as the sender
	signers, err := cdc.GetReflectMsgSigners(reflectMsg)
	if err != nil {
		return nil, err
	}
	if len(signers) > 0 && signers[0] != nil {
		addrStr, err := cdc.InterfaceRegistry().SigningContext().AddressCodec().BytesToString(signers[0])
		if err != nil {
			return nil, err
		}
		msgEvent = msgEvent.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeySender, addrStr))
	}

	// verify that events have no module attribute set
	if _, found := events.GetAttributes(sdk.AttributeKeyModule); !found {
		if moduleName := sdk.GetModuleNameFromTypeURL(eventMsgName); moduleName != "" {
			msgEvent = msgEvent.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyModule, moduleName))
		}
	}

	return sdk.Events{msgEvent}.AppendEvents(events), nil
}

// PrepareProposalVerifyTx performs transaction verification when a proposer is
// creating a block proposal during PrepareProposal. Any state committed to the
// PrepareProposal state internally will be discarded. <nil, err> will be
// returned if the transaction cannot be encoded. <bz, nil> will be returned if
// the transaction is valid, otherwise <bz, err> will be returned.
func (app *BaseApp) PrepareProposalVerifyTx(tx sdk.Tx) ([]byte, error) {
	bz, err := app.txEncoder(tx)
	if err != nil {
		return nil, err
	}

	_, _, _, err = app.runTx(execModePrepareProposal, bz)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// ProcessProposalVerifyTx performs transaction verification when receiving a
// block proposal during ProcessProposal. Any state committed to the
// ProcessProposal state internally will be discarded. <nil, err> will be
// returned if the transaction cannot be decoded. <Tx, nil> will be returned if
// the transaction is valid, otherwise <Tx, err> will be returned.
func (app *BaseApp) ProcessProposalVerifyTx(txBz []byte) (sdk.Tx, error) {
	tx, err := app.txDecoder(txBz)
	if err != nil {
		return nil, err
	}

	_, _, _, err = app.runTx(execModeProcessProposal, txBz)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (app *BaseApp) TxDecode(txBytes []byte) (sdk.Tx, error) {
	return app.txDecoder(txBytes)
}

func (app *BaseApp) TxEncode(tx sdk.Tx) ([]byte, error) {
	return app.txEncoder(tx)
}

// Close is called in start cmd to gracefully cleanup resources.
func (app *BaseApp) Close() error {
	var errs []error

	// Close app.db (opened by cosmos-sdk/server/start.go call to openDB)
	if app.db != nil {
		app.logger.Info("Closing application.db")
		if err := app.db.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Close app.snapshotManager
	// - opened when app chains use cosmos-sdk/server/util.go/DefaultBaseappOptions (boilerplate)
	// - which calls cosmos-sdk/server/util.go/GetSnapshotStore
	// - which is passed to baseapp/options.go/SetSnapshot
	// - to set app.snapshotManager = snapshots.NewManager
	if app.snapshotManager != nil {
		app.logger.Info("Closing snapshots/metadata.db")
		if err := app.snapshotManager.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// GetBaseApp returns the pointer to itself.
func (app *BaseApp) GetBaseApp() *BaseApp {
	return app
}
