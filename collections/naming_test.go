package collections

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections/codec"
)

func TestNaming(t *testing.T) {
	expectKeyCodecName(t, "u16", Uint16Key.WithName("u16"))
	expectKeyCodecName(t, "u32", Uint32Key.WithName("u32"))
	expectKeyCodecName(t, "u64", Uint64Key.WithName("u64"))
	expectKeyCodecName(t, "i32", Int32Key.WithName("i32"))
	expectKeyCodecName(t, "i64", Int64Key.WithName("i64"))
	expectKeyCodecName(t, "str", StringKey.WithName("str"))
	expectKeyCodecName(t, "bytes", BytesKey.WithName("bytes"))
	expectKeyCodecName(t, "bool", BoolKey.WithName("bool"))

	expectValueCodecName(t, "vu16", Uint16Value.WithName("vu16"))
	expectValueCodecName(t, "vu32", Uint32Value.WithName("vu32"))
	expectValueCodecName(t, "vu64", Uint64Value.WithName("vu64"))
	expectValueCodecName(t, "vi32", Int32Value.WithName("vi32"))
	expectValueCodecName(t, "vi64", Int64Value.WithName("vi64"))
	expectValueCodecName(t, "vstr", StringValue.WithName("vstr"))
	expectValueCodecName(t, "vbytes", BytesValue.WithName("vbytes"))
	expectValueCodecName(t, "vbool", BoolValue.WithName("vbool"))

	expectKeyCodecNames(t, NamedPairKeyCodec[bool, string]("abc", BoolKey, "def", StringKey), "abc", "def")
	expectKeyCodecNames(t, NamedTripleKeyCodec[bool, string, int32]("abc", BoolKey, "def", StringKey, "ghi", Int32Key), "abc", "def", "ghi")
	expectKeyCodecNames(t, NamedQuadKeyCodec[bool, string, int32, uint64]("abc", BoolKey, "def", StringKey, "ghi", Int32Key, "jkl", Uint64Key), "abc", "def", "ghi", "jkl")
}

func expectKeyCodecName[T any](t *testing.T, name string, cdc codec.KeyCodec[T]) {
	t.Helper()
	schema, err := codec.KeySchemaCodec(cdc)
	require.NoError(t, err)
	require.Equal(t, 1, len(schema.Fields))
	require.Equal(t, name, schema.Fields[0].Name)
}

func expectValueCodecName[T any](t *testing.T, name string, cdc codec.ValueCodec[T]) {
	t.Helper()
	schema, err := codec.ValueSchemaCodec(cdc)
	require.NoError(t, err)
	require.Equal(t, 1, len(schema.Fields))
	require.Equal(t, name, schema.Fields[0].Name)
}

func expectKeyCodecNames[T any](t *testing.T, cdc codec.KeyCodec[T], names ...string) {
	t.Helper()
	schema, err := codec.KeySchemaCodec(cdc)
	require.NoError(t, err)
	require.Equal(t, len(names), len(schema.Fields))
	for i, name := range names {
		require.Equal(t, name, schema.Fields[i].Name)
	}
}
