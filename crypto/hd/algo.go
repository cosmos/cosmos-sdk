package hd

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	ethsecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1/eth"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/go-bip39"
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

var (
	// Secp256k1 uses the Bitcoin secp256k1 ECDSA parameters.
	Secp256k1 = secp256k1Algo{keyType: Secp256k1Type}
	// Secp256k1 uses the Ethereum secp256k1 ECDSA parameters.
	EthSecp256k1 = secp256k1Algo{keyType: EthSecp256k1Type}
)

type (
	DeriveFn   func(mnemonic, bip39Passphrase, hdPath string) ([]byte, error)
	GenerateFn func(bz []byte) types.PrivKey
)

// secp256k1Algo represents a key derivation algorithmn.
type secp256k1Algo struct {
	keyType PubKeyType
}

// Name returns the keyType of the algorithmn.
func (s secp256k1Algo) Name() PubKeyType {
	return s.keyType
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

		return DerivePrivateKeyForPath(masterPriv, ch, hdPath)
	}
}

// Generate generates a secp256k1 private key from the given bytes.
func (s secp256k1Algo) Generate() GenerateFn {
	return func(bz []byte) types.PrivKey {
		bzArr := make([]byte, secp256k1.PrivKeySize)
		copy(bzArr, bz)
		if s.keyType == EthSecp256k1Type {
			return &ethsecp256k1.PrivKey{Key: bz}
		}
		return &secp256k1.PrivKey{Key: bz}
	}
}
