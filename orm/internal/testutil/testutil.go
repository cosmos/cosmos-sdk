package testutil

import (
	"fmt"
	"math"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormfield"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
)

// TestFieldSpec defines a test field against the testpb.ExampleTable message.
type TestFieldSpec struct {
	FieldName protoreflect.Name
	Gen       *rapid.Generator
}

var TestFieldSpecs = []TestFieldSpec{
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
		rapid.SliceOfN(rapid.Byte(), 0, math.MaxUint32),
	},
	{
		"i32",
		rapid.Int32(),
	},
	{
		"f32",
		rapid.Uint32(),
	},
	{
		"s32",
		rapid.Int32(),
	},
	{
		"sf32",
		rapid.Int32(),
	},
	{
		"i64",
		rapid.Int64(),
	},
	{
		"f64",
		rapid.Uint64(),
	},
	{
		"s64",
		rapid.Int64(),
	},
	{
		"sf64",
		rapid.Int64(),
	},
	{
		"b",
		rapid.Bool(),
	},
	{
		"ts",
		rapid.Custom(func(t *rapid.T) protoreflect.Message {
			seconds := rapid.Int64Range(-9999999999, 9999999999).Draw(t, "seconds").(int64)
			nanos := rapid.Int32Range(0, 999999999).Draw(t, "nanos").(int32)
			return (&timestamppb.Timestamp{
				Seconds: seconds,
				Nanos:   nanos,
			}).ProtoReflect()
		}),
	},
	{
		"dur",
		rapid.Custom(func(t *rapid.T) protoreflect.Message {
			seconds := rapid.Int64Range(0, 315576000000).Draw(t, "seconds").(int64)
			nanos := rapid.Int32Range(0, 999999999).Draw(t, "nanos").(int32)
			return (&durationpb.Duration{
				Seconds: seconds,
				Nanos:   nanos,
			}).ProtoReflect()
		}),
	},
	{
		"e",
		rapid.Int32().Map(func(x int32) protoreflect.EnumNumber {
			return protoreflect.EnumNumber(x)
		}),
	},
}

func MakeTestCodec(fname protoreflect.Name, nonTerminal bool) (ormfield.Codec, error) {
	field := GetTestField(fname)
	if field == nil {
		return nil, fmt.Errorf("can't find field %s", fname)
	}
	return ormfield.GetCodec(field, nonTerminal)
}

func GetTestField(fname protoreflect.Name) protoreflect.FieldDescriptor {
	a := &testpb.ExampleTable{}
	return a.ProtoReflect().Descriptor().Fields().ByName(fname)
}

type TestKeyCodec struct {
	KeySpecs []TestFieldSpec
	Codec    *ormkv.KeyCodec
}

func TestFieldSpecsGen(minLen, maxLen int) *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) []TestFieldSpec {
		xs := rapid.SliceOfNDistinct(rapid.IntRange(0, len(TestFieldSpecs)-1), minLen, maxLen, func(i int) int { return i }).
			Draw(t, "fieldSpecIndexes").([]int)

		var specs []TestFieldSpec

		for _, x := range xs {
			spec := TestFieldSpecs[x]
			specs = append(specs, spec)
		}

		return specs
	})
}

func TestKeyCodecGen(minLen, maxLen int) *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) TestKeyCodec {
		specs := TestFieldSpecsGen(minLen, maxLen).Draw(t, "fieldSpecs").([]TestFieldSpec)

		var fields []protoreflect.Name
		for _, spec := range specs {
			fields = append(fields, spec.FieldName)
		}

		prefix := rapid.SliceOfN(rapid.Byte(), 0, 5).Draw(t, "prefix").([]byte)

		msgType := (&testpb.ExampleTable{}).ProtoReflect().Type()
		cdc, err := ormkv.NewKeyCodec(prefix, msgType, fields)
		if err != nil {
			panic(err)
		}

		return TestKeyCodec{
			Codec:    cdc,
			KeySpecs: specs,
		}
	})
}

func (k TestKeyCodec) Draw(t *rapid.T, id string) []protoreflect.Value {
	n := len(k.KeySpecs)
	keyValues := make([]protoreflect.Value, n)
	for i, k := range k.KeySpecs {
		keyValues[i] = protoreflect.ValueOf(k.Gen.Draw(t, fmt.Sprintf("%s[%d]", id, i)))
	}
	return keyValues
}

var GenA = rapid.Custom(func(t *rapid.T) *testpb.ExampleTable {
	a := &testpb.ExampleTable{}
	ref := a.ProtoReflect()
	for _, spec := range TestFieldSpecs {
		field := GetTestField(spec.FieldName)
		value := spec.Gen.Draw(t, string(spec.FieldName))
		ref.Set(field, protoreflect.ValueOf(value))
	}
	return a
})
