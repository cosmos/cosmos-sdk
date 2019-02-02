package gaskv_test

import (
	"fmt"
	"testing"

	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/gaskv"
	stypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/stretchr/testify/require"
)

func newGasKVStore() stypes.KVStore {
	meter := stypes.NewGasMeter(10000)
	mem := dbadapter.Store{dbm.NewMemDB()}
	return gaskv.NewStore(mem, meter, stypes.KVGasConfig())
}

func bz(s string) []byte { return []byte(s) }

func keyFmt(i int) []byte { return bz(fmt.Sprintf("key%0.8d", i)) }
func valFmt(i int) []byte { return bz(fmt.Sprintf("value%0.8d", i)) }

func TestGasKVStoreBasic(t *testing.T) {
	mem := dbadapter.Store{dbm.NewMemDB()}
	meter := stypes.NewGasMeter(10000)
	st := gaskv.NewStore(mem, meter, stypes.KVGasConfig())
	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	st.Set(keyFmt(1), valFmt(1))
	require.Equal(t, valFmt(1), st.Get(keyFmt(1)))
	st.Delete(keyFmt(1))
	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	require.Equal(t, meter.GasConsumed(), stypes.Gas(6429))
}

func TestGasKVStoreIterator(t *testing.T) {
	mem := dbadapter.Store{dbm.NewMemDB()}
	meter := stypes.NewGasMeter(10000)
	st := gaskv.NewStore(mem, meter, stypes.KVGasConfig())
	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	require.Empty(t, st.Get(keyFmt(2)), "Expected `key2` to be empty")
	st.Set(keyFmt(1), valFmt(1))
	st.Set(keyFmt(2), valFmt(2))
	iterator := st.Iterator(nil, nil)
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
	require.Equal(t, meter.GasConsumed(), stypes.Gas(6987))
}

func TestGasKVStoreOutOfGasSet(t *testing.T) {
	mem := dbadapter.Store{dbm.NewMemDB()}
	meter := stypes.NewGasMeter(0)
	st := gaskv.NewStore(mem, meter, stypes.KVGasConfig())
	require.Panics(t, func() { st.Set(keyFmt(1), valFmt(1)) }, "Expected out-of-gas")
}

func TestGasKVStoreOutOfGasIterator(t *testing.T) {
	mem := dbadapter.Store{dbm.NewMemDB()}
	meter := stypes.NewGasMeter(20000)
	st := gaskv.NewStore(mem, meter, stypes.KVGasConfig())
	st.Set(keyFmt(1), valFmt(1))
	iterator := st.Iterator(nil, nil)
	iterator.Next()
	require.Panics(t, func() { iterator.Value() }, "Expected out-of-gas")
}
