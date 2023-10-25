package accounts

import (
	"context"
	"encoding/json"

	"cosmossdk.io/x/accounts/cli"
	v1 "cosmossdk.io/x/accounts/v1"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
)

const ModuleName = "accounts"

var (
	_ module.HasName     = AppModule{}
	_ module.HasGenesis  = AppModule{}
	_ module.HasServices = AppModule{}
)

func NewAppModule(k Keeper) AppModule {
	return AppModule{k: k}
}

type AppModule struct {
	k Keeper
}

func (m AppModule) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

func (m AppModule) RegisterInterfaces(registry types.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, v1.MsgServiceDesc())
}

func (m AppModule) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

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

func (AppModule) Name() string { return ModuleName }

func (AppModule) GetTxCmd() *cobra.Command {
	cmds := cli.GeTxCmds()
	cmd := &cobra.Command{
		Use:                        ModuleName,
		Short:                      "accounts module tx commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(cmds...)
	return cmd
}

func (AppModule) GetQueryCmd() *cobra.Command {
	cmds := cli.GetQueryCmds()
	cmd := &cobra.Command{
		Use:                        ModuleName,
		Short:                      "accounts module query commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(cmds...)
	return cmd
}
