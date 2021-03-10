package sr25519_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	tmsr25519 "github.com/tendermint/tendermint/crypto/sr25519"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sr25519 "github.com/cosmos/cosmos-sdk/crypto/keys/sr25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

func TestSignAndValidateSr25519(t *testing.T) {
	privKey := sr25519.GenPrivKey()
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(1000)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifySignature(msg, sig))

	// ----
	// Mutate the signature, just one bit.
	// TODO: Replace this with a much better fuzzer, tendermint/ed25519/issues/10
	sig[7] ^= byte(0x01)
	assert.False(t, pubKey.VerifySignature(msg, sig))
}

func TestPubKeyEquals(t *testing.T) {
	sr25519PubKey := sr25519.GenPrivKey().PubKey().(*sr25519.PubKey)

	testCases := []struct {
		msg      string
		pubKey   cryptotypes.PubKey
		other    cryptotypes.PubKey
		expectEq bool
	}{
		{
			"different bytes",
			sr25519PubKey,
			sr25519.GenPrivKey().PubKey(),
			false,
		},
		{
			"equals",
			sr25519PubKey,
			&sr25519.PubKey{
				Key: sr25519PubKey.Key,
			},
			true,
		},
		{
			"different types",
			sr25519PubKey,
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

func TestAddressSr25519(t *testing.T) {
	pk := sr25519.PubKey{[]byte{125, 80, 29, 208, 159, 53, 119, 198, 73, 53, 187, 33, 199, 144, 62, 255, 1, 235, 117, 96, 128, 211, 17, 45, 34, 64, 189, 165, 33, 182, 54, 206}}
	addr := pk.Address()
	require.Len(t, addr, 20, "Address must be 20 bytes long")
}

func TestPrivKeyEquals(t *testing.T) {
	sr25519PrivKey := sr25519.GenPrivKey()

	testCases := []struct {
		msg      string
		privKey  cryptotypes.PrivKey
		other    cryptotypes.PrivKey
		expectEq bool
	}{
		{
			"different bytes",
			sr25519PrivKey,
			sr25519.GenPrivKey(),
			false,
		},
		{
			"equals",
			sr25519PrivKey,
			&sr25519.PrivKey{
				Key: sr25519PrivKey.Key,
			},
			true,
		},
		{
			"different types",
			sr25519PrivKey,
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
	privKey := sr25519.GenPrivKey()
	pubKey := privKey.PubKey().(*sr25519.PubKey)

	testCases := []struct {
		desc      string
		msg       codec.AminoMarshaler
		typ       interface{}
		expBinary []byte
		expJSON   string
	}{
		{
			"sr25519 private key",
			privKey,
			&sr25519.PrivKey{},
			append([]byte{64}, privKey.Bytes()...), // Length-prefixed.
			"\"" + base64.StdEncoding.EncodeToString(privKey.Bytes()) + "\"",
		},
		{
			"sr25519 public key",
			pubKey,
			&sr25519.PubKey{},
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
	tmPrivKey := tmsr25519.GenPrivKey()
	tmPubKey := tmPrivKey.PubKey()
	// Create our own keys, with the same private key as Tendermint's.
	privKey := &sr25519.PrivKey{Key: []byte(tmPrivKey)}
	pubKey := privKey.PubKey().(*sr25519.PubKey)

	testCases := []struct {
		desc      string
		tmKey     interface{}
		ourKey    interface{}
		marshalFn func(o interface{}) ([]byte, error)
	}{
		{
			"sr25519 private key, binary",
			tmPrivKey,
			privKey,
			aminoCdc.MarshalBinaryBare,
		},
		{
			"sr25519 private key, JSON",
			tmPrivKey,
			privKey,
			aminoCdc.MarshalJSON,
		},
		{
			"sr25519 public key, binary",
			tmPubKey,
			pubKey,
			aminoCdc.MarshalBinaryBare,
		},
		{
			"sr25519 public key, JSON",
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
