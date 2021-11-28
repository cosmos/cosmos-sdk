package ormkv_test

import (
	"bytes"
	"testing"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"

	"google.golang.org/protobuf/reflect/protoreflect"
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/orm/internal/testutil"
)

func TestKeyCodec(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		key := testutil.TestKeyCodecGen.Draw(t, "key").(testutil.TestKeyCodec)
		for i := 0; i < 100; i++ {
			keyValues := key.Draw(t, "values")

			bz1 := assertEncDecKey(t, key, keyValues)

			if key.Codec.IsFullyOrdered() {
				// check if ordered keys have ordered encodings
				keyValues2 := key.Draw(t, "values2")
				bz2 := assertEncDecKey(t, key, keyValues2)
				// bytes comparison should equal comparison of values
				assert.Equal(t, key.Codec.CompareValues(keyValues, keyValues2), bytes.Compare(bz1, bz2))
			}
		}
	})
}

func assertEncDecKey(t *rapid.T, key testutil.TestKeyCodec, keyValues []protoreflect.Value) []byte {
	bz, err := key.Codec.Encode(keyValues)
	assert.NilError(t, err)
	keyValues2, err := key.Codec.Decode(bytes.NewReader(bz))
	assert.NilError(t, err)
	assert.Equal(t, 0, key.Codec.CompareValues(keyValues, keyValues2))
	return bz
}

func TestCompareValues(t *testing.T) {
	cdc, err := ormkv.NewKeyCodec(nil, []protoreflect.FieldDescriptor{
		testutil.GetTestField("u32"),
		testutil.GetTestField("str"),
		testutil.GetTestField("i32"),
	})
	assert.NilError(t, err)

	tests := []struct {
		name    string
		values1 []protoreflect.Value
		values2 []protoreflect.Value
		expect  int
	}{
		{
			"eq",
			ValuesOf(uint32(0), "abc", int32(-3)),
			ValuesOf(uint32(0), "abc", int32(-3)),
			0,
		},
		{
			"eq prefix 0",
			ValuesOf(),
			ValuesOf(),
			0,
		},
		{
			"eq prefix 1",
			ValuesOf(uint32(0)),
			ValuesOf(uint32(0)),
			0,
		},
		{
			"eq prefix 2",
			ValuesOf(uint32(0), "abc"),
			ValuesOf(uint32(0), "abc"),
			0,
		},
		{
			"lt1",
			ValuesOf(uint32(0), "abc", int32(-3)),
			ValuesOf(uint32(1), "abc", int32(-3)),
			-1,
		},
		{
			"lt2",
			ValuesOf(uint32(1), "abb", int32(-3)),
			ValuesOf(uint32(1), "abc", int32(-3)),
			-1,
		},
		{
			"lt3",
			ValuesOf(uint32(1), "abb", int32(-4)),
			ValuesOf(uint32(1), "abb", int32(-3)),
			-1,
		},
		{
			"less prefix 0",
			ValuesOf(),
			ValuesOf(uint32(1), "abb", int32(-4)),
			-1,
		},
		{
			"less prefix 1",
			ValuesOf(uint32(1)),
			ValuesOf(uint32(1), "abb", int32(-4)),
			-1,
		},
		{
			"less prefix 2",
			ValuesOf(uint32(1), "abb"),
			ValuesOf(uint32(1), "abb", int32(-4)),
			-1,
		},
		{
			"gt1",
			ValuesOf(uint32(2), "abb", int32(-4)),
			ValuesOf(uint32(1), "abb", int32(-4)),
			1,
		},
		{
			"gt2",
			ValuesOf(uint32(2), "abc", int32(-4)),
			ValuesOf(uint32(2), "abb", int32(-4)),
			1,
		},
		{
			"gt3",
			ValuesOf(uint32(2), "abc", int32(1)),
			ValuesOf(uint32(2), "abc", int32(-3)),
			1,
		},
		{
			"gt prefix 0",
			ValuesOf(uint32(2), "abc", int32(-3)),
			ValuesOf(),
			1,
		},
		{
			"gt prefix 1",
			ValuesOf(uint32(2), "abc", int32(-3)),
			ValuesOf(uint32(2)),
			1,
		},
		{
			"gt prefix 2",
			ValuesOf(uint32(2), "abc", int32(-3)),
			ValuesOf(uint32(2), "abc"),
			1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(
				t, test.expect,
				cdc.CompareValues(test.values1, test.values2),
			)
		})
	}
}

func ValuesOf(values ...interface{}) []protoreflect.Value {
	n := len(values)
	res := make([]protoreflect.Value, n)
	for i := 0; i < n; i++ {
		res[i] = protoreflect.ValueOf(values[i])
	}
	return res
}
