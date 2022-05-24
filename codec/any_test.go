package codec_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func NewTestInterfaceRegistry() types.InterfaceRegistry {
	registry := types.NewInterfaceRegistry()
	registry.RegisterInterface("Animal", (*testdata.Animal)(nil))
	registry.RegisterImplementations(
		(*testdata.Animal)(nil),
		&testdata.Dog{},
		&testdata.Cat{},
	)
	return registry
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
	bz, err := ccfg.Marshaler.MarshalJSON(pkAny)
	require.NoError(err)

	var pkAny2 codectypes.Any
	err = ccfg.Marshaler.UnmarshalJSON(bz, &pkAny2)
	require.NoError(err)
	// Before getting a cached value we need to unpack it.
	// Normally this happens in types which implement UnpackInterfaces
	var pkI cryptotypes.PubKey
	err = ccfg.InterfaceRegistry.UnpackAny(&pkAny2, &pkI)
	require.NoError(err)
	var pk2 = pkAny2.GetCachedValue().(cryptotypes.PubKey)
	require.True(pk2.Equals(pk))

	// **** test binary serialization ****

	bz, err = ccfg.Marshaler.Marshal(pkAny)
	require.NoError(err)

	var pkAny3 codectypes.Any
	err = ccfg.Marshaler.Unmarshal(bz, &pkAny3)
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

	bz, err := ccfg.Marshaler.MarshalInterfaceJSON(pk)
	require.NoError(err)

	var pk3 cryptotypes.PubKey
	err = ccfg.Marshaler.UnmarshalInterfaceJSON(bz, &pk3)
	require.NoError(err)
	require.True(pk3.Equals(pk))

	// ** Check unmarshal using JSONCodec **
	// Unpacking won't work straightforward s Any type
	// Any can't implement UnpackInterfacesMessage interface. So Any is not
	// automatically unpacked and we won't get a value.
	var pkAny codectypes.Any
	err = ccfg.Marshaler.UnmarshalJSON(bz, &pkAny)
	require.NoError(err)
	ifc := pkAny.GetCachedValue()
	require.Nil(ifc)

	// **** test binary serialization ****

	bz, err = ccfg.Marshaler.MarshalInterface(pk)
	require.NoError(err)

	var pk2 cryptotypes.PubKey
	err = ccfg.Marshaler.UnmarshalInterface(bz, &pk2)
	require.NoError(err)
	require.True(pk2.Equals(pk))
}
