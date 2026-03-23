package store

import (
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/store/cache"
	"cosmossdk.io/store/rootmulti"
	"cosmossdk.io/store/types"
)

func NewCommitMultiStore(db dbm.DB, logger log.Logger) types.CommitMultiStore {
	return rootmulti.NewStore(db, logger)
}

func NewCommitKVStoreCacheManager() types.MultiStorePersistentCache {
	return cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)
}
