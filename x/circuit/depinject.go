package circuit

import (
	modulev1 "cosmossdk.io/api/cosmos/circuit/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/circuit/keeper"
	"cosmossdk.io/x/circuit/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

	AddressCodec address.Codec
}

type ModuleOutputs struct {
	depinject.Out

	CircuitKeeper  keeper.Keeper
	Module         appmodule.AppModule
	BaseappOptions runtime.BaseAppOption
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(types.GovModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	authorityAddr, err := in.AddressCodec.BytesToString(authority)
	if err != nil {
		panic(err)
	}

	circuitkeeper := keeper.NewKeeper(
		in.Environment,
		in.Cdc,
		authorityAddr,
		in.AddressCodec,
	)
	m := NewAppModule(in.Cdc, circuitkeeper)

	baseappOpt := func(app *baseapp.BaseApp) {
		app.SetCircuitBreaker(&circuitkeeper)
	}

	return ModuleOutputs{CircuitKeeper: circuitkeeper, Module: m, BaseappOptions: baseappOpt}
}
