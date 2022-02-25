package baseapp

import (
	"fmt"
	"io"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/snapshots"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/store/v2/multi"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// File for storing in-package BaseApp optional functions,
// for options that need access to non-exported fields of the BaseApp

// SetPruning sets a pruning option on the multistore associated with the app
func SetPruning(opts sdk.PruningOptions) StoreOption {
	return func(config *multi.StoreParams, _ uint64) error { config.Pruning = opts; return nil }
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

// SetSnapshotInterval sets the snapshot interval.
func SetSnapshotInterval(interval uint64) AppOptionFunc {
	return func(app *BaseApp) { app.SetSnapshotInterval(interval) }
}

// SetSnapshotKeepRecent sets the recent snapshots to keep.
func SetSnapshotKeepRecent(keepRecent uint32) AppOptionFunc {
	return func(app *BaseApp) { app.SetSnapshotKeepRecent(keepRecent) }
}

// SetSnapshotStore sets the snapshot store.
func SetSnapshotStore(snapshotStore *snapshots.Store) AppOptionOrdered {
	return AppOptionOrdered{
		func(app *BaseApp) { app.SetSnapshotStore(snapshotStore) },
		OptionOrderAfterStore,
	}
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

func (app *BaseApp) SetTxHandler(txHandler tx.Handler) {
	if app.sealed {
		panic("SetTxHandler() on sealed BaseApp")
	}

	app.txHandler = txHandler
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

// SetSnapshotStore sets the snapshot store.
func (app *BaseApp) SetSnapshotStore(snapshotStore *snapshots.Store) {
	if app.sealed {
		panic("SetSnapshotStore() on sealed BaseApp")
	}
	if snapshotStore == nil {
		app.snapshotManager = nil
		return
	}
	app.snapshotManager = snapshots.NewManager(snapshotStore, app.store, nil)
}

// SetSnapshotInterval sets the snapshot interval.
func (app *BaseApp) SetSnapshotInterval(snapshotInterval uint64) {
	if app.sealed {
		panic("SetSnapshotInterval() on sealed BaseApp")
	}
	app.snapshotInterval = snapshotInterval
}

// SetSnapshotKeepRecent sets the number of recent snapshots to keep.
func (app *BaseApp) SetSnapshotKeepRecent(snapshotKeepRecent uint32) {
	if app.sealed {
		panic("SetSnapshotKeepRecent() on sealed BaseApp")
	}
	app.snapshotKeepRecent = snapshotKeepRecent
}

// SetInterfaceRegistry sets the InterfaceRegistry.
func (app *BaseApp) SetInterfaceRegistry(registry types.InterfaceRegistry) {
	app.interfaceRegistry = registry
	app.grpcQueryRouter.SetInterfaceRegistry(registry)
}

// SetStreamingService is used to set a streaming service into the BaseApp hooks and load the listeners into the multistore
func (app *BaseApp) SetStreamingService(s StreamingService) {
	// add the listeners for each StoreKey
	for key, lis := range s.Listeners() {
		app.store.AddListeners(key, lis)
	}
	// register the StreamingService within the BaseApp
	// BaseApp will pass BeginBlock, DeliverTx, and EndBlock requests and responses to the streaming services to update their ABCI context
	app.abciListeners = append(app.abciListeners, s)
}
