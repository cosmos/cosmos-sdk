package vesting

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"

	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var _ appmodule.AppModule = AppModule{}

// AppModule implementing the AppModule interface.
type AppModule struct {
	accountKeeper keeper.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
}

// NewAppModule returns a new vesting AppModule.
func NewAppModule(ak keeper.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		accountKeeper:  ak,
		bankKeeper:     bk,
		stakingKeeper:  sk,
	}
}

// RegisterInvariants performs a no-op; there are no invariants to enforce.
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the module's message router and handler.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.accountKeeper, am.bankKeeper, am.stakingKeeper))
}

// QuerierRoute returns an empty string as the module contains no query
// functionality.
func (AppModule) QuerierRoute() string { return "" }

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(am.accountKeeper, am.bankKeeper, am.stakingKeeper))
}

// LegacyQuerierHandler performs a no-op.
func (am AppModule) LegacyQuerierHandler(_ *codec.LegacyAmino) sdk.Querier {
	return nil
}

// InitGenesis performs a no-op.
func (am AppModule) InitGenesis(_ sdk.Context, _ codec.JSONCodec, _ json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// RegisterLegacyAminoCodec registers the module's types with the given codec.
func (AppModule) RegisterLegacyAminoCodec(registrar registry.AminoRegistrar) {
	types.RegisterLegacyAminoCodec(registrar)
}

// RegisterInterfaces registers the module's interfaces and implementations with
// the given interface registry.
func (AppModule) RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	types.RegisterInterfaces(registrar)
}

// ConsensusVersion implements HasConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }
