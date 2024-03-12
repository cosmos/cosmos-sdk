package crisis

import (
	"github.com/spf13/cast"

	modulev1 "cosmossdk.io/api/cosmos/crisis/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	authtypes "cosmossdk.io/x/auth/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"
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

	Config       *modulev1.Module
	StoreService store.KVStoreService
	Codec        codec.Codec
	AppOpts      servertypes.AppOptions `optional:"true"`

	BankKeeper   types.SupplyKeeper
	AddressCodec address.Codec
}

type ModuleOutputs struct {
	depinject.Out

	Module       appmodule.AppModule
	CrisisKeeper *keeper.Keeper
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	var invalidCheckPeriod uint
	if in.AppOpts != nil {
		invalidCheckPeriod = cast.ToUint(in.AppOpts.Get(server.FlagInvCheckPeriod))
	}

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
		in.Codec,
		in.StoreService,
		invalidCheckPeriod,
		in.BankKeeper,
		feeCollectorName,
		authority.String(),
		in.AddressCodec,
	)

	var skipGenesisInvariants bool
	if in.AppOpts != nil {
		skipGenesisInvariants = cast.ToBool(in.AppOpts.Get(FlagSkipGenesisInvariants))
	}

	m := NewAppModule(k, in.Codec, skipGenesisInvariants)

	return ModuleOutputs{CrisisKeeper: k, Module: m}
}
