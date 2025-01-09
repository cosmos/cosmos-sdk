package upgrade

import (
	"github.com/spf13/cast"

	modulev1 "cosmossdk.io/api/cosmos/upgrade/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/codec"
	coreserver "cosmossdk.io/core/server"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/upgrade/keeper"
	"cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	// flagUnsafeSkipUpgrades is a custom flag that allows the user to skip upgrades
	// It is used in a baseapp chain.
	flagUnsafeSkipUpgrades = "unsafe-skip-upgrades"
	// flagUnsafeSkipUpgradesV2 is a custom flag that allows the user to skip upgrades
	// It is used in a v2 chain.
	flagUnsafeSkipUpgradesV2 = "server.unsafe-skip-upgrades"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(&modulev1.Module{},
		appconfig.Provide(ProvideModule, ProvideConfig),
		appconfig.Invoke(PopulateVersionMap),
	)
}

func ProvideConfig(key depinject.OwnModuleKey) coreserver.ModuleConfigMap {
	return coreserver.ModuleConfigMap{
		Module: depinject.ModuleKey(key).Name(),
		Config: coreserver.ConfigMap{
			flagUnsafeSkipUpgrades:   []int{},
			flagUnsafeSkipUpgradesV2: []int{},
			flags.FlagHome:           "",
		},
	}
}

type ModuleInputs struct {
	depinject.In

	Config             *modulev1.Module
	ConfigMap          coreserver.ConfigMap
	Environment        appmodule.Environment
	Cdc                codec.Codec
	AddressCodec       address.Codec
	AppVersionModifier coreserver.VersionModifier
	ConsensusKeeper    types.ConsensusKeeper
}

type ModuleOutputs struct {
	depinject.Out

	UpgradeKeeper *keeper.Keeper
	Module        appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	var (
		homePath           string
		skipUpgradeHeights = make(map[int64]bool)
	)

	skipUpgrades, ok := in.ConfigMap[flagUnsafeSkipUpgradesV2] // check v2
	if !ok || skipUpgrades == nil {
		skipUpgrades, ok = in.ConfigMap[flagUnsafeSkipUpgrades] // check v1
		if !ok || skipUpgrades == nil {
			skipUpgrades = []int{}
		}
	}

	heights := cast.ToIntSlice(skipUpgrades) // safe to use cast here as we've handled nil case
	for _, h := range heights {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath = cast.ToString(in.ConfigMap[flags.FlagHome])

	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(types.GovModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	authorityStr, err := in.AddressCodec.BytesToString(authority)
	if err != nil {
		panic(err)
	}

	// set the governance module account as the authority for conducting upgrades
	k := keeper.NewKeeper(in.Environment, skipUpgradeHeights, in.Cdc, homePath, in.AppVersionModifier, authorityStr, in.ConsensusKeeper)
	m := NewAppModule(k)

	return ModuleOutputs{UpgradeKeeper: k, Module: m}
}

func PopulateVersionMap(upgradeKeeper *keeper.Keeper, modules map[string]appmodule.AppModule) {
	if upgradeKeeper == nil {
		return
	}

	upgradeKeeper.SetInitVersionMap(module.NewManagerFromMap(modules).GetVersionMap())
}
