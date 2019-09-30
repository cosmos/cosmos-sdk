package upgrade

import (
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	"github.com/cosmos/cosmos-sdk/x/upgrade/client/cli"
	"github.com/cosmos/cosmos-sdk/x/upgrade/client/rest"
	abci "github.com/tendermint/tendermint/abci/types"
)

// module codec
var moduleCdc = codec.New()

func init() {
	upgrade.RegisterCodec(moduleCdc)
}

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// app module basics object
type AppModuleBasic struct{}

// module name
func (AppModuleBasic) Name() string {
	return upgrade.ModuleName
}

// register module codec
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	upgrade.RegisterCodec(cdc)
}

// default genesis state
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return moduleCdc.MustMarshalJSON(upgrade.DefaultGenesisState())
}

// module validate genesis
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data upgrade.GenesisState
	err := moduleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return upgrade.ValidateGenesis(data)
}

func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, r *mux.Router) {
	rest.RegisterRoutes(ctx, r, moduleCdc, upgrade.StoreKey)
}

// GetQueryCmd returns the cli query commands for this module
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Querying commands for the upgrade module",
	}
	queryCmd.AddCommand(client.GetCommands(
		cli.GetPlanCmd(upgrade.StoreKey, cdc),
		cli.GetAppliedHeightCmd(upgrade.StoreKey, cdc),
	)...)

	return queryCmd

}

// GetTxCmd returns the transaction commands for this module
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade transaction subcommands",
	}
	txCmd.AddCommand(client.PostCommands()...)
	return txCmd
}

// app module
type AppModule struct {
	AppModuleBasic
	keeper upgrade.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper upgrade.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// module name
func (AppModule) Name() string {
	return upgrade.ModuleName
}

// register invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// module message route name
func (AppModule) Route() string { return "" }

// module handler
func (am AppModule) NewHandler() sdk.Handler { return nil }

// module querier route name
func (AppModule) QuerierRoute() string { return "" }

// module querier
func (am AppModule) NewQuerierHandler() sdk.Querier { return nil }

// module init-genesis
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState upgrade.GenesisState
	moduleCdc.MustUnmarshalJSON(data, &genesisState)
	upgrade.InitGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// module export genesis
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := upgrade.ExportGenesis(ctx, am.keeper)
	return moduleCdc.MustMarshalJSON(gs)
}

// module begin-block
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	am.keeper.BeginBlocker(ctx, req)
}

// module end-block
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
