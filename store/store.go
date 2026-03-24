package store

import (
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/store/v2/cache"
	"github.com/cosmos/cosmos-sdk/store/v2/legacy/rootmulti"
	"github.com/cosmos/cosmos-sdk/store/v2/types"
)

func NewCommitMultiStore(db dbm.DB, logger log.Logger) types.CommitMultiStore {
	return rootmulti.NewStore(db, logger)
}

func NewCommitKVStoreCacheManager() types.MultiStorePersistentCache {
	return cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)
}
