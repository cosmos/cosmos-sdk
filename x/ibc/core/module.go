package ibc

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	ibcclient "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/simulation"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/types"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the ibc module.
type AppModuleBasic struct{}

var _ module.AppModuleBasic = AppModuleBasic{}

// Name returns the ibc module's name.
func (AppModuleBasic) Name() string {
	return host.ModuleName
}

// RegisterLegacyAminoCodec does nothing. IBC does not support amino.
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// DefaultGenesis returns default genesis state as raw bytes for the ibc
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the ibc module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", host.ModuleName, err)
	}

	return gs.Validate()
}

// RegisterRESTRoutes does nothing. IBC does not support legacy REST routes.
func (AppModuleBasic) RegisterRESTRoutes(client.Context, *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the ibc module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	clienttypes.RegisterQueryHandlerClient(context.Background(), mux, clienttypes.NewQueryClient(clientCtx))
	connectiontypes.RegisterQueryHandlerClient(context.Background(), mux, connectiontypes.NewQueryClient(clientCtx))
	channeltypes.RegisterQueryHandlerClient(context.Background(), mux, channeltypes.NewQueryClient(clientCtx))
}

// GetTxCmd returns the root tx command for the ibc module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns no root query command for the ibc module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// RegisterInterfaces registers module concrete types into protobuf Any.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// AppModule implements an application module for the ibc module.
type AppModule struct {
	AppModuleBasic
	keeper *keeper.Keeper

	// create localhost by default
	createLocalhost bool
}

// NewAppModule creates a new AppModule object
func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{
		keeper: k,
	}
}

// Name returns the ibc module's name.
func (AppModule) Name() string {
	return host.ModuleName
}

// RegisterInvariants registers the ibc module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// TODO:
}

// Route returns the message routing key for the ibc module.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(host.RouterKey, NewHandler(*am.keeper))
}

// QuerierRoute returns the ibc module's querier route name.
func (AppModule) QuerierRoute() string {
	return host.QuerierRoute
}

// LegacyQuerierHandler returns nil. IBC does not support the legacy querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	clienttypes.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	connectiontypes.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	channeltypes.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	types.RegisterQueryService(cfg.QueryServer(), am.keeper)
}

// InitGenesis performs genesis initialization for the ibc module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, bz json.RawMessage) []abci.ValidatorUpdate {
	var gs types.GenesisState
	err := cdc.UnmarshalJSON(bz, &gs)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal %s genesis state: %s", host.ModuleName, err))
	}
	InitGenesis(ctx, *am.keeper, am.createLocalhost, &gs)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the ibc
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	return cdc.MustMarshalJSON(ExportGenesis(ctx, *am.keeper))
}

// BeginBlock returns the begin blocker for the ibc module.
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	ibcclient.BeginBlocker(ctx, am.keeper.ClientKeeper)
}

// EndBlock returns the end blocker for the ibc module. It returns no validator
// updates.
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the ibc module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents doesn't return any content functions for governance proposals.
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams returns nil since IBC doesn't register parameter changes.
func (AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {
	return nil
}

// RegisterStoreDecoder registers a decoder for ibc module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[host.StoreKey] = simulation.NewDecodeStore(*am.keeper)
}

// WeightedOperations returns the all the ibc module operations with their respective weights.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
