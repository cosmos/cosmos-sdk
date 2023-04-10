package hd

import (
	"github.com/cosmos/go-bip39"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

// PubKeyType defines an algorithm to derive key-pairs which can be used for cryptographic signing.
type PubKeyType string

const (
	// MultiType implies that a pubkey is a multisignature
	MultiType = PubKeyType("multi")
	// Secp256k1Type uses the Bitcoin secp256k1 ECDSA parameters.
	Secp256k1Type = PubKeyType("secp256k1")
	// EthSecp256k1Type uses the Ethereum secp256k1 ECDSA parameters.
	EthSecp256k1Type = PubKeyType("ethsecp256k1")
	// Ed25519Type represents the Ed25519Type signature system.
	// It is currently not supported for end-user keys (wallets/ledgers).
	Ed25519Type = PubKeyType("ed25519")
	// Sr25519Type represents the Sr25519Type signature system.
	Sr25519Type = PubKeyType("sr25519")
)

// Secp256k1 uses the Bitcoin secp256k1 ECDSA parameters.
var Secp256k1 = secp256k1Algo{}

type (
	DeriveFn   func(mnemonic, bip39Passphrase, hdPath string) ([]byte, error)
	GenerateFn func(bz []byte) types.PrivKey
)

type WalletGenerator interface {
	Derive(mnemonic, bip39Passphrase, hdPath string) ([]byte, error)
	Generate(bz []byte) types.PrivKey
}

type secp256k1Algo struct{}

func (s secp256k1Algo) Name() PubKeyType {
	return Secp256k1Type
}

// Derive derives and returns the secp256k1 private key for the given seed and HD path.
func (s secp256k1Algo) Derive() DeriveFn {
	return func(mnemonic, bip39Passphrase, hdPath string) ([]byte, error) {
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
		if err != nil {
			return nil, err
		}

		masterPriv, ch := ComputeMastersFromSeed(seed)
		if len(hdPath) == 0 {
			return masterPriv[:], nil
		}
		derivedKey, err := DerivePrivateKeyForPath(masterPriv, ch, hdPath)

		return derivedKey, err
	}
}

// Generate generates a secp256k1 private key from the given bytes.
func (s secp256k1Algo) Generate() GenerateFn {
	return func(bz []byte) types.PrivKey {
		bzArr := make([]byte, secp256k1.PrivKeySize)
		copy(bzArr, bz)

		return &secp256k1.PrivKey{Key: bzArr}
	}
}

var (
	// EthSecp256k1 uses the Bitcoin secp256k1 ECDSA parameters.
	EthSecp256k1 = ethSecp256k1Algo{}
)

type ethSecp256k1Algo struct{}

// Name returns eth_secp256k1
func (s ethSecp256k1Algo) Name() PubKeyType {
	return EthSecp256k1Type
}

// `Derive` derives and returns the eth_secp256k1 private key for the given mnemonic and HD path.
func (s ethSecp256k1Algo) Derive() DeriveFn {
	return func(mnemonic, bip39Passphrase, hdPath string) ([]byte, error) {
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
		if err != nil {
			return nil, err
		}

		masterPriv, ch := ComputeMastersFromSeed(seed)
		if len(hdPath) == 0 {
			return ECDSAify(masterPriv[:])
		}

		derivedKey, err := DerivePrivateKeyForPath(masterPriv, ch, hdPath)
		if err != nil {
			return nil, err
		}

		return ECDSAify(derivedKey)
	}
}

// Generate generates a eth_secp256k1 private key from the given bytes.
func (s ethSecp256k1Algo) Generate() GenerateFn {
	return func(bz []byte) types.PrivKey {
		bzArr := make([]byte, secp256k1.EthPrivKeySize)
		copy(bzArr, bz)
		return &secp256k1.EthPrivKey{
			Key: bzArr,
		}
	}
}

// `ECDSAify` converts a private key to an ECDSA private key.
func ECDSAify(key []byte) ([]byte, error) {
	// Convert the private key to an ECDSA private key.
	x, err := secp256k1.EthPrivKey{Key: key}.ToECDSA()
	if err != nil {
		return nil, err
	}

	// Return the private key as a byte slice.
	return secp256k1.FromECDSA(x), nil
}
