package store

import (
	"testing"

	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tmlibs/db"
)

func TestCacheKVStore(t *testing.T) {
	mem := dbm.NewMemDB()
	st := NewCacheKVStore(mem)

	require.Empty(t, st.Get(bz("key1")), "Expected `key1` to be empty")

	mem.Set(bz("key1"), bz("value1"))
	st.Set(bz("key1"), bz("value1"))
	require.Equal(t, bz("value1"), st.Get(bz("key1")))

	st.Set(bz("key1"), bz("value2"))
	require.Equal(t, bz("value2"), st.Get(bz("key1")))
	require.Equal(t, bz("value1"), mem.Get(bz("key1")))

	st.Write()
	require.Equal(t, bz("value2"), mem.Get(bz("key1")))

	st.Write()
	st.Write()
	require.Equal(t, bz("value2"), mem.Get(bz("key1")))

	st = NewCacheKVStore(mem)
	st.Delete(bz("key1"))
	require.Empty(t, st.Get(bz("key1")))
	require.Equal(t, mem.Get(bz("key1")), bz("value2"))

	st.Write()
	require.Empty(t, st.Get(bz("key1")), "Expected `key1` to be empty")
	require.Empty(t, mem.Get(bz("key1")), "Expected `key1` to be empty")
}

func TestCacheKVStoreNested(t *testing.T) {
	mem := dbm.NewMemDB()
	st := NewCacheKVStore(mem)
	st.Set(bz("key1"), bz("value1"))

	require.Empty(t, mem.Get(bz("key1")))
	require.Equal(t, bz("value1"), st.Get(bz("key1")))
	st2 := NewCacheKVStore(st)
	require.Equal(t, bz("value1"), st2.Get(bz("key1")))

	st2.Set(bz("key1"), bz("VALUE2"))
	require.Equal(t, []byte(nil), mem.Get(bz("key1")))
	require.Equal(t, bz("value1"), st.Get(bz("key1")))
	require.Equal(t, bz("VALUE2"), st2.Get(bz("key1")))

	st2.Write()
	require.Equal(t, []byte(nil), mem.Get(bz("key1")))
	require.Equal(t, bz("VALUE2"), st.Get(bz("key1")))

	st.Write()
	require.Equal(t, bz("VALUE2"), mem.Get(bz("key1")))

}

func bz(s string) []byte { return []byte(s) }
