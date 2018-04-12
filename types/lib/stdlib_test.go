package stdlib

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dbm "github.com/tendermint/tmlibs/db"

	abci "github.com/tendermint/abci/types"

	store "github.com/cosmos/cosmos-sdk/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

type S struct {
	I int64
	B bool
}

func defaultComponents(key sdk.StoreKey) (sdk.Context, *wire.Codec) {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, nil)
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
	assert.Equal(t, int64(1), lm.Len(ctx))
	lm.Get(ctx, int64(0), &res)
	assert.Equal(t, val, res)

	val = S{2, false}
	lm.Set(ctx, int64(0), val)
	lm.Get(ctx, int64(0), &res)
	assert.Equal(t, val, res)

	lm.Iterate(ctx, &res, func(ctx sdk.Context, index int64) (brk bool) {
		lm.Set(ctx, index, S{res.I + 1, !res.B})
		return
	})
	lm.Get(ctx, int64(0), &res)
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
	assert.Panics(t, func() { qm.Peek(ctx, &res) })

	qm.Push(ctx, S{1, true})
	qm.Push(ctx, S{2, true})
	qm.Push(ctx, S{3, true})
	qm.Iterate(ctx, &res, func(ctx sdk.Context) (brk bool) {
		if res.I == 3 {
			brk = true
		}
		return
	})

	assert.False(t, qm.IsEmpty(ctx))
	qm.Pop(ctx)
	assert.True(t, qm.IsEmpty(ctx))
}
