package xsalsa20symmetric

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/nacl/secretbox"
)

// TODO, make this into a struct that implements crypto.Symmetric.

const (
	nonceLen  = 24
	secretLen = 32
)

var ErrCiphertextDecrypt = errors.New("ciphertext decryption failed")

// EncryptSymmetric encrypts the given plaintext with the given secret using XSalsa20-Poly1305.
// The secret must be exactly 32 bytes long. For secure key derivation, use
// something like Sha256(Bcrypt(passphrase)). The returned ciphertext includes
// a random 24-byte nonce and is (secretbox.Overhead + 24) bytes longer than the plaintext.
func EncryptSymmetric(plaintext, secret []byte) (ciphertext []byte) {
	if len(secret) != secretLen {
		panic(fmt.Sprintf("Secret must be 32 bytes long, got len %v", len(secret)))
	}
	nonce := randBytes(nonceLen)
	nonceArr := [nonceLen]byte{}
	copy(nonceArr[:], nonce)
	secretArr := [secretLen]byte{}
	copy(secretArr[:], secret)
	ciphertext = make([]byte, nonceLen+secretbox.Overhead+len(plaintext))
	copy(ciphertext, nonce)
	secretbox.Seal(ciphertext[nonceLen:nonceLen], plaintext, &nonceArr, &secretArr)
	return ciphertext
}

// DecryptSymmetric decrypts the given ciphertext with the given secret using XSalsa20-Poly1305.
// The secret must be exactly 32 bytes long and must match the one used for encryption.
// Returns the original plaintext or an error if decryption fails (e.g., wrong key,
// corrupted ciphertext, or insufficient length).
func DecryptSymmetric(ciphertext, secret []byte) (plaintext []byte, err error) {
	if len(secret) != secretLen {
		panic(fmt.Sprintf("Secret must be 32 bytes long, got len %v", len(secret)))
	}
	if len(ciphertext) <= secretbox.Overhead+nonceLen {
		return nil, errors.New("ciphertext is too short")
	}
	nonce := ciphertext[:nonceLen]
	nonceArr := [nonceLen]byte{}
	copy(nonceArr[:], nonce)
	secretArr := [secretLen]byte{}
	copy(secretArr[:], secret)
	plaintext = make([]byte, len(ciphertext)-nonceLen-secretbox.Overhead)
	_, ok := secretbox.Open(plaintext[:0], ciphertext[nonceLen:], &nonceArr, &secretArr)
	if !ok {
		return nil, ErrCiphertextDecrypt
	}
	return plaintext, nil
}

// randBytes generates cryptographically secure random bytes using the OS's randomness source.
// Panics if the underlying crypto/rand.Read fails, which should only happen in
// exceptional circumstances (e.g., system entropy exhaustion).
func randBytes(numBytes int) []byte {
	b := make([]byte, numBytes)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
