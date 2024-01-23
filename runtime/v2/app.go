package runtime

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"golang.org/x/exp/slices"

	runtimev2 "cosmossdk.io/api/cosmos/app/runtime/v2"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
	coreappmanager "cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/mempool"
	"cosmossdk.io/server/v2/core/store"
	servertx "cosmossdk.io/server/v2/core/transaction"
	"cosmossdk.io/server/v2/stf"
	storetypes "cosmossdk.io/store/types"
	authtx "cosmossdk.io/x/auth/tx"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// App is a wrapper around AppManager and ModuleManager that can be used in hybrid
// app.go/app config scenarios or directly as a servertypes.Application instance.
// To get an instance of *App, *AppBuilder must be requested as a dependency
// in a container which declares the runtime module and the AppBuilder.Build()
// method must be called.
//
// App can be used to create a hybrid app.go setup where some configuration is
// done declaratively with an app config and the rest of it is done the old way.
// See simapp/app_.go for an example of this setup.
type App struct {
	*appmanager.AppManager[servertx.Tx]

	// app manager dependencies
	stf                 *stf.STF[servertx.Tx]
	msgRouterBuilder    *stf.MsgRouterBuilder
	mempool             mempool.Mempool[servertx.Tx]
	prepareBlockHandler coreappmanager.PrepareHandler[servertx.Tx]
	verifyBlockHandler  coreappmanager.ProcessHandler[servertx.Tx]
	db                  store.Store

	// app configuration
	logger    log.Logger
	config    *runtimev2.Module
	appConfig *appv1alpha1.Config

	// modules configuration
	configurator      module.Configurator
	storeKeys         []storetypes.StoreKey
	interfaceRegistry codectypes.InterfaceRegistry
	cdc               codec.Codec
	amino             *codec.LegacyAmino
	moduleManager     *MMv2
	basicManager      module.BasicManager

	// Unknowns
	// initChainer is the init chainer function defined by the app config.
	// this is only required if the chain wants to add special InitChainer logic.
	initChainer sdk.InitChainer
}

// RegisterStores registers the provided store keys.
// This method should only be used for registering extra stores
// wiich is necessary for modules that not registered using the app config.
// To be used in combination of RegisterModules.
func (a *App) RegisterStores(keys ...storetypes.StoreKey) error {
	a.storeKeys = append(a.storeKeys, keys...)
	// a.MountStores(keys...)

	return nil
}

// Load finishes all initialization operations and loads the app.
func (a *App) Load() error {
	// set defaults
	if a.mempool == nil {
		// a.mempool = mempool.NewBasicMempool()
	}

	if a.prepareBlockHandler == nil {
		// a.prepareBlockHandler = appmanager.DefaultPrepareBlockHandler
	}

	if a.verifyBlockHandler == nil {
		// a.verifyBlockHandler = appmanager.DefaultProcessBlockHandler
	}

	appManagerBuilder := appmanager.Builder[servertx.Tx]{
		STF:                 a.stf,
		DB:                  a.db,
		ValidateTxGasLimit:  a.config.GasConfig.ValidateTxGasLimit,
		QueryGasLimit:       a.config.GasConfig.QueryGasLimit,
		SimulationGasLimit:  a.config.GasConfig.SimulationGasLimit,
		PrepareBlockHandler: a.prepareBlockHandler,
		VerifyBlockHandler:  a.verifyBlockHandler,
	}

	appManager, err := appManagerBuilder.Build()
	if err != nil {
		return err
	}

	a.AppManager = appManager

	return nil
}

// InitChainer initializes the chain.
func (a *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState map[string]json.RawMessage
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	return a.moduleManager.InitGenesis(ctx, a.cdc, genesisState)
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (a *App) RegisterAPIRoutes(apiSvr *api.Server, _ config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new CometBFT queries routes from grpc-gateway. // TODO probably server should do that while taking an app.
	// cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	a.basicManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (a *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(a.GRPCQueryRouter(), clientCtx, a.Simulate, a.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (a *App) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(a)
	cmtservice.RegisterTendermintService(
		clientCtx,
		a.GRPCQueryRouter(),
		a.interfaceRegistry,
		cmtApp.Query,
	)
}

// RegisterNodeService registers the node gRPC service on the app gRPC router.
func (a *App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, a.GRPCQueryRouter(), cfg)
}

// Configurator returns the app's configurator.
func (a *App) Configurator() module.Configurator {
	return a.configurator
}

// LoadHeight loads a particular height
func (a *App) LoadHeight(height int64) error {
	return nil
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (a *App) DefaultGenesis() map[string]json.RawMessage {
	return a.basicManager.DefaultGenesis(a.cdc)
}

// GetStoreKeys returns all the stored store keys.
func (a *App) GetStoreKeys() []storetypes.StoreKey {
	return a.storeKeys
}

// UnsafeFindStoreKey fetches a registered StoreKey from the App in linear time.
//
// NOTE: This should only be used in testing.
func (a *App) UnsafeFindStoreKey(storeKey string) storetypes.StoreKey {
	i := slices.IndexFunc(a.storeKeys, func(s storetypes.StoreKey) bool { return s.Name() == storeKey })
	if i == -1 {
		return nil
	}

	return a.storeKeys[i]
}

var _ servertypes.Application = &App{}
