package distribution

import (
	modulev1 "cosmossdk.io/api/cosmos/distribution/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/distribution/keeper"
	"cosmossdk.io/x/distribution/types"
	staking "cosmossdk.io/x/staking/types"

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
	Environment appmodule.Environment
	Cdc         codec.Codec

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
	StakingKeeper types.StakingKeeper
	PoolKeeper    types.PoolKeeper
}

type ModuleOutputs struct {
	depinject.Out

	DistrKeeper keeper.Keeper
	Module      appmodule.AppModule
	Hooks       staking.StakingHooksWrapper
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	feeCollectorName := in.Config.FeeCollectorName
	if feeCollectorName == "" {
		feeCollectorName = authtypes.FeeCollectorName
	}

	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(types.GovModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	k := keeper.NewKeeper(
		in.Cdc,
		in.Environment,
		in.AccountKeeper,
		in.BankKeeper,
		in.StakingKeeper,
		in.PoolKeeper,
		feeCollectorName,
		authority.String(),
	)

	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BankKeeper, in.StakingKeeper, in.PoolKeeper)

	return ModuleOutputs{
		DistrKeeper: k,
		Module:      m,
		Hooks:       staking.StakingHooksWrapper{StakingHooks: k.Hooks()},
	}
}
