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

const mainKeyHeader = "header"

// App - The ABCI application
type App struct {
	logger log.Logger

	// App name from abci.Info
	name string

	// DeliverTx (main) state
	store types.MultiStore

	// CheckTx state
	storeCheck types.CacheMultiStore

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

func (app *App) SetStore(store types.MultiStore) {
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
	storeCheck := store.CacheMultiStore()

	// Main store should exist.
	if store.GetKVStore("main") == nil {
		return errors.New("App expects MultiStore with 'main' KVStore")
	}

	// If we've committed before, we expect main://<mainKeyHeader>.
	if !lastCommitID.IsZero() {
		headerBytes, ok := main.Get(mainKeyHeader)
		if !ok {
			errStr := fmt.Sprintf("Version > 0 but missing key %s", mainKeyHeader)
			return errors.New(errStr)
		}
		err = proto.Unmarshal(headerBytes, header)
		if err != nil {
			return errors.Wrap(err, "Failed to parse Header")
		}
		if header.Height != lastCommitID.Version {
			errStr := fmt.Sprintf("Expected main://%s.Height %v but got %v", mainKeyHeader, version, headerHeight)
			return errors.New(errStr)
		}
	}

	// Set App state.
	app.header = header
	app.storeCheck = app.store.CacheMultiStore()
	app.valUpdates = nil

	return nil
}

//----------------------------------------

// Implements ABCI
func (app *App) Info(req abci.RequestInfo) abci.ResponseInfo {

	lastCommitID := app.store.LastCommitID()

	return abci.ResponseInfo{
		Data:             app.Name,
		LastBlockHeight:  lastCommitID.Version,
		LastBlockAppHash: lastCommitID.Hash,
	}
}

// Implements ABCI
func (app *App) SetOption(req abci.RequestSetOption) abci.ResponseSetOption {
	return "Not Implemented"
}

// Implements ABCI
func (app *App) InitChain(req abci.RequestInitChain) abci.ResponseInitChain {
	// TODO: Use req.Validators
}

// Implements ABCI
func (app *App) Query(req abci.RequestQuery) abci.ResponseQuery {
	// TODO: See app/query.go
}

// Implements ABCI
func (app *App) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	app.header = req.Header
}

// Implements ABCI
func (app *App) CheckTx(txBytes []byte) abci.ResponseCheckTx {

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
func (app *App) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {

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
	// XXX Update to res.Updates.
	res.ValidatorUpdates = app.valUpdates
	app.valUpdates = nil
	return
}

// Implements ABCI
func (app *App) Commit() (res abci.ResponseCommit) {
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
