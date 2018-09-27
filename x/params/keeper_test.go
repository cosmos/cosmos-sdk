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

type invalid struct{}

type s struct {
	I int
}

func createTestCodec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	cdc.RegisterConcrete(s{}, "test/s", nil)
	cdc.RegisterConcrete(invalid{}, "test/invalid", nil)
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

	for i, kv := range kvs {
		require.NotPanics(t, func() { store.Set(ctx, []byte(kv.key), kv.param) }, "store.Set panics, tc #%d", i)
	}

	for i, kv := range kvs {
		var param int64
		require.NotPanics(t, func() { store.Get(ctx, []byte(kv.key), &param) }, "store.Get panics, tc #%d", i)
		require.Equal(t, kv.param, param, "stored param not equal, tc #%d", i)
	}

	cdc := codec.New()
	for i, kv := range kvs {
		var param int64
		bz := store.GetRaw(ctx, []byte(kv.key))
		err := cdc.UnmarshalJSON(bz, &param)
		require.Nil(t, err, "err is not nil, tc #%d", i)
		require.Equal(t, kv.param, param, "stored param not equal, tc #%d", i)
	}

	for i, kv := range kvs {
		var param bool
		require.Panics(t, func() { store.Get(ctx, []byte(kv.key), &param) }, "invalid store.Get not panics, tc #%d", i)
	}

	for i, kv := range kvs {
		require.Panics(t, func() { store.Set(ctx, []byte(kv.key), true) }, "invalid store.Set not panics, tc #%d", i)
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
		{"int", sdk.NewInt(1), *new(sdk.Int), new(sdk.Int)},
		{"uint", sdk.NewUint(1), *new(sdk.Uint), new(sdk.Uint)},
		{"dec", sdk.NewDec(1), *new(sdk.Dec), new(sdk.Dec)},
		{"struct", s{1}, s{0}, new(s)},
	}

	for i, kv := range kvs {
		require.False(t, store.Modified(ctx, []byte(kv.key)), "store.Modified returns true before setting, tc #%d", i)
		require.NotPanics(t, func() { store.Set(ctx, []byte(kv.key), kv.param) }, "store.Set panics, tc #%d", i)
		require.True(t, store.Modified(ctx, []byte(kv.key)), "store.Modified returns false after setting, tc #%d", i)
	}

	for i, kv := range kvs {
		require.NotPanics(t, func() { store.GetIfExists(ctx, []byte("invalid"), kv.ptr) }, "store.GetIfExists panics when no value exists, tc #%d", i)
		require.Equal(t, kv.zero, reflect.ValueOf(kv.ptr).Elem().Interface(), "store.GetIfExists unmarshalls when no value exists, tc #%d", i)
		require.Panics(t, func() { store.Get(ctx, []byte("invalid"), kv.ptr) }, "invalid store.Get not panics when no value exists, tc #%d", i)
		require.Equal(t, kv.zero, reflect.ValueOf(kv.ptr).Elem().Interface(), "invalid store.Get unmarshalls when no value exists, tc #%d", i)

		require.NotPanics(t, func() { store.GetIfExists(ctx, []byte(kv.key), kv.ptr) }, "store.GetIfExists panics, tc #%d", i)
		require.Equal(t, kv.param, reflect.ValueOf(kv.ptr).Elem().Interface(), "stored param not equal, tc #%d", i)
		require.NotPanics(t, func() { store.Get(ctx, []byte(kv.key), kv.ptr) }, "store.Get panics, tc #%d", i)
		require.Equal(t, kv.param, reflect.ValueOf(kv.ptr).Elem().Interface(), "stored param not equal, tc #%d", i)

		require.Panics(t, func() { store.Get(ctx, []byte("invalid"), kv.ptr) }, "invalid store.Get not panics when no value exists, tc #%d", i)
		require.Equal(t, kv.param, reflect.ValueOf(kv.ptr).Elem().Interface(), "invalid store.Get unmarshalls when no value existt, tc #%d", i)

		require.Panics(t, func() { store.Get(ctx, []byte(kv.key), nil) }, "invalid store.Get not panics when the pointer is nil, tc #%d", i)
		require.Panics(t, func() { store.Get(ctx, []byte(kv.key), new(invalid)) }, "invalid store.Get not panics when the pointer is different type, tc #%d", i)
	}
}
