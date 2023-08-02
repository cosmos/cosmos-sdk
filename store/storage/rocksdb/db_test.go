package rocksdb

import (
	"fmt"
	"sort"
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

	lv, err := db.GetLatestVersion()
	require.NoError(t, err)
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

func TestDatabase_Batch(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)

	batch := db.NewBatch(1)

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

	batch := db.NewBatch(1)

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

func TestDatabase_StartIterator(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)

	batch := db.NewBatch(1)

	keys := make([]string, 100)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		err = batch.Set(storeKey1, []byte(key), []byte("value"))
		require.NoError(t, err)

		keys[i] = key
	}

	sort.Strings(keys)

	err = batch.Write()
	require.NoError(t, err)

	iter, err := db.NewIterator(storeKey1, 1, []byte("key0"), nil)
	require.NoError(t, err)

	defer iter.Close()

	var i int
	for ; iter.Valid(); iter.Next() {
		require.Equal(t, []byte(keys[i]), iter.Key())
		require.Equal(t, []byte("value"), iter.Value())
		i++
	}
}
