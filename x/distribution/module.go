package distribution

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

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
func (a AppModule) RegisterInvariants(ir sdk.InvariantRouter) {
	RegisterInvariants(ir, a.keeper)
}

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
func (a AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) (sdk.Tags, error) {
	BeginBlocker(ctx, req, a.keeper)
	return sdk.EmptyTags(), nil
}

// module end-block
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) ([]abci.ValidatorUpdate, sdk.Tags, error) {
	return []abci.ValidatorUpdate{}, sdk.EmptyTags(), nil
}
