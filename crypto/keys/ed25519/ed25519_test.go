package ed25519_test

import (
	stded25519 "crypto/ed25519"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	tmed25519 "github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	ed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
)

func TestSignAndValidateEd25519(t *testing.T) {
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(1000)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifySignature(msg, sig))

	// ----
	// Test cross packages verification
	stdPrivKey := stded25519.PrivateKey(privKey.Key)
	stdPubKey := stdPrivKey.Public().(stded25519.PublicKey)

	assert.Equal(t, stdPubKey, pubKey.(*ed25519.PubKey).Key)
	assert.Equal(t, stdPrivKey, privKey.Key)
	assert.True(t, stded25519.Verify(stdPubKey, msg, sig))
	sig2 := stded25519.Sign(stdPrivKey, msg)
	assert.True(t, pubKey.VerifySignature(msg, sig2))

	// ----
	// Mutate the signature, just one bit.
	// TODO: Replace this with a much better fuzzer, tendermint/ed25519/issues/10
	sig[7] ^= byte(0x01)
	assert.False(t, pubKey.VerifySignature(msg, sig))
}

func TestPubKeyEquals(t *testing.T) {
	ed25519PubKey := ed25519.GenPrivKey().PubKey().(*ed25519.PubKey)

	testCases := []struct {
		msg      string
		pubKey   cryptotypes.PubKey
		other    cryptotypes.PubKey
		expectEq bool
	}{
		{
			"different bytes",
			ed25519PubKey,
			ed25519.GenPrivKey().PubKey(),
			false,
		},
		{
			"equals",
			ed25519PubKey,
			&ed25519.PubKey{
				Key: ed25519PubKey.Key,
			},
			true,
		},
		{
			"different types",
			ed25519PubKey,
			secp256k1.GenPrivKey().PubKey(),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			eq := tc.pubKey.Equals(tc.other)
			require.Equal(t, eq, tc.expectEq)
		})
	}
}

func TestPrivKeyEquals(t *testing.T) {
	ed25519PrivKey := ed25519.GenPrivKey()

	testCases := []struct {
		msg      string
		privKey  cryptotypes.PrivKey
		other    cryptotypes.PrivKey
		expectEq bool
	}{
		{
			"different bytes",
			ed25519PrivKey,
			ed25519.GenPrivKey(),
			false,
		},
		{
			"equals",
			ed25519PrivKey,
			&ed25519.PrivKey{
				Key: ed25519PrivKey.Key,
			},
			true,
		},
		{
			"different types",
			ed25519PrivKey,
			secp256k1.GenPrivKey(),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			eq := tc.privKey.Equals(tc.other)
			require.Equal(t, eq, tc.expectEq)
		})
	}
}

func TestMarshalAmino(t *testing.T) {
	aminoCdc := codec.NewLegacyAmino()
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey().(*ed25519.PubKey)

	testCases := []struct {
		desc      string
		msg       codec.AminoMarshaler
		typ       interface{}
		expBinary []byte
		expJSON   string
	}{
		{
			"ed25519 private key",
			privKey,
			&ed25519.PrivKey{},
			append([]byte{64}, privKey.Bytes()...), // Length-prefixed.
			"\"" + base64.StdEncoding.EncodeToString(privKey.Bytes()) + "\"",
		},
		{
			"ed25519 public key",
			pubKey,
			&ed25519.PubKey{},
			append([]byte{32}, pubKey.Bytes()...), // Length-prefixed.
			"\"" + base64.StdEncoding.EncodeToString(pubKey.Bytes()) + "\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Do a round trip of encoding/decoding binary.
			bz, err := aminoCdc.MarshalBinaryBare(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.expBinary, bz)

			err = aminoCdc.UnmarshalBinaryBare(bz, tc.typ)
			require.NoError(t, err)

			require.Equal(t, tc.msg, tc.typ)

			// Do a round trip of encoding/decoding JSON.
			bz, err = aminoCdc.MarshalJSON(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.expJSON, string(bz))

			err = aminoCdc.UnmarshalJSON(bz, tc.typ)
			require.NoError(t, err)

			require.Equal(t, tc.msg, tc.typ)
		})
	}
}

func TestMarshalAmino_BackwardsCompatibility(t *testing.T) {
	aminoCdc := codec.NewLegacyAmino()
	// Create Tendermint keys.
	tmPrivKey := tmed25519.GenPrivKey()
	tmPubKey := tmPrivKey.PubKey()
	// Create our own keys, with the same private key as Tendermint's.
	privKey := &ed25519.PrivKey{Key: []byte(tmPrivKey)}
	pubKey := privKey.PubKey().(*ed25519.PubKey)

	testCases := []struct {
		desc      string
		tmKey     interface{}
		ourKey    interface{}
		marshalFn func(o interface{}) ([]byte, error)
	}{
		{
			"ed25519 private key, binary",
			tmPrivKey,
			privKey,
			aminoCdc.MarshalBinaryBare,
		},
		{
			"ed25519 private key, JSON",
			tmPrivKey,
			privKey,
			aminoCdc.MarshalJSON,
		},
		{
			"ed25519 public key, binary",
			tmPubKey,
			pubKey,
			aminoCdc.MarshalBinaryBare,
		},
		{
			"ed25519 public key, JSON",
			tmPubKey,
			pubKey,
			aminoCdc.MarshalJSON,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Make sure Amino encoding override is not breaking backwards compatibility.
			bz1, err := tc.marshalFn(tc.tmKey)
			require.NoError(t, err)
			bz2, err := tc.marshalFn(tc.ourKey)
			require.NoError(t, err)
			require.Equal(t, bz1, bz2)
		})
	}
}

// TODO - maybe we should move the tests below to `codec_test` package, WDYT?
func TestMarshalProto(t *testing.T) {
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
	// we before getting a cached value we need to unpack it.
	// Normally this happens in in types which implement UnpackInterfaces
	var pkI cryptotypes.PubKey
	err = ccfg.InterfaceRegistry.UnpackAny(&pkAny2, &pkI)
	require.NoError(err)
	var pk2 = pkAny2.GetCachedValue().(cryptotypes.PubKey)
	require.True(pk2.Equals(pk))

	// **** test binary serialization ****

	bz, err = ccfg.Marshaler.MarshalBinaryBare(pkAny)
	fmt.Println(bz)
	require.NoError(err)

	var pkAny3 codectypes.Any
	err = ccfg.Marshaler.UnmarshalBinaryBare(bz, &pkAny3)
	require.NoError(err)
	err = ccfg.InterfaceRegistry.UnpackAny(&pkAny3, &pkI)
	require.NoError(err)
	var pk3 = pkAny3.GetCachedValue().(cryptotypes.PubKey)
	require.True(pk3.Equals(pk))
}

func TestMarshalProto2(t *testing.T) {
	require := require.New(t)
	ccfg := simapp.MakeTestEncodingConfig()
	privKey := ed25519.GenPrivKey()
	pk := privKey.PubKey()

	// **** test JSON serialization ****

	bz, err := ccfg.Marshaler.MarshalInterfaceJSON(pk)
	require.NoError(err)

	var pk3 cryptotypes.PubKey
	am := codec.NewJSONAnyMarshaler(ccfg.Marshaler, ccfg.InterfaceRegistry)
	err = codec.UnmarshalIfcJSON(am, &pk3, bz)
	require.NoError(err)
	require.True(pk3.Equals(pk))

	// ** Check unmarshal using JSONMarshaler **
	// Unpacking won't work straightforward s Any type
	// Any can't implement UnpackInterfacesMessage insterface. So Any is not
	// automatically unpacked and we won't get a value.
	var pkAny codectypes.Any
	err = ccfg.Marshaler.UnmarshalJSON(bz, &pkAny)
	require.NoError(err)
	ifc := pkAny.GetCachedValue()
	require.Nil(ifc)

	// **** test binary serialization ****

	bz, err = codec.MarshalIfc(ccfg.Marshaler, pk)
	require.NoError(err)

	var pk2 cryptotypes.PubKey
	err = codec.UnmarshalIfc(ccfg.Marshaler, &pk2, bz)
	require.NoError(err)
	require.True(pk2.Equals(pk))
}
