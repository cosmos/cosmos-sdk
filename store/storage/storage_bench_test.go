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

func BenchmarkSet(b *testing.B) {
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
				key := make([]byte, 128)
				val := make([]byte, 128)

				_, err = rng.Read(key)
				require.NoError(b, err)
				_, err = rng.Read(val)
				require.NoError(b, err)

				b.StartTimer()
				err := db.Set(storeKey1, 1, key, val)
				require.NoError(b, err)
			}
		})
	}
}

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

		batch, err := db.NewBatch(1)
		require.NoError(b, err)

		for i := 0; i < numKeyVals; i++ {
			err = batch.Set(storeKey1, keys[i], vals[i])
			require.NoError(b, err)
		}

		require.NoError(b, batch.Write())

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

func BenchmarkBatch(b *testing.B) {
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

				batch, err := db.NewBatch(uint64(b.N + 1))
				require.NoError(b, err)

				for j := 0; j < 1000; j++ {
					b.StopTimer()
					key := make([]byte, 128)
					val := make([]byte, 128)

					_, err = rng.Read(key)
					require.NoError(b, err)
					_, err = rng.Read(val)
					require.NoError(b, err)

					b.StartTimer()
					err = batch.Set(storeKey1, key, val)
					require.NoError(b, err)
				}

				err = batch.Write()
				require.NoError(b, err)
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

		batch, err := db.NewBatch(1)
		require.NoError(b, err)

		for i := 0; i < numKeyVals; i++ {
			err = batch.Set(storeKey1, keys[i], vals[i])
			require.NoError(b, err)
		}

		require.NoError(b, batch.Write())

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
