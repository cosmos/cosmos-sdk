package store

import (
	"cosmossdk.io/store/types"
)

type (
	Store            = types.Store
	Committer        = types.Committer
	MultiStore       = types.MultiStore
	CacheMultiStore  = types.CacheMultiStore
	CommitMultiStore = types.CommitMultiStore
	KVStore          = types.KVStore
	Iterator         = types.Iterator
	CacheKVStore     = types.CacheKVStore
	CacheWrapper     = types.CacheWrapper
	CacheWrap        = types.CacheWrap
	CommitID         = types.CommitID
	Key              = types.StoreKey
	Type             = types.StoreType
	Queryable        = types.Queryable
	Gas              = types.Gas
	GasMeter         = types.GasMeter
	GasConfig        = types.GasConfig
)
