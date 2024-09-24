package mint

import (
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
	)
}

type ModuleInputs struct {
	depinject.In

	ModuleKey              depinject.OwnModuleKey
	Config                 *modulev1.Module
	Environment            appmodule.Environment
	Cdc                    codec.Codec
	MintFn                 types.MintFn                 `optional:"true"`
	InflationCalculationFn types.InflationCalculationFn `optional:"true"` // deprecated

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
	StakingKeeper types.StakingKeeper `optional:true`
}

type ModuleOutputs struct {
	depinject.Out

	MintKeeper keeper.Keeper
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
	if in.MintFn == nil && in.StakingKeeper != nil {
		panic("custom minting function or staking keeper must be supplied or available")
	} else if in.MintFn == nil {
		in.MintFn = keeper.DefaultMintFn(types.DefaultInflationCalculationFn, in.StakingKeeper, k)
	}
	if err := k.SetMintFn(in.MintFn); err != nil {
		panic(err)
	}

	m := NewAppModule(in.Cdc, k, in.AccountKeeper)

	return ModuleOutputs{MintKeeper: k, Module: m, EpochHooks: epochstypes.EpochHooksWrapper{EpochHooks: m}}
}
