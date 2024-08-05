//go:build rocksdb
// +build rocksdb

package commitment_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	dbm "cosmossdk.io/store/v2/db"
)

var (
	storeKeys  = []string{"store1", "store2", "store3"}
	dbBackends = map[string]func(dataDir string) (corestore.KVStoreWithBatch, error){
		"rocksdb_opts": func(dataDir string) (corestore.KVStoreWithBatch, error) {
			return dbm.NewRocksDB("test", dataDir)
		},
		"pebbledb_opts": func(dataDir string) (corestore.KVStoreWithBatch, error) {
			return dbm.NewPebbleDB("test", dataDir)
		},
		"goleveldb_opts": func(dataDir string) (corestore.KVStoreWithBatch, error) {
			return dbm.NewGoLevelDB("test", dataDir, nil)
		},
	}
	rng        = rand.New(rand.NewSource(543210))
	changesets = make([]*corestore.Changeset, 1000)
)

func init() {
	for i := 0; i < 1000; i++ {
		cs := corestore.NewChangeset()
		for _, storeKey := range storeKeys {
			for j := 0; j < 100; j++ {
				key := make([]byte, 16)
				val := make([]byte, 16)

				_, err := rng.Read(key)
				if err != nil {
					panic(err)
				}
				_, err = rng.Read(val)
				if err != nil {
					panic(err)
				}

				cs.AddKVPair([]byte(storeKey), corestore.KVPair{Key: key, Value: val})
			}
		}
		changesets[i] = cs
	}
}

func getCommitStore(b *testing.B, db corestore.KVStoreWithBatch) *commitment.CommitStore {
	b.Helper()
	multiTrees := make(map[string]commitment.Tree)
	for _, storeKey := range storeKeys {
		prefixDB := dbm.NewPrefixDB(db, []byte(storeKey))
		multiTrees[storeKey] = iavl.NewIavlTree(prefixDB, coretesting.NewNopLogger(), iavl.DefaultConfig())
	}

	sc, err := commitment.NewCommitStore(multiTrees, nil, db, coretesting.NewNopLogger())
	require.NoError(b, err)

	return sc
}

func BenchmarkCommit(b *testing.B) {
	for ty, fn := range dbBackends {
		b.Run(fmt.Sprintf("backend_%s", ty), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			b.StopTimer()
			for i := 0; i < b.N; i++ {
				db, err := fn(b.TempDir())
				require.NoError(b, err)
				sc := getCommitStore(b, db)
				b.StartTimer()
				for j, cs := range changesets {
					require.NoError(b, sc.WriteChangeset(cs))
					_, err := sc.Commit(uint64(j + 1))
					require.NoError(b, err)
				}
				b.StopTimer()
				require.NoError(b, db.Close())
			}
		})
	}
}

func BenchmarkGetProof(b *testing.B) {
	for ty, fn := range dbBackends {
		db, err := fn(b.TempDir())
		require.NoError(b, err)
		sc := getCommitStore(b, db)

		b.Run(fmt.Sprintf("backend_%s", ty), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			b.StopTimer()
			// commit some changesets
			for i, cs := range changesets {
				require.NoError(b, sc.WriteChangeset(cs))
				_, err = sc.Commit(uint64(i + 1))
				require.NoError(b, err)
			}
			b.StartTimer()

			for i := 0; i < b.N; i++ {
				// non-existing proof
				p, err := sc.GetProof([]byte(storeKeys[0]), 500, []byte("key-1-1"))
				require.NoError(b, err)
				require.NotNil(b, p)
				// existing proof
				p, err = sc.GetProof([]byte(storeKeys[1]), 500, changesets[499].Changes[1].StateChanges[1].Key)
				require.NoError(b, err)
				require.NotNil(b, p)
			}
		})
		require.NoError(b, db.Close())
	}
}
