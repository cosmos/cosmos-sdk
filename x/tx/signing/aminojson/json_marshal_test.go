package aminojson_test

import (
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-proto/rapidproto"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"pgregory.net/rapid"

	"gotest.tools/v3/assert"

	"cosmossdk.io/x/tx/signing/aminojson"

	"cosmossdk.io/x/tx/signing/aminojson/internal/testpb"
)

func marshalLegacy(msg proto.Message) ([]byte, error) {
	cdc := amino.NewCodec()
	cdc.RegisterConcrete(&testpb.ABitOfEverything{}, "ABitOfEverything", nil)
	cdc.RegisterConcrete(&testpb.NestedMessage{}, "NestedMessage", nil)
	return cdc.MarshalJSON(msg)
}

func TestAminoJSON_EdgeCases(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterConcrete(&testpb.ABitOfEverything{}, "ABitOfEverything", nil)
	cdc.RegisterConcrete(&testpb.NestedMessage{}, "NestedMessage", nil)
	aj := aminojson.NewAminoJSON()

	cases := map[string]struct {
		msg       proto.Message
		shouldErr bool
	}{
		"empty":      {msg: &testpb.ABitOfEverything{}},
		"single map": {msg: &testpb.WithAMap{StrMap: map[string]string{"foo": "bar"}}, shouldErr: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			bz, err := aj.Marshal(tc.msg)

			if tc.shouldErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			rv := reflect.New(reflect.TypeOf(tc.msg).Elem()).Elem()
			msg2 := rv.Addr().Interface().(proto.Message)

			legacyBz, err := cdc.MarshalJSON(tc.msg)
			assert.NilError(t, err)

			require.Equal(t, string(legacyBz), string(bz))

			goProtoJSON, err := protojson.Marshal(tc.msg)
			assert.NilError(t, err)
			err = cdc.UnmarshalJSON(bz, msg2)
			assert.NilError(t, err, "unmarshal failed: %s vs %s", legacyBz, goProtoJSON)
		})
	}
}

func TestAminoJSON(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterConcrete(&testpb.ABitOfEverything{}, "ABitOfEverything", nil)
	cdc.RegisterConcrete(&testpb.NestedMessage{}, "NestedMessage", nil)

	msg := &testpb.ABitOfEverything{
		Message: &testpb.NestedMessage{
			Foo: "test",
			Bar: 0, // this is the default value and should be omitted from output
		},
		Enum:     testpb.AnEnum_ONE,
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
	}
	bz, err := aminojson.NewAminoJSON().Marshal(msg)
	assert.NilError(t, err)
	legacyBz, err := cdc.MarshalJSON(msg)
	assert.NilError(t, err)
	require.Equal(t, string(legacyBz), string(bz))
}

func TestRapid(t *testing.T) {
	gen := rapidproto.MessageGenerator(&testpb.ABitOfEverything{}, rapidproto.GeneratorOptions{})
	rapid.Check(t, func(t *rapid.T) {
		msg := gen.Draw(t, "msg")
		bz, err := aminojson.NewAminoJSON().Marshal(msg)
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
	cdc := amino.NewCodec()
	cdc.RegisterConcrete(&testpb.ABitOfEverything{}, "ABitOfEverything", nil)
	cdc.RegisterConcrete(&testpb.NestedMessage{}, "NestedMessage", nil)

	message2 := message.ProtoReflect().New().Interface()
	goProtoJSON, err := cdc.MarshalJSON(message)
	assert.NilError(t, err)
	err = cdc.UnmarshalJSON(marshaledBytes, message2)
	assert.NilError(t, err, "%s vs %s", string(marshaledBytes), string(goProtoJSON))
}
