package appmodule

import "cosmossdk.io/depinject"

type AppModule interface {
	depinject.OnePerModuleType
	IsAppModule()
}
