package bankv2

import (
	"fmt"
	"maps"
	"slices"
	"sort"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/bank/v2/keeper"
	"cosmossdk.io/x/bank/v2/types"
	moduletypes "cosmossdk.io/x/bank/v2/types/module"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(
		&moduletypes.Module{},
		appconfig.Provide(ProvideModule),
		appconfig.Invoke(InvokeSetSendRestrictions),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *moduletypes.Module
	Cdc          codec.Codec
	Environment  appmodule.Environment
	AddressCodec address.Codec
}

type ModuleOutputs struct {
	depinject.Out

	Keeper *keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// default to governance authority if not provided
	authority := sdkaddress.Module(types.GovModuleName)
	if in.Config.Authority != "" {
		bz, err := in.AddressCodec.StringToBytes(in.Config.Authority)
		if err != nil { // module name
			authority = sdkaddress.Module(in.Config.Authority)
		} else { // actual address
			authority = bz
		}
	}

	k := keeper.NewKeeper(authority, in.AddressCodec, in.Environment, in.Cdc)
	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{
		Keeper: k,
		Module: m,
	}
}

func InvokeSetSendRestrictions(
	config *moduletypes.Module,
	keeper keeper.Keeper,
	restrictions map[string]types.SendRestrictionFn,
) error {
	if config == nil {
		return nil
	}

	modules := slices.Collect(maps.Keys(restrictions))
	order := config.RestrictionsOrder
	if len(order) == 0 {
		order = modules
		sort.Strings(order)
	}

	if len(order) != len(modules) {
		return fmt.Errorf("len(restrictions order: %v) != len(restriction modules: %v)", order, modules)
	}

	if len(modules) == 0 {
		return nil
	}

	for _, module := range order {
		restriction, ok := restrictions[module]
		if !ok {
			return fmt.Errorf("can't find send restriction for module %s", module)
		}

		keeper.AppendGlobalSendRestriction(restriction)
	}

	return nil
}
