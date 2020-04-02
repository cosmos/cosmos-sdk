package keyring

import (
	"fmt"

	"github.com/cosmos/go-bip39"
	"github.com/tendermint/tendermint/crypto"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
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

type SignatureAlgo interface {
	Name() pubKeyType
	DeriveKey() DeriveKeyFn
	PrivKeyGen() PrivKeyGenFn
}

func NewSigningAlgoFromString(str string) (SignatureAlgo, error) {
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

func (s secp256k1Algo) DeriveKey() DeriveKeyFn {
	return Secp256k1DeriveKey
}

func (s secp256k1Algo) PrivKeyGen() PrivKeyGenFn {
	return Secp256k1PrivKeyGen
}

var (
	// Secp256k1 uses the Bitcoin secp256k1 ECDSA parameters.
	AltSecp256k1 = secp256k1Algo{}
)

type SigningAlgoList []SignatureAlgo

func (l SigningAlgoList) Contains(algo SignatureAlgo) bool {
	for _, cAlgo := range l {
		if cAlgo.Name() == algo.Name() {
			return true
		}
	}

	return false
}

// Secp256k1PrivKeyGen generates a secp256k1 private key from the given bytes
func Secp256k1PrivKeyGen(bz []byte) tmcrypto.PrivKey {
	var bzArr [32]byte
	copy(bzArr[:], bz)
	return secp256k1.PrivKeySecp256k1(bzArr)
}

// Secp256k1DeriveKey derives and returns the secp256k1 private key for the given seed and HD path.
func Secp256k1DeriveKey(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error) {
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

type DeriveKeyFn func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error)
type PrivKeyGenFn func(bz []byte) crypto.PrivKey
