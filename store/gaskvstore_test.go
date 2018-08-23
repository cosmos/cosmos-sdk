package store

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func newGasKVStore() KVStore {
	meter := sdk.NewGasMeter(1000)
	mem := dbStoreAdapter{dbm.NewMemDB()}
	return NewGasKVStore(meter, mem)
}

func TestGasKVStoreBasic(t *testing.T) {
	mem := dbStoreAdapter{dbm.NewMemDB()}
	meter := sdk.NewGasMeter(1000)
	st := NewGasKVStore(meter, mem)
	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	st.Set(keyFmt(1), valFmt(1))
	require.Equal(t, valFmt(1), st.Get(keyFmt(1)))
	st.Delete(keyFmt(1))
	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	require.Equal(t, meter.GasConsumed(), sdk.Gas(183))
}

func TestGasKVStoreIterator(t *testing.T) {
	mem := dbStoreAdapter{dbm.NewMemDB()}
	meter := sdk.NewGasMeter(1000)
	st := NewGasKVStore(meter, mem)
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
	require.Equal(t, meter.GasConsumed(), sdk.Gas(356))
}

func TestGasKVStoreOutOfGasSet(t *testing.T) {
	mem := dbStoreAdapter{dbm.NewMemDB()}
	meter := sdk.NewGasMeter(0)
	st := NewGasKVStore(meter, mem)
	require.Panics(t, func() { st.Set(keyFmt(1), valFmt(1)) }, "Expected out-of-gas")
}

func TestGasKVStoreOutOfGasIterator(t *testing.T) {
	mem := dbStoreAdapter{dbm.NewMemDB()}
	meter := sdk.NewGasMeter(200)
	st := NewGasKVStore(meter, mem)
	st.Set(keyFmt(1), valFmt(1))
	iterator := st.Iterator(nil, nil)
	iterator.Next()
	require.Panics(t, func() { iterator.Value() }, "Expected out-of-gas")
}
