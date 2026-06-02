package blockstm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/test-go/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

func TestMVMemoryViewDelete(t *testing.T) {
	ctx := context.Background()
	stores := map[storetypes.StoreKey]int{
		StoreKeyAuth: 0,
	}
	storage := NewMultiMemDB(stores)
	mv := NewMVMemory(16, stores, MultiStoreToCachedStorage(storage, stores), nil)

	mview := mv.View(ctx, 0)
	view := mview.GetKVStore(StoreKeyAuth)
	view.Set(Key("a"), []byte("1"))
	view.Set(Key("b"), []byte("1"))
	view.Set(Key("c"), []byte("1"))
	require.True(t, mv.Record(TxnVersion{0, 0}, mview))

	mview = mv.View(ctx, 1)
	view = mview.GetKVStore(StoreKeyAuth)
	view.Delete(Key("a"))
	view.Set(Key("b"), []byte("2"))
	require.True(t, mv.Record(TxnVersion{1, 0}, mview))

	mview = mv.View(ctx, 2)
	view = mview.GetKVStore(StoreKeyAuth)
	require.Nil(t, view.Get(Key("a")))
	require.False(t, view.Has(Key("a")))
}

func TestMVMemoryViewHasStoragePath(t *testing.T) {
	ctx := context.Background()
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0}

	parent := NewMultiMemDB(stores)
	parent.GetKVStore(StoreKeyAuth).Set([]byte("present"), []byte("v"))

	counted := &countingStorage[[]byte]{GStorage: parent.GetKVStore(StoreKeyAuth)}
	storages := []Storage{NewGCachedStorage[[]byte](counted, storetypes.BytesIsZero)}
	mv := NewMVMemory(4, stores, storages, NewScheduler(4))

	mview := mv.View(ctx, 0)
	view := mview.GetKVStore(StoreKeyAuth)
	require.True(t, view.Has([]byte("present")))
	require.False(t, view.Has([]byte("missing")))

	require.EqualValues(t, 2, counted.gets.Load()+counted.hasOps.Load(),
		"Has() should consult underlying storage once per distinct key")

	rs := (*mview.ReadSet())[0]
	require.Empty(t, rs.Reads, "Has() must not emit value-versioned descriptors")
	require.Equal(t, []HasDescriptor{
		{Key: []byte("present"), Exists: true, FromStorage: true},
		{Key: []byte("missing"), Exists: false, FromStorage: true},
	}, rs.HasReads)
}

func TestMVMemoryViewHasConflictReduction(t *testing.T) {
	key := []byte("k")
	cases := []struct {
		name      string
		mutate    func(storetypes.KVStore)
		wantValid bool
	}{
		{"same_existence_revalidates", func(s storetypes.KVStore) { s.Set(key, []byte("v0-new")) }, true},
		{"existence_flip_invalidates", func(s storetypes.KVStore) { s.Delete(key) }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0}
			storage := NewMultiMemDB(stores)
			storage.GetKVStore(StoreKeyAuth).Set(key, []byte("v0"))

			mv := NewMVMemory(4, stores, MultiStoreToCachedStorage(storage, stores), NewScheduler(4))

			// Txn 2 observes Has(key)==true via storage.
			view2 := mv.View(ctx, 2)
			require.True(t, view2.GetKVStore(StoreKeyAuth).Has(key))
			mv.Record(TxnVersion{2, 0}, view2)

			// Txn 0 mutates the same key; existence may or may not flip.
			view0 := mv.View(ctx, 0)
			tc.mutate(view0.GetKVStore(StoreKeyAuth))
			mv.Record(TxnVersion{0, 0}, view0)

			require.Equal(t, tc.wantValid, mv.ValidateReadSet(ctx, 2))
		})
	}
}

func TestReadSetKeyAliasing(t *testing.T) {
	ctx := context.Background()
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0}
	storage := NewMultiMemDB(stores)
	mv := NewMVMemory(4, stores, MultiStoreToCachedStorage(storage, stores), NewScheduler(4))

	mview := mv.View(ctx, 0)
	view := mview.GetKVStore(StoreKeyAuth)

	getKey := []byte("get-key")
	hasKey := []byte("has-key")
	_ = view.Get(getKey)
	_ = view.Has(hasKey)

	rs := (*mview.ReadSet())[0]
	require.True(t, &rs.Reads[0].Key[0] == &getKey[0],
		"Get should store the caller's key slice without copying")
	require.True(t, &rs.HasReads[0].Key[0] == &hasKey[0],
		"Has should store the caller's key slice without copying")
}

func TestMVMemoryViewHasWaitsOnEstimate(t *testing.T) {
	ctx := context.Background()
	key := []byte("k")
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0}
	storage := NewMultiMemDB(stores)
	scheduler := NewScheduler(4)
	mv := NewMVMemory(4, stores, MultiStoreToCachedStorage(storage, stores), scheduler)

	// Txn 0 wrote key then was re-marked ESTIMATE (mimics block-stm aborting a writer).
	view0 := mv.View(ctx, 0)
	view0.GetKVStore(StoreKeyAuth).Set(key, []byte("v"))
	mv.Record(TxnVersion{0, 0}, view0)
	mv.ConvertWritesToEstimates(0)

	existsCh := make(chan bool, 1)
	go func() {
		scheduler.TryIncarnate(2)
		view2 := mv.View(ctx, 2)
		exists := view2.GetKVStore(StoreKeyAuth).Has(key)
		mv.Record(TxnVersion{2, 0}, view2)
		existsCh <- exists
	}()

	select {
	case <-existsCh:
		t.Fatal("Has() should block on Txn 0's ESTIMATE")
	case <-time.After(50 * time.Millisecond):
	}

	// Txn 0 re-executes with a delete; existence flips to false.
	view0 = mv.View(ctx, 0)
	view0.GetKVStore(StoreKeyAuth).Delete(key)
	mv.Record(TxnVersion{0, 1}, view0)
	scheduler.TryIncarnate(0)
	scheduler.FinishExecution(TxnVersion{0, 1}, false)

	select {
	case exists := <-existsCh:
		require.False(t, exists, "Has() should observe Txn 0's delete after wake-up")
	case <-time.After(time.Second):
		t.Fatal("Has() did not wake up after Txn 0 finished")
	}
	require.True(t, mv.ValidateReadSet(ctx, 2))
}

func TestMVMemoryViewIteration(t *testing.T) {
	ctx := context.Background()
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0}
	storage := NewMultiMemDB(stores)
	mv := NewMVMemory(16, stores, MultiStoreToCachedStorage(storage, stores), nil)
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
		mview := mv.View(ctx, TxnIndex(i))
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
			view := mv.View(ctx, tc.index).GetKVStore(StoreKeyAuth)
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
