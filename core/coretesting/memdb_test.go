package coretesting

import (
	"fmt"
	"testing"

	"cosmossdk.io/core/store"
	"github.com/stretchr/testify/require"
)

func TestMemDB(t *testing.T) {
	var db store.KVStore = newMemDB()

	key, value := []byte("key"), []byte("value")
	require.NoError(t, db.Set(key, value))
	val, err := db.Get(key)
	require.NoError(t, err)
	require.Equal(t, value, val)
	require.NoError(t, db.Delete(key))
	has, err := db.Has(key)
	require.NoError(t, err)
	require.False(t, has)

	// test iter
	makeKey := func(i int) []byte {
		return []byte(fmt.Sprintf("key_%d", i))
	}
	for i := 0; i < 10; i++ {
		require.NoError(t, db.Set(makeKey(i), makeKey(i)))
	}

	iter, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	iter.Next()
	key = iter.Key()
	value = iter.Value()
	require.Equal(t, makeKey(0), key)
	require.Equal(t, makeKey(0), value)
	require.NoError(t, iter.Error())
	iter.Next()
	key, value = iter.Key(), iter.Value()
	require.Equal(t, makeKey(1), key)
	require.Equal(t, makeKey(1), value)
	require.NoError(t, iter.Close())

	// test reverse iter
	iter, err = db.ReverseIterator(nil, nil)
	require.NoError(t, err)
	key = iter.Key()
	value = iter.Value()
	require.Equal(t, makeKey(9), key)
	require.Equal(t, makeKey(9), value)
	require.NoError(t, iter.Error())
	iter.Next()
	key, value = iter.Key(), iter.Value()
	require.Equal(t, makeKey(8), key)
	require.Equal(t, makeKey(8), value)
	require.NoError(t, iter.Close())
}
