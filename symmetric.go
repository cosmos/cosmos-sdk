package crypto

import (
	"errors"

	. "github.com/tendermint/go-common"
	"golang.org/x/crypto/nacl/secretbox"
)

const nonceLen = 24
const secretLen = 32

// secret must be 32 bytes long. Use something like Sha256(Bcrypt(passphrase))
// The ciphertext is (secretbox.Overhead + 24) bytes longer than the plaintext.
// NOTE: call crypto.MixEntropy() first.
func EncryptSymmetric(plaintext []byte, secret []byte) (ciphertext []byte) {
	if len(secret) != secretLen {
		PanicSanity(Fmt("Secret must be 32 bytes long, got len %v", len(secret)))
	}
	nonce := CRandBytes(nonceLen)
	nonceArr := [nonceLen]byte{}
	copy(nonceArr[:], nonce)
	secretArr := [secretLen]byte{}
	copy(secretArr[:], secret)
	ciphertext = make([]byte, nonceLen+secretbox.Overhead+len(plaintext))
	copy(ciphertext, nonce)
	secretbox.Seal(ciphertext[nonceLen:nonceLen], plaintext, &nonceArr, &secretArr)
	return ciphertext
}

// secret must be 32 bytes long. Use something like Sha256(Bcrypt(passphrase))
// The ciphertext is (secretbox.Overhead + 24) bytes longer than the plaintext.
func DecryptSymmetric(ciphertext []byte, secret []byte) (plaintext []byte, err error) {
	if len(secret) != secretLen {
		PanicSanity(Fmt("Secret must be 32 bytes long, got len %v", len(secret)))
	}
	if len(ciphertext) <= secretbox.Overhead+nonceLen {
		return nil, errors.New("Ciphertext is too short")
	}
	nonce := ciphertext[:nonceLen]
	nonceArr := [nonceLen]byte{}
	copy(nonceArr[:], nonce)
	secretArr := [secretLen]byte{}
	copy(secretArr[:], secret)
	plaintext = make([]byte, len(ciphertext)-nonceLen-secretbox.Overhead)
	_, ok := secretbox.Open(plaintext[:0], ciphertext[nonceLen:], &nonceArr, &secretArr)
	if !ok {
		return nil, errors.New("Ciphertext decryption failed")
	}
	return plaintext, nil
}
