package keys

import (
	"bytes"
	"crypto/subtle"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/sr25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSr25519Compatibility(t *testing.T) {
	tmPrivKey := sr25519.GenPrivKey()
	tmPubKey := tmPrivKey.PubKey()

	privKey, err := GenPrivKey(SR25519)
	require.NoError(t, err)
	pubKey := privKey.PubKey()

	require.Equal(t, sdk.AccAddress(tmPubKey.Address()).String(), sdk.AccAddress(pubKey.Address()).String())
	require.True(t, bytes.Equal(tmPubKey.Bytes()[:], pubKey.Bytes()))
	require.True(t, subtle.ConstantTimeCompare(privKey.Bytes()[:], privKey.Bytes()) == 1)
}

func TestSignAndValidateSr25519(t *testing.T) {
	privKey, err := GenPrivKey(SR25519)
	require.NoError(t, err)
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.NoError(t, err)

	// Test the signature
	require.True(t, pubKey.VerifyBytes(msg, sig))
	require.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	// TODO: Replace this with a much better fuzzer, tendermint/ed25519/issues/10
	sig[7] ^= byte(0x01)

	require.False(t, pubKey.VerifyBytes(msg, sig))
}
