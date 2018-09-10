package params

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func defaultContext(key sdk.StoreKey, tkey sdk.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tkey, sdk.StoreTypeTransient, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	return ctx
}

type s struct{}

func createTestCodec() *codec.Codec {
	cdc := codec.NewCodec()
	sdk.RegisterWire(cdc)
	cdc.RegisterConcrete(s{}, "test/s", nil)
	return cdc
}

func TestKeeper(t *testing.T) {
	kvs := []struct {
		key   Key
		param int64
	}{
		{NewKey("key1"), 10},
		{NewKey("key2"), 55},
		{NewKey("key3"), 182},
		{NewKey("key4"), 17582},
		{NewKey("key5"), 2768554},
		{NewKey("space1", "key1"), 1157279},
		{NewKey("space1", "key2"), 9058701},
	}

	skey := sdk.NewKVStoreKey("test")
	tkey := sdk.NewTransientStoreKey("transient_test")
	ctx := defaultContext(skey, tkey)
	space := NewKeeper(codec.NewCodec(), skey, tkey).Subspace("test")

	for _, kv := range kvs {
		require.NotPanics(t, func() { space.Set(ctx, kv.key, kv.param) })
	}

	for _, kv := range kvs {
		var param int64
		require.NotPanics(t, func() { space.Get(ctx, kv.key, &param) })
		require.Equal(t, kv.param, param)
	}

	cdc := codec.New()
	for _, kv := range kvs {
		var param int64
		bz := space.GetRaw(ctx, kv.key)
		err := cdc.UnmarshalJSON(bz, &param)
		require.Nil(t, err)
		require.Equal(t, kv.param, param)
	}

	for _, kv := range kvs {
		var param bool
		require.Panics(t, func() { space.Get(ctx, kv.key, &param) })
	}

	for _, kv := range kvs {
		require.Panics(t, func() { space.Set(ctx, kv.key, true) })
	}
}

func TestGet(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	tkey := sdk.NewTransientStoreKey("transient_test")
	ctx := defaultContext(key, tkey)
	keeper := NewKeeper(createTestCodec(), key, tkey)

	space := keeper.Subspace("test")

	kvs := []struct {
		key   Key
		param interface{}
		zero  interface{}
		ptr   interface{}
	}{
		{NewKey("string"), "test", "", new(string)},
		{NewKey("bool"), true, false, new(bool)},
		{NewKey("int16"), int16(1), int16(0), new(int16)},
		{NewKey("int32"), int32(1), int32(0), new(int32)},
		{NewKey("int64"), int64(1), int64(0), new(int64)},
		{NewKey("uint16"), uint16(1), uint16(0), new(uint16)},
		{NewKey("uint32"), uint32(1), uint32(0), new(uint32)},
		{NewKey("uint64"), uint64(1), uint64(0), new(uint64)},
		/*
			{NewKey("int"), sdk.NewInt(1), *new(sdk.Int), new(sdk.Int)},
			{NewKey("uint"), sdk.NewUint(1), *new(sdk.Uint), new(sdk.Uint)},
			{NewKey("dec"), sdk.NewDec(1), *new(sdk.Dec), new(sdk.Dec)},
		*/
	}

	for _, kv := range kvs {
		require.NotPanics(t, func() { space.Set(ctx, kv.key, kv.param) })
	}

	for _, kv := range kvs {
		require.NotPanics(t, func() { space.GetIfExists(ctx, NewKey("invalid"), kv.ptr) })
		require.Equal(t, kv.zero, reflect.ValueOf(kv.ptr).Elem().Interface())
		require.Panics(t, func() { space.Get(ctx, NewKey("invalid"), kv.ptr) })
		require.Equal(t, kv.zero, reflect.ValueOf(kv.ptr).Elem().Interface())

		require.NotPanics(t, func() { space.GetIfExists(ctx, kv.key, kv.ptr) })
		require.Equal(t, kv.param, reflect.ValueOf(kv.ptr).Elem().Interface())
		require.NotPanics(t, func() { space.Get(ctx, kv.key, kv.ptr) })
		require.Equal(t, kv.param, reflect.ValueOf(kv.ptr).Elem().Interface())

		require.Panics(t, func() { space.Get(ctx, NewKey("invalid"), kv.ptr) })
		require.Equal(t, kv.param, reflect.ValueOf(kv.ptr).Elem().Interface())

		require.Panics(t, func() { space.Get(ctx, kv.key, nil) })
		require.Panics(t, func() { space.Get(ctx, kv.key, new(s)) })
	}
}
