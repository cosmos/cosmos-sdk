package benchmark

import "cosmossdk.io/core/appmodule"

var _ appmodule.AppModule = AppModule{}

type AppModule struct {
	collector *KVServiceCollector
}

func (a AppModule) IsOnePerModuleType() {}

func (a AppModule) IsAppModule() {}
