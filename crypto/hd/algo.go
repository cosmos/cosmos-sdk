package hd

import (
	bip39 "github.com/cosmos/go-bip39"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/sr25519"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	substratebip39 "github.com/sikkatech/go-substrate-bip39"
)

// PubKeyType defines an algorithm to derive key-pairs which can be used for cryptographic signing.
type PubKeyType string

const (
	// MultiType implies that a pubkey is a multisignature
	MultiType = PubKeyType("multi")
	// Secp256k1Type uses the Bitcoin secp256k1 ECDSA parameters.
	Secp256k1Type = PubKeyType("secp256k1")
	// Ed25519Type represents the Ed25519Type signature system.
	// It is currently not supported for end-user keys (wallets/ledgers).
	Ed25519Type = PubKeyType("ed25519")
	// Sr25519Type represents the Sr25519Type signature system.
	Sr25519Type = PubKeyType("sr25519")
)

var (
	// Secp256k1 uses the Bitcoin secp256k1 ECDSA parameters.
	Secp256k1 = secp256k1Algo{}
	// Sr25519 represents the Sr25519Type signature system.
	Sr25519 = sr25519Algo{}
)

type DeriveFn func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error)
type GenerateFn func(bz []byte) types.PrivKey

type WalletGenerator interface {
	Derive(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error)
	Generate(bz []byte) types.PrivKey
}

type secp256k1Algo struct {
}

func (s secp256k1Algo) Name() PubKeyType {
	return Secp256k1Type
}

// Derive derives and returns the secp256k1 private key for the given seed and HD path.
func (s secp256k1Algo) Derive() DeriveFn {
	return func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error) {
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
		var bzArr = make([]byte, secp256k1.PrivKeySize)
		copy(bzArr, bz)

		return &secp256k1.PrivKey{Key: bzArr}
	}
}

type sr25519Algo struct {
}

func (s sr25519Algo) Name() PubKeyType {
	return Sr25519Type
}

// Derive derives and returns the sr25519 private key for the given seed and HD path.
func (s sr25519Algo) Derive() DeriveFn {
	return func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error) {
		seed, err := substratebip39.SeedFromMnemonic(mnemonic, bip39Passphrase)
		return seed[:], err
	}
}

// Generate generates a sr25519 private key from the given bytes.
func (s sr25519Algo) Generate() GenerateFn {
	return func(bz []byte) types.PrivKey {
		return &sr25519.PrivKey{Key: bz}
	}
}
