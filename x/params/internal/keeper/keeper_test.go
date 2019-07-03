package params

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
		{"key6", 1157279},
		{"key7", 9058701},
	}

	table := NewKeyTable(
		[]byte("key1"), int64(0),
		[]byte("key2"), int64(0),
		[]byte("key3"), int64(0),
		[]byte("key4"), int64(0),
		[]byte("key5"), int64(0),
		[]byte("key6"), int64(0),
		[]byte("key7"), int64(0),
		[]byte("extra1"), bool(false),
		[]byte("extra2"), string(""),
	)

	cdc, ctx, skey, _, keeper := testComponents()

	store := prefix.NewStore(ctx.KVStore(skey), []byte("test/"))
	space := keeper.Subspace("test").WithKeyTable(table)

	// Set params
	for i, kv := range kvs {
		require.NotPanics(t, func() { space.Set(ctx, []byte(kv.key), kv.param) }, "space.Set panics, tc #%d", i)
	}

	// Test space.Get
	for i, kv := range kvs {
		var param int64
		require.NotPanics(t, func() { space.Get(ctx, []byte(kv.key), &param) }, "space.Get panics, tc #%d", i)
		require.Equal(t, kv.param, param, "stored param not equal, tc #%d", i)
	}

	// Test space.GetRaw
	for i, kv := range kvs {
		var param int64
		bz := space.GetRaw(ctx, []byte(kv.key))
		err := cdc.UnmarshalJSON(bz, &param)
		require.Nil(t, err, "err is not nil, tc #%d", i)
		require.Equal(t, kv.param, param, "stored param not equal, tc #%d", i)
	}

	// Test store.Get equals space.Get
	for i, kv := range kvs {
		var param int64
		bz := store.Get([]byte(kv.key))
		require.NotNil(t, bz, "KVStore.Get returns nil, tc #%d", i)
		err := cdc.UnmarshalJSON(bz, &param)
		require.NoError(t, err, "UnmarshalJSON returns error, tc #%d", i)
		require.Equal(t, kv.param, param, "stored param not equal, tc #%d", i)
	}

	// Test invalid space.Get
	for i, kv := range kvs {
		var param bool
		require.Panics(t, func() { space.Get(ctx, []byte(kv.key), &param) }, "invalid space.Get not panics, tc #%d", i)
	}

	// Test invalid space.Set
	for i, kv := range kvs {
		require.Panics(t, func() { space.Set(ctx, []byte(kv.key), true) }, "invalid space.Set not panics, tc #%d", i)
	}

	// Test GetSubspace
	for i, kv := range kvs {
		var gparam, param int64
		gspace, ok := keeper.GetSubspace("test")
		require.True(t, ok, "cannot retrieve subspace, tc #%d", i)

		require.NotPanics(t, func() { gspace.Get(ctx, []byte(kv.key), &gparam) })
		require.NotPanics(t, func() { space.Get(ctx, []byte(kv.key), &param) })
		require.Equal(t, gparam, param, "GetSubspace().Get not equal with space.Get, tc #%d", i)

		require.NotPanics(t, func() { gspace.Set(ctx, []byte(kv.key), int64(i)) })
		require.NotPanics(t, func() { space.Get(ctx, []byte(kv.key), &param) })
		require.Equal(t, int64(i), param, "GetSubspace().Set not equal with space.Get, tc #%d", i)
	}
}

func indirect(ptr interface{}) interface{} {
	return reflect.ValueOf(ptr).Elem().Interface()
}

func TestSubspace(t *testing.T) {
	cdc, ctx, key, _, keeper := testComponents()

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

	table := NewKeyTable(
		[]byte("string"), string(""),
		[]byte("bool"), bool(false),
		[]byte("int16"), int16(0),
		[]byte("int32"), int32(0),
		[]byte("int64"), int64(0),
		[]byte("uint16"), uint16(0),
		[]byte("uint32"), uint32(0),
		[]byte("uint64"), uint64(0),
		[]byte("int"), sdk.Int{},
		[]byte("uint"), sdk.Uint{},
		[]byte("dec"), sdk.Dec{},
		[]byte("struct"), s{},
	)

	store := prefix.NewStore(ctx.KVStore(key), []byte("test/"))
	space := keeper.Subspace("test").WithKeyTable(table)

	// Test space.Set, space.Modified
	for i, kv := range kvs {
		require.False(t, space.Modified(ctx, []byte(kv.key)), "space.Modified returns true before setting, tc #%d", i)
		require.NotPanics(t, func() { space.Set(ctx, []byte(kv.key), kv.param) }, "space.Set panics, tc #%d", i)
		require.True(t, space.Modified(ctx, []byte(kv.key)), "space.Modified returns false after setting, tc #%d", i)
	}

	// Test space.Get, space.GetIfExists
	for i, kv := range kvs {
		require.NotPanics(t, func() { space.GetIfExists(ctx, []byte("invalid"), kv.ptr) }, "space.GetIfExists panics when no value exists, tc #%d", i)
		require.Equal(t, kv.zero, indirect(kv.ptr), "space.GetIfExists unmarshalls when no value exists, tc #%d", i)
		require.Panics(t, func() { space.Get(ctx, []byte("invalid"), kv.ptr) }, "invalid space.Get not panics when no value exists, tc #%d", i)
		require.Equal(t, kv.zero, indirect(kv.ptr), "invalid space.Get unmarshalls when no value exists, tc #%d", i)

		require.NotPanics(t, func() { space.GetIfExists(ctx, []byte(kv.key), kv.ptr) }, "space.GetIfExists panics, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "stored param not equal, tc #%d", i)
		require.NotPanics(t, func() { space.Get(ctx, []byte(kv.key), kv.ptr) }, "space.Get panics, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "stored param not equal, tc #%d", i)

		require.Panics(t, func() { space.Get(ctx, []byte("invalid"), kv.ptr) }, "invalid space.Get not panics when no value exists, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "invalid space.Get unmarshalls when no value existt, tc #%d", i)

		require.Panics(t, func() { space.Get(ctx, []byte(kv.key), nil) }, "invalid space.Get not panics when the pointer is nil, tc #%d", i)
		require.Panics(t, func() { space.Get(ctx, []byte(kv.key), new(invalid)) }, "invalid space.Get not panics when the pointer is different type, tc #%d", i)
	}

	// Test store.Get equals space.Get
	for i, kv := range kvs {
		bz := store.Get([]byte(kv.key))
		require.NotNil(t, bz, "store.Get() returns nil, tc #%d", i)
		err := cdc.UnmarshalJSON(bz, kv.ptr)
		require.NoError(t, err, "cdc.UnmarshalJSON() returns error, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "stored param not equal, tc #%d", i)
	}
}

type paramJSON struct {
	Param1 int64  `json:"param1,omitempty"`
	Param2 string `json:"param2,omitempty"`
}

func TestJSONUpdate(t *testing.T) {
	_, ctx, _, _, keeper := testComponents()

	key := []byte("key")

	space := keeper.Subspace("test").WithKeyTable(NewKeyTable(key, paramJSON{}))

	var param paramJSON

	space.Update(ctx, key, []byte(`{"param1": "10241024"}`))
	space.Get(ctx, key, &param)
	require.Equal(t, paramJSON{10241024, ""}, param)

	space.Update(ctx, key, []byte(`{"param2": "helloworld"}`))
	space.Get(ctx, key, &param)
	require.Equal(t, paramJSON{10241024, "helloworld"}, param)

	space.Update(ctx, key, []byte(`{"param1": "20482048"}`))
	space.Get(ctx, key, &param)
	require.Equal(t, paramJSON{20482048, "helloworld"}, param)

	space.Update(ctx, key, []byte(`{"param1": "40964096", "param2": "goodbyeworld"}`))
	space.Get(ctx, key, &param)
	require.Equal(t, paramJSON{40964096, "goodbyeworld"}, param)
}
