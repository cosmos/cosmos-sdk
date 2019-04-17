package crisis

import (
	"encoding/json"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// name of this module
const ModuleName = "crisis"

// app module basics object
type AppModuleBasic struct{}

var _ sdk.AppModuleBasic = AppModuleBasic{}

// module name
func (AppModuleBasic) Name() string {
	return ModuleName
}

// module name
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

// module name
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return moduleCdc.MustMarshalJSON(DefaultGenesisState())
}

// module validate genesis
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data GenesisState
	err := moduleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return ValidateGenesis(data)
}

//___________________________
// app module for bank
type AppModule struct {
	AppModuleBasic
	keeper Keeper
	logger log.Logger
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper Keeper, logger log.Logger) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
		logger:         logger,
	}
}

var _ sdk.AppModule = AppModule{}

// register invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRouter) {}

// module querier route name
func (AppModule) Route() string {
	return RouterKey
}

// module handler
func (a AppModule) NewHandler() sdk.Handler {
	return NewHandler(a.keeper)
}

// module querier route name
func (AppModule) QuerierRoute() string { return "" }

// module querier
func (AppModule) NewQuerierHandler() sdk.Querier { return nil }

// module init-genesis
func (a AppModule) InitGenesis(ctx sdk.Context, _ json.RawMessage) []abci.ValidatorUpdate {
	a.keeper.AssertInvariants(ctx, a.logger)
	return []abci.ValidatorUpdate{}
}

// module export genesis
func (a AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := ExportGenesis(ctx, a.keeper)
	return moduleCdc.MustMarshalJSON(gs)
}

// module begin-block
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) sdk.Tags {
	return sdk.EmptyTags()
}

// module end-block
func (a AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) ([]abci.ValidatorUpdate, sdk.Tags) {
	EndBlocker(ctx, a.keeper, a.logger)
	return []abci.ValidatorUpdate{}, sdk.EmptyTags()
}
