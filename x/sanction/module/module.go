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
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/client/cli"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/sanction/simulation"
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
	accKeeper  sanction.AccountKeeper
	bankKeeper sanction.BankKeeper
	govKeeper  sanction.GovKeeper
	registry   cdctypes.InterfaceRegistry
}

func NewAppModule(cdc codec.Codec, sanctionKeeper keeper.Keeper, accKeeper sanction.AccountKeeper, bankKeeper sanction.BankKeeper, govKeeper sanction.GovKeeper, registry cdctypes.InterfaceRegistry) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         sanctionKeeper,
		accKeeper:      accKeeper,
		bankKeeper:     bankKeeper,
		govKeeper:      govKeeper,
		registry:       registry,
	}
}

type AppModuleBasic struct {
	cdc codec.Codec
}

func (AppModuleBasic) Name() string {
	return sanction.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the sanction module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(sanction.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the sanction module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data sanction.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", sanction.ModuleName, err)
	}
	return data.Validate()
}

// GetQueryCmd returns the cli query commands for the sanction module
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.QueryCmd()
}

// GetTxCmd returns the transaction commands for the sanction module
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the sanction module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *runtime.ServeMux) {
	if err := sanction.RegisterQueryHandlerClient(context.Background(), mux, sanction.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers the sanction module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	sanction.RegisterInterfaces(registry)
}

// RegisterLegacyAminoCodec registers the sanction module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	sanction.RegisterLegacyAminoCodec(cdc)
}

// RegisterInvariants does nothing, there are no invariants to enforce for the sanction module.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Deprecated: Route returns the message routing key for the sanction module, empty.
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// Deprecated: QuerierRoute returns the route we respond to for abci queries, "".
func (AppModule) QuerierRoute() string { return "" }

// Deprecated: LegacyQuerierHandler returns the sanction module sdk.Querier (nil).
func (am AppModule) LegacyQuerierHandler(_ *codec.LegacyAmino) sdk.Querier {
	return nil
}

// InitGenesis performs genesis initialization for the sanction module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState sanction.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the sanction module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// RegisterServices registers a gRPC query service to respond to the sanction-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	sanction.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	sanction.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// ____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the sanction module.
func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns all the sanction content functions used to
// simulate governance proposals.
func (am AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	// This is for stuff that uses the v1beta1 gov module's Content interface.
	// This module use the v1 gov stuff, though, so there's nothing to do here.
	// It's all handled in the WeightedOperations.
	return nil
}

// RandomizedParams creates randomized sanction param changes for the simulator.
func (AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {
	// While the x/sanction module does have "Params", it doesn't use the x/params module.
	// So there's nothing to return here.
	return nil
}

// RegisterStoreDecoder registers a decoder for sanction module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[sanction.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the sanction module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc, codec.NewProtoCodec(am.registry),
		am.accKeeper, am.bankKeeper, am.govKeeper, am.keeper,
	)
}
