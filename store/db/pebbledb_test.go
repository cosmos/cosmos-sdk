package db_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/store/v2/db"
	"github.com/stretchr/testify/require"
)

func TestPebbleDB(t *testing.T) {
	db, err := db.NewPebbleDB(t.TempDir())
	require.NoError(t, err)

	bz, err := db.Get([]byte("key001"))
	require.NoError(t, err)
	require.Nil(t, bz)

	ok, err := db.Has([]byte("key001"))
	require.NoError(t, err)
	require.False(t, ok)

	batch := db.NewBatch()
	for i := 1; i <= 100; i++ {
		require.NoError(t, batch.Set([]byte(fmt.Sprintf("key%03d", i)), []byte(fmt.Sprintf("val%03d", i))))
	}
	require.NoError(t, batch.Write())

	bz, err = db.Get([]byte("key001"))
	require.NoError(t, err)
	require.Equal(t, bz, []byte("val001"))

	ok, err = db.Has([]byte("key001"))
	require.NoError(t, err)
	require.True(t, ok)

	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	require.NotNil(t, itr)

	i := 1
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		value := itr.Value()
		require.Equal(t, key, []byte(fmt.Sprintf("key%03d", i)))
		require.Equal(t, value, []byte(fmt.Sprintf("val%03d", i)))

		i++
	}

	require.NoError(t, itr.Close())

	rItr, err := db.ReverseIterator(nil, nil)
	require.NoError(t, err)
	require.NotNil(t, rItr)

	i = 100
	for ; rItr.Valid(); rItr.Next() {
		key := rItr.Key()
		value := rItr.Value()
		require.Equal(t, key, []byte(fmt.Sprintf("key%03d", i)))
		require.Equal(t, value, []byte(fmt.Sprintf("val%03d", i)))

		i--
	}

	require.NoError(t, rItr.Close())

	require.NoError(t, db.Close())
}
