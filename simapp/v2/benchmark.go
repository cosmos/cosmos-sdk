//go:build benchmark

package simapp

import (
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	benchmarkmodulev1 "cosmossdk.io/api/cosmos/benchmark/module/v1"
	"cosmossdk.io/depinject/appconfig"
	benchmark "cosmossdk.io/tools/benchmark/module"
)

func init() {
	// WARNING!
	// Enabling this module will produce 3M keys in the genesis state for the benchmark module.
	// Will also enable processing of benchmark transactions which can easily overwhelm the system.
	ModuleConfig.Modules = append(ModuleConfig.Modules, &appv1alpha1.ModuleConfig{
		Name: benchmark.ModuleName,
		Config: appconfig.WrapAny(&benchmarkmodulev1.Module{
			GenesisParams: &benchmarkmodulev1.GeneratorParams{
				Seed:         34,
				BucketCount:  3,
				GenesisCount: 3_000_000,
				KeyMean:      64,
				KeyStdDev:    12,
				ValueMean:    1024,
				ValueStdDev:  256,
			},
		}),
	})
}
