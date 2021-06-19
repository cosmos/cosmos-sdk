package store

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/store/types"
	"go.uber.org/dig"
)

type Inputs struct {
	dig.In
}

type Outputs struct {
	dig.Out

	app.KVStoreKeyProvider
	app.TransientStoreKeyProvider
	app.MemoryStoreKeyProvider

	BaseAppOption func(*baseapp.BaseApp) `group:"init"`
}

func provider() Outputs {
	kvStoreKeys := map[string]*types.KVStoreKey{}
	transientKeys := map[string]*types.TransientStoreKey{}
	memKeys := map[string]*types.MemoryStoreKey{}

	return Outputs{
		KVStoreKeyProvider: func(key app.ModuleKey) *types.KVStoreKey {
			name := key.ID().Name()
			storeKey := types.NewKVStoreKey(name)
			kvStoreKeys[name] = storeKey
			return storeKey
		},
		TransientStoreKeyProvider: func(key app.ModuleKey) *types.TransientStoreKey {
			name := key.ID().Name()
			storeKey := types.NewTransientStoreKey(name)
			transientKeys[name] = storeKey
			return storeKey
		},
		MemoryStoreKeyProvider: func(key app.ModuleKey) *types.MemoryStoreKey {
			name := key.ID().Name()
			storeKey := types.NewMemoryStoreKey(name)
			memKeys[name] = storeKey
			return storeKey
		},
		BaseAppOption: func(baseApp *baseapp.BaseApp) {
			baseApp.MountKVStores(kvStoreKeys)
			baseApp.MountTransientStores(transientKeys)
			baseApp.MountMemoryStores(memKeys)
		},
	}
}

var Module = container.Provide(provider)
