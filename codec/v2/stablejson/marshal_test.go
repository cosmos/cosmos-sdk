package stablejson_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/codec/v2/internal/testpb"
	"cosmossdk.io/codec/v2/stablejson"
)

func TestStableJSON(t *testing.T) {
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
		Repeated:    nil,
		String_:     "",
		Bool:        false,
		Bytes:       nil,
		I32:         0,
		F32:         0,
		U32:         0,
		Si32:        0,
		Sf32:        0,
		I64:         0,
		F64:         0,
		U64:         0,
		Si64:        0,
		Sf64:        0,
		Float:       0,
		Double:      0,
		Any:         nil,
		Timestamp:   nil,
		Duration:    nil,
		Struct:      nil,
		BoolValue:   nil,
		BytesValue:  nil,
		DoubleValue: nil,
		FloatValue:  nil,
		Int32Value:  nil,
		Int64Value:  nil,
		StringValue: nil,
		Uint32Value: nil,
		Uint64Value: nil,
		FieldMask:   nil,
		ListValue:   nil,
		Value:       nil,
		NullValue:   0,
		Empty:       nil,
	}
	bz, err := stablejson.Marshal(msg)
	assert.NilError(t, err)
	assert.Equal(t,
		`{}`,
		string(bz))
}
