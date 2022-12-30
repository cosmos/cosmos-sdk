package store

import (
	"bytes"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store/cache"
	"github.com/cosmos/cosmos-sdk/store/metrics"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

func NewCommitMultiStore(db dbm.DB, logger log.Logger, metricGatherer metrics.StoreMetrics) types.CommitMultiStore {
	return rootmulti.NewStore(db, logger, metricGatherer)
}

func NewCommitKVStoreCacheManager() types.MultiStorePersistentCache {
	return cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)
}

// DecoderRegistry defines each of the modules store decoders. Used for ImportExport
// simulation.
type DecoderRegistry map[string]func(kvA, kvB kv.Pair) string

// DiffKVStores compares two KVstores and returns all the key/value pairs
// that differ from one another. It also skips value comparison for a set of provided prefixes.
func DiffKVStores(a KVStore, b KVStore, prefixesToSkip [][]byte) (kvAs, kvBs []kv.Pair) {
	iterA := a.Iterator(nil, nil)

	defer iterA.Close()

	iterB := b.Iterator(nil, nil)

	defer iterB.Close()

	for {
		if !iterA.Valid() && !iterB.Valid() {
			return kvAs, kvBs
		}

		var kvA, kvB kv.Pair
		if iterA.Valid() {
			kvA = kv.Pair{Key: iterA.Key(), Value: iterA.Value()}

			iterA.Next()
		}

		if iterB.Valid() {
			kvB = kv.Pair{Key: iterB.Key(), Value: iterB.Value()}
		}

		compareValue := true

		for _, prefix := range prefixesToSkip {
			// Skip value comparison if we matched a prefix
			if bytes.HasPrefix(kvA.Key, prefix) {
				compareValue = false
				break
			}
		}

		if !compareValue {
			// We're skipping this key due to an exclusion prefix.  If it's present in B, iterate past it.  If it's
			// absent don't iterate.
			if bytes.Equal(kvA.Key, kvB.Key) {
				iterB.Next()
			}
			continue
		}

		// always iterate B when comparing
		iterB.Next()

		if !bytes.Equal(kvA.Key, kvB.Key) || !bytes.Equal(kvA.Value, kvB.Value) {
			kvAs = append(kvAs, kvA)
			kvBs = append(kvBs, kvB)
		}
	}
}
