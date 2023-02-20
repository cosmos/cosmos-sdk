package store

import (
	"github.com/cometbft/cometbft/libs/log"
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/store/cache"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	"cosmossdk.io/store/types"
)

func NewCommitMultiStore(db dbm.DB, logger log.Logger, metricGatherer metrics.StoreMetrics) types.CommitMultiStore {
	return rootmulti.NewStore(db, logger, metricGatherer)
}

func NewCommitKVStoreCacheManager() types.MultiStorePersistentCache {
	return cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)
}
