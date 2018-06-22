package lib

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	abci "github.com/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

type S struct {
	I uint64
	B bool
}

func defaultComponents(key sdk.StoreKey) (sdk.Context, *wire.Codec) {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	cdc := wire.NewCodec()
	return ctx, cdc
}

func TestList(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx, cdc := defaultComponents(key)
	store := ctx.KVStore(key)
	lm := NewList(cdc, store, nil)

	val := S{1, true}
	var res S

	lm.Push(val)
	assert.Equal(t, uint64(1), lm.Len())
	lm.Get(uint64(0), &res)
	assert.Equal(t, val, res)

	val = S{2, false}
	lm.Set(uint64(0), val)
	lm.Get(uint64(0), &res)
	assert.Equal(t, val, res)

	val = S{100, false}
	lm.Push(val)
	assert.Equal(t, uint64(2), lm.Len())
	lm.Get(uint64(1), &res)
	assert.Equal(t, val, res)

	lm.Delete(uint64(1))
	assert.Equal(t, uint64(2), lm.Len())

	lm.Iterate(&res, func(index uint64) (brk bool) {
		var temp S
		lm.Get(index, &temp)
		assert.Equal(t, temp, res)

		assert.True(t, index != 1)
		return
	})

	lm.Iterate(&res, func(index uint64) (brk bool) {
		lm.Set(index, S{res.I + 1, !res.B})
		return
	})

	lm.Get(uint64(0), &res)
	assert.Equal(t, S{3, true}, res)
}

func TestQueue(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx, cdc := defaultComponents(key)
	store := ctx.KVStore(key)

	qm := NewQueue(cdc, store, nil)

	val := S{1, true}
	var res S

	qm.Push(val)
	qm.Peek(&res)
	assert.Equal(t, val, res)

	qm.Pop()
	empty := qm.IsEmpty()

	assert.True(t, empty)
	assert.NotNil(t, qm.Peek(&res))

	qm.Push(S{1, true})
	qm.Push(S{2, true})
	qm.Push(S{3, true})
	qm.Flush(&res, func() (brk bool) {
		if res.I == 3 {
			brk = true
		}
		return
	})

	assert.False(t, qm.IsEmpty())

	qm.Pop()
	assert.True(t, qm.IsEmpty())
}

func TestOptions(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx, cdc := defaultComponents(key)
	store := ctx.KVStore(key)

	keys := &LinearKeys{
		LengthKey: []byte{0xDE, 0xAD},
		ElemKey:   []byte{0xBE, 0xEF},
		TopKey:    []byte{0x12, 0x34},
	}
	linear := NewLinear(cdc, store, keys)

	for i := 0; i < 10; i++ {
		linear.Push(i)
	}

	var len uint64
	var top uint64
	var expected int
	var actual int

	// Checking keys.LengthKey
	err := cdc.UnmarshalBinary(store.Get(keys.LengthKey), &len)
	assert.Nil(t, err)
	assert.Equal(t, len, linear.Len())

	// Checking keys.ElemKey
	for i := 0; i < 10; i++ {
		linear.Get(uint64(i), &expected)
		bz := store.Get(append(keys.ElemKey, []byte(fmt.Sprintf("%020d", i))...))
		err = cdc.UnmarshalBinary(bz, &actual)
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	}

	linear.Pop()

	err = cdc.UnmarshalBinary(store.Get(keys.TopKey), &top)
	assert.Nil(t, err)
	assert.Equal(t, top, linear.getTop())

}
