package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

func newGasKVStore() KVStore {
	meter := NewGasMeter(1000)
	mem := dbStoreAdapter{dbm.NewMemDB()}
	return NewGasKVStore(meter, mem)
}

func TestGasKVStoreBasic(t *testing.T) {
	mem := dbStoreAdapter{dbm.NewMemDB()}
	meter := NewGasMeter(1000)
	st := NewGasKVStore(meter, mem)

	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")

	mem.Set(keyFmt(1), valFmt(1))
	st.Set(keyFmt(1), valFmt(1))
	require.Equal(t, valFmt(1), st.Get(keyFmt(1)))
}

func TestGasKVStoreOutOfGas(t *testing.T) {
	mem := dbStoreAdapter{dbm.NewMemDB()}
	meter := NewGasMeter(0)
	st := NewGasKVStore(meter, mem)
	require.Panics(t, func() { st.Set(keyFmt(1), valFmt(1)) }, "Expected out-of-gas")
}

func keyFmt(i int) []byte { return bz(cmn.Fmt("key%0.8d", i)) }
func valFmt(i int) []byte { return bz(cmn.Fmt("value%0.8d", i)) }
func bz(s string) []byte  { return []byte(s) }

type dbStoreAdapter struct {
	dbm.DB
}

// Implements Store.
func (dbStoreAdapter) GetStoreType() StoreType {
	return StoreTypeDB
}

// Implements KVStore.
func (dsa dbStoreAdapter) CacheWrap() CacheWrap {
	panic("unsupported")
}

func (dsa dbStoreAdapter) SubspaceIterator(prefix []byte) Iterator {
	return dsa.Iterator(prefix, PrefixEndBytes(prefix))
}

func (dsa dbStoreAdapter) ReverseSubspaceIterator(prefix []byte) Iterator {
	return dsa.ReverseIterator(prefix, PrefixEndBytes(prefix))
}

// dbm.DB implements KVStore so we can CacheKVStore it.
var _ KVStore = dbStoreAdapter{dbm.DB(nil)}
