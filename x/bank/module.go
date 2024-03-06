package bank

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/x/bank/client/cli"
	"cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/simulation"
	"cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// ConsensusVersion defines the current x/bank module consensus version.
const ConsensusVersion = 4

var (
	_ module.HasName                  = AppModule{}
	_ module.HasAminoCodec            = AppModule{}
	_ module.HasGRPCGateway           = AppModule{}
	_ appmodule.HasRegisterInterfaces = AppModule{}
	_ module.AppModuleSimulation      = AppModule{}
	_ module.HasGenesis               = AppModule{}
	_ module.HasInvariants            = AppModule{}

	_ appmodule.AppModule     = AppModule{}
	_ appmodule.HasServices   = AppModule{}
	_ appmodule.HasMigrations = AppModule{}
)

// AppModule implements an application module for the bank module.
type AppModule struct {
	cdc           codec.Codec
	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, accountKeeper types.AccountKeeper) AppModule {
	return AppModule{
		cdc:           cdc,
		keeper:        keeper,
		accountKeeper: accountKeeper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Name returns the bank module's name.
func (AppModule) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the bank module's types on the LegacyAmino codec.
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the bank module.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the bank module.
func (AppModule) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// RegisterInterfaces registers interfaces and implementations of the bank module.
func (AppModule) RegisterInterfaces(registry registry.LegacyRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(registrar, keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(registrar, am.keeper)

	return nil
}

// RegisterMigrations registers the bank module's migrations.
func (am AppModule) RegisterMigrations(mr appmodule.MigrationRegistrar) error {
	m := keeper.NewMigrator(am.keeper.(keeper.BaseKeeper))

	if err := mr.Register(types.ModuleName, 1, m.Migrate1to2); err != nil {
		return fmt.Errorf("failed to migrate x/bank from version 1 to 2: %w", err)
	}

	if err := mr.Register(types.ModuleName, 2, m.Migrate2to3); err != nil {
		return fmt.Errorf("failed to migrate x/bank from version 2 to 3: %w", err)
	}

	if err := mr.Register(types.ModuleName, 3, m.Migrate3to4); err != nil {
		return fmt.Errorf("failed to migrate x/bank from version 3 to 4: %w", err)
	}

	return nil
}

// RegisterInvariants registers the bank module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// DefaultGenesis returns default genesis state as raw bytes for the bank module.
func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the bank module.
func (AppModule) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return data.Validate()
}

// InitGenesis performs genesis initialization for the bank module.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, data json.RawMessage) {
	start := time.Now()
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	telemetry.MeasureSince(start, "InitGenesis", "crisis", "unmarshal")

	am.keeper.InitGenesis(ctx, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the bank
// module.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the bank module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs()
}

// RegisterStoreDecoder registers a decoder for supply module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[types.StoreKey] = simtypes.NewStoreDecoderFuncFromCollectionsSchema(am.keeper.(keeper.BaseKeeper).Schema)
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc, simState.TxConfig, am.accountKeeper, am.keeper,
	)
}
