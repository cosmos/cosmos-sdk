package runtime

import (
	"encoding/json"

	"github.com/gogo/protobuf/grpc"
	abci "github.com/tendermint/tendermint/abci/types"

	runtimev1 "github.com/cosmos/cosmos-sdk/api/cosmos/base/runtime/v1"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type App struct {
	*baseapp.BaseApp
	config              *runtimev1.Module
	builder             *appBuilder
	mm                  *module.Manager
	beginBlockers       []func(sdk.Context, abci.RequestBeginBlock)
	endBlockers         []func(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
	baseAppOptions      []BaseAppOption
	txHandler           tx.Handler
	msgServiceRegistrar grpc.Server
}

func (a App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState map[string]json.RawMessage
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	// TODO: app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return a.mm.InitGenesis(ctx, a.builder.cdc, genesisState)
}

func (a App) ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs []string) (servertypes.ExportedApp, error) {
	//TODO implement me
	panic("implement me")
}

func (a App) SimulationManager() *module.SimulationManager {
	//TODO implement me
	panic("implement me")
}

var _ SimappLikeApp = &App{}

func (a App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	basics := module.BasicManager{}

	for name, mod := range a.mm.Modules {
		basics[name] = mod
	}

	basics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

func (a App) RegisterTxService(clientCtx client.Context) {}

func (a App) RegisterTendermintService(clientCtx client.Context) {}

var _ servertypes.Application = &App{}

type SimappLikeApp interface {
	// Exports the state of the application for a genesis file.
	ExportAppStateAndValidators(
		forZeroHeight bool, jailAllowedAddrs []string,
	) (servertypes.ExportedApp, error)

	// Helper for the simulation framework.
	SimulationManager() *module.SimulationManager
}
