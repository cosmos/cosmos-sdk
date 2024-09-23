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
	StakingKeeper types.StakingKeeper
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

	if in.MintFn == nil {
		panic("mintFn cannot be nil")
	}

	k := keeper.NewKeeper(
		in.Cdc,
		in.Environment,
		in.AccountKeeper,
		in.BankKeeper,
		feeCollectorName,
		as,
	)

	k.SetMintFn(in.MintFn)

	m := NewAppModule(in.Cdc, k, in.AccountKeeper)

	return ModuleOutputs{MintKeeper: k, Module: m, EpochHooks: epochstypes.EpochHooksWrapper{EpochHooks: m}}
}
