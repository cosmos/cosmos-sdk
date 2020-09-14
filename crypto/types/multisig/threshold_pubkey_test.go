package multisig_test

import (
	"math/rand"
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/sr25519"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// This tests multisig functionality, but it expects the first k signatures to be valid
// TODO: Adapt it to give more flexibility about first k signatures being valid
func TestThresholdMultisigValidCases(t *testing.T) {
	pkSet1, sigSet1 := generatePubKeysAndSignatures(5, []byte{1, 2, 3, 4})
	cases := []struct {
		msg            []byte
		k              int
		pubkeys        []crypto.PubKey
		signingIndices []int
		// signatures should be the same size as signingIndices.
		signatures           []signing.SignatureData
		passAfterKSignatures []bool
	}{
		{
			msg:                  []byte{1, 2, 3, 4},
			k:                    2,
			pubkeys:              pkSet1,
			signingIndices:       []int{0, 3, 1},
			signatures:           sigSet1,
			passAfterKSignatures: []bool{false},
		},
	}
	for tcIndex, tc := range cases {
		multisigKey := multisig.NewPubKeyMultisigThreshold(tc.k, tc.pubkeys)
		multisignature := multisig.NewMultisig(len(tc.pubkeys))
		signBytesFn := func(mode signing.SignMode) ([]byte, error) { return tc.msg, nil }

		for i := 0; i < tc.k-1; i++ {
			signingIndex := tc.signingIndices[i]
			require.NoError(
				t,
				multisig.AddSignatureFromPubKey(multisignature, tc.signatures[signingIndex], tc.pubkeys[signingIndex], tc.pubkeys),
			)
			require.Error(
				t,
				multisigKey.VerifyMultisignature(signBytesFn, multisignature),
				"multisig passed when i < k, tc %d, i %d", tcIndex, i,
			)
			require.NoError(
				t,
				multisig.AddSignatureFromPubKey(multisignature, tc.signatures[signingIndex], tc.pubkeys[signingIndex], tc.pubkeys),
			)
			require.Equal(
				t,
				i+1,
				len(multisignature.Signatures),
				"adding a signature for the same pubkey twice increased signature count by 2, tc %d", tcIndex,
			)
		}
		require.Error(
			t,
			multisigKey.VerifyMultisignature(signBytesFn, multisignature),
			"multisig passed with k - 1 sigs, tc %d", tcIndex,
		)
		require.NoError(
			t,
			multisig.AddSignatureFromPubKey(
				multisignature,
				tc.signatures[tc.signingIndices[tc.k]],
				tc.pubkeys[tc.signingIndices[tc.k]],
				tc.pubkeys,
			),
		)
		require.NoError(
			t,
			multisigKey.VerifyMultisignature(signBytesFn, multisignature),
			"multisig failed after k good signatures, tc %d", tcIndex,
		)

		for i := tc.k + 1; i < len(tc.signingIndices); i++ {
			signingIndex := tc.signingIndices[i]

			require.NoError(
				t,
				multisig.AddSignatureFromPubKey(multisignature, tc.signatures[signingIndex], tc.pubkeys[signingIndex], tc.pubkeys),
			)
			require.Equal(
				t,
				tc.passAfterKSignatures[i-(tc.k)-1],
				multisigKey.VerifyMultisignature(func(mode signing.SignMode) ([]byte, error) {
					return tc.msg, nil
				}, multisignature),
				"multisig didn't verify as expected after k sigs, tc %d, i %d", tcIndex, i,
			)
			require.NoError(
				t,
				multisig.AddSignatureFromPubKey(multisignature, tc.signatures[signingIndex], tc.pubkeys[signingIndex], tc.pubkeys),
			)
			require.Equal(
				t,
				i+1,
				len(multisignature.Signatures),
				"adding a signature for the same pubkey twice increased signature count by 2, tc %d", tcIndex,
			)
		}
	}
}

// TODO: Fully replace this test with table driven tests
func TestThresholdMultisigDuplicateSignatures(t *testing.T) {
	msg := []byte{1, 2, 3, 4, 5}
	pubkeys, sigs := generatePubKeysAndSignatures(5, msg)
	multisigKey := multisig.NewPubKeyMultisigThreshold(2, pubkeys)
	multisignature := multisig.NewMultisig(5)
	signBytesFn := func(mode signing.SignMode) ([]byte, error) { return msg, nil }

	require.Error(t, multisigKey.VerifyMultisignature(signBytesFn, multisignature))
	multisig.AddSignatureFromPubKey(multisignature, sigs[0], pubkeys[0], pubkeys)
	// Add second signature manually
	multisignature.Signatures = append(multisignature.Signatures, sigs[0])
	require.Error(t, multisigKey.VerifyMultisignature(signBytesFn, multisignature))
}

func TestMultiSigPubKeyEquality(t *testing.T) {
	pubKey1 := secp256k1.GenPrivKey().PubKey()
	pubKey2 := secp256k1.GenPrivKey().PubKey()
	pubkeys := []crypto.PubKey{pubKey1, pubKey2}
	multisigKey := multisig.NewPubKeyMultisigThreshold(2, pubkeys)
	var other multisig.PubKey

	testCases := []struct {
		msg      string
		malleate func()
		expectEq bool
	}{
		{
			"equals",
			func() {
				var otherPubKey multisig.PubKeyMultisigThreshold
				multisig.Cdc.MustUnmarshalBinaryBare(multisigKey.Bytes(), &otherPubKey)
				other = otherPubKey
			},
			true,
		},
		{
			"ensure that reordering pubkeys is treated as a different pubkey",
			func() {
				pubkeysCpy := make([]crypto.PubKey, 2)
				copy(pubkeysCpy, pubkeys)
				pubkeysCpy[0] = pubkeys[1]
				pubkeysCpy[1] = pubkeys[0]
				other = multisig.NewPubKeyMultisigThreshold(2, pubkeysCpy)
			},
			false,
		},
		{
			"equals with proto pub key",
			func() {
				anyPubKeys := make([]*codectypes.Any, len(pubkeys))

				for i := 0; i < len(pubkeys); i++ {
					any, err := codectypes.NewAnyWithValue(pubkeys[i].(proto.Message))
					require.NoError(t, err)
					anyPubKeys[i] = any
				}
				other = &kmultisig.LegacyAminoMultisigThresholdPubKey{Threshold: 2, PubKeys: anyPubKeys}
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			tc.malleate()
			eq := multisigKey.Equals(other)
			require.Equal(t, eq, tc.expectEq)
		})
	}
}

func TestAddress(t *testing.T) {
	msg := []byte{1, 2, 3, 4}
	pubkeys, _ := generatePubKeysAndSignatures(5, msg)
	multisigKey := multisig.NewPubKeyMultisigThreshold(2, pubkeys)
	require.Len(t, multisigKey.Address().Bytes(), 20)
}

func TestPubKeyMultisigThresholdAminoToIface(t *testing.T) {
	msg := []byte{1, 2, 3, 4}
	pubkeys, _ := generatePubKeysAndSignatures(5, msg)
	multisigKey := multisig.NewPubKeyMultisigThreshold(2, pubkeys)

	ab, err := multisig.Cdc.MarshalBinaryLengthPrefixed(multisigKey)
	require.NoError(t, err)
	// like other crypto.Pubkey implementations (e.g. ed25519.PubKeyMultisigThreshold),
	// PubKeyMultisigThreshold should be deserializable into a crypto.PubKeyMultisigThreshold:
	var pubKey crypto.PubKey
	err = multisig.Cdc.UnmarshalBinaryLengthPrefixed(ab, &pubKey)
	require.NoError(t, err)

	require.Equal(t, multisigKey, pubKey)
}

func TestMultiSignature(t *testing.T) {
	msg := []byte{1, 2, 3, 4}
	pk, sig := generateNestedMultiSignature(3, msg)
	signBytesFn := func(mode signing.SignMode) ([]byte, error) { return msg, nil }
	err := pk.VerifyMultisignature(signBytesFn, sig)
	require.NoError(t, err)
}

func TestMultiSigMigration(t *testing.T) {
	msg := []byte{1, 2, 3, 4}
	pkSet, sigs := generatePubKeysAndSignatures(2, msg)
	multisignature := multisig.NewMultisig(2)

	multisigKey := multisig.NewPubKeyMultisigThreshold(2, pkSet)
	signBytesFn := func(mode signing.SignMode) ([]byte, error) { return msg, nil }

	cdc := codec.NewLegacyAmino()

	err := multisig.AddSignatureFromPubKey(multisignature, sigs[0], pkSet[0], pkSet)

	// create a StdSignature for msg, and convert it to sigV2
	sig := authtypes.StdSignature{PubKey: pkSet[1], Signature: msg}
	sigV2, err := authtypes.StdSignatureToSignatureV2(cdc, sig)
	require.NoError(t, multisig.AddSignatureV2(multisignature, sigV2, pkSet))

	require.NoError(t, err)
	require.NotNil(t, sigV2)

	require.NoError(t, multisigKey.VerifyMultisignature(signBytesFn, multisignature))
}

func TestAddSignatureFromPubKeyNilCheck(t *testing.T) {
	pkSet, sigs := generatePubKeysAndSignatures(5, []byte{1, 2, 3, 4})
	multisignature := multisig.NewMultisig(5)

	//verify no error is returned with all non-nil values
	err := multisig.AddSignatureFromPubKey(multisignature, sigs[0], pkSet[0], pkSet)
	require.NoError(t, err)
	//verify error is returned when key value is nil
	err = multisig.AddSignatureFromPubKey(multisignature, sigs[0], pkSet[0], nil)
	require.Error(t, err)
	//verify error is returned when pubkey value is nil
	err = multisig.AddSignatureFromPubKey(multisignature, sigs[0], nil, pkSet)
	require.Error(t, err)
	//verify error is returned when signature value is nil
	err = multisig.AddSignatureFromPubKey(multisignature, nil, pkSet[0], pkSet)
	require.Error(t, err)
	//verify error is returned when multisignature value is nil
	err = multisig.AddSignatureFromPubKey(nil, sigs[0], pkSet[0], pkSet)
	require.Error(t, err)
}

func generatePubKeysAndSignatures(n int, msg []byte) (pubkeys []crypto.PubKey, signatures []signing.SignatureData) {
	pubkeys = make([]crypto.PubKey, n)
	signatures = make([]signing.SignatureData, n)
	for i := 0; i < n; i++ {
		var privkey crypto.PrivKey
		switch rand.Int63() % 3 {
		case 0:
			privkey = ed25519.GenPrivKey()
		case 1:
			privkey = secp256k1.GenPrivKey()
		case 2:
			privkey = sr25519.GenPrivKey()
		}
		pubkeys[i] = privkey.PubKey()
		sig, _ := privkey.Sign(msg)
		signatures[i] = &signing.SingleSignatureData{Signature: sig}
	}
	return
}

func generateNestedMultiSignature(n int, msg []byte) (multisig.PubKey, *signing.MultiSignatureData) {
	pubkeys := make([]crypto.PubKey, n)
	signatures := make([]signing.SignatureData, n)
	bitArray := types.NewCompactBitArray(n)
	for i := 0; i < n; i++ {
		nestedPks, nestedSigs := generatePubKeysAndSignatures(5, msg)
		nestedBitArray := types.NewCompactBitArray(5)
		for j := 0; j < 5; j++ {
			nestedBitArray.SetIndex(j, true)
		}
		nestedSig := &signing.MultiSignatureData{
			BitArray:   nestedBitArray,
			Signatures: nestedSigs,
		}
		signatures[i] = nestedSig
		pubkeys[i] = multisig.NewPubKeyMultisigThreshold(5, nestedPks)
		bitArray.SetIndex(i, true)
	}
	return multisig.NewPubKeyMultisigThreshold(n, pubkeys), &signing.MultiSignatureData{
		BitArray:   bitArray,
		Signatures: signatures,
	}
}
