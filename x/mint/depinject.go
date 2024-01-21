package mint

import (
	modulev1 "cosmossdk.io/api/cosmos/mint/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/mint/keeper"
	"cosmossdk.io/x/mint/types"

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

	ModuleKey    depinject.OwnModuleKey
	Config       *modulev1.Module
	StoreService store.KVStoreService
	Cdc          codec.Codec

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
	StakingKeeper types.StakingKeeper
}

type ModuleOutputs struct {
	depinject.Out

	MintKeeper keeper.Keeper
	Module     appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	feeCollectorName := in.Config.FeeCollectorName
	if feeCollectorName == "" {
		feeCollectorName = authtypes.FeeCollectorName
	}

	k := keeper.NewKeeper(
		in.Cdc,
		in.StoreService,
		in.StakingKeeper,
		in.AccountKeeper,
		in.BankKeeper,
		feeCollectorName,
	)

	// when no inflation calculation function is provided it will use the default types.DefaultInflationCalculationFn
	m := NewAppModule(in.Cdc, k, in.AccountKeeper)

	return ModuleOutputs{MintKeeper: k, Module: m}
}
