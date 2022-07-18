package configurator

import (
	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	paramsmodulev1 "cosmossdk.io/api/cosmos/params/module/v1"
	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
)

var beginBlockOrder = []string{
	"mint",
	"staking",
	"auth",
	"bank",
	"params",
}

var endBlockersOrder = []string{
	"staking",
	"auth",
	"bank",
	"mint",
	"params",
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

func NewAppConfig(opts ...ModuleOption) depinject.Config {
	options := make(map[string]*appv1alpha1.ModuleConfig)
	for _, opt := range opts {
		opt(options)
	}

	beginBlockers := make([]string, 0)
	endBlockers := make([]string, 0)
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
				OverrideStoreKeys: overrides,
			}),
		},
	}

	for _, m := range options {
		modules = append(modules, m)
	}

	return appconfig.Compose(&appv1alpha1.Config{Modules: modules})
}
