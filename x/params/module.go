package params

import (
	"encoding/json"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/params/simulation"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
	_ module.InterfaceModule     = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the params module.
type AppModuleBasic struct{}

// Name returns the params module's name.
func (AppModuleBasic) Name() string {
	return proposal.ModuleName
}

// RegisterCodec registers the params module's types for the given codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	proposal.RegisterCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the params
// module.
func (AppModuleBasic) DefaultGenesis(_ codec.JSONMarshaler) json.RawMessage { return nil }

// ValidateGenesis performs genesis state validation for the params module.
func (AppModuleBasic) ValidateGenesis(_ codec.JSONMarshaler, _ json.RawMessage) error { return nil }

// RegisterRESTRoutes registers the REST routes for the params module.
func (AppModuleBasic) RegisterRESTRoutes(_ context.CLIContext, _ *mux.Router) {}

// GetTxCmd returns no root tx command for the params module.
func (AppModuleBasic) GetTxCmd(_ context.CLIContext) *cobra.Command { return nil }

// GetQueryCmd returns no root query command for the params module.
func (AppModuleBasic) GetQueryCmd(_ *codec.Codec) *cobra.Command { return nil }

func (am AppModuleBasic) RegisterInterfaceTypes(registry types.InterfaceRegistry) {
	proposal.RegisterInterfaces(registry)
}

//____________________________________________________________________________

// AppModule implements an application module for the distribution module.
type AppModule struct {
	AppModuleBasic

	keeper Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(k Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (am AppModule) NewHandler() sdk.Handler { return nil }

// InitGenesis performs a no-op.
func (am AppModule) InitGenesis(_ sdk.Context, _ codec.JSONMarshaler, _ json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func (AppModule) Route() string { return "" }

// GenerateGenesisState performs a no-op.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {}

// QuerierRoute returns the x/param module's querier route name.
func (AppModule) QuerierRoute() string { return QuerierRoute }

// NewQuerierHandler returns the x/params querier handler.
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// ProposalContents returns all the params content functions used to
// simulate governance proposals.
func (am AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return simulation.ProposalContents(simState.ParamChanges)
}

// RandomizedParams creates randomized distribution param changes for the simulator.
func (AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return nil
}

// RegisterStoreDecoder doesn't register any type.
func (AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}

// ExportGenesis performs a no-op.
func (am AppModule) ExportGenesis(_ sdk.Context, _ codec.JSONMarshaler) json.RawMessage {
	return nil
}

// BeginBlock performs a no-op.
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock performs a no-op.
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
