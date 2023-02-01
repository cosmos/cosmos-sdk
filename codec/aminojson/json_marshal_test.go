package aminojson_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"pgregory.net/rapid"

	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	"cosmossdk.io/api/cosmos/crypto/ed25519"
	distapi "cosmossdk.io/api/cosmos/distribution/v1beta1"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/aminojson"
	"github.com/cosmos/cosmos-sdk/testutil/rapidproto"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/codec/aminojson/internal/testpb"
)

type pubKeyEd25519 struct {
	ed25519.PubKey
}

var _ codec.AminoMarshaler = pubKeyEd25519{}

func (pubKey pubKeyEd25519) MarshalAmino() ([]byte, error) {
	return pubKey.Key, nil
}

func (pubKey pubKeyEd25519) UnmarshalAmino([]byte) error {
	panic("not implemented")
}

func (pubKey pubKeyEd25519) UnmarshalAminoJSON([]byte) error {
	panic("not implemented")
}

// MarshalAminoJSON overrides Amino JSON marshalling.
func (pubKey pubKeyEd25519) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return pubKey.MarshalAmino()
}

func marshalLegacy(msg proto.Message) ([]byte, error) {
	cdc := amino.NewCodec()
	return cdc.MarshalJSON(msg)
}

func TestAminoJSON_LegacyParity(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterConcrete(authtypes.Params{}, "cosmos-sdk/x/auth/Params", nil)
	cdc.RegisterConcrete(disttypes.MsgWithdrawDelegatorReward{}, "cosmos-sdk/x/distribution/MsgWithdrawDelegatorReward", nil)

	cases := map[string]struct {
		gogo   any
		pulsar proto.Message
	}{
		"auth/params": {gogo: &authtypes.Params{TxSigLimit: 10}, pulsar: &authapi.Params{TxSigLimit: 10}},
		// TODO
		// treat
		// (gogoproto.nullable)     = false,
		// (amino.dont_omitempty)   = true
		"distribution/delegator_starting_info": {
			gogo:   &disttypes.DelegatorStartingInfo{},
			pulsar: &distapi.DelegatorStartingInfo{}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gogoBytes, err := cdc.MarshalJSON(tc.gogo)
			require.NoError(t, err)

			pulsarBytes, err := aminojson.MarshalAmino(tc.pulsar)
			require.NoError(t, err)

			require.Equal(t, string(gogoBytes), string(pulsarBytes), "gogo: %s vs pulsar: %s", gogoBytes, pulsarBytes)
		})
	}
}

func TestAminoJSON_EdgeCases(t *testing.T) {
	simpleAny, err := anypb.New(&testpb.NestedMessage{Foo: "any type nested", Bar: 11})
	require.NoError(t, err)

	cdc := amino.NewCodec()
	cdc.RegisterConcrete((*pubKeyEd25519)(nil), "tendermint/PubKeyEd25519", nil)
	pubkey := &pubKeyEd25519{}
	pubkey.Key = []byte("key")

	cases := map[string]struct {
		msg       proto.Message
		shouldErr bool
	}{
		"empty":         {msg: &testpb.ABitOfEverything{}},
		"single map":    {msg: &testpb.WithAMap{StrMap: map[string]string{"foo": "bar"}}, shouldErr: true},
		"any type":      {msg: &testpb.ABitOfEverything{Any: simpleAny}},
		"zero duration": {msg: &testpb.ABitOfEverything{Duration: durationpb.New(time.Second * 0)}},
		"key_field":     {msg: pubkey},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			bz, err := aminojson.MarshalAmino(tc.msg)

			if tc.shouldErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			rv := reflect.New(reflect.TypeOf(tc.msg).Elem()).Elem()
			msg2 := rv.Addr().Interface().(proto.Message)

			legacyBz, err := cdc.MarshalJSON(tc.msg)
			assert.NilError(t, err)

			fmt.Printf("legacy: %s vs %s\n", legacyBz, bz)
			assert.Equal(t, string(legacyBz), string(bz), "legacy: %s vs %s", legacyBz, bz)

			goProtoJson, err := protojson.Marshal(tc.msg)
			err = cdc.UnmarshalJSON(bz, msg2.(proto.Message))
			assert.NilError(t, err, "unmarshal failed: %s vs %s", legacyBz, goProtoJson)
		})
	}
}

func TestAminoJSON(t *testing.T) {
	a, err := anypb.New(&testpb.NestedMessage{
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
	bz, err := aminojson.MarshalAmino(msg)
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
		bz, err := aminojson.MarshalAmino(msg)
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
	goProtoJson, err := cdc.MarshalJSON(message)
	assert.NilError(t, err)
	err = cdc.UnmarshalJSON(marshaledBytes, message2)
	assert.NilError(t, err, "%s vs %s", string(marshaledBytes), string(goProtoJson))
}
