package store

import (
	"github.com/cosmos/cosmos-sdk/types"
)

// Import cosmos-sdk/types/store.go for convenience.
type MultiStore = types.MultiStore
type CacheMultiStore = types.CacheMultiStore
type CommitStore = types.CommitStore
type Committer = types.Committer
type CommitMultiStore = types.CommitMultiStore
type CommitStoreLoader = types.CommitStoreLoader
type KVStore = types.KVStore
type Iterator = types.Iterator
type CacheKVStore = types.CacheKVStore
type CacheWrapper = types.CacheWrapper
type CacheWrap = types.CacheWrap
type CommitID = types.CommitID
type SubstoreKey = types.SubstoreKey
