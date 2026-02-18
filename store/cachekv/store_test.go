package cachekv_test

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/dbadapter"
)

func newStoreWithParent() (*cachekv.Store, dbadapter.Store) {
	parent := dbadapter.Store{DB: dbm.NewMemDB()}
	return cachekv.NewStore(parent), parent
}

func TestGet_FromParentWhenClean(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("parent_value"))

	got := st.Get([]byte("key"))
	require.Equal(t, []byte("parent_value"), got)
}

func TestGet_FromWriteMap(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("parent_value"))

	st.Set([]byte("key"), []byte("cached_value"))

	got := st.Get([]byte("key"))
	require.Equal(t, []byte("cached_value"), got)
}

func TestGet_DeletedKeyReturnsNil(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("parent_value"))

	st.Delete([]byte("key"))

	got := st.Get([]byte("key"))
	require.Nil(t, got)
}

func TestHas_CachedValue(t *testing.T) {
	st, _ := newStoreWithParent()
	st.Set([]byte("key"), []byte("value"))

	require.True(t, st.Has([]byte("key")))
}

func TestHas_DeletedKey(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("value"))

	st.Delete([]byte("key"))

	require.False(t, st.Has([]byte("key")))
}

func TestSet_MarksStoreDirty(t *testing.T) {
	st, _ := newStoreWithParent()

	_, count := st.Updates()
	require.Equal(t, 0, count)

	st.Set([]byte("key"), []byte("value"))

	_, count = st.Updates()
	require.Equal(t, 1, count)
}

func TestDelete_MarksStoreDirty(t *testing.T) {
	st, _ := newStoreWithParent()

	_, count := st.Updates()
	require.Equal(t, 0, count)

	st.Delete([]byte("key"))

	_, count = st.Updates()
	require.Equal(t, 1, count)
}

// --- Input Validation ---

func TestSet_PanicsOnInvalidInput(t *testing.T) {
	st, _ := newStoreWithParent()

	require.Panics(t, func() { st.Set(nil, []byte("value")) }, "nil key")
	require.Panics(t, func() { st.Set([]byte(""), []byte("value")) }, "empty key")
	require.Panics(t, func() { st.Set([]byte("key"), nil) }, "nil value")
}

func TestGet_PanicsOnInvalidKey(t *testing.T) {
	st, _ := newStoreWithParent()

	require.Panics(t, func() { st.Get(nil) })
	require.Panics(t, func() { st.Get([]byte("")) })
}

func TestHas_PanicsOnInvalidKey(t *testing.T) {
	st, _ := newStoreWithParent()

	require.Panics(t, func() { st.Has(nil) })
	require.Panics(t, func() { st.Has([]byte("")) })
}

func TestDelete_PanicsOnInvalidKey(t *testing.T) {
	st, _ := newStoreWithParent()

	require.Panics(t, func() { st.Delete(nil) })
	require.Panics(t, func() { st.Delete([]byte("")) })
}

// --- Write Method ---

func TestWrite_FlushesToParent(t *testing.T) {
	st, parent := newStoreWithParent()

	st.Set([]byte("key"), []byte("value"))
	require.False(t, parent.Has([]byte("key")))

	st.Write()

	require.Equal(t, []byte("value"), parent.Get([]byte("key")))
}

func TestWrite_DeletesPropagated(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("value"))

	st.Delete([]byte("key"))
	require.True(t, parent.Has([]byte("key")))

	st.Write()

	require.False(t, parent.Has([]byte("key")))
}

func TestWrite_ClearsWriteMap(t *testing.T) {
	st, _ := newStoreWithParent()

	st.Set([]byte("key"), []byte("value"))
	_, count := st.Updates()
	require.Equal(t, 1, count)

	st.Write()

	// after write, store is clean again (dirty=false path in Updates)
	_, count = st.Updates()
	require.Equal(t, 0, count)
}

func TestWrite_Overwriting(t *testing.T) {
	st, parent := newStoreWithParent()

	st.Set([]byte("key"), []byte("first"))
	st.Set([]byte("key"), []byte("second"))
	st.Set([]byte("key"), []byte("third"))

	st.Write()

	require.Equal(t, []byte("third"), parent.Get([]byte("key")))
}

func TestWrite_SetThenDelete(t *testing.T) {
	st, parent := newStoreWithParent()

	st.Set([]byte("key"), []byte("value"))
	st.Delete([]byte("key"))

	st.Write()

	require.False(t, parent.Has([]byte("key")))
}

func TestWrite_OverwritesParent(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("original"))

	st.Set([]byte("key"), []byte("new_value"))

	st.Write()

	require.Equal(t, []byte("new_value"), parent.Get([]byte("key")))
}

// --- Updates() Method ---

func TestUpdates_EmptyWhenClean(t *testing.T) {
	st, _ := newStoreWithParent()

	updates, count := st.Updates()
	require.Equal(t, 0, count)

	collected := 0
	updates(func(k []byte, v []byte) bool {
		collected++
		return true
	})
	require.Equal(t, 0, collected)
}

func TestUpdates_ReturnsAllChanges(t *testing.T) {
	st, _ := newStoreWithParent()

	st.Set([]byte("a"), []byte("1"))
	st.Set([]byte("b"), []byte("2"))
	st.Delete([]byte("c"))

	updates, count := st.Updates()
	require.Equal(t, 3, count)

	result := make(map[string][]byte)
	updates(func(k []byte, v []byte) bool {
		result[string(k)] = v
		return true
	})

	require.Equal(t, []byte("1"), result["a"])
	require.Equal(t, []byte("2"), result["b"])
	require.Nil(t, result["c"]) // deletion marked with nil
}

func TestUpdates_CorrectCount(t *testing.T) {
	st, _ := newStoreWithParent()

	st.Set([]byte("a"), []byte("1"))
	st.Set([]byte("b"), []byte("2"))
	st.Set([]byte("a"), []byte("updated")) // same key, should not increase count

	_, count := st.Updates()
	require.Equal(t, 2, count)
}

// --- CacheWrap Hierarchy ---

func TestCacheWrap_ReturnsNewStore(t *testing.T) {
	st, _ := newStoreWithParent()

	wrapped := st.CacheWrap()

	require.IsType(t, &cachekv.Store{}, wrapped)
	require.NotSame(t, st, wrapped)
}

func TestCacheWrap_IsolatesWrites(t *testing.T) {
	st, _ := newStoreWithParent()
	st.Set([]byte("key"), []byte("original"))

	child := st.CacheWrap().(*cachekv.Store)
	child.Set([]byte("key"), []byte("modified"))
	child.Set([]byte("new"), []byte("value"))

	// parent should not see child changes
	require.Equal(t, []byte("original"), st.Get([]byte("key")))
	require.False(t, st.Has([]byte("new")))

	// child sees its own changes
	require.Equal(t, []byte("modified"), child.Get([]byte("key")))
	require.True(t, child.Has([]byte("new")))
}

func TestCacheWrap_NestedWrite(t *testing.T) {
	st, parent := newStoreWithParent()

	child := st.CacheWrap().(*cachekv.Store)
	child.Set([]byte("key"), []byte("value"))

	// write child to parent cache
	child.Write()
	require.Equal(t, []byte("value"), st.Get([]byte("key")))
	require.False(t, parent.Has([]byte("key")))

	// write parent cache to underlying
	st.Write()
	require.Equal(t, []byte("value"), parent.Get([]byte("key")))
}
