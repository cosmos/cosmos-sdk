package capability

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}

	// TODO: Enable simulation once concrete types are defined.
	// _ module.AppModuleSimulation = AppModuleSimulation{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface for the capability module.
type AppModuleBasic struct {
}

func NewAppModuleBasic() AppModuleBasic {
	return AppModuleBasic{}
}

// Name returns the capability module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers the capability module's types to the provided codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

// DefaultGenesis returns the capability module's default genesis state.
func (AppModuleBasic) DefaultGenesis(_ codec.JSONMarshaler) json.RawMessage { return []byte("{}") }

// ValidateGenesis performs genesis state validation for the capability module.
func (AppModuleBasic) ValidateGenesis(_ codec.JSONMarshaler, _ json.RawMessage) error { return nil }

// RegisterRESTRoutes registers the capability module's REST service handlers.
func (a AppModuleBasic) RegisterRESTRoutes(_ context.CLIContext, _ *mux.Router) {}

// GetTxCmd returns the capability module's root tx command.
func (a AppModuleBasic) GetTxCmd(_ *codec.Codec) *cobra.Command { return nil }

// GetTxCmd returns the capability module's root query command.
func (AppModuleBasic) GetQueryCmd(_ *codec.Codec) *cobra.Command { return nil }

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface for the capability module.
type AppModule struct {
	AppModuleBasic

	keeper Keeper
}

func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(),
		keeper:         keeper,
	}
}

// Name returns the capability module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// Route returns the capability module's message routing key.
func (AppModule) Route() string { return "" }

// QuerierRoute returns the capability module's query routing key.
func (AppModule) QuerierRoute() string { return "" }

// NewHandler returns the capability module's message Handler.
func (am AppModule) NewHandler() sdk.Handler { return nil }

// NewQuerierHandler returns the capability module's Querier.
func (am AppModule) NewQuerierHandler() sdk.Querier { return nil }

// RegisterInvariants registers the capability module's invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs the capability module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(_ sdk.Context, _ codec.JSONMarshaler, _ json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the capability module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(_ sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	return am.DefaultGenesis(cdc)
}

// BeginBlock executes all ABCI BeginBlock logic respective to the capability module.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock executes all ABCI EndBlock logic respective to the capability module. It
// returns no validator updates.
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
