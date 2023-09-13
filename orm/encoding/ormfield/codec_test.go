package ormfield_test

import (
	"bytes"
	"fmt"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"cosmossdk.io/orm/encoding/ormfield"
	"cosmossdk.io/orm/internal/testutil"
	"cosmossdk.io/orm/types/ormerrors"
)

func TestCodec(t *testing.T) {
	for _, ks := range testutil.TestFieldSpecs {
		testCodec(t, ks)
	}
}

func testCodec(t *testing.T, spec testutil.TestFieldSpec) {
	t.Helper()
	t.Run(fmt.Sprintf("%s %v", spec.FieldName, false), func(t *testing.T) {
		testCodecNT(t, spec.FieldName, spec.Gen, false)
	})
	t.Run(fmt.Sprintf("%s %v", spec.FieldName, true), func(t *testing.T) {
		testCodecNT(t, spec.FieldName, spec.Gen, true)
	})
}

func testCodecNT(t *testing.T, fname protoreflect.Name, generator *rapid.Generator[any], nonTerminal bool) {
	t.Helper()
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
	assert.ErrorContains(t, err, ormerrors.InvalidKeyField.Error())
	_, err = ormfield.GetCodec(testutil.GetTestField("repeated"), false)
	assert.ErrorContains(t, err, ormerrors.InvalidKeyField.Error())
	_, err = ormfield.GetCodec(testutil.GetTestField("map"), false)
	assert.ErrorContains(t, err, ormerrors.InvalidKeyField.Error())
	_, err = ormfield.GetCodec(testutil.GetTestField("msg"), false)
	assert.ErrorContains(t, err, ormerrors.InvalidKeyField.Error())
	_, err = ormfield.GetCodec(testutil.GetTestField("oneof"), false)
	assert.ErrorContains(t, err, ormerrors.InvalidKeyField.Error())
}

func TestCompactUInt32(t *testing.T) {
	var lastBz []byte
	testEncodeDecode := func(x uint32, expectedLen int) {
		bz := ormfield.EncodeCompactUint32(x)
		assert.Equal(t, expectedLen, len(bz))
		y, err := ormfield.DecodeCompactUint32(bytes.NewReader(bz))
		assert.NilError(t, err)
		assert.Equal(t, x, y)
		assert.Assert(t, bytes.Compare(lastBz, bz) < 0)
		lastBz = bz
	}

	testEncodeDecode(64, 2)
	testEncodeDecode(16383, 2)
	testEncodeDecode(16384, 3)
	testEncodeDecode(4194303, 3)
	testEncodeDecode(4194304, 4)
	testEncodeDecode(1073741823, 4)
	testEncodeDecode(1073741824, 5)

	// randomized tests
	rapid.Check(t, func(t *rapid.T) {
		x := rapid.Uint32().Draw(t, "x")
		y := rapid.Uint32().Draw(t, "y")

		bx := ormfield.EncodeCompactUint32(x)
		by := ormfield.EncodeCompactUint32(y)

		cmp := bytes.Compare(bx, by)
		switch {
		case x < y:
			assert.Equal(t, -1, cmp)
		case x == y:
			assert.Equal(t, 0, cmp)
		default:
			assert.Equal(t, 1, cmp)
		}

		x2, err := ormfield.DecodeCompactUint32(bytes.NewReader(bx))
		assert.NilError(t, err)
		assert.Equal(t, x, x2)
		y2, err := ormfield.DecodeCompactUint32(bytes.NewReader(by))
		assert.NilError(t, err)
		assert.Equal(t, y, y2)
	})
}

func TestCompactUInt64(t *testing.T) {
	var lastBz []byte
	testEncodeDecode := func(x uint64, expectedLen int) {
		bz := ormfield.EncodeCompactUint64(x)
		assert.Equal(t, expectedLen, len(bz))
		y, err := ormfield.DecodeCompactUint64(bytes.NewReader(bz))
		assert.NilError(t, err)
		assert.Equal(t, x, y)
		assert.Assert(t, bytes.Compare(lastBz, bz) < 0)
		lastBz = bz
	}

	testEncodeDecode(64, 2)
	testEncodeDecode(16383, 2)
	testEncodeDecode(16384, 4)
	testEncodeDecode(4194303, 4)
	testEncodeDecode(4194304, 4)
	testEncodeDecode(1073741823, 4)
	testEncodeDecode(1073741824, 6)
	testEncodeDecode(70368744177663, 6)
	testEncodeDecode(70368744177664, 9)

	// randomized tests
	rapid.Check(t, func(t *rapid.T) {
		x := rapid.Uint64().Draw(t, "x")
		y := rapid.Uint64().Draw(t, "y")

		bx := ormfield.EncodeCompactUint64(x)
		by := ormfield.EncodeCompactUint64(y)

		cmp := bytes.Compare(bx, by)
		switch {
		case x < y:
			assert.Equal(t, -1, cmp)
		case x == y:
			assert.Equal(t, 0, cmp)
		default:
			assert.Equal(t, 1, cmp)
		}

		x2, err := ormfield.DecodeCompactUint64(bytes.NewReader(bx))
		assert.NilError(t, err)
		assert.Equal(t, x, x2)
		y2, err := ormfield.DecodeCompactUint64(bytes.NewReader(by))
		assert.NilError(t, err)
		assert.Equal(t, y, y2)
	})
}
