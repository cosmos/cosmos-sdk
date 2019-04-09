package gov

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// name of this module
const ModuleName = "gov"

// app module
type AppModule struct {
	keeper Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		keeper: keeper,
	}
}

var _ sdk.AppModule = AppModule{}

// module name
func (AppModule) Name() string {
	return ModuleName
}

// register app codec
func (AppModule) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

// register invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRouter) {}

// module message route name
func (AppModule) Route() string {
	return RouterKey
}

// module handler
func (a AppModule) NewHandler() sdk.Handler {
	return NewHandler(a.keeper)
}

// module querier route name
func (AppModule) QuerierRoute() string {
	return QuerierRoute
}

// module querier
func (a AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(a.keeper)
}

// module begin-block
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) (sdk.Tags, error) {
	return sdk.EmptyTags(), nil
}

// module end-block
func (a AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) ([]abci.ValidatorUpdate, sdk.Tags, error) {
	tags := EndBlocker(ctx, a.keeper)
	return []abci.ValidatorUpdate{}, tags, nil
}
