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
// * ante handler
// * genesis
// * cli
// * rest
// -> changes to auth, etc

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
	panic("not implemented!")
	// return ModuleCdc.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the delegation module.
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	// TODO
	return nil
	// var data GenesisState
	// err := ModuleCdc.UnmarshalJSON(bz, &data)
	// if err != nil {
	// 	return err
	// }
	// return ValidateGenesis(data)
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
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// TODO?
	// RegisterInvariants(ir, am.keeper)
}

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
	panic("not implemented!")
	// return NewQuerier(am.keeper)
}

// InitGenesis performs genesis initialization for the delegation module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	// TODO
	// var genesisState GenesisState
	// ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	// InitGenesis(ctx, am.keeper, am.ak, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the delegation
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	panic("not implemented!")
	// gs := ExportGenesis(ctx, am.keeper)
	// return ModuleCdc.MustMarshalJSON(gs)
}

// BeginBlock returns the begin blocker for the delegation module.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock returns the end blocker for the delegation module. It returns no validator
// updates.
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
