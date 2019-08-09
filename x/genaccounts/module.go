package genaccounts

import (
	"encoding/json"

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
)

var (
	_ module.AppModuleGenesis    = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModuleSimulation{}
)

// AppModuleBasic defines the basic application module used by the genesis accounts module.
type AppModuleBasic struct{}

// Name returns the genesis accounts module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers the genesis accounts module's types for the given codec.
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

// AppModuleSimulation defines the module simulation functions used by the genesis accounts module.
type AppModuleSimulation struct{}

// RegisterStoreDecoder performs a no-op.
func (AppModuleSimulation) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// GenerateGenesisState creates a randomized GenState of the genesis accounts module.
func (AppModuleSimulation) GenerateGenesisState(cdc *codec.Codec, r *rand.Rand, genesisState map[string]json.RawMessage) {
	simulation.GenGenesisAccounts(cdc, r, genesisState)
}

//____________________________________________________________________________

// AppModule implements an application module for the genesis accounts module.
type AppModule struct {
	AppModuleBasic
	AppModuleSimulation

	accountKeeper types.AccountKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(accountKeeper types.AccountKeeper) module.AppModule {

	return module.NewGenesisOnlyAppModule(AppModule{
		AppModuleBasic:      AppModuleBasic{},
		AppModuleSimulation: AppModuleSimulation{},
		accountKeeper:       accountKeeper,
	})
}

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
