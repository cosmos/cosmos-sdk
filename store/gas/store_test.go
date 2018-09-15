package gas

import (
	"testing"

	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func newGasTank(limit sdk.Gas) *sdk.GasTank {
	return sdk.NewGasTank(sdk.NewGasMeter(limit), sdk.KVGasConfig())
}

func newGasKVStore() sdk.KVStore {
	tank := newGasTank(1000)
	mem := dbadapter.NewStore(dbm.NewMemDB())
	return NewStore(tank, mem)
}

func TestGasKVStoreBasic(t *testing.T) {
	tank := newGasTank(1000)
	mem := dbadapter.NewStore(dbm.NewMemDB())
	st := NewStore(tank, mem)
	require.Empty(t, st.Get(utils.KeyFmt(1)), "Expected `key1` to be empty")
	st.Set(utils.KeyFmt(1), utils.ValFmt(1))
	require.Equal(t, utils.ValFmt(1), st.Get(utils.KeyFmt(1)))
	st.Delete(utils.KeyFmt(1))
	require.Empty(t, st.Get(utils.KeyFmt(1)), "Expected `key1` to be empty")
	require.Equal(t, tank.GasConsumed(), sdk.Gas(193))
}

func TestGasKVStoreIterator(t *testing.T) {
	tank := newGasTank(1000)
	mem := dbadapter.NewStore(dbm.NewMemDB())
	st := NewStore(tank, mem)
	require.Empty(t, st.Get(utils.KeyFmt(1)), "Expected `key1` to be empty")
	require.Empty(t, st.Get(utils.KeyFmt(2)), "Expected `key2` to be empty")
	st.Set(utils.KeyFmt(1), utils.ValFmt(1))
	st.Set(utils.KeyFmt(2), utils.ValFmt(2))
	iterator := st.Iterator(nil, nil)
	ka := iterator.Key()
	require.Equal(t, ka, utils.KeyFmt(1))
	va := iterator.Value()
	require.Equal(t, va, utils.ValFmt(1))
	iterator.Next()
	kb := iterator.Key()
	require.Equal(t, kb, utils.KeyFmt(2))
	vb := iterator.Value()
	require.Equal(t, vb, utils.ValFmt(2))
	iterator.Next()
	require.False(t, iterator.Valid())
	require.Panics(t, iterator.Next)
	require.Equal(t, tank.GasConsumed(), sdk.Gas(384))
}

func TestGasKVStoreOutOfGasSet(t *testing.T) {
	tank := newGasTank(0)
	mem := dbadapter.NewStore(dbm.NewMemDB())
	st := NewStore(tank, mem)
	require.Panics(t, func() { st.Set(utils.KeyFmt(1), utils.ValFmt(1)) }, "Expected out-of-gas")
}

func TestGasKVStoreOutOfGasIterator(t *testing.T) {
	tank := newGasTank(200)
	mem := dbadapter.NewStore(dbm.NewMemDB())
	st := NewStore(tank, mem)
	st.Set(utils.KeyFmt(1), utils.ValFmt(1))
	iterator := st.Iterator(nil, nil)
	iterator.Next()
	require.Panics(t, func() { iterator.Value() }, "Expected out-of-gas")
}

func testGasKVStoreWrap(t *testing.T, store sdk.KVStore) {
	meter := sdk.NewGasMeter(10000)
	tank := sdk.NewGasTank(meter, sdk.GasConfig{HasCostFlat: 10})

	store = NewStore(tank, store)
	require.Equal(t, int64(0), meter.GasConsumed())

	store.Has([]byte("key"))
	require.Equal(t, int64(10), meter.GasConsumed())

	tank = sdk.NewGasTank(meter, sdk.GasConfig{HasCostFlat: 20})
	store = NewStore(tank, store)

	store.Has([]byte("key"))
	require.Equal(t, int64(40), meter.GasConsumed())
}

// XXX: make it stop using iavl or move it to iavl
/*
func TestGasKVStoreWrap(t *testing.T) {
	db := dbm.NewMemDB()
	tree, _ := newTree(t, db)
	iavl := newIAVLStore(tree, numRecent, storeEvery)
	testGasKVStoreWrap(t, iavl)

	st := cache.NewStore(iavl)
	testGasKVStoreWrap(t, st)

	pref := prefix.NewStore(st, []byte("prefix"))
	testGasKVStoreWrap(t, pref)

	dsa := dbadapter.NewStore(dbm.NewMemDB())
	testGasKVStoreWrap(t, dsa)

	ts := transient.NewStore()
	testGasKVStoreWrap(t, ts)
}
*/
