package privkey

import (
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
)

var (
	// Secp256k1 uses the Bitcoin secp256k1 ECDSA parameters.
	Secp256k1 = secp256k1Algo{}
)

type DeriveKeyFn func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error)
type GenerateFn func(bz []byte) crypto.PrivKey

type HDWallet interface {
	Derive(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error)
	Generate(bz []byte) crypto.PrivKey
}

type secp256k1Algo struct {
}

func (s secp256k1Algo) Name() PubKeyType {
	return Secp256k1Type
}

// Derive derives and returns the secp256k1 private key for the given seed and HD path.
func (s secp256k1Algo) Derive() DeriveKeyFn {
	return func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error) {
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
		if err != nil {
			return nil, err
		}

		masterPriv, ch := hd.ComputeMastersFromSeed(seed)
		if len(hdPath) == 0 {
			return masterPriv[:], nil
		}
		derivedKey, err := hd.DerivePrivateKeyForPath(masterPriv, ch, hdPath)
		return derivedKey[:], err
	}
}

// Generate generates a secp256k1 private key from the given bytes
func (s secp256k1Algo) Generate() GenerateFn {
	return func(bz []byte) crypto.PrivKey {
		var bzArr [32]byte
		copy(bzArr[:], bz)
		return secp256k1.PrivKeySecp256k1(bzArr)
	}
}
