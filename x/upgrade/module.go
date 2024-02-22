package upgrade

import (
	"context"
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/x/upgrade/client/cli"
	"cosmossdk.io/x/upgrade/keeper"
	"cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func init() {
	types.RegisterLegacyAminoCodec(codec.NewLegacyAmino())
}

// ConsensusVersion defines the current x/upgrade module consensus version.
const ConsensusVersion uint64 = 3

var (
	_ module.HasName               = AppModule{}
	_ module.HasAminoCodec         = AppModule{}
	_ module.HasGRPCGateway        = AppModule{}
	_ module.HasRegisterInterfaces = AppModule{}
	_ module.HasGenesis            = AppModule{}

	_ appmodule.AppModule     = AppModule{}
	_ appmodule.HasPreBlocker = AppModule{}
	_ appmodule.HasServices   = AppModule{}
	_ appmodule.HasMigrations = AppModule{}
)

// AppModule implements the sdk.AppModule interface
type AppModule struct {
	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper *keeper.Keeper) AppModule {
	return AppModule{
		keeper: keeper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// Name returns the ModuleName
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the upgrade types on the LegacyAmino codec
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the upgrade module.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the CLI transaction commands for this module
func (AppModule) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// RegisterInterfaces registers interfaces and implementations of the upgrade module.
func (AppModule) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(registrar, keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(registrar, am.keeper)

	return nil
}

// RegisterMigrations registers module migrations
func (am AppModule) RegisterMigrations(mr appmodule.MigrationRegistrar) error {
	m := keeper.NewMigrator(am.keeper)
	if err := mr.Register(types.ModuleName, 1, m.Migrate1to2); err != nil {
		return fmt.Errorf("failed to migrate x/%s from version 1 to 2: %w", types.ModuleName, err)
	}

	if err := mr.Register(types.ModuleName, 2, m.Migrate2to3); err != nil {
		return fmt.Errorf("failed to migrate x/%s from version 2 to 3: %w", types.ModuleName, err)
	}

	return nil
}

// DefaultGenesis is an empty object
func (AppModule) DefaultGenesis(_ codec.JSONCodec) json.RawMessage {
	return []byte("{}")
}

// ValidateGenesis is always successful, as we ignore the value
func (AppModule) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, _ json.RawMessage) error {
	return nil
}

// InitGenesis is ignored, no sense in serializing future upgrades
func (am AppModule) InitGenesis(ctx context.Context, _ codec.JSONCodec, _ json.RawMessage) {
	// set version map automatically if available
	if versionMap := am.keeper.GetInitVersionMap(); versionMap != nil {
		// chains can still use a custom init chainer for setting the version map
		// this means that we need to combine the manually wired modules version map with app wiring enabled modules version map
		moduleVM, err := am.keeper.GetModuleVersionMap(ctx)
		if err != nil {
			panic(err)
		}

		for name, version := range moduleVM {
			if _, ok := versionMap[name]; !ok {
				versionMap[name] = version
			}
		}

		err = am.keeper.SetModuleVersionMap(ctx, versionMap)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis is always empty, as InitGenesis does nothing either
func (am AppModule) ExportGenesis(_ context.Context, cdc codec.JSONCodec) json.RawMessage {
	return am.DefaultGenesis(cdc)
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// PreBlock calls the upgrade module hooks
//
// CONTRACT: this is called *before* all other modules' BeginBlock functions
func (am AppModule) PreBlock(ctx context.Context) (appmodule.ResponsePreBlock, error) {
	return am.keeper.PreBlocker(ctx)
}
