package baseapp

import (
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"github.com/pkg/errors"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/wire"
)

// Key to store the header in the DB itself.
// Use the db directly instead of a store to avoid
// conflicts with handlers writing to the store
// and to avoid affecting the Merkle root.
var dbHeaderKey = []byte("header")

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
	Logger     log.Logger
	name       string               // application name from abci.Info
	db         dbm.DB               // common DB backend
	cms        sdk.CommitMultiStore // Main (uncached) state
	router     Router               // handle any kind of message
	codespacer *sdk.Codespacer      // handle module codespacing
	txDecoder  sdk.TxDecoder        // unmarshal []byte into sdk.Tx

	anteHandler sdk.AnteHandler // ante handler for fee and auth

	// may be nil
	initChainer      sdk.InitChainer  // initialize state with validators and state blob
	beginBlocker     sdk.BeginBlocker // logic to run before any txs
	endBlocker       sdk.EndBlocker   // logic to run after all txs, and to determine valset changes
	addrPeerFilter   sdk.PeerFilter   // filter peers by address and port
	pubkeyPeerFilter sdk.PeerFilter   // filter peers by public key

	//--------------------
	// Volatile
	// checkState is set on initialization and reset on Commit.
	// deliverState is set in InitChain and BeginBlock and cleared on Commit.
	// See methods setCheckState and setDeliverState.
	checkState       *state                  // for CheckTx
	deliverState     *state                  // for DeliverTx
	signedValidators []abci.SigningValidator // absent validators from begin block

	// flag for sealing
	sealed bool
}

var _ abci.Application = (*BaseApp)(nil)

// NewBaseApp returns a reference to an initialized BaseApp.
//
// TODO: Determine how to use a flexible and robust configuration paradigm that
// allows for sensible defaults while being highly configurable
// (e.g. functional options).
//
// NOTE: The db is used to store the version number for now.
// Accepts a user-defined txDecoder
// Accepts variable number of option functions, which act on the BaseApp to set configuration choices
func NewBaseApp(name string, logger log.Logger, db dbm.DB, txDecoder sdk.TxDecoder, options ...func(*BaseApp)) *BaseApp {
	app := &BaseApp{
		Logger:     logger,
		name:       name,
		db:         db,
		cms:        store.NewCommitMultiStore(db),
		router:     NewRouter(),
		codespacer: sdk.NewCodespacer(),
		txDecoder:  txDecoder,
	}

	// Register the undefined & root codespaces, which should not be used by
	// any modules.
	app.codespacer.RegisterOrPanic(sdk.CodespaceRoot)
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

// Register the next available codespace through the baseapp's codespacer, starting from a default
func (app *BaseApp) RegisterCodespace(codespace sdk.CodespaceType) sdk.CodespaceType {
	return app.codespacer.RegisterNext(codespace)
}

// Mount a store to the provided key in the BaseApp multistore
func (app *BaseApp) MountStoresIAVL(keys ...*sdk.KVStoreKey) {
	for _, key := range keys {
		app.MountStore(key, sdk.StoreTypeIAVL)
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
func (app *BaseApp) LoadLatestVersion(mainKey sdk.StoreKey) error {
	err := app.cms.LoadLatestVersion()
	if err != nil {
		return err
	}
	return app.initFromStore(mainKey)
}

// load application version
func (app *BaseApp) LoadVersion(version int64, mainKey sdk.StoreKey) error {
	err := app.cms.LoadVersion(version)
	if err != nil {
		return err
	}
	return app.initFromStore(mainKey)
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
func (app *BaseApp) initFromStore(mainKey sdk.StoreKey) error {
	// main store should exist.
	// TODO: we don't actually need the main store here
	main := app.cms.GetKVStore(mainKey)
	if main == nil {
		return errors.New("baseapp expects MultiStore with 'main' KVStore")
	}
	// Needed for `gaiad export`, which inits from store but never calls initchain
	app.setCheckState(abci.Header{})

	app.Seal()

	return nil
}

// NewContext returns a new Context with the correct store, the given header, and nil txBytes.
func (app *BaseApp) NewContext(isCheckTx bool, header abci.Header) sdk.Context {
	if isCheckTx {
		return sdk.NewContext(app.checkState.ms, header, true, app.Logger)
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

func (app *BaseApp) setCheckState(header abci.Header) {
	ms := app.cms.CacheMultiStore()
	app.checkState = &state{
		ms:  ms,
		ctx: sdk.NewContext(ms, header, true, app.Logger),
	}
}

func (app *BaseApp) setDeliverState(header abci.Header) {
	ms := app.cms.CacheMultiStore()
	app.deliverState = &state{
		ms:  ms,
		ctx: sdk.NewContext(ms, header, false, app.Logger),
	}
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
// InitChain runs the initialization logic directly on the CommitMultiStore and commits it.
func (app *BaseApp) InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain) {
	// Initialize the deliver state and check state with ChainID and run initChain
	app.setDeliverState(abci.Header{ChainID: req.ChainId})
	app.setCheckState(abci.Header{ChainID: req.ChainId})

	if app.initChainer == nil {
		return
	}
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
				Code:  uint32(sdk.ABCICodeOK),
				Value: []byte(version.GetVersion()),
			}
		default:
			result = sdk.ErrUnknownRequest(fmt.Sprintf("Unknown query: %s", path)).Result()
		}

		// Encode with json
		value := wire.Cdc.MustMarshalBinary(result)
		return abci.ResponseQuery{
			Code:  uint32(sdk.ABCICodeOK),
			Value: value,
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
		app.deliverState.ctx = app.deliverState.ctx.WithBlockHeader(req.Header)
	}

	if app.beginBlocker != nil {
		res = app.beginBlocker(app.deliverState.ctx, req)
	}

	// set the signed validators for addition to context in deliverTx
	app.signedValidators = req.Validators
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
		GasWanted: result.GasWanted,
		GasUsed:   result.GasUsed,
		Fee: cmn.KI64Pair{
			[]byte(result.FeeDenom),
			result.FeeAmount,
		},
		Tags: result.Tags,
	}
}

// Implements ABCI
func (app *BaseApp) DeliverTx(txBytes []byte) (res abci.ResponseDeliverTx) {
	// Decode the Tx.
	var result sdk.Result
	var tx, err = app.txDecoder(txBytes)
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
		Data:      result.Data,
		Log:       result.Log,
		GasWanted: result.GasWanted,
		GasUsed:   result.GasUsed,
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
			err = err.WithDefaultCodespace(sdk.CodespaceRoot)
			return err
		}
	}

	return nil
}

func (app *BaseApp) getContextForAnte(mode runTxMode, txBytes []byte) (ctx sdk.Context) {
	// Get the context
	if mode == runTxModeCheck || mode == runTxModeSimulate {
		ctx = app.checkState.ctx.WithTxBytes(txBytes)
	} else {
		ctx = app.deliverState.ctx.WithTxBytes(txBytes)
		ctx = ctx.WithSigningValidators(app.signedValidators)
	}

	return
}

// Iterates through msgs and executes them
func (app *BaseApp) runMsgs(ctx sdk.Context, msgs []sdk.Msg, mode runTxMode) (result sdk.Result) {
	// accumulate results
	logs := make([]string, 0, len(msgs))
	var data []byte   // NOTE: we just append them all (?!)
	var tags sdk.Tags // also just append them all
	var code sdk.ABCICodeType
	for msgIdx, msg := range msgs {
		// Match route.
		msgType := msg.Type()
		handler := app.router.Route(msgType)
		if handler == nil {
			return sdk.ErrUnknownRequest("Unrecognized Msg type: " + msgType).Result()
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
		tags = append(tags, msgResult.Tags...)

		// Stop execution and return on first failed message.
		if !msgResult.IsOK() {
			logs = append(logs, fmt.Sprintf("Msg %d failed: %s", msgIdx, msgResult.Log))
			code = msgResult.Code
			break
		}

		// Construct usable logs in multi-message transactions.
		logs = append(logs, fmt.Sprintf("Msg %d: %s", msgIdx, msgResult.Log))
	}

	// Set the final gas values.
	result = sdk.Result{
		Code:    code,
		Data:    data,
		Log:     strings.Join(logs, "\n"),
		GasUsed: ctx.GasMeter().GasConsumed(),
		// TODO: FeeAmount/FeeDenom
		Tags: tags,
	}

	return result
}

// Returns the applicantion's deliverState if app is in runTxModeDeliver,
// otherwise it returns the application's checkstate.
func getState(app *BaseApp, mode runTxMode) *state {
	if mode == runTxModeCheck || mode == runTxModeSimulate {
		return app.checkState
	}

	return app.deliverState
}

// runTx processes a transaction. The transactions is proccessed via an
// anteHandler. txBytes may be nil in some cases, eg. in tests. Also, in the
// future we may support "internal" transactions.
func (app *BaseApp) runTx(mode runTxMode, txBytes []byte, tx sdk.Tx) (result sdk.Result) {
	// NOTE: GasWanted should be returned by the AnteHandler. GasUsed is
	// determined by the GasMeter. We need access to the context to get the gas
	// meter so we initialize upfront.
	var gasWanted int64
	ctx := app.getContextForAnte(mode, txBytes)

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

	var msgs = tx.GetMsgs()

	err := validateBasicTxMsgs(msgs)
	if err != nil {
		return err.Result()
	}

	// run the ante handler
	if app.anteHandler != nil {
		newCtx, result, abort := app.anteHandler(ctx, tx)
		if abort {
			return result
		}
		if !newCtx.IsZero() {
			ctx = newCtx
		}

		gasWanted = result.GasWanted
	}

	// Keep the state in a transient CacheWrap in case processing the messages
	// fails.
	msCache := getState(app, mode).CacheMultiStore()
	if msCache.TracingEnabled() {
		msCache = msCache.WithTracingContext(sdk.TraceContext(
			map[string]interface{}{"txHash": cmn.HexBytes(tmhash.Sum(txBytes)).String()},
		)).(sdk.CacheMultiStore)
	}

	ctx = ctx.WithMultiStore(msCache)
	result = app.runMsgs(ctx, msgs, mode)
	result.GasWanted = gasWanted

	// only update state if all messages pass and we're not in a simulation
	if result.IsOK() && mode != runTxModeSimulate {
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
	/*
		// Write the latest Header to the store
			headerBytes, err := proto.Marshal(&header)
			if err != nil {
				panic(err)
			}
			app.db.SetSync(dbHeaderKey, headerBytes)
	*/

	// Write the Deliver state and commit the MultiStore
	app.deliverState.ms.Write()
	commitID := app.cms.Commit()
	// TODO: this is missing a module identifier and dumps byte array
	app.Logger.Debug("Commit synced",
		"commit", commitID,
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
