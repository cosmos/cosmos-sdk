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
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	cdc.RegisterConcrete(s{}, "test/s", nil)
	return cdc
}

func TestKeeper(t *testing.T) {
	kvs := []struct {
		key   string
		param int64
	}{
		{"key1", 10},
		{"key2", 55},
		{"key3", 182},
		{"key4", 17582},
		{"key5", 2768554},
		{"store1/key1", 1157279},
		{"store1/key2", 9058701},
	}

	skey := sdk.NewKVStoreKey("test")
	tkey := sdk.NewTransientStoreKey("transient_test")
	ctx := defaultContext(skey, tkey)
	store := NewKeeper(codec.New(), skey, tkey).Substore("test")

	for _, kv := range kvs {
		require.NotPanics(t, func() { store.Set(ctx, kv.key, kv.param) })
	}

	for _, kv := range kvs {
		var param int64
		require.NotPanics(t, func() { store.Get(ctx, kv.key, &param) })
		require.Equal(t, kv.param, param)
	}

	cdc := codec.New()
	for _, kv := range kvs {
		var param int64
		bz := store.GetRaw(ctx, kv.key)
		err := cdc.UnmarshalJSON(bz, &param)
		require.Nil(t, err)
		require.Equal(t, kv.param, param)
	}

	for _, kv := range kvs {
		var param bool
		require.Panics(t, func() { store.Get(ctx, kv.key, &param) })
	}

	for _, kv := range kvs {
		require.Panics(t, func() { store.Set(ctx, kv.key, true) })
	}
}

func TestGet(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	tkey := sdk.NewTransientStoreKey("transient_test")
	ctx := defaultContext(key, tkey)
	keeper := NewKeeper(createTestCodec(), key, tkey)

	store := keeper.Substore("test")

	kvs := []struct {
		key   string
		param interface{}
		zero  interface{}
		ptr   interface{}
	}{
		{"string", "test", "", new(string)},
		{"bool", true, false, new(bool)},
		{"int16", int16(1), int16(0), new(int16)},
		{"int32", int32(1), int32(0), new(int32)},
		{"int64", int64(1), int64(0), new(int64)},
		{"uint16", uint16(1), uint16(0), new(uint16)},
		{"uint32", uint32(1), uint32(0), new(uint32)},
		{"uint64", uint64(1), uint64(0), new(uint64)},
		/*
			{NewKey("int"), sdk.NewInt(1), *new(sdk.Int), new(sdk.Int)},
			{NewKey("uint"), sdk.NewUint(1), *new(sdk.Uint), new(sdk.Uint)},
			{NewKey("dec"), sdk.NewDec(1), *new(sdk.Dec), new(sdk.Dec)},
		*/
	}

	for _, kv := range kvs {
		require.NotPanics(t, func() { store.Set(ctx, kv.key, kv.param) })
	}

	for _, kv := range kvs {
		require.NotPanics(t, func() { store.GetIfExists(ctx, "invalid", kv.ptr) })
		require.Equal(t, kv.zero, reflect.ValueOf(kv.ptr).Elem().Interface())
		require.Panics(t, func() { store.Get(ctx, "invalid", kv.ptr) })
		require.Equal(t, kv.zero, reflect.ValueOf(kv.ptr).Elem().Interface())

		require.NotPanics(t, func() { store.GetIfExists(ctx, kv.key, kv.ptr) })
		require.Equal(t, kv.param, reflect.ValueOf(kv.ptr).Elem().Interface())
		require.NotPanics(t, func() { store.Get(ctx, kv.key, kv.ptr) })
		require.Equal(t, kv.param, reflect.ValueOf(kv.ptr).Elem().Interface())

		require.Panics(t, func() { store.Get(ctx, "invalid", kv.ptr) })
		require.Equal(t, kv.param, reflect.ValueOf(kv.ptr).Elem().Interface())

		require.Panics(t, func() { store.Get(ctx, kv.key, nil) })
		require.Panics(t, func() { store.Get(ctx, kv.key, new(s)) })
	}
}
