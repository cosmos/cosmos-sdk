package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dbm "github.com/tendermint/tmlibs/db"

	abci "github.com/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/store"
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
	lm := sdk.NewListMapper(cdc, key)

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
}

func TestQueueMapper(t *testing.T) {
	key := sdk.NewKVStoreKey("queue")
	ctx, cdc := defaultComponents(key)
	qm := sdk.NewQueueMapper(cdc, key)

	val := S{1, true}
	var res S

	qm.Push(ctx, val)
	qm.Peek(ctx, &res)
	assert.Equal(t, val, res)

	qm.Pop(ctx)
	empty := qm.IsEmpty(ctx)

	assert.Equal(t, true, empty)

	assert.Panics(t, func() { qm.Peek(ctx, &res) })
}
