package crypto

import (
	"bytes"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestSimple(t *testing.T) {

	MixEntropy([]byte("someentropy"))

	plaintext := []byte("sometext")
	secret := []byte("somesecretoflengththirtytwo===32")
	ciphertext := EncryptSymmetric(plaintext, secret)

	plaintext2, err := DecryptSymmetric(ciphertext, secret)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(plaintext, plaintext2) {
		t.Errorf("Decrypted plaintext was %X, expected %X", plaintext2, plaintext)
	}

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
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(plaintext, plaintext2) {
		t.Errorf("Decrypted plaintext was %X, expected %X", plaintext2, plaintext)
	}

}
