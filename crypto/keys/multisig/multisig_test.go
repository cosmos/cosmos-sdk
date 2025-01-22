package multisig_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	_ "github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

func TestNewMultiSig(t *testing.T) {
	require := require.New(t)
	pk1 := secp256k1.GenPrivKey().PubKey()
	pks := []cryptotypes.PubKey{pk1, pk1}

	require.NotNil(kmultisig.NewLegacyAminoPubKey(1, pks),
		"Should support not unique public keys")
}

func TestAddress(t *testing.T) {
	pubKeys := generatePubKeys(5)
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
	signBytesFn := func(mode apisigning.SignMode) ([]byte, error) { return msg, nil }

	testCases := []struct {
		msg        string
		malleate   func(*require.Assertions)
		expectPass bool
	}{
		{
			"nested multisignature",
			func(require *require.Assertions) {
				genPk, genSig := generateNestedMultiSignature(3, msg)
				sig = genSig
				pk = genPk
			},
			true,
		},
		{
			"wrong size for sig bit array",
			func(require *require.Assertions) {
				pubKeys := generatePubKeys(3)
				pk = kmultisig.NewLegacyAminoPubKey(3, pubKeys)
				sig = multisig.NewMultisig(1)
			},
			false,
		},
		{
			"single signature data, expects the first k signatures to be valid",
			func(require *require.Assertions) {
				k := 2
				signingIndices := []int{0, 3, 1}
				pubKeys, sigs := generatePubKeysAndSignatures(8, msg)
				pk = kmultisig.NewLegacyAminoPubKey(k, pubKeys)
				sig = multisig.NewMultisig(len(pubKeys))
				signBytesFn := func(mode apisigning.SignMode) ([]byte, error) { return msg, nil }

				for i := 0; i < k-1; i++ {
					signingIndex := signingIndices[i]
					require.NoError(
						multisig.AddSignatureFromPubKey(sig, sigs[signingIndex], pubKeys[signingIndex], pubKeys),
					)
					require.Error(
						pk.VerifyMultisignature(signBytesFn, sig),
						"multisig passed when i < k, i %d", i,
					)
					require.NoError(
						multisig.AddSignatureFromPubKey(sig, sigs[signingIndex], pubKeys[signingIndex], pubKeys),
					)
					require.Equal(
						i+1,
						len(sig.Signatures),
						"adding a signature for the same pubkey twice increased signature count by 2, index %d", i,
					)
				}
				require.Error(
					pk.VerifyMultisignature(signBytesFn, sig),
					"multisig passed with k - 1 sigs",
				)
				require.NoError(
					multisig.AddSignatureFromPubKey(
						sig,
						sigs[signingIndices[k]],
						pubKeys[signingIndices[k]],
						pubKeys,
					),
				)
				require.NoError(
					pk.VerifyMultisignature(signBytesFn, sig),
					"multisig failed after k good signatures",
				)
			},
			true,
		},
		{
			"duplicate signatures",
			func(require *require.Assertions) {
				pubKeys, sigs := generatePubKeysAndSignatures(5, msg)
				pk = kmultisig.NewLegacyAminoPubKey(2, pubKeys)
				sig = multisig.NewMultisig(5)

				require.Error(pk.VerifyMultisignature(signBytesFn, sig))
				require.NoError(multisig.AddSignatureFromPubKey(sig, sigs[0], pubKeys[0], pubKeys))
				// Add second signature manually
				sig.Signatures = append(sig.Signatures, sigs[0])
			},
			false,
		},
		{
			"duplicated key",
			func(require *require.Assertions) {
				// here we test an edge case where we create a multi sig with two same
				// keys. It  should work.
				pubkeys, sigs := generatePubKeysAndSignatures(3, msg)
				pubkeys[1] = pubkeys[0]
				pk = kmultisig.NewLegacyAminoPubKey(2, pubkeys)
				sig = multisig.NewMultisig(len(pubkeys))
				multisig.AddSignature(sig, sigs[0], 0)
				multisig.AddSignature(sig, sigs[0], 1)
			},
			true,
		},
		{
			"same key used twice",
			func(require *require.Assertions) {
				pubkeys, sigs := generatePubKeysAndSignatures(3, msg)
				pk = kmultisig.NewLegacyAminoPubKey(2, pubkeys)
				sig = multisig.NewMultisig(len(pubkeys))
				multisig.AddSignature(sig, sigs[0], 0)
				multisig.AddSignature(sig, sigs[0], 1)
			},
			false,
		},
		{
			"unable to verify signature",
			func(require *require.Assertions) {
				pubKeys := generatePubKeys(2)
				_, sigs := generatePubKeysAndSignatures(2, msg)
				pk = kmultisig.NewLegacyAminoPubKey(2, pubKeys)
				sig = multisig.NewMultisig(2)
				require.NoError(multisig.AddSignatureFromPubKey(sig, sigs[0], pubKeys[0], pubKeys))
				require.NoError(multisig.AddSignatureFromPubKey(sig, sigs[1], pubKeys[1], pubKeys))
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			tc.malleate(require.New(t))
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
	pkSet, sigs := generatePubKeysAndSignatures(3, msg)
	multisignature := multisig.NewMultisig(3)

	multisigKey := kmultisig.NewLegacyAminoPubKey(2, pkSet)
	signBytesFn := func(mode apisigning.SignMode) ([]byte, error) { return msg, nil }

	cdc := codec.NewLegacyAmino()

	require.NoError(t, multisig.AddSignatureFromPubKey(multisignature, sigs[0], pkSet[0], pkSet))

	// create a StdSignature for msg, and convert it to sigV2
	sig := legacytx.StdSignature{PubKey: pkSet[1], Signature: sigs[1].(*signing.SingleSignatureData).Signature} //nolint:staticcheck // SA1019: legacytx.StdSignature is deprecated: use Tx.Msgs, Signatures and Memo instead.
	sigV2, err := legacytx.StdSignatureToSignatureV2(cdc, sig)
	require.NoError(t, multisig.AddSignatureV2(multisignature, sigV2, pkSet))

	require.NoError(t, err)
	require.NotNil(t, sigV2)

	require.NoError(t, multisigKey.VerifyMultisignature(signBytesFn, multisignature))
}

func TestPubKeyMultisigThresholdAminoToIface(t *testing.T) {
	pubkeys := generatePubKeys(5)
	multisigKey := kmultisig.NewLegacyAminoPubKey(2, pubkeys)

	ab, err := legacy.Cdc.MarshalLengthPrefixed(multisigKey)
	require.NoError(t, err)
	// like other cryptotypes.Pubkey implementations (e.g. ed25519.PubKey),
	// LegacyAminoPubKey should be deserializable into a cryptotypes.LegacyAminoPubKey:
	var pubKey kmultisig.LegacyAminoPubKey
	err = legacy.Cdc.UnmarshalLengthPrefixed(ab, &pubKey)
	require.NoError(t, err)

	require.Equal(t, multisigKey.Equals(&pubKey), true)
}

func generatePubKeys(n int) []cryptotypes.PubKey {
	pks := make([]cryptotypes.PubKey, n)
	for i := 0; i < n; i++ {
		pks[i] = secp256k1.GenPrivKey().PubKey()
	}
	return pks
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

// generateNestedMultiSignature creates a nested multisig structure for testing purposes.
// It generates a top-level multisig with 'n' sub-multisigs, where each sub-multisig contains
// 5 individual signatures.
//
// Parameters:
// - n: number of nested multisigs to generate (determines the size of the top-level multisig)
// - msg: the message to be signed
//
// Returns:
// - multisig.PubKey: the top-level multisig public key containing n sub-multisig public keys
// - *signing.MultiSignatureData: the corresponding signature data structure containing n sub-multisig signatures
//
// Each sub-multisig is configured with:
// - threshold of 5 (requires all 4 signatures)
// - 5 individual secp256k1 public keys and their corresponding signatures
// All signature bits are set to true to simulate a fully signed multisig.
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
		pubKeys[i] = kmultisig.NewLegacyAminoPubKey(4, nestedPks)
		bitArray.SetIndex(i, true)
	}
	return kmultisig.NewLegacyAminoPubKey(n-1, pubKeys), &signing.MultiSignatureData{
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

func TestDisplay(t *testing.T) {
	require := require.New(t)
	pubKeys := generatePubKeys(3)
	msig := kmultisig.NewLegacyAminoPubKey(2, pubKeys)

	require.NotEmpty(msig.String())
	var cdc codec.Codec
	err := depinject.Inject(
		depinject.Configs(
			configurator.NewAppConfig(),
			depinject.Supply(log.NewNopLogger(),
				func() address.Codec { return addresscodec.NewBech32Codec("cosmos") },
				func() address.ValidatorAddressCodec { return addresscodec.NewBech32Codec("cosmosvaloper") },
				func() address.ConsensusAddressCodec { return addresscodec.NewBech32Codec("cosmosvalcons") },
			),
		), &cdc)
	require.NoError(err)
	bz, err := cdc.MarshalInterfaceJSON(msig)
	require.NoError(err)
	expectedPrefix := `{"@type":"/cosmos.crypto.multisig.LegacyAminoPubKey","threshold":2,"public_keys":[{"@type":"/cosmos.crypto.secp256k1.PubKey"`
	require.True(strings.HasPrefix(string(bz), expectedPrefix))
	// Example output:
	// {"@type":"/cosmos.crypto.multisig.LegacyAminoPubKey","threshold":2,"public_keys":[{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AymUY3J2HKIyy9cbpGKcBFUTuDQsRH9NO/orKF/0WQ76"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AkvnCDzSYF+tQV/FoI217V7CDIRPzjJj7zBE2nw7x3xT"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A0yiqgcM5EB1i0h79+sQp+C0jLPFnT3+dFmdZmGa+H1s"}]}
}

func TestAminoBinary(t *testing.T) {
	pubkeys := generatePubKeys(2)
	msig := kmultisig.NewLegacyAminoPubKey(2, pubkeys)

	// Do a round-trip key->bytes->key.
	bz, err := legacy.Cdc.Marshal(msig)
	require.NoError(t, err)
	var newMsig cryptotypes.PubKey
	err = legacy.Cdc.Unmarshal(bz, &newMsig)
	require.NoError(t, err)
	require.Equal(t, msig.Threshold, newMsig.(*kmultisig.LegacyAminoPubKey).Threshold)
}

func TestAminoMarshalJSON(t *testing.T) {
	pubkeys := generatePubKeys(2)
	multisigKey := kmultisig.NewLegacyAminoPubKey(2, pubkeys)
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

func TestVerifyMultisignatureNMRule(t *testing.T) {
	makeTestKeysAndSignatures := func(n int, msg []byte) ([]cryptotypes.PubKey, []signing.SignatureData) {
		pubKeys := make([]cryptotypes.PubKey, n)
		sigs := make([]signing.SignatureData, n)

		for i := 0; i < n; i++ {
			// Generate private key and get its public key
			privKey := secp256k1.GenPrivKey()
			pubKeys[i] = privKey.PubKey()

			// Create real signature
			sig, err := privKey.Sign(msg)
			require.NoError(t, err)
			sigs[i] = &signing.SingleSignatureData{
				Signature: sig,
			}
		}
		return pubKeys, sigs
	}

	tests := []struct {
		name      string
		n         int    // number of keys
		m         uint32 // threshold
		expectErr string
	}{
		{
			name:      "valid case: N=3 M=2",
			n:         3,
			m:         2,
			expectErr: "",
		},
		{
			name:      "invalid: M=0",
			n:         3,
			m:         0,
			expectErr: "invalid threshold: must be > 0, got 0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := []byte("test message")
			pubKeys, sigs := makeTestKeysAndSignatures(tc.n, msg)

			// Create the key directly instead of using NewLegacyAminoPubKey
			pubKeysAny := make([]*types.Any, len(pubKeys))
			for i, key := range pubKeys {
				anyKey, err := types.NewAnyWithValue(key)
				require.NoError(t, err)
				pubKeysAny[i] = anyKey
			}

			multisigKey := &kmultisig.LegacyAminoPubKey{
				Threshold: tc.m,
				PubKeys:   pubKeysAny,
			}

			// Create a dummy signature with a proper bit array
			bitArray := cryptotypes.NewCompactBitArray(tc.n)
			// Set the first M bits to simulate M signatures
			for i := 0; i < int(tc.m) && i < tc.n; i++ {
				bitArray.SetIndex(i, true)
			}

			multiSig := &signing.MultiSignatureData{
				BitArray:   bitArray,
				Signatures: make([]signing.SignatureData, 0, tc.m),
			}

			// Fill in dummy signatures
			for i := 0; i < int(tc.m) && i < tc.n; i++ {
				multiSig.Signatures = append(multiSig.Signatures, sigs[i])
			}

			// Create getSignBytes function that returns our test message
			getSignBytes := func(mode apisigning.SignMode) ([]byte, error) {
				return msg, nil
			}

			err := multisigKey.VerifyMultisignature(getSignBytes, multiSig)

			if tc.expectErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErr)
			}
		})
	}
}
