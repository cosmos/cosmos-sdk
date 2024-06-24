package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNaming(t *testing.T) {
	require.Equal(t, "u16", Uint16Key.WithName("u16").Name())
	require.Equal(t, "u32", Uint32Key.WithName("u32").Name())
	require.Equal(t, "u64", Uint64Key.WithName("u64").Name())
	require.Equal(t, "i32", Int32Key.WithName("i32").Name())
	require.Equal(t, "i64", Int64Key.WithName("i64").Name())
	require.Equal(t, "str", StringKey.WithName("str").Name())
	require.Equal(t, "bytes", BytesKey.WithName("bytes").Name())
	require.Equal(t, "bool", BoolKey.WithName("bool").Name())

	require.Equal(t, "vu16", Uint16Value.WithName("vu16").Name())
	require.Equal(t, "vu32", Uint32Value.WithName("vu32").Name())
	require.Equal(t, "vu64", Uint64Value.WithName("vu64").Name())
	require.Equal(t, "vi32", Int32Value.WithName("vi32").Name())
	require.Equal(t, "vi64", Int64Value.WithName("vi64").Name())
	require.Equal(t, "vstr", StringValue.WithName("vstr").Name())
	require.Equal(t, "vbytes", BytesValue.WithName("vbytes").Name())
	require.Equal(t, "vbool", BoolValue.WithName("vbool").Name())

	require.Equal(t, "abc,def", NamedPairKeyCodec[bool, string]("abc", BoolKey, "def", StringKey).Name())
	require.Equal(t, "abc,def,ghi", NamedTripleKeyCodec[bool, string, int32]("abc", BoolKey, "def", StringKey, "ghi", Int32Key).Name())
}
