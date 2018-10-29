package list

import (
	"math/rand"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
)

type S struct {
	I uint64
	B bool
}

func defaultComponents(key sdk.KVStoreKey) (sdk.Context, *codec.Codec) {
	db := dbm.NewMemDB()
	cms := rootmulti.NewStore(db)
	cms.MountKVStoreWithDB(key, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	cdc := codec.New()
	return ctx, cdc
}

func TestList(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx, cdc := defaultComponents(key)
	store := ctx.KVStore(key)
	lm := New(cdc, store)

	val := S{1, true}
	var res S

	lm.Push(val)
	require.Equal(t, uint64(1), lm.Len())
	lm.Get(uint64(0), &res)
	require.Equal(t, val, res)

	val = S{2, false}
	lm.Set(uint64(0), val)
	lm.Get(uint64(0), &res)
	require.Equal(t, val, res)

	val = S{100, false}
	lm.Push(val)
	require.Equal(t, uint64(2), lm.Len())
	lm.Get(uint64(1), &res)
	require.Equal(t, val, res)

	lm.Delete(uint64(1))
	require.Equal(t, uint64(2), lm.Len())

	lm.Iterate(&res, func(index uint64) (brk bool) {
		var temp S
		lm.Get(index, &temp)
		require.Equal(t, temp, res)

		require.True(t, index != 1)
		return
	})

	lm.Iterate(&res, func(index uint64) (brk bool) {
		lm.Set(index, S{res.I + 1, !res.B})
		return
	})

	lm.Get(uint64(0), &res)
	require.Equal(t, S{3, true}, res)
}

func TestRandom(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx, cdc := defaultComponents(key)
	store := ctx.KVStore(key)
	list := New(cdc, store)
	mocklist := []uint32{}

	for i := 0; i < 100; i++ {
		item := rand.Uint32()
		list.Push(item)
		mocklist = append(mocklist, item)
	}

	for k, v := range mocklist {
		var i uint32
		require.NotPanics(t, func() { list.Get(uint64(k), &i) })
		require.Equal(t, v, i)
	}
}
