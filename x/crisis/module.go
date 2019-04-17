package crisis

import (
	"encoding/json"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// name of this module
const ModuleName = "crisis"

// app module for bank
type AppModule struct {
	keeper Keeper
	logger log.Logger
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper Keeper, logger log.Logger) AppModule {
	return AppModule{
		keeper: keeper,
		logger: logger,
	}
}

var _ sdk.AppModule = AppModule{}

// module name
func (AppModule) Name() string {
	return ModuleName
}

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
func (a AppModule) InitGenesis(ctx sdk.Context, _ json.RawMessage) ([]abci.ValidatorUpdate, error) {
	a.keeper.AssertInvariants(ctx, a.logger)
	return []abci.ValidatorUpdate{}, nil
}

// module validate genesis
func (AppModule) ValidateGenesis(bz json.RawMessage) error {
	var data GenesisState
	err := moduleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return ValidateGenesis(data)
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
