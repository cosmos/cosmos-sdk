package app

import (
	"bytes"
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/types"
)

var mainHeaderKey = "header"

// App - The ABCI application
type App struct {
	logger log.Logger

	// App name from abci.Info
	name string

	// Main (uncached) state
	store types.CommitMultiStore

	// CheckTx state, a cache-wrap of store.
	storeCheck types.CacheMultiStore

	// DeliverTx state, a cache-wrap of store.
	storeDeliver types.CacheMultiStore

	// Current block header
	header *abci.Header

	// Handler for CheckTx and DeliverTx.
	handler types.Handler

	// Cached validator changes from DeliverTx
	valUpdates []abci.Validator
}

var _ abci.Application = &App{}

func NewApp(name string) *App {
	return &App{
		name:   name,
		logger: makeDefaultLogger(),
	}
}

func (app *App) Name() string {
	return app.name
}

func (app *App) SetCommitStore(store types.CommitMultiStore) {
	app.store = store
}

func (app *App) SetHandler(handler types.Handler) {
	app.handler = handler
}

func (app *App) LoadLatestVersion() error {
	store := app.store
	store.LoadLatestVersion()
	return app.initFromStore()
}

func (app *App) LoadVersion(version int64) error {
	store := app.store
	store.LoadVersion(version)
	return app.initFromStore()
}

// Initializes the remaining logic from app.store.
func (app *App) initFromStore() error {
	store := app.store
	lastCommitID := store.LastCommitID()
	main := store.GetKVStore("main")
	header := (*abci.Header)(nil)

	// Main store should exist.
	if store.GetKVStore("main") == nil {
		return errors.New("App expects MultiStore with 'main' KVStore")
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

	// Set App state.
	app.header = header
	app.storeCheck = nil
	app.storeDeliver = nil
	app.valUpdates = nil

	return nil
}

//----------------------------------------

// Implements ABCI
func (app *App) Info(req abci.RequestInfo) abci.ResponseInfo {

	lastCommitID := app.store.LastCommitID()

	return abci.ResponseInfo{
		Data:             app.name,
		LastBlockHeight:  lastCommitID.Version,
		LastBlockAppHash: lastCommitID.Hash,
	}
}

// Implements ABCI
func (app *App) SetOption(req abci.RequestSetOption) (res abci.ResponseSetOption) {
	// TODO: Implement
	return
}

// Implements ABCI
func (app *App) InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain) {
	// TODO: Use req.Validators
	return
}

// Implements ABCI
func (app *App) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	// TODO: See app/query.go
	return
}

// Implements ABCI
func (app *App) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	app.header = req.Header
	app.storeDeliver = app.store.CacheMultiStore()
	app.storeCheck = app.store.CacheMultiStore()
	return
}

// Implements ABCI
func (app *App) CheckTx(txBytes []byte) (res abci.ResponseCheckTx) {

	// Initialize arguments to Handler.
	var isCheckTx = true
	var ctx = types.NewContext(app.header, isCheckTx, txBytes)
	var store = app.store
	var tx Tx = nil // nil until a decorator parses one.

	// Run the handler.
	var result = app.handler(ctx, app.store, tx)

	// Tell the blockchain engine (i.e. Tendermint).
	return abci.ResponseDeliverTx{
		Code:      result.Code,
		Data:      result.Data,
		Log:       result.Log,
		Gas:       result.Gas,
		FeeDenom:  result.FeeDenom,
		FeeAmount: result.FeeAmount,
	}
}

// Implements ABCI
func (app *App) DeliverTx(txBytes []byte) (res abci.ResponseDeliverTx) {

	// Initialize arguments to Handler.
	var isCheckTx = false
	var ctx = types.NewContext(app.header, isCheckTx, txBytes)
	var store = app.store
	var tx Tx = nil // nil until a decorator parses one.

	// Run the handler.
	var result = app.handler(ctx, app.store, tx)

	// After-handler hooks.
	if result.Code == abci.CodeType_OK {
		app.valUpdates = append(app.valUpdates, result.ValUpdate)
	} else {
		// Even though the Code is not OK, there will be some side effects,
		// like those caused by fee deductions or sequence incrementations.
	}

	// Tell the blockchain engine (i.e. Tendermint).
	return abci.ResponseDeliverTx{
		Code: result.Code,
		Data: result.Data,
		Log:  result.Log,
		Tags: result.Tags,
	}
}

// Implements ABCI
func (app *App) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	res.ValidatorUpdates = app.valUpdates
	app.valUpdates = nil
	return
}

// Implements ABCI
func (app *App) Commit() (res abci.ResponseCommit) {
	app.storeDeliver.Write()
	commitID := app.store.Commit()
	app.logger.Debug("Commit synced",
		"commit", commitID,
	)
	return abci.NewResultOK(hash, "")
}

//----------------------------------------
// Misc.

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
