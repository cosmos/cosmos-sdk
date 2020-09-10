package keys_test

import (
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/sr25519"
)

func TestPubKeyEquals(t *testing.T) {
	secp256K1PubKey := secp256k1.GenPrivKey().PubKey().(secp256k1.PubKey)
	secp256K1PbPubKey := &keys.Secp256K1PubKey{Key: secp256K1PubKey}

	testCases := []struct {
		msg      string
		pubKey   crypto.PubKey
		other    tmcrypto.PubKey
		expectEq bool
	}{
		{
			"secp256k1 pb different bytes",
			secp256K1PbPubKey,
			&keys.Secp256K1PubKey{
				Key: secp256k1.GenPrivKey().PubKey().(secp256k1.PubKey),
			},
			false,
		},
		{
			"secp256k1 pb equals",
			secp256K1PbPubKey,
			&keys.Secp256K1PubKey{
				Key: secp256K1PubKey,
			},
			true,
		},
		{
			"secp256k1 different types",
			secp256K1PbPubKey,
			sr25519.GenPrivKey().PubKey(),
			false,
		},
		{
			"secp256k1 different bytes",
			secp256K1PbPubKey,
			secp256k1.GenPrivKey().PubKey(),
			false,
		},
		{
			"secp256k1 equals",
			secp256K1PbPubKey,
			secp256K1PubKey,
			true,
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
	secp256K1PrivKey := secp256k1.GenPrivKey()
	secp256K1PbPrivKey := &keys.Secp256K1PrivKey{Key: secp256K1PrivKey}

	testCases := []struct {
		msg      string
		privKey  crypto.PrivKey
		other    tmcrypto.PrivKey
		expectEq bool
	}{
		{
			"secp256k1 pb different bytes",
			secp256K1PbPrivKey,
			&keys.Secp256K1PrivKey{
				Key: secp256k1.GenPrivKey(),
			},
			false,
		},
		{
			"secp256k1 pb equals",
			secp256K1PbPrivKey,
			&keys.Secp256K1PrivKey{
				Key: secp256K1PrivKey,
			},
			true,
		},
		{
			"secp256k1 different types",
			secp256K1PbPrivKey,
			sr25519.GenPrivKey(),
			false,
		},
		{
			"secp256k1 different bytes",
			secp256K1PbPrivKey,
			secp256k1.GenPrivKey(),
			false,
		},
		{
			"secp256k1 equals",
			secp256K1PbPrivKey,
			secp256K1PrivKey,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			eq := tc.privKey.Equals(tc.other)
			require.Equal(t, eq, tc.expectEq)
		})
	}
}

func TestSignAndVerifySignature(t *testing.T) {
	testCases := []struct {
		msg     string
		privKey crypto.PrivKey
	}{
		{
			"secp256k1",
			&keys.Secp256K1PrivKey{Key: secp256k1.GenPrivKey()},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			pubKey := tc.privKey.PubKey()
			msg := tmcrypto.CRandBytes(128)
			sig, err := tc.privKey.Sign(msg)
			require.Nil(t, err)

			assert.True(t, pubKey.VerifySignature(msg, sig))

			sig[7] ^= byte(0x01)

			assert.False(t, pubKey.VerifySignature(msg, sig))
		})
	}

}

func TestMarshalAmino(t *testing.T) {
	aminoCdc := codec.NewLegacyAmino()
	privKey := secp256k1.GenPrivKey()

	testCases := []struct {
		desc      string
		msg       codec.AminoMarshaler
		expBinary []byte
		expJSON   string
	}{
		{
			"secp256k1 private key",
			&keys.Secp256K1PrivKey{Key: privKey},
			append([]byte{32}, privKey.Bytes()...), // Length-prefixed.
			"\"" + base64.StdEncoding.EncodeToString(privKey.Bytes()) + "\"",
		},
		{
			"secp256k1 public key",
			&keys.Secp256K1PubKey{Key: privKey.PubKey().(secp256k1.PubKey)},
			append([]byte{33}, privKey.PubKey().Bytes()...),
			"\"" + base64.StdEncoding.EncodeToString(privKey.PubKey().Bytes()) + "\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Do a round trip of encoding/decoding binary.
			bz, err := aminoCdc.MarshalBinaryBare(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.expBinary, bz)

			// Create a new empty value of the same type as `tc.msg`.
			newMsg := reflect.New(reflect.TypeOf(tc.msg).Elem()).Interface().(codec.AminoMarshaler)
			err = aminoCdc.UnmarshalBinaryBare(bz, newMsg)
			require.NoError(t, err)

			require.Equal(t, tc.msg, newMsg)

			// Do a round trip of encoding/decoding JSON.
			bz, err = aminoCdc.MarshalJSON(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.expJSON, string(bz))

			// Create a new empty value of the same type as `tc.msg`.
			newMsg = reflect.New(reflect.TypeOf(tc.msg).Elem()).Interface().(codec.AminoMarshaler)
			err = aminoCdc.UnmarshalJSON(bz, newMsg)
			require.NoError(t, err)

			require.Equal(t, tc.msg, newMsg)
		})
	}
}
