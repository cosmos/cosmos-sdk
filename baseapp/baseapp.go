package baseapp

import (
	"bytes"
	"fmt"
	"os"
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

// BaseApp - The ABCI application
type BaseApp struct {
	logger log.Logger

	// Application name from abci.Info
	name string

	// Common DB backend
	db dbm.DB

	// Main (uncached) state
	cms sdk.CommitMultiStore

	// Unmarshal []byte into sdk.Tx
	txDecoder sdk.TxDecoder

	// Ante handler for fee and auth.
	defaultAnteHandler sdk.AnteHandler

	// Handle any kind of message.
	router Router

	//--------------------
	// Volatile

	// CheckTx state, a cache-wrap of `.cms`.
	msCheck sdk.CacheMultiStore

	// DeliverTx state, a cache-wrap of `.cms`.
	msDeliver sdk.CacheMultiStore

	// Current block header
	header *abci.Header

	// Cached validator changes from DeliverTx.
	valUpdates []abci.Validator
}

var _ abci.Application = &BaseApp{}

func NewBaseApp(name string) *BaseApp {
	var baseapp = &BaseApp{
		logger: makeDefaultLogger(),
		name:   name,
		db:     nil,
		cms:    nil,
		router: NewRouter(),
	}
	baseapp.initDB()
	baseapp.initMultiStore()
	return baseapp
}

// Create the underlying leveldb datastore which will
// persist the Merkle tree inner & leaf nodes.
func (app *BaseApp) initDB() {
	db, err := dbm.NewGoLevelDB(app.name, "data")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	app.db = db
}

func (app *BaseApp) initMultiStore() {
	cms := store.NewCommitMultiStore(app.db)
	app.cms = cms
}

func (app *BaseApp) Name() string {
	return app.name
}

func (app *BaseApp) MountStore(key sdk.StoreKey, typ sdk.StoreType) {
	app.cms.MountStoreWithDB(key, typ, app.db)
}

func (app *BaseApp) SetTxDecoder(txDecoder sdk.TxDecoder) {
	app.txDecoder = txDecoder
}

func (app *BaseApp) SetDefaultAnteHandler(ah sdk.AnteHandler) {
	app.defaultAnteHandler = ah
}

func (app *BaseApp) Router() Router {
	return app.router
}

/* TODO consider:
func (app *BaseApp) SetBeginBlocker(...) {}
func (app *BaseApp) SetEndBlocker(...) {}
func (app *BaseApp) SetInitStater(...) {}
*/

func (app *BaseApp) LoadLatestVersion(mainKey sdk.StoreKey) error {
	app.cms.LoadLatestVersion()
	return app.initFromStore(mainKey)
}

func (app *BaseApp) LoadVersion(version int64, mainKey sdk.StoreKey) error {
	app.cms.LoadVersion(version)
	return app.initFromStore(mainKey)
}

// The last CommitID of the multistore.
func (app *BaseApp) LastCommitID() sdk.CommitID {
	return app.cms.LastCommitID()
}

// The last commited block height.
func (app *BaseApp) LastBlockHeight() int64 {
	return app.cms.LastCommitID().Version
}

// Initializes the remaining logic from app.cms.
func (app *BaseApp) initFromStore(mainKey sdk.StoreKey) error {
	var lastCommitID = app.cms.LastCommitID()
	var main = app.cms.GetKVStore(mainKey)
	var header *abci.Header

	// Main store should exist.
	if main == nil {
		return errors.New("BaseApp expects MultiStore with 'main' KVStore")
	}

	// If we've committed before, we expect main://<mainHeaderKey>.
	if !lastCommitID.IsZero() {
		headerBytes := main.Get(mainHeaderKey)
		if len(headerBytes) == 0 {
			errStr := fmt.Sprintf("Version > 0 but missing key %s", mainHeaderKey)
			return errors.New(errStr)
		}
		err := proto.Unmarshal(headerBytes, header)
		if err != nil {
			return errors.Wrap(err, "Failed to parse Header")
		}
		lastVersion := lastCommitID.Version
		if header.Height != lastVersion {
			errStr := fmt.Sprintf("Expected main://%s.Height %v but got %v", mainHeaderKey, lastVersion, header.Height)
			return errors.New(errStr)
		}
	}

	// Set BaseApp state.
	app.header = header
	app.msCheck = nil
	app.msDeliver = nil
	app.valUpdates = nil

	return nil
}

//----------------------------------------

// Implements ABCI.
func (app *BaseApp) Info(req abci.RequestInfo) abci.ResponseInfo {

	lastCommitID := app.cms.LastCommitID()

	return abci.ResponseInfo{
		Data:             app.name,
		LastBlockHeight:  lastCommitID.Version,
		LastBlockAppHash: lastCommitID.Hash,
	}
}

// Implements ABCI.
func (app *BaseApp) SetOption(req abci.RequestSetOption) (res abci.ResponseSetOption) {
	// TODO: Implement
	return
}

// Implements ABCI.
func (app *BaseApp) InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain) {
	// TODO: Use req.Validators
	return
}

// Implements ABCI.
func (app *BaseApp) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	// TODO: See app/query.go
	return
}

// Implements ABCI.
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	// NOTE: For consistency we should unset these upon EndBlock.
	app.header = &req.Header
	app.msDeliver = app.cms.CacheMultiStore()
	app.msCheck = app.cms.CacheMultiStore()
	app.valUpdates = nil
	return
}

// Implements ABCI.
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

// Implements ABCI.
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

// txBytes may be nil in some cases, for example, when tx is
// coming from TestApp.  Also, in the future we may support
// "internal" transactions.
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

	// Construct a Context.
	var ctx = app.newContext(isCheckTx, txBytes)

	// TODO: override default ante handler w/ custom ante handler.

	// Run the ante handler.
	newCtx, result, abort := app.defaultAnteHandler(ctx, tx)
	if isCheckTx || abort {
		return result
	}
	if !newCtx.IsZero() {
		ctx = newCtx
	}

	// CacheWrap app.msDeliver in case it fails.
	msCache := app.getMultiStore(isCheckTx).CacheMultiStore()
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

// Implements ABCI.
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	res.ValidatorUpdates = app.valUpdates
	app.valUpdates = nil
	app.header = nil
	app.msDeliver = nil
	app.msCheck = nil
	return
}

// Implements ABCI.
func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	app.msDeliver.Write()
	commitID := app.cms.Commit()
	app.logger.Debug("Commit synced",
		"commit", commitID,
	)
	return abci.ResponseCommit{
		Data: commitID.Hash,
	}
}

//----------------------------------------
// Misc.

func (app *BaseApp) getMultiStore(isCheckTx bool) sdk.MultiStore {
	if isCheckTx {
		return app.msCheck
	} else {
		return app.msDeliver
	}
}

// Return index of list with validator of same PubKey, or -1 if no match
func pubKeyIndex(val *abci.Validator, list []*abci.Validator) int {
	for i, v := range list {
		if bytes.Equal(val.PubKey, v.PubKey) {
			return i
		}
	}
	return -1
}

// Make a simple default logger
// TODO: Make log capturable for each transaction, and return it in
// ResponseDeliverTx.Log and ResponseCheckTx.Log.
func makeDefaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
}
