package baseapp

import (
	"fmt"
	"io"

	"github.com/cosmos/cosmos-sdk/codec/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2alpha1"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/multi"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// File for storing in-package BaseApp optional functions,
// for options that need access to non-exported fields of the BaseApp

// SetPruning sets a pruning option on the multistore associated with the app
func SetPruning(opts pruningtypes.PruningOptions) StoreOption {
	return func(config *multi.StoreParams, _ uint64) error { config.Pruning = opts; return nil }
}

func SetSubstoreTracer(w io.Writer) StoreOption {
	return func(cfg *multi.StoreParams, v uint64) error {
		cfg.SetTracer(w)
		return nil
	}
}

func SetTracerFor(skey storetypes.StoreKey, w io.Writer) StoreOption {
	return func(cfg *multi.StoreParams, v uint64) error {
		cfg.SetTracerFor(skey, w)
		return nil
	}
}

// SetSubstoreKVPair sets a key, value pair for the given substore inside a multistore
// Only works for v2alpha1/multi
func SetSubstoreKVPair(skeyName string, key, val []byte) AppOptionFunc {
	return func(bapp *BaseApp) { bapp.cms.(*multi.Store).SetSubstoreKVPair(skeyName, key, val) }
}

// SetMinGasPrices returns an option that sets the minimum gas prices on the app.
func SetMinGasPrices(gasPricesStr string) AppOptionFunc {
	gasPrices, err := sdk.ParseDecCoins(gasPricesStr)
	if err != nil {
		panic(fmt.Sprintf("invalid minimum gas prices: %v", err))
	}

	return func(bapp *BaseApp) { bapp.setMinGasPrices(gasPrices) }
}

// SetHaltHeight returns a BaseApp option function that sets the halt block height.
func SetHaltHeight(blockHeight uint64) AppOptionFunc {
	return func(bap *BaseApp) { bap.setHaltHeight(blockHeight) }
}

// SetInitialHeight returns a BaseApp option function that sets the initial block height.
func SetInitialHeight(blockHeight int64) AppOptionFunc {
	return func(bap *BaseApp) { bap.setInitialHeight(blockHeight) }
}

// SetHaltTime returns a BaseApp option function that sets the halt block time.
func SetHaltTime(haltTime uint64) AppOptionFunc {
	return func(bap *BaseApp) { bap.setHaltTime(haltTime) }
}

// SetMinRetainBlocks returns a BaseApp option function that sets the minimum
// block retention height value when determining which heights to prune during
// ABCI Commit.
func SetMinRetainBlocks(minRetainBlocks uint64) AppOptionFunc {
	return func(bapp *BaseApp) { bapp.setMinRetainBlocks(minRetainBlocks) }
}

// SetTrace will turn on or off trace flag
func SetTrace(trace bool) AppOptionFunc {
	return func(app *BaseApp) { app.setTrace(trace) }
}

// SetIndexEvents provides a BaseApp option function that sets the events to index.
func SetIndexEvents(ie []string) AppOptionFunc {
	return func(app *BaseApp) { app.setIndexEvents(ie) }
}

// SetInterBlockCache provides a BaseApp option function that sets the
// inter-block cache.
func SetInterBlockCache(cache sdk.MultiStorePersistentCache) AppOptionFunc {
	opt := func(cfg *multi.StoreParams, v uint64) error {
		cfg.PersistentCache = cache
		return nil
	}
	return func(app *BaseApp) { app.storeOpts = append(app.storeOpts, opt) }
}

// SetSubstores registers substores according to app configuration
func SetSubstores(keys ...storetypes.StoreKey) StoreOption {
	return func(config *multi.StoreParams, _ uint64) error {
		for _, key := range keys {
			typ, err := storetypes.StoreKeyToType(key)
			if err != nil {
				return err
			}
			if err = config.RegisterSubstore(key, typ); err != nil {
				return err
			}
		}
		return nil
	}
}

func SetSubstoresFromMaps(
	keys map[string]*storetypes.KVStoreKey,
	tkeys map[string]*storetypes.TransientStoreKey,
	memKeys map[string]*storetypes.MemoryStoreKey,
) StoreOption {
	return func(params *multi.StoreParams, _ uint64) error {
		if err := multi.RegisterSubstoresFromMap(params, keys); err != nil {
			return err
		}
		if err := multi.RegisterSubstoresFromMap(params, tkeys); err != nil {
			return err
		}
		if err := multi.RegisterSubstoresFromMap(params, memKeys); err != nil {
			return err
		}
		return nil
	}
}

// SetSnapshot sets the snapshot store.
func SetSnapshot(snapshotStore *snapshots.Store, opts snapshottypes.SnapshotOptions) AppOption {
	return AppOptionOrdered{
		func(app *BaseApp) { app.SetSnapshot(snapshotStore, opts) },
		OptionOrderAfterStore,
	}
}

func (app *BaseApp) SetName(name string) {
	if app.sealed {
		panic("SetName() on sealed BaseApp")
	}

	app.name = name
}

// SetParamStore sets a parameter store on the BaseApp.
func (app *BaseApp) SetParamStore(ps ParamStore) {
	if app.sealed {
		panic("SetParamStore() on sealed BaseApp")
	}

	app.paramStore = ps
}

// SetVersion sets the application's version string.
func (app *BaseApp) SetVersion(v string) {
	if app.sealed {
		panic("SetVersion() on sealed BaseApp")
	}
	app.version = v
}

// SetProtocolVersion sets the application's protocol version
func (app *BaseApp) SetProtocolVersion(v uint64) {
	app.appVersion = v
}

func (app *BaseApp) SetInitChainer(initChainer sdk.InitChainer) {
	if app.sealed {
		panic("SetInitChainer() on sealed BaseApp")
	}

	app.initChainer = initChainer
}

func (app *BaseApp) SetBeginBlocker(beginBlocker sdk.BeginBlocker) {
	if app.sealed {
		panic("SetBeginBlocker() on sealed BaseApp")
	}

	app.beginBlocker = beginBlocker
}

func (app *BaseApp) SetEndBlocker(endBlocker sdk.EndBlocker) {
	if app.sealed {
		panic("SetEndBlocker() on sealed BaseApp")
	}

	app.endBlocker = endBlocker
}

func (app *BaseApp) SetAnteHandler(ah sdk.AnteHandler) {
	if app.sealed {
		panic("SetAnteHandler() on sealed BaseApp")
	}

	app.anteHandler = ah
}

func (app *BaseApp) SetPostHandler(ph sdk.AnteHandler) {
	if app.sealed {
		panic("SetPostHandler() on sealed BaseApp")
	}

	app.postHandler = ph
}

func (app *BaseApp) SetAddrPeerFilter(pf sdk.PeerFilter) {
	if app.sealed {
		panic("SetAddrPeerFilter() on sealed BaseApp")
	}

	app.addrPeerFilter = pf
}

func (app *BaseApp) SetIDPeerFilter(pf sdk.PeerFilter) {
	if app.sealed {
		panic("SetIDPeerFilter() on sealed BaseApp")
	}

	app.idPeerFilter = pf
}

func (app *BaseApp) SetFauxMerkleMode() {
	if app.sealed {
		panic("SetFauxMerkleMode() on sealed BaseApp")
	}

	app.fauxMerkleMode = true
}

// SetCommitMultiStoreTracer sets the store tracer on the BaseApp's underlying
// CommitMultiStore.
func (app *BaseApp) SetCommitMultiStoreTracer(w io.Writer) {
	opt := func(cfg *multi.StoreParams, v uint64) error {
		cfg.TraceWriter = w
		return nil
	}
	app.storeOpts = append(app.storeOpts, opt)
}

// SetRouter allows us to customize the router.
func (app *BaseApp) SetRouter(router sdk.Router) {
	if app.sealed {
		panic("SetRouter() on sealed BaseApp")
	}
	app.router = router
}

// SetSnapshot sets the snapshot store and options.
func (app *BaseApp) SetSnapshot(snapshotStore *snapshots.Store, opts snapshottypes.SnapshotOptions) {
	if app.sealed {
		panic("SetSnapshot() on sealed BaseApp")
	}
	if snapshotStore == nil || opts.Interval == snapshottypes.SnapshotIntervalOff {
		app.snapshotManager = nil
		return
	}
	app.cms.SetSnapshotInterval(opts.Interval)
	app.snapshotManager = snapshots.NewManager(snapshotStore, opts, app.cms, nil, app.logger)
}

// SetInterfaceRegistry sets the InterfaceRegistry.
func (app *BaseApp) SetInterfaceRegistry(registry types.InterfaceRegistry) {
	app.interfaceRegistry = registry
	app.grpcQueryRouter.SetInterfaceRegistry(registry)
	app.msgServiceRouter.SetInterfaceRegistry(registry)
}

// SetStreamingService is used to set a streaming service into the BaseApp hooks and load the listeners into the multistore
func (app *BaseApp) SetStreamingService(s StreamingService) {
	// add the listeners for each StoreKey
	for key, lis := range s.Listeners() {
		app.cms.AddListeners(key, lis)
	}
	// register the StreamingService within the BaseApp
	// BaseApp will pass BeginBlock, DeliverTx, and EndBlock requests and responses to the streaming services to update their ABCI context
	app.abciListeners = append(app.abciListeners, s)
}

// SetTxDecoder sets the TxDecoder if it wasn't provided in the BaseApp constructor.
func (app *BaseApp) SetTxDecoder(txDecoder sdk.TxDecoder) {
	app.txDecoder = txDecoder
}
