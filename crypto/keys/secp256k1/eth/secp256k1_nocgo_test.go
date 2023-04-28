//go:build !libsecp256k1_sdk
// +build !libsecp256k1_sdk

package eth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"

	secp "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1/internal/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

func pubkeyToAddress(pub ecdsa.PublicKey) []byte {
	if pub.X == nil || pub.Y == nil {
		return nil
	}
	return Keccak256(elliptic.Marshal(secp.S256(), pub.X, pub.Y)[1:])[12:]
}

func TestPrivKey(t *testing.T) {
	// validate type and equality
	privKey, err := GenerateKey()
	require.NoError(t, err)
	require.Implements(t, (*cryptotypes.PrivKey)(nil), privKey)

	// validate inequality
	privKey2, err := GenerateKey()
	require.NoError(t, err)
	require.False(t, privKey.Equals(privKey2))

	// validate Ethereum address equality
	addr := privKey.PubKey().Address()
	key, err := privKey.ToECDSA()
	require.NoError(t, err)
	expectedAddr := pubkeyToAddress(key.PublicKey)
	require.Equal(t, expectedAddr, addr.Bytes())

	// validate we can sign some bytes
	msg := []byte("hello world")
	sigHash := Keccak256(msg)
	expectedSig, err := secp.Sign(sigHash, privKey.Bytes())
	require.NoError(t, err)

	sig, err := privKey.Sign(sigHash)
	require.NoError(t, err)
	require.Equal(t, expectedSig, sig)
}

func TestPrivKey_PubKey(t *testing.T) {
	privKey, err := GenerateKey()
	require.NoError(t, err)

	// validate type and equality
	pubKey := &PubKey{
		Key: privKey.PubKey().Bytes(),
	}
	require.Implements(t, (*cryptotypes.PubKey)(nil), pubKey)

	// validate inequality
	privKey2, err := GenerateKey()
	require.NoError(t, err)
	require.False(t, pubKey.Equals(privKey2.PubKey()))

	// validate signature
	msg := []byte("hello world")
	sig, err := privKey.Sign(msg)
	require.NoError(t, err)

	require.True(t, pubKey.VerifySignature(msg, sig))
}

func TestMarshalAmino(t *testing.T) {
	aminoCdc := codec.NewLegacyAmino()
	privKey, err := GenerateKey()
	require.NoError(t, err)

	pubKey := privKey.PubKey().(*PubKey)

	testCases := []struct {
		desc      string
		msg       codec.AminoMarshaler
		typ       interface{}
		expBinary []byte
		expJSON   string
	}{
		{
			"ethsecp256k1 private key",
			privKey,
			&PrivKey{},
			append([]byte{32}, privKey.Bytes()...), // Length-prefixed.
			"\"" + base64.StdEncoding.EncodeToString(privKey.Bytes()) + "\"",
		},
		{
			"ethsecp256k1 public key",
			pubKey,
			&PubKey{},
			append([]byte{33}, pubKey.Bytes()...), // Length-prefixed.
			"\"" + base64.StdEncoding.EncodeToString(pubKey.Bytes()) + "\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Do a round trip of encoding/decoding binary.
			bz, err := aminoCdc.Marshal(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.expBinary, bz)

			err = aminoCdc.Unmarshal(bz, tc.typ)
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
