package configurator

import (
	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	consensusmodulev1 "cosmossdk.io/api/cosmos/consensus/module/v1"
	distrmodulev1 "cosmossdk.io/api/cosmos/distribution/module/v1"
	feegrantmodulev1 "cosmossdk.io/api/cosmos/feegrant/module/v1"
	genutilmodulev1 "cosmossdk.io/api/cosmos/genutil/module/v1"
	govmodulev1 "cosmossdk.io/api/cosmos/gov/module/v1"
	mintmodulev1 "cosmossdk.io/api/cosmos/mint/module/v1"
	paramsmodulev1 "cosmossdk.io/api/cosmos/params/module/v1"
	slashingmodulev1 "cosmossdk.io/api/cosmos/slashing/module/v1"
	stakingmodulev1 "cosmossdk.io/api/cosmos/staking/module/v1"
	txconfigv1 "cosmossdk.io/api/cosmos/tx/config/v1"
	vestingmodulev1 "cosmossdk.io/api/cosmos/vesting/module/v1"
	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
)

var beginBlockOrder = []string{
	"upgrade",
	"capability",
	"mint",
	"distribution",
	"slashing",
	"evidence",
	"staking",
	"auth",
	"bank",
	"gov",
	"crisis",
	"genutil",
	"authz",
	"feegrant",
	"nft",
	"group",
	"params",
	"consensus",
	"vesting",
}

var endBlockersOrder = []string{
	"crisis",
	"gov",
	"staking",
	"capability",
	"auth",
	"bank",
	"distribution",
	"slashing",
	"mint",
	"genutil",
	"evidence",
	"authz",
	"feegrant",
	"nft",
	"group",
	"params",
	"consensus",
	"upgrade",
	"vesting",
}

var initGenesisOrder = []string{
	"capability",
	"auth",
	"bank",
	"distribution",
	"staking",
	"slashing",
	"gov",
	"mint",
	"crisis",
	"genutil",
	"evidence",
	"authz",
	"feegrant",
	"nft",
	"group",
	"params",
	"consensus",
	"upgrade",
	"vesting",
}

type appConfig struct {
	moduleConfigs  map[string]*appv1alpha1.ModuleConfig
	setInitGenesis bool
}

type ModuleOption func(config *appConfig)

func BankModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["bank"] = &appv1alpha1.ModuleConfig{
			Name:   "bank",
			Config: appconfig.WrapAny(&bankmodulev1.Module{}),
		}
	}
}

func AuthModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["auth"] = &appv1alpha1.ModuleConfig{
			Name: "auth",
			Config: appconfig.WrapAny(&authmodulev1.Module{
				Bech32Prefix: "cosmos",
				ModuleAccountPermissions: []*authmodulev1.ModuleAccountPermission{
					{Account: "fee_collector"},
					{Account: "distribution"},
					{Account: "mint", Permissions: []string{"minter"}},
					{Account: "bonded_tokens_pool", Permissions: []string{"burner", "staking"}},
					{Account: "not_bonded_tokens_pool", Permissions: []string{"burner", "staking"}},
					{Account: "gov", Permissions: []string{"burner"}},
				},
			}),
		}
	}
}

func ParamsModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["params"] = &appv1alpha1.ModuleConfig{
			Name:   "params",
			Config: appconfig.WrapAny(&paramsmodulev1.Module{}),
		}
	}
}

func TxModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["tx"] = &appv1alpha1.ModuleConfig{
			Name:   "tx",
			Config: appconfig.WrapAny(&txconfigv1.Config{}),
		}
	}
}

func StakingModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["staking"] = &appv1alpha1.ModuleConfig{
			Name:   "staking",
			Config: appconfig.WrapAny(&stakingmodulev1.Module{}),
		}
	}
}

func SlashingModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["slashing"] = &appv1alpha1.ModuleConfig{
			Name:   "slashing",
			Config: appconfig.WrapAny(&slashingmodulev1.Module{}),
		}
	}
}

func GenutilModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["genutil"] = &appv1alpha1.ModuleConfig{
			Name:   "genutil",
			Config: appconfig.WrapAny(&genutilmodulev1.Module{}),
		}
	}
}

func DistributionModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["distribution"] = &appv1alpha1.ModuleConfig{
			Name:   "distribution",
			Config: appconfig.WrapAny(&distrmodulev1.Module{}),
		}
	}
}

func FeegrantModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["feegrant"] = &appv1alpha1.ModuleConfig{
			Name:   "feegrant",
			Config: appconfig.WrapAny(&feegrantmodulev1.Module{}),
		}
	}
}

func VestingModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["vesting"] = &appv1alpha1.ModuleConfig{
			Name:   "vesting",
			Config: appconfig.WrapAny(&vestingmodulev1.Module{}),
		}
	}
}

func GovModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["gov"] = &appv1alpha1.ModuleConfig{
			Name:   "gov",
			Config: appconfig.WrapAny(&govmodulev1.Module{}),
		}
	}
}

func ConsensusModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["consensus"] = &appv1alpha1.ModuleConfig{
			Name:   "consensus",
			Config: appconfig.WrapAny(&consensusmodulev1.Module{}),
		}
	}
}

func MintModule() ModuleOption {
	return func(config *appConfig) {
		config.moduleConfigs["mint"] = &appv1alpha1.ModuleConfig{
			Name:   "mint",
			Config: appconfig.WrapAny(&mintmodulev1.Module{}),
			GolangBindings: []*appv1alpha1.GolangBinding{
				{
					InterfaceType:  "github.com/cosmos/cosmos-sdk/x/mint/types/types.StakingKeeper",
					Implementation: "github.com/cosmos/cosmos-sdk/x/staking/keeper/*keeper.Keeper",
				},
			},
		}
	}
}

func OmitInitGenesis() ModuleOption {
	return func(config *appConfig) {
		config.setInitGenesis = false
	}
}

func NewAppConfig(opts ...ModuleOption) depinject.Config {
	cfg := &appConfig{
		moduleConfigs:  make(map[string]*appv1alpha1.ModuleConfig),
		setInitGenesis: true,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	beginBlockers := make([]string, 0)
	endBlockers := make([]string, 0)
	initGenesis := make([]string, 0)
	overrides := make([]*runtimev1alpha1.StoreKeyConfig, 0)

	for _, s := range beginBlockOrder {
		if _, ok := cfg.moduleConfigs[s]; ok {
			beginBlockers = append(beginBlockers, s)
		}
	}

	for _, s := range endBlockersOrder {
		if _, ok := cfg.moduleConfigs[s]; ok {
			endBlockers = append(endBlockers, s)
		}
	}

	for _, s := range initGenesisOrder {
		if _, ok := cfg.moduleConfigs[s]; ok {
			initGenesis = append(initGenesis, s)
		}
	}

	if _, ok := cfg.moduleConfigs["auth"]; ok {
		overrides = append(overrides, &runtimev1alpha1.StoreKeyConfig{ModuleName: "auth", KvStoreKey: "acc"})
	}

	runtimeConfig := &runtimev1alpha1.Module{
		AppName:           "TestApp",
		BeginBlockers:     beginBlockers,
		EndBlockers:       endBlockers,
		OverrideStoreKeys: overrides,
	}
	if cfg.setInitGenesis {
		runtimeConfig.InitGenesis = initGenesis
	}

	modules := []*appv1alpha1.ModuleConfig{{
		Name:   "runtime",
		Config: appconfig.WrapAny(runtimeConfig),
	}}

	for _, m := range cfg.moduleConfigs {
		modules = append(modules, m)
	}

	return appconfig.Compose(&appv1alpha1.Config{Modules: modules})
}
