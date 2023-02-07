package aminojson

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	goamino "github.com/tendermint/go-amino"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"pgregory.net/rapid"

	"cosmossdk.io/api/amino"
	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	"cosmossdk.io/x/tx/aminojson"
	"cosmossdk.io/x/tx/rapidproto"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func Test_newGogoMessage(t *testing.T) {
	ma := &authtypes.ModuleAccount{}
	rma := newGogoMessage(reflect.TypeOf(ma).Elem())
	require.NotPanics(t, func() {
		x := rma.(*authtypes.ModuleAccount)
		require.NotNil(t, x.Address)
	})
}

func TestTypeIndex(t *testing.T) {
	ti := newTypeIndex(msgTypes)
	require.Equal(t, len(msgTypes), len(ti.gogoFields))
	require.Equal(t, len(msgTypes), len(ti.pulsarFields))
	for k, v := range ti.pulsarFields {
		require.Equal(t, len(v), len(ti.gogoFields[ti.pulsarToGogo[k]]), "failed on type: %s", k)
	}
}

func TestAminoJSON_AllTypes(t *testing.T) {
	ti := newTypeIndex(msgTypes)
	cdc := goamino.NewCodec()
	aj := aminojson.NewAminoJSON()
	for _, tt := range msgTypes {
		desc := tt.pulsar.ProtoReflect().Descriptor()
		opts := desc.Options()
		if !proto.HasExtension(opts, amino.E_Name) {
			fmt.Printf("WARN: missing name extension for %s", desc.FullName())
			continue
		}
		name := proto.GetExtension(opts, amino.E_Name).(string)
		cdc.RegisterConcrete(tt.gogo, name, nil)
	}

	params := &authapi.Params{}
	genOpts := rapidproto.GeneratorOptions{
		AnyTypeURLs: []string{string(params.ProtoReflect().Descriptor().FullName())},
		Resolver:    protoregistry.GlobalTypes,
	}

	for _, tt := range msgTypes {
		gen := rapidproto.MessageGenerator(tt.pulsar, genOpts)
		fmt.Printf("testing %s\n", tt.pulsar.ProtoReflect().Descriptor().FullName())
		rapid.Check(t, func(t *rapid.T) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic: %+v\n", r)
					t.FailNow()
				}
			}()
			msg := gen.Draw(t, "msg")
			postFixPulsarMessage(msg)
			//goMsg := reflect.New(reflect.TypeOf(tt.gogo).Elem()).Interface().(gogoproto.Message)
			goMsg := newGogoMessage(reflect.TypeOf(tt.gogo).Elem())
			ti.deepClone(msg, goMsg)
			gogobz, err := cdc.MarshalJSON(goMsg)
			require.NoError(t, err, "failed to marshal gogo message")
			pulsarbz, err := aj.MarshalAmino(msg)
			if !bytes.Equal(gogobz, pulsarbz) {
				require.Fail(t, fmt.Sprintf("marshalled messages not equal, are the unmarshalled messages semantically equivalent?\nmarshalled gogo: %s != %s\nunmarshalled gogo: %v vs %v", string(gogobz), string(pulsarbz), goMsg, msg))
			}
			require.Equal(t, string(gogobz), string(pulsarbz))
		})
	}
}
