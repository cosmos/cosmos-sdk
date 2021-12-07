package ormkv_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"

	"google.golang.org/protobuf/reflect/protoreflect"
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/orm/internal/testutil"
)

func TestKeyCodec(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		key := testutil.TestKeyCodecGen(0, 5).Draw(t, "key").(testutil.TestKeyCodec)
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
	cdc, err := ormkv.NewKeyCodec(nil,
		(&testpb.A{}).ProtoReflect().Descriptor(),
		[]protoreflect.Name{"u32", "str", "i32"})
	assert.NilError(t, err)

	tests := []struct {
		name       string
		values1    []protoreflect.Value
		values2    []protoreflect.Value
		expect     int
		validRange bool
	}{
		{
			"eq",
			testutil.ValuesOf(uint32(0), "abc", int32(-3)),
			testutil.ValuesOf(uint32(0), "abc", int32(-3)),
			0,
			false,
		},
		{
			"eq prefix 0",
			testutil.ValuesOf(),
			testutil.ValuesOf(),
			0,
			false,
		},
		{
			"eq prefix 1",
			testutil.ValuesOf(uint32(0)),
			testutil.ValuesOf(uint32(0)),
			0,
			false,
		},
		{
			"eq prefix 2",
			testutil.ValuesOf(uint32(0), "abc"),
			testutil.ValuesOf(uint32(0), "abc"),
			0,
			false,
		},
		{
			"lt1",
			testutil.ValuesOf(uint32(0), "abc", int32(-3)),
			testutil.ValuesOf(uint32(1), "abc", int32(-3)),
			-1,
			true,
		},
		{
			"lt2",
			testutil.ValuesOf(uint32(1), "abb", int32(-3)),
			testutil.ValuesOf(uint32(1), "abc", int32(-3)),
			-1,
			true,
		},
		{
			"lt3",
			testutil.ValuesOf(uint32(1), "abb", int32(-4)),
			testutil.ValuesOf(uint32(1), "abb", int32(-3)),
			-1,
			true,
		},
		{
			"less prefix 0",
			testutil.ValuesOf(),
			testutil.ValuesOf(uint32(1), "abb", int32(-4)),
			-1,
			true,
		},
		{
			"less prefix 1",
			testutil.ValuesOf(uint32(1)),
			testutil.ValuesOf(uint32(1), "abb", int32(-4)),
			-1,
			true,
		},
		{
			"less prefix 2",
			testutil.ValuesOf(uint32(1), "abb"),
			testutil.ValuesOf(uint32(1), "abb", int32(-4)),
			-1,
			true,
		},
		{
			"gt1",
			testutil.ValuesOf(uint32(2), "abb", int32(-4)),
			testutil.ValuesOf(uint32(1), "abb", int32(-4)),
			1,
			false,
		},
		{
			"gt2",
			testutil.ValuesOf(uint32(2), "abc", int32(-4)),
			testutil.ValuesOf(uint32(2), "abb", int32(-4)),
			1,
			false,
		},
		{
			"gt3",
			testutil.ValuesOf(uint32(2), "abc", int32(1)),
			testutil.ValuesOf(uint32(2), "abc", int32(-3)),
			1,
			false,
		},
		{
			"gt prefix 0",
			testutil.ValuesOf(uint32(2), "abc", int32(-3)),
			testutil.ValuesOf(),
			1,
			true,
		},
		{
			"gt prefix 1",
			testutil.ValuesOf(uint32(2), "abc", int32(-3)),
			testutil.ValuesOf(uint32(2)),
			1,
			true,
		},
		{
			"gt prefix 2",
			testutil.ValuesOf(uint32(2), "abc", int32(-3)),
			testutil.ValuesOf(uint32(2), "abc"),
			1,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(
				t, test.expect,
				cdc.CompareValues(test.values1, test.values2),
			)
			// CheckValidRangeIterationKeys should give comparable results
			err := cdc.CheckValidRangeIterationKeys(test.values1, test.values2)
			if test.validRange {
				assert.NilError(t, err)
			} else {
				assert.ErrorContains(t, err, "")
			}
		})
	}
}

func TestDecodePrefixKey(t *testing.T) {
	cdc, err := ormkv.NewKeyCodec(nil,
		(&testpb.A{}).ProtoReflect().Descriptor(),
		[]protoreflect.Name{"u32", "str", "bz", "i32"})

	assert.NilError(t, err)
	tests := []struct {
		name   string
		values []protoreflect.Value
	}{
		{
			"1",
			testutil.ValuesOf(uint32(5), "abc"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bz, err := cdc.Encode(test.values)
			assert.NilError(t, err)
			values, err := cdc.Decode(bytes.NewReader(bz))
			assert.ErrorType(t, err, io.EOF)
			assert.Equal(t, 0, cdc.CompareValues(test.values, values))
		})
	}
}

func TestValidRangeIterationKeys(t *testing.T) {
	cdc, err := ormkv.NewKeyCodec(nil,
		(&testpb.A{}).ProtoReflect().Descriptor(),
		[]protoreflect.Name{"u32", "str", "bz", "i32"})
	assert.NilError(t, err)

	tests := []struct {
		name      string
		values1   []protoreflect.Value
		values2   []protoreflect.Value
		expectErr bool
	}{
		{
			"1 eq",
			testutil.ValuesOf(uint32(0)),
			testutil.ValuesOf(uint32(0)),
			true,
		},
		{
			"1 lt",
			testutil.ValuesOf(uint32(0)),
			testutil.ValuesOf(uint32(1)),
			false,
		},
		{
			"1 gt",
			testutil.ValuesOf(uint32(1)),
			testutil.ValuesOf(uint32(0)),
			true,
		},
		{
			"1,2 lt",
			testutil.ValuesOf(uint32(0)),
			testutil.ValuesOf(uint32(0), "abc"),
			false,
		},
		{
			"1,2 gt",
			testutil.ValuesOf(uint32(0), "abc"),
			testutil.ValuesOf(uint32(0)),
			false,
		},
		{
			"1,2,3",
			testutil.ValuesOf(uint32(0)),
			testutil.ValuesOf(uint32(0), "abc", []byte{1, 2}),
			true,
		},
		{
			"1,2,3,4 lt",
			testutil.ValuesOf(uint32(0), "abc", []byte{1, 2}, int32(-1)),
			testutil.ValuesOf(uint32(0), "abc", []byte{1, 2}, int32(1)),
			false,
		},
		{
			"too long",
			testutil.ValuesOf(uint32(0), "abc", []byte{1, 2}, int32(-1)),
			testutil.ValuesOf(uint32(0), "abc", []byte{1, 2}, int32(1), int32(1)),
			true,
		},
		{
			"1,2,3,4 eq",
			testutil.ValuesOf(uint32(0), "abc", []byte{1, 2}, int32(1)),
			testutil.ValuesOf(uint32(0), "abc", []byte{1, 2}, int32(1)),
			true,
		},
		{
			"1,2,3,4 bz err",
			testutil.ValuesOf(uint32(0), "abc", []byte{1, 2}, int32(-1)),
			testutil.ValuesOf(uint32(0), "abc", []byte{1, 2, 3}, int32(1)),
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := cdc.CheckValidRangeIterationKeys(test.values1, test.values2)
			if test.expectErr {
				assert.ErrorContains(t, err, "")
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestGetSet(t *testing.T) {
	cdc, err := ormkv.NewKeyCodec(nil,
		(&testpb.A{}).ProtoReflect().Descriptor(),
		[]protoreflect.Name{"u32", "str", "i32"})
	assert.NilError(t, err)

	var a testpb.A
	values := testutil.ValuesOf(uint32(4), "abc", int32(1))
	cdc.SetValues(a.ProtoReflect(), values)
	values2 := cdc.GetValues(a.ProtoReflect())
	assert.Equal(t, 0, cdc.CompareValues(values, values2))
	bz, err := cdc.Encode(values)
	assert.NilError(t, err)
	values3, bz2, err := cdc.EncodeFromMessage(a.ProtoReflect())
	assert.NilError(t, err)
	assert.Equal(t, 0, cdc.CompareValues(values, values3))
	assert.Assert(t, bytes.Equal(bz, bz2))
}
