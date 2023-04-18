package testutil

import (
	"fmt"
	"math"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pgregory.net/rapid"

	"cosmossdk.io/orm/encoding/ormfield"
	"cosmossdk.io/orm/encoding/ormkv"
	"cosmossdk.io/orm/internal/testpb"
)

// TestFieldSpec defines a test field against the testpb.ExampleTable message.
type TestFieldSpec struct {
	FieldName protoreflect.Name
	Gen       *rapid.Generator[any]
}

var TestFieldSpecs = []TestFieldSpec{
	{
		"u32",
		rapid.Uint32().AsAny(),
	},
	{
		"u64",
		rapid.Uint64().AsAny(),
	},
	{
		"str",
		rapid.String().Filter(func(x string) bool {
			// filter out null terminators
			return strings.IndexByte(x, 0) < 0
		}).AsAny(),
	},
	{
		"bz",
		rapid.SliceOfN(rapid.Byte(), 0, math.MaxUint32).AsAny(),
	},
	{
		"i32",
		rapid.Int32().AsAny(),
	},
	{
		"f32",
		rapid.Uint32().AsAny(),
	},
	{
		"s32",
		rapid.Int32().AsAny(),
	},
	{
		"sf32",
		rapid.Int32().AsAny(),
	},
	{
		"i64",
		rapid.Int64().AsAny(),
	},
	{
		"f64",
		rapid.Uint64().AsAny(),
	},
	{
		"s64",
		rapid.Int64().AsAny(),
	},
	{
		"sf64",
		rapid.Int64().AsAny(),
	},
	{
		"b",
		rapid.Bool().AsAny(),
	},
	{
		"ts",
		rapid.Custom(func(t *rapid.T) protoreflect.Message {
			isNil := rapid.Float32().Draw(t, "isNil")
			if isNil >= 0.95 { // draw a nil 5% of the time
				return nil
			}
			seconds := rapid.Int64Range(ormfield.TimestampSecondsMin, ormfield.TimestampSecondsMax).Draw(t, "seconds")
			nanos := rapid.Int32Range(0, ormfield.TimestampNanosMax).Draw(t, "nanos")
			return (&timestamppb.Timestamp{
				Seconds: seconds,
				Nanos:   nanos,
			}).ProtoReflect()
		}).AsAny(),
	},
	{
		"dur",
		rapid.Custom(func(t *rapid.T) protoreflect.Message {
			isNil := rapid.Float32().Draw(t, "isNil")
			if isNil >= 0.95 { // draw a nil 5% of the time
				return nil
			}
			seconds := rapid.Int64Range(ormfield.DurationNanosMin, ormfield.DurationNanosMax).Draw(t, "seconds")
			nanos := rapid.Int32Range(0, ormfield.DurationNanosMax).Draw(t, "nanos")
			if seconds < 0 {
				nanos = -nanos
			}
			return (&durationpb.Duration{
				Seconds: seconds,
				Nanos:   nanos,
			}).ProtoReflect()
		}).AsAny(),
	},
	{
		"e",
		rapid.Map(rapid.Int32(), func(x int32) protoreflect.EnumNumber {
			return protoreflect.EnumNumber(x)
		}).AsAny(),
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

func TestFieldSpecsGen(minLen, maxLen int) *rapid.Generator[[]TestFieldSpec] {
	return rapid.Custom(func(t *rapid.T) []TestFieldSpec {
		xs := rapid.SliceOfNDistinct(rapid.IntRange(0, len(TestFieldSpecs)-1), minLen, maxLen, func(i int) int { return i }).
			Draw(t, "fieldSpecIndexes")

		var specs []TestFieldSpec

		for _, x := range xs {
			spec := TestFieldSpecs[x]
			specs = append(specs, spec)
		}

		return specs
	})
}

func TestKeyCodecGen(minLen, maxLen int) *rapid.Generator[TestKeyCodec] {
	return rapid.Custom(func(t *rapid.T) TestKeyCodec {
		specs := TestFieldSpecsGen(minLen, maxLen).Draw(t, "fieldSpecs")

		var fields []protoreflect.Name
		for _, spec := range specs {
			fields = append(fields, spec.FieldName)
		}

		prefix := rapid.SliceOfN(rapid.Byte(), 0, 5).Draw(t, "prefix")

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
		if value != nil {
			ref.Set(field, protoreflect.ValueOf(value))
		}
	}
	return a
})
