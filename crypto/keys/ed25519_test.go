package keys

import (
	"bytes"
	"crypto/subtle"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestSignAndValidateEd25519(t *testing.T) {

	privKey, err := GenPrivKey(ED25519)
	require.NoError(t, err)
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.NoError(t, err)

	// Test the signature
	require.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	// TODO: Replace this with a much better fuzzer, tendermint/ed25519/issues/10
	sig[7] ^= byte(0x01)

	require.False(t, pubKey.VerifyBytes(msg, sig))
}

func TestEd25519Compatibility(t *testing.T) {
	tmPrivKey := ed25519.GenPrivKey()
	tmPubKey := tmPrivKey.PubKey()

	privKey, err := GenPrivKey(ED25519)
	require.NoError(t, err)
	pubKey := privKey.PubKey()

	require.Equal(t, sdk.AccAddress(tmPubKey.Address()).String(), sdk.AccAddress(pubKey.Address()).String())
	require.True(t, bytes.Equal(tmPubKey.Bytes()[:], pubKey.Bytes()))
	require.True(t, subtle.ConstantTimeCompare(privKey.Bytes()[:], privKey.Bytes()) == 1)
}
