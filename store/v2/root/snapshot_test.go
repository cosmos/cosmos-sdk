package root

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"math/rand"
	"sort"
	"strings"
	"testing"
)

func multiStoreConfig(t *testing.T, stores int) StoreConfig {
	opts := DefaultStoreConfig()
	opts.Pruning = types.PruneNothing

	for i := 0; i < stores; i++ {
		sKey := types.NewKVStoreKey(fmt.Sprintf("store%d", i))
		require.NoError(t, opts.ReservePrefix(sKey.Name(), types.StoreTypePersistent))
	}

	return opts
}

func newMultiStoreWithGeneratedData(t *testing.T, db dbm.DBConnection, stores int, storeKeys uint64) *Store {
	cfg := multiStoreConfig(t, stores)
	store, err := NewStore(db, cfg)
	require.NoError(t, err)
	r := rand.New(rand.NewSource(49872768940)) // Fixed seed for deterministic tests

	var sKeys []string
	for sKey := range store.schema {
		sKeys = append(sKeys, sKey)
	}

	sort.Slice(sKeys, func(i, j int) bool {
		return strings.Compare(sKeys[i], sKeys[j]) == -1
	})

	for _, sKey := range sKeys {
		sStore, err := store.getSubstore(sKey)
		require.NoError(t, err)
		for i := uint64(0); i < storeKeys; i++ {
			k := make([]byte, 8)
			v := make([]byte, 1024)
			binary.BigEndian.PutUint64(k, i)
			_, err := r.Read(v)
			if err != nil {
				panic(err)
			}
			sStore.Set(k, v)
		}
	}
	store.Commit()
	return store
}

func newMultiStoreWithBasicData(t *testing.T, db dbm.DBConnection, stores int) *Store {
	cfg := multiStoreConfig(t, stores)
	store, err := NewStore(db, cfg)
	require.NoError(t, err)

	for sKey := range store.schema {
		sStore, err := store.getSubstore(sKey)
		require.NoError(t, err)
		for k, v := range alohaData {
			sStore.Set([]byte(k), []byte(v))
		}
	}

	store.Commit()
	return store
}

func newMultiStore(t *testing.T, db dbm.DBConnection, stores int) *Store {
	cfg := multiStoreConfig(t, stores)
	store, err := NewStore(db, cfg)
	require.NoError(t, err)
	return store
}

func TestMultistoreSnapshot_Errors(t *testing.T) {
	store := newMultiStoreWithBasicData(t, memdb.NewDB(), 4)
	testcases := map[string]struct {
		height     uint64
		format     uint32
		expectType error
	}{
		"0 height":       {0, snapshottypes.CurrentFormat, nil},
		"0 format":       {1, 0, snapshottypes.ErrUnknownFormat},
		"unknown format": {1, 9, snapshottypes.ErrUnknownFormat},
	}
	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			_, err := store.Snapshot(tc.height, tc.format)
			require.Error(t, err)
			if tc.expectType != nil {
				assert.True(t, errors.Is(err, tc.expectType))
			}
		})
	}
}

func TestMultistoreRestore_Errors(t *testing.T) {
	store := newMultiStoreWithBasicData(t, memdb.NewDB(), 4)
	testcases := map[string]struct {
		height     uint64
		format     uint32
		expectType error
	}{
		"0 height":       {0, snapshottypes.CurrentFormat, nil},
		"0 format":       {1, 0, snapshottypes.ErrUnknownFormat},
		"unknown format": {1, 9, snapshottypes.ErrUnknownFormat},
	}
	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			err := store.Restore(tc.height, tc.format, nil, nil)
			require.Error(t, err)
			if tc.expectType != nil {
				assert.True(t, errors.Is(err, tc.expectType))
			}
		})
	}
}

func TestMultistoreSnapshot_Checksum(t *testing.T) {
	store := newMultiStoreWithGeneratedData(t, memdb.NewDB(), 5, 10000)
	version := uint64(store.LastCommitID().Version)

	testcases := []struct {
		format      uint32
		chunkHashes []string
	}{
		{1, []string{
			"28b9dd52156e7c46f42d6c2b390350be3c635a54446f6a6a553e1a6ecca5efca",
			"8c32e05f312cf2dee6b7d2bdb41e1a2bb2372697f25504e676af1718245d8b63",
			"05dfef0e32c34ef3900300f9de51f228d7fb204fa8f4e4d0d1529f083d122029",
			"77d30aeeb427b0bdcedf3639adde1e822c15233d652782e171125280875aa492",
			"c00c3801da889ea4370f0e647ffe1e291bd47f500e2a7269611eb4cc198b993f",
			"3af4440d732225317644fa814dd8c0fb52adb7bf9046631af092af2c8cf9b512",
		}},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(fmt.Sprintf("Format %v", tc.format), func(t *testing.T) {
			chunks, err := store.Snapshot(version, tc.format)
			require.NoError(t, err)
			hashes := []string{}
			hasher := sha256.New()
			for chunk := range chunks {
				hasher.Reset()
				_, err := io.Copy(hasher, chunk)
				require.NoError(t, err)
				hashes = append(hashes, hex.EncodeToString(hasher.Sum(nil)))
			}
			assert.Equal(t, tc.chunkHashes, hashes, "Snapshot output for format %v has changed", tc.format)
		})
	}
}

func TestMultistoreSnapshotRestore(t *testing.T) {
	source := newMultiStoreWithGeneratedData(t, memdb.NewDB(), 3, 4)
	target := newMultiStore(t, memdb.NewDB(), 0)
	require.Equal(t, source.LastCommitID().Version, int64(1))
	version := uint64(source.LastCommitID().Version)
	// check for target store restore
	require.Equal(t, target.LastCommitID().Version, int64(0))

	chunks, err := source.Snapshot(version, snapshottypes.CurrentFormat)
	require.NoError(t, err)
	ready := make(chan struct{})
	err = target.Restore(version, snapshottypes.CurrentFormat, chunks, ready)
	require.NoError(t, err)
	assert.EqualValues(t, struct{}{}, <-ready)

	assert.Equal(t, source.LastCommitID(), target.LastCommitID())

	for sKey := range source.schema {
		sourceSubStore, err := source.getSubstore(sKey)
		require.NoError(t, err)
		targetSubStore, err := target.getSubstore(sKey)
		require.NoError(t, err)
		require.Equal(t, sourceSubStore, targetSubStore)
	}

	// checking snapshot restore for store with existing saved version
	target2 := newMultiStoreWithBasicData(t, memdb.NewDB(), 0)
	ready2 := make(chan struct{})
	err = target2.Restore(version, snapshottypes.CurrentFormat, chunks, ready2)
	require.Error(t, err)

	// checking snapshot restoring for store with existed schema and without existing versions
	target3 := newMultiStore(t, memdb.NewDB(), 4)
	ready3 := make(chan struct{})
	chunks, err = source.Snapshot(version, snapshottypes.CurrentFormat)
	require.NoError(t, err)
	err = target3.Restore(version, snapshottypes.CurrentFormat, chunks, ready3)
	require.Error(t, err)
}

func BenchmarkMultistoreSnapshot100K(b *testing.B) {
	benchmarkMultistoreSnapshot(b, 10, 10000)
}

func BenchmarkMultistoreSnapshot1M(b *testing.B) {
	benchmarkMultistoreSnapshot(b, 10, 100000)
}

func BenchmarkMultistoreSnapshotRestore100K(b *testing.B) {
	benchmarkMultistoreSnapshotRestore(b, 10, 10000)
}

func BenchmarkMultistoreSnapshotRestore1M(b *testing.B) {
	benchmarkMultistoreSnapshotRestore(b, 10, 100000)
}

func benchmarkMultistoreSnapshot(b *testing.B, stores int, storeKeys uint64) {
	b.Skip("Noisy with slow setup time, please see https://github.com/cosmos/cosmos-sdk/issues/8855.")

	b.ReportAllocs()
	b.StopTimer()
	source := newMultiStoreWithGeneratedData(nil, memdb.NewDB(), stores, storeKeys)

	version := source.LastCommitID().Version
	require.EqualValues(b, 1, version)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		target := newMultiStore(nil, memdb.NewDB(), stores)
		require.EqualValues(b, 0, target.LastCommitID().Version)

		chunks, err := source.Snapshot(uint64(version), snapshottypes.CurrentFormat)
		require.NoError(b, err)
		for reader := range chunks {
			_, err := io.Copy(io.Discard, reader)
			require.NoError(b, err)
			err = reader.Close()
			require.NoError(b, err)
		}
	}
}

func benchmarkMultistoreSnapshotRestore(b *testing.B, stores int, storeKeys uint64) {
	b.Skip("Noisy with slow setup time, please see https://github.com/cosmos/cosmos-sdk/issues/8855.")

	b.ReportAllocs()
	b.StopTimer()
	source := newMultiStoreWithGeneratedData(nil, memdb.NewDB(), stores, storeKeys)
	version := uint64(source.LastCommitID().Version)
	require.EqualValues(b, 1, version)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		target := newMultiStore(nil, memdb.NewDB(), stores)
		require.EqualValues(b, 0, target.LastCommitID().Version)

		chunks, err := source.Snapshot(version, snapshottypes.CurrentFormat)
		require.NoError(b, err)
		err = target.Restore(version, snapshottypes.CurrentFormat, chunks, nil)
		require.NoError(b, err)
		require.Equal(b, source.LastCommitID(), target.LastCommitID())
	}
}
