package baseapp

import (
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
)

// Key to store the consensus params in the main store.
var mainConsensusParamsKey = []byte("consensus_params")

// Enum mode for app.runTx
type runTxMode uint8

const (
	// Check a transaction
	runTxModeCheck runTxMode = iota
	// Simulate a transaction
	runTxModeSimulate runTxMode = iota
	// Deliver a transaction
	runTxModeDeliver runTxMode = iota
)

// BaseApp reflects the ABCI application implementation.
type BaseApp struct {
	// initialized on creation
	Logger      log.Logger
	name        string               // application name from abci.Info
	db          dbm.DB               // common DB backend
	cms         sdk.CommitMultiStore // Main (uncached) state
	router      Router               // handle any kind of message
	queryRouter QueryRouter          // router for redirecting query calls
	txDecoder   sdk.TxDecoder        // unmarshal []byte into sdk.Tx

	// set upon LoadVersion or LoadLatestVersion.
	mainKey *sdk.KVStoreKey // Main KVStore in cms

	// may be nil
	anteHandler      sdk.AnteHandler  // ante handler for fee and auth
	initChainer      sdk.InitChainer  // initialize state with validators and state blob
	beginBlocker     sdk.BeginBlocker // logic to run before any txs
	endBlocker       sdk.EndBlocker   // logic to run after all txs, and to determine valset changes
	addrPeerFilter   sdk.PeerFilter   // filter peers by address and port
	pubkeyPeerFilter sdk.PeerFilter   // filter peers by public key
	fauxMerkleMode   bool             // if true, IAVL MountStores uses MountStoresDB for simulation speed.

	//--------------------
	// Volatile
	// checkState is set on initialization and reset on Commit.
	// deliverState is set in InitChain and BeginBlock and cleared on Commit.
	// See methods setCheckState and setDeliverState.
	checkState   *state          // for CheckTx
	deliverState *state          // for DeliverTx
	voteInfos    []abci.VoteInfo // absent validators from begin block

	// consensus params
	// TODO move this in the future to baseapp param store on main store.
	consensusParams *abci.ConsensusParams

	// spam prevention
	minimumFees sdk.Coins

	// flag for sealing
	sealed bool
}

var _ abci.Application = (*BaseApp)(nil)

// NewBaseApp returns a reference to an initialized BaseApp.
//
// NOTE: The db is used to store the version number for now.
// Accepts a user-defined txDecoder
// Accepts variable number of option functions, which act on the BaseApp to set configuration choices
func NewBaseApp(name string, logger log.Logger, db dbm.DB, txDecoder sdk.TxDecoder, options ...func(*BaseApp)) *BaseApp {
	app := &BaseApp{
		Logger:         logger,
		name:           name,
		db:             db,
		cms:            store.NewCommitMultiStore(db),
		router:         NewRouter(),
		queryRouter:    NewQueryRouter(),
		txDecoder:      txDecoder,
		fauxMerkleMode: false,
	}
	for _, option := range options {
		option(app)
	}
	return app
}

// BaseApp Name
func (app *BaseApp) Name() string {
	return app.name
}

// SetCommitMultiStoreTracer sets the store tracer on the BaseApp's underlying
// CommitMultiStore.
func (app *BaseApp) SetCommitMultiStoreTracer(w io.Writer) {
	app.cms.WithTracer(w)
}

// Mount IAVL or DB stores to the provided keys in the BaseApp multistore
func (app *BaseApp) MountStores(keys ...*sdk.KVStoreKey) {
	for _, key := range keys {
		if !app.fauxMerkleMode {
			app.MountStore(key, sdk.StoreTypeIAVL)
		} else {
			// StoreTypeDB doesn't do anything upon commit, and it doesn't
			// retain history, but it's useful for faster simulation.
			app.MountStore(key, sdk.StoreTypeDB)
		}
	}
}

// Mount stores to the provided keys in the BaseApp multistore
func (app *BaseApp) MountStoresTransient(keys ...*sdk.TransientStoreKey) {
	for _, key := range keys {
		app.MountStore(key, sdk.StoreTypeTransient)
	}
}

// Mount a store to the provided key in the BaseApp multistore, using a specified DB
func (app *BaseApp) MountStoreWithDB(key sdk.StoreKey, typ sdk.StoreType, db dbm.DB) {
	app.cms.MountStoreWithDB(key, typ, db)
}

// Mount a store to the provided key in the BaseApp multistore, using the default DB
func (app *BaseApp) MountStore(key sdk.StoreKey, typ sdk.StoreType) {
	app.cms.MountStoreWithDB(key, typ, nil)
}

// load latest application version
// panics if called more than once on a running baseapp
func (app *BaseApp) LoadLatestVersion(mainKey *sdk.KVStoreKey) error {
	err := app.cms.LoadLatestVersion()
	if err != nil {
		return err
	}
	return app.initFromMainStore(mainKey)
}

// load application version
// panics if called more than once on a running baseapp
func (app *BaseApp) LoadVersion(version int64, mainKey *sdk.KVStoreKey) error {
	err := app.cms.LoadVersion(version)
	if err != nil {
		return err
	}
	return app.initFromMainStore(mainKey)
}

// the last CommitID of the multistore
func (app *BaseApp) LastCommitID() sdk.CommitID {
	return app.cms.LastCommitID()
}

// the last committed block height
func (app *BaseApp) LastBlockHeight() int64 {
	return app.cms.LastCommitID().Version
}

// initializes the remaining logic from app.cms
func (app *BaseApp) initFromMainStore(mainKey *sdk.KVStoreKey) error {

	// main store should exist.
	mainStore := app.cms.GetKVStore(mainKey)
	if mainStore == nil {
		return errors.New("baseapp expects MultiStore with 'main' KVStore")
	}

	// memoize mainKey
	if app.mainKey != nil {
		panic("app.mainKey expected to be nil; duplicate init?")
	}
	app.mainKey = mainKey

	// load consensus params from the main store
	consensusParamsBz := mainStore.Get(mainConsensusParamsKey)
	if consensusParamsBz != nil {
		var consensusParams = &abci.ConsensusParams{}
		err := proto.Unmarshal(consensusParamsBz, consensusParams)
		if err != nil {
			panic(err)
		}
		app.setConsensusParams(consensusParams)
	} else {
		// It will get saved later during InitChain.
		// TODO assert that InitChain hasn't yet been called.
	}

	// Needed for `gaiad export`, which inits from store but never calls initchain
	app.setCheckState(abci.Header{})

	app.Seal()

	return nil
}

// SetMinimumFees sets the minimum fees.
func (app *BaseApp) SetMinimumFees(fees sdk.Coins) { app.minimumFees = fees }

// NewContext returns a new Context with the correct store, the given header, and nil txBytes.
func (app *BaseApp) NewContext(isCheckTx bool, header abci.Header) sdk.Context {
	if isCheckTx {
		return sdk.NewContext(app.checkState.ms, header, true, app.Logger).WithMinimumFees(app.minimumFees)
	}
	return sdk.NewContext(app.deliverState.ms, header, false, app.Logger)
}

type state struct {
	ms  sdk.CacheMultiStore
	ctx sdk.Context
}

func (st *state) CacheMultiStore() sdk.CacheMultiStore {
	return st.ms.CacheMultiStore()
}

func (st *state) Context() sdk.Context {
	return st.ctx
}

func (app *BaseApp) setCheckState(header abci.Header) {
	ms := app.cms.CacheMultiStore()
	app.checkState = &state{
		ms:  ms,
		ctx: sdk.NewContext(ms, header, true, app.Logger).WithMinimumFees(app.minimumFees),
	}
}

func (app *BaseApp) setDeliverState(header abci.Header) {
	ms := app.cms.CacheMultiStore()
	app.deliverState = &state{
		ms:  ms,
		ctx: sdk.NewContext(ms, header, false, app.Logger),
	}
}

// setConsensusParams memoizes the consensus params.
func (app *BaseApp) setConsensusParams(consensusParams *abci.ConsensusParams) {
	app.consensusParams = consensusParams
}

// setConsensusParams stores the consensus params to the main store.
func (app *BaseApp) storeConsensusParams(consensusParams *abci.ConsensusParams) {
	consensusParamsBz, err := proto.Marshal(consensusParams)
	if err != nil {
		panic(err)
	}
	mainStore := app.cms.GetKVStore(app.mainKey)
	mainStore.Set(mainConsensusParamsKey, consensusParamsBz)
}

// getMaximumBlockGas gets the maximum gas from the consensus params.
func (app *BaseApp) getMaximumBlockGas() (maxGas uint64) {
	if app.consensusParams == nil || app.consensusParams.BlockSize == nil {
		return 0
	}
	return uint64(app.consensusParams.BlockSize.MaxGas)
}

//______________________________________________________________________________

// ABCI

// Implements ABCI
func (app *BaseApp) Info(req abci.RequestInfo) abci.ResponseInfo {
	lastCommitID := app.cms.LastCommitID()

	return abci.ResponseInfo{
		Data:             app.name,
		LastBlockHeight:  lastCommitID.Version,
		LastBlockAppHash: lastCommitID.Hash,
	}
}

// Implements ABCI
func (app *BaseApp) SetOption(req abci.RequestSetOption) (res abci.ResponseSetOption) {
	// TODO: Implement
	return
}

// Implements ABCI
// InitChain runs the initialization logic directly on the CommitMultiStore.
func (app *BaseApp) InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain) {

	// Stash the consensus params in the cms main store and memoize.
	if req.ConsensusParams != nil {
		app.setConsensusParams(req.ConsensusParams)
		app.storeConsensusParams(req.ConsensusParams)
	}

	// Initialize the deliver state and check state with ChainID and run initChain
	app.setDeliverState(abci.Header{ChainID: req.ChainId})
	app.setCheckState(abci.Header{ChainID: req.ChainId})

	if app.initChainer == nil {
		return
	}

	// add block gas meter for any genesis transactions (allow infinite gas)
	app.deliverState.ctx = app.deliverState.ctx.
		WithBlockGasMeter(sdk.NewInfiniteGasMeter())

	res = app.initChainer(app.deliverState.ctx, req)

	// NOTE: we don't commit, but BeginBlock for block 1
	// starts from this deliverState
	return
}

// Filter peers by address / port
func (app *BaseApp) FilterPeerByAddrPort(info string) abci.ResponseQuery {
	if app.addrPeerFilter != nil {
		return app.addrPeerFilter(info)
	}
	return abci.ResponseQuery{}
}

// Filter peers by public key
func (app *BaseApp) FilterPeerByPubKey(info string) abci.ResponseQuery {
	if app.pubkeyPeerFilter != nil {
		return app.pubkeyPeerFilter(info)
	}
	return abci.ResponseQuery{}
}

// Splits a string path using the delimter '/'.  i.e. "this/is/funny" becomes []string{"this", "is", "funny"}
func splitPath(requestPath string) (path []string) {
	path = strings.Split(requestPath, "/")
	// first element is empty string
	if len(path) > 0 && path[0] == "" {
		path = path[1:]
	}
	return path
}

// Implements ABCI.
// Delegates to CommitMultiStore if it implements Queryable
func (app *BaseApp) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	path := splitPath(req.Path)
	if len(path) == 0 {
		msg := "no query path provided"
		return sdk.ErrUnknownRequest(msg).QueryResult()
	}
	switch path[0] {
	// "/app" prefix for special application queries
	case "app":
		return handleQueryApp(app, path, req)
	case "store":
		return handleQueryStore(app, path, req)
	case "p2p":
		return handleQueryP2P(app, path, req)
	case "custom":
		return handleQueryCustom(app, path, req)
	}

	msg := "unknown query path"
	return sdk.ErrUnknownRequest(msg).QueryResult()
}

func handleQueryApp(app *BaseApp, path []string, req abci.RequestQuery) (res abci.ResponseQuery) {
	if len(path) >= 2 {
		var result sdk.Result
		switch path[1] {
		case "simulate":
			txBytes := req.Data
			tx, err := app.txDecoder(txBytes)
			if err != nil {
				result = err.Result()
			} else {
				result = app.Simulate(tx)
			}
		case "version":
			return abci.ResponseQuery{
				Code:      uint32(sdk.CodeOK),
				Codespace: string(sdk.CodespaceRoot),
				Value:     []byte(version.GetVersion()),
			}
		default:
			result = sdk.ErrUnknownRequest(fmt.Sprintf("Unknown query: %s", path)).Result()
		}

		// Encode with json
		value := codec.Cdc.MustMarshalBinaryLengthPrefixed(result)
		return abci.ResponseQuery{
			Code:      uint32(sdk.CodeOK),
			Codespace: string(sdk.CodespaceRoot),
			Value:     value,
		}
	}
	msg := "Expected second parameter to be either simulate or version, neither was present"
	return sdk.ErrUnknownRequest(msg).QueryResult()
}

func handleQueryStore(app *BaseApp, path []string, req abci.RequestQuery) (res abci.ResponseQuery) {
	// "/store" prefix for store queries
	queryable, ok := app.cms.(sdk.Queryable)
	if !ok {
		msg := "multistore doesn't support queries"
		return sdk.ErrUnknownRequest(msg).QueryResult()
	}
	req.Path = "/" + strings.Join(path[1:], "/")
	return queryable.Query(req)
}

// nolint: unparam
func handleQueryP2P(app *BaseApp, path []string, req abci.RequestQuery) (res abci.ResponseQuery) {
	// "/p2p" prefix for p2p queries
	if len(path) >= 4 {
		if path[1] == "filter" {
			if path[2] == "addr" {
				return app.FilterPeerByAddrPort(path[3])
			}
			if path[2] == "pubkey" {
				// TODO: this should be changed to `id`
				// NOTE: this changed in tendermint and we didn't notice...
				return app.FilterPeerByPubKey(path[3])
			}
		} else {
			msg := "Expected second parameter to be filter"
			return sdk.ErrUnknownRequest(msg).QueryResult()
		}
	}

	msg := "Expected path is p2p filter <addr|pubkey> <parameter>"
	return sdk.ErrUnknownRequest(msg).QueryResult()
}

func handleQueryCustom(app *BaseApp, path []string, req abci.RequestQuery) (res abci.ResponseQuery) {
	// path[0] should be "custom" because "/custom" prefix is required for keeper queries.
	// the queryRouter routes using path[1]. For example, in the path "custom/gov/proposal", queryRouter routes using "gov"
	if len(path) < 2 || path[1] == "" {
		return sdk.ErrUnknownRequest("No route for custom query specified").QueryResult()
	}
	querier := app.queryRouter.Route(path[1])
	if querier == nil {
		return sdk.ErrUnknownRequest(fmt.Sprintf("no custom querier found for route %s", path[1])).QueryResult()
	}

	// Cache wrap the commit-multistore for safety.
	ctx := sdk.NewContext(app.cms.CacheMultiStore(), app.checkState.ctx.BlockHeader(), true, app.Logger).
		WithMinimumFees(app.minimumFees)

	// Passes the rest of the path as an argument to the querier.
	// For example, in the path "custom/gov/proposal/test", the gov querier gets []string{"proposal", "test"} as the path
	resBytes, err := querier(ctx, path[2:], req)
	if err != nil {
		return abci.ResponseQuery{
			Code:      uint32(err.Code()),
			Codespace: string(err.Codespace()),
			Log:       err.ABCILog(),
		}
	}
	return abci.ResponseQuery{
		Code:  uint32(sdk.CodeOK),
		Value: resBytes,
	}
}

// BeginBlock implements the ABCI application interface.
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	if app.cms.TracingEnabled() {
		app.cms.ResetTraceContext()
		app.cms.WithTracingContext(sdk.TraceContext(
			map[string]interface{}{"blockHeight": req.Header.Height},
		))
	}

	// Initialize the DeliverTx state. If this is the first block, it should
	// already be initialized in InitChain. Otherwise app.deliverState will be
	// nil, since it is reset on Commit.
	if app.deliverState == nil {
		app.setDeliverState(req.Header)
	} else {
		// In the first block, app.deliverState.ctx will already be initialized
		// by InitChain. Context is now updated with Header information.
		app.deliverState.ctx = app.deliverState.ctx.
			WithBlockHeader(req.Header).
			WithBlockHeight(req.Header.Height)
	}

	// add block gas meter
	var gasMeter sdk.GasMeter
	if maxGas := app.getMaximumBlockGas(); maxGas > 0 {
		gasMeter = sdk.NewGasMeter(maxGas)
	} else {
		gasMeter = sdk.NewInfiniteGasMeter()
	}
	app.deliverState.ctx = app.deliverState.ctx.WithBlockGasMeter(gasMeter)

	if app.beginBlocker != nil {
		res = app.beginBlocker(app.deliverState.ctx, req)
	}

	// set the signed validators for addition to context in deliverTx
	// TODO: communicate this result to the address to pubkey map in slashing
	app.voteInfos = req.LastCommitInfo.GetVotes()
	return
}

// CheckTx implements ABCI
// CheckTx runs the "basic checks" to see whether or not a transaction can possibly be executed,
// first decoding, then the ante handler (which checks signatures/fees/ValidateBasic),
// then finally the route match to see whether a handler exists. CheckTx does not run the actual
// Msg handler function(s).
func (app *BaseApp) CheckTx(txBytes []byte) (res abci.ResponseCheckTx) {
	// Decode the Tx.
	var result sdk.Result
	var tx, err = app.txDecoder(txBytes)
	if err != nil {
		result = err.Result()
	} else {
		result = app.runTx(runTxModeCheck, txBytes, tx)
	}

	return abci.ResponseCheckTx{
		Code:      uint32(result.Code),
		Data:      result.Data,
		Log:       result.Log,
		GasWanted: int64(result.GasWanted), // TODO: Should type accept unsigned ints?
		GasUsed:   int64(result.GasUsed),   // TODO: Should type accept unsigned ints?
		Tags:      result.Tags,
	}
}

// Implements ABCI
func (app *BaseApp) DeliverTx(txBytes []byte) (res abci.ResponseDeliverTx) {

	// Decode the Tx.
	var tx, err = app.txDecoder(txBytes)
	var result sdk.Result
	if err != nil {
		result = err.Result()
	} else {
		result = app.runTx(runTxModeDeliver, txBytes, tx)
	}

	// Even though the Result.Code is not OK, there are still effects,
	// namely fee deductions and sequence incrementing.

	// Tell the blockchain engine (i.e. Tendermint).
	return abci.ResponseDeliverTx{
		Code:      uint32(result.Code),
		Codespace: string(result.Codespace),
		Data:      result.Data,
		Log:       result.Log,
		GasWanted: int64(result.GasWanted), // TODO: Should type accept unsigned ints?
		GasUsed:   int64(result.GasUsed),   // TODO: Should type accept unsigned ints?
		Tags:      result.Tags,
	}
}

// Basic validator for msgs
func validateBasicTxMsgs(msgs []sdk.Msg) sdk.Error {
	if msgs == nil || len(msgs) == 0 {
		// TODO: probably shouldn't be ErrInternal. Maybe new ErrInvalidMessage, or ?
		return sdk.ErrInternal("Tx.GetMsgs() must return at least one message in list")
	}

	for _, msg := range msgs {
		// Validate the Msg.
		err := msg.ValidateBasic()
		if err != nil {
			return err
		}
	}

	return nil
}

// retrieve the context for the tx w/ txBytes and other memoized values.
func (app *BaseApp) getContextForTx(mode runTxMode, txBytes []byte) (ctx sdk.Context) {
	ctx = app.getState(mode).ctx.
		WithTxBytes(txBytes).
		WithVoteInfos(app.voteInfos).
		WithConsensusParams(app.consensusParams)
	if mode == runTxModeSimulate {
		ctx, _ = ctx.CacheContext()
	}
	return
}

// Iterates through msgs and executes them
func (app *BaseApp) runMsgs(ctx sdk.Context, msgs []sdk.Msg, mode runTxMode) (result sdk.Result) {
	// accumulate results
	logs := make([]string, 0, len(msgs))
	var data []byte   // NOTE: we just append them all (?!)
	var tags sdk.Tags // also just append them all
	var code sdk.CodeType
	var codespace sdk.CodespaceType
	for msgIdx, msg := range msgs {
		// Match route.
		msgRoute := msg.Route()
		handler := app.router.Route(msgRoute)
		if handler == nil {
			return sdk.ErrUnknownRequest("Unrecognized Msg type: " + msgRoute).Result()
		}

		var msgResult sdk.Result
		// Skip actual execution for CheckTx
		if mode != runTxModeCheck {
			msgResult = handler(ctx, msg)
		}

		// NOTE: GasWanted is determined by ante handler and
		// GasUsed by the GasMeter

		// Append Data and Tags
		data = append(data, msgResult.Data...)
		tags = append(tags, sdk.MakeTag(sdk.TagAction, []byte(msg.Type())))
		tags = append(tags, msgResult.Tags...)

		// Stop execution and return on first failed message.
		if !msgResult.IsOK() {
			logs = append(logs, fmt.Sprintf("Msg %d failed: %s", msgIdx, msgResult.Log))
			code = msgResult.Code
			codespace = msgResult.Codespace
			break
		}

		// Construct usable logs in multi-message transactions.
		logs = append(logs, fmt.Sprintf("Msg %d: %s", msgIdx, msgResult.Log))
	}

	// Set the final gas values.
	result = sdk.Result{
		Code:      code,
		Codespace: codespace,
		Data:      data,
		Log:       strings.Join(logs, "\n"),
		GasUsed:   ctx.GasMeter().GasConsumed(),
		// TODO: FeeAmount/FeeDenom
		Tags: tags,
	}

	return result
}

// Returns the applicantion's deliverState if app is in runTxModeDeliver,
// otherwise it returns the application's checkstate.
func (app *BaseApp) getState(mode runTxMode) *state {
	if mode == runTxModeCheck || mode == runTxModeSimulate {
		return app.checkState
	}

	return app.deliverState
}

// cacheTxContext returns a new context based off of the provided context with
// a cache wrapped multi-store.
func (app *BaseApp) cacheTxContext(ctx sdk.Context, txBytes []byte) (
	sdk.Context, sdk.CacheMultiStore) {

	ms := ctx.MultiStore()
	// TODO: https://github.com/cosmos/cosmos-sdk/issues/2824
	msCache := ms.CacheMultiStore()
	if msCache.TracingEnabled() {
		msCache = msCache.WithTracingContext(
			sdk.TraceContext(
				map[string]interface{}{
					"txHash": fmt.Sprintf("%X", tmhash.Sum(txBytes)),
				},
			),
		).(sdk.CacheMultiStore)
	}

	return ctx.WithMultiStore(msCache), msCache
}

// runTx processes a transaction. The transactions is proccessed via an
// anteHandler. The provided txBytes may be nil in some cases, eg. in tests. For
// further details on transaction execution, reference the BaseApp SDK
// documentation.
func (app *BaseApp) runTx(mode runTxMode, txBytes []byte, tx sdk.Tx) (result sdk.Result) {
	// NOTE: GasWanted should be returned by the AnteHandler. GasUsed is
	// determined by the GasMeter. We need access to the context to get the gas
	// meter so we initialize upfront.
	var gasWanted uint64

	ctx := app.getContextForTx(mode, txBytes)
	ms := ctx.MultiStore()

	// only run the tx if there is block gas remaining
	if mode == runTxModeDeliver && ctx.BlockGasMeter().IsOutOfGas() {
		result = sdk.ErrOutOfGas("no block gas left to run tx").Result()
		return
	}

	defer func() {
		if r := recover(); r != nil {
			switch rType := r.(type) {
			case sdk.ErrorOutOfGas:
				log := fmt.Sprintf("out of gas in location: %v", rType.Descriptor)
				result = sdk.ErrOutOfGas(log).Result()
			default:
				log := fmt.Sprintf("recovered: %v\nstack:\n%v", r, string(debug.Stack()))
				result = sdk.ErrInternal(log).Result()
			}
		}

		result.GasWanted = gasWanted
		result.GasUsed = ctx.GasMeter().GasConsumed()
	}()

	// If BlockGasMeter() panics it will be caught by the above recover and
	// return an error - in any case BlockGasMeter will consume gas past
	// the limit.
	// NOTE: this must exist in a separate defer function for the
	//       above recovery to recover from this one
	defer func() {
		if mode == runTxModeDeliver {
			ctx.BlockGasMeter().ConsumeGas(
				ctx.GasMeter().GasConsumedToLimit(), "block gas meter")
		}
	}()

	var msgs = tx.GetMsgs()
	if err := validateBasicTxMsgs(msgs); err != nil {
		return err.Result()
	}

	// Execute the ante handler if one is defined.
	if app.anteHandler != nil {
		var anteCtx sdk.Context
		var msCache sdk.CacheMultiStore

		// Cache wrap context before anteHandler call in case it aborts.
		// This is required for both CheckTx and DeliverTx.
		// https://github.com/cosmos/cosmos-sdk/issues/2772
		// NOTE: Alternatively, we could require that anteHandler ensures that
		// writes do not happen if aborted/failed.  This may have some
		// performance benefits, but it'll be more difficult to get right.
		anteCtx, msCache = app.cacheTxContext(ctx, txBytes)

		newCtx, result, abort := app.anteHandler(anteCtx, tx, (mode == runTxModeSimulate))
		if abort {
			return result
		}
		if !newCtx.IsZero() {
			// At this point, newCtx.MultiStore() is cache wrapped,
			// or something else replaced by anteHandler.
			// We want the original ms, not one which was cache-wrapped
			// for the ante handler.
			ctx = newCtx.WithMultiStore(ms)
		}
		msCache.Write()
		gasWanted = result.GasWanted
	}

	if mode == runTxModeCheck {
		return
	}

	// Create a new context based off of the existing context with a cache wrapped
	// multi-store in case message processing fails.
	runMsgCtx, msCache := app.cacheTxContext(ctx, txBytes)
	result = app.runMsgs(runMsgCtx, msgs, mode)
	result.GasWanted = gasWanted

	if mode == runTxModeSimulate {
		return
	}

	// only update state if all messages pass
	if result.IsOK() {
		msCache.Write()
	}

	return
}

// EndBlock implements the ABCI application interface.
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	if app.deliverState.ms.TracingEnabled() {
		app.deliverState.ms = app.deliverState.ms.ResetTraceContext().(sdk.CacheMultiStore)
	}

	if app.endBlocker != nil {
		res = app.endBlocker(app.deliverState.ctx, req)
	}

	return
}

// Implements ABCI
func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	header := app.deliverState.ctx.BlockHeader()

	// Write the Deliver state and commit the MultiStore
	app.deliverState.ms.Write()
	commitID := app.cms.Commit()
	// TODO: this is missing a module identifier and dumps byte array
	app.Logger.Debug("Commit synced",
		"commit", fmt.Sprintf("%X", commitID),
	)

	// Reset the Check state to the latest committed
	// NOTE: safe because Tendermint holds a lock on the mempool for Commit.
	// Use the header from this latest block.
	app.setCheckState(header)

	// Empty the Deliver state
	app.deliverState = nil

	return abci.ResponseCommit{
		Data: commitID.Hash,
	}
}
