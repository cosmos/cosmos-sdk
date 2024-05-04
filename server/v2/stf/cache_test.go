package stf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"

	appmanager "cosmossdk.io/core/app"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/mock"
)

var (
	tempCache    = NewModuleContainer()
	actorNameStr = "test"
)

func TestMockCacheCtx(t *testing.T) {
	storeService := NewStoreService(actorNameStr)
	ctx := NewExecutionContext()
	item := mock.NewMockItem(storeService, []byte("item"), "item")

	// set
	err := item.Set(ctx, []byte("1000"))
	require.NoError(t, err)
	// Check if item was cached
	cacheContainer := ctx.Cache.GetContainer([]byte(actorNameStr))
	v, ok := cacheContainer.Get([]byte("item"))
	require.True(t, ok)
	require.Equal(t, []byte("1000"), v.([]byte))

	// get
	// Remove item from cache
	ctx.Cache.GetContainer([]byte(actorNameStr)).Remove([]byte("item"))

	i, err := item.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, []byte("1000"), i)

	// item.Get() should add item to cache
	cacheContainer = ctx.Cache.GetContainer([]byte(actorNameStr))
	v, ok = cacheContainer.Get([]byte("item"))
	require.True(t, ok)
	require.Equal(t, []byte("1000"), v.(uint64))

	// has
	has, err := item.Has(ctx)
	require.NoError(t, err)
	require.True(t, has)

	// remove
	err = item.Remove(ctx)
	require.NoError(t, err)
	has, err = item.Has(ctx)
	require.NoError(t, err)
	require.False(t, has)
	// Check if item was removed from cache
	cacheContainer = ctx.Cache.GetContainer([]byte(actorNameStr))
	v, ok = cacheContainer.Get([]byte("item"))
	require.False(t, ok)
}

func TestSTFCache(t *testing.T) {
	state := mock.DB()
	mockTx := mock.Tx{
		Sender:   []byte("sender"),
		Msg:      wrapperspb.Bool(true), // msg does not matter at all because our handler does nothing.
		GasLimit: 100_000,
	}

	s := NewSTF(
		func(ctx context.Context, msg transaction.Type) (msgResp transaction.Type, err error) {
			return nil, cacheSet(t, ctx, []byte("1"), "exec")
		},
		func(ctx context.Context, msg transaction.Type) (msgResp transaction.Type, err error) {
			return nil, err
		},
		func(ctx context.Context, txs []mock.Tx) error { return nil },
		func(ctx context.Context) error {
			return cacheSet(t, ctx, []byte("2"), "begin-block")
		},
		func(ctx context.Context) error {
			return cacheSet(t, ctx, []byte("3"), "end-block")
		},
		func(ctx context.Context, tx mock.Tx) error {
			return cacheSet(t, ctx, []byte("4"), "validate")
		},
		func(ctx context.Context) ([]appmodulev2.ValidatorUpdate, error) { return nil, nil },
		func(ctx context.Context, tx mock.Tx, success bool) error {
			return cacheSet(t, ctx, []byte("5"), "post-tx-exec")
		},
		branch.DefaultNewWriterMap,
	)

	t.Run("begin and end block", func(t *testing.T) {
		ctx := NewExecutionContext()
		_, _, err := s.DeliverBlock(ctx, &appmanager.BlockRequest[mock.Tx]{Txs: []mock.Tx{}}, state)
		require.NoError(t, err)
		cacheHas(t, tempCache, []byte("2"), "begin-block")
		cacheHas(t, tempCache, []byte("3"), "end-block")
	})

	t.Run("basic tx", func(t *testing.T) {
		_, _, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Txs: []mock.Tx{mockTx},
		}, state)
		require.NoError(t, err)
		cacheHas(t, tempCache, []byte("4"), "validate")
		cacheHas(t, tempCache, []byte("1"), "exec")
		cacheHas(t, tempCache, []byte("5"), "post-tx-exec")
	})
}

func cacheSet(t *testing.T, ctx context.Context, prefix []byte, v string) error {
	t.Helper()

	storeService := NewStoreService(actorNameStr)
	item := mock.NewMockItem(storeService, prefix, "item")

	err := item.Set(ctx, []byte(v))
	require.NoError(t, err)

	eCtx := GetExecutionContext(ctx)
	tempCache = eCtx.Cache

	return nil
}

func cacheHas(t *testing.T, cache ModuleContainer, prefix []byte, expected string) {
	t.Helper()
	v, ok := cache.GetContainer([]byte(actorNameStr)).Get(prefix)
	require.True(t, ok)
	require.Equal(t, expected, v.(string))
}
