package runtime

import (
	"cosmossdk.io/core/appmodule"
	storetypes "cosmossdk.io/store/types"
)

func NewEnvironment(kvKeys []storetypes.KVStoreKey, memKeys []storetypes.MemoryStoreKey) *appmodule.Environment {
	env := &appmodule.Environment{}
	for _, storeKey := range kvKeys {
		storeKey := storeKey // TODO remove in go 1.22
		env.KvStoreService[storeKey.Name()] = NewKVStoreService(&storeKey)
	}
	for _, memKey := range memKeys {
		memKey := memKey // TODO remove in go 1.22
		env.MemStoreService[memKey.Name()] = NewMemStoreService(&memKey)
	}
	env.EventService = EventService{}
	env.HeaderService = HeaderService{}
	env.BranchService = BranchService{}
	env.GasService = GasService{}

	return env
}
