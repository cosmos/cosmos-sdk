package keyring

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"
)

// pubKeyType defines an algorithm to derive key-pairs which can be used for cryptographic signing.
type pubKeyType string

const (
	// MultiAlgo implies that a pubkey is a multisignature
	MultiAlgo = pubKeyType("multi")
	// Secp256k1 uses the Bitcoin secp256k1 ECDSA parameters.
	Secp256k1 = pubKeyType("secp256k1")
	// Ed25519 represents the Ed25519 signature system.
	// It is currently not supported for end-user keys (wallets/ledgers).
	Ed25519 = pubKeyType("ed25519")
	// Sr25519 represents the Sr25519 signature system.
	Sr25519 = pubKeyType("sr25519")
)

// IsSupportedAlgorithm returns whether the signing algorithm is in the passed-in list of supported algorithms.
func IsSupportedAlgorithm(supported []pubKeyType, algo pubKeyType) bool {
	for _, supportedAlgo := range supported {
		if algo == supportedAlgo {
			return true
		}
	}
	return false
}

type SigningAlgo interface {
	Name() pubKeyType
	DeriveKey() AltDeriveKeyFunc
	PrivKeyGen() AltPrivKeyGenFunc
}

type AltDeriveKeyFunc func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error)
type AltPrivKeyGenFunc func(bz []byte) crypto.PrivKey

func NewSigningAlgoFromString(str string) (SigningAlgo, error) {
	if str != string(AltSecp256k1.Name()) {
		return nil, fmt.Errorf("provided algorithm `%s` is not supported", str)
	}

	return AltSecp256k1, nil
}

type secp256k1Algo struct {
}

func (s secp256k1Algo) Name() pubKeyType {
	return Secp256k1
}

func (s secp256k1Algo) DeriveKey() AltDeriveKeyFunc {
	return SecpDeriveKey
}

func (s secp256k1Algo) PrivKeyGen() AltPrivKeyGenFunc {
	return SecpPrivKeyGen
}

var (
	// Secp256k1 uses the Bitcoin secp256k1 ECDSA parameters.
	AltSecp256k1 = secp256k1Algo{}
)

type AltSigningAlgoList []SigningAlgo

func (l AltSigningAlgoList) Contains(algo SigningAlgo) bool {
	for _, cAlgo := range l {
		if cAlgo.Name() == algo.Name() {
			return true
		}
	}

	return false
}
