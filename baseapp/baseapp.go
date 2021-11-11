package baseapp

import (
	"context"
	"fmt"
	// "reflect"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec/types"
	dbm "github.com/cosmos/cosmos-sdk/db"
	// dbutil "github.com/cosmos/cosmos-sdk/internal/db"
	"github.com/cosmos/cosmos-sdk/snapshots"
	// "github.com/cosmos/cosmos-sdk/store"
	// "github.com/cosmos/cosmos-sdk/store/rootmulti"
	stypes "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/store/v2/flat"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

const (
	runTxModeCheck    runTxMode = iota // Check a transaction
	runTxModeReCheck                   // Recheck a (pending) transaction after a commit
	runTxModeSimulate                  // Simulate a transaction
	runTxModeDeliver                   // Deliver a transaction
)

var (
	_ abci.Application = (*BaseApp)(nil)
)

type (
	// Enum mode for app.runTx
	runTxMode uint8

	// StoreLoader defines a customizable function to control how we load the CommitRootStore
	// from disk. This is useful for state migration, when loading a datastore written with
	// an older version of the software. In particular, if a module changed the substore key name
	// (or removed a substore) between two versions of the software.
	// StoreLoader func(ms sdk.CommitRootStore) error

	// StoreLoader func(initialVersion uint64) (sdk.CommitRootStore, error)
	// StoreLoader func(sdk.RootStoreConfig) (sdk.CommitRootStore, error)
	// StoreLoader func(*sdk.RootStoreConfig, uint64) (sdk.CommitRootStore, error)
	StoreOption      func(*sdk.RootStoreConfig, uint64) error
	StoreConstructor func(dbm.DBConnection, sdk.RootStoreConfig) (sdk.CommitRootStore, error)
)

// BaseApp reflects the ABCI application implementation.
type BaseApp struct { // nolint: maligned
	// initialized on creation
	logger    log.Logger
	name      string // application name from abci.Info
	db        dbm.DBConnection
	storeCtor StoreConstructor
	storeOpts []StoreOption       // options to configure root store
	store     sdk.CommitRootStore // Main (uncached) state
	// storeLoader StoreLoader         // function to handle store loading
	queryRouter       sdk.QueryRouter  // router for redirecting query calls
	grpcQueryRouter   *GRPCQueryRouter // router for redirecting gRPC query calls
	interfaceRegistry types.InterfaceRegistry
	txDecoder         sdk.TxDecoder // unmarshal []byte into sdk.Tx

	txHandler      tx.Handler       // txHandler for {Deliver,Check}Tx and simulations
	initChainer    sdk.InitChainer  // initialize state with validators and state blob
	beginBlocker   sdk.BeginBlocker // logic to run before any txs
	endBlocker     sdk.EndBlocker   // logic to run after all txs, and to determine valset changes
	addrPeerFilter sdk.PeerFilter   // filter peers by address and port
	idPeerFilter   sdk.PeerFilter   // filter peers by node ID
	fauxMerkleMode bool             // if true, IAVL MountStores uses MountStoresDB for simulation speed.

	// manages snapshots, i.e. dumps of app state at certain intervals
	snapshotManager    *snapshots.Manager
	snapshotInterval   uint64 // block interval between state sync snapshots
	snapshotKeepRecent uint32 // recent state sync snapshots to keep

	// volatile states:
	//
	// checkState is set on InitChain and reset on Commit
	// deliverState is set on InitChain and BeginBlock and set to nil on Commit
	checkState   *state // for CheckTx
	deliverState *state // for DeliverTx

	// absent validators from begin block
	voteInfos []abci.VoteInfo

	// paramStore is used to query for ABCI consensus parameters from an
	// application parameter store.
	paramStore ParamStore

	// The minimum gas prices a validator is willing to accept for processing a
	// transaction. This is mainly used for DoS and spam prevention.
	minGasPrices sdk.DecCoins

	// initialHeight is the initial height at which we start the baseapp
	initialHeight int64

	// flag for sealing options and parameters to a BaseApp
	sealed bool

	// block height at which to halt the chain and gracefully shutdown
	haltHeight uint64

	// minimum block time (in Unix seconds) at which to halt the chain and gracefully shutdown
	haltTime uint64

	// minRetainBlocks defines the minimum block height offset from the current
	// block being committed, such that all blocks past this offset are pruned
	// from Tendermint. It is used as part of the process of determining the
	// ResponseCommit.RetainHeight value during ABCI Commit. A value of 0 indicates
	// that no blocks should be pruned.
	//
	// Note: Tendermint block pruning is dependant on this parameter in conunction
	// with the unbonding (safety threshold) period, state pruning and state sync
	// snapshot parameters to determine the correct minimum value of
	// ResponseCommit.RetainHeight.
	minRetainBlocks uint64

	// application's version string
	version string

	// application's protocol version that increments on every upgrade
	// if BaseApp is passed to the upgrade keeper's NewKeeper method.
	appVersion uint64

	// trace set will return full stack traces for errors in ABCI Log field
	trace bool

	// indexEvents defines the set of events in the form {eventType}.{attributeKey},
	// which informs Tendermint what to index. If empty, all events will be indexed.
	indexEvents map[string]struct{}

	// abciListeners for hooking into the ABCI message processing of the BaseApp
	// and exposing the requests and responses to external consumers
	abciListeners []ABCIListener
}

// OptionOrder represents the required ordering for options that are order dependent
type OptionOrder int

const (
	OptionOrderDefault = iota
	OptionOrderAfterStore
)

// AppOption is a configuration option for a BaseApp
type AppOption interface {
	Apply(*BaseApp)
	Order() OptionOrder
}

type AppOptionFunc func(*BaseApp)

type AppOptionOrdered struct {
	AppOptionFunc
	order OptionOrder
}

func (opt AppOptionOrdered) Order() OptionOrder { return opt.order }

func (opt AppOptionFunc) Apply(app *BaseApp) { opt(app) }
func (opt AppOptionFunc) Order() OptionOrder { return OptionOrderDefault }

func (opt StoreOption) Apply(app *BaseApp) { app.storeOpts = append(app.storeOpts, opt) }
func (opt StoreOption) Order() OptionOrder { return OptionOrderDefault }

// NewBaseApp returns a reference to an initialized BaseApp. It accepts a
// variadic number of option functions, which act on the BaseApp to set
// configuration choices.
//
// NOTE: The db is used to store the version number for now.
func NewBaseApp(
	name string,
	logger log.Logger,
	db dbm.DBConnection,
	txDecoder sdk.TxDecoder,
	options ...AppOption,
) *BaseApp {
	app := &BaseApp{
		logger:          logger,
		name:            name,
		db:              db,
		storeCtor:       DefaultStoreConstructor,
		queryRouter:     NewQueryRouter(),
		grpcQueryRouter: NewGRPCQueryRouter(),
		txDecoder:       txDecoder,
		fauxMerkleMode:  false,
	}

	var afterStoreOpts []AppOption
	for _, option := range options {
		if int(option.Order()) > int(OptionOrderDefault) {
			afterStoreOpts = append(afterStoreOpts, option)
		} else {
			option.Apply(app)
		}
	}

	err := app.loadStore()
	if err != nil {
		panic(err)
	}
	for _, option := range afterStoreOpts {
		option.Apply(app)
	}

	// // TODO: conditional loading of multistore/rootstore
	// if true {
	// 	var err error
	// 	opts := store.RootStoreConfig{PersistentCache: app.interBlockCache}
	// 	app.store, err = store.NewCommitRootStore(db, opts)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// } else {
	// 	// app.store = nil store.NewCommitMultiStore(dbutil.ConnectionAsTmdb(db))
	// 	app.store = store.MultiStoreAsRootStore(db)
	// }

	return app
}

// Name returns the name of the BaseApp.
func (app *BaseApp) Name() string {
	return app.name
}

// AppVersion returns the application's protocol version.
func (app *BaseApp) AppVersion() uint64 {
	return app.appVersion
}

// Version returns the application's version string.
func (app *BaseApp) Version() string {
	return app.version
}

// Logger returns the logger of the BaseApp.
func (app *BaseApp) Logger() log.Logger {
	return app.logger
}

// Trace returns the boolean value for logging error stack traces.
func (app *BaseApp) Trace() bool {
	return app.trace
}

func (app *BaseApp) loadStore() error {
	versions, err := app.db.Versions()
	if err != nil {
		return err
	}
	latest := versions.Last()
	config := flat.DefaultRootStoreConfig()
	for _, opt := range app.storeOpts {
		opt(&config, latest)
	}
	app.store, err = app.storeCtor(app.db, config)
	if err != nil {
		return fmt.Errorf("failed to load latest version: %w", err)
	}
	return nil
}

func (app *BaseApp) CloseStore() error {
	return app.store.Close()
}

// DefaultStoreConstructor attempts to create a new store, but loads from existing data if present.
func DefaultStoreConstructor(db dbm.DBConnection, config sdk.RootStoreConfig) (stypes.CommitRootStore, error) {
	return flat.NewRootStore(db, config)
}

// LastCommitID returns the last CommitID of the multistore.
func (app *BaseApp) LastCommitID() stypes.CommitID {
	return app.store.LastCommitID()
}

// LastBlockHeight returns the last committed block height.
func (app *BaseApp) LastBlockHeight() int64 {
	return app.store.LastCommitID().Version
}

// Init sets the check state and seals the app. It will panic if
// called more than once on a running BaseApp.
func (app *BaseApp) Init() error {
	if app.sealed {
		panic("cannot call Init: baseapp already sealed")
	}

	// needed for the export command which inits from store but never calls initchain
	app.setCheckState(tmproto.Header{})
	app.Seal()

	// make sure the snapshot interval is a multiple of the pruning KeepEvery interval
	if app.snapshotManager != nil && app.snapshotInterval > 0 {
		pruningOpts := app.store.GetPruning()
		if pruningOpts.KeepEvery > 0 && app.snapshotInterval%pruningOpts.KeepEvery != 0 {
			return fmt.Errorf(
				"state sync snapshot interval %v must be a multiple of pruning keep every interval %v",
				app.snapshotInterval, pruningOpts.KeepEvery)
		}
	}

	return nil
}

func (app *BaseApp) setMinGasPrices(gasPrices sdk.DecCoins) {
	app.minGasPrices = gasPrices
}

func (app *BaseApp) setHaltHeight(haltHeight uint64) {
	app.haltHeight = haltHeight
}

func (app *BaseApp) setHaltTime(haltTime uint64) {
	app.haltTime = haltTime
}

func (app *BaseApp) setMinRetainBlocks(minRetainBlocks uint64) {
	app.minRetainBlocks = minRetainBlocks
}

func (app *BaseApp) setTrace(trace bool) {
	app.trace = trace
}

func (app *BaseApp) setIndexEvents(ie []string) {
	app.indexEvents = make(map[string]struct{})

	for _, e := range ie {
		app.indexEvents[e] = struct{}{}
	}
}

// QueryRouter returns the QueryRouter of a BaseApp.
func (app *BaseApp) QueryRouter() sdk.QueryRouter { return app.queryRouter }

// Seal seals a BaseApp. It prohibits any further modifications to a BaseApp.
func (app *BaseApp) Seal() { app.sealed = true }

// IsSealed returns true if the BaseApp is sealed and false otherwise.
func (app *BaseApp) IsSealed() bool { return app.sealed }

// setCheckState sets the BaseApp's checkState with a branched multi-store
// (i.e. a CacheMultiStore) and a new Context with the same multi-store branch,
// provided header, and minimum gas prices set. It is set on InitChain and reset
// on Commit.
func (app *BaseApp) setCheckState(header tmproto.Header) {
	ms := app.store.CacheRootStore()
	app.checkState = &state{
		ms:  ms,
		ctx: sdk.NewContext(ms, header, true, app.logger).WithMinGasPrices(app.minGasPrices),
	}
}

// setDeliverState sets the BaseApp's deliverState with a branched multi-store
// (i.e. a CacheMultiStore) and a new Context with the same multi-store branch,
// and provided header. It is set on InitChain and BeginBlock and set to nil on
// Commit.
func (app *BaseApp) setDeliverState(header tmproto.Header) {
	ms := app.store.CacheRootStore()
	app.deliverState = &state{
		ms:  ms,
		ctx: sdk.NewContext(ms, header, false, app.logger),
	}
}

// GetConsensusParams returns the current consensus parameters from the BaseApp's
// ParamStore. If the BaseApp has no ParamStore defined, nil is returned.
func (app *BaseApp) GetConsensusParams(ctx sdk.Context) *abci.ConsensusParams {
	if app.paramStore == nil {
		return nil
	}

	cp := new(abci.ConsensusParams)

	if app.paramStore.Has(ctx, ParamStoreKeyBlockParams) {
		var bp abci.BlockParams

		app.paramStore.Get(ctx, ParamStoreKeyBlockParams, &bp)
		cp.Block = &bp
	}

	if app.paramStore.Has(ctx, ParamStoreKeyEvidenceParams) {
		var ep tmproto.EvidenceParams

		app.paramStore.Get(ctx, ParamStoreKeyEvidenceParams, &ep)
		cp.Evidence = &ep
	}

	if app.paramStore.Has(ctx, ParamStoreKeyValidatorParams) {
		var vp tmproto.ValidatorParams

		app.paramStore.Get(ctx, ParamStoreKeyValidatorParams, &vp)
		cp.Validator = &vp
	}

	return cp
}

// StoreConsensusParams sets the consensus parameters to the baseapp's param store.
func (app *BaseApp) StoreConsensusParams(ctx sdk.Context, cp *abci.ConsensusParams) {
	if app.paramStore == nil {
		panic("cannot store consensus params with no params store set")
	}

	if cp == nil {
		return
	}

	app.paramStore.Set(ctx, ParamStoreKeyBlockParams, cp.Block)
	app.paramStore.Set(ctx, ParamStoreKeyEvidenceParams, cp.Evidence)
	app.paramStore.Set(ctx, ParamStoreKeyValidatorParams, cp.Validator)
	// We're explicitly not storing the Tendermint app_version in the param store. It's
	// stored instead in the x/upgrade store, with its own bump logic.
}

// getMaximumBlockGas gets the maximum gas from the consensus params. It panics
// if maximum block gas is less than negative one and returns zero if negative
// one.
func (app *BaseApp) getMaximumBlockGas(ctx sdk.Context) uint64 {
	cp := app.GetConsensusParams(ctx)
	if cp == nil || cp.Block == nil {
		return 0
	}

	maxGas := cp.Block.MaxGas

	switch {
	case maxGas < -1:
		panic(fmt.Sprintf("invalid maximum block gas: %d", maxGas))

	case maxGas == -1:
		return 0

	default:
		return uint64(maxGas)
	}
}

func (app *BaseApp) validateHeight(req abci.RequestBeginBlock) error {
	if req.Header.Height < 1 {
		return fmt.Errorf("invalid height: %d", req.Header.Height)
	}

	// expectedHeight holds the expected height to validate.
	var expectedHeight int64
	if app.LastBlockHeight() == 0 && app.initialHeight > 1 {
		// In this case, we're validating the first block of the chain (no
		// previous commit). The height we're expecting is the initial height.
		expectedHeight = app.initialHeight
	} else {
		// This case can means two things:
		// - either there was already a previous commit in the store, in which
		// case we increment the version from there,
		// - or there was no previous commit, and initial version was not set,
		// in which case we start at version 1.
		expectedHeight = app.LastBlockHeight() + 1
	}

	if req.Header.Height != expectedHeight {
		return fmt.Errorf("invalid height: %d; expected: %d", req.Header.Height, expectedHeight)
	}

	return nil
}

// Returns the applications's deliverState if app is in runTxModeDeliver,
// otherwise it returns the application's checkstate.
func (app *BaseApp) getState(mode runTxMode) *state {
	if mode == runTxModeDeliver {
		return app.deliverState
	}

	return app.checkState
}

// retrieve the context for the tx w/ txBytes and other memoized values.
func (app *BaseApp) getContextForTx(mode runTxMode, txBytes []byte) context.Context {
	ctx := app.getState(mode).ctx.
		WithTxBytes(txBytes).
		WithVoteInfos(app.voteInfos)

	ctx = ctx.WithConsensusParams(app.GetConsensusParams(ctx))

	if mode == runTxModeReCheck {
		ctx = ctx.WithIsReCheckTx(true)
	}

	if mode == runTxModeSimulate {
		ctx, _ = ctx.CacheContext()
	}

	return sdk.WrapSDKContext(ctx)
}
