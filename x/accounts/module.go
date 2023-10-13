package accounts

import (
	"context"
	"encoding/json"

	v1 "cosmossdk.io/x/accounts/v1"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

const Name = "accounts"

var (
	_ module.AppModuleBasic = AppModule{}
	_ module.HasGenesis     = AppModule{}
	_ module.HasServices    = AppModule{}
)

func NewAppModule(k Keeper) AppModule {
	return AppModule{k: k}
}

type AppModule struct {
	k Keeper
}

// App module services

func (m AppModule) RegisterServices(configurator module.Configurator) {
	v1.RegisterQueryServer(configurator.QueryServer(), NewQueryServer(m.k))
	v1.RegisterMsgServer(configurator.MsgServer(), NewMsgServer(m.k))
}

// App module genesis

func (AppModule) DefaultGenesis(jsonCodec codec.JSONCodec) json.RawMessage {
	return jsonCodec.MustMarshalJSON(&v1.GenesisState{})
}

func (AppModule) ValidateGenesis(jsonCodec codec.JSONCodec, config client.TxEncodingConfig, message json.RawMessage) error {
	return nil
}

func (m AppModule) InitGenesis(ctx context.Context, jsonCodec codec.JSONCodec, message json.RawMessage) {
	gs := &v1.GenesisState{}
	jsonCodec.MustUnmarshalJSON(message, gs)
	err := m.k.ImportState(ctx, gs)
	if err != nil {
		panic(err)
	}
}

func (m AppModule) ExportGenesis(ctx context.Context, jsonCodec codec.JSONCodec) json.RawMessage {
	gs, err := m.k.ExportState(ctx)
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
