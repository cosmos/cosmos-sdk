package aminojson_test

import (
	"github.com/tendermint/go-amino"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"

	"github.com/cosmos/cosmos-sdk/codec/aminojson"
	"github.com/cosmos/cosmos-sdk/codec/aminojson/internal/testpb"
)

func TestAminoJSON(t *testing.T) {
	a, err := anypb.New(&testpb.ABitOfEverything{
		I32: 10,
		Str: "abc",
	})
	assert.NilError(t, err)
	msg := &testpb.ABitOfEverything{
		Message: &testpb.NestedMessage{
			Foo: "test",
			Bar: 0, // this is the default value and should be omitted from output
		},
		Enum: testpb.AnEnum_ONE,
		StrMap: map[string]string{
			"foo": "abc",
			"bar": "def",
		},
		Int32Map: map[int32]string{
			-3: "xyz",
			0:  "abc",
			10: "qrs",
		},
		BoolMap: map[bool]string{
			true:  "T",
			false: "F",
		},
		Repeated: []int32{3, -7, 2, 6, 4},
		Str:      `abcxyz"foo"def`,
		Bool:     true,
		Bytes:    []byte{0, 1, 2, 3},
		I32:      -15,
		F32:      1001,
		U32:      1200,
		Si32:     -376,
		Sf32:     -1000,
		I64:      14578294827584932,
		F64:      9572348124213523654,
		U64:      4759492485,
		Si64:     -59268425823934,
		Sf64:     -659101379604211154,
		//Float:     1.0,
		//Double:    5235.2941,
		Any:       a,
		Timestamp: timestamppb.New(time.Date(2022, 1, 1, 12, 31, 0, 0, time.UTC)),
		Duration:  durationpb.New(time.Second * 3000),
		Struct: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"null": structpb.NewNullValue(),
				"num":  structpb.NewNumberValue(3.76),
				"str":  structpb.NewStringValue("abc"),
				"bool": structpb.NewBoolValue(true),
				"nested struct": structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
					"a": structpb.NewStringValue("abc"),
				}}),
				"struct list": structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{
					structpb.NewStringValue("xyz"),
					structpb.NewBoolValue(false),
					structpb.NewNumberValue(-9),
				}}),
				"empty": {},
			},
		},
		BoolValue:  &wrapperspb.BoolValue{Value: true},
		BytesValue: &wrapperspb.BytesValue{Value: []byte{0, 1, 2, 3}},
		//DoubleValue: &wrapperspb.DoubleValue{Value: 1.324},
		//FloatValue:  &wrapperspb.FloatValue{Value: -1.0},
		Int32Value:  &wrapperspb.Int32Value{Value: 10},
		Int64Value:  &wrapperspb.Int64Value{Value: -376923457},
		StringValue: &wrapperspb.StringValue{Value: "gfedcba"},
		Uint32Value: &wrapperspb.UInt32Value{Value: 37492},
		Uint64Value: &wrapperspb.UInt64Value{Value: 1892409137358391},
		FieldMask:   &fieldmaskpb.FieldMask{Paths: []string{"a.b", "a.c", "b"}},
		ListValue: &structpb.ListValue{Values: []*structpb.Value{
			structpb.NewNumberValue(1.1),
			structpb.NewStringValue("qrs"),
		}},
		Value:     structpb.NewStringValue("a value"),
		NullValue: structpb.NullValue_NULL_VALUE,
		Empty:     &emptypb.Empty{},
	}
	bz, err := aminojson.MarshalAmino(msg)
	assert.NilError(t, err)
	cdc := amino.NewCodec()
	legacyBz, err := cdc.MarshalJSON(msg)
	golden.Assert(t, string(legacyBz), "example1.json")
	golden.Assert(t, string(bz), "example1.json")
}

/*
func TestRapid(t *testing.T) {
	gen := rapidproto.MessageGenerator(&testpb.ABitOfEverything{}, rapidproto.GeneratorOptions{})
	rapid.Check(t, func(t *rapid.T) {
		msg := gen.Draw(t, "msg")
		bz, err := aminojson.Marshal(msg)
		assert.NilError(t, err)
		checkInvariants(t, msg, bz)
	})
}

func checkInvariants(t *rapid.T, message proto.Message, marshaledBytes []byte) {
	checkRoundTrip(t, message, marshaledBytes)
	checkJsonNoWhitespace(t, marshaledBytes)
}

func checkRoundTrip(t *rapid.T, message proto.Message, marshaledBytes []byte) {
	message2 := message.ProtoReflect().New().Interface()
	goProtoJson, err := protojson.Marshal(message)
	assert.NilError(t, err)
	assert.NilError(t, protojson.UnmarshalOptions{}.Unmarshal(marshaledBytes, message2), "%s vs %s", string(marshaledBytes), string(goProtoJson))
	// TODO: assert.DeepEqual(t, message, message2, protocmp.Transform())
}

func checkJsonInvariants(t *testing.T, message proto.Message, unmarshaledJson map[string]interface{}) {
}

func checkJsonNoWhitespace(t *rapid.T, marshaledBytes []byte) {
}

func checkJsonFieldsOrdered(t *testing.T, message proto.Message, unmarshaledJson map[string]interface{}) {
}

func checkJsonMapsOrdered(t *testing.T, message proto.Message, unmarshaledJson map[string]interface{}) {
}

func checkJsonStringMapsOrdered(t *testing.T, message proto.Message, unmarshaledJson map[string]interface{}) {
}

func checkJsonNumericMapsOrdered(t *testing.T, message proto.Message, unmarshaledJson map[string]interface{}) {
}

func checkJsonBoolMapsOrdered(t *testing.T, message proto.Message, unmarshaledJson map[string]interface{}) {
}
*/
