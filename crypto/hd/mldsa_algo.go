package hd

import (
	"crypto/sha256"

	"github.com/cloudflare/circl/sign/mldsa/mldsa44"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/go-bip39"
)

const (
	MLDSA44Type = PubKeyType("mldsa44")
)

type mldsaAlgo struct{}

var Mldsa44 = mldsaAlgo{}

func (s mldsaAlgo) Name() PubKeyType {
	return MLDSA44Type
}

// Derive derives and returns the mldsa44 private key for the given seed and HD path.
func (s mldsaAlgo) Derive() DeriveFn {
	return func(mnemonic, bip39Passphrase, hdPath string) ([]byte, error) {
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
		if err != nil {
			return nil, err
		}
		shaSeed := sha256.Sum256(seed)
		scheme := mldsa44.Scheme()
		_, privateKey := scheme.DeriveKey(shaSeed[:])
		return privateKey.MarshalBinary()
	}
}

// Generate generates a mldsa44 private key from the given bytes.
func (s mldsaAlgo) Generate() GenerateFn {
	return func(bz []byte) types.PrivKey {
		scheme := mldsa44.Scheme()
		bzArr := make([]byte, scheme.PrivateKeySize())
		copy(bzArr, bz)
		return &mldsa.PrivKey{
			Key: bzArr,
		}
	}
}
