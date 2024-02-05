package vesting

import (
	"context"
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"
	v1 "cosmossdk.io/x/accounts/v1"
	"cosmossdk.io/x/accounts/vesting/cli"
	"cosmossdk.io/x/accounts/vesting/keeper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

const (
	ModuleName = "vesting"
	StoreKey   = ModuleName
)

// ModuleAccountAddress defines the x/accounts module address.
var ModuleAccountAddress = address.Module(ModuleName)

const (
	ConsensusVersion = 1
)

var (
	_ appmodule.AppModule        = AppModule{}
	_ module.HasName             = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
)

func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{k: k}
}

type AppModule struct {
	k keeper.Keeper
}

func (m AppModule) IsAppModule() {}

func (m AppModule) IsOnePerModuleType() {}

func (m AppModule) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

func (m AppModule) RegisterInterfaces(registry types.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, v1.MsgServiceDesc())
}

func (m AppModule) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

// App module services

func (m AppModule) RegisterServices(configurator module.Configurator) {
	// v1.RegisterQueryServer(configurator.QueryServer(), NewQueryServer(m.k))
	// v1.RegisterMsgServer(configurator.MsgServer(), NewMsgServer(m.k))
}

// App module genesis

func (AppModule) DefaultGenesis(jsonCodec codec.JSONCodec) json.RawMessage {
	return jsonCodec.MustMarshalJSON(&v1.GenesisState{})
}

func (AppModule) ValidateGenesis(jsonCodec codec.JSONCodec, config client.TxEncodingConfig, message json.RawMessage) error {
	return nil
}

func (m AppModule) InitGenesis(ctx context.Context, jsonCodec codec.JSONCodec, message json.RawMessage) {

}

func (m AppModule) ExportGenesis(ctx context.Context, jsonCodec codec.JSONCodec) json.RawMessage {
	return nil
}

func (AppModule) Name() string { return ModuleName }

func (AppModule) GetTxCmd() *cobra.Command {
	return cli.TxCmd(ModuleName)
}

func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }
