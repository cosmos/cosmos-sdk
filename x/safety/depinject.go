package safety

import (
	modulev1 "cosmossdk.io/api/cosmos/safety/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/safety/keeper"
)

func init() {
	appconfig.RegisterModule(&modulev1.Module{},
		appconfig.Provide(ProvideModule),
	)
}

// ModuleInputs defines the x/safety depinject inputs.
type ModuleInputs struct {
	depinject.In

	Config       *modulev1.Module
	Cdc          codec.Codec
	StoreService store.KVStoreService
	AddressCodec address.Codec
}

// ModuleOutputs defines the x/safety depinject outputs.
type ModuleOutputs struct {
	depinject.Out

	SafetyKeeper  keeper.Keeper
	Module        appmodule.AppModule
	BaseAppOption runtime.BaseAppOption
}

// ProvideModule builds the x/safety keeper and module, and a BaseApp option
// that installs the keeper as the BaseApp circuit breaker.
func ProvideModule(in ModuleInputs) ModuleOutputs {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	if in.Config.Authority != "" {
		if bz, err := in.AddressCodec.StringToBytes(in.Config.Authority); err == nil {
			authority = bz
		} else {
			authority = authtypes.NewModuleAddress(in.Config.Authority)
		}
	}
	authorityStr, err := in.AddressCodec.BytesToString(authority)
	if err != nil {
		panic(err)
	}

	k := keeper.NewKeeper(in.Cdc, in.StoreService, in.AddressCodec, in.Cdc.InterfaceRegistry(), authorityStr)
	m := NewAppModule(k)

	return ModuleOutputs{
		SafetyKeeper: k,
		Module:       m,
		BaseAppOption: func(app *baseapp.BaseApp) {
			app.SetCircuitBreaker(k)
		},
	}
}
