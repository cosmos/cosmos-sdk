package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
)

func TestContextGetOpShouldNeverPanic(t *testing.T) {
	var ms types.MultiStore
	ctx := types.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	indices := []int64{
		-10, 1, 0, 10, 20,
	}

	for _, index := range indices {
		_, _ = ctx.GetOp(index)
	}
}

func defaultContext(key types.StoreKey) types.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, types.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := types.NewContext(cms, abci.Header{}, false, nil, log.NewNopLogger())
	return ctx
}

func TestCacheContext(t *testing.T) {
	key := types.NewKVStoreKey(t.Name())
	k1 := []byte("hello")
	v1 := []byte("world")
	k2 := []byte("key")
	v2 := []byte("value")

	ctx := defaultContext(key)
	store := ctx.KVStore(key)
	store.Set(k1, v1)
	assert.Equal(t, v1, store.Get(k1))
	assert.Nil(t, store.Get(k2))

	cctx, write := ctx.CacheContext()
	cstore := cctx.KVStore(key)
	assert.Equal(t, v1, cstore.Get(k1))
	assert.Nil(t, cstore.Get(k2))

	cstore.Set(k2, v2)
	assert.Equal(t, v2, cstore.Get(k2))
	assert.Nil(t, store.Get(k2))

	write()

	assert.Equal(t, v2, store.Get(k2))
}
