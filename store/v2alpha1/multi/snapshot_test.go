package multi

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	"github.com/cosmos/cosmos-sdk/store/types"
)

func multiStoreConfig(t *testing.T, stores int) StoreConfig {
	opts := DefaultStoreConfig()
	opts.Pruning = pruningtypes.NewPruningOptions(pruningtypes.PruningNothing)

	for i := 0; i < stores; i++ {
		sKey := types.NewKVStoreKey(fmt.Sprintf("store%d", i))
		require.NoError(t, opts.RegisterSubstore(sKey.Name(), types.StoreTypePersistent))
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
		expectType error
	}{
		"0 height": {0, snapshottypes.ErrInvalidSnapshotVersion},
		"1 height": {1, nil},
	}
	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			chunks := make(chan io.ReadCloser)
			streamWriter := snapshots.NewStreamWriter(chunks)
			err := store.Snapshot(tc.height, streamWriter)
			if tc.expectType != nil {
				assert.True(t, errors.Is(err, tc.expectType))
			}
		})
	}
}

func TestMultistoreRestore_Errors(t *testing.T) {
	store := newMultiStoreWithBasicData(t, memdb.NewDB(), 4)
	testcases := map[string]struct {
		height          uint64
		format          uint32
		expectErrorType error
	}{
		"0 height":       {0, snapshottypes.CurrentFormat, nil},
		"0 format":       {1, 0, snapshottypes.ErrUnknownFormat},
		"unknown format": {1, 9, snapshottypes.ErrUnknownFormat},
	}
	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			_, err := store.Restore(tc.height, tc.format, nil)
			require.Error(t, err)
			if tc.expectErrorType != nil {
				assert.True(t, errors.Is(err, tc.expectErrorType))
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
			"b0635a30d94d56b6cd1073fbfa109fa90b194d0ff2397659b00934c844a1f6fb",
			"8c32e05f312cf2dee6b7d2bdb41e1a2bb2372697f25504e676af1718245d8b63",
			"05dfef0e32c34ef3900300f9de51f228d7fb204fa8f4e4d0d1529f083d122029",
			"77d30aeeb427b0bdcedf3639adde1e822c15233d652782e171125280875aa492",
			"c00c3801da889ea4370f0e647ffe1e291bd47f500e2a7269611eb4cc198b993f",
			"6d565eb28776631f3e3e764decd53436c3be073a8a01fa5434afd539f9ae6eda",
		}},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(fmt.Sprintf("Format %v", tc.format), func(t *testing.T) {
			chunks := make(chan io.ReadCloser, 100)
			hashes := []string{}
			go func() {
				streamWriter := snapshots.NewStreamWriter(chunks)
				defer streamWriter.Close()
				require.NotNil(t, streamWriter)
				err := store.Snapshot(version, streamWriter)
				require.NoError(t, err)
			}()
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
	target := newMultiStore(t, memdb.NewDB(), 3)
	require.Equal(t, source.LastCommitID().Version, int64(1))
	version := uint64(source.LastCommitID().Version)
	// check for target store restore
	require.Equal(t, target.LastCommitID().Version, int64(0))

	dummyExtensionItem := snapshottypes.SnapshotItem{
		Item: &snapshottypes.SnapshotItem_Extension{
			Extension: &snapshottypes.SnapshotExtensionMeta{
				Name:   "test",
				Format: 1,
			},
		},
	}

	chunks := make(chan io.ReadCloser, 100)
	go func() {
		streamWriter := snapshots.NewStreamWriter(chunks)
		require.NotNil(t, streamWriter)
		defer streamWriter.Close()
		err := source.Snapshot(version, streamWriter)
		require.NoError(t, err)
		// write an extension metadata
		err = streamWriter.WriteMsg(&dummyExtensionItem)
		require.NoError(t, err)
	}()

	streamReader, err := snapshots.NewStreamReader(chunks)
	require.NoError(t, err)
	nextItem, err := target.Restore(version, snapshottypes.CurrentFormat, streamReader)
	require.NoError(t, err)
	require.Equal(t, *dummyExtensionItem.GetExtension(), *nextItem.GetExtension())

	assert.Equal(t, source.LastCommitID(), target.LastCommitID())

	for sKey := range source.schema {
		sourceSubStore, err := source.getSubstore(sKey)
		require.NoError(t, err)
		targetSubStore, err := target.getSubstore(sKey)
		require.NoError(t, err)
		require.Equal(t, sourceSubStore, targetSubStore)
	}

	// checking snapshot restoring for store with existed schema and without existing versions
	target3 := newMultiStore(t, memdb.NewDB(), 4)
	chunks3 := make(chan io.ReadCloser, 100)
	go func() {
		streamWriter3 := snapshots.NewStreamWriter(chunks3)
		require.NotNil(t, streamWriter3)
		defer streamWriter3.Close()
		err := source.Snapshot(version, streamWriter3)
		require.NoError(t, err)
	}()
	streamReader3, err := snapshots.NewStreamReader(chunks3)
	require.NoError(t, err)
	_, err = target3.Restore(version, snapshottypes.CurrentFormat, streamReader3)
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

		chunks := make(chan io.ReadCloser)
		go func() {
			streamWriter := snapshots.NewStreamWriter(chunks)
			require.NotNil(b, streamWriter)
			err := source.Snapshot(uint64(version), streamWriter)
			require.NoError(b, err)
		}()
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

		chunks := make(chan io.ReadCloser)
		go func() {
			writer := snapshots.NewStreamWriter(chunks)
			require.NotNil(b, writer)
			err := source.Snapshot(version, writer)
			require.NoError(b, err)
		}()

		reader, err := snapshots.NewStreamReader(chunks)
		require.NoError(b, err)
		_, err = target.Restore(version, snapshottypes.CurrentFormat, reader)
		require.NoError(b, err)
		require.Equal(b, source.LastCommitID(), target.LastCommitID())
	}
}
