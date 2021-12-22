package codec_test

import (
	"bytes"
	"fmt"
	gogopb "github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	testdatav2 "github.com/cosmos/cosmos-sdk/testutil/testdata/v2"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
)

func NewTestInterfaceRegistry() types.InterfaceRegistry {
	registry := types.NewInterfaceRegistry()
	registry.RegisterInterface("Animal", (*testdata.Animal)(nil))
	registry.RegisterImplementations(
		(*testdata.Animal)(nil),
		&testdata.Dog{},
		&testdata.Cat{},
		&testdatav2.Snake{},
	)
	return registry
}

func TestRegistry(t *testing.T) {
	NewTestInterfaceRegistry()
}

func TestMarshalAny(t *testing.T) {
	registry := types.NewInterfaceRegistry()

	cdc := codec.NewProtoCodec(registry)

	kitty := &testdata.Cat{Moniker: "Kitty"}
	bz, err := cdc.MarshalInterface(kitty)
	require.NoError(t, err)

	var animal testdata.Animal

	// empty registry should fail
	err = cdc.UnmarshalInterface(bz, &animal)
	require.Error(t, err)

	// wrong type registration should fail
	registry.RegisterImplementations((*testdata.Animal)(nil), &testdata.Dog{})
	err = cdc.UnmarshalInterface(bz, &animal)
	require.Error(t, err)

	// should pass
	registry = NewTestInterfaceRegistry()
	cdc = codec.NewProtoCodec(registry)
	err = cdc.UnmarshalInterface(bz, &animal)
	require.NoError(t, err)
	require.Equal(t, kitty, animal)

	// nil should fail
	registry = NewTestInterfaceRegistry()
	err = cdc.UnmarshalInterface(bz, nil)
	require.Error(t, err)
}

func TestMarshalProtoPubKey(t *testing.T) {
	require := require.New(t)
	ccfg := simapp.MakeTestEncodingConfig()
	privKey := ed25519.GenPrivKey()
	pk := privKey.PubKey()

	// **** test JSON serialization ****
	pkAny, err := codectypes.NewAnyWithValue(pk)
	require.NoError(err)
	bz, err := ccfg.Codec.MarshalJSON(pkAny)
	require.NoError(err)

	var pkAny2 codectypes.Any
	err = ccfg.Codec.UnmarshalJSON(bz, &pkAny2)
	require.NoError(err)
	// Before getting a cached value we need to unpack it.
	// Normally this happens in types which implement UnpackInterfaces
	var pkI cryptotypes.PubKey
	err = ccfg.InterfaceRegistry.UnpackAny(&pkAny2, &pkI)
	require.NoError(err)
	var pk2 = pkAny2.GetCachedValue().(cryptotypes.PubKey)
	require.True(pk2.Equals(pk))

	// **** test binary serialization ****

	bz, err = ccfg.Codec.Marshal(pkAny)
	require.NoError(err)

	var pkAny3 codectypes.Any
	err = ccfg.Codec.Unmarshal(bz, &pkAny3)
	require.NoError(err)
	err = ccfg.InterfaceRegistry.UnpackAny(&pkAny3, &pkI)
	require.NoError(err)
	var pk3 = pkAny3.GetCachedValue().(cryptotypes.PubKey)
	require.True(pk3.Equals(pk))
}

// TestMarshalProtoInterfacePubKey tests PubKey marshaling using (Un)marshalInterface
// helper functions
func TestMarshalProtoInterfacePubKey(t *testing.T) {
	require := require.New(t)
	ccfg := simapp.MakeTestEncodingConfig()
	privKey := ed25519.GenPrivKey()
	pk := privKey.PubKey()

	// **** test JSON serialization ****

	bz, err := ccfg.Codec.MarshalInterfaceJSON(pk)
	require.NoError(err)

	var pk3 cryptotypes.PubKey
	err = ccfg.Codec.UnmarshalInterfaceJSON(bz, &pk3)
	require.NoError(err)
	require.True(pk3.Equals(pk))

	// ** Check unmarshal using JSONCodec **
	// Unpacking won't work straightforward s Any type
	// Any can't implement UnpackInterfacesMessage interface. So Any is not
	// automatically unpacked and we won't get a value.
	var pkAny codectypes.Any
	err = ccfg.Codec.UnmarshalJSON(bz, &pkAny)
	require.NoError(err)
	ifc := pkAny.GetCachedValue()
	require.Nil(ifc)

	// **** test binary serialization ****

	bz, err = ccfg.Codec.MarshalInterface(pk)
	require.NoError(err)

	var pk2 cryptotypes.PubKey
	err = ccfg.Codec.UnmarshalInterface(bz, &pk2)
	require.NoError(err)
	require.True(pk2.Equals(pk))
}

func TestMarshalAnyV2(t *testing.T) {
	registry := NewTestInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	snakey := testdatav2.Snake{Name: "Snakey", Age: 42}
	bz, err := cdc.MarshalInterface(&snakey)
	require.NoError(t, err)

	var animal testdata.Animal

	// should pass
	err = cdc.UnmarshalInterface(bz, &animal)
	require.NoError(t, err)
	assert.DeepEqual(t, &snakey, animal, protocmp.Transform())
	require.Equal(t, snakey.Greet(), animal.Greet())

	// get Any for snake
	any, err := codectypes.NewAnyWithValue(&snakey)
	require.NoError(t, err)

	// wrap Any into HasAnimal
	hasAnimal := testdata.HasAnimal{
		Animal: any,
		X:      15,
	}

	// marshal this message
	bz, err = cdc.MarshalJSON(&hasAnimal)
	require.NoError(t, err)

	// unmarshal this message
	var hasAnimal2 testdata.HasAnimal
	err = cdc.UnmarshalJSON(bz, &hasAnimal2)
	require.NoError(t, err)

	// nil should fail
	registry = NewTestInterfaceRegistry()
	err = cdc.UnmarshalInterface(bz, nil)
	require.Error(t, err)
}

type FakeRegistry struct {
	typeMap map[string]proto.Message
}

var _ jsonpb.AnyResolver = &FakeRegistry{}

func (f *FakeRegistry) RegisterInterface(msg proto.Message) {
	fmt.Println("regstering: ", proto.MessageName(msg))
	f.typeMap["/"+proto.MessageName(msg)] = msg
}

func (f *FakeRegistry) Resolve(typeURL string) (proto.Message, error) {

	if msg, ok := f.typeMap[typeURL]; ok {
		return msg, nil
	}
	return nil, fmt.Errorf("no msg found for type url %s\n", typeURL)
}

func TestMarshalInterfacev2(t *testing.T) {
	require := require.New(t)
	privKey := ed25519.GenPrivKey()
	ccfg := simapp.MakeTestEncodingConfig()
	registry := &FakeRegistry{typeMap: make(map[string]proto.Message)}
	pk := privKey.PubKey()
	registry.RegisterInterface(pk)

	// we mimick the internals of MarshalInterfaceJSON here so we can pass in our fake test registry!
	// with the resolver, this works fine.
	any, err := types.NewAnyWithValue(pk)
	require.NoError(err)
	var protoJM = &jsonpb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: registry}
	buf := new(bytes.Buffer)
	err = protoJM.Marshal(buf, any)
	require.NoError(err)

	// now lets set the resolver to nil, which will give us an error when marshaling
	protoJM.AnyResolver = nil
	buf = new(bytes.Buffer)
	err = protoJM.Marshal(buf, any)
	require.Error(err) // error! cant find this type!
	require.Contains(err.Error(), "not found")

	// now lets use gogo's jsonpb marshaler without an AnyResolver, we see it still works!
	var gogoJM = &gogopb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: nil}
	buf = new(bytes.Buffer)
	err = gogoJM.Marshal(buf, any)
	require.NoError(err)

	// works with or without a resolver
	gogoJM.AnyResolver = ccfg.InterfaceRegistry
	buf = new(bytes.Buffer)
	err = gogoJM.Marshal(buf, any)
	require.NoError(err)
}
