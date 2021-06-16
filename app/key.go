package app

import "github.com/cosmos/cosmos-sdk/types"

type KVStoreKeyProvider func(ModuleKey) *types.KVStoreKey
type TransientStoreKeyProvider func(ModuleKey) *types.TransientStoreKey
