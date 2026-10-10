package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/address"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1eth"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func rekeyAddrCodec() address.Codec { return addresscodec.NewBech32Codec("cosmos") }

func genMlDsa65(t *testing.T) (*mldsa65.PrivKey, cryptotypes.PubKey) {
	t.Helper()
	sk, err := mldsa65.GenPrivKey()
	require.NoError(t, err)
	return &sk, sk.PubKey()
}

func proofDoc(t *testing.T, chainID string, accNum uint64, addr string, newPk cryptotypes.PubKey) types.ChangePubKeyProofDoc {
	t.Helper()
	anyPk, err := codectypes.NewAnyWithValue(newPk)
	require.NoError(t, err)
	return types.ChangePubKeyProofDoc{ChainId: chainID, AccountNumber: accNum, Address: addr, NewPubKey: anyPk}
}

func signBytesFor(t *testing.T, doc types.ChangePubKeyProofDoc, newPk cryptotypes.PubKey) []byte {
	t.Helper()
	bz, err := types.ChangePubKeyProofSignBytes(rekeyAddrCodec(), doc, newPk)
	require.NoError(t, err)
	return bz
}

func singleSig(t *testing.T, sk cryptotypes.PrivKey, msg []byte) *signing.SingleSignatureData {
	t.Helper()
	sig, err := sk.Sign(msg)
	require.NoError(t, err)
	return &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, Signature: sig}
}

func TestValidateRekeyPubKey_AcceptsSupportedTypes(t *testing.T) {
	params := types.DefaultParams()
	current := secp256k1.GenPrivKey().PubKey()

	_, mlPk := genMlDsa65(t)
	_, mlPk2 := genMlDsa65(t)
	r1, err := secp256r1.GenPrivKey()
	require.NoError(t, err)

	mixed := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{
		secp256k1.GenPrivKey().PubKey(), mlPk, mlPk2,
	})

	for name, pk := range map[string]cryptotypes.PubKey{
		"secp256k1":      secp256k1.GenPrivKey().PubKey(),
		"mldsa65":        mlPk,
		"ed25519":        ed25519.GenPrivKey().PubKey(),
		"secp256r1":      r1.PubKey(),
		"mixed multisig": mixed,
		"multisig with ed25519 subkey": kmultisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{
			ed25519.GenPrivKey().PubKey(), secp256k1.GenPrivKey().PubKey(),
		}),
	} {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, types.ValidateRekeyPubKey(current, pk, params))
			// The ordering contract: a validated key never panics in Address().
			require.NotPanics(t, func() { pk.Address() })
		})
	}
}

func TestValidateRekeyPubKey_Rejects(t *testing.T) {
	params := types.DefaultParams()
	params.TxSigLimit = 7
	current := secp256k1.GenPrivKey().PubKey()

	modCred, err := types.NewModuleCredential("gov", []byte("x"))
	require.NoError(t, err)

	genKeys := func(n int) []cryptotypes.PubKey {
		keys := make([]cryptotypes.PubKey, n)
		for i := range keys {
			keys[i] = secp256k1.GenPrivKey().PubKey()
		}
		return keys
	}

	nested := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{
		kmultisig.NewLegacyAminoPubKey(2, genKeys(4)),
		kmultisig.NewLegacyAminoPubKey(2, genKeys(4)),
	})

	// kmultisig.AminoCdc does not register secp256r1, so a multisig holding
	// one panics in Address().
	r1, err := secp256r1.GenPrivKey()
	require.NoError(t, err)
	r1Msig := kmultisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{r1.PubKey(), current})

	cases := map[string]cryptotypes.PubKey{
		"nil":                            nil,
		"multisig with secp256r1 subkey": r1Msig,
		"nested multisig with secp256r1 subkey": kmultisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{
			r1Msig, secp256k1.GenPrivKey().PubKey(),
		}),
		"module credential":         modCred,
		"secp256k1eth":              secp256k1eth.GenPrivKey().PubKey(),
		"equals current":            &secp256k1.PubKey{Key: current.Bytes()},
		"multisig over sig limit":   kmultisig.NewLegacyAminoPubKey(2, genKeys(8)),
		"nested multisig over cap":  nested,
		"multisig with eth subkey":  kmultisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{secp256k1eth.GenPrivKey().PubKey(), current}),
		"multisig zero threshold":   &kmultisig.LegacyAminoPubKey{Threshold: 0, PubKeys: kmultisig.NewLegacyAminoPubKey(1, genKeys(2)).PubKeys},
		"multisig unmeetable":       &kmultisig.LegacyAminoPubKey{Threshold: 3, PubKeys: kmultisig.NewLegacyAminoPubKey(1, genKeys(2)).PubKeys},
		"multisig empty":            &kmultisig.LegacyAminoPubKey{Threshold: 1},
		"secp256k1 wrong length":    &secp256k1.PubKey{Key: []byte{1, 2, 3}},
		"ed25519 wrong length":      &ed25519.PubKey{Key: []byte{1, 2, 3}},
		"mldsa65 wrong length":      &mldsa65.PubKey{Key: []byte{1, 2, 3}},
		"secp256r1 without key":     &secp256r1.PubKey{},
		"typed nil secp256k1 value": (*secp256k1.PubKey)(nil),
		"multisig with unpacked any": &kmultisig.LegacyAminoPubKey{Threshold: 1, PubKeys: []*codectypes.Any{
			{TypeUrl: "/cosmos.crypto.secp256k1.PubKey", Value: []byte{0x0a, 0x01, 0x01}},
		}},
		"multisig with nil any": &kmultisig.LegacyAminoPubKey{Threshold: 1, PubKeys: []*codectypes.Any{nil}},
	}
	for name, pk := range cases {
		t.Run(name, func(t *testing.T) {
			err := types.ValidateRekeyPubKey(current, pk, params)
			require.ErrorIs(t, err, types.ErrInvalidNewPubKey)
		})
	}
}

// TestValidateRekeyPubKey_SignatureTreeLimits checks that a new pubkey whose
// signature tree the ante handler cannot flatten is rejected, so an account
// cannot rekey itself into a key it can never sign with.
func TestValidateRekeyPubKey_SignatureTreeLimits(t *testing.T) {
	current := secp256k1.GenPrivKey().PubKey()
	k := func() cryptotypes.PubKey { return secp256k1.GenPrivKey().PubKey() }
	msig := func(threshold int, keys ...cryptotypes.PubKey) cryptotypes.PubKey {
		return kmultisig.NewLegacyAminoPubKey(threshold, keys)
	}

	params := types.DefaultParams()

	// Two multisig levels: the leaf signatures sit at the deepest level the
	// ante handler flattens.
	require.NoError(t, types.ValidateRekeyPubKey(current, msig(2, msig(1, k(), k()), k()), params))

	// Three multisig levels with a single leaf: only the depth limit rejects it.
	err := types.ValidateRekeyPubKey(current, msig(1, msig(1, msig(1, k()))), params)
	require.ErrorIs(t, err, types.ErrInvalidNewPubKey)

	// A flat multisig wider than the per-level limit: only the breadth limit
	// rejects it.
	wide := make([]cryptotypes.PubKey, types.MaxSignatureTreeBreadth+1)
	for i := range wide {
		wide[i] = k()
	}
	params.TxSigLimit = uint64(len(wide)) + 1
	err = types.ValidateRekeyPubKey(current, msig(1, wide...), params)
	require.ErrorIs(t, err, types.ErrInvalidNewPubKey)
	require.NoError(t, types.ValidateRekeyPubKey(current, msig(1, wide[1:]...), params))
}

func TestValidateRekeyPubKey_NilCurrent(t *testing.T) {
	_, mlPk := genMlDsa65(t)
	require.NoError(t, types.ValidateRekeyPubKey(nil, mlPk, types.DefaultParams()))
}

func TestVerifyChangePubKeyProof_Single(t *testing.T) {
	addrA := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"

	skK1 := secp256k1.GenPrivKey()
	skMl, _ := genMlDsa65(t)

	for name, sk := range map[string]cryptotypes.PrivKey{"secp256k1": skK1, "mldsa65": skMl} {
		t.Run(name, func(t *testing.T) {
			pk := sk.PubKey()
			msg := signBytesFor(t, proofDoc(t, "a", 5, addrA, pk), pk)
			proof := singleSig(t, sk, msg)
			require.NoError(t, types.VerifyChangePubKeyProof(pk, msg, proof))

			flipped := append([]byte(nil), proof.Signature...)
			flipped[0] ^= 0x01
			bad := &signing.SingleSignatureData{SignMode: proof.SignMode, Signature: flipped}
			require.ErrorIs(t, types.VerifyChangePubKeyProof(pk, msg, bad), types.ErrInvalidPubKeyProof)

			// wrong sign mode is rejected even with a valid signature
			wrongMode := &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT, Signature: proof.Signature}
			require.ErrorIs(t, types.VerifyChangePubKeyProof(pk, msg, wrongMode), types.ErrInvalidPubKeyProof)

			// a multisig proof for a single key is rejected
			require.ErrorIs(t, types.VerifyChangePubKeyProof(pk, msg, multisig.NewMultisig(1)), types.ErrInvalidPubKeyProof)
			// nil proof is rejected
			require.ErrorIs(t, types.VerifyChangePubKeyProof(pk, msg, nil), types.ErrInvalidPubKeyProof)
		})
	}
}

func TestVerifyChangePubKeyProof_Multisig(t *testing.T) {
	addrA := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"

	skK1 := secp256k1.GenPrivKey()
	skMl1, _ := genMlDsa65(t)
	skMl2, _ := genMlDsa65(t)
	sks := []cryptotypes.PrivKey{skK1, skMl1, skMl2}
	pks := []cryptotypes.PubKey{skK1.PubKey(), skMl1.PubKey(), skMl2.PubKey()}
	msigPk := kmultisig.NewLegacyAminoPubKey(2, pks)

	msg := signBytesFor(t, proofDoc(t, "a", 5, addrA, msigPk), msigPk)

	build := func(idx ...int) *signing.MultiSignatureData {
		m := multisig.NewMultisig(len(pks))
		for _, i := range idx {
			require.NoError(t, multisig.AddSignatureFromPubKey(m, singleSig(t, sks[i], msg), pks[i], pks))
		}
		return m
	}

	require.NoError(t, types.VerifyChangePubKeyProof(msigPk, msg, build(0, 2)))
	require.NoError(t, types.VerifyChangePubKeyProof(msigPk, msg, build(1, 2)))
	require.ErrorIs(t, types.VerifyChangePubKeyProof(msigPk, msg, build(1)), types.ErrInvalidPubKeyProof)

	// a single signature for a multisig key is rejected
	require.ErrorIs(t, types.VerifyChangePubKeyProof(msigPk, msg, singleSig(t, skK1, msg)), types.ErrInvalidPubKeyProof)
}

func TestVerifyChangePubKeyProof_Replay(t *testing.T) {
	addrA := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	addrB := "cosmos1qgpqyqszqgpqyqszqgpqyqszqgpqyqszrh8mx2"

	sk := secp256k1.GenPrivKey()
	pk := sk.PubKey()

	msgA := signBytesFor(t, proofDoc(t, "a", 5, addrA, pk), pk)
	proof := singleSig(t, sk, msgA)
	require.NoError(t, types.VerifyChangePubKeyProof(pk, msgA, proof))

	for name, doc := range map[string]types.ChangePubKeyProofDoc{
		"other chain":   proofDoc(t, "b", 5, addrA, pk),
		"other account": proofDoc(t, "a", 6, addrA, pk),
		"other address": proofDoc(t, "a", 5, addrB, pk),
	} {
		t.Run(name, func(t *testing.T) {
			msg := signBytesFor(t, doc, pk)
			require.ErrorIs(t, types.VerifyChangePubKeyProof(pk, msg, proof), types.ErrInvalidPubKeyProof)
		})
	}
}

func TestChangePubKeyProofSignBytes_Golden(t *testing.T) {
	pk := secp256k1.GenPrivKeyFromSecret([]byte("rekey-golden")).PubKey()
	doc := proofDoc(t, "test-chain", 5, "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu", pk)

	bz := signBytesFor(t, doc, pk)
	require.Equal(t, goldenChangePubKeyProofSignBytes, hex.EncodeToString(bz), "sign bytes: %s", bz)
}

// goldenChangePubKeyProofSignBytes is the hex of:
// {"account_number":"0","chain_id":"","fee":{"amount":[],"gas":"0"},"memo":"","msgs":[{"type":"sign/MsgSignData","value":{"data":"<base64 proto(doc)>","signer":"cosmos1k9t0vy03agccf9wacyh9ygehuegrfph8jjzc3p"}}],"sequence":"0"}
const goldenChangePubKeyProofSignBytes = "7b226163636f756e745f6e756d626572223a2230222c22636861696e5f6964223a22222c22666565223a7b22616d6f756e74223a5b5d2c22676173223a2230227d2c226d656d6f223a22222c226d736773223a5b7b2274797065223a227369676e2f4d73675369676e44617461222c2276616c7565223a7b2264617461223a22436770305a584e304c574e6f59576c75454155614c574e7663323176637a467865584278654842784f58466a636e4e7a656d637963485a3463545a79637a42366357637a65586c6a4e577836646a643464534a47436838765932397a6257397a4c6d4e79655842306279357a5a574e774d6a5532617a4575554856695332563545694d4b49514f666a69707566526342474a616f50726147504b6f34387076763061585133743348646830784934345565413d3d222c227369676e6572223a22636f736d6f73316b397430767930336167636366397761637968397967656875656772667068386a6a7a633370227d7d5d2c2273657175656e6365223a2230227d"

func TestDecodeChangePubKeyProof(t *testing.T) {
	single := &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, Signature: []byte("sig")}
	bz, err := signing.SignatureDataToProto(single).Marshal()
	require.NoError(t, err)
	got, err := types.DecodeChangePubKeyProof(bz)
	require.NoError(t, err)
	require.Equal(t, single, got)

	pks := []cryptotypes.PubKey{secp256k1.GenPrivKey().PubKey(), secp256k1.GenPrivKey().PubKey()}
	multi := multisig.NewMultisig(2)
	require.NoError(t, multisig.AddSignatureFromPubKey(multi, single, pks[1], pks))
	bz, err = signing.SignatureDataToProto(multi).Marshal()
	require.NoError(t, err)
	got, err = types.DecodeChangePubKeyProof(bz)
	require.NoError(t, err)
	require.Equal(t, multi, got)

	for name, bad := range map[string][]byte{
		"garbage": {0xff, 0xff, 0xff},
		"empty":   {},
		// a multi whose only nested signature has no sum set
		"multi with empty nested": mustMarshalProto(t, &signing.SignatureDescriptor_Data{
			Sum: &signing.SignatureDescriptor_Data_Multi_{Multi: &signing.SignatureDescriptor_Data_Multi{
				Bitarray:   multi.BitArray,
				Signatures: []*signing.SignatureDescriptor_Data{{}},
			}},
		}),
		// a multi with no bit array
		"multi without bitarray": mustMarshalProto(t, &signing.SignatureDescriptor_Data{
			Sum: &signing.SignatureDescriptor_Data_Multi_{Multi: &signing.SignatureDescriptor_Data_Multi{}},
		}),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := types.DecodeChangePubKeyProof(bad)
			require.ErrorIs(t, err, types.ErrInvalidPubKeyProof)
		})
	}
}

func mustMarshalProto(t *testing.T, m *signing.SignatureDescriptor_Data) []byte {
	t.Helper()
	bz, err := m.Marshal()
	require.NoError(t, err)
	return bz
}

// nestedMultiProof returns a proof with levels nested MultiSignatureData
// levels around a single signature.
func nestedMultiProof(t *testing.T, levels int) []byte {
	t.Helper()
	data := &signing.SignatureDescriptor_Data{
		Sum: &signing.SignatureDescriptor_Data_Single_{Single: &signing.SignatureDescriptor_Data_Single{
			Mode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, Signature: []byte("sig"),
		}},
	}
	for range levels {
		data = &signing.SignatureDescriptor_Data{
			Sum: &signing.SignatureDescriptor_Data_Multi_{Multi: &signing.SignatureDescriptor_Data_Multi{
				Bitarray:   multisig.NewMultisig(1).BitArray,
				Signatures: []*signing.SignatureDescriptor_Data{data},
			}},
		}
	}
	return mustMarshalProto(t, data)
}

func TestDecodeChangePubKeyProof_NestingDepth(t *testing.T) {
	_, err := types.DecodeChangePubKeyProof(nestedMultiProof(t, types.MaxChangePubKeyProofDepth))
	require.NoError(t, err)

	_, err = types.DecodeChangePubKeyProof(nestedMultiProof(t, types.MaxChangePubKeyProofDepth+1))
	require.ErrorIs(t, err, types.ErrInvalidPubKeyProof)

	_, err = types.DecodeChangePubKeyProof(nestedMultiProof(t, 1000))
	require.ErrorIs(t, err, types.ErrInvalidPubKeyProof)
}
