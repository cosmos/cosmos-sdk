package internal

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/types"
)

func KVStoreKeyProvider(scope container.Scope) *types.KVStoreKey {
	return types.NewKVStoreKey(string(scope))
}

func ConfiguratorProvider(scope container.Scope) app.Configurator {
	panic("TODO")
}
