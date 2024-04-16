package vesting

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/x/auth/keeper"
	"cosmossdk.io/x/auth/vesting/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var (
	_ module.AppModule = AppModule{}
	_ module.HasName   = AppModule{}

	_ appmodule.AppModule = AppModule{}
)

// AppModule implementing the AppModule interface.
type AppModule struct {
	accountKeeper keeper.AccountKeeper
	bankKeeper    types.BankKeeper
}

func NewAppModule(ak keeper.AccountKeeper, bk types.BankKeeper) AppModule {
	return AppModule{
		accountKeeper: ak,
		bankKeeper:    bk,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Name returns the module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterCodec registers the module's types with the given codec.
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interfaces and implementations with
// the given interface registry.
func (AppModule) RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	types.RegisterInterfaces(registrar)
}

// ConsensusVersion implements HasConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }
