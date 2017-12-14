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

// BaseApp - The ABCI application
type BaseApp struct {
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
	valSetDiff []abci.Validator
}

var _ abci.Application = &BaseApp{}

// CONTRACT: There exists a "main" KVStore.
func NewBaseApp(name string, store CommitMultiStore) (*BaseApp, error) {

	if store.GetKVStore("main") == nil {
		return nil, errors.New("BaseApp expects MultiStore with 'main' KVStore")
	}

	logger := makeDefaultLogger()
	lastCommitID := store.LastCommitID()
	curVersion := store.CurrentVersion()
	main := store.GetKVStore("main")
	header := (*abci.Header)(nil)
	storeCheck := store.CacheMultiStore()

	// SANITY
	if curVersion != lastCommitID.Version+1 {
		panic("CurrentVersion != LastCommitID.Version+1")
	}

	// If we've committed before, we expect store.GetKVStore("main").Get("header")
	if !lastCommitID.IsZero() {
		headerBytes, ok := main.Get(mainKeyHeader)
		if !ok {
			return nil, errors.New(fmt.Sprintf("Version > 0 but missing key %s", mainKeyHeader))
		}
		err = proto.Unmarshal(headerBytes, header)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse Header")
		}

		// SANITY: Validate Header
		if header.Height != curVersion-1 {
			errStr := fmt.Sprintf("Expected header.Height %v but got %v", version, headerHeight)
			panic(errStr)
		}
	}

	return &BaseApp{
		logger:     logger,
		name:       name,
		store:      store,
		storeCheck: storeCheck,
		header:     header,
		hander:     nil, // set w/ .WithHandler()
		valSetDiff: nil,
	}
}

func (app *BaseApp) WithHandler(handler sdk.Handler) *BaseApp {
	app.handler = handler
}

//----------------------------------------

// DeliverTx - ABCI - dispatches to the handler
func (app *BaseApp) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {

	// Initialize arguments to Handler.
	var isCheckTx = false
	var ctx = sdk.NewContext(app.header, isCheckTx, txBytes)
	var store = app.store
	var tx Tx = nil // nil until a decorator parses one.

	// Run the handler.
	var result = app.handler(ctx, app.store, tx)

	// After-handler hooks.
	// TODO move to app.afterHandler(...).
	if result.Code == abci.CodeType_OK {
		// XXX No longer "diff", we need to replace old entries.
		app.ValSetDiff = append(app.ValSetDiff, result.ValSetDiff)
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
func (app *BaseApp) CheckTx(txBytes []byte) abci.ResponseCheckTx {

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
func (app *BaseApp) Info(req abci.RequestInfo) abci.ResponseInfo {

	lastCommitID := app.store.LastCommitID()

	return abci.ResponseInfo{
		Data:             app.Name,
		LastBlockHeight:  lastCommitID.Version,
		LastBlockAppHash: lastCommitID.Hash,
	}
}

// SetOption - ABCI
func (app *BaseApp) SetOption(key string, value string) string {
	return "Not Implemented"
}

// Query - ABCI
func (app *BaseApp) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
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
func (app *BaseApp) Commit() (res abci.Result) {
	commitID := app.store.Commit()
	app.logger.Debug("Commit synced",
		"commit", commitID,
	)
	return abci.NewResultOK(hash, "")
}

// InitChain - ABCI
func (app *BaseApp) InitChain(req abci.RequestInitChain) {}

// BeginBlock - ABCI
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) {
	app.header = req.Header
}

// EndBlock - ABCI
// Returns a list of all validator changes made in this block
func (app *BaseApp) EndBlock(height uint64) (res abci.ResponseEndBlock) {
	// XXX Update to res.Updates.
	res.Diffs = app.valSetDiff
	app.valSetDiff = nil
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
