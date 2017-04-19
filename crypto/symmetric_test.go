package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/crypto/bcrypt"
)

func TestSimple(t *testing.T) {

	MixEntropy([]byte("someentropy"))

	plaintext := []byte("sometext")
	secret := []byte("somesecretoflengththirtytwo===32")
	ciphertext := EncryptSymmetric(plaintext, secret)
	plaintext2, err := DecryptSymmetric(ciphertext, secret)

	require.Nil(t, err, "%+v", err)
	assert.Equal(t, plaintext, plaintext2)
}

func TestSimpleWithKDF(t *testing.T) {

	MixEntropy([]byte("someentropy"))

	plaintext := []byte("sometext")
	secretPass := []byte("somesecret")
	secret, err := bcrypt.GenerateFromPassword(secretPass, 12)
	if err != nil {
		t.Error(err)
	}
	secret = Sha256(secret)

	ciphertext := EncryptSymmetric(plaintext, secret)
	plaintext2, err := DecryptSymmetric(ciphertext, secret)

	require.Nil(t, err, "%+v", err)
	assert.Equal(t, plaintext, plaintext2)
}
