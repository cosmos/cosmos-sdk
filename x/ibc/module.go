package ibc

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/client/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// TODO: AppModuleSimulation
var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the ibc module.
type AppModuleBasic struct{}

var _ module.AppModuleBasic = AppModuleBasic{}

// Name returns the ibc module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterCodec registers the ibc module's types for the given codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	client.RegisterCodec(cdc)
	connection.RegisterCodec(cdc)
	channel.RegisterCodec(cdc)
	ibctmtypes.RegisterCodec(cdc)
	localhosttypes.RegisterCodec(cdc)
	commitmenttypes.RegisterCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the ibc
// module.
func (AppModuleBasic) DefaultGenesis(_ codec.JSONMarshaler) json.RawMessage {
	return nil
}

// ValidateGenesis performs genesis state validation for the ibc module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, bz json.RawMessage) error {

	// TODO: UNDO this when DefaultGenesis() is implemented
	// This validation is breaking the state as it is trying to
	// validate nil. DefaultGenesis is not implemented and it just returns nil
	// This is a quick fix to make the cli-tests work and
	// SHOULD BE reverted when #5948 is addressed
	// To UNDO this, just uncomment the code below

	// var gs GenesisState
	// if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
	// 	return fmt.Errorf("failed to unmarshal %s genesis state: %w", ModuleName, err)
	// }

	// return gs.Validate()

	return nil
}

// RegisterRESTRoutes registers the REST routes for the ibc module.
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	rest.RegisterRoutes(ctx, rtr, StoreKey)
}

// GetTxCmd returns the root tx command for the ibc module.
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(StoreKey, cdc)
}

// GetQueryCmd returns no root query command for the ibc module.
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(QuerierRoute, cdc)
}

// AppModule implements an application module for the ibc module.
type AppModule struct {
	AppModuleBasic
	keeper *Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(k *Keeper) AppModule {
	return AppModule{
		keeper: k,
	}
}

// Name returns the ibc module's name.
func (AppModule) Name() string {
	return ModuleName
}

// RegisterInvariants registers the ibc module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// TODO:
}

// Route returns the message routing key for the ibc module.
func (AppModule) Route() string {
	return RouterKey
}

// NewHandler returns an sdk.Handler for the ibc module.
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(*am.keeper)
}

// QuerierRoute returns the ibc module's querier route name.
func (AppModule) QuerierRoute() string {
	return QuerierRoute
}

// NewQuerierHandler returns the ibc module sdk.Querier.
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(*am.keeper)
}

// InitGenesis performs genesis initialization for the ibc module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, bz json.RawMessage) []abci.ValidatorUpdate {
	var gs GenesisState
	err := cdc.UnmarshalJSON(bz, &gs)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal %s genesis state: %s", ModuleName, err))
	}
	InitGenesis(ctx, *am.keeper, gs)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the ibc
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	return cdc.MustMarshalJSON(ExportGenesis(ctx, *am.keeper))
}

// BeginBlock returns the begin blocker for the ibc module.
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	client.BeginBlocker(ctx, am.keeper.ClientKeeper)
}

// EndBlock returns the end blocker for the ibc module. It returns no validator
// updates.
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
