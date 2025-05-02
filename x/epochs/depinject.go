package epochs

import (
	"fmt"
	"maps"
	"slices"

	modulev1 "cosmossdk.io/api/cosmos/epochs/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/epochs/keeper"
	"github.com/cosmos/cosmos-sdk/x/epochs/types"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(&modulev1.Module{},
		appconfig.Provide(ProvideModule),
		appconfig.Invoke(InvokeSetHooks),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *modulev1.Module
	Cdc          codec.Codec
	StoreService store.KVStoreService
}

type ModuleOutputs struct {
	depinject.Out

	EpochKeeper keeper.Keeper
	Module      appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(in.StoreService, in.Cdc)
	m := NewAppModule(k)
	return ModuleOutputs{EpochKeeper: k, Module: m}
}

func InvokeSetHooks(keeper keeper.Keeper, hooks map[string]types.EpochHooksWrapper) error {
	if hooks == nil {
		return nil
	}

	// Default ordering is lexical by module name.
	// Explicit ordering can be added to the module config if required.
	modNames := slices.Sorted(maps.Keys(hooks))
	var multiHooks types.MultiEpochHooks
	for _, modName := range modNames {
		hook, ok := hooks[modName]
		if !ok {
			return fmt.Errorf("can't find epoch hooks for module %s", modName)
		}
		multiHooks = append(multiHooks, hook)
	}

	keeper.SetHooks(multiHooks)
	return nil
}
