package testutil

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"

	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormvalue"
)

type TestKeyPartSpec struct {
	FieldName protoreflect.Name
	Gen       *rapid.Generator
}

var TestKeyPartSpecs = []TestKeyPartSpec{
	{
		"u32",
		rapid.Uint32(),
	},
	{
		"u64",
		rapid.Uint64(),
	},
	{
		"str",
		rapid.String().Filter(func(x string) bool {
			// filter out null terminators
			return strings.IndexByte(x, 0) < 0
		}),
	},
	{
		"bz",
		rapid.SliceOfN(rapid.Byte(), 0, 255),
	},
}

func MakeTestPartCodec(fname protoreflect.Name, nonTerminal bool) (ormvalue.Codec, error) {
	return ormvalue.MakeCodec(GetTestField(fname), nonTerminal)
}

func GetTestField(fname protoreflect.Name) protoreflect.FieldDescriptor {
	a := &testpb.A{}
	return a.ProtoReflect().Descriptor().Fields().ByName(fname)
}
