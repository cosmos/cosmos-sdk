package ffi

import (
	"sync"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"google.golang.org/grpc"
)

type module struct {
	kvStoreService store.KVStoreService
}

func (m module) RegisterServices(registrar grpc.ServiceRegistrar) error {
	panic("implement me")
}

func (m module) IsOnePerModuleType() {}

func (m module) IsAppModule() {}

var _ appmodule.AppModule = module{}
var _ appmodule.HasServices = module{}

var moduleMap = &sync.Map{}

func resolveModule(moduleId uint32) *module {
	m, ok := moduleMap.Load(moduleId)
	if !ok {
		panic("invalid module")
	}
	return m.(*module)
}
