package multisig_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

func generatePubKeys(n int) []cryptotypes.PubKey {
	pks := make([]cryptotypes.PubKey, n)
	for i := 0; i < n; i++ {
		pks[i] = secp256k1.GenPrivKey().PubKey()
	}

	return pks
}

func TestAddress(t *testing.T) {
	msg := []byte{1, 2, 3, 4}
	pubKeys, _ := generatePubKeysAndSignatures(5, msg)
	multisigKey := kmultisig.NewLegacyAminoPubKey(2, pubKeys)

	require.Len(t, multisigKey.Address().Bytes(), 20)
}

func TestEquals(t *testing.T) {
	pubKey1 := secp256k1.GenPrivKey().PubKey()
	pubKey2 := secp256k1.GenPrivKey().PubKey()

	multisigKey := kmultisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{pubKey1, pubKey2})
	otherMultisigKey := kmultisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{pubKey1, multisigKey})

	testCases := []struct {
		msg      string
		other    cryptotypes.PubKey
		expectEq bool
	}{
		{
			"equals with proto pub key",
			&kmultisig.LegacyAminoPubKey{Threshold: 1, PubKeys: multisigKey.PubKeys},
			true,
		},
		{
			"different threshold",
			&kmultisig.LegacyAminoPubKey{Threshold: 2, PubKeys: multisigKey.PubKeys},
			false,
		},
		{
			"different pub keys length",
			&kmultisig.LegacyAminoPubKey{Threshold: 1, PubKeys: []*types.Any{multisigKey.PubKeys[0]}},
			false,
		},
		{
			"different pub keys",
			otherMultisigKey,
			false,
		},
		{
			"different types",
			secp256k1.GenPrivKey().PubKey(),
			false,
		},
		{
			"ensure that reordering pubkeys is treated as a different pubkey",
			reorderPubKey(multisigKey),
			false,
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
				genPk, genSig := generateNestedMultiSignature(3, msg)
				sig = genSig
				pk = genPk
			},
			true,
		},
		{
			"wrong size for sig bit array",
			func() {
				pubKeys, _ := generatePubKeysAndSignatures(3, msg)
				pk = kmultisig.NewLegacyAminoPubKey(3, pubKeys)
				sig = multisig.NewMultisig(1)
			},
			false,
		},
		{
			"single signature data, expects the first k signatures to be valid",
			func() {
				k := 2
				signingIndices := []int{0, 3, 1}
				pubKeys, sigs := generatePubKeysAndSignatures(5, msg)
				pk = kmultisig.NewLegacyAminoPubKey(k, pubKeys)
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
					require.Equal(
						t,
						i+1,
						len(sig.Signatures),
						"adding a signature for the same pubkey twice increased signature count by 2, index %d", i,
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
				require.NoError(
					t,
					pk.VerifyMultisignature(signBytesFn, sig),
					"multisig failed after k good signatures",
				)
			},
			true,
		},
		{
			"duplicate signatures",
			func() {
				pubKeys, sigs := generatePubKeysAndSignatures(5, msg)
				pk = kmultisig.NewLegacyAminoPubKey(2, pubKeys)
				sig = multisig.NewMultisig(5)

				require.Error(t, pk.VerifyMultisignature(signBytesFn, sig))
				multisig.AddSignatureFromPubKey(sig, sigs[0], pubKeys[0], pubKeys)
				// Add second signature manually
				sig.Signatures = append(sig.Signatures, sigs[0])
			},
			false,
		},
		{
			"unable to verify signature",
			func() {
				pubKeys, _ := generatePubKeysAndSignatures(2, msg)
				_, sigs := generatePubKeysAndSignatures(2, msg)
				pk = kmultisig.NewLegacyAminoPubKey(2, pubKeys)
				sig = multisig.NewMultisig(2)
				multisig.AddSignatureFromPubKey(sig, sigs[0], pubKeys[0], pubKeys)
				multisig.AddSignatureFromPubKey(sig, sigs[1], pubKeys[1], pubKeys)
			},
			false,
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

func TestAddSignatureFromPubKeyNilCheck(t *testing.T) {
	pkSet, sigs := generatePubKeysAndSignatures(5, []byte{1, 2, 3, 4})
	multisignature := multisig.NewMultisig(5)

	// verify no error is returned with all non-nil values
	err := multisig.AddSignatureFromPubKey(multisignature, sigs[0], pkSet[0], pkSet)
	require.NoError(t, err)
	// verify error is returned when key value is nil
	err = multisig.AddSignatureFromPubKey(multisignature, sigs[0], pkSet[0], nil)
	require.Error(t, err)
	// verify error is returned when pubkey value is nil
	err = multisig.AddSignatureFromPubKey(multisignature, sigs[0], nil, pkSet)
	require.Error(t, err)
	// verify error is returned when signature value is nil
	err = multisig.AddSignatureFromPubKey(multisignature, nil, pkSet[0], pkSet)
	require.Error(t, err)
	// verify error is returned when multisignature value is nil
	err = multisig.AddSignatureFromPubKey(nil, sigs[0], pkSet[0], pkSet)
	require.Error(t, err)
}

func TestMultiSigMigration(t *testing.T) {
	msg := []byte{1, 2, 3, 4}
	pkSet, sigs := generatePubKeysAndSignatures(2, msg)
	multisignature := multisig.NewMultisig(2)

	multisigKey := kmultisig.NewLegacyAminoPubKey(2, pkSet)
	signBytesFn := func(mode signing.SignMode) ([]byte, error) { return msg, nil }

	cdc := codec.NewLegacyAmino()

	require.NoError(t, multisig.AddSignatureFromPubKey(multisignature, sigs[0], pkSet[0], pkSet))

	// create a StdSignature for msg, and convert it to sigV2
	sig := legacytx.StdSignature{PubKey: pkSet[1], Signature: sigs[1].(*signing.SingleSignatureData).Signature}
	sigV2, err := legacytx.StdSignatureToSignatureV2(cdc, sig)
	require.NoError(t, multisig.AddSignatureV2(multisignature, sigV2, pkSet))

	require.NoError(t, err)
	require.NotNil(t, sigV2)

	require.NoError(t, multisigKey.VerifyMultisignature(signBytesFn, multisignature))
}

func TestPubKeyMultisigThresholdAminoToIface(t *testing.T) {
	msg := []byte{1, 2, 3, 4}
	pubkeys, _ := generatePubKeysAndSignatures(5, msg)
	multisigKey := kmultisig.NewLegacyAminoPubKey(2, pubkeys)

	ab, err := legacy.Cdc.MarshalBinaryLengthPrefixed(multisigKey)
	require.NoError(t, err)
	// like other cryptotypes.Pubkey implementations (e.g. ed25519.PubKey),
	// LegacyAminoPubKey should be deserializable into a cryptotypes.LegacyAminoPubKey:
	var pubKey kmultisig.LegacyAminoPubKey
	err = legacy.Cdc.UnmarshalBinaryLengthPrefixed(ab, &pubKey)
	require.NoError(t, err)

	require.Equal(t, multisigKey.Equals(&pubKey), true)
}

func generatePubKeysAndSignatures(n int, msg []byte) (pubKeys []cryptotypes.PubKey, signatures []signing.SignatureData) {
	pubKeys = make([]cryptotypes.PubKey, n)
	signatures = make([]signing.SignatureData, n)

	for i := 0; i < n; i++ {
		privkey := secp256k1.GenPrivKey()
		pubKeys[i] = privkey.PubKey()

		sig, _ := privkey.Sign(msg)
		signatures[i] = &signing.SingleSignatureData{Signature: sig}
	}
	return
}

func generateNestedMultiSignature(n int, msg []byte) (multisig.PubKey, *signing.MultiSignatureData) {
	pubKeys := make([]cryptotypes.PubKey, n)
	signatures := make([]signing.SignatureData, n)
	bitArray := cryptotypes.NewCompactBitArray(n)
	for i := 0; i < n; i++ {
		nestedPks, nestedSigs := generatePubKeysAndSignatures(5, msg)
		nestedBitArray := cryptotypes.NewCompactBitArray(5)
		for j := 0; j < 5; j++ {
			nestedBitArray.SetIndex(j, true)
		}
		nestedSig := &signing.MultiSignatureData{
			BitArray:   nestedBitArray,
			Signatures: nestedSigs,
		}
		signatures[i] = nestedSig
		pubKeys[i] = kmultisig.NewLegacyAminoPubKey(5, nestedPks)
		bitArray.SetIndex(i, true)
	}
	return kmultisig.NewLegacyAminoPubKey(n, pubKeys), &signing.MultiSignatureData{
		BitArray:   bitArray,
		Signatures: signatures,
	}
}

func reorderPubKey(pk *kmultisig.LegacyAminoPubKey) (other *kmultisig.LegacyAminoPubKey) {
	pubkeysCpy := make([]*types.Any, len(pk.PubKeys))
	copy(pubkeysCpy, pk.PubKeys)
	pubkeysCpy[0] = pk.PubKeys[1]
	pubkeysCpy[1] = pk.PubKeys[0]
	other = &kmultisig.LegacyAminoPubKey{Threshold: 2, PubKeys: pubkeysCpy}
	return
}

func TestAminoBinary(t *testing.T) {
	pubKey1 := secp256k1.GenPrivKey().PubKey()
	pubKey2 := secp256k1.GenPrivKey().PubKey()
	multisigKey := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{pubKey1, pubKey2})

	// Do a round-trip key->bytes->key.
	bz, err := legacy.Cdc.MarshalBinaryBare(multisigKey)
	require.NoError(t, err)
	var newMultisigKey cryptotypes.PubKey
	err = legacy.Cdc.UnmarshalBinaryBare(bz, &newMultisigKey)
	require.NoError(t, err)
	require.Equal(t, multisigKey.Threshold, newMultisigKey.(*kmultisig.LegacyAminoPubKey).Threshold)
}

func TestAminoMarshalJSON(t *testing.T) {
	pubKey1 := secp256k1.GenPrivKey().PubKey()
	pubKey2 := secp256k1.GenPrivKey().PubKey()
	multisigKey := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{pubKey1, pubKey2})

	bz, err := legacy.Cdc.MarshalJSON(multisigKey)
	require.NoError(t, err)

	// Note the quotes around `"2"`. They are present because we are overriding
	// the Amino JSON marshaling of LegacyAminoPubKey (using tmMultisig).
	// Without the override, there would not be any quotes.
	require.Contains(t, string(bz), "\"threshold\":\"2\"")
}

func TestAminoUnmarshalJSON(t *testing.T) {
	// This is a real multisig from the Akash chain. It has been exported from
	// v0.39, hence the `threshold` field as a string.
	// We are testing that when unmarshaling this JSON into a LegacyAminoPubKey
	// with amino, there's no error.
	// ref: https://github.com/cosmos/cosmos-sdk/issues/8776
	pkJSON := `{
	"type": "tendermint/PubKeyMultisigThreshold",
	"value": {
		"pubkeys": [
			{
			"type": "tendermint/PubKeySecp256k1",
			"value": "AzYxq2VNeD10TyABwOgV36OVWDIMn8AtI4OFA0uQX2MK"
			},
			{
			"type": "tendermint/PubKeySecp256k1",
			"value": "A39cdsrm00bTeQ3RVZVqjkH8MvIViO9o99c8iLiNO35h"
			},
			{
			"type": "tendermint/PubKeySecp256k1",
			"value": "A/uLLCZph8MkFg2tCxqSMGwFfPHdt1kkObmmrqy9aiYD"
			},
			{
			"type": "tendermint/PubKeySecp256k1",
			"value": "A4mOMhM5gPDtBAkAophjRs6uDGZm4tD4Dbok3ai4qJi8"
			},
			{
			"type": "tendermint/PubKeySecp256k1",
			"value": "A90icFucrjNNz2SAdJWMApfSQcARIqt+M2x++t6w5fFs"
			}
		],
		"threshold": "3"
	}
}`

	cdc := codec.NewLegacyAmino()
	cryptocodec.RegisterCrypto(cdc)

	var pk cryptotypes.PubKey
	err := cdc.UnmarshalJSON([]byte(pkJSON), &pk)
	require.NoError(t, err)
	lpk := pk.(*kmultisig.LegacyAminoPubKey)
	require.Equal(t, uint32(3), lpk.Threshold)
	require.Equal(t, 5, len(pk.(*kmultisig.LegacyAminoPubKey).PubKeys))

	for _, key := range pk.(*kmultisig.LegacyAminoPubKey).PubKeys {
		require.NotNil(t, key)
		pk := secp256k1.PubKey{}
		err := pk.Unmarshal(key.Value)
		require.NoError(t, err)
	}
}

func TestProtoMarshalJSON(t *testing.T) {
	require := require.New(t)
	pubkeys := generatePubKeys(3)
	msig := kmultisig.NewLegacyAminoPubKey(2, pubkeys)

	registry := types.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	bz, err := cdc.MarshalInterfaceJSON(msig)
	require.NoError(err)

	var pk2 cryptotypes.PubKey
	err = cdc.UnmarshalInterfaceJSON(bz, &pk2)
	require.NoError(err)
	require.True(pk2.Equals(msig))
}
