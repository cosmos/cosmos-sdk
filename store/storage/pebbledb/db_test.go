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

	for i := uint64(1); i <= 1001; i++ {
		err = db.SetLatestVersion(i)
		require.NoError(t, err)

		lv, err = db.GetLatestVersion()
		require.NoError(t, err)
		require.Equal(t, i, lv)
	}
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

	for i := uint64(1); i <= 100; i++ {
		err := db.Set(storeKey1, i, []byte("key"), []byte(fmt.Sprintf("value%03d", i)))
		require.NoError(t, err)
	}

	for i := uint64(1); i <= 100; i++ {
		bz, err := db.Get(storeKey1, i, []byte("key"))
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("value%03d", i), string(bz))
	}
}

func TestDatabase_GetVersionedKey(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	// store a key at version 1
	err = db.Set(storeKey1, 1, []byte("key"), []byte("value"))
	require.NoError(t, err)

	// assume chain progresses to version 10 w/o any changes to key
	bz, err := db.Get(storeKey1, 10, []byte("key"))
	require.NoError(t, err)
	require.Equal(t, []byte("value"), bz)

	ok, err := db.Has(storeKey1, 10, []byte("key"))
	require.NoError(t, err)
	require.True(t, ok)
}

func TestDatabase_Batch(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	batch, err := db.NewBatch(1)
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		err = batch.Set(storeKey1, []byte(fmt.Sprintf("key%03d", i)), []byte("value"))
		require.NoError(t, err)
	}

	for i := 0; i < 100; i++ {
		if i%10 == 0 {
			err = batch.Delete(storeKey1, []byte(fmt.Sprintf("key%03d", i)))
			require.NoError(t, err)
		}
	}

	require.NotZero(t, batch.Size())

	err = batch.Write()
	require.NoError(t, err)

	lv, err := db.GetLatestVersion()
	require.NoError(t, err)
	require.Equal(t, uint64(1), lv)

	for i := 0; i < 1; i++ {
		ok, err := db.Has(storeKey1, 1, []byte(fmt.Sprintf("key%03d", i)))
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

func TestDatabase_IteratorEmptyDomain(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	iter, err := db.NewIterator(storeKey1, 1, []byte{}, []byte{})
	require.Error(t, err)
	require.Nil(t, iter)
}

func TestDatabase_IteratorClose(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	iter, err := db.NewIterator(storeKey1, 1, []byte("key000"), nil)
	require.NoError(t, err)
	iter.Close()

	require.False(t, iter.Valid())
	require.Panics(t, func() { iter.Close() })
}

func TestDatabase_IteratorDomain(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	testCases := map[string]struct {
		start, end []byte
	}{
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
	itr, err := db.NewIterator(storeKey1, 1, []byte("key000"), nil)
	require.NoError(t, err)

	defer itr.Close()

	var i, count int
	for ; itr.Valid(); itr.Next() {
		require.Equal(t, []byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))
		require.Equal(t, []byte(fmt.Sprintf("val%03d", i)), itr.Value())

		i++
		count++
	}
	require.Equal(t, 100, count)
	require.NoError(t, itr.Error())

	// seek past domain, which should make the iterator invalid and produce an error
	require.False(t, itr.Next())
	require.False(t, itr.Valid())

	// iterator with with a start and end domain
	itr2, err := db.NewIterator(storeKey1, 1, []byte("key010"), []byte("key019"))
	require.NoError(t, err)

	defer itr2.Close()

	i, count = 10, 0
	for ; itr2.Valid(); itr2.Next() {
		require.Equal(t, []byte(fmt.Sprintf("key%03d", i)), itr2.Key())
		require.Equal(t, []byte(fmt.Sprintf("val%03d", i)), itr2.Value())

		i++
		count++
	}
	require.Equal(t, 9, count)
	require.NoError(t, itr2.Error())

	// seek past domain, which should make the iterator invalid and produce an error
	require.False(t, itr2.Next())
	require.False(t, itr2.Valid())

	// start must be <= end
	iter3, err := db.NewIterator(storeKey1, 1, []byte("key020"), []byte("key019"))
	require.Error(t, err)
	require.Nil(t, iter3)
}

func TestDatabase_IteratorMultiVersion(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	// for versions 1-49, set all 10 keys
	for v := uint64(1); v < 50; v++ {
		b, err := db.NewBatch(v)
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key%03d", i)
			val := fmt.Sprintf("val%03d-%03d", i, v)

			require.NoError(t, b.Set(storeKey1, []byte(key), []byte(val)))
		}

		require.NoError(t, b.Write())
	}

	// for versions 50-100, only update even keys
	for v := uint64(50); v <= 100; v++ {
		b, err := db.NewBatch(v)
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				key := fmt.Sprintf("key%03d", i)
				val := fmt.Sprintf("val%03d-%03d", i, v)

				require.NoError(t, b.Set(storeKey1, []byte(key), []byte(val)))
			}
		}

		require.NoError(t, b.Write())
	}

	itr, err := db.NewIterator(storeKey1, 69, []byte("key000"), nil)
	require.NoError(t, err)

	defer itr.Close()

	// All keys should be present; All odd keys should have a value that reflects
	// version 49, and all even keys should have a value that reflects the desired
	// version, 69.
	var i, count int
	for ; itr.Valid(); itr.Next() {
		require.Equal(t, []byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))

		if i%2 == 0 {
			require.Equal(t, []byte(fmt.Sprintf("val%03d-%03d", i, 69)), itr.Value())
		} else {
			require.Equal(t, []byte(fmt.Sprintf("val%03d-%03d", i, 49)), itr.Value())
		}

		i = (i + 1) % 10
		count++
	}
	require.Equal(t, 10, count)
	require.NoError(t, itr.Error())
}

func TestDatabase_ReverseIterator(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	require.Panics(t, func() { _, _ = db.NewReverseIterator(storeKey1, 1, []byte("key000"), nil) })
}
