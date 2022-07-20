package store

import (
	dbm "github.com/cosmos/cosmos-sdk/db"

	"github.com/cosmos/cosmos-sdk/store/cache"
	types "github.com/cosmos/cosmos-sdk/store/v2alpha1"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/multi"
)

func NewCommitMultiStore(db dbm.Connection) types.CommitMultiStore {
	store, err := multi.NewV1MultiStoreAsV2(db, multi.DefaultStoreParams())
	if err != nil {
		panic(err)
	}
	return store
}

func NewCommitKVStoreCacheManager() types.MultiStorePersistentCache {
	return cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)
}
