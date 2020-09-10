package keys_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/types"
	keys "github.com/cosmos/cosmos-sdk/crypto/keys"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/stretchr/testify/require"
)

func TestAddress(t *testing.T) {
	msg := []byte{1, 2, 3, 4}
	pubKeys, _ := generatePubKeysAndSignatures(5, msg)
	anyPubKeys, err := keys.PackPubKeys(pubKeys)
	require.NoError(t, err)
	multisigKey := &keys.LegacyAminoMultisigThresholdPubKey{K: 2, PubKeys: anyPubKeys}

	require.Len(t, multisigKey.Address().Bytes(), 20)
}

func TestEquals(t *testing.T) {
	pubKey1 := secp256k1.GenPrivKey().PubKey()
	pubKey2 := secp256k1.GenPrivKey().PubKey()

	pbPubKey1 := &keys.Secp256K1PubKey{Key: pubKey1.(secp256k1.PubKey)}
	pbPubKey2 := &keys.Secp256K1PubKey{Key: pubKey2.(secp256k1.PubKey)}
	anyPubKeys, err := keys.PackPubKeys([]tmcrypto.PubKey{pbPubKey1, pbPubKey2})
	require.NoError(t, err)
	multisigKey := keys.LegacyAminoMultisigThresholdPubKey{K: 1, PubKeys: anyPubKeys}

	otherPubKeys, err := keys.PackPubKeys([]tmcrypto.PubKey{pbPubKey1, &multisigKey})
	require.NoError(t, err)
	otherMultisigKey := keys.LegacyAminoMultisigThresholdPubKey{K: 1, PubKeys: otherPubKeys}

	testCases := []struct {
		msg      string
		other    tmcrypto.PubKey
		expectEq bool
	}{
		{
			"equals with proto pub key",
			&keys.LegacyAminoMultisigThresholdPubKey{K: 1, PubKeys: anyPubKeys},
			true,
		},
		{
			"different threshold",
			&keys.LegacyAminoMultisigThresholdPubKey{K: 2, PubKeys: anyPubKeys},
			false,
		},
		{
			"different pub keys length",
			&keys.LegacyAminoMultisigThresholdPubKey{K: 1, PubKeys: []*types.Any{anyPubKeys[0]}},
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
			multisig.NewPubKeyMultisigThreshold(1, []tmcrypto.PubKey{pubKey1, pubKey2}),
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
	var (
		pk  multisig.PubKey
		sig *signing.MultiSignatureData
	)
	msg := []byte{1, 2, 3, 4}
	signBytesFn := func(mode signing.SignMode) ([]byte, error) { return msg, nil }

	testCases := []struct {
		msg        string
		malleate   func()
		expectPass bool
	}{
		{
			"nested multisignature",
			func() {
				genPk, genSig, err := generateNestedMultiSignature(3, msg)
				require.NoError(t, err)
				sig = genSig
				pk = genPk
			},
			true,
		},
		{
			"wrong size for sig bit array",
			func() {
				pubKeys, _ := generatePubKeysAndSignatures(3, msg)
				anyPubKeys, err := keys.PackPubKeys(pubKeys)
				require.NoError(t, err)
				pk = &keys.LegacyAminoMultisigThresholdPubKey{K: 3, PubKeys: anyPubKeys}
				sig = multisig.NewMultisig(1)
			},
			false,
		},
		{
			"single signature data",
			func() {
				k := 2
				signingIndices := []int{0, 3, 1}
				pubKeys, sigs := generatePubKeysAndSignatures(5, msg)
				anyPubKeys, err := keys.PackPubKeys(pubKeys)
				require.NoError(t, err)
				pk = &keys.LegacyAminoMultisigThresholdPubKey{K: uint32(k), PubKeys: anyPubKeys}
				sig = multisig.NewMultisig(len(pubKeys))
				signBytesFn := func(mode signing.SignMode) ([]byte, error) { return msg, nil }

				for i := 0; i < k-1; i++ {
					signingIndex := signingIndices[i]
					require.NoError(
						t,
						multisig.AddSignatureFromPubKey(sig, sigs[signingIndex], pubKeys[signingIndex], pubKeys),
					)
					require.Error(
						t,
						pk.VerifyMultisignature(signBytesFn, sig),
						"multisig passed when i < k, i %d", i,
					)
					require.NoError(
						t,
						multisig.AddSignatureFromPubKey(sig, sigs[signingIndex], pubKeys[signingIndex], pubKeys),
					)
				}
				require.Error(
					t,
					pk.VerifyMultisignature(signBytesFn, sig),
					"multisig passed with k - 1 sigs",
				)
				require.NoError(
					t,
					multisig.AddSignatureFromPubKey(
						sig,
						sigs[signingIndices[k]],
						pubKeys[signingIndices[k]],
						pubKeys,
					),
				)
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			tc.malleate()
			err := pk.VerifyMultisignature(signBytesFn, sig)
			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func generatePubKeysAndSignatures(n int, msg []byte) (pubKeys []tmcrypto.PubKey, signatures []signing.SignatureData) {
	pubKeys = make([]tmcrypto.PubKey, n)
	signatures = make([]signing.SignatureData, n)

	for i := 0; i < n; i++ {
		privkey := &keys.Secp256K1PrivKey{Key: secp256k1.GenPrivKey()}
		pubKeys[i] = privkey.PubKey()

		sig, _ := privkey.Sign(msg)
		signatures[i] = &signing.SingleSignatureData{Signature: sig}
	}
	return
}

func generateNestedMultiSignature(n int, msg []byte) (multisig.PubKey, *signing.MultiSignatureData, error) {
	pubKeys := make([]tmcrypto.PubKey, n)
	signatures := make([]signing.SignatureData, n)
	bitArray := crypto.NewCompactBitArray(n)
	for i := 0; i < n; i++ {
		nestedPks, nestedSigs := generatePubKeysAndSignatures(5, msg)
		nestedBitArray := crypto.NewCompactBitArray(5)
		for j := 0; j < 5; j++ {
			nestedBitArray.SetIndex(j, true)
		}
		nestedSig := &signing.MultiSignatureData{
			BitArray:   nestedBitArray,
			Signatures: nestedSigs,
		}
		signatures[i] = nestedSig
		anyNestedPks, err := keys.PackPubKeys(nestedPks)
		if err != nil {
			return nil, nil, err
		}
		pubKeys[i] = &keys.LegacyAminoMultisigThresholdPubKey{K: 5, PubKeys: anyNestedPks}
		bitArray.SetIndex(i, true)
	}
	anyPubKeys, err := keys.PackPubKeys(pubKeys)
	if err != nil {
		return nil, nil, err
	}
	return &keys.LegacyAminoMultisigThresholdPubKey{K: uint32(n), PubKeys: anyPubKeys}, &signing.MultiSignatureData{
		BitArray:   bitArray,
		Signatures: signatures,
	}, nil
}
