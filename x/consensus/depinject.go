package consensus

import (
	modulev1 "cosmossdk.io/api/cosmos/consensus/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	authtypes "cosmossdk.io/x/auth/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/x/consensus/keeper"
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

	Config      *modulev1.Module
	Cdc         codec.Codec
	Environment appmodule.Environment
}

type ModuleOutputs struct {
	depinject.Out

	Keeper        keeper.Keeper
	Module        appmodule.AppModule
	BaseAppOption runtime.BaseAppOption
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress("gov")
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	k := keeper.NewKeeper(in.Cdc, in.Environment, authority.String())
	m := NewAppModule(in.Cdc, k)
	baseappOpt := func(app *baseapp.BaseApp) {
		app.SetParamStore(k.ParamsStore)
	}

	return ModuleOutputs{
		Keeper:        k,
		Module:        m,
		BaseAppOption: baseappOpt,
	}
}
