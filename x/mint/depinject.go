package mint

import (
	"fmt"

	modulev1 "cosmossdk.io/api/cosmos/mint/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	epochstypes "cosmossdk.io/x/epochs/types"
	"cosmossdk.io/x/mint/keeper"
	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/codec"
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

	Config      *modulev1.Module
	Environment appmodule.Environment
	Cdc         codec.Codec

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	MintKeeper *keeper.Keeper
	Module     appmodule.AppModule
	EpochHooks epochstypes.EpochHooksWrapper
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

	as, err := in.AccountKeeper.AddressCodec().BytesToString(authority)
	if err != nil {
		panic(err)
	}

	k := keeper.NewKeeper(
		in.Cdc,
		in.Environment,
		in.AccountKeeper,
		in.BankKeeper,
		feeCollectorName,
		as,
	)

	m := NewAppModule(in.Cdc, k, in.AccountKeeper)

	return ModuleOutputs{MintKeeper: k, Module: m, EpochHooks: epochstypes.EpochHooksWrapper{EpochHooks: m}}
}

func InvokeSetMintFn(
	mintKeeper *keeper.Keeper,
	stakingKeeper types.StakingKeeper,
	mintFn types.MintFn,
	inflationCalculationFn types.InflationCalculationFn,
) error {
	if mintFn == nil && stakingKeeper == nil && inflationCalculationFn == nil {
		return fmt.Errorf("custom minting function, inflation calculation function or staking keeper must be supplied or available")
	} else if mintFn != nil && inflationCalculationFn != nil {
		return fmt.Errorf("cannot set both custom minting function and inflation calculation function")
	} else if mintFn == nil {
		if inflationCalculationFn != nil && stakingKeeper != nil {
			mintFn = keeper.DefaultMintFn(inflationCalculationFn, stakingKeeper, mintKeeper)
		} else {
			mintFn = keeper.DefaultMintFn(types.DefaultInflationCalculationFn, stakingKeeper, mintKeeper)
		}
	}

	return mintKeeper.SetMintFn(mintFn)
}
