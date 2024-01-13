package runtime

import (
	"cosmossdk.io/core/appmodule"
	storetypes "cosmossdk.io/store/types"
)

func NewEnvironment(storeKey *storetypes.KVStoreKey, memKey *storetypes.MemoryStoreKey) appmodule.Environment {
	env := appmodule.Environment{}
	if storeKey != nil {
		env.KvStoreService = NewKVStoreService(storeKey)
	}

	if memKey != nil {
		env.MemStoreService = NewMemStoreService(memKey)
	}

	env.EventService = EventService{}
	env.HeaderService = HeaderService{}
	env.BranchService = BranchService{}
	env.GasService = GasService{}

	return env
}
