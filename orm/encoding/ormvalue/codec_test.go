package ormvalue_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormvalue"
	"github.com/cosmos/cosmos-sdk/orm/internal/testutil"

	"gotest.tools/assert"

	"google.golang.org/protobuf/reflect/protoreflect"
	"pgregory.net/rapid"
)

func TestPartCodec(t *testing.T) {
	for _, ks := range testutil.TestKeyPartSpecs {
		testKeyPartCodec(t, ks)
	}
}
func testKeyPartCodec(t *testing.T, spec testutil.TestKeyPartSpec) {
	t.Run(fmt.Sprintf("%s %v", spec.FieldName, false), func(t *testing.T) {
		testKeyPartCodecNT(t, spec.FieldName, spec.Gen, false)
	})
	t.Run(fmt.Sprintf("%s %v", spec.FieldName, true), func(t *testing.T) {
		testKeyPartCodecNT(t, spec.FieldName, spec.Gen, true)
	})
}

func testKeyPartCodecNT(t *testing.T, fname protoreflect.Name, generator *rapid.Generator, nonTerminal bool) {
	cdc, err := testutil.MakeTestPartCodec(fname, nonTerminal)
	assert.NilError(t, err)
	rapid.Check(t, func(t *rapid.T) {
		x := protoreflect.ValueOf(generator.Draw(t, string(fname)))
		bz1 := assertEncDecPart(t, x, cdc)
		if cdc.IsOrdered() {
			y := protoreflect.ValueOf(generator.Draw(t, string(fname+"2")))
			bz2 := assertEncDecPart(t, y, cdc)
			assert.Equal(t, cdc.Compare(x, y), bytes.Compare(bz1, bz2))
		}
	})
}

func assertEncDecPart(t *rapid.T, x protoreflect.Value, cdc ormvalue.Codec) []byte {
	buf := &bytes.Buffer{}
	err := cdc.Encode(x, buf)
	assert.NilError(t, err)
	bz := buf.Bytes()
	size, err := cdc.Size(x)
	assert.NilError(t, err)
	assert.Equal(t, size, len(bz))
	fixedSize := cdc.FixedSize()
	if fixedSize > 0 {
		assert.Equal(t, fixedSize, size)
	}
	y, err := cdc.Decode(bytes.NewReader(bz))
	assert.NilError(t, err)
	assert.Equal(t, 0, cdc.Compare(x, y))
	return bz
}
