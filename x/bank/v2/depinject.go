package bankv2

import (
	bankv2api "cosmossdk.io/api/cosmos/bank/module/v2"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/bank/v2/keeper"
	"cosmossdk.io/x/bank/v2/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(
		&bankv2api.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *bankv2api.Module
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
