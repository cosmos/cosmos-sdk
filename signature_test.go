package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestSignAndValidateEd25519(t *testing.T) {

	privKey := GenPrivKeyEd25519()
	pubKey, err := privKey.PubKey()
	require.Nil(t, err)

	msg := CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	sigEd := sig.(SignatureEd25519)
	sigEd[7] ^= byte(0x01)
	sig = sigEd

	assert.False(t, pubKey.VerifyBytes(msg, sig))
}

func TestSignAndValidateSecp256k1(t *testing.T) {
	privKey := GenPrivKeySecp256k1()
	pubKey, err := privKey.PubKey()
	require.Nil(t, err)

	msg := CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	assert.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	sigEd := sig.(SignatureSecp256k1)
	sigEd[3] ^= byte(0x01)
	sig = sigEd

	assert.False(t, pubKey.VerifyBytes(msg, sig))
}
