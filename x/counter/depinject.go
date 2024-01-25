package counter

import (
	modulev1 "cosmossdk.io/api/cosmos/counter/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"

	"github.com/cosmos/cosmos-sdk/x/counter/keeper"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(
		&modulev1.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *modulev1.Module
	StoreService storetypes.KVStoreService
	EventManager event.Service
}

type ModuleOutputs struct {
	depinject.Out

	Keeper keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(in.StoreService, in.EventManager)
	m := NewAppModule(k)

	return ModuleOutputs{
		Keeper: k,
		Module: m,
	}
}
