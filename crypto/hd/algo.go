package hd

import (
	"crypto/elliptic"

	"github.com/btcsuite/btcd/btcec"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

// PubKeyType defines an algorithm to derive key-pairs which can be used for cryptographic signing.
type PubKeyType string

// Signature scheme types
const (
	MultiType     = PubKeyType("multi")
	Secp256k1Type = PubKeyType("secp256k1")
	Secp256r1Type = PubKeyType(secp256r1.Name)

	Ed25519Type = PubKeyType("ed25519")
	Sr25519Type = PubKeyType("sr25519")
)

// Secp256 algos use the Bip39 derivation mechanism
var (
	Secp256k1 = ecdsaAlgo{btcec.S256(), Secp256k1Type, genSecp256k1}
	Secp256r1 = ecdsaAlgo{elliptic.P256(), Secp256r1Type, genSecp256r1}
)

type DeriveFn func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error)
type GenerateFn func(bz []byte) types.PrivKey

type WalletGenerator interface {
	Derive(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error)
	Generate(bz []byte) types.PrivKey
}

type ecdsaAlgo struct {
	curve elliptic.Curve
	name  PubKeyType
	gen   GenerateFn
}

// Name returns signature scheme name
func (s ecdsaAlgo) Name() PubKeyType { return s.name }

// Derive derives and returns the secp256k1 private key for the given seed and HD path.
func (s ecdsaAlgo) Derive() DeriveFn {
	return s.derive
}

func (s ecdsaAlgo) derive(mnemonic, bip39Passphrase, hdPath string) ([]byte, error) {
	m, ch, err := masterPrivKey(mnemonic, bip39Passphrase)
	if err != nil {
		return nil, err
	}
	if len(hdPath) == 0 {
		return m[:], nil
	}
	return DeriveECDSAPrivKey(s.curve, m, ch, hdPath)
}

// Generate generates a secp256k1 private key from the given bytes.
func (s ecdsaAlgo) Generate() GenerateFn {
	return s.gen
}

func genSecp256k1(bz []byte) types.PrivKey {
	var bzArr = make([]byte, secp256k1.PrivKeySize)
	copy(bzArr, bz)
	return &secp256k1.PrivKey{Key: bzArr}
}

func genSecp256r1(bz []byte) types.PrivKey {
	key, err := secp256r1.NewPrivKeyFromSecret(bz)
	if err != nil {
		panic(err)
	}
	return key
}
