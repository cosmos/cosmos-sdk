package store

import (
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/store/v2/cache"
	"github.com/cosmos/cosmos-sdk/store/v2/metrics"
	"github.com/cosmos/cosmos-sdk/store/v2/rootmulti"
	"github.com/cosmos/cosmos-sdk/store/v2/types"
)

func NewCommitMultiStore(db dbm.DB, logger log.Logger, metricGatherer metrics.StoreMetrics) types.CommitMultiStore {
	return rootmulti.NewStore(db, logger, metricGatherer)
}

func NewCommitKVStoreCacheManager() types.MultiStorePersistentCache {
	return cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)
}
