package delegation

import (
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// TODO:
// * genesis
// * cli
// * rest
// * periodic fee
// -> change StdFee instead of StdTx, etc?

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the delegation module.
type AppModuleBasic struct{}

// Name returns the delegation module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers the delegation module's types for the given codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the delegation
// module.
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return []byte("[]")
}

// ValidateGenesis performs genesis state validation for the delegation module.
func (a AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	_, err := a.getValidatedGenesis(bz)
	return err
}

func (a AppModuleBasic) getValidatedGenesis(bz json.RawMessage) (GenesisState, error) {
	cdc := codec.New()
	a.RegisterCodec(cdc)

	var data GenesisState
	err := cdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return nil, err
	}
	return data, data.ValidateBasic()
}

// RegisterRESTRoutes registers the REST routes for the delegation module.
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	// TODO
	// rest.RegisterRoutes(ctx, rtr)
}

// GetTxCmd returns the root tx command for the delegation module.
func (AppModuleBasic) GetTxCmd(_ *codec.Codec) *cobra.Command {
	// TODO
	return nil
}

// GetQueryCmd returns no root query command for the delegation module.
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	// TODO
	return nil
	// return cli.GetQueryCmd(cdc)
}

//____________________________________________________________________________

// AppModule implements an application module for the delegation module.
type AppModule struct {
	AppModuleBasic
	keeper Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// Name returns the delegation module's name.
func (AppModule) Name() string {
	return ModuleName
}

// RegisterInvariants registers the delegation module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route returns the message routing key for the delegation module.
func (AppModule) Route() string {
	return RouterKey
}

// NewHandler returns an sdk.Handler for the delegation module.
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// QuerierRoute returns the delegation module's querier route name.
func (AppModule) QuerierRoute() string {
	return QuerierRoute
}

// NewQuerierHandler returns the delegation module sdk.Querier.
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// InitGenesis performs genesis initialization for the delegation module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	genesisState, err := am.getValidatedGenesis(data)
	if err != nil {
		panic(err)
	}
	err = InitGenesis(ctx, am.keeper, genesisState)
	if err != nil {
		panic(err)
	}
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the delegation
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs, err := ExportGenesis(ctx, am.keeper)
	if err != nil {
		panic(err)
	}
	return ModuleCdc.MustMarshalJSON(gs)
}

// BeginBlock returns the begin blocker for the delegation module.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock returns the end blocker for the delegation module. It returns no validator
// updates.
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
