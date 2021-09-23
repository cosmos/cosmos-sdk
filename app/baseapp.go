package app

import (
	"fmt"
	io "io"
	"net/http"
	"path/filepath"
	"reflect"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/store"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Name string

var BaseAppProvider = container.Options(
	container.AutoGroupTypes(reflect.TypeOf(func(*baseapp.BaseApp) {})),
	container.AutoGroupTypes(reflect.TypeOf(func(types.AppOptions) func(*baseapp.BaseApp) { return nil })),
	container.Provide(provideBaseApp),
)

type baseAppInput struct {
	container.In

	Name         Name          `optional:"true"`
	TxDecoder    sdk.TxDecoder `optional:"true"`
	TypeRegistry codectypes.TypeRegistry
	Options      []func(*baseapp.BaseApp)

	// AppOptOptions are functions which provide a BaseApp option based on some AppOptions provided at runtime
	AppOptOptions []func(types.AppOptions) func(*baseapp.BaseApp)
}

type app struct {
	*baseapp.BaseApp

	handlers     map[string]Handler
	typeRegistry codectypes.TypeRegistry
}

var _ types.Application

func provideBaseApp(inputs baseAppInput) types.AppCreator {
	return func(logger log.Logger, db dbm.DB, tracer io.Writer, appOpts types.AppOptions) types.Application {
		name := inputs.Name
		if name == "" {
			name = "simapp"
		}

		txDecoder := inputs.TxDecoder
		if txDecoder == nil {
			txDecoder = func(txBytes []byte) (sdk.Tx, error) {
				return nil, fmt.Errorf("no TxDecoder, can't decode transactions")
			}
		}

		var cache sdk.MultiStorePersistentCache

		if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
			cache = store.NewCommitKVStoreCacheManager()
		}

		pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
		if err != nil {
			panic(err)
		}

		snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
		snapshotDB, err := sdk.NewLevelDB("metadata", snapshotDir)
		if err != nil {
			panic(err)
		}
		snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
		if err != nil {
			panic(err)
		}

		opts := []func(*baseapp.BaseApp){
			baseapp.SetPruning(pruningOpts),
			baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
			baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
			baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
			baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
			baseapp.SetInterBlockCache(cache),
			baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
			baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
			baseapp.SetSnapshotStore(snapshotStore),
			baseapp.SetSnapshotInterval(cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval))),
			baseapp.SetSnapshotKeepRecent(cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent))),
		}

		for _, appOptOpt := range inputs.AppOptOptions {
			opts = append(opts, appOptOpt(appOpts))
		}

		opts = append(opts, inputs.Options...)

		baseApp := baseapp.NewBaseApp(string(name), logger, db, txDecoder, opts...)

		if tracer != nil {
			baseApp.SetCommitMultiStoreTracer(tracer)
		}

		return &app{
			BaseApp:      baseApp,
			typeRegistry: inputs.TypeRegistry,
		}
	}
}

func (a app) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	//// ** TODO: Register legacy and grpc-gateway routes for all modules.
	//for _, b := range a.moduleBasics {
	//	b.RegisterRESTRoutes(clientCtx, apiSvr.Router)
	//}
	//
	//for _, b := range a.moduleBasics {
	//	b.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	//}

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		registerSwaggerAPI(apiSvr.Router)
	}
}

// RegisterSwaggerAPI registers swagger route with API Server
func registerSwaggerAPI(rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

func (a app) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(a.GRPCQueryRouter(), clientCtx, a.Simulate, a.typeRegistry)
}

func (a app) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(a.GRPCQueryRouter(), clientCtx, a.typeRegistry)
}
