package mint

import (
	"fmt"

	modulev1 "cosmossdk.io/api/cosmos/mint/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/moduleaccounts"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	epochstypes "cosmossdk.io/x/epochs/types"
	"cosmossdk.io/x/mint/keeper"
	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(&modulev1.Module{},
		appconfig.Provide(ProvideModule),
		appconfig.Invoke(InvokeSetMintFn),
	)
}

type ModuleInputs struct {
	depinject.In

	Config                *modulev1.Module
	Environment           appmodule.Environment
	Cdc                   codec.Codec
	ModuleAccountsService moduleaccounts.Service
	AddressCdc            address.Codec

	BankKeeper types.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	MintKeeper     *keeper.Keeper
	Module         appmodule.AppModule
	EpochHooks     epochstypes.EpochHooksWrapper
	ModuleAccounts []runtime.ModuleAccount
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

	as, err := in.AddressCdc.BytesToString(authority)
	if err != nil {
		panic(err)
	}

	k := keeper.NewKeeper(
		in.Cdc,
		in.Environment,
		in.BankKeeper,
		feeCollectorName,
		as,
		in.ModuleAccountsService,
	)

	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{
		MintKeeper:     k,
		Module:         m,
		EpochHooks:     epochstypes.EpochHooksWrapper{EpochHooks: m},
		ModuleAccounts: []runtime.ModuleAccount{runtime.NewModuleAccount(types.ModuleName, "minter")},
	}
}

func InvokeSetMintFn(mintKeeper *keeper.Keeper, mintFn types.MintFn, stakingKeeper types.StakingKeeper) error {
	if mintFn == nil && stakingKeeper == nil {
		return fmt.Errorf("custom minting function or staking keeper must be supplied or available")
	} else if mintFn == nil {
		mintFn = keeper.DefaultMintFn(types.DefaultInflationCalculationFn, stakingKeeper, mintKeeper)
	}

	return mintKeeper.SetMintFn(mintFn)
}
