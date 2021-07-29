package app

import (
	"fmt"
	io "io"
	"net/http"
	"reflect"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types/module"
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
	container.OnePerScopeTypes(reflect.TypeOf((*module.AppModuleBasic)(nil)).Elem()),
	container.Provide(provideBaseApp),
)

type baseAppInput struct {
	container.StructArgs

	Name         Name          `optional:"true"`
	TxDecoder    sdk.TxDecoder `optional:"true"`
	TypeRegistry codectypes.TypeRegistry
	ModuleBasics map[string]module.AppModuleBasic
	Options      []func(*baseapp.BaseApp)
}

type app struct {
	*baseapp.BaseApp

	moduleBasics map[string]module.AppModuleBasic
	typeRegistry codectypes.TypeRegistry
}

var _ types.Application

func provideBaseApp(inputs baseAppInput) types.AppCreator {
	return func(logger log.Logger, db dbm.DB, tracer io.Writer, appOptions types.AppOptions) types.Application {
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

		baseApp := baseapp.NewBaseApp(string(name), logger, db, txDecoder, inputs.Options...)

		if tracer != nil {
			baseApp.SetCommitMultiStoreTracer(tracer)
		}

		return &app{
			BaseApp:      baseApp,
			moduleBasics: inputs.ModuleBasics,
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

	// Register legacy and grpc-gateway routes for all modules.
	for _, b := range a.moduleBasics {
		b.RegisterRESTRoutes(clientCtx, apiSvr.Router)
	}

	for _, b := range a.moduleBasics {
		b.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	}

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
