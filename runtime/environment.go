package runtime

import (
	"cosmossdk.io/core/appmodule"
	storetypes "cosmossdk.io/store/types"
)

func NewEnvironment(kvKeys []storetypes.KVStoreKey, memKeys []storetypes.MemoryStoreKey) *appmodule.Environment {
	env := &appmodule.Environment{}
	for _, storeKey := range kvKeys {
		env.KvStoreService[storeKey.Name()] = NewKVStoreService(&storeKey)
	}
	for _, memKey := range memKeys {
		env.MemStoreService[memKey.Name()] = NewMemStoreService(&memKey)
	}
	env.EventService = EventService{}
	env.HeaderService = HeaderService{}
	env.BranchService = BranchService{}

	//   GasService      gas.Service

	return env
}
