package dbtest

import (
	"fmt"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/cosmos-sdk/db"
)

type Loader func(*testing.T, string) dbm.DBConnection

func ikey(i int) []byte { return []byte(fmt.Sprintf("key-%03d", i)) }
func ival(i int) []byte { return []byte(fmt.Sprintf("val-%03d", i)) }

func DoTestGetSetHasDelete(t *testing.T, load Loader) {
	t.Helper()
	db := load(t, t.TempDir())

	var txn dbm.DBReadWriter
	var view dbm.DBReader

	view = db.Reader()
	require.NotNil(t, view)

	// A nonexistent key should return nil.
	value, err := view.Get([]byte("a"))
	require.NoError(t, err)
	require.Nil(t, value)

	ok, err := view.Has([]byte("a"))
	require.NoError(t, err)
	require.False(t, ok)

	txn = db.ReadWriter()

	// Set and get a value.
	err = txn.Set([]byte("a"), []byte{0x01})
	require.NoError(t, err)

	ok, err = txn.Has([]byte("a"))
	require.NoError(t, err)
	require.True(t, ok)

	value, err = txn.Get([]byte("a"))
	require.NoError(t, err)
	require.Equal(t, []byte{0x01}, value)

	// New value is not visible from another txn.
	ok, err = view.Has([]byte("a"))
	require.NoError(t, err)
	require.False(t, ok)

	// Deleting a non-existent value is fine.
	err = txn.Delete([]byte("x"))
	require.NoError(t, err)

	// Delete a value.
	err = txn.Delete([]byte("a"))
	require.NoError(t, err)

	value, err = txn.Get([]byte("a"))
	require.NoError(t, err)
	require.Nil(t, value)

	err = txn.Set([]byte("b"), []byte{0x02})
	require.NoError(t, err)

	view.Discard()
	require.NoError(t, txn.Commit())

	txn = db.ReadWriter()

	// Verify committed values.
	value, err = txn.Get([]byte("b"))
	require.NoError(t, err)
	require.Equal(t, []byte{0x02}, value)

	ok, err = txn.Has([]byte("a"))
	require.NoError(t, err)
	require.False(t, ok)

	// Setting, getting, and deleting an empty key should error.
	_, err = txn.Get([]byte{})
	require.Equal(t, dbm.ErrKeyEmpty, err)
	_, err = txn.Get(nil)
	require.Equal(t, dbm.ErrKeyEmpty, err)

	_, err = txn.Has([]byte{})
	require.Equal(t, dbm.ErrKeyEmpty, err)
	_, err = txn.Has(nil)
	require.Equal(t, dbm.ErrKeyEmpty, err)

	err = txn.Set([]byte{}, []byte{0x01})
	require.Equal(t, dbm.ErrKeyEmpty, err)
	err = txn.Set(nil, []byte{0x01})
	require.Equal(t, dbm.ErrKeyEmpty, err)

	err = txn.Delete([]byte{})
	require.Equal(t, dbm.ErrKeyEmpty, err)
	err = txn.Delete(nil)
	require.Equal(t, dbm.ErrKeyEmpty, err)

	// Setting a nil value should error, but an empty value is fine.
	err = txn.Set([]byte("x"), nil)
	require.Equal(t, dbm.ErrValueNil, err)

	err = txn.Set([]byte("x"), []byte{})
	require.NoError(t, err)

	value, err = txn.Get([]byte("x"))
	require.NoError(t, err)
	require.Equal(t, []byte{}, value)

	require.NoError(t, txn.Commit())

	require.NoError(t, db.Close())
}

func DoTestIterators(t *testing.T, load Loader) {
	t.Helper()
	db := load(t, t.TempDir())
	type entry struct {
		key []byte
		val string
	}
	entries := []entry{
		{[]byte{0}, "0"},
		{[]byte{0, 0}, "0 0"},
		{[]byte{0, 1}, "0 1"},
		{[]byte{0, 2}, "0 2"},
		{[]byte{1}, "1"},
	}
	txn := db.ReadWriter()
	for _, e := range entries {
		require.NoError(t, txn.Set(e.key, []byte(e.val)))
	}
	require.NoError(t, txn.Commit())

	testRange := func(t *testing.T, iter dbm.Iterator, expected []string) {
		i := 0
		for ; iter.Next(); i++ {
			expectedValue := expected[i]
			value := iter.Value()
			require.EqualValues(t, expectedValue, string(value), "i=%v", i)
		}
		require.Equal(t, len(expected), i)
	}

	type testCase struct {
		start, end []byte
		expected   []string
	}

	view := db.Reader()

	iterCases := []testCase{
		{nil, nil, []string{"0", "0 0", "0 1", "0 2", "1"}},
		{[]byte{0x00}, nil, []string{"0", "0 0", "0 1", "0 2", "1"}},
		{[]byte{0x00}, []byte{0x00, 0x01}, []string{"0", "0 0"}},
		{[]byte{0x00}, []byte{0x01}, []string{"0", "0 0", "0 1", "0 2"}},
		{[]byte{0x00, 0x01}, []byte{0x01}, []string{"0 1", "0 2"}},
		{nil, []byte{0x01}, []string{"0", "0 0", "0 1", "0 2"}},
	}
	for i, tc := range iterCases {
		t.Logf("Iterator case %d: [%v, %v)", i, tc.start, tc.end)
		it, err := view.Iterator(tc.start, tc.end)
		require.NoError(t, err)
		testRange(t, it, tc.expected)
		it.Close()
	}

	reverseCases := []testCase{
		{nil, nil, []string{"1", "0 2", "0 1", "0 0", "0"}},
		{[]byte{0x00}, nil, []string{"1", "0 2", "0 1", "0 0", "0"}},
		{[]byte{0x00}, []byte{0x00, 0x01}, []string{"0 0", "0"}},
		{[]byte{0x00}, []byte{0x01}, []string{"0 2", "0 1", "0 0", "0"}},
		{[]byte{0x00, 0x01}, []byte{0x01}, []string{"0 2", "0 1"}},
		{nil, []byte{0x01}, []string{"0 2", "0 1", "0 0", "0"}},
	}
	for i, tc := range reverseCases {
		t.Logf("ReverseIterator case %d: [%v, %v)", i, tc.start, tc.end)
		it, err := view.ReverseIterator(tc.start, tc.end)
		require.NoError(t, err)
		testRange(t, it, tc.expected)
		it.Close()
	}

	view.Discard()
	require.NoError(t, db.Close())
}

func DoTestVersioning(t *testing.T, load Loader) {
	t.Helper()
	db := load(t, t.TempDir())
	view := db.Reader()
	require.NotNil(t, view)

	// Write, then read different versions
	txn := db.ReadWriter()
	require.NoError(t, txn.Set([]byte("0"), []byte("a")))
	require.NoError(t, txn.Set([]byte("1"), []byte("b")))
	require.NoError(t, txn.Commit())
	v1, err := db.SaveNextVersion()
	require.NoError(t, err)

	txn = db.ReadWriter()
	require.NoError(t, txn.Set([]byte("0"), []byte("c")))
	require.NoError(t, txn.Delete([]byte("1")))
	require.NoError(t, txn.Set([]byte("2"), []byte("c")))
	require.NoError(t, txn.Commit())
	v2, err := db.SaveNextVersion()
	require.NoError(t, err)

	// Skip to a future version
	v3 := (v2 + 2)
	require.NoError(t, db.SaveVersion(v3))

	// Try to save to a past version
	err = db.SaveVersion(v2)
	require.Error(t, err)

	// Verify existing versions
	versions, err := db.Versions()
	require.NoError(t, err)
	require.Equal(t, 3, versions.Count())
	var all []uint64
	for it := versions.Iterator(); it.Next(); {
		all = append(all, it.Value())
	}
	sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
	require.Equal(t, []uint64{v1, v2, v3}, all)
	require.Equal(t, v3, versions.Last())

	view, err = db.ReaderAt(v1)
	require.NoError(t, err)
	require.NotNil(t, view)
	val, err := view.Get([]byte("0"))
	require.Equal(t, []byte("a"), val)
	require.NoError(t, err)
	val, err = view.Get([]byte("1"))
	require.Equal(t, []byte("b"), val)
	require.NoError(t, err)
	has, err := view.Has([]byte("2"))
	require.False(t, has)

	view, err = db.ReaderAt(v2)
	require.NoError(t, err)
	require.NotNil(t, view)
	val, err = view.Get([]byte("0"))
	require.Equal(t, []byte("c"), val)
	require.NoError(t, err)
	val, err = view.Get([]byte("2"))
	require.Equal(t, []byte("c"), val)
	require.NoError(t, err)
	has, err = view.Has([]byte("1"))
	require.False(t, has)

	// Try to read an invalid version
	view, err = db.ReaderAt(versions.Last() + 1)
	require.Equal(t, dbm.ErrVersionDoesNotExist, err)

	require.NoError(t, db.DeleteVersion(v2))
	// Try to read a deleted version
	view, err = db.ReaderAt(v2)
	require.Equal(t, dbm.ErrVersionDoesNotExist, err)

	// Ensure latest version is accurate
	prev := v3
	for i := 0; i < 10; i++ {
		w := db.Writer()
		require.NoError(t, w.Set(ikey(i), ival(i)))
		require.NoError(t, w.Commit())
		ver, err := db.SaveNextVersion()
		require.NoError(t, err)
		require.Equal(t, prev+1, ver)
		versions, err := db.Versions()
		require.NoError(t, err)
		require.Equal(t, ver, versions.Last())
		prev = ver
	}

	require.NoError(t, db.Close())
}

func DoTestTransactions(t *testing.T, load Loader, multipleWriters bool) {
	t.Helper()
	db := load(t, t.TempDir())
	// Both methods should work in a DBWriter context
	writerFuncs := []func() dbm.DBWriter{
		db.Writer,
		func() dbm.DBWriter { return db.ReadWriter() },
	}

	for _, getWriter := range writerFuncs {
		// Uncommitted records are not saved
		t.Run("no commit", func(t *testing.T) {
			t.Helper()
			view := db.Reader()
			defer view.Discard()
			tx := getWriter()
			defer tx.Discard()
			require.NoError(t, tx.Set([]byte("0"), []byte("a")))
			v, err := view.Get([]byte("0"))
			require.NoError(t, err)
			require.Nil(t, v)
		})

		// Try to commit version with open txns
		t.Run("open transactions", func(t *testing.T) {
			t.Helper()
			tx := getWriter()
			tx.Set([]byte("2"), []byte("a"))
			_, err := db.SaveNextVersion()
			require.Equal(t, dbm.ErrOpenTransactions, err)
			tx.Discard()
		})

		// Continue only if the backend supports multiple concurrent writers
		if !multipleWriters {
			continue
		}

		// Writing separately to same key causes a conflict
		t.Run("write conflict", func(t *testing.T) {
			t.Helper()
			tx1 := getWriter()
			tx2 := db.ReadWriter()
			tx2.Get([]byte("1"))
			require.NoError(t, tx1.Set([]byte("1"), []byte("b")))
			require.NoError(t, tx2.Set([]byte("1"), []byte("c")))
			require.NoError(t, tx1.Commit())
			require.Error(t, tx2.Commit())
		})

		// Writing from concurrent txns
		t.Run("concurrent transactions", func(t *testing.T) {
			t.Helper()
			var wg sync.WaitGroup
			setkv := func(k, v []byte) {
				defer wg.Done()
				tx := getWriter()
				require.NoError(t, tx.Set(k, v))
				require.NoError(t, tx.Commit())
			}
			n := 10
			wg.Add(n)
			for i := 0; i < n; i++ {
				go setkv(ikey(i), ival(i))
			}
			wg.Wait()
			view := db.Reader()
			defer view.Discard()
			v, err := view.Get(ikey(0))
			require.NoError(t, err)
			require.Equal(t, ival(0), v)
		})

	}
	require.NoError(t, db.Close())
}

// Tests reloading a saved DB from disk.
func DoTestReloadDB(t *testing.T, load Loader) {
	t.Helper()
	dirname := t.TempDir()
	db := load(t, dirname)

	txn := db.Writer()
	for i := 0; i < 100; i++ {
		require.NoError(t, txn.Set(ikey(i), ival(i)))
	}
	require.NoError(t, txn.Commit())
	first, err := db.SaveNextVersion()
	require.NoError(t, err)

	txn = db.Writer()
	for i := 0; i < 50; i++ { // overwrite some values
		require.NoError(t, txn.Set(ikey(i), ival(i*10)))
	}
	require.NoError(t, txn.Commit())
	last, err := db.SaveNextVersion()
	require.NoError(t, err)

	txn = db.Writer()
	for i := 100; i < 150; i++ {
		require.NoError(t, txn.Set(ikey(i), ival(i)))
	}
	require.NoError(t, txn.Commit())
	db.Close()

	// Reload and check each saved version
	db = load(t, dirname)

	view, err := db.ReaderAt(first)
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		v, err := view.Get(ikey(i))
		require.NoError(t, err)
		require.Equal(t, ival(i), v)
	}
	view.Discard()

	view, err = db.ReaderAt(last)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		v, err := view.Get(ikey(i))
		require.NoError(t, err)
		require.Equal(t, ival(i*10), v)
	}
	for i := 50; i < 100; i++ {
		v, err := view.Get(ikey(i))
		require.NoError(t, err)
		require.Equal(t, ival(i), v)
	}
	view.Discard()

	// Load working version
	view = db.Reader()
	for i := 100; i < 150; i++ {
		v, err := view.Get(ikey(i))
		require.NoError(t, err)
		require.Equal(t, ival(i), v)
	}
	view.Discard()

	require.NoError(t, db.Close())
}
