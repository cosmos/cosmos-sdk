package baseapp

import (
	"github.com/pkg/errors"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

// Key to store the header in the DB itself.
// Use the db directly instead of a store to avoid
// conflicts with handlers writing to the store
// and to avoid affecting the Merkle root.
var dbHeaderKey = []byte("header")

// The ABCI application
type BaseApp struct {
	// initialized on creation
	Logger     log.Logger
	name       string           // application name from abci.Info
	db         dbm.DB           // common DB backend
	cms        CommitMultiStore // Main (uncached) state
	codespacer *Codespacer      // handle module codespacing

	// must be set
	checkTxer   CheckTxer   // initialize state with validators and state blob
	deliverTxer DeliverTxer // logic to run before any txs

	// may be nil
	initChainer  InitChainer  // initialize state with validators and state blob
	beginBlocker BeginBlocker // logic to run before any txs
	endBlocker   EndBlocker   // logic to run after all txs, and to determine valset changes

	//--------------------
	// Volatile
	// checkState is set on initialization and reset on Commit.
	// deliverState is set in InitChain and BeginBlock and cleared on Commit.
	// See methods setCheckState and setDeliverState.
	// .valUpdates accumulate in DeliverTx and are reset in BeginBlock.
	// QUESTION: should we put valUpdates in the deliverState.ctx?
	checkState   *state           // for CheckTx
	deliverState *state           // for DeliverTx
	valUpdates   []abci.Validator // cached validator changes from DeliverTx
}

var _ abci.Application = (*BaseApp)(nil)

// Create and name new BaseApp
// NOTE: The db is used to store the version number for now.
func NewBaseApp(name string, logger log.Logger, db dbm.DB, cms CommitMultiStore) *BaseApp {
	app := &BaseApp{
		Logger:     logger,
		name:       name,
		db:         db,
		cms:        cms,
		codespacer: NewCodespacer(),
	}
	app.codespacer.RegisterOrPanic(CodespaceRoot)
	app.codespacer.RegisterOrPanic(CodespaceUndefined)

	return app
}

// BaseApp Name
func (app *BaseApp) Name() string {
	return app.name
}

// Register the next available codespace through the baseapp's codespacer, starting from a default
func (app *BaseApp) RegisterCodespace(codespace CodespaceType) CodespaceType {
	return app.codespacer.RegisterNext(codespace)
}

// Mount a store to the provided key in the BaseApp multistore
func (app *BaseApp) MountStoresIAVL(keys ...*KVStoreKey) {
	for _, key := range keys {
		app.MountStore(key, StoreTypeIAVL)
	}
}

// Mount a store to the provided key in the BaseApp multistore, using a specified DB
func (app *BaseApp) MountStoreWithDB(key StoreKey, typ StoreType, db dbm.DB) {
	app.cms.MountStoreWithDB(key, typ, db)
}

// Mount a store to the provided key in the BaseApp multistore, using the default DB
func (app *BaseApp) MountStore(key StoreKey, typ StoreType) {
	app.cms.MountStoreWithDB(key, typ, nil)
}

// nolint - Set functions
func (app *BaseApp) SetInitChainer(initChainer InitChainer) {
	app.initChainer = initChainer
}
func (app *BaseApp) SetBeginBlocker(beginBlocker BeginBlocker) {
	app.beginBlocker = beginBlocker
}
func (app *BaseApp) SetEndBlocker(endBlocker EndBlocker) {
	app.endBlocker = endBlocker
}
func (app *BaseApp) SetCheckTxer(checkTxer CheckTxer) {
	app.checkTxer = checkTxer
}
func (app *BaseApp) SetDeliverTxer(deliverTxer DeliverTxer) {
	app.deliverTxer = deliverTxer
}

// load latest application version
func (app *BaseApp) LoadLatestVersion(mainKey StoreKey) error {
	app.cms.LoadLatestVersion()
	return app.initFromStore(mainKey)
}

// load application version
func (app *BaseApp) LoadVersion(version int64, mainKey StoreKey) error {
	app.cms.LoadVersion(version)
	return app.initFromStore(mainKey)
}

// the last CommitID of the multistore
func (app *BaseApp) LastCommitID() CommitID {
	return app.cms.LastCommitID()
}

// the last commited block height
func (app *BaseApp) LastBlockHeight() int64 {
	return app.cms.LastCommitID().Version
}

// initializes the remaining logic from app.cms
func (app *BaseApp) initFromStore(mainKey StoreKey) error {

	// main store should exist.
	// TODO: we don't actually need the main store here
	main := app.cms.GetKVStore(mainKey)
	if main == nil {
		return errors.New("BaseApp expects MultiStore with 'main' KVStore")
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
				return errors.Wrap(err, "Failed to parse Header")
			}
			lastVersion := lastCommitID.Version
			if header.Height != lastVersion {
				errStr := fmt.Sprintf("Expected db://%s.Height %v but got %v", dbHeaderKey, lastVersion, header.Height)
				return errors.New(errStr)
			}
		}
	*/

	// initialize Check state
	app.setCheckState(abci.Header{})

	return nil
}

// NewContext returns a new Context with the correct store, the given header, and nil txBytes.
func (app *BaseApp) NewContext(isCheckTx bool, header abci.Header) Context {
	if isCheckTx {
		return NewContext(app.checkState.ms, header, true, nil, app.Logger)
	}
	return NewContext(app.deliverState.ms, header, false, nil, app.Logger)
}

type state struct {
	ms  CacheMultiStore
	ctx Context
}

func (st *state) CacheMultiStore() CacheMultiStore {
	return st.ms.CacheMultiStore()
}

func (app *BaseApp) setCheckState(header abci.Header) {
	ms := app.cms.CacheMultiStore()
	app.checkState = &state{
		ms:  ms,
		ctx: NewContext(ms, header, true, nil, app.Logger),
	}
}

func (app *BaseApp) setDeliverState(header abci.Header) {
	ms := app.cms.CacheMultiStore()
	app.deliverState = &state{
		ms:  ms,
		ctx: NewContext(ms, header, false, nil, app.Logger),
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
	if app.initChainer == nil {
		return
	}

	// Initialize the deliver state and run initChain
	app.setDeliverState(abci.Header{})
	app.initChainer(app.deliverState.ctx, req) // no error

	// NOTE: we don't commit, but BeginBlock for block 1
	// starts from this deliverState
	return
}

// Implements ABCI.
// Delegates to CommitMultiStore if it implements Queryable
func (app *BaseApp) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	queryable, ok := app.cms.(Queryable)
	if !ok {
		msg := "application doesn't support queries"
		return ErrUnknownRequest(msg).QueryResult()
	}
	return queryable.Query(req)
}

// Implements ABCI
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	// Initialize the DeliverTx state.
	// If this is the first block, it should already
	// be initialized in InitChain. It may also be nil
	// if this is a test and InitChain was never called.
	if app.deliverState == nil {
		app.setDeliverState(req.Header)
	}
	app.valUpdates = nil
	if app.beginBlocker != nil {
		res = app.beginBlocker(app.deliverState.ctx, req)
	}
	return
}

// Implements ABCI
func (app *BaseApp) CheckTx(txBytes []byte) abci.ResponseCheckTx {
	var result Result
	if app.checkTxer != nil {
		result = app.checkTxer(app.checkState.ctx.WithTxBytes(txBytes), txBytes)
	} else {
		panic("set checkTxer")
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
func (app *BaseApp) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {
	var result Result
	if app.deliverTxer != nil {
		result = app.deliverTxer(app.deliverState.ctx.WithTxBytes(txBytes), txBytes)
	} else {
		panic("set deliverTxer")
	}

	// After-handler hooks.
	if result.IsOK() {
		app.valUpdates = append(app.valUpdates, result.ValidatorUpdates...)
	} else {
		// Even though the Result.Code is not OK, there are still effects,
		// for example fee deductions and sequence incrementing.
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

// Implements ABCI
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	if app.endBlocker != nil {
		res = app.endBlocker(app.deliverState.ctx, req)
	} else {
		res.ValidatorUpdates = app.valUpdates
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
