package aminojson_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"pgregory.net/rapid"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-proto/any"

	"cosmossdk.io/x/tx/aminojson"
	"cosmossdk.io/x/tx/rapidproto"

	"cosmossdk.io/x/tx/aminojson/internal/testpb"
)

func marshalLegacy(msg proto.Message) ([]byte, error) {
	cdc := amino.NewCodec()
	return cdc.MarshalJSON(msg)
}

func TestAminoJSON_EdgeCases(t *testing.T) {
	simpleAny, err := anypb.New(&testpb.NestedMessage{Foo: "any type nested", Bar: 11})
	require.NoError(t, err)

	cdc := amino.NewCodec()
	aj := aminojson.NewAminoJSON()

	cases := map[string]struct {
		msg       proto.Message
		shouldErr bool
	}{
		"empty":         {msg: &testpb.ABitOfEverything{}},
		"single map":    {msg: &testpb.WithAMap{StrMap: map[string]string{"foo": "bar"}}, shouldErr: true},
		"any type":      {msg: &testpb.ABitOfEverything{Any: simpleAny}},
		"zero duration": {msg: &testpb.ABitOfEverything{Duration: durationpb.New(time.Second * 0)}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			bz, err := aj.MarshalAmino(tc.msg)

			if tc.shouldErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			rv := reflect.New(reflect.TypeOf(tc.msg).Elem()).Elem()
			msg2 := rv.Addr().Interface().(proto.Message)

			legacyBz, err := cdc.MarshalJSON(tc.msg)
			assert.NilError(t, err)

			assert.Equal(t, string(legacyBz), string(bz), "legacy: %s vs %s", legacyBz, bz)

			goProtoJSON, err := protojson.Marshal(tc.msg)
			err = cdc.UnmarshalJSON(bz, msg2)
			assert.NilError(t, err, "unmarshal failed: %s vs %s", legacyBz, goProtoJSON)
		})
	}
}

func TestAminoJSON(t *testing.T) {
	a, err := any.New(&testpb.NestedMessage{
		Foo: "any type nested",
		Bar: 11,
	})

	assert.NilError(t, err)
	msg := &testpb.ABitOfEverything{
		Message: &testpb.NestedMessage{
			Foo: "test",
			Bar: 0, // this is the default value and should be omitted from output
		},
		Enum:      testpb.AnEnum_ONE,
		Repeated:  []int32{3, -7, 2, 6, 4},
		Str:       `abcxyz"foo"def`,
		Bool:      true,
		Bytes:     []byte{0, 1, 2, 3},
		I32:       -15,
		F32:       1001,
		U32:       1200,
		Si32:      -376,
		Sf32:      -1000,
		I64:       14578294827584932,
		F64:       9572348124213523654,
		U64:       4759492485,
		Si64:      -59268425823934,
		Sf64:      -659101379604211154,
		Any:       a,
		Timestamp: timestamppb.New(time.Date(2022, 1, 1, 12, 31, 0, 0, time.UTC)),
		Duration:  durationpb.New(time.Second * 3000),
	}
	bz, err := aminojson.NewAminoJSON().MarshalAmino(msg)
	assert.NilError(t, err)
	cdc := amino.NewCodec()
	legacyBz, err := cdc.MarshalJSON(msg)
	assert.NilError(t, err)
	require.Equal(t, string(legacyBz), string(bz), "%s vs legacy: %s", string(bz), string(legacyBz))
}

func TestRapid(t *testing.T) {
	gen := rapidproto.MessageGenerator(&testpb.ABitOfEverything{}, rapidproto.GeneratorOptions{})
	rapid.Check(t, func(t *rapid.T) {
		msg := gen.Draw(t, "msg")
		bz, err := aminojson.NewAminoJSON().MarshalAmino(msg)
		assert.NilError(t, err)
		checkInvariants(t, msg, bz)
	})
}

func checkInvariants(t *rapid.T, message proto.Message, marshaledBytes []byte) {
	checkLegacyParity(t, message, marshaledBytes)
	checkRoundTrip(t, message, marshaledBytes)
}

func checkLegacyParity(t *rapid.T, message proto.Message, marshaledBytes []byte) {
	legacyBz, err := marshalLegacy(message)
	assert.NilError(t, err)
	require.Equal(t, string(legacyBz), string(marshaledBytes), "%s vs legacy: %s", string(marshaledBytes),
		string(legacyBz))
}

func checkRoundTrip(t *rapid.T, message proto.Message, marshaledBytes []byte) {
	message2 := message.ProtoReflect().New().Interface()
	cdc := amino.NewCodec()
	goProtoJSON, err := cdc.MarshalJSON(message)
	assert.NilError(t, err)
	err = cdc.UnmarshalJSON(marshaledBytes, message2)
	assert.NilError(t, err, "%s vs %s", string(marshaledBytes), string(goProtoJSON))
}
