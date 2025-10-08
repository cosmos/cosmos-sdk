package hd

import (
	"fmt"

	"github.com/cloudflare/circl/sign/mldsa/mldsa44"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/go-bip39"
)

const (
	MLDSAType = PubKeyType("mldsa44")
)

type mldsaAlgo struct{}

func (s mldsaAlgo) Name() PubKeyType {
	return MLDSAType
}

// Derive derives and returns the mldsa44 private key for the given seed and HD path.
func (s mldsaAlgo) Derive() DeriveFn {
	return func(mnemonic, bip39Passphrase, hdPath string) ([]byte, error) {
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
		if err != nil {
			return nil, err
		}
		scheme := mldsa44.Scheme()
		_, privateKey := scheme.DeriveKey(seed)
		return privateKey.MarshalBinary()
	}
}

// Generate generates a mldsa44 private key from the given bytes.
func (s mldsaAlgo) Generate() GenerateFn {
	return func(bz []byte) types.PrivKey {
		scheme := mldsa44.Scheme()
		_, privateKey, err := scheme.GenerateKey()
		if err != nil {
			panic(err)
		}
		data, err := privateKey.MarshalBinary()
		if err != nil {
			panic(fmt.Errorf("failed to marshal generated private key: %w", err))
		}
		return &mldsa.PrivKey{
			Key: data,
		}
	}
}
