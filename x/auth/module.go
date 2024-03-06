package auth

import (
	"context"
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/x/auth/keeper"
	"cosmossdk.io/x/auth/simulation"
	"cosmossdk.io/x/auth/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// ConsensusVersion defines the current x/auth module consensus version.
const (
	ConsensusVersion = 5
	GovModuleName    = "gov"
)

var (
	_ module.AppModuleSimulation = AppModule{}
	_ module.HasName             = AppModule{}
	_ module.HasGenesis          = AppModule{}

	_ appmodule.AppModule     = AppModule{}
	_ appmodule.HasServices   = AppModule{}
	_ appmodule.HasMigrations = AppModule{}
)

// AppModule implements an application module for the auth module.
type AppModule struct {
	accountKeeper     keeper.AccountKeeper
	randGenAccountsFn types.RandomGenesisAccountsFn
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, accountKeeper keeper.AccountKeeper, randGenAccountsFn types.RandomGenesisAccountsFn) AppModule {
	return AppModule{
		accountKeeper:     accountKeeper,
		randGenAccountsFn: randGenAccountsFn,
	}
}

// Name returns the auth module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the auth module's types for the given codec.
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the auth module.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers interfaces and implementations of the auth module.
func (AppModule) RegisterInterfaces(registry registry.LegacyRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(registrar, keeper.NewMsgServerImpl(am.accountKeeper))
	types.RegisterQueryServer(registrar, keeper.NewQueryServer(am.accountKeeper))

	return nil
}

// RegisterMigrations registers module migrations
func (am AppModule) RegisterMigrations(mr appmodule.MigrationRegistrar) error {
	m := keeper.NewMigrator(am.accountKeeper)
	if err := mr.Register(types.ModuleName, 1, m.Migrate1to2); err != nil {
		return fmt.Errorf("failed to migrate x/%s from version 1 to 2: %w", types.ModuleName, err)
	}

	if err := mr.Register(types.ModuleName, 2, m.Migrate2to3); err != nil {
		return fmt.Errorf("failed to migrate x/%s from version 2 to 3: %w", types.ModuleName, err)
	}
	if err := mr.Register(types.ModuleName, 3, m.Migrate3to4); err != nil {
		return fmt.Errorf("failed to migrate x/%s from version 3 to 4: %w", types.ModuleName, err)
	}
	if err := mr.Register(types.ModuleName, 4, m.Migrate4To5); err != nil {
		return fmt.Errorf("failed to migrate x/%s from version 4 to 5: %w", types.ModuleName, err)
	}

	return nil
}

// DefaultGenesis returns default genesis state as raw bytes for the auth module.
func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the auth module.
func (AppModule) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return types.ValidateGenesis(data)
}

// InitGenesis performs genesis initialization for the auth module.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	err := am.accountKeeper.InitGenesis(ctx, genesisState)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the auth
// module.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	gs, err := am.accountKeeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the auth module
func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState, am.randGenAccountsFn)
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs()
}

// RegisterStoreDecoder registers a decoder for auth module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[types.StoreKey] = simtypes.NewStoreDecoderFuncFromCollectionsSchema(am.accountKeeper.Schema)
}

// WeightedOperations doesn't return any auth module operation.
func (AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
