package module

import (
	"encoding/json"

	"cosmossdk.io/x/accounts"
	v1 "cosmossdk.io/x/accounts/v1"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

const Name = "accounts"

var (
	_ module.AppModuleBasic = AppModule{}
	_ module.HasGenesis     = AppModule{}
	_ module.HasServices    = AppModule{}
)

func NewAppModule(k accounts.Keeper) AppModule {
	return AppModule{k: k}
}

type AppModule struct {
	k accounts.Keeper
}

// App module services

func (m AppModule) RegisterServices(configurator module.Configurator) {
	v1.RegisterQueryServer(configurator.QueryServer(), accounts.NewQueryServer(m.k))
	v1.RegisterMsgServer(configurator.MsgServer(), accounts.NewMsgServer(m.k))
}

// App module genesis

func (AppModule) DefaultGenesis(jsonCodec codec.JSONCodec) json.RawMessage {
	return jsonCodec.MustMarshalJSON(&v1.GenesisState{})
}

func (AppModule) ValidateGenesis(jsonCodec codec.JSONCodec, config client.TxEncodingConfig, message json.RawMessage) error {
	return nil
}

func (m AppModule) InitGenesis(context sdk.Context, jsonCodec codec.JSONCodec, message json.RawMessage) {
	gs := &v1.GenesisState{}
	jsonCodec.MustUnmarshalJSON(message, gs)
	err := m.k.ImportState(context, gs)
	if err != nil {
		panic(err)
	}
}

func (m AppModule) ExportGenesis(context sdk.Context, jsonCodec codec.JSONCodec) json.RawMessage {
	gs, err := m.k.ExportState(context)
	if err != nil {
		panic(err)
	}
	return jsonCodec.MustMarshalJSON(gs)
}

// App module basic

func (AppModule) Name() string { return Name }

func (AppModule) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

func (AppModule) RegisterInterfaces(_ types.InterfaceRegistry) {}

func (AppModule) RegisterGRPCGatewayRoutes(context client.Context, mux *runtime.ServeMux) {}
