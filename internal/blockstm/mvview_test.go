package blockstm

import (
	"fmt"
	"testing"

	"github.com/test-go/testify/require"

	storetypes "cosmossdk.io/store/types"
)

func TestMVMemoryViewDelete(t *testing.T) {
	stores := map[storetypes.StoreKey]int{
		StoreKeyAuth: 0,
	}
	storage := NewMultiMemDB(stores)
	mv := NewMVMemory(16, stores, storage, nil)

	mview := mv.View(0)
	view := mview.GetKVStore(StoreKeyAuth)
	view.Set(Key("a"), []byte("1"))
	view.Set(Key("b"), []byte("1"))
	view.Set(Key("c"), []byte("1"))
	require.True(t, mv.Record(TxnVersion{0, 0}, mview))

	mview = mv.View(1)
	view = mview.GetKVStore(StoreKeyAuth)
	view.Delete(Key("a"))
	view.Set(Key("b"), []byte("2"))
	require.True(t, mv.Record(TxnVersion{1, 0}, mview))

	mview = mv.View(2)
	view = mview.GetKVStore(StoreKeyAuth)
	require.Nil(t, view.Get(Key("a")))
	require.False(t, view.Has(Key("a")))
}

func TestMVMemoryViewIteration(t *testing.T) {
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0}
	storage := NewMultiMemDB(stores)
	mv := NewMVMemory(16, stores, storage, nil)
	{
		parentState := []KVPair{
			{Key("a"), []byte("1")},
			{Key("A"), []byte("1")},
		}
		parent := storage.GetKVStore(StoreKeyAuth)
		for _, kv := range parentState {
			parent.Set(kv.Key, kv.Value)
		}
	}

	sets := [][]KVPair{
		{{Key("a"), []byte("1")}, {Key("b"), []byte("1")}, {Key("c"), []byte("1")}},
		{{Key("b"), []byte("2")}, {Key("c"), []byte("2")}, {Key("d"), []byte("2")}},
		{{Key("c"), []byte("3")}, {Key("d"), []byte("3")}, {Key("e"), []byte("3")}},
		{{Key("d"), []byte("4")}, {Key("f"), []byte("4")}},
		{{Key("e"), []byte("5")}, {Key("f"), []byte("5")}, {Key("g"), []byte("5")}},
		{{Key("f"), []byte("6")}, {Key("g"), []byte("6")}, {Key("a"), []byte("6")}},
	}
	deletes := [][]Key{
		{},
		{},
		{Key("a")},
		{Key("A"), Key("e")},
		{},
		{Key("b"), Key("c"), Key("d")},
	}

	for i, pairs := range sets {
		mview := mv.View(TxnIndex(i))
		view := mview.GetKVStore(StoreKeyAuth)
		for _, kv := range pairs {
			view.Set(kv.Key, kv.Value)
		}
		for _, key := range deletes[i] {
			view.Delete(key)
		}
		require.True(t, mv.Record(TxnVersion{TxnIndex(i), 0}, mview))
	}

	testCases := []struct {
		index      TxnIndex
		start, end Key
		ascending  bool
		expect     []KVPair
	}{
		{2, nil, nil, true, []KVPair{
			{Key("A"), []byte("1")},
			{Key("a"), []byte("1")},
			{Key("b"), []byte("2")},
			{Key("c"), []byte("2")},
			{Key("d"), []byte("2")},
		}},
		{3, nil, nil, true, []KVPair{
			{Key("A"), []byte("1")},
			{Key("b"), []byte("2")},
			{Key("c"), []byte("3")},
			{Key("d"), []byte("3")},
			{Key("e"), []byte("3")},
		}},
		{3, nil, nil, false, []KVPair{
			{Key("e"), []byte("3")},
			{Key("d"), []byte("3")},
			{Key("c"), []byte("3")},
			{Key("b"), []byte("2")},
			{Key("A"), []byte("1")},
		}},
		{4, nil, nil, true, []KVPair{
			{Key("b"), []byte("2")},
			{Key("c"), []byte("3")},
			{Key("d"), []byte("4")},
			{Key("f"), []byte("4")},
		}},
		{5, nil, nil, true, []KVPair{
			{Key("b"), []byte("2")},
			{Key("c"), []byte("3")},
			{Key("d"), []byte("4")},
			{Key("e"), []byte("5")},
			{Key("f"), []byte("5")},
			{Key("g"), []byte("5")},
		}},
		{6, nil, nil, true, []KVPair{
			{Key("a"), []byte("6")},
			{Key("e"), []byte("5")},
			{Key("f"), []byte("6")},
			{Key("g"), []byte("6")},
		}},
		{6, Key("e"), Key("g"), true, []KVPair{
			{Key("e"), []byte("5")},
			{Key("f"), []byte("6")},
		}},
		{6, Key("e"), Key("g"), false, []KVPair{
			{Key("f"), []byte("6")},
			{Key("e"), []byte("5")},
		}},
		{6, Key("b"), nil, true, []KVPair{
			{Key("e"), []byte("5")},
			{Key("f"), []byte("6")},
			{Key("g"), []byte("6")},
		}},
		{6, Key("b"), nil, false, []KVPair{
			{Key("g"), []byte("6")},
			{Key("f"), []byte("6")},
			{Key("e"), []byte("5")},
		}},
		{6, nil, Key("g"), true, []KVPair{
			{Key("a"), []byte("6")},
			{Key("e"), []byte("5")},
			{Key("f"), []byte("6")},
		}},
		{6, nil, Key("g"), false, []KVPair{
			{Key("f"), []byte("6")},
			{Key("e"), []byte("5")},
			{Key("a"), []byte("6")},
		}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("version-%d", tc.index), func(t *testing.T) {
			view := mv.View(tc.index).GetKVStore(StoreKeyAuth)
			var iter storetypes.Iterator
			if tc.ascending {
				iter = view.Iterator(tc.start, tc.end)
			} else {
				iter = view.ReverseIterator(tc.start, tc.end)
			}
			require.Equal(t, tc.expect, CollectIterator(iter))
			require.NoError(t, iter.Close())
		})
	}
}

func CollectIterator[V any](iter storetypes.GIterator[V]) []GKVPair[V] {
	var res []GKVPair[V]
	for iter.Valid() {
		res = append(res, GKVPair[V]{iter.Key(), iter.Value()})
		iter.Next()
	}
	return res
}
