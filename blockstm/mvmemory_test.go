package blockstm

import (
	"testing"

	"github.com/test-go/testify/require"

	storetypes "cosmossdk.io/store/types"
)

func TestMVMemoryRecord(t *testing.T) {
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0}
	storage := NewMultiMemDB(stores)
	scheduler := NewScheduler(16)
	mv := NewMVMemory(16, stores, storage, scheduler)

	var views []*MultiMVMemoryView
	for i := TxnIndex(0); i < 3; i++ {
		view := mv.View(i)
		store := view.GetKVStore(StoreKeyAuth)

		_ = store.Get([]byte("a"))
		_ = store.Get([]byte("d"))
		store.Set([]byte("a"), []byte("1"))
		store.Set([]byte("b"), []byte("1"))
		store.Set([]byte("c"), []byte("1"))

		views = append(views, view)
	}

	for i, view := range views {
		wroteNewLocation := mv.Record(TxnVersion{TxnIndex(i), 0}, view)
		require.True(t, wroteNewLocation)
	}

	require.True(t, mv.ValidateReadSet(0))
	require.False(t, mv.ValidateReadSet(1))
	require.False(t, mv.ValidateReadSet(2))

	// abort 2 and 3
	mv.ConvertWritesToEstimates(1)
	mv.ConvertWritesToEstimates(2)

	resultCh := make(chan struct{}, 1)
	go func() {
		// set correct status for the Suspend call
		scheduler.TryIncarnate(3)

		view := mv.View(3)
		store := view.GetKVStore(StoreKeyAuth)
		// will wait for tx 2
		store.Get([]byte("a"))
		wroteNewLocation := mv.Record(TxnVersion{3, 1}, view)
		require.False(t, wroteNewLocation)
		require.True(t, mv.ValidateReadSet(3))
		resultCh <- struct{}{}
	}()

	{
		data := mv.GetMVStore(0).(*MVData)
		value, version, estimate := data.Read(Key("a"), 1)
		require.False(t, estimate)
		require.Equal(t, []byte("1"), value)
		require.Equal(t, TxnVersion{0, 0}, version)

		_, version, estimate = data.Read(Key("a"), 2)
		require.True(t, estimate)
		require.Equal(t, TxnIndex(1), version.Index)

		_, version, estimate = data.Read(Key("a"), 3)
		require.True(t, estimate)
		require.Equal(t, TxnIndex(2), version.Index)
	}

	// rerun tx 1
	{
		view := mv.View(1)
		store := view.GetKVStore(StoreKeyAuth)

		_ = store.Get([]byte("a"))
		_ = store.Get([]byte("d"))
		store.Set([]byte("a"), []byte("2"))
		store.Set([]byte("b"), []byte("2"))
		store.Set([]byte("c"), []byte("2"))

		wroteNewLocation := mv.Record(TxnVersion{1, 1}, view)
		require.False(t, wroteNewLocation)
		require.True(t, mv.ValidateReadSet(1))
	}

	// rerun tx 2
	// don't write `c` this time
	{
		version := TxnVersion{2, 1}
		view := mv.View(version.Index)
		store := view.GetKVStore(StoreKeyAuth)

		_ = store.Get([]byte("a"))
		_ = store.Get([]byte("d"))
		store.Set([]byte("a"), []byte("3"))
		store.Set([]byte("b"), []byte("3"))

		wroteNewLocation := mv.Record(version, view)
		require.False(t, wroteNewLocation)
		require.True(t, mv.ValidateReadSet(2))

		scheduler.TryIncarnate(version.Index)
		scheduler.FinishExecution(version, wroteNewLocation)

		// wait for dependency to finish
		<-resultCh
	}

	// run tx 3
	{
		view := mv.View(3)
		store := view.GetKVStore(StoreKeyAuth)

		_ = store.Get([]byte("a"))

		wroteNewLocation := mv.Record(TxnVersion{3, 1}, view)
		require.False(t, wroteNewLocation)
		require.True(t, mv.ValidateReadSet(3))
	}

	{
		data := mv.GetMVStore(0).(*MVData)
		value, version, estimate := data.Read(Key("a"), 2)
		require.False(t, estimate)
		require.Equal(t, []byte("2"), value)
		require.Equal(t, TxnVersion{1, 1}, version)

		value, version, estimate = data.Read(Key("a"), 3)
		require.False(t, estimate)
		require.Equal(t, []byte("3"), value)
		require.Equal(t, TxnVersion{2, 1}, version)

		value, version, estimate = data.Read(Key("c"), 3)
		require.False(t, estimate)
		require.Equal(t, []byte("2"), value)
		require.Equal(t, TxnVersion{1, 1}, version)
	}
}

func TestMVMemoryDelete(t *testing.T) {
	nonceKey, balanceKey := []byte("nonce"), []byte("balance")

	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0, StoreKeyBank: 1}
	storage := NewMultiMemDB(stores)
	{
		// genesis state
		authStore := storage.GetKVStore(StoreKeyAuth)
		authStore.Set(nonceKey, []byte{0})
		bankStore := storage.GetKVStore(StoreKeyBank)
		bankStore.Set(balanceKey, []byte{100})
	}
	scheduler := NewScheduler(16)
	mv := NewMVMemory(16, stores, storage, scheduler)

	genMockTx := func(txNonce int) func(*MultiMVMemoryView) bool {
		return func(view *MultiMVMemoryView) bool {
			bankStore := view.GetKVStore(StoreKeyBank)
			balance := int(bankStore.Get(balanceKey)[0])
			if balance < 50 {
				// insurfficient balance
				return false
			}

			authStore := view.GetKVStore(StoreKeyAuth)
			nonce := int(authStore.Get(nonceKey)[0])
			// do a set no matter what
			authStore.Set(nonceKey, []byte{byte(nonce)})
			if nonce != txNonce {
				// invalid nonce
				return false
			}

			authStore.Set(nonceKey, []byte{byte(nonce + 1)})
			bankStore.Set(balanceKey, []byte{byte(balance - 50)})
			return true
		}
	}

	tx0, tx1, tx2 := genMockTx(0), genMockTx(1), genMockTx(2)

	view0 := mv.View(0)
	require.True(t, tx0(view0))
	view1 := mv.View(1)
	require.False(t, tx1(view1))
	view2 := mv.View(2)
	require.False(t, tx2(view2))

	require.True(t, mv.Record(TxnVersion{1, 0}, view1))
	require.True(t, mv.Record(TxnVersion{2, 0}, view2))
	require.True(t, mv.Record(TxnVersion{0, 0}, view0))

	require.True(t, mv.ValidateReadSet(0))
	require.False(t, mv.ValidateReadSet(1))
	mv.ConvertWritesToEstimates(1)
	require.False(t, mv.ValidateReadSet(2))
	mv.ConvertWritesToEstimates(2)

	// re-execute tx 1 and 2
	view1 = mv.View(1)
	require.True(t, tx1(view1))
	mv.Record(TxnVersion{1, 1}, view1)
	require.True(t, mv.ValidateReadSet(1))

	view2 = mv.View(2)
	// tx 2 fail due to insufficient balance, but stm validation is successful.
	require.False(t, tx2(view2))
	mv.Record(TxnVersion{2, 1}, view2)
	require.True(t, mv.ValidateReadSet(2))

	mv.WriteSnapshot(storage)
	{
		authStore := storage.GetKVStore(StoreKeyAuth)
		require.Equal(t, []byte{2}, authStore.Get(nonceKey))
		bankStore := storage.GetKVStore(StoreKeyBank)
		require.Equal(t, []byte{0}, bankStore.Get(balanceKey))
	}
}

func TestMVMemoryIteration(t *testing.T) {
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0}
	storage := NewMultiMemDB(stores)
	mv := NewMVMemory(16, stores, storage, nil)

	view := mv.View(0)
	store := view.GetKVStore(StoreKeyAuth)

	{
		iter := store.Iterator(nil, nil)
		kvs := CollectIterator(iter)
		iter.Close()
		require.Empty(t, kvs)
	}

	store.Set([]byte("a"), []byte("1"))
	store.Set([]byte("b"), []byte("1"))
	store.Set([]byte("c"), []byte("1"))
	require.True(t, mv.Record(TxnVersion{0, 0}, view))

	view = mv.View(1)
	store = view.GetKVStore(StoreKeyAuth)

	{
		iter := store.Iterator(nil, nil)
		kvs := CollectIterator(iter)
		iter.Close()
		require.Equal(t, []KVPair{
			{[]byte("a"), []byte("1")},
			{[]byte("b"), []byte("1")},
			{[]byte("c"), []byte("1")},
		}, kvs)
	}

	store.Set([]byte("b"), []byte("2"))
	store.Set([]byte("c"), []byte("2"))
	store.Set([]byte("d"), []byte("2"))
	require.True(t, mv.Record(TxnVersion{1, 0}, view))

	view = mv.View(2)
	store = view.GetKVStore(StoreKeyAuth)

	{
		iter := store.Iterator(nil, nil)
		kvs := CollectIterator(iter)
		iter.Close()
		require.Equal(t, []KVPair{
			{[]byte("a"), []byte("1")},
			{[]byte("b"), []byte("2")},
			{[]byte("c"), []byte("2")},
			{[]byte("d"), []byte("2")},
		}, kvs)
	}
	store.Set([]byte("c"), []byte("3"))
	store.Set([]byte("d"), []byte("3"))
	store.Set([]byte("e"), []byte("3"))
	require.True(t, mv.Record(TxnVersion{2, 0}, view))

	require.True(t, mv.ValidateReadSet(0))
	require.True(t, mv.ValidateReadSet(1))
	require.True(t, mv.ValidateReadSet(2))
}
