package module

import (
	"context"
	"encoding/json"
	"math/rand"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/client/cli"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/simulation"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic defines the basic application module used by the feegrant module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the feegrant module's name.
func (AppModuleBasic) Name() string {
	return feegrant.ModuleName
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	feegrant.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	feegrant.RegisterQueryServer(cfg.QueryServer(), am.keeper)
	m := keeper.NewMigrator(am.keeper)
	err := cfg.RegisterMigration(feegrant.ModuleName, 1, m.Migrate1to2)
	if err != nil {
		panic(err)
	}
}

// RegisterLegacyAminoCodec registers the feegrant module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	feegrant.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the feegrant module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	feegrant.RegisterInterfaces(registry)
}

// LegacyQuerierHandler returns the feegrant module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// DefaultGenesis returns default genesis state as raw bytes for the feegrant
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(feegrant.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the feegrant module.
func (a AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data feegrant.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		sdkerrors.Wrapf(err, "failed to unmarshal %s genesis state", feegrant.ModuleName)
	}

	return feegrant.ValidateGenesis(data)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the feegrant module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *runtime.ServeMux) {
	if err := feegrant.RegisterQueryHandlerClient(context.Background(), mux, feegrant.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the feegrant module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns no root query command for the feegrant module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements an application module for the feegrant module.
type AppModule struct {
	AppModuleBasic
	keeper        keeper.Keeper
	accountKeeper feegrant.AccountKeeper
	bankKeeper    feegrant.BankKeeper
	registry      cdctypes.InterfaceRegistry
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, ak feegrant.AccountKeeper, bk feegrant.BankKeeper, keeper keeper.Keeper, registry cdctypes.InterfaceRegistry) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		accountKeeper:  ak,
		bankKeeper:     bk,
		registry:       registry,
	}
}

// Name returns the feegrant module's name.
func (AppModule) Name() string {
	return feegrant.ModuleName
}

// RegisterInvariants registers the feegrant module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Deprecated: Route returns the message routing key for the feegrant module.
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// NewHandler returns an sdk.Handler for the feegrant module.
func (am AppModule) NewHandler() sdk.Handler {
	return nil
}

// QuerierRoute returns the feegrant module's querier route name.
func (AppModule) QuerierRoute() string {
	return ""
}

// InitGenesis performs genesis initialization for the feegrant module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) []abci.ValidatorUpdate {
	var gs feegrant.GenesisState
	cdc.MustUnmarshalJSON(bz, &gs)

	err := am.keeper.InitGenesis(ctx, &gs)
	if err != nil {
		panic(err)
	}
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the feegrant
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}

	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 2 }

// EndBlock returns the end blocker for the feegrant module. It returns no validator
// updates.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	EndBlocker(ctx, am.keeper)
	return []abci.ValidatorUpdate{}
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the feegrant module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns all the feegrant content functions used to
// simulate governance proposals.
func (AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized feegrant param changes for the simulator.
func (AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return nil
}

// RegisterStoreDecoder registers a decoder for feegrant module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[feegrant.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns all the feegrant module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc, am.accountKeeper, am.bankKeeper, am.keeper,
	)
}
