package lib

import (
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
	ctx := sdk.NewContext(cms, abci.Header{}, false, nil, log.NewNopLogger())
	cdc := wire.NewCodec()
	return ctx, cdc
}

func TestListMapper(t *testing.T) {
	key := sdk.NewKVStoreKey("list")
	ctx, cdc := defaultComponents(key)
	lm := NewListMapper(cdc, key, "data")

	val := S{1, true}
	var res S

	lm.Push(ctx, val)
	assert.Equal(t, uint64(1), lm.Len(ctx))
	lm.Get(ctx, uint64(0), &res)
	assert.Equal(t, val, res)

	val = S{2, false}
	lm.Set(ctx, uint64(0), val)
	lm.Get(ctx, uint64(0), &res)
	assert.Equal(t, val, res)

	val = S{100, false}
	lm.Push(ctx, val)
	assert.Equal(t, uint64(2), lm.Len(ctx))
	lm.Get(ctx, uint64(1), &res)
	assert.Equal(t, val, res)

	lm.Delete(ctx, uint64(1))
	assert.Equal(t, uint64(2), lm.Len(ctx))

	lm.IterateRead(ctx, &res, func(ctx sdk.Context, index uint64) (brk bool) {
		var temp S
		lm.Get(ctx, index, &temp)
		assert.Equal(t, temp, res)

		assert.True(t, index != 1)
		return
	})

	lm.IterateWrite(ctx, &res, func(ctx sdk.Context, index uint64) (brk bool) {
		lm.Set(ctx, index, S{res.I + 1, !res.B})
		return
	})

	lm.Get(ctx, uint64(0), &res)
	assert.Equal(t, S{3, true}, res)
}

func TestQueueMapper(t *testing.T) {
	key := sdk.NewKVStoreKey("queue")
	ctx, cdc := defaultComponents(key)
	qm := NewQueueMapper(cdc, key, "data")

	val := S{1, true}
	var res S

	qm.Push(ctx, val)
	qm.Peek(ctx, &res)
	assert.Equal(t, val, res)

	qm.Pop(ctx)
	empty := qm.IsEmpty(ctx)

	assert.True(t, empty)
	assert.NotNil(t, qm.Peek(ctx, &res))

	qm.Push(ctx, S{1, true})
	qm.Push(ctx, S{2, true})
	qm.Push(ctx, S{3, true})
	qm.Flush(ctx, &res, func(ctx sdk.Context) (brk bool) {
		if res.I == 3 {
			brk = true
		}
		return
	})

	assert.False(t, qm.IsEmpty(ctx))

	qm.Pop(ctx)
	assert.True(t, qm.IsEmpty(ctx))
}
