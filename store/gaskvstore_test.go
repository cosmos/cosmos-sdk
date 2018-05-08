package store

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tmlibs/db"
)

func newGasKVStore() KVStore {
	meter := sdk.NewGasMeter(1000)
	mem := dbStoreAdapter{dbm.NewMemDB()}
	return NewGasKVStore(meter, mem)
}

func TestGasKVStore(t *testing.T) {
	mem := dbStoreAdapter{dbm.NewMemDB()}
	meter := sdk.NewGasMeter(1000)
	st := NewGasKVStore(meter, mem)

	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")

	mem.Set(keyFmt(1), valFmt(1))
	st.Set(keyFmt(1), valFmt(1))
	require.Equal(t, valFmt(1), st.Get(keyFmt(1)))
}
