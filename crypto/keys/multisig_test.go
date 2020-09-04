package keys_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/types"
	keys "github.com/cosmos/cosmos-sdk/crypto/keys"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	proto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/stretchr/testify/require"
)

func TestAddress(t *testing.T) {
	msg := []byte{1, 2, 3, 4}
	pubKeys, _ := generatePubKeysAndSignatures(5, msg)
	anyPubKeys, err := packPubKeys(pubKeys)
	require.NoError(t, err)
	multisigKey := &keys.MultisigThresholdPubKey{K: 2, PubKeys: anyPubKeys}

	require.Len(t, multisigKey.Address().Bytes(), 20)
}

func TestEquals(t *testing.T) {
	pubKey1 := secp256k1.GenPrivKey().PubKey()
	pubKey2 := secp256k1.GenPrivKey().PubKey()

	pbPubKey1 := &keys.Secp256K1PubKey{Key: pubKey1.(secp256k1.PubKey)}
	pbPubKey2 := &keys.Secp256K1PubKey{Key: pubKey2.(secp256k1.PubKey)}
	anyPubKeys, err := packPubKeys([]crypto.PubKey{pbPubKey1, pbPubKey2})
	require.NoError(t, err)
	multisigKey := keys.MultisigThresholdPubKey{K: 1, PubKeys: anyPubKeys}

	otherPubKeys, err := packPubKeys([]crypto.PubKey{pbPubKey1, &multisigKey})
	require.NoError(t, err)
	otherMultisigKey := keys.MultisigThresholdPubKey{K: 1, PubKeys: otherPubKeys}

	testCases := []struct {
		msg      string
		other    crypto.PubKey
		expectEq bool
	}{
		{
			"equals with proto pub key",
			&keys.MultisigThresholdPubKey{K: 1, PubKeys: anyPubKeys},
			true,
		},
		{
			"different threshold",
			&keys.MultisigThresholdPubKey{K: 2, PubKeys: anyPubKeys},
			false,
		},
		{
			"different pub keys length",
			&keys.MultisigThresholdPubKey{K: 1, PubKeys: []*types.Any{anyPubKeys[0]}},
			false,
		},
		{
			"different pub keys",
			&otherMultisigKey,
			false,
		},
		{
			"different types",
			secp256k1.GenPrivKey().PubKey(),
			false,
		},
		{
			"equals with amino pub key",
			multisig.NewPubKeyMultisigThreshold(1, []crypto.PubKey{pubKey1, pubKey2}),
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			eq := multisigKey.Equals(tc.other)
			require.Equal(t, eq, tc.expectEq)
		})
	}
}

func TestVerifyMultisignature(t *testing.T) {
	// TODO
}

func generatePubKeysAndSignatures(n int, msg []byte) (pubKeys []crypto.PubKey, signatures []signing.SignatureData) {
	pubKeys = make([]crypto.PubKey, n)
	signatures = make([]signing.SignatureData, n)

	for i := 0; i < n; i++ {
		privkey := &keys.Secp256K1PrivKey{Key: secp256k1.GenPrivKey()}
		pubKeys[i] = privkey.PubKey()

		sig, _ := privkey.Sign(msg)
		signatures[i] = &signing.SingleSignatureData{Signature: sig}
	}
	return
}

func packPubKeys(pubKeys []crypto.PubKey) ([]*types.Any, error) {
	anyPubKeys := make([]*types.Any, len(pubKeys))

	for i := 0; i < len(pubKeys); i++ {
		any, err := types.NewAnyWithValue(pubKeys[i].(proto.Message))
		if err != nil {
			return nil, err
		}
		anyPubKeys[i] = any
	}
	return anyPubKeys, nil
}
