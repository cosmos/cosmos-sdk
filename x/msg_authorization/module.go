package msg_authorization

import (
	"encoding/json"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/gogo/protobuf/grpc"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/keeper"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/types"
)

// module codec
// var moduleCdc = codec.New()

// func init() {
// 	RegisterCodec(moduleCdc)
// }

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

type AppModuleBasic struct {
	cdc codec.Marshaler
}

// Name returns the ModuleName
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the distribution module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis is an empty object
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return nil
}

// ValidateGenesis is always successful, as we ignore the value
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

// RegisterRESTRoutes registers all REST query handlers
func (AppModuleBasic) RegisterRESTRoutes(clientCtx sdkclient.Context, r *mux.Router) {
	// rest.RegisterRoutes(clientCtx, r)
}

// RegisterGRPCRoutes registers the gRPC Gateway routes for the staking module.
func (AppModuleBasic) RegisterGRPCRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// TODO: Add grpc querier
	// types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

//GetQueryCmd returns the cli query commands for this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	// TODO
	return nil
}

// GetTxCmd returns the transaction commands for this module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	// TODO
	return nil
}

// AppModule implements the sdk.AppModule interface
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Marshaler, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// Name returns the msg_authorization module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants does nothing, there are no invariants to enforce
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// // Route is empty, as we do not handle Messages (just proposals)
// func (AppModule) Route() string { return types.RouterKey }

// Route returns the message routing key for the staking module.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.keeper))
}

func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// QuerierRoute returns the route we respond to for abci queries
func (AppModule) QuerierRoute() string { return types.QuerierRoute }

// LegacyQuerierHandler returns the staking module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

/// NewQuerierHandler registers a query handler to respond to the module-specific queries
func (am AppModule) NewQuerierHandler() sdk.Querier {
	//return NewQuerier(am.keeper)
	return nil
}

// RegisterQueryService registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterQueryService(server grpc.Server) {
	// TODO:
	// querier := keeper.Querier{Keeper: am.keeper}
	// types.RegisterQueryServer(server, querier)
}

// InitGenesis is ignored, no sense in serializing future upgrades
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis is always empty, as InitGenesis does nothing either
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	return am.DefaultGenesis(cdc)
}

// DefaultGenesis is an empty object
func (am AppModule) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return nil
}

func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {

}

// EndBlock does nothing
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

//____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the evidence module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns all the evidence content functions used to
// simulate governance proposals.
func (am AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized evidence param changes for the simulator.
func (AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return nil
}

// RegisterStoreDecoder registers a decoder for evidence module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	// sdr[types.StoreKey] = simulation.NewDecodeStore(am.keeper)
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
