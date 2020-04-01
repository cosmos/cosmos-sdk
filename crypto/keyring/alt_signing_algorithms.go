package keyring

import "github.com/tendermint/tendermint/crypto"

type AltSigningAlgo interface {
	Name() pubKeyType
	DeriveKey() AltDeriveKeyFunc
	PrivKeyGen() AltPrivKeyGenFunc
}

type AltDeriveKeyFunc func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error)
type AltPrivKeyGenFunc func(bz []byte) crypto.PrivKey

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

type AltSigningAlgoList []AltSigningAlgo

func (l AltSigningAlgoList) Contains(algo AltSigningAlgo) bool {
	for _, cAlgo := range l {
		if cAlgo.Name() == algo.Name() {
			return true
		}
	}

	return false
}
