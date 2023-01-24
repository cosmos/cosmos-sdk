package module

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/quarantine/client/cli"
	"github.com/cosmos/cosmos-sdk/x/quarantine/keeper"
	"github.com/cosmos/cosmos-sdk/x/quarantine/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

type AppModule struct {
	AppModuleBasic
	keeper     keeper.Keeper
	accKeeper  quarantine.AccountKeeper
	bankKeeper quarantine.BankKeeper
	registry   cdctypes.InterfaceRegistry
}

func NewAppModule(cdc codec.Codec, quarantineKeeper keeper.Keeper, accKeeper quarantine.AccountKeeper, bankKeeper quarantine.BankKeeper, registry cdctypes.InterfaceRegistry) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         quarantineKeeper,
		accKeeper:      accKeeper,
		bankKeeper:     bankKeeper,
		registry:       registry,
	}
}

type AppModuleBasic struct {
	cdc codec.Codec
}

func (AppModuleBasic) Name() string {
	return quarantine.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the quarantine module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(quarantine.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the quarantine module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data quarantine.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", quarantine.ModuleName, err)
	}
	return data.Validate()
}

// GetQueryCmd returns the cli query commands for the quarantine module
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.QueryCmd()
}

// GetTxCmd returns the transaction commands for the quarantine module
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.TxCmd()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the quarantine module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *runtime.ServeMux) {
	if err := quarantine.RegisterQueryHandlerClient(context.Background(), mux, quarantine.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers the quarantine module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	quarantine.RegisterInterfaces(registry)
}

// RegisterLegacyAminoCodec registers the quarantine module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	quarantine.RegisterLegacyAminoCodec(cdc)
}

// Name returns the quarantine module's name.
func (AppModule) Name() string {
	return quarantine.ModuleName
}

// RegisterInvariants does nothing, there are no invariants to enforce for the quarantine module.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// Deprecated: Route returns the message routing key for the quarantine module, empty.
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// Deprecated: QuerierRoute returns the route we respond to for abci queries, "".
func (AppModule) QuerierRoute() string { return "" }

// Deprecated: LegacyQuerierHandler returns the quarantine module sdk.Querier (nil).
func (am AppModule) LegacyQuerierHandler(_ *codec.LegacyAmino) sdk.Querier {
	return nil
}

// InitGenesis performs genesis initialization for the quarantine module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState quarantine.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the quarantine module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// RegisterServices registers a gRPC query service to respond to the quarantine-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	quarantine.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	quarantine.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// ____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the quarantine module.
func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState, am.keeper.GetFundsHolder())
}

// ProposalContents returns all the quarantine content functions used to
// simulate governance proposals.
func (am AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized quarantine param changes for the simulator.
func (AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {
	return nil
}

// RegisterStoreDecoder registers a decoder for quarantine module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[quarantine.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the quarantine module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc,
		am.accKeeper, am.bankKeeper, am.keeper, am.cdc,
	)
}
