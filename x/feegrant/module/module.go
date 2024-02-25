package module

import (
	"context"
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/errors"
	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/feegrant/client/cli"
	"cosmossdk.io/x/feegrant/keeper"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var (
	_ module.HasName               = AppModule{}
	_ module.HasAminoCodec         = AppModule{}
	_ module.HasGRPCGateway        = AppModule{}
	_ module.HasRegisterInterfaces = AppModule{}
	_ module.AppModuleSimulation   = AppModule{}
	_ module.HasGenesis            = AppModule{}

	_ appmodule.AppModule     = AppModule{}
	_ appmodule.HasEndBlocker = AppModule{}
	_ appmodule.HasServices   = AppModule{}
	_ appmodule.HasMigrations = AppModule{}
)

// AppModule implements an application module for the feegrant module.
type AppModule struct {
	cdc      codec.Codec
	registry cdctypes.InterfaceRegistry

	keeper        keeper.Keeper
	accountKeeper feegrant.AccountKeeper
	bankKeeper    feegrant.BankKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, ak feegrant.AccountKeeper, bk feegrant.BankKeeper, keeper keeper.Keeper, registry cdctypes.InterfaceRegistry) AppModule {
	return AppModule{
		cdc:           cdc,
		keeper:        keeper,
		accountKeeper: ak,
		bankKeeper:    bk,
		registry:      registry,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// Name returns the feegrant module's name.
func (AppModule) Name() string {
	return feegrant.ModuleName
}

// RegisterLegacyAminoCodec registers the feegrant module's types for the given codec.
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	feegrant.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the feegrant module's interface types
func (AppModule) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	feegrant.RegisterInterfaces(registry)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the feegrant module.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *gwruntime.ServeMux) {
	if err := feegrant.RegisterQueryHandlerClient(context.Background(), mux, feegrant.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the feegrant module.
func (AppModule) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	feegrant.RegisterMsgServer(registrar, keeper.NewMsgServerImpl(am.keeper))
	feegrant.RegisterQueryServer(registrar, am.keeper)

	return nil
}

// RegisterMigrations registers module migrations.
func (am AppModule) RegisterMigrations(mr appmodule.MigrationRegistrar) error {
	m := keeper.NewMigrator(am.keeper)

	if err := mr.Register(feegrant.ModuleName, 1, m.Migrate1to2); err != nil {
		return fmt.Errorf("failed to migrate x/feegrant from version 1 to 2: %w", err)
	}

	return nil
}

// DefaultGenesis returns default genesis state as raw bytes for the feegrant module.
func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(feegrant.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the feegrant module.
func (AppModule) ValidateGenesis(cdc codec.JSONCodec, config sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data feegrant.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return errors.Wrapf(err, "failed to unmarshal %s genesis state", feegrant.ModuleName)
	}

	return feegrant.ValidateGenesis(data)
}

// InitGenesis performs genesis initialization for the feegrant module.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	var gs feegrant.GenesisState
	cdc.MustUnmarshalJSON(bz, &gs)

	err := am.keeper.InitGenesis(ctx, &gs)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the feegrant module.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}

	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 2 }

// EndBlock returns the end blocker for the feegrant module.
func (am AppModule) EndBlock(ctx context.Context) error {
	return EndBlocker(ctx, am.keeper)
}
