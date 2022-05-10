package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/gogo/protobuf/grpc"
	abci "github.com/tendermint/tendermint/abci/types"

	runtimev1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/app/runtime/v1alpha1"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// App is a wrapper around BaseApp and ModuleManager that can be used in hybrid
// app.go/app config scenarios or directly as a servertypes.Application instance.
type App struct {
	*baseapp.BaseApp
	ModuleManager       *module.Manager
	config              *runtimev1alpha1.Module
	privateState        *privateState
	beginBlockers       []func(sdk.Context, abci.RequestBeginBlock)
	endBlockers         []func(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
	baseAppOptions      []BaseAppOption
	txHandler           tx.Handler
	msgServiceRegistrar grpc.Server
}

// RegisterModules registers the provided modules with the module manager and
// the basic module manager. This is the primary hook for integrating with
// modules which are not registered using the app config.
func (a *App) RegisterModules(modules ...module.AppModule) error {
	for _, appModule := range modules {
		name := appModule.Name()
		if _, ok := a.ModuleManager.Modules[name]; ok {
			return fmt.Errorf("AppModule named %q already exists", name)
		}
		a.ModuleManager.Modules[name] = appModule

		if _, ok := a.privateState.basicManager[name]; ok {
			return fmt.Errorf("AppModuleBasic named %q already exists", name)
		}

		a.privateState.basicManager[name] = appModule
		appModule.RegisterInterfaces(a.privateState.interfaceRegistry)
		appModule.RegisterLegacyAminoCodec(a.privateState.amino)

	}
	return nil
}

// Load finishes all initialization operations and loads the app.
func (a *App) Load(loadLatest bool) error {
	configurator := module.NewConfigurator(a.privateState.cdc, a.msgServiceRegistrar, a.GRPCQueryRouter())
	a.ModuleManager.RegisterServices(configurator)

	if len(a.config.InitGenesis) != 0 {
		a.ModuleManager.SetOrderInitGenesis(a.config.InitGenesis...)
		a.SetInitChainer(a.InitChainer)
	}

	if len(a.config.BeginBlockers) != 0 {
		a.ModuleManager.SetOrderBeginBlockers(a.config.BeginBlockers...)
		a.SetBeginBlocker(a.ModuleManager.BeginBlock)
	}

	if len(a.config.EndBlockers) != 0 {
		a.ModuleManager.SetOrderEndBlockers(a.config.EndBlockers...)
		a.SetEndBlocker(a.ModuleManager.EndBlock)
	}

	if loadLatest {
		if err := a.LoadLatestVersion(); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState map[string]json.RawMessage
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	// TODO: app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return a.ModuleManager.InitGenesis(ctx, a.privateState.cdc, genesisState)
}

func (a *App) ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs []string) (servertypes.ExportedApp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *App) SimulationManager() *module.SimulationManager {
	//TODO implement me
	panic("implement me")
}

var _ simappLikeApp = &App{}

func (a *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	basics := module.BasicManager{}

	for name, mod := range a.ModuleManager.Modules {
		basics[name] = mod
	}

	basics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

func (a *App) RegisterTxService(clientCtx client.Context) {}

func (a *App) RegisterTendermintService(clientCtx client.Context) {}

var _ servertypes.Application = &App{}

type simappLikeApp interface {
	// Exports the state of the application for a genesis file.
	ExportAppStateAndValidators(
		forZeroHeight bool, jailAllowedAddrs []string,
	) (servertypes.ExportedApp, error)

	// Helper for the simulation framework.
	SimulationManager() *module.SimulationManager
}
