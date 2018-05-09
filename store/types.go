package store

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
)

// Import cosmos-sdk/types/store.go for convenience.
// nolint
type Store = baseapp.Store
type Committer = baseapp.Committer
type CommitStore = baseapp.CommitStore
type MultiStore = baseapp.MultiStore
type CacheMultiStore = baseapp.CacheMultiStore
type CommitMultiStore = baseapp.CommitMultiStore
type KVStore = baseapp.KVStore
type Iterator = baseapp.Iterator
type CacheKVStore = baseapp.CacheKVStore
type CommitKVStore = baseapp.CommitKVStore
type CacheWrapper = baseapp.CacheWrapper
type CacheWrap = baseapp.CacheWrap
type CommitID = baseapp.CommitID
type StoreKey = baseapp.StoreKey
type StoreType = baseapp.StoreType
type Queryable = baseapp.Queryable
