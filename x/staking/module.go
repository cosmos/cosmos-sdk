package staking

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/codec"
	abci "github.com/tendermint/tendermint/abci/types"
)

// app module basics object
type AppModuleBasic struct{}

var _ sdk.AppModuleBasic = AppModuleBasic{}

// module name
func (AppModuleBasic) Name() string {
	return ModuleName
}

// module name
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) string {
	return RegisterCodec(cdc)
}

// module name
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return moduleCdc.MustMarshalJSON(DefaultGenesisState())
}

//___________________________
// app module
type AppModule struct {
	AppModuleBasic
	keeper      Keeper
	fcKeeper    FeeCollectionKeeper
	distrKeeper DistributionKeeper
	accKeeper   auth.AccountKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper Keeper, fcKeeper types.FeeCollectionKeeper,
	distrKeeper types.DistributionKeeper, accKeeper auth.AccountKeeper) AppModule {

	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
		fcKeeper:       fcKeeper,
		distrKeeper:    distrKeeper,
		accKeeper:      accKeeper,
	}
}

var _ sdk.AppModule = AppModule{}

// module name
func (AppModule) Name() string {
	return ModuleName
}

// register invariants
func (a AppModule) RegisterInvariants(ir sdk.InvariantRouter) {
	RegisterInvariants(ir, a.keeper, a.fcKeeper, a.distrKeeper, a.accKeeper)
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

// module init-genesis
func (a AppModule) InitGenesis(_ sdk.Context, _ json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// module validate genesis
func (AppModule) ValidateGenesis(bz json.RawMessage) error {
	var data GenesisState
	err := ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return ValidateGenesis(data)
}

// module export genesis
func (a AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := ExportGenesis(ctx, a.keeper)
	return ModuleCdc.MustMarshalJSON(gs)
}

// module begin-block
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) sdk.Tags {
	return sdk.EmptyTags()
}

// module end-block
func (a AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) ([]abci.ValidatorUpdate, sdk.Tags) {
	validatorUpdates, tags := EndBlocker(ctx, a.keeper)
	return validatorUpdates, tags
}
