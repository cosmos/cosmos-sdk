package keys_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/sr25519"
)

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
			"LegacyAminoMultisigThresholdPubKey",
			func() {
				pubKey1 := secp256k1.GenPrivKey().PubKey()
				pubKey2 := secp256k1.GenPrivKey().PubKey()

				pbPubKey1 := &keys.Secp256K1PubKey{Key: pubKey1.(secp256k1.PubKey)}
				pbPubKey2 := &keys.Secp256K1PubKey{Key: pubKey2.(secp256k1.PubKey)}
				anyPubKeys, err := packPubKeys([]crypto.PubKey{pbPubKey1, pbPubKey2})
				require.NoError(t, err)

				pbPubKey = &keys.LegacyAminoMultisigThresholdPubKey{K: 1, PubKeys: anyPubKeys}
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

				pbPubKey = &keys.LegacyAminoMultisigThresholdPubKey{K: 1, PubKeys: anyPubKeys}
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
