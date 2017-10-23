package cryptostore

import (
	"github.com/pkg/errors"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/bcrypt"
)

var (
	// SecretBox uses the algorithm from NaCL to store secrets securely
	SecretBox Encoder = secretbox{}
	// Noop doesn't do any encryption, should only be used in test code
	Noop Encoder = noop{}
)

// Encoder is used to encrypt any key with a passphrase for storage.
//
// This should use a well-designed symetric encryption algorithm
type Encoder interface {
	Encrypt(privKey crypto.PrivKey, passphrase string) (saltBytes []byte, encBytes []byte, err error)
	Decrypt(saltBytes []byte, encBytes []byte, passphrase string) (privKey crypto.PrivKey, err error)
}

type secretbox struct{}

func (e secretbox) Encrypt(privKey crypto.PrivKey, passphrase string) (saltBytes []byte, encBytes []byte, err error) {
	if passphrase == "" {
		return nil, privKey.Bytes(), nil
	}

	saltBytes = crypto.CRandBytes(16)
	key, err := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), 14) // TODO parameterize.  14 is good today (2016)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Couldn't generate bcrypt key from passphrase.")
	}
	key = crypto.Sha256(key) // Get 32 bytes
	privKeyBytes := privKey.Bytes()
	return saltBytes, crypto.EncryptSymmetric(privKeyBytes, key), nil
}

func (e secretbox) Decrypt(saltBytes []byte, encBytes []byte, passphrase string) (privKey crypto.PrivKey, err error) {
	privKeyBytes := encBytes
	// NOTE: Some keys weren't encrypted with a passphrase and hence we have the conditional
	if passphrase != "" {
		key, err := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), 14) // TODO parameterize.  14 is good today (2016)
		if err != nil {
			return crypto.PrivKey{}, errors.Wrap(err, "Invalid Passphrase")
		}
		key = crypto.Sha256(key) // Get 32 bytes
		privKeyBytes, err = crypto.DecryptSymmetric(encBytes, key)
		if err != nil {
			return crypto.PrivKey{}, errors.Wrap(err, "Invalid Passphrase")
		}
	}
	privKey, err = crypto.PrivKeyFromBytes(privKeyBytes)
	if err != nil {
		return crypto.PrivKey{}, errors.Wrap(err, "Private Key")
	}
	return privKey, nil
}

type noop struct{}

func (n noop) Encrypt(key crypto.PrivKey, passphrase string) (saltBytes []byte, encBytes []byte, err error) {
	return []byte{}, key.Bytes(), nil
}

func (n noop) Decrypt(saltBytes []byte, encBytes []byte, passphrase string) (privKey crypto.PrivKey, err error) {
	return crypto.PrivKeyFromBytes(encBytes)
}
