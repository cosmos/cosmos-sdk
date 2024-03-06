package evidence

import (
	"context"
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"
	eviclient "cosmossdk.io/x/evidence/client"
	"cosmossdk.io/x/evidence/client/cli"
	"cosmossdk.io/x/evidence/keeper"
	"cosmossdk.io/x/evidence/simulation"
	"cosmossdk.io/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

var (
	_ module.HasName                  = AppModule{}
	_ module.HasAminoCodec            = AppModule{}
	_ module.HasGRPCGateway           = AppModule{}
	_ appmodule.HasRegisterInterfaces = AppModule{}
	_ module.AppModuleSimulation      = AppModule{}
	_ module.HasGenesis               = AppModule{}

	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
)

const ConsensusVersion = 1

// AppModule implements the AppModule interface for the evidence module.
type AppModule struct {
	evidenceHandlers []eviclient.EvidenceHandler
	keeper           keeper.Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(keeper keeper.Keeper, evidenceHandlers ...eviclient.EvidenceHandler) AppModule {
	return AppModule{
		keeper:           keeper,
		evidenceHandlers: evidenceHandlers,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Name returns the evidence module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the evidence module's types to the LegacyAmino codec.
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the evidence module.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the evidence module's root tx command.
func (am AppModule) GetTxCmd() *cobra.Command {
	evidenceCLIHandlers := make([]*cobra.Command, len(am.evidenceHandlers))
	for i, evidenceHandler := range am.evidenceHandlers {
		evidenceCLIHandlers[i] = evidenceHandler.CLIHandler()
	}

	return cli.GetTxCmd(evidenceCLIHandlers)
}

// RegisterInterfaces registers the evidence module's interface types
func (AppModule) RegisterInterfaces(registry registry.LegacyRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(registrar, keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(registrar, keeper.NewQuerier(&am.keeper))
	return nil
}

// DefaultGenesis returns the evidence module's default genesis state.
func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the evidence module.
func (AppModule) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return gs.Validate()
}

// InitGenesis performs the evidence module's genesis initialization
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	var gs types.GenesisState
	err := cdc.UnmarshalJSON(bz, &gs)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal %s genesis state: %s", types.ModuleName, err))
	}

	InitGenesis(ctx, am.keeper, &gs)
}

// ExportGenesis returns the evidence module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(ExportGenesis(ctx, am.keeper))
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// BeginBlock executes all ABCI BeginBlock logic respective to the evidence module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	return am.keeper.BeginBlocker(ctx)
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the evidence module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for evidence module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[types.StoreKey] = simtypes.NewStoreDecoderFuncFromCollectionsSchema(am.keeper.Schema)
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
