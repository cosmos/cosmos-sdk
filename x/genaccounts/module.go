package genaccounts

import (
	"encoding/json"
	"math/rand"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/genaccounts/internal/types"
	"github.com/cosmos/cosmos-sdk/x/genaccounts/simulation"
	sim "github.com/cosmos/cosmos-sdk/x/simulation"
)

var (
	_ module.AppModuleGenesis    = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the genesis accounts module.
type AppModuleBasic struct{}

// Name returns the genesis accounts module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec performs a no-op.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {}

// DefaultGenesis returns default genesis state as raw bytes for the genesis accounts
// module.
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(GenesisState{})
}

// ValidateGenesis performs genesis state validation for the genesis accounts module.
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data GenesisState
	err := ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return ValidateGenesis(data)
}

// RegisterRESTRoutes registers no REST routes for the genesis accounts module.
func (AppModuleBasic) RegisterRESTRoutes(_ context.CLIContext, _ *mux.Router) {}

// GetTxCmd returns no root tx command for the genesis accounts module.
func (AppModuleBasic) GetTxCmd(_ *codec.Codec) *cobra.Command { return nil }

// GetQueryCmd returns no root query command for the genesis accounts module.
func (AppModuleBasic) GetQueryCmd(_ *codec.Codec) *cobra.Command { return nil }

// extra function from sdk.AppModuleBasic

// IterateGenesisAccounts iterates over the genesis accounts and perform an operation at each of them
// - to used by other modules
func (AppModuleBasic) IterateGenesisAccounts(cdc *codec.Codec, appGenesis map[string]json.RawMessage,
	iterateFn func(exported.Account) (stop bool)) {

	genesisState := GetGenesisStateFromAppState(cdc, appGenesis)
	for _, genAcc := range genesisState {
		acc := genAcc.ToAccount()
		if iterateFn(acc) {
			break
		}
	}
}

//____________________________________________________________________________

// AppModule implements an application module for the genesis accounts module.
type AppModule struct {
	AppModuleBasic

	accountKeeper types.AccountKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(accountKeeper types.AccountKeeper) AppModule {

	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		accountKeeper:  accountKeeper,
	}
}

// RegisterInvariants is a placeholder function register no invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route empty module message route
func (AppModule) Route() string { return "" }

// NewHandler returns an empty module handler
func (AppModule) NewHandler() sdk.Handler { return nil }

// QuerierRoute returns an empty module querier route
func (AppModule) QuerierRoute() string { return "" }

// NewQuerierHandler returns an empty module querier
func (AppModule) NewQuerierHandler() sdk.Querier { return nil }

// InitGenesis performs genesis initialization for the genesis accounts module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, ModuleCdc, am.accountKeeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the genesis accounts
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := ExportGenesis(ctx, am.accountKeeper)
	return ModuleCdc.MustMarshalJSON(gs)
}

// BeginBlock returns an empty module begin-block
func (AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {}

// EndBlock returns an empty module end-block
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

//____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the genesis accounts module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents doesn't return any content functions for governance proposals.
func (AppModule) ProposalContents(_ module.SimulationState) []sim.WeightedProposalContent {
	return nil
}

// RandomizedParams doesn't create randomized genaccounts param changes for the simulator.
func (AppModule) RandomizedParams(_ *rand.Rand) []sim.ParamChange {
	return nil
}

// RegisterStoreDecoder performs a no-op.
func (AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations doesn't return any auth module operation.
func (AppModule) WeightedOperations(_ module.SimulationState) []sim.WeightedOperation {
	return nil
}
