package baseapp

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/pkg/errors"

	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
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

// The ABCI application
type BaseApp struct {
	// initialized on creation
	Logger     log.Logger
	Config     *cfg.Config
	name       string               // application name from abci.Info
	cdc        *wire.Codec          // Amino codec
	db         dbm.DB               // common DB backend
	cms        sdk.CommitMultiStore // Main (uncached) state
	router     Router               // handle any kind of message
	codespacer *sdk.Codespacer      // handle module codespacing

	// must be set
	txDecoder   sdk.TxDecoder   // unmarshal []byte into sdk.Tx
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
}

var _ abci.Application = (*BaseApp)(nil)

// Create and name new BaseApp
// NOTE: The db is used to store the version number for now.
func NewBaseApp(name string, cdc *wire.Codec, ctx *sdk.ServerContext, db dbm.DB) *BaseApp {
	app := &BaseApp{
		Logger:     ctx.Logger,
		Config:     ctx.Config,
		name:       name,
		cdc:        cdc,
		db:         db,
		cms:        store.NewCommitMultiStore(db),
		router:     NewRouter(),
		codespacer: sdk.NewCodespacer(),
		txDecoder:  defaultTxDecoder(cdc),
	}
	// Register the undefined & root codespaces, which should not be used by any modules
	app.codespacer.RegisterOrPanic(sdk.CodespaceRoot)
	return app
}

// BaseApp Name
func (app *BaseApp) Name() string {
	return app.name
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

// Set the txDecoder function
func (app *BaseApp) SetTxDecoder(txDecoder sdk.TxDecoder) {
	app.txDecoder = txDecoder
}

// default custom logic for transaction decoding
func defaultTxDecoder(cdc *wire.Codec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx = auth.StdTx{}

		if len(txBytes) == 0 {
			return nil, sdk.ErrTxDecode("txBytes are empty")
		}

		// StdTx.Msg is an interface. The concrete types
		// are registered by MakeTxCodec
		err := cdc.UnmarshalBinary(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxDecode("").TraceSDK(err.Error())
		}
		return tx, nil
	}
}

// nolint - Set functions
func (app *BaseApp) SetInitChainer(initChainer sdk.InitChainer) {
	app.initChainer = initChainer
}
func (app *BaseApp) SetBeginBlocker(beginBlocker sdk.BeginBlocker) {
	app.beginBlocker = beginBlocker
}
func (app *BaseApp) SetEndBlocker(endBlocker sdk.EndBlocker) {
	app.endBlocker = endBlocker
}
func (app *BaseApp) SetAnteHandler(ah sdk.AnteHandler) {
	app.anteHandler = ah
}
func (app *BaseApp) SetAddrPeerFilter(pf sdk.PeerFilter) {
	app.addrPeerFilter = pf
}
func (app *BaseApp) SetPubKeyPeerFilter(pf sdk.PeerFilter) {
	app.pubkeyPeerFilter = pf
}
func (app *BaseApp) Router() Router { return app.router }

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

	// XXX: Do we really need the header? What does it have that we want
	// here that's not already in the CommitID ? If an app wants to have it,
	// they can do so in their BeginBlocker. If we force it in baseapp,
	// then either we force the AppHash to change with every block (since the header
	// will be in the merkle store) or we can't write the state and the header to the
	// db atomically without doing some surgery on the store interfaces ...

	// if we've committed before, we expect <dbHeaderKey> to exist in the db
	/*
		var lastCommitID = app.cms.LastCommitID()
		var header abci.Header

		if !lastCommitID.IsZero() {
			headerBytes := app.db.Get(dbHeaderKey)
			if len(headerBytes) == 0 {
				errStr := fmt.Sprintf("Version > 0 but missing key %s", dbHeaderKey)
				return errors.New(errStr)
			}
			err := proto.Unmarshal(headerBytes, &header)
			if err != nil {
				return errors.Wrap(err, "failed to parse Header")
			}
			lastVersion := lastCommitID.Version
			if header.Height != lastVersion {
				errStr := fmt.Sprintf("expected db://%s.Height %v but got %v", dbHeaderKey, lastVersion, header.Height)
				return errors.New(errStr)
			}
		}
	*/

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
	app.initChainer(app.deliverState.ctx, req) // no error

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

// Implements ABCI.
// Delegates to CommitMultiStore if it implements Queryable
func (app *BaseApp) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	path := strings.Split(req.Path, "/")
	// first element is empty string
	if len(path) > 0 && path[0] == "" {
		path = path[1:]
	}
	// "/app" prefix for special application queries
	if len(path) >= 2 && path[0] == "app" {
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
		value := app.cdc.MustMarshalBinary(result)
		return abci.ResponseQuery{
			Code:  uint32(sdk.ABCICodeOK),
			Value: value,
		}
	}
	// "/store" prefix for store queries
	if len(path) >= 1 && path[0] == "store" {
		queryable, ok := app.cms.(sdk.Queryable)
		if !ok {
			msg := "multistore doesn't support queries"
			return sdk.ErrUnknownRequest(msg).QueryResult()
		}
		req.Path = "/" + strings.Join(path[1:], "/")
		return queryable.Query(req)
	}
	// "/p2p" prefix for p2p queries
	if len(path) >= 4 && path[0] == "p2p" {
		if path[1] == "filter" {
			if path[2] == "addr" {
				return app.FilterPeerByAddrPort(path[3])
			}
			if path[2] == "pubkey" {
				return app.FilterPeerByPubKey(path[3])
			}
		}
	}
	msg := "unknown query path"
	return sdk.ErrUnknownRequest(msg).QueryResult()
}

// Implements ABCI
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	// Initialize the DeliverTx state.
	// If this is the first block, it should already
	// be initialized in InitChain.
	// Otherwise app.deliverState will be nil, since it
	// is reset on Commit.
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

// Implements ABCI
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

// nolint - Mostly for testing
func (app *BaseApp) Check(tx sdk.Tx) (result sdk.Result) {
	return app.runTx(runTxModeCheck, nil, tx)
}

// nolint - full tx execution
func (app *BaseApp) Simulate(tx sdk.Tx) (result sdk.Result) {
	return app.runTx(runTxModeSimulate, nil, tx)
}

// nolint
func (app *BaseApp) Deliver(tx sdk.Tx) (result sdk.Result) {
	return app.runTx(runTxModeDeliver, nil, tx)
}

// txBytes may be nil in some cases, eg. in tests.
// Also, in the future we may support "internal" transactions.
func (app *BaseApp) runTx(mode runTxMode, txBytes []byte, tx sdk.Tx) (result sdk.Result) {
	// Handle any panics.
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case sdk.ErrorOutOfGas:
				log := fmt.Sprintf("out of gas in location: %v", r.(sdk.ErrorOutOfGas).Descriptor)
				result = sdk.ErrOutOfGas(log).Result()
			default:
				log := fmt.Sprintf("recovered: %v\nstack:\n%v", r, string(debug.Stack()))
				result = sdk.ErrInternal(log).Result()
			}
		}
	}()

	// Get the Msg.
	var msgs = tx.GetMsgs()
	if msgs == nil || len(msgs) == 0 {
		return sdk.ErrInternal("Tx.GetMsgs() must return at least one message in list").Result()
	}

	for _, msg := range msgs {
		// Validate the Msg
		err := msg.ValidateBasic()
		if err != nil {
			err = err.WithDefaultCodespace(sdk.CodespaceRoot)
			return err.Result()
		}
	}

	// Get the context
	var ctx sdk.Context
	if mode == runTxModeCheck || mode == runTxModeSimulate {
		ctx = app.checkState.ctx.WithTxBytes(txBytes)
	} else {
		ctx = app.deliverState.ctx.WithTxBytes(txBytes)
		ctx = ctx.WithSigningValidators(app.signedValidators)
	}

	// Simulate a DeliverTx for gas calculation
	if mode == runTxModeSimulate {
		ctx = ctx.WithIsCheckTx(false)
	}

	// Run the ante handler.
	if app.anteHandler != nil {
		newCtx, result, abort := app.anteHandler(ctx, tx)
		if abort {
			return result
		}
		if !newCtx.IsZero() {
			ctx = newCtx
		}
	}

	// Get the correct cache
	var msCache sdk.CacheMultiStore
	if mode == runTxModeCheck || mode == runTxModeSimulate {
		// CacheWrap app.checkState.ms in case it fails.
		msCache = app.checkState.CacheMultiStore()
		ctx = ctx.WithMultiStore(msCache)
	} else {
		// CacheWrap app.deliverState.ms in case it fails.
		msCache = app.deliverState.CacheMultiStore()
		ctx = ctx.WithMultiStore(msCache)
	}

	finalResult := sdk.Result{}
	var logs []string
	for i, msg := range msgs {
		// Match route.
		msgType := msg.Type()
		handler := app.router.Route(msgType)
		if handler == nil {
			return sdk.ErrUnknownRequest("Unrecognized Msg type: " + msgType).Result()
		}

		result = handler(ctx, msg)

		// Set gas utilized
		finalResult.GasUsed += ctx.GasMeter().GasConsumed()
		finalResult.GasWanted += result.GasWanted

		// Append Data and Tags
		finalResult.Data = append(finalResult.Data, result.Data...)
		finalResult.Tags = append(finalResult.Tags, result.Tags...)

		// Construct usable logs in multi-message transactions. Messages are 1-indexed in logs.
		logs = append(logs, fmt.Sprintf("Msg %d: %s", i+1, finalResult.Log))

		// Stop execution and return on first failed message.
		if !result.IsOK() {
			if len(msgs) == 1 {
				return result
			}
			result.GasUsed = finalResult.GasUsed
			if i == 0 {
				result.Log = fmt.Sprintf("Msg 1 failed: %s", result.Log)
			} else {
				result.Log = fmt.Sprintf("Msg 1-%d Passed. Msg %d failed: %s", i, i+1, result.Log)
			}
			return result
		}
	}

	// If not a simulated run and result was successful, write to app.checkState.ms or app.deliverState.ms
	// Only update state if all messages pass.
	if mode != runTxModeSimulate && result.IsOK() {
		msCache.Write()
	}

	finalResult.Log = strings.Join(logs, "\n")

	return finalResult
}

// Implements ABCI
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
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
