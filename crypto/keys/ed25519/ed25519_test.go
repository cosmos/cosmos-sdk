package ed25519_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/sr25519"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

func TestSignAndValidateEd25519(t *testing.T) {

	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifySignature(msg, sig))

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
		other    crypto.PubKey
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
			sr25519.GenPrivKey().PubKey(),
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
		other    crypto.PrivKey
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
			sr25519.GenPrivKey(),
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
