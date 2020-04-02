package keyring

import (
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/types"
	bip39 "github.com/cosmos/go-bip39"
)

var fundraiserPath = types.GetConfig().GetFullFundraiserPath()

// StdPrivKeyGen is the default PrivKeyGen function in the keybase.
// For now, it only supports Secp256k1
func StdPrivKeyGen(bz []byte, algo pubKeyType) (tmcrypto.PrivKey, error) {
	if algo == Secp256k1 {
		return SecpPrivKeyGen(bz), nil
	}
	return nil, ErrUnsupportedSigningAlgo
}

// SecpPrivKeyGen generates a secp256k1 private key from the given bytes
func SecpPrivKeyGen(bz []byte) tmcrypto.PrivKey {
	var bzArr [32]byte
	copy(bzArr[:], bz)
	return secp256k1.PrivKeySecp256k1(bzArr)
}

// StdDeriveKey is the default DeriveKey function in the keybase.
// For now, it only supports Secp256k1
func StdDeriveKey(mnemonic string, bip39Passphrase, hdPath string, algo pubKeyType) ([]byte, error) {
	if algo == Secp256k1 {
		return SecpDeriveKey(mnemonic, bip39Passphrase, hdPath)
	}
	return nil, ErrUnsupportedSigningAlgo
}

// SecpDeriveKey derives and returns the secp256k1 private key for the given seed and HD path.
func SecpDeriveKey(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error) {
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
