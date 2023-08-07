package pebbledb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	storeKey1 = "store1"
)

func TestDatabase_Close(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

func TestDatabase_LatestVersion(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	lv, err := db.GetLatestVersion()
	require.Error(t, err)
	require.Zero(t, lv)

	expected := uint64(1)

	err = db.SetLatestVersion(expected)
	require.NoError(t, err)

	lv, err = db.GetLatestVersion()
	require.NoError(t, err)
	require.Equal(t, expected, lv)
}

func TestDatabase_CRUD(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	ok, err := db.Has(storeKey1, 1, []byte("key"))
	require.NoError(t, err)
	require.False(t, ok)

	err = db.Set(storeKey1, 1, []byte("key"), []byte("value"))
	require.NoError(t, err)

	ok, err = db.Has(storeKey1, 1, []byte("key"))
	require.NoError(t, err)
	require.True(t, ok)

	val, err := db.Get(storeKey1, 1, []byte("key"))
	require.NoError(t, err)
	require.Equal(t, []byte("value"), val)

	err = db.Delete(storeKey1, 1, []byte("key"))
	require.NoError(t, err)

	ok, err = db.Has(storeKey1, 1, []byte("key"))
	require.NoError(t, err)
	require.False(t, ok)

	val, err = db.Get(storeKey1, 1, []byte("key"))
	require.NoError(t, err)
	require.Nil(t, val)
}

func TestDatabase_VersionedKeys(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	for i := 1; i <= 100; i++ {
		err := db.Set(storeKey1, uint64(i), []byte("key"), []byte(fmt.Sprintf("value%03d", i)))
		require.NoError(t, err)
	}

	for i := 1; i <= 100; i++ {
		bz, err := db.Get(storeKey1, uint64(i), []byte("key"))
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("value%03d", i), string(bz))
	}
}

func TestDatabase_Batch(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	batch, err := db.NewBatch(1)
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		err = batch.Set(storeKey1, []byte(fmt.Sprintf("key%d", i)), []byte("value"))
		require.NoError(t, err)
	}

	for i := 0; i < 100; i++ {
		if i%10 == 0 {
			err = batch.Delete(storeKey1, []byte(fmt.Sprintf("key%d", i)))
			require.NoError(t, err)
		}
	}

	require.NotZero(t, batch.Size())

	err = batch.Write()
	require.NoError(t, err)

	lv, err := db.GetLatestVersion()
	require.NoError(t, err)
	require.Equal(t, uint64(1), lv)

	for i := 0; i < 100; i++ {
		ok, err := db.Has(storeKey1, 1, []byte(fmt.Sprintf("key%d", i)))
		require.NoError(t, err)

		if i%10 == 0 {
			require.False(t, ok)
		} else {
			require.True(t, ok)
		}
	}
}

func TestDatabase_ResetBatch(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	batch, err := db.NewBatch(1)
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		err = batch.Set(storeKey1, []byte(fmt.Sprintf("key%d", i)), []byte("value"))
		require.NoError(t, err)
	}

	for i := 0; i < 100; i++ {
		if i%10 == 0 {
			err = batch.Delete(storeKey1, []byte(fmt.Sprintf("key%d", i)))
			require.NoError(t, err)
		}
	}

	require.NotZero(t, batch.Size())
	batch.Reset()
	require.NotPanics(t, func() { batch.Reset() })

	// There is an initial cost of 12 bytes for the batch header
	require.LessOrEqual(t, batch.Size(), 12)
}

func TestDatabase_IteratorDomain(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	testCases := map[string]struct {
		start, end []byte
	}{
		"empty domain": {},
		"start without end domain": {
			start: []byte("key010"),
		},
		"start and end domain": {
			start: []byte("key010"),
			end:   []byte("key020"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			iter, err := db.NewIterator(storeKey1, 1, tc.start, tc.end)
			require.NoError(t, err)

			defer iter.Close()

			start, end := iter.Domain()
			require.Equal(t, tc.start, start)
			require.Equal(t, tc.end, end)
		})
	}
}

func TestDatabase_Iterator(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	batch, err := db.NewBatch(1)
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099
		err = batch.Set(storeKey1, []byte(key), []byte(val))
		require.NoError(t, err)
	}

	err = batch.Write()
	require.NoError(t, err)

	// iterator without an end key
	iter, err := db.NewIterator(storeKey1, 1, []byte("key000"), nil)
	require.NoError(t, err)

	defer iter.Close()

	var i, count int
	for ; iter.Valid(); iter.Next() {
		require.Equal(t, []byte(fmt.Sprintf("key%03d", i)), iter.Key())
		require.Equal(t, []byte(fmt.Sprintf("val%03d", i)), iter.Value())

		i++
		count++
	}
	require.Equal(t, 100, count)

	// iterator with with a start and end domain
	iter2, err := db.NewIterator(storeKey1, 1, []byte("key010"), []byte("key019"))
	require.NoError(t, err)

	defer iter2.Close()

	i, count = 10, 0
	for ; iter2.Valid(); iter2.Next() {
		require.Equal(t, []byte(fmt.Sprintf("key%03d", i)), iter2.Key())
		require.Equal(t, []byte(fmt.Sprintf("val%03d", i)), iter2.Value())

		i++
		count++
	}
	require.Equal(t, 9, count)

	// start must be <= end
	iter3, err := db.NewIterator(storeKey1, 1, []byte("key020"), []byte("key019"))
	require.Error(t, err)
	require.Nil(t, iter3)
}
