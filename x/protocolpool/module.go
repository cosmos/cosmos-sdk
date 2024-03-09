package protocolpool

import (
	"context"
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/x/protocolpool/keeper"
	"cosmossdk.io/x/protocolpool/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// ConsensusVersion defines the current x/protocolpool module consensus version.
const ConsensusVersion = 1

var (
	_ module.HasName                  = AppModule{}
	_ module.HasAminoCodec            = AppModule{}
	_ module.HasGRPCGateway           = AppModule{}
	_ appmodule.HasRegisterInterfaces = AppModule{}
	_ module.AppModule                = AppModule{}
	_ module.AppModuleSimulation      = AppModule{}
	_ module.HasGenesis               = AppModule{}

	_ appmodule.AppModule   = AppModule{}
	_ appmodule.HasServices = AppModule{}
)

// AppModule implements an application module for the pool module
type AppModule struct {
	cdc           codec.Codec
	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper,
	accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper,
) AppModule {
	return AppModule{
		cdc:           cdc,
		keeper:        keeper,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// Name returns the pool module's name.
func (AppModule) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the pool module's types on the LegacyAmino codec.
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers interfaces and implementations of the bank module.
func (AppModule) RegisterInterfaces(registry registry.LegacyRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(registrar, keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(registrar, keeper.NewQuerier(am.keeper))

	return nil
}

// DefaultGenesis returns default genesis state as raw bytes for the protocolpool module.
func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the protocolpool module.
func (AppModule) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return types.ValidateGenesis(&data)
}

// InitGenesis performs genesis initialization for the protocolpool module.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	if err := am.keeper.InitGenesis(ctx, &genesisState); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the protocolpool module.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }
