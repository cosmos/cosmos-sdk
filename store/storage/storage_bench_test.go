//go:build rocksdb
// +build rocksdb

package storage

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/storage/pebbledb"
	"cosmossdk.io/store/v2/storage/rocksdb"
	"cosmossdk.io/store/v2/storage/sqlite"
	"github.com/stretchr/testify/require"
)

var (
	backends = map[string]func(dataDir string) (store.VersionedDatabase, error){
		"rocksdb_versiondb_opts": func(dataDir string) (store.VersionedDatabase, error) {
			return rocksdb.New(dataDir)
		},
		"pebbledb_default_opts": func(dataDir string) (store.VersionedDatabase, error) {
			return pebbledb.New(dataDir)
		},
		"btree_sqlite": func(dataDir string) (store.VersionedDatabase, error) {
			return sqlite.New(dataDir)
		},
	}
	rng = rand.New(rand.NewSource(567320))
)

func BenchmarkGet(b *testing.B) {
	numKeyVals := 1_000_000
	keys := make([][]byte, numKeyVals)
	vals := make([][]byte, numKeyVals)
	for i := 0; i < numKeyVals; i++ {
		key := make([]byte, 128)
		val := make([]byte, 128)

		_, err := rng.Read(key)
		require.NoError(b, err)
		_, err = rng.Read(val)
		require.NoError(b, err)

		keys[i] = key
		vals[i] = val
	}

	for ty, fn := range backends {
		db, err := fn(b.TempDir())
		require.NoError(b, err)
		defer func() {
			_ = db.Close()
		}()

		cs := new(store.Changeset)
		for i := 0; i < numKeyVals; i++ {
			cs.AddKVPair(store.KVPair{StoreKey: storeKey1, Key: keys[i], Value: vals[i]})
		}

		require.NoError(b, db.ApplyChangeset(1, cs))

		b.Run(fmt.Sprintf("backend_%s", ty), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				key := keys[rng.Intn(len(keys))]

				b.StartTimer()
				_, err = db.Get(storeKey1, 1, key)
				require.NoError(b, err)
			}
		})
	}
}

func BenchmarkApplyChangeset(b *testing.B) {
	for ty, fn := range backends {
		db, err := fn(b.TempDir())
		require.NoError(b, err)
		defer func() {
			_ = db.Close()
		}()

		b.Run(fmt.Sprintf("backend_%s", ty), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()

				cs := new(store.Changeset)
				for j := 0; j < 1000; j++ {
					key := make([]byte, 128)
					val := make([]byte, 128)

					_, err = rng.Read(key)
					require.NoError(b, err)
					_, err = rng.Read(val)
					require.NoError(b, err)

					cs.AddKVPair(store.KVPair{StoreKey: storeKey1, Key: key, Value: val})
				}

				b.StartTimer()
				require.NoError(b, db.ApplyChangeset(uint64(b.N+1), cs))
			}
		})
	}
}

func BenchmarkIterate(b *testing.B) {
	numKeyVals := 1_000_000
	keys := make([][]byte, numKeyVals)
	vals := make([][]byte, numKeyVals)
	for i := 0; i < numKeyVals; i++ {
		key := make([]byte, 128)
		val := make([]byte, 128)

		_, err := rng.Read(key)
		require.NoError(b, err)
		_, err = rng.Read(val)
		require.NoError(b, err)

		keys[i] = key
		vals[i] = val

	}

	for ty, fn := range backends {
		db, err := fn(b.TempDir())
		require.NoError(b, err)
		defer func() {
			_ = db.Close()
		}()

		b.StopTimer()

		cs := new(store.Changeset)
		for i := 0; i < numKeyVals; i++ {
			cs.AddKVPair(store.KVPair{StoreKey: storeKey1, Key: keys[i], Value: vals[i]})
		}

		require.NoError(b, db.ApplyChangeset(1, cs))

		sort.Slice(keys, func(i, j int) bool {
			return bytes.Compare(keys[i], keys[j]) < 0
		})

		b.Run(fmt.Sprintf("backend_%s", ty), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()

				itr, err := db.Iterator(storeKey1, 1, keys[0], nil)
				require.NoError(b, err)

				b.StartTimer()

				for ; itr.Valid(); itr.Next() {
					_ = itr.Key()
					_ = itr.Value()
				}

				require.NoError(b, itr.Error())
			}
		})
	}
}
