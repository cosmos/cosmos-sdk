package compat

import (
	"context"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/app/cli"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func RegisterAppModuleBasic(configurator cli.Configurator, module module.AppModuleBasic) {
	configurator.SetTxCommand(module.GetTxCmd())
	configurator.SetQueryCommand(module.GetQueryCmd())
	configurator.SetGenesisHandler(module)
}

func RegisterAppModule(configurator app.Configurator, module module.AppModule) {
	module.RegisterServices(configurator)
	configurator.RegisterBeginBlocker(func(ctx context.Context, req abci.RequestBeginBlock) {
		module.BeginBlock(types.UnwrapSDKContext(ctx), req)
	})
	configurator.RegisterEndBlocker(func(ctx context.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
		return module.EndBlock(types.UnwrapSDKContext(ctx), req)
	})
	configurator.RegisterGenesisHandler(genesisWrapper{module})
}

type genesisWrapper struct {
	module.AppModule
}

func (g genesisWrapper) InitGenesis(ctx context.Context, codec codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	return g.AppModule.InitGenesis(types.UnwrapSDKContext(ctx), codec, message)
}

func (g genesisWrapper) ExportGenesis(ctx context.Context, codec codec.JSONCodec) json.RawMessage {
	return g.AppModule.ExportGenesis(types.UnwrapSDKContext(ctx), codec)
}

var _ app.GenesisHandler = genesisWrapper{}
