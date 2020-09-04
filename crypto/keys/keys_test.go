package keys_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	proto "github.com/gogo/protobuf/proto"
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
		other    crypto.PubKey
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
		other    crypto.PrivKey
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

func TestProtoPubKeyToAminoPubKey(t *testing.T) {
	var (
		pbPubKey    proto.Message
		aminoPubKey crypto.PubKey
	)
	testCases := []struct {
		msg        string
		malleate   func()
		expectPass bool
	}{
		{
			"Secp256K1PubKey",
			func() {
				aminoPubKey = secp256k1.GenPrivKey().PubKey()
				pbPubKey = &keys.Secp256K1PubKey{Key: aminoPubKey.(secp256k1.PubKey)}
			},
			true,
		},
		{
			"MultisigThresholdPubKey",
			func() {
				pubKey1 := secp256k1.GenPrivKey().PubKey()
				pubKey2 := secp256k1.GenPrivKey().PubKey()

				pbPubKey1 := &keys.Secp256K1PubKey{Key: pubKey1.(secp256k1.PubKey)}
				pbPubKey2 := &keys.Secp256K1PubKey{Key: pubKey2.(secp256k1.PubKey)}
				anyPubKeys, err := packPubKeys([]crypto.PubKey{pbPubKey1, pbPubKey2})
				require.NoError(t, err)

				pbPubKey = &keys.MultisigThresholdPubKey{K: 1, PubKeys: anyPubKeys}
				aminoPubKey = multisig.NewPubKeyMultisigThreshold(1, []crypto.PubKey{pubKey1, pubKey2})
			},
			true,
		},
		{
			"unknown type",
			func() {
				pbPubKey = &testdata.Dog{}
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			tc.malleate()
			pubKey, err := keys.ProtoPubKeyToAminoPubKey(pbPubKey)
			if tc.expectPass {
				require.NoError(t, err)
				require.Equal(t, aminoPubKey, pubKey)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAminoPubKeyToProtoPubKey(t *testing.T) {
	var (
		pbPubKey    proto.Message
		aminoPubKey crypto.PubKey
	)
	testCases := []struct {
		msg        string
		malleate   func()
		expectPass bool
	}{
		{
			"secp256k1.PubKey",
			func() {
				aminoPubKey = secp256k1.GenPrivKey().PubKey()
				pbPubKey = &keys.Secp256K1PubKey{Key: aminoPubKey.(secp256k1.PubKey)}
			},
			true,
		},
		{
			"multisig.PubKeyMultisigThreshold",
			func() {
				pubKey1 := secp256k1.GenPrivKey().PubKey()
				pubKey2 := secp256k1.GenPrivKey().PubKey()

				pbPubKey1 := &keys.Secp256K1PubKey{Key: pubKey1.(secp256k1.PubKey)}
				pbPubKey2 := &keys.Secp256K1PubKey{Key: pubKey2.(secp256k1.PubKey)}
				anyPubKeys, err := packPubKeys([]crypto.PubKey{pbPubKey1, pbPubKey2})
				require.NoError(t, err)

				pbPubKey = &keys.MultisigThresholdPubKey{K: 1, PubKeys: anyPubKeys}
				aminoPubKey = multisig.NewPubKeyMultisigThreshold(1, []crypto.PubKey{pubKey1, pubKey2})
			},
			true,
		},
		{
			"unknown type",
			func() {
				aminoPubKey = sr25519.GenPrivKey().PubKey()
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			tc.malleate()
			pubKey, err := keys.AminoPubKeyToProtoPubKey(aminoPubKey)
			if tc.expectPass {
				require.NoError(t, err)
				require.Equal(t, pbPubKey, pubKey)
			} else {
				require.Error(t, err)
			}
		})
	}
}
