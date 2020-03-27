package keyring

import "github.com/tendermint/tendermint/crypto"

type AltSigningAlgo struct {
	Name       SigningAlgo
	DeriveKey  AltDeriveKeyFunc
	PrivKeyGen AltPrivKeyGenFunc
}

type AltDeriveKeyFunc func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error)
type AltPrivKeyGenFunc func(bz []byte) crypto.PrivKey

var (
	// Secp256k1 uses the Bitcoin secp256k1 ECDSA parameters.
	AltSecp256k1 = AltSigningAlgo{Name: Secp256k1, DeriveKey: SecpDeriveKey, PrivKeyGen: SecpPrivKeyGen}
)

type AltSigningAlgoList []AltSigningAlgo

func (l AltSigningAlgoList) Contains(algo AltSigningAlgo) bool {
	for _, cAlgo := range l {
		if cAlgo.Name == algo.Name {
			return true
		}
	}

	return false
}
