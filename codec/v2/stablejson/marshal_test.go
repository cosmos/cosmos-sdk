package stablejson_test

import (
	"testing"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"gotest.tools/v3/assert"

	"cosmossdk.io/codec/v2/internal/testpb"
	"cosmossdk.io/codec/v2/stablejson"
)

func TestStableJSON(t *testing.T) {
	a, err := anypb.New(&testpb.ABitOfEverything{
		I32: 10,
		Str: "abc",
	})
	assert.NilError(t, err)
	msg := &testpb.ABitOfEverything{
		//Message: &testpb.NestedMessage{
		//	Foo: "test",
		//	Bar: 0, // this is the default value and should be omitted from output
		//},
		//Enum: testpb.AnEnum_ONE,
		//StrMap: map[string]string{
		//	"foo": "abc",
		//	"bar": "def",
		//},
		//Int32Map: map[int32]string{
		//	-3: "xyz",
		//	0:  "abc",
		//	10: "qrs",
		//},
		//BoolMap: map[bool]string{
		//	true:  "T",
		//	false: "F",
		//},
		//Repeated:  []int32{3, -7, 2, 6, 4},
		//Str:       `abcxyz"foo"def`,
		//Bool:      true,
		//Bytes:     []byte{0, 1, 2, 3},
		//I32:       -15,
		//F32:       1001,
		//U32:       1200,
		//Si32:      -376,
		//Sf32:      -1000,
		//I64:       14578294827584932,
		//F64:       9572348124213523654,
		//U64:       4759492485,
		//Si64:      -59268425823934,
		//Sf64:      -659101379604211154,
		//Float:     1.0,
		//Double:    5235.2941,
		Any: a,
		//Timestamp: timestamppb.New(time.Date(2022, 1, 1, 12, 31, 0, 0, time.UTC)),
		//Duration:  durationpb.New(time.Second * 3000),
		Struct: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"null": structpb.NewNullValue(),
				"num":  structpb.NewNumberValue(3.76),
				"str":  structpb.NewStringValue("abc"),
				"bool": structpb.NewBoolValue(true),
				"struct": structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
					"a": structpb.NewStringValue("abc"),
				}}),
				"list": structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{
					structpb.NewStringValue("xyz"),
					structpb.NewBoolValue(false),
					structpb.NewNumberValue(-9),
				}}),
			},
		},
		//BoolValue:   &wrapperspb.BoolValue{Value: true},
		//BytesValue:  &wrapperspb.BytesValue{Value: []byte{0, 1, 2, 3}},
		//DoubleValue: &wrapperspb.DoubleValue{Value: 1.324},
		//FloatValue:  &wrapperspb.FloatValue{Value: -1.0},
		//Int32Value:  &wrapperspb.Int32Value{Value: 10},
		//Int64Value:  &wrapperspb.Int64Value{Value: -376923457},
		//StringValue: &wrapperspb.StringValue{Value: "gfedcba"},
		//Uint32Value: &wrapperspb.UInt32Value{Value: 37492},
		//Uint64Value: &wrapperspb.UInt64Value{Value: 1892409137358391},
		//FieldMask:   &fieldmaskpb.FieldMask{Paths: []string{"f", "a", "b"}},
		//ListValue:   &structpb.ListValue{Values: nil},
		//Value:       &structpb.Value{},
		//NullValue:   structpb.NullValue_NULL_VALUE,
		//Empty:       &emptypb.Empty{},
	}
	bz, err := stablejson.Marshal(msg)
	assert.NilError(t, err)
	assert.Equal(t,
		`{}`,
		string(bz))
}
