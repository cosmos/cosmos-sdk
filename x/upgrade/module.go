package upgrade

import (
	"context"
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
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
	_ module.AppModuleBasic = AppModule{}
	_ module.HasGenesis     = AppModule{}
	_ module.HasServices    = AppModule{}

	_ appmodule.AppModule     = AppModule{}
	_ appmodule.HasPreBlocker = AppModule{}
)

// AppModuleBasic implements the sdk.AppModuleBasic interface
type AppModuleBasic struct{}

// Name returns the ModuleName
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the upgrade types on the LegacyAmino codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the upgrade module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the CLI transaction commands for this module
func (ab AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// RegisterInterfaces registers interfaces and implementations of the upgrade module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// AppModule implements the sdk.AppModule interface
type AppModule struct {
	AppModuleBasic
	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper *keeper.Keeper, ac address.Codec) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := keeper.NewMigrator(am.keeper)
	err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2)
	if err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", types.ModuleName, err))
	}
	err = cfg.RegisterMigration(types.ModuleName, 2, m.Migrate2to3)
	if err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 2 to 3: %v", types.ModuleName, err))
	}
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

// DefaultGenesis is an empty object
func (AppModuleBasic) DefaultGenesis(_ codec.JSONCodec) json.RawMessage {
	return []byte("{}")
}

// ValidateGenesis is always successful, as we ignore the value
func (AppModuleBasic) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, _ json.RawMessage) error {
	return nil
}

// ExportGenesis is always empty, as InitGenesis does nothing either
func (am AppModule) ExportGenesis(_ context.Context, cdc codec.JSONCodec) json.RawMessage {
	return am.DefaultGenesis(cdc)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// PreBlock calls the upgrade module hooks
//
// CONTRACT: this is called *before* all other modules' BeginBlock functions
func (am AppModule) PreBlock(ctx context.Context) (appmodule.ResponsePreBlock, error) {
	return PreBlocker(ctx, am.keeper)
}
