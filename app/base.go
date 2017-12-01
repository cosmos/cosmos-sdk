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
	"github.com/cosmos/cosmos-sdk/util"
)

const mainKeyHeader = "header"

// BaseApp - The ABCI application
type BaseApp struct {
	logger log.Logger

	// App name from abci.Info
	name string

	// DeliverTx (main) state
	ms MultiStore

	// CheckTx state
	msCheck CacheMultiStore

	// Cached validator changes from DeliverTx
	pending []*abci.Validator

	// Parser for the tx.
	txParser sdk.TxParser

	// Handler for CheckTx and DeliverTx.
	handler sdk.Handler
}

var _ abci.Application = &BaseApp{}

// CONTRACT: There exists a "main" KVStore.
func NewBaseApp(name string, ms MultiStore) (*BaseApp, error) {

	if ms.GetKVStore("main") == nil {
		return nil, errors.New("BaseApp expects MultiStore with 'main' KVStore")
	}

	logger := makeDefaultLogger()
	lastCommitID := ms.LastCommitID()
	curVersion := ms.CurrentVersion()
	main := ms.GetKVStore("main")
	header := (*abci.Header)(nil)
	msCheck := ms.CacheMultiStore()

	// SANITY
	if curVersion != lastCommitID.Version+1 {
		panic("CurrentVersion != LastCommitID.Version+1")
	}

	// If we've committed before, we expect ms.GetKVStore("main").Get("header")
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
		logger:  logger,
		name:    name,
		ms:      ms,
		msCheck: msCheck,
		pending: nil,
		header:  header,
	}
}

func (app *BaseApp) SetTxParser(parser TxParser) {
	app.txParser = parser
}

func (app *BaseApp) SetHandler(handler sdk.Handler) {
	app.handler = handler
}

//----------------------------------------

// DeliverTx - ABCI - dispatches to the handler
func (app *BaseApp) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {

	// TODO: use real context on refactor
	ctx := util.MockContext(
		app.GetChainID(),
		app.WorkingHeight(),
	)

	// Parse the transaction
	tx, err := app.parseTxFn(ctx, txBytes)
	if err != nil {
		err := sdk.TxParseError("").WithCause(err)
		return sdk.ResponseDeliverTxFromErr(err)
	}

	// Make handler deal with it
	data, err := app.handler.DeliverTx(ctx, app.ms, tx)
	if err != nil {
		return sdk.ResponseDeliverTxFromErr(err)
	}

	app.AddValChange(res.Diff)

	return abci.ResponseDeliverTx{
		Code: abci.CodeType_OK,
		Data: data,
		Log:  "", // TODO add log from ctx.logger
	}
}

// CheckTx - ABCI - dispatches to the handler
func (app *BaseApp) CheckTx(txBytes []byte) abci.ResponseCheckTx {

	// TODO: use real context on refactor
	ctx := util.MockContext(
		app.GetChainID(),
		app.WorkingHeight(),
	)

	// Parse the transaction
	tx, err := app.parseTxFn(ctx, txBytes)
	if err != nil {
		err := sdk.TxParseError("").WithCause(err)
		return sdk.ResponseCheckTxFromErr(err)
	}

	// Make handler deal with it
	data, err := app.handler.CheckTx(ctx, app.ms, tx)
	if err != nil {
		return sdk.ResponseCheckTx(err)
	}

	return abci.ResponseCheckTx{
		Code: abci.CodeType_OK,
		Data: data,
		Log:  "", // TODO add log from ctx.logger
	}
}

// Info - ABCI
func (app *BaseApp) Info(req abci.RequestInfo) abci.ResponseInfo {

	lastCommitID := app.ms.LastCommitID()

	return abci.ResponseInfo{
		Data:             app.Name,
		LastBlockHeight:  lastCommitID.Version,
		LastBlockAppHash: lastCommitID.Hash,
	}
}

// SetOption - ABCI
func (app *StoreApp) SetOption(key string, value string) string {
	return "Not Implemented"
}

// Query - ABCI
func (app *StoreApp) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	/* TODO

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
func (app *StoreApp) Commit() (res abci.Result) {
	/*
		hash, err := app.state.Commit(app.height)
		if err != nil {
			// die if we can't commit, not to recover
			panic(err)
		}
		app.logger.Debug("Commit synced",
			"height", app.height,
			"hash", fmt.Sprintf("%X", hash),
		)

		if app.state.Size() == 0 {
			return abci.NewResultOK(nil, "Empty hash for empty tree")
		}
		return abci.NewResultOK(hash, "")
	*/
}

// InitChain - ABCI
func (app *StoreApp) InitChain(req abci.RequestInitChain) {}

// BeginBlock - ABCI
func (app *StoreApp) BeginBlock(req abci.RequestBeginBlock) {
	// TODO
}

// EndBlock - ABCI
// Returns a list of all validator changes made in this block
func (app *StoreApp) EndBlock(height uint64) (res abci.ResponseEndBlock) {
	// TODO: cleanup in case a validator exists multiple times in the list
	res.Diffs = app.pending
	app.pending = nil
	return
}

// AddValChange is meant to be called by apps on DeliverTx
// results, this is added to the cache for the endblock
// changeset
func (app *StoreApp) AddValChange(diffs []*abci.Validator) {
	for _, d := range diffs {
		idx := pubKeyIndex(d, app.pending)
		if idx >= 0 {
			app.pending[idx] = d
		} else {
			app.pending = append(app.pending, d)
		}
	}
}

// return index of list with validator of same PubKey, or -1 if no match
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
