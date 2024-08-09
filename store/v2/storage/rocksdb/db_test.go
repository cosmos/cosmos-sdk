//go:build rocksdb
// +build rocksdb

package rocksdb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2/storage"
)

var storeKey1 = []byte("store1")

func TestStorageTestSuite(t *testing.T) {
	s := &storage.StorageTestSuite{
		NewDB: func(dir string) (*storage.StorageStore, error) {
			db, err := New(dir)
			return storage.NewStorageStore(db, coretesting.NewNopLogger()), err
		},
		EmptyBatchSize: 12,
		SkipTests:      []string{"TestUpgradable_Prune"},
	}
	suite.Run(t, s)
}

func TestDatabase_ReverseIterator(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	batch := NewBatch(db, 1)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		require.NoError(t, batch.Set(storeKey1, []byte(key), []byte(val)))
	}

	require.NoError(t, batch.Write())

	// reverse iterator without an end key
	iter, err := db.ReverseIterator(storeKey1, 1, []byte("key000"), nil)
	require.NoError(t, err)

	defer iter.Close()

	i, count := 99, 0
	for ; iter.Valid(); iter.Next() {
		require.Equal(t, []byte(fmt.Sprintf("key%03d", i)), iter.Key())
		require.Equal(t, []byte(fmt.Sprintf("val%03d", i)), iter.Value())

		i--
		count++
	}
	require.Equal(t, 100, count)
	require.NoError(t, iter.Error())

	// seek past domain, which should make the iterator invalid and produce an error
	require.False(t, iter.Valid())

	// reverse iterator with a start and end domain
	iter2, err := db.ReverseIterator(storeKey1, 1, []byte("key010"), []byte("key019"))
	require.NoError(t, err)

	defer iter2.Close()

	i, count = 18, 0
	for ; iter2.Valid(); iter2.Next() {
		require.Equal(t, []byte(fmt.Sprintf("key%03d", i)), iter2.Key())
		require.Equal(t, []byte(fmt.Sprintf("val%03d", i)), iter2.Value())

		i--
		count++
	}
	require.Equal(t, 9, count)
	require.NoError(t, iter2.Error())

	// seek past domain, which should make the iterator invalid and produce an error
	require.False(t, iter2.Valid())

	// start must be <= end
	iter3, err := db.ReverseIterator(storeKey1, 1, []byte("key020"), []byte("key019"))
	require.Error(t, err)
	require.Nil(t, iter3)
}
