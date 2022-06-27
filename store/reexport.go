package store

import (
	v1 "github.com/cosmos/cosmos-sdk/store/types"
	v2 "github.com/cosmos/cosmos-sdk/store/v2alpha1"
)

// Import cosmos-sdk/types/store.go for convenience.
type (
	Store            = v1.Store
	Committer        = v1.Committer
	CommitStore      = v1.CommitStore
	MultiStore       = v2.MultiStore
	CacheMultiStore  = v2.CacheMultiStore
	CommitMultiStore = v2.CommitMultiStore
	KVStore          = v1.KVStore
	KVPair           = v1.KVPair
	Iterator         = v1.Iterator
	CacheKVStore     = v1.CacheKVStore
	CommitKVStore    = v1.CommitKVStore
	CacheWrapper     = v1.CacheWrapper
	CacheWrap        = v1.CacheWrap
	CommitID         = v1.CommitID
	Key              = v1.StoreKey
	Type             = v1.StoreType
	Queryable        = v1.Queryable
	TraceContext     = v1.TraceContext
	Gas              = v1.Gas
	GasMeter         = v1.GasMeter
	GasConfig        = v1.GasConfig
)
