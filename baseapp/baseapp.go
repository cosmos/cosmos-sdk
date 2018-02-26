package baseapp

import (
	"fmt"
	"runtime/debug"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var mainHeaderKey = []byte("header")

// The ABCI application
type BaseApp struct {
	// initialized on creation
	logger log.Logger
	name   string               // application name from abci.Info
	db     dbm.DB               // common DB backend
	cms    sdk.CommitMultiStore // Main (uncached) state
	router Router               // handle any kind of message

	// must be set
	txDecoder   sdk.TxDecoder   // unmarshal []byte into sdk.Tx
	anteHandler sdk.AnteHandler // ante handler for fee and auth

	// may be nil
	initChainer  sdk.InitChainer  // initialize state with validators and state blob
	beginBlocker sdk.BeginBlocker // logic to run before any txs
	endBlocker   sdk.EndBlocker   // logic to run after all txs, and to determine valset changes

	//--------------------
	// Volatile
	// .msCheck and .ctxCheck are set on initialization and reset on Commit.
	// .msDeliver and .ctxDeliver are (re-)set on BeginBlock.
	// .valUpdates accumulate in DeliverTx and reset in BeginBlock.
	// QUESTION: should we put valUpdates in the ctxDeliver?

	msCheck    sdk.CacheMultiStore // CheckTx state, a cache-wrap of `.cms`
	msDeliver  sdk.CacheMultiStore // DeliverTx state, a cache-wrap of `.cms`
	ctxCheck   sdk.Context         // CheckTx context
	ctxDeliver sdk.Context         // DeliverTx context
	valUpdates []abci.Validator    // cached validator changes from DeliverTx
}

var _ abci.Application = &BaseApp{}

// Create and name new BaseApp
func NewBaseApp(name string, logger log.Logger, db dbm.DB, txDecoder sdk.TxDecoder,
	ah sdk.AnteHandler) *BaseApp {

	return &BaseApp{
		logger:      logger,
		name:        name,
		db:          db,
		cms:         store.NewCommitMultiStore(db),
		router:      NewRouter(),
		txDecoder:   txDecoder,
		anteHandler: ah,
	}
}

// BaseApp Name
func (app *BaseApp) Name() string {
	return app.name
}

// Mount a store to the provided key in the BaseApp multistore
func (app *BaseApp) MountStoresIAVL(keys ...*sdk.KVStoreKey) {
	for _, key := range keys {
		app.MountStore(key, sdk.StoreTypeIAVL)
	}
}

// Mount a store to the provided key in the BaseApp multistore
func (app *BaseApp) MountStore(key sdk.StoreKey, typ sdk.StoreType) {
	app.cms.MountStoreWithDB(key, typ, app.db)
}

func (app *BaseApp) SetInitChainer(initChainer sdk.InitChainer) {
	app.initChainer = initChainer
}
func (app *BaseApp) SetBeginBlocker(beginBlocker sdk.BeginBlocker) {
	app.beginBlocker = beginBlocker
}
func (app *BaseApp) SetEndBlocker(endBlocker sdk.EndBlocker) {
	app.endBlocker = endBlocker
}

// nolint - Get functions
func (app *BaseApp) Router() Router { return app.router }

// load latest application version
func (app *BaseApp) LoadLatestVersion(mainKey sdk.StoreKey) error {
	app.cms.LoadLatestVersion()
	return app.initFromStore(mainKey)
}

// load application version
func (app *BaseApp) LoadVersion(version int64, mainKey sdk.StoreKey) error {
	app.cms.LoadVersion(version)
	return app.initFromStore(mainKey)
}

// the last CommitID of the multistore
func (app *BaseApp) LastCommitID() sdk.CommitID {
	return app.cms.LastCommitID()
}

// the last commited block height
func (app *BaseApp) LastBlockHeight() int64 {
	return app.cms.LastCommitID().Version
}

// initializes the remaining logic from app.cms
func (app *BaseApp) initFromStore(mainKey sdk.StoreKey) error {
	var lastCommitID = app.cms.LastCommitID()
	var main = app.cms.GetKVStore(mainKey)
	var header abci.Header

	// main store should exist.
	if main == nil {
		return errors.New("BaseApp expects MultiStore with 'main' KVStore")
	}

	// if we've committed before, we expect main://<mainHeaderKey>
	if !lastCommitID.IsZero() {
		headerBytes := main.Get(mainHeaderKey)
		if len(headerBytes) == 0 {
			errStr := fmt.Sprintf("Version > 0 but missing key %s", mainHeaderKey)
			return errors.New(errStr)
		}
		err := proto.Unmarshal(headerBytes, &header)
		if err != nil {
			return errors.Wrap(err, "Failed to parse Header")
		}
		lastVersion := lastCommitID.Version
		if header.Height != lastVersion {
			errStr := fmt.Sprintf("Expected main://%s.Height %v but got %v", mainHeaderKey, lastVersion, header.Height)
			return errors.New(errStr)
		}
	}

	// initialize Check state
	app.msCheck = app.cms.CacheMultiStore()
	app.ctxCheck = app.NewContext(true, abci.Header{})

	return nil
}

// NewContext returns a new Context with the correct store, the given header, and nil txBytes.
func (app *BaseApp) NewContext(isCheckTx bool, header abci.Header) sdk.Context {
	if isCheckTx {
		return sdk.NewContext(app.msCheck, header, true, nil)
	}
	return sdk.NewContext(app.msDeliver, header, false, nil)
}

//----------------------------------------
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
	if app.initChainer == nil {
		// TODO: should we have some default handling of validators?
		return
	}

	// make a context for the initialization.
	// NOTE: we're writing to the cms directly, without a CacheWrap
	ctx := sdk.NewContext(app.cms, abci.Header{}, false, nil)

	res = app.initChainer(ctx, req)
	// TODO: handle error https://github.com/cosmos/cosmos-sdk/issues/468

	// XXX this commits everything and bumps the version.
	// https://github.com/cosmos/cosmos-sdk/issues/442#issuecomment-366470148
	app.cms.Commit()

	return
}

// Implements ABCI.
// Delegates to CommitMultiStore if it implements Queryable
func (app *BaseApp) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	queryable, ok := app.cms.(sdk.Queryable)
	if !ok {
		msg := "application doesn't support queries"
		return sdk.ErrUnknownRequest(msg).Result().ToQuery()
	}
	return queryable.Query(req)
}

// Implements ABCI
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	app.msDeliver = app.cms.CacheMultiStore()
	app.ctxDeliver = app.NewContext(false, req.Header)
	app.valUpdates = nil
	if app.beginBlocker != nil {
		res = app.beginBlocker(app.ctxDeliver, req)
	}
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
		result = app.runTx(true, txBytes, tx)
	}

	return abci.ResponseCheckTx{
		Code:      uint32(result.Code),
		Data:      result.Data,
		Log:       result.Log,
		GasWanted: result.GasWanted,
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
		result = app.runTx(false, txBytes, tx)
	}

	// After-handler hooks.
	if result.IsOK() {
		app.valUpdates = append(app.valUpdates, result.ValidatorUpdates...)
	} else {
		// Even though the Code is not OK, there will be some side
		// effects, like those caused by fee deductions or sequence
		// incrementations.
	}

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

// Mostly for testing
func (app *BaseApp) Check(tx sdk.Tx) (result sdk.Result) {
	return app.runTx(true, nil, tx)
}
func (app *BaseApp) Deliver(tx sdk.Tx) (result sdk.Result) {
	return app.runTx(false, nil, tx)
}

// txBytes may be nil in some cases, eg. in tests.
// Also, in the future we may support "internal" transactions.
func (app *BaseApp) runTx(isCheckTx bool, txBytes []byte, tx sdk.Tx) (result sdk.Result) {

	// Handle any panics.
	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("Recovered: %v\nstack:\n%v", r, string(debug.Stack()))
			result = sdk.ErrInternal(log).Result()
		}
	}()

	// Get the Msg.
	var msg = tx.GetMsg()
	if msg == nil {
		return sdk.ErrInternal("Tx.GetMsg() returned nil").Result()
	}

	// Validate the Msg.
	err := msg.ValidateBasic()
	if err != nil {
		return err.Result()
	}

	// Get the context
	var ctx sdk.Context
	if isCheckTx {
		ctx = app.ctxCheck.WithTxBytes(txBytes)
	} else {
		ctx = app.ctxDeliver.WithTxBytes(txBytes)
	}

	// Run the ante handler.
	newCtx, result, abort := app.anteHandler(ctx, tx)
	if isCheckTx || abort {
		return result
	}
	if !newCtx.IsZero() {
		ctx = newCtx
	}

	// CacheWrap app.msDeliver in case it fails.
	msCache := app.msDeliver.CacheMultiStore()
	ctx = ctx.WithMultiStore(msCache)

	// Match and run route.
	msgType := msg.Type()
	handler := app.router.Route(msgType)
	result = handler(ctx, msg)

	// If result was successful, write to app.msDeliver or app.msCheck.
	if result.IsOK() {
		msCache.Write()
	}

	return result
}

// Implements ABCI
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	if app.endBlocker != nil {
		res = app.endBlocker(app.ctxDeliver, req)
	} else {
		res.ValidatorUpdates = app.valUpdates
	}
	return
}

// Implements ABCI
func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	// Write the Deliver state and commit the MultiStore
	app.msDeliver.Write()
	commitID := app.cms.Commit()
	app.logger.Debug("Commit synced",
		"commit", commitID,
	)

	// Reset the Check state
	// NOTE: safe because Tendermint holds a lock on the mempool for Commit.
	// Use the header from this latest block.
	header := app.ctxDeliver.BlockHeader()
	app.msCheck = app.cms.CacheMultiStore()
	app.ctxCheck = app.NewContext(true, header)

	return abci.ResponseCommit{
		Data: commitID.Hash,
	}
}
