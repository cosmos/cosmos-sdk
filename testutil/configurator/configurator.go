package configurator

import (
	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	distrmodulev1 "cosmossdk.io/api/cosmos/distribution/module/v1"
	feegrantmodulev1 "cosmossdk.io/api/cosmos/feegrant/module/v1"
	genutilmodulev1 "cosmossdk.io/api/cosmos/genutil/module/v1"
	paramsmodulev1 "cosmossdk.io/api/cosmos/params/module/v1"
	stakingmodulev1 "cosmossdk.io/api/cosmos/staking/module/v1"
	txmodulev1 "cosmossdk.io/api/cosmos/tx/module/v1"
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
	"upgrade",
	"vesting",
}

type ModuleOption func(options map[string]*appv1alpha1.ModuleConfig)

func BankModule() ModuleOption {
	return func(options map[string]*appv1alpha1.ModuleConfig) {
		options["bank"] = &appv1alpha1.ModuleConfig{
			Name:   "bank",
			Config: appconfig.WrapAny(&bankmodulev1.Module{}),
		}
	}
}

func AuthModule() ModuleOption {
	return func(options map[string]*appv1alpha1.ModuleConfig) {
		options["auth"] = &appv1alpha1.ModuleConfig{
			Name: "auth",
			Config: appconfig.WrapAny(&authmodulev1.Module{
				Bech32Prefix: "cosmos",
				ModuleAccountPermissions: []*authmodulev1.ModuleAccountPermission{
					{Account: "fee_collector"},
					{Account: "distribution"},
					{Account: "mint", Permissions: []string{"minter"}},
					{Account: "bonded_tokens_pool", Permissions: []string{"burner", "staking"}},
					{Account: "not_bonded_tokens_pool", Permissions: []string{"burner", "staking"}},
				},
			}),
		}
	}
}

func ParamsModule() ModuleOption {
	return func(options map[string]*appv1alpha1.ModuleConfig) {
		options["params"] = &appv1alpha1.ModuleConfig{
			Name:   "params",
			Config: appconfig.WrapAny(&paramsmodulev1.Module{}),
		}
	}
}

func TxModule() ModuleOption {
	return func(options map[string]*appv1alpha1.ModuleConfig) {
		options["tx"] = &appv1alpha1.ModuleConfig{
			Name:   "tx",
			Config: appconfig.WrapAny(&txmodulev1.Module{}),
		}
	}
}

func StakingModule() ModuleOption {
	return func(options map[string]*appv1alpha1.ModuleConfig) {
		options["staking"] = &appv1alpha1.ModuleConfig{
			Name:   "staking",
			Config: appconfig.WrapAny(&stakingmodulev1.Module{}),
		}
	}
}

func GenutilModule() ModuleOption {
	return func(options map[string]*appv1alpha1.ModuleConfig) {
		options["genutil"] = &appv1alpha1.ModuleConfig{
			Name:   "genutil",
			Config: appconfig.WrapAny(&genutilmodulev1.Module{}),
		}
	}
}

func DistributionModule() ModuleOption {
	return func(options map[string]*appv1alpha1.ModuleConfig) {
		options["distribution"] = &appv1alpha1.ModuleConfig{
			Name:   "distribution",
			Config: appconfig.WrapAny(&distrmodulev1.Module{}),
		}
	}
}

func FeegrantModule() ModuleOption {
	return func(options map[string]*appv1alpha1.ModuleConfig) {
		options["feegrant"] = &appv1alpha1.ModuleConfig{
			Name:   "feegrant",
			Config: appconfig.WrapAny(&feegrantmodulev1.Module{}),
		}
	}
}

func VestingModule() ModuleOption {
	return func(options map[string]*appv1alpha1.ModuleConfig) {
		options["vesting"] = &appv1alpha1.ModuleConfig{
			Name:   "vesting",
			Config: appconfig.WrapAny(&vestingmodulev1.Module{}),
		}
	}
}

func NewAppConfig(opts ...ModuleOption) depinject.Config {
	options := make(map[string]*appv1alpha1.ModuleConfig)
	for _, opt := range opts {
		opt(options)
	}

	beginBlockers := make([]string, 0)
	endBlockers := make([]string, 0)
	initGenesis := make([]string, 0)
	overrides := make([]*runtimev1alpha1.StoreKeyConfig, 0)

	for _, s := range beginBlockOrder {
		if _, ok := options[s]; ok {
			beginBlockers = append(beginBlockers, s)
		}
	}

	for _, s := range endBlockersOrder {
		if _, ok := options[s]; ok {
			endBlockers = append(endBlockers, s)
		}
	}

	for _, s := range initGenesisOrder {
		if _, ok := options[s]; ok {
			initGenesis = append(initGenesis, s)
		}
	}

	if _, ok := options["auth"]; ok {
		overrides = append(overrides, &runtimev1alpha1.StoreKeyConfig{ModuleName: "auth", KvStoreKey: "acc"})
	}

	modules := []*appv1alpha1.ModuleConfig{
		{
			Name: "runtime",
			Config: appconfig.WrapAny(&runtimev1alpha1.Module{
				AppName:           "TestApp",
				BeginBlockers:     beginBlockers,
				EndBlockers:       endBlockers,
				InitGenesis:       initGenesis,
				OverrideStoreKeys: overrides,
			}),
		},
	}

	for _, m := range options {
		modules = append(modules, m)
	}

	return appconfig.Compose(&appv1alpha1.Config{Modules: modules})
}
