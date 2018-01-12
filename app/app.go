package app

import (
	"bytes"
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var mainHeaderKey = []byte("header")

// App - The ABCI application
type App struct {
	logger log.Logger

	// App name from abci.Info
	name string

	// Main (uncached) state
	ms sdk.CommitMultiStore

	// Unmarshal []byte into sdk.Tx
	txDecoder sdk.TxDecoder

	// Ante handler for fee and auth.
	defaultAnteHandler sdk.AnteHandler

	// Handle any kind of message.
	router Router

	//--------------------
	// Volatile

	// CheckTx state, a cache-wrap of `.ms`.
	msCheck sdk.CacheMultiStore

	// DeliverTx state, a cache-wrap of `.ms`.
	msDeliver sdk.CacheMultiStore

	// Current block header
	header abci.Header

	// Cached validator changes from DeliverTx.
	valUpdates []abci.Validator
}

var _ abci.Application = &App{}

func NewApp(name string, ms sdk.CommitMultiStore) *App {
	return &App{
		logger: makeDefaultLogger(),
		name:   name,
		ms:     ms,
		router: NewRouter(),
	}
}

func (app *App) Name() string {
	return app.name
}

func (app *App) SetTxDecoder(txDecoder sdk.TxDecoder) {
	app.txDecoder = txDecoder
}

func (app *App) SetDefaultAnteHandler(ah sdk.AnteHandler) {
	app.defaultAnteHandler = ah
}

func (app *App) Router() Router {
	return app.router
}

/* TODO consider:
func (app *App) SetBeginBlocker(...) {}
func (app *App) SetEndBlocker(...) {}
func (app *App) SetInitStater(...) {}
*/

func (app *App) LoadLatestVersion(mainKey sdk.SubstoreKey) error {
	app.ms.LoadLatestVersion()
	return app.initFromStore(mainKey)
}

func (app *App) LoadVersion(version int64, mainKey sdk.SubstoreKey) error {
	app.ms.LoadVersion(version)
	return app.initFromStore(mainKey)
}

// The last CommitID of the multistore.
func (app *App) LastCommitID() sdk.CommitID {
	return app.ms.LastCommitID()
}

// The last commited block height.
func (app *App) LastBlockHeight() int64 {
	return app.ms.LastCommitID().Version
}

// Initializes the remaining logic from app.ms.
func (app *App) initFromStore(mainKey sdk.SubstoreKey) error {
	lastCommitID := app.ms.LastCommitID()
	main := app.ms.GetKVStore(mainKey)
	header := abci.Header{}

	// Main store should exist.
	if main == nil {
		return errors.New("App expects MultiStore with 'main' KVStore")
	}

	// If we've committed before, we expect main://<mainHeaderKey>.
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

	// Set App state.
	app.header = header
	app.msCheck = nil
	app.msDeliver = nil
	app.valUpdates = nil

	return nil
}

//----------------------------------------

// Implements ABCI
func (app *App) Info(req abci.RequestInfo) abci.ResponseInfo {

	lastCommitID := app.ms.LastCommitID()

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
	app.msDeliver = app.ms.CacheMultiStore()
	app.msCheck = app.ms.CacheMultiStore()
	return
}

// Implements ABCI
func (app *App) CheckTx(txBytes []byte) (res abci.ResponseCheckTx) {

	result := app.runTx(true, txBytes)

	return abci.ResponseCheckTx{
		Code:      result.Code,
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
func (app *App) DeliverTx(txBytes []byte) (res abci.ResponseDeliverTx) {

	result := app.runTx(false, txBytes)

	// After-handler hooks.
	if result.Code == abci.CodeTypeOK {
		app.valUpdates = append(app.valUpdates, result.ValidatorUpdates...)
	} else {
		// Even though the Code is not OK, there will be some side
		// effects, like those caused by fee deductions or sequence
		// incrementations.
	}

	// Tell the blockchain engine (i.e. Tendermint).
	return abci.ResponseDeliverTx{
		Code:      result.Code,
		Data:      result.Data,
		Log:       result.Log,
		GasWanted: result.GasWanted,
		GasUsed:   result.GasUsed,
		Tags:      result.Tags,
	}
}

func (app *App) runTx(isCheckTx bool, txBytes []byte) (result sdk.Result) {

	// Handle any panics.
	defer func() {
		if r := recover(); r != nil {
			result = sdk.Result{
				Code: 1, // TODO
				Log:  fmt.Sprintf("Recovered: %v\n", r),
			}
		}
	}()

	var store sdk.MultiStore
	if isCheckTx {
		store = app.msCheck
	} else {
		store = app.msDeliver
	}

	// Initialize arguments to Handler.
	var ctx = sdk.NewContext(
		store,
		app.header,
		isCheckTx,
		txBytes,
	)

	// Decode the Tx.
	tx, err := app.txDecoder(txBytes)
	if err != nil {
		return sdk.Result{
			Code: 1, //  TODO
		}
	}

	// Run the ante handler.
	ctx, result, abort := app.defaultAnteHandler(ctx, tx)
	if isCheckTx || abort {
		return result
	}

	// Match and run route.
	msgType := tx.Type()
	handler := app.router.Route(msgType)
	result = handler(ctx, tx)

	return result
}

// Implements ABCI
func (app *App) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	res.ValidatorUpdates = app.valUpdates
	app.valUpdates = nil
	return
}

// Implements ABCI
func (app *App) Commit() (res abci.ResponseCommit) {
	app.msDeliver.Write()
	commitID := app.ms.Commit()
	app.logger.Debug("Commit synced",
		"commit", commitID,
	)
	return abci.ResponseCommit{
		Data: commitID.Hash,
	}
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
