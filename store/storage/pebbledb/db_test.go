package pebbledb

import (
	"fmt"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/storage"
)

const (
	storeKey1 = "store1"
)

func TestStorageTestSuite(t *testing.T) {
	s := &storage.StorageTestSuite{
		NewDB: func(dir string) (store.VersionedDatabase, error) {
			return New(dir)
		},
		EmptyBatchSize: 12,
		SkipTests: []string{
			"TestStorageTestSuite/TestDatabase_Prune",
		},
	}
	suite.Run(t, s)
}

func TestDatabase_ReverseIterator(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	cs := new(store.Changeset)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		cs.AddKVPair(store.KVPair{StoreKey: storeKey1, Key: []byte(key), Value: []byte(val)})
	}

	// save the changeset under multiple versions
	require.NoError(t, db.ApplyChangeset(1, cs))
	require.NoError(t, db.ApplyChangeset(2, cs))

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	version := uint64(1)
	lowerBound := MVCCEncode(prependStoreKey(storeKey1, []byte("key000")), 0)

	var upperBound []byte
	upperBound = MVCCEncode(prependStoreKey(storeKey1, []byte("key099")), 0)

	itr, err := db.storage.NewIter(&pebble.IterOptions{LowerBound: lowerBound, UpperBound: upperBound})
	require.NoError(t, err)
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	if itr.Last() {
		firstKey, _, ok := SplitMVCCKey(itr.Key())
		require.True(t, ok)

		valid := itr.SeekLT(MVCCEncode(firstKey, version+1))
		require.True(t, valid)
		printIteratorCursor(t, itr)
	}
}

func printIteratorCursor(t *testing.T, itr *pebble.Iterator) {
	key, vBz, ok := SplitMVCCKey(itr.Key())
	require.True(t, ok)

	version, err := decodeUint64Ascending(vBz)
	require.NoError(t, err)

	val, tombBz, ok := SplitMVCCKey(itr.Value())
	require.True(t, ok)

	var tombstone uint64
	if len(tombBz) > 0 {
		tombstone, err = decodeUint64Ascending(vBz)
		require.NoError(t, err)
	}

	fmt.Printf("KEY: %s, VALUE: %s, VERSION: %d, TOMBSTONE: %d\n", key, val, version, tombstone)
}
