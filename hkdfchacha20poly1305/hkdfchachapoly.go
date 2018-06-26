// Package hkdfchacha20poly1305 creates an AEAD using hkdf, chacha20, and poly1305
// When sealing and opening, the hkdf is used to obtain the nonce and subkey for
// chacha20. Other than the change for the how the subkey and nonce for chacha
// are obtained, this is the same as chacha20poly1305
package hkdfchacha20poly1305

import (
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

type hkdfchacha20poly1305 struct {
	key [KeySize]byte
}

const (
	// KeySize is the size of the key used by this AEAD, in bytes.
	KeySize = 32
	// NonceSize is the size of the nonce used with this AEAD, in bytes.
	NonceSize = 24
	// TagSize is the size added from poly1305
	TagSize = 16
	// MaxPlaintextSize is the max size that can be passed into a single call of Seal
	MaxPlaintextSize = (1 << 38) - 64
	// MaxCiphertextSize is the max size that can be passed into a single call of Open,
	// this differs from plaintext size due to the tag
	MaxCiphertextSize = (1 << 38) - 48
	// HkdfInfo is the parameter used internally for Hkdf's info parameter.
	HkdfInfo = "TENDERMINT_SECRET_CONNECTION_FRAME_KEY_DERIVE"
)

//New xChaChapoly1305 AEAD with 24 byte nonces
func New(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize {
		return nil, errors.New("chacha20poly1305: bad key length")
	}
	ret := new(hkdfchacha20poly1305)
	copy(ret.key[:], key)
	return ret, nil

}
func (c *hkdfchacha20poly1305) NonceSize() int {
	return NonceSize
}

func (c *hkdfchacha20poly1305) Overhead() int {
	return TagSize
}

func (c *hkdfchacha20poly1305) Seal(dst, nonce, plaintext, additionalData []byte) []byte {
	if len(nonce) != NonceSize {
		panic("hkdfchacha20poly1305: bad nonce length passed to Seal")
	}

	if uint64(len(plaintext)) > MaxPlaintextSize {
		panic("hkdfchacha20poly1305: plaintext too large")
	}

	subKey, chachaNonce := getSubkeyAndChachaNonceFromHkdf(&c.key, &nonce)

	aead, err := chacha20poly1305.New(subKey[:])
	if err != nil {
		panic("hkdfchacha20poly1305: failed to initialize chacha20poly1305")
	}

	return aead.Seal(dst, chachaNonce[:], plaintext, additionalData)
}

func (c *hkdfchacha20poly1305) Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	if len(nonce) != NonceSize {
		return nil, errors.New("hkdfchacha20poly1305: bad nonce length passed to Open")
	}
	if uint64(len(ciphertext)) > MaxCiphertextSize {
		return nil, errors.New("hkdfchacha20poly1305: ciphertext too large")
	}

	subKey, chachaNonce := getSubkeyAndChachaNonceFromHkdf(&c.key, &nonce)

	aead, err := chacha20poly1305.New(subKey[:])
	if err != nil {
		panic("hkdfchacha20poly1305: failed to initialize chacha20poly1305")
	}

	return aead.Open(dst, chachaNonce[:], ciphertext, additionalData)
}

func getSubkeyAndChachaNonceFromHkdf(cKey *[32]byte, nonce *[]byte) (
	subKey [KeySize]byte, chachaNonce [chacha20poly1305.NonceSize]byte) {
	hash := sha256.New
	hkdf := hkdf.New(hash, (*cKey)[:], *nonce, []byte(HkdfInfo))
	_, err := io.ReadFull(hkdf, subKey[:])
	if err != nil {
		panic("hkdfchacha20poly1305: failed to read subkey from hkdf")
	}
	_, err = io.ReadFull(hkdf, chachaNonce[:])
	if err != nil {
		panic("hkdfchacha20poly1305: failed to read chachaNonce from hkdf")
	}
	return
}
