package codec_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
)

// NewTestInterfaceRegistry creates a basic InterfaceRegistry for testing
// Animal interfaces, mapping Dog and Cat implementations to the Animal interface.
func NewTestInterfaceRegistry() codectypes.InterfaceRegistry {
	registry := codectypes.NewInterfaceRegistry()
	registry.RegisterInterface("Animal", (*testdata.Animal)(nil))
	registry.RegisterImplementations(
		(*testdata.Animal)(nil),
		&testdata.Dog{},
		&testdata.Cat{},
	)
	return registry
}

// TestMarshalAny verifies the correct serialization and deserialization of
// concrete types wrapped as `Any` interfaces using MarshalInterface/UnmarshalInterface.
func TestMarshalAny(t *testing.T) {
	// Registry containing only the Cat implementation
	catRegistry := codectypes.NewInterfaceRegistry()
	catRegistry.RegisterImplementations((*testdata.Animal)(nil), &testdata.Cat{})

	// Empty registry for comparison tests
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	kitty := &testdata.Cat{Moniker: "Kitty"}
	var animal testdata.Animal

	// 1. Marshal failure: Should fail if the concrete type's interface is not registered in the codec
	emptyBz, err := cdc.MarshalInterface(kitty)
	require.ErrorContains(t, err, "does not have a registered interface")

	// 2. Marshal success: Should succeed with the registry containing the Cat implementation
	catBz, err := codec.NewProtoCodec(catRegistry).MarshalInterface(kitty)
	require.NoError(t, err)
	require.NotEmpty(t, catBz)

	// 3. Unmarshal failure (No implementations): Should error in a codec with an empty registry
	err = cdc.UnmarshalInterface(catBz, &animal)
	require.ErrorContains(t, err, "no registered implementations of type testdata.Animal")

	// 4. Unmarshal empty bytes: Should return nil interface, but no error
	err = cdc.UnmarshalInterface(emptyBz, &animal)
	require.Nil(t, animal)
	require.NoError(t, err)

	// 5. Unmarshal failure (Wrong implementation): Should fail if a different implementation is registered
	registry.RegisterImplementations((*testdata.Animal)(nil), &testdata.Dog{})
	err = cdc.UnmarshalInterface(catBz, &animal)
	require.Error(t, err)

	// 6. Unmarshal success: Should pass with the full registry containing Cat
	registry = NewTestInterfaceRegistry()
	cdc = codec.NewProtoCodec(registry)
	err = cdc.UnmarshalInterface(catBz, &animal)
	require.NoError(t, err)
	require.Equal(t, kitty, animal)

	// 7. Unmarshal failure (Nil target): Should fail if the target pointer is nil
	_ = NewTestInterfaceRegistry()
	err = cdc.UnmarshalInterface(catBz, nil)
	require.Error(t, err)
}

// TestMarshalProtoPubKey tests JSON and binary marshaling of a PubKey wrapped in a raw `Any` type.
func TestMarshalProtoPubKey(t *testing.T) {
	require := require.New(t)
	// Use test configuration to ensure necessary interfaces (e.g., PubKey) are registered
	ccfg := testutil.MakeTestEncodingConfig()
	privKey := ed25519.GenPrivKey()
	pk := privKey.PubKey()

	// --- Test JSON serialization (Raw Any) ---

	pkAny, err := codectypes.NewAnyWithValue(pk)
	require.NoError(err)
	bz, err := ccfg.Codec.MarshalJSON(pkAny) // Marshals the raw Any struct
	require.NoError(err)

	var pkAny2 codectypes.Any
	err = ccfg.Codec.UnmarshalJSON(bz, &pkAny2)
	require.NoError(err)
	
	// Unpacking is required for raw Any type to populate the CachedValue
	var pkI cryptotypes.PubKey
	err = ccfg.InterfaceRegistry.UnpackAny(&pkAny2, &pkI)
	require.NoError(err)
	
	// Check the unpacked value
	pk2 := pkAny2.GetCachedValue().(cryptotypes.PubKey)
	require.True(pk2.Equals(pk))

	// --- Test Binary serialization (Raw Any) ---

	bz, err = ccfg.Codec.Marshal(pkAny)
	require.NoError(err)

	var pkAny3 codectypes.Any
	err = ccfg.Codec.Unmarshal(bz, &pkAny3)
	require.NoError(err)
	
	// Unpack is still necessary for binary
	err = ccfg.InterfaceRegistry.UnpackAny(&pkAny3, &pkI)
	require.NoError(err)
	pk3 := pkAny3.GetCachedValue().(cryptotypes.PubKey)
	require.True(pk3.Equals(pk))
}

// TestMarshalProtoInterfacePubKey tests PubKey marshaling using the explicit
// (Un)marshalInterface helper functions, which simplify the Any handling.
func TestMarshalProtoInterfacePubKey(t *testing.T) {
	require := require.New(t)
	ccfg := testutil.MakeTestEncodingConfig()
	privKey := ed25519.GenPrivKey()
	pk := privKey.PubKey()

	// --- Test JSON serialization (MarshalInterfaceJSON) ---

	// MarshalInterfaceJSON automatically wraps the interface in Any
	bz, err := ccfg.Codec.MarshalInterfaceJSON(pk)
	require.NoError(err)

	var pk3 cryptotypes.PubKey
	// UnmarshalInterfaceJSON automatically unpacks the Any into the target interface
	err = ccfg.Codec.UnmarshalInterfaceJSON(bz, &pk3)
	require.NoError(err)
	require.True(pk3.Equals(pk))

	// --- Check unmarshal using standard JSONCodec (Demonstrates why helper is needed) ---
	// Unmarshaling into a raw Any type using standard JSONCodec will not auto-unpack
	// since Any doesn't implement UnpackInterfacesMessage.
	var pkAny codectypes.Any
	err = ccfg.Codec.UnmarshalJSON(bz, &pkAny)
	require.NoError(err)
	ifc := pkAny.GetCachedValue()
	require.Nil(ifc) // Cached value is nil until UnpackAny is called manually

	// --- Test Binary serialization (MarshalInterface) ---

	// MarshalInterface automatically wraps the interface in Any
	bz, err = ccfg.Codec.MarshalInterface(pk)
	require.NoError(err)

	var pk2 cryptotypes.PubKey
	// UnmarshalInterface automatically unpacks the Any into the target interface
	err = ccfg.Codec.UnmarshalInterface(bz, &pk2)
	require.NoError(err)
	require.True(pk2.Equals(pk))
}
