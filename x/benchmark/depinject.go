package benchmark

import "cosmossdk.io/depinject/appconfig"

func init() {
	appconfig.RegisterModule(&Module{},
		appconfig.Provide(ProvideModule),
	)
}
