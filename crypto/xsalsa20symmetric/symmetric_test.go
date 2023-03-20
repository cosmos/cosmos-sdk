package xsalsa20symmetric

import (
	"testing"

	"github.com/cometbft/cometbft/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"
)

func TestSimple(t *testing.T) {
	plaintext := []byte("sometext")
	secret := []byte("somesecretoflengththirtytwo===32")
	ciphertext := EncryptSymmetric(plaintext, secret)
	plaintext2, err := DecryptSymmetric(ciphertext, secret)

	require.NoError(t, err, "%+v", err)
	assert.Equal(t, plaintext, plaintext2)
}

func TestSimpleWithKDF(t *testing.T) {
	plaintext := []byte("sometext")
	secretPass := []byte("somesecret")
	saltBytes := crypto.CRandBytes(16)
	secret := argon2.IDKey(secretPass, saltBytes, 1, 64*1024, 4, 32)

	ciphertext := EncryptSymmetric(plaintext, secret)
	plaintext2, err := DecryptSymmetric(ciphertext, secret)

	require.NoError(t, err, "%+v", err)
	assert.Equal(t, plaintext, plaintext2)
}
