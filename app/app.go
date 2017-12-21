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

	sdk "github.com/cosmos/cosmos-sdk"
)

const mainKeyHeader = "header"

// App - The ABCI application
type App struct {
	logger log.Logger

	// App name from abci.Info
	name string

	// DeliverTx (main) state
	store MultiStore

	// CheckTx state
	storeCheck CacheMultiStore

	// Current block header
	header *abci.Header

	// Handler for CheckTx and DeliverTx.
	handler sdk.Handler

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

func (app *App) SetStore(store MultiStore) {
	app.store = store
}

func (app *App) SetHandler(handler Handler) {
	app.handler = handler
}

func (app *App) LoadLatestVersion() error {
	store := app.store
	store.LoadLastVersion()
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

// DeliverTx - ABCI - dispatches to the handler
func (app *App) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {

	// Initialize arguments to Handler.
	var isCheckTx = false
	var ctx = sdk.NewContext(app.header, isCheckTx, txBytes)
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

// CheckTx - ABCI - dispatches to the handler
func (app *App) CheckTx(txBytes []byte) abci.ResponseCheckTx {

	// Initialize arguments to Handler.
	var isCheckTx = true
	var ctx = sdk.NewContext(app.header, isCheckTx, txBytes)
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

// Info - ABCI
func (app *App) Info(req abci.RequestInfo) abci.ResponseInfo {

	lastCommitID := app.store.LastCommitID()

	return abci.ResponseInfo{
		Data:             app.Name,
		LastBlockHeight:  lastCommitID.Version,
		LastBlockAppHash: lastCommitID.Hash,
	}
}

// SetOption - ABCI
func (app *App) SetOption(key string, value string) string {
	return "Not Implemented"
}

// Query - ABCI
func (app *App) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	/*
		XXX Make this work with MultiStore.
		XXX It will require some interfaces updates in store/types.go.

		if len(reqQuery.Data) == 0 {
			resQuery.Log = "Query cannot be zero length"
			resQuery.Code = abci.CodeType_EncodingError
			return
		}

		// set the query response height to current
		tree := app.state.Committed()

		height := reqQuery.Height
		if height == 0 {
			// TODO: once the rpc actually passes in non-zero
			// heights we can use to query right after a tx
			// we must retrun most recent, even if apphash
			// is not yet in the blockchain

			withProof := app.CommittedHeight() - 1
			if tree.Tree.VersionExists(withProof) {
				height = withProof
			} else {
				height = app.CommittedHeight()
			}
		}
		resQuery.Height = height

		switch reqQuery.Path {
		case "/store", "/key": // Get by key
			key := reqQuery.Data // Data holds the key bytes
			resQuery.Key = key
			if reqQuery.Prove {
				value, proof, err := tree.GetVersionedWithProof(key, height)
				if err != nil {
					resQuery.Log = err.Error()
					break
				}
				resQuery.Value = value
				resQuery.Proof = proof.Bytes()
			} else {
				value := tree.Get(key)
				resQuery.Value = value
			}

		default:
			resQuery.Code = abci.CodeType_UnknownRequest
			resQuery.Log = cmn.Fmt("Unexpected Query path: %v", reqQuery.Path)
		}
		return
	*/
}

// Commit implements abci.Application
func (app *App) Commit() (res abci.Result) {
	commitID := app.store.Commit()
	app.logger.Debug("Commit synced",
		"commit", commitID,
	)
	return abci.NewResultOK(hash, "")
}

// InitChain - ABCI
func (app *App) InitChain(req abci.RequestInitChain) {}

// BeginBlock - ABCI
func (app *App) BeginBlock(req abci.RequestBeginBlock) {
	app.header = req.Header
}

// EndBlock - ABCI
// Returns a list of all validator changes made in this block
func (app *App) EndBlock(height uint64) (res abci.ResponseEndBlock) {
	// XXX Update to res.Updates.
	res.Diffs = app.valUpdates
	app.valUpdates = nil
	return
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

// InitState - used to setup state (was SetOption)
// to be call from setting up the genesis file
func (app *InitApp) InitState(module, key, value string) error {
	state := app.Append()
	logger := app.Logger().With("module", module, "key", key)

	if module == sdk.ModuleNameBase {
		if key == sdk.ChainKey {
			app.info.SetChainID(state, value)
			return nil
		}
		logger.Error("Invalid genesis option")
		return fmt.Errorf("Unknown base option: %s", key)
	}

	log, err := app.initState.InitState(logger, state, module, key, value)
	if err != nil {
		logger.Error("Invalid genesis option", "err", err)
	} else {
		logger.Info(log)
	}
	return err
}

// InitChain - ABCI - sets the initial validators
func (app *InitApp) InitChain(req abci.RequestInitChain) {
	// return early if no InitValidator registered
	if app.initVals == nil {
		return
	}

	logger, store := app.Logger(), app.Append()
	app.initVals.InitValidators(logger, store, req.Validators)
}
