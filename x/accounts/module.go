package accounts

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/x/accounts/cli"
	v1 "cosmossdk.io/x/accounts/v1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

const (
	ModuleName = "accounts"
	StoreKey   = "_" + ModuleName // unfortunately accounts collides with auth store key
)

// ModuleAccountAddress defines the x/accounts module address.
var ModuleAccountAddress = address.Module(ModuleName)

const (
	ConsensusVersion = 1
)

var (
	_ appmodule.AppModule   = AppModule{}
	_ appmodule.HasServices = AppModule{}

	_ module.HasName             = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
)

func NewAppModule(k Keeper) AppModule {
	return AppModule{k: k}
}

type AppModule struct {
	k Keeper
}

func (m AppModule) IsAppModule() {}

func (AppModule) Name() string { return ModuleName }

func (m AppModule) RegisterInterfaces(registry registry.LegacyRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, v1.MsgServiceDesc())
}

// App module services

func (m AppModule) RegisterServices(registar grpc.ServiceRegistrar) error {
	v1.RegisterQueryServer(registar, NewQueryServer(m.k))
	v1.RegisterMsgServer(registar, NewMsgServer(m.k))

	return nil
}

// App module genesis

func (AppModule) DefaultGenesis(jsonCodec codec.JSONCodec) json.RawMessage {
	return jsonCodec.MustMarshalJSON(&v1.GenesisState{})
}

func (AppModule) ValidateGenesis(jsonCodec codec.JSONCodec, config client.TxEncodingConfig, message json.RawMessage) error {
	gs := &v1.GenesisState{}
	if err := jsonCodec.UnmarshalJSON(message, gs); err != nil {
		return err
	}
	// Add validation logic for gs here
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

func (AppModule) GetTxCmd() *cobra.Command {
	return cli.TxCmd(ModuleName)
}

func (AppModule) GetQueryCmd() *cobra.Command {
	return cli.QueryCmd(ModuleName)
}

func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }
