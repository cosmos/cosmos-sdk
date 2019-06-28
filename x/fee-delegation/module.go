package fee_delegation

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ModuleName is the name of this module
const ModuleName = "fee-delegation"
const StoreKey = ModuleName

// AppModuleBasic defines the internal data for the module
// ----------------------------------------------------------------------------
type AppModuleBasic struct{}

var _ module.AppModuleBasic = AppModuleBasic{}

// Name define the name of the module
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers the types needed for amino encoding/decoding
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

// DefaultGenesis creates the default genesis state for testing
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return moduleCodec.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis validates the genesis state
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data GenesisState
	err := moduleCodec.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return ValidateGenesis(data)
}

// register rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	// rest.RegisterRoutes(ctx, rtr, StoreKey)
}

// get the root tx command of this module
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   ModuleName,
		Short: "Fee delegation transaction subcommands",
	}

	txCmd.AddCommand(client.PostCommands(
		GetCmdDelegateFees(cdc),
		GetCmdRevokeDelegatedFees(cdc),
	)...)

	return txCmd
}

// get the root query command of this module
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return GetQueryCmd(ModuleName, cdc)
}

// AppModule defines external data for the module
// ----------------------------------------------------------------------------
type AppModule struct {
	AppModuleBasic
	keeper Keeper
}

// NewAppModule creates a new app module
func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// RegisterInvariants enforces registering of invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route defines the key for the route
func (AppModule) Route() string {
	return RouterKey
}

// NewHandler creates the handler for the delegation module
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// QuerierRoute defines the querier route
func (AppModule) QuerierRoute() string {
	return QuerierRoute
}

// NewQuerierHandler creates a new querier handler
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// InitGenesis enforces the creation of the genesis state for the delegation module
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	moduleCodec.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis enforces exporting this module's data to a genesis file
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return moduleCodec.MustMarshalJSON(gs)
}

// BeginBlock runs before a block is processed
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {
}

// EndBlock runs at the end of each block
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
