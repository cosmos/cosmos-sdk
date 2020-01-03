package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestGetPubKeySecp256k1(t *testing.T) {
	tmpk := (secp256k1.GenPrivKey().PubKey()).(secp256k1.PubKeySecp256k1)
	testCases := []struct {
		input    *types.PublicKey
		expected secp256k1.PubKeySecp256k1
		err      error
	}{
		{
			&types.PublicKey{&types.PublicKey_Secp256K1{tmpk[:]}},
			tmpk,
			nil,
		},
		{
			&types.PublicKey{&types.PublicKey_Secp256K1{append(tmpk[:], tmpk[:]...)}},
			secp256k1.PubKeySecp256k1{},
			types.ErrInvalidPubKeySecp256k1Length,
		},
		{
			&types.PublicKey{&types.PublicKey_Ed25519{}},
			secp256k1.PubKeySecp256k1{},
			types.ErrInvalidPubKeySecp256k1,
		},
	}

	for _, tc := range testCases {
		pk, err := tc.input.GetPubKeySecp256k1()
		require.Equal(t, tc.err, err)
		require.True(t, pk.Equals(tc.expected))
	}
}

func TestGetPubKeyEd25519(t *testing.T) {
	tmpk := (ed25519.GenPrivKey().PubKey()).(ed25519.PubKeyEd25519)
	testCases := []struct {
		input    *types.PublicKey
		expected ed25519.PubKeyEd25519
		err      error
	}{
		{
			&types.PublicKey{&types.PublicKey_Ed25519{tmpk[:]}},
			tmpk,
			nil,
		},
		{
			&types.PublicKey{&types.PublicKey_Ed25519{append(tmpk[:], tmpk[:]...)}},
			ed25519.PubKeyEd25519{},
			types.ErrInvalidPubKeyEd25519Length,
		},
		{
			&types.PublicKey{&types.PublicKey_Secp256K1{}},
			ed25519.PubKeyEd25519{},
			types.ErrInvalidPubKeyEd25519,
		},
	}

	for _, tc := range testCases {
		pk, err := tc.input.GetPubKeyEd25519()
		require.Equal(t, tc.err, err)
		require.True(t, pk.Equals(tc.expected))
	}
}

func TestAddress(t *testing.T) {
	ed25519pk := (ed25519.GenPrivKey().PubKey()).(ed25519.PubKeyEd25519)
	secp256k1pk := (secp256k1.GenPrivKey().PubKey()).(secp256k1.PubKeySecp256k1)
	secp256k1pk2 := (secp256k1.GenPrivKey().PubKey()).(secp256k1.PubKeySecp256k1)
	multisigpk := multisig.NewPubKeyMultisigThreshold(2, []tmcrypto.PubKey{secp256k1pk, secp256k1pk2})

	testCases := []struct {
		pubkey tmcrypto.PubKey
		input  *types.PublicKey
	}{
		{
			ed25519pk, &types.PublicKey{&types.PublicKey_Ed25519{ed25519pk[:]}},
		},
		{
			secp256k1pk, &types.PublicKey{&types.PublicKey_Secp256K1{secp256k1pk[:]}},
		},
		{
			multisigpk,
			&types.PublicKey{
				&types.PublicKey_Multisig{
					Multisig: &types.MultiSig{
						K: 2,
						Pubkeys: []*types.PublicKey{
							{&types.PublicKey_Secp256K1{secp256k1pk[:]}},
							{&types.PublicKey_Secp256K1{secp256k1pk2[:]}},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.pubkey.Address(), tc.input.Address())
	}
}
