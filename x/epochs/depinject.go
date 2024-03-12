package epochs

import (
	modulev1 "cosmossdk.io/api/cosmos/epochs/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/epochs/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(&modulev1.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config      *modulev1.Module
	Cdc         codec.Codec
	Environment appmodule.Environment
}

type ModuleOutputs struct {
	depinject.Out

	EpochKeeper keeper.Keeper
	Module      appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(in.Environment, in.Cdc)
	m := NewAppModule(in.Cdc, k)
	return ModuleOutputs{EpochKeeper: k, Module: m}
}
