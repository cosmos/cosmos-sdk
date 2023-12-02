package cgo

import (
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
