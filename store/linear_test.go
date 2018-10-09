package store

import (
	//	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type S struct {
	I uint64
	B bool
}

func defaultComponents(key sdk.StoreKey) (sdk.Context, *codec.Codec) {
	db := dbm.NewMemDB()
	cms := NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	cdc := codec.New()
	return ctx, cdc
}

func TestList(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx, cdc := defaultComponents(key)
	store := ctx.KVStore(key)
	lm := NewList(cdc, store)

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

func TestQueue(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx, cdc := defaultComponents(key)
	store := ctx.KVStore(key)

	qm := NewQueue(cdc, store)

	val := S{1, true}
	var res S

	qm.Push(val)
	qm.Peek(&res)
	require.Equal(t, val, res)

	qm.Pop()
	empty := qm.IsEmpty()

	require.True(t, empty)
	require.NotNil(t, qm.Peek(&res))

	qm.Push(S{1, true})
	qm.Push(S{2, true})
	qm.Push(S{3, true})
	qm.Flush(&res, func() (brk bool) {
		if res.I == 3 {
			brk = true
		}
		return
	})

	require.False(t, qm.IsEmpty())

	qm.Pop()
	require.True(t, qm.IsEmpty())
}

func TestKeys(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx, cdc := defaultComponents(key)
	store := ctx.KVStore(key)
	queue := NewQueue(cdc, store)

	for i := 0; i < 10; i++ {
		queue.Push(i)
	}

	var len uint64
	var top uint64
	var expected int
	var actual int

	// Checking keys.LengthKey
	err := cdc.UnmarshalBinary(store.Get(LengthKey()), &len)
	require.Nil(t, err)
	require.Equal(t, len, queue.List.Len())

	// Checking keys.ElemKey
	for i := 0; i < 10; i++ {
		queue.List.Get(uint64(i), &expected)
		bz := store.Get(ElemKey(uint64(i)))
		err = cdc.UnmarshalBinary(bz, &actual)
		require.Nil(t, err)
		require.Equal(t, expected, actual)
	}

	queue.Pop()

	err = cdc.UnmarshalBinary(store.Get(TopKey()), &top)
	require.Nil(t, err)
	require.Equal(t, top, queue.getTop())

}

/*
func TestListRandom(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx, cdc := defaultComponents(key)
	store := ctx.KVStore(key)
	list := NewList(cdc, store)
	mocklist := []uint32{}

	for i := 0; i < 10000; {
		limit := rand.Int31() % 1000
		for j := int32(0); j < limit; j++ {
			item := rand.Uint32()
			list.Push(item)
			mocklist = append(mocklist, item)
		}

		require.Equal()
	}
}
*/
