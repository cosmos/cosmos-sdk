package testutil

import (
	"orm/encoding/ormvalue"
	"strings"

	"pgregory.net/rapid"
)

type TestKeyPartSpec struct {
	FieldName string
	Gen       *rapid.Generator
}

var TestKeyPartSpecs = []TestKeyPartSpec{
	{
		"UINT32",
		rapid.Uint32(),
	},
	{
		"UINT64",
		rapid.Uint64(),
	},
	{
		"STRING",
		rapid.String().Filter(func(x string) bool {
			// filter out null terminators
			return strings.IndexByte(x, 0) < 0
		}),
	},
	{
		"BYTES",
		rapid.SliceOfN(rapid.Byte(), 0, 255),
	},
}

func MakeTestPartCodec(fname string, nonTerminal bool) (ormvalue.Codec, error) {
	return ormvalue.MakeCodec(GetTestField(fname), nonTerminal)
}
