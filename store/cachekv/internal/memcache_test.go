package internal

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestGetSetDelete(t *testing.T) {
	db := NewMemCache()

	// A nonexistent key should return nil.
	value, found := db.Get([]byte("a"))
	require.Nil(t, value)
	require.False(t, found)

	// Set and get a value.
	db.Set([]byte("a"), []byte{0x01}, true)
	db.Set([]byte("b"), []byte{0x02}, false)
	value, found = db.Get([]byte("a"))
	require.Equal(t, []byte{0x01}, value)
	require.True(t, found)

	value, found = db.Get([]byte("b"))
	require.Equal(t, []byte{0x02}, value)
	require.True(t, found)

	var dirties [][]byte
	db.ScanDirtyItems(func(k, v []byte) {
		dirties = append(dirties, k)
	})
	require.Equal(t, [][]byte{[]byte("a")}, dirties)
}

func TestDBIterator(t *testing.T) {
	db := NewMemCache()

	for i := 0; i < 10; i++ {
		if i != 6 { // but skip 6.
			db.Set(int642Bytes(int64(i)), []byte{}, false)
		}
	}

	// Blank iterator keys should panic
	require.Panics(t, func() {
		db.ReverseIterator([]byte{}, nil)
	})
	require.Panics(t, func() {
		db.ReverseIterator(nil, []byte{})
	})

	itr := db.Iterator(nil, nil)
	verifyIterator(t, itr, []int64{0, 1, 2, 3, 4, 5, 7, 8, 9}, "forward iterator")

	ritr := db.ReverseIterator(nil, nil)
	verifyIterator(t, ritr, []int64{9, 8, 7, 5, 4, 3, 2, 1, 0}, "reverse iterator")

	itr = db.Iterator(nil, int642Bytes(0))
	verifyIterator(t, itr, []int64(nil), "forward iterator to 0")

	ritr = db.ReverseIterator(int642Bytes(10), nil)
	verifyIterator(t, ritr, []int64(nil), "reverse iterator from 10 (ex)")

	itr = db.Iterator(int642Bytes(0), nil)
	verifyIterator(t, itr, []int64{0, 1, 2, 3, 4, 5, 7, 8, 9}, "forward iterator from 0")

	itr = db.Iterator(int642Bytes(1), nil)
	verifyIterator(t, itr, []int64{1, 2, 3, 4, 5, 7, 8, 9}, "forward iterator from 1")

	ritr = db.ReverseIterator(nil, int642Bytes(10))
	verifyIterator(t, ritr,
		[]int64{9, 8, 7, 5, 4, 3, 2, 1, 0}, "reverse iterator from 10 (ex)")

	ritr = db.ReverseIterator(nil, int642Bytes(9))
	verifyIterator(t, ritr,
		[]int64{8, 7, 5, 4, 3, 2, 1, 0}, "reverse iterator from 9 (ex)")

	ritr = db.ReverseIterator(nil, int642Bytes(8))
	verifyIterator(t, ritr,
		[]int64{7, 5, 4, 3, 2, 1, 0}, "reverse iterator from 8 (ex)")

	itr = db.Iterator(int642Bytes(5), int642Bytes(6))
	verifyIterator(t, itr, []int64{5}, "forward iterator from 5 to 6")

	itr = db.Iterator(int642Bytes(5), int642Bytes(7))
	verifyIterator(t, itr, []int64{5}, "forward iterator from 5 to 7")

	itr = db.Iterator(int642Bytes(5), int642Bytes(8))
	verifyIterator(t, itr, []int64{5, 7}, "forward iterator from 5 to 8")

	itr = db.Iterator(int642Bytes(6), int642Bytes(7))
	verifyIterator(t, itr, []int64(nil), "forward iterator from 6 to 7")

	itr = db.Iterator(int642Bytes(6), int642Bytes(8))
	verifyIterator(t, itr, []int64{7}, "forward iterator from 6 to 8")

	itr = db.Iterator(int642Bytes(7), int642Bytes(8))
	verifyIterator(t, itr, []int64{7}, "forward iterator from 7 to 8")

	ritr = db.ReverseIterator(int642Bytes(4), int642Bytes(5))
	verifyIterator(t, ritr, []int64{4}, "reverse iterator from 5 (ex) to 4")

	ritr = db.ReverseIterator(int642Bytes(4), int642Bytes(6))
	verifyIterator(t, ritr,
		[]int64{5, 4}, "reverse iterator from 6 (ex) to 4")

	ritr = db.ReverseIterator(int642Bytes(4), int642Bytes(7))
	verifyIterator(t, ritr,
		[]int64{5, 4}, "reverse iterator from 7 (ex) to 4")

	ritr = db.ReverseIterator(int642Bytes(5), int642Bytes(6))
	verifyIterator(t, ritr, []int64{5}, "reverse iterator from 6 (ex) to 5")

	ritr = db.ReverseIterator(int642Bytes(5), int642Bytes(7))
	verifyIterator(t, ritr, []int64{5}, "reverse iterator from 7 (ex) to 5")

	ritr = db.ReverseIterator(int642Bytes(6), int642Bytes(7))
	verifyIterator(t, ritr,
		[]int64(nil), "reverse iterator from 7 (ex) to 6")

	ritr = db.ReverseIterator(int642Bytes(10), nil)
	verifyIterator(t, ritr, []int64(nil), "reverse iterator to 10")

	ritr = db.ReverseIterator(int642Bytes(6), nil)
	verifyIterator(t, ritr, []int64{9, 8, 7}, "reverse iterator to 6")

	ritr = db.ReverseIterator(int642Bytes(5), nil)
	verifyIterator(t, ritr, []int64{9, 8, 7, 5}, "reverse iterator to 5")

	ritr = db.ReverseIterator(int642Bytes(8), int642Bytes(9))
	verifyIterator(t, ritr, []int64{8}, "reverse iterator from 9 (ex) to 8")

	ritr = db.ReverseIterator(int642Bytes(2), int642Bytes(4))
	verifyIterator(t, ritr,
		[]int64{3, 2}, "reverse iterator from 4 (ex) to 2")

	ritr = db.ReverseIterator(int642Bytes(4), int642Bytes(2))
	verifyIterator(t, ritr,
		[]int64(nil), "reverse iterator from 2 (ex) to 4")

	// Ensure that the iterators don't panic with an empty database.
	db2 := NewMemCache()

	itr = db2.Iterator(nil, nil)
	verifyIterator(t, itr, nil, "forward iterator with empty db")

	ritr = db2.ReverseIterator(nil, nil)
	verifyIterator(t, ritr, nil, "reverse iterator with empty db")
}

func verifyIterator(t *testing.T, itr *memIterator, expected []int64, msg string) {
	i := 0
	for itr.Valid() {
		key := itr.Key()
		require.Equal(t, expected[i], bytes2Int64(key), "iterator: %d mismatches", i)
		itr.Next()
		i++
	}
	require.Equal(t, i, len(expected), "expected to have fully iterated over all the elements in iter")
	require.NoError(t, itr.Close())
}

func int642Bytes(i int64) []byte {
	return sdk.Uint64ToBigEndian(uint64(i))
}

func bytes2Int64(buf []byte) int64 {
	return int64(sdk.BigEndianToUint64(buf))
}
