package protocolpool

import (
	"context"
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/simulation"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

// ConsensusVersion defines the current x/protocolpool module consensus version.
const ConsensusVersion = 1

var (
	_ module.AppModuleSimulation = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}
	_ module.AppModule           = AppModule{}

	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
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

// Name returns the protocolpool module's name.
// Deprecated: kept for legacy reasons.
func (AppModule) Name() string { return types.ModuleName }

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers interfaces and implementations of the protocolpool module.
func (AppModule) RegisterInterfaces(ir codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(ir)
}

func (am AppModule) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(amino)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(configurator module.Configurator) {
	types.RegisterMsgServer(configurator, keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(configurator, keeper.NewQuerier(am.keeper))
}

// DefaultGenesis returns default genesis state as raw bytes for the protocolpool module.
func (am AppModule) DefaultGenesis(_ codec.JSONCodec) json.RawMessage {
	data, err := am.cdc.MarshalJSON(types.DefaultGenesisState())
	if err != nil {
		panic(err)
	}
	return data
}

// ValidateGenesis performs genesis state validation for the protocolpool module.
func (am AppModule) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := am.cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return data.Validate()
}

// InitGenesis performs genesis initialization for the protocolpool module.
func (am AppModule) InitGenesis(ctx sdk.Context, _ codec.JSONCodec, data json.RawMessage) {
	var genesisState types.GenesisState
	am.cdc.MustUnmarshalJSON(data, &genesisState)

	if err := am.keeper.InitGenesis(ctx, &genesisState); err != nil {
		panic(fmt.Errorf("failed to init genesis state: %w", err))
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the protocolpool module.
func (am AppModule) ExportGenesis(ctx sdk.Context, _ codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to export genesis state: %w", err))
	}
	return am.cdc.MustMarshalJSON(gs)
}

// BeginBlock implements appmodule.HasBeginBlocker.
func (am AppModule) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return am.keeper.BeginBlocker(sdkCtx)
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// ____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the protocolpool module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	// GenerateGenesisState creates a randomized GenState of the protocolpool module.
	simulation.RandomizedGenState(simState)
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (AppModule) ProposalMsgs(_ module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs()
}

// RegisterStoreDecoder registers a decoder for protocolpool module's types
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {
}

func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams,
		simState.TxConfig,
		am.accountKeeper,
		am.bankKeeper,
		am.keeper,
	)
}
