package store

import (
	"testing"

	dbm "github.com/tendermint/tendermint/libs/db"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func newGasKVStore() KVStore {
	meter := sdk.NewGasMeter(1000)
	mem := dbStoreAdapter{dbm.NewMemDB()}
	return NewGasKVStore(meter, sdk.DefaultGasConfig(), mem)
}

func TestGasKVStoreBasic(t *testing.T) {
	mem := dbStoreAdapter{dbm.NewMemDB()}
	meter := sdk.NewGasMeter(1000)
	st := NewGasKVStore(meter, sdk.DefaultGasConfig(), mem)
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
	st := NewGasKVStore(meter, sdk.DefaultGasConfig(), mem)
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
	st := NewGasKVStore(meter, sdk.DefaultGasConfig(), mem)
	require.Panics(t, func() { st.Set(keyFmt(1), valFmt(1)) }, "Expected out-of-gas")
}

func TestGasKVStoreOutOfGasIterator(t *testing.T) {
	mem := dbStoreAdapter{dbm.NewMemDB()}
	meter := sdk.NewGasMeter(200)
	st := NewGasKVStore(meter, sdk.DefaultGasConfig(), mem)
	st.Set(keyFmt(1), valFmt(1))
	iterator := st.Iterator(nil, nil)
	iterator.Next()
	require.Panics(t, func() { iterator.Value() }, "Expected out-of-gas")
}

func testGasKVStoreWrap(t *testing.T, store KVStore) {
	meter := sdk.NewGasMeter(10000)

	store = store.Gas(meter, sdk.GasConfig{HasCost: 10})
	require.Equal(t, int64(0), meter.GasConsumed())

	store.Has([]byte("key"))
	require.Equal(t, int64(10), meter.GasConsumed())

	store = store.Gas(meter, sdk.GasConfig{HasCost: 20})

	store.Has([]byte("key"))
	require.Equal(t, int64(40), meter.GasConsumed())
}

func TestGasKVStoreWrap(t *testing.T) {
	db := dbm.NewMemDB()
	tree, _ := newTree(t, db)
	iavl := newIAVLStore(tree, numRecent, storeEvery)
	testGasKVStoreWrap(t, iavl)

	st := NewCacheKVStore(iavl)
	testGasKVStoreWrap(t, st)

	pref := st.Prefix([]byte("prefix"))
	testGasKVStoreWrap(t, pref)

	dsa := dbStoreAdapter{dbm.NewMemDB()}
	testGasKVStoreWrap(t, dsa)

	ts := newTransientStore()
	testGasKVStoreWrap(t, ts)

}
