package ormfield_test

import (
	"bytes"
	"fmt"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormfield"
	"github.com/cosmos/cosmos-sdk/orm/internal/testutil"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

func TestCodec(t *testing.T) {
	for _, ks := range testutil.TestFieldSpecs {
		testCodec(t, ks)
	}
}

func testCodec(t *testing.T, spec testutil.TestFieldSpec) {
	t.Run(fmt.Sprintf("%s %v", spec.FieldName, false), func(t *testing.T) {
		testCodecNT(t, spec.FieldName, spec.Gen, false)
	})
	t.Run(fmt.Sprintf("%s %v", spec.FieldName, true), func(t *testing.T) {
		testCodecNT(t, spec.FieldName, spec.Gen, true)
	})
}

func testCodecNT(t *testing.T, fname protoreflect.Name, generator *rapid.Generator, nonTerminal bool) {
	cdc, err := testutil.MakeTestCodec(fname, nonTerminal)
	assert.NilError(t, err)
	rapid.Check(t, func(t *rapid.T) {
		x := protoreflect.ValueOf(generator.Draw(t, string(fname)))
		bz1 := checkEncodeDecodeSize(t, x, cdc)
		if cdc.IsOrdered() {
			y := protoreflect.ValueOf(generator.Draw(t, fmt.Sprintf("%s 2", fname)))
			bz2 := checkEncodeDecodeSize(t, y, cdc)
			assert.Equal(t, cdc.Compare(x, y), bytes.Compare(bz1, bz2))
		}
	})
}

func checkEncodeDecodeSize(t *rapid.T, x protoreflect.Value, cdc ormfield.Codec) []byte {
	buf := &bytes.Buffer{}
	err := cdc.Encode(x, buf)
	assert.NilError(t, err)
	bz := buf.Bytes()
	size, err := cdc.ComputeBufferSize(x)
	assert.NilError(t, err)
	assert.Assert(t, size >= len(bz))
	fixedSize := cdc.FixedBufferSize()
	if fixedSize > 0 {
		assert.Equal(t, fixedSize, size)
	}
	y, err := cdc.Decode(bytes.NewReader(bz))
	assert.NilError(t, err)
	assert.Equal(t, 0, cdc.Compare(x, y))
	return bz
}

func TestUnsupportedFields(t *testing.T) {
	_, err := ormfield.GetCodec(nil, false)
	assert.ErrorContains(t, err, ormerrors.UnsupportedKeyField.Error())
	_, err = ormfield.GetCodec(testutil.GetTestField("repeated"), false)
	assert.ErrorContains(t, err, ormerrors.UnsupportedKeyField.Error())
	_, err = ormfield.GetCodec(testutil.GetTestField("map"), false)
	assert.ErrorContains(t, err, ormerrors.UnsupportedKeyField.Error())
	_, err = ormfield.GetCodec(testutil.GetTestField("msg"), false)
	assert.ErrorContains(t, err, ormerrors.UnsupportedKeyField.Error())
	_, err = ormfield.GetCodec(testutil.GetTestField("oneof"), false)
	assert.ErrorContains(t, err, ormerrors.UnsupportedKeyField.Error())
}

func TestNTBytesTooLong(t *testing.T) {
	cdc, err := ormfield.GetCodec(testutil.GetTestField("bz"), true)
	assert.NilError(t, err)
	buf := &bytes.Buffer{}
	bz := protoreflect.ValueOfBytes(make([]byte, 256))
	assert.ErrorContains(t, cdc.Encode(bz, buf), ormerrors.BytesFieldTooLong.Error())
	_, err = cdc.ComputeBufferSize(bz)
	assert.ErrorContains(t, err, ormerrors.BytesFieldTooLong.Error())
}

func TestDefaultValue(t *testing.T) {
	tests := []struct {
		name        string
		nonTerminal bool
		value       interface{}
	}{
		{name: "u32", value: uint32(0)},
		{name: "u64", value: uint64(0)},
		{name: "i32", value: int32(0)},
		{name: "i64", value: int64(0)},
		{name: "str", value: ""},
		{name: "str", nonTerminal: true, value: ""},
		{name: "bz", value: []byte{}},
		{name: "bz", nonTerminal: true, value: []byte{}},
		{name: "b", value: false},
		{name: "ts", value: (&timestamppb.Timestamp{}).ProtoReflect()},
		{name: "dur", value: (&durationpb.Duration{}).ProtoReflect()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field := testutil.GetTestField(protoreflect.Name(test.name))
			cdc, err := ormfield.GetCodec(field, test.nonTerminal)
			assert.NilError(t, err)
			assert.Equal(t, 0, cdc.Compare(protoreflect.ValueOf(test.value), cdc.DefaultValue()))
		})
	}
}
