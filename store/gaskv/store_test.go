package gaskv_test

import (
	"fmt"
	"testing"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/gaskv"
	"github.com/cosmos/cosmos-sdk/store/types"

	"github.com/stretchr/testify/require"
)

func bz(s string) []byte { return []byte(s) }

func keyFmt(i int) []byte { return bz(fmt.Sprintf("key%0.8d", i)) }
func valFmt(i int) []byte { return bz(fmt.Sprintf("value%0.8d", i)) }

func TestGasKVStoreBasic(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	meter := types.NewGasMeter(10000)
	st := gaskv.NewStore(mem, meter, types.KVGasConfig())

	require.Equal(t, types.StoreTypeDB, st.GetStoreType())
	require.Panics(t, func() { st.CacheWrap() })
	require.Panics(t, func() { st.CacheWrapWithTrace(nil, nil) })

	require.Panics(t, func() { st.Set(nil, []byte("value")) }, "setting a nil key should panic")
	require.Panics(t, func() { st.Set([]byte(""), []byte("value")) }, "setting an empty key should panic")

	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	st.Set(keyFmt(1), valFmt(1))
	require.Equal(t, valFmt(1), st.Get(keyFmt(1)))
	st.Delete(keyFmt(1))
	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	require.Equal(t, meter.GasConsumed(), types.Gas(6429))
}

func TestGasKVStoreIterator(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	meter := types.NewGasMeter(10000)
	st := gaskv.NewStore(mem, meter, types.KVGasConfig())
	require.False(t, st.Has(keyFmt(1)))
	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	require.Empty(t, st.Get(keyFmt(2)), "Expected `key2` to be empty")
	st.Set(keyFmt(1), valFmt(1))
	require.True(t, st.Has(keyFmt(1)))
	st.Set(keyFmt(2), valFmt(2))

	iterator := st.Iterator(nil, nil)
	start, end := iterator.Domain()
	require.Nil(t, start)
	require.Nil(t, end)
	require.NoError(t, iterator.Error())

	t.Cleanup(func() {
		if err := iterator.Close(); err != nil {
			t.Fatal(err)
		}
	})
	ka := iterator.Key()
	require.Equal(t, ka, keyFmt(1))
	va := iterator.Value()
	require.Equal(t, va, valFmt(1))
	iterator.Next()
	kb := iterator.Key()
	require.Equal(t, kb, keyFmt(2))
	vb := iterator.Value()
	require.Equal(t, vb, valFmt(2))
	iterator.Next()
	require.False(t, iterator.Valid())
	require.Panics(t, iterator.Next)
	require.NoError(t, iterator.Error())

	reverseIterator := st.ReverseIterator(nil, nil)
	t.Cleanup(func() {
		if err := reverseIterator.Close(); err != nil {
			t.Fatal(err)
		}
	})
	require.Equal(t, reverseIterator.Key(), keyFmt(2))
	reverseIterator.Next()
	require.Equal(t, reverseIterator.Key(), keyFmt(1))
	reverseIterator.Next()
	require.False(t, reverseIterator.Valid())
	require.Panics(t, reverseIterator.Next)

	require.Equal(t, types.Gas(9194), meter.GasConsumed())
}

func TestGasKVStoreOutOfGasSet(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	meter := types.NewGasMeter(0)
	st := gaskv.NewStore(mem, meter, types.KVGasConfig())
	require.Panics(t, func() { st.Set(keyFmt(1), valFmt(1)) }, "Expected out-of-gas")
}

func TestGasKVStoreOutOfGasIterator(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	meter := types.NewGasMeter(20000)
	st := gaskv.NewStore(mem, meter, types.KVGasConfig())
	st.Set(keyFmt(1), valFmt(1))
	iterator := st.Iterator(nil, nil)
	iterator.Next()
	require.Panics(t, func() { iterator.Value() }, "Expected out-of-gas")
}
