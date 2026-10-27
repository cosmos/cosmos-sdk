package hd

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

// MlDsa65 is the ML-DSA-65 (FIPS 204) account key algorithm.
var MlDsa65 = mldsa65Algo{}

type mldsa65Algo struct{}

func (mldsa65Algo) Name() PubKeyType {
	return MlDsa65Type
}

// Derive reuses the secp256k1 BIP32 derivation. It returns the 32-byte
// BIP32-derived key for the given mnemonic, passphrase, and HD path, which is
// used directly as the ML-DSA-65 keygen seed (mldsa65 SeedSize == 32). This
// gives mnemonic recovery and per-path account separation without inventing a
// new derivation scheme.
func (mldsa65Algo) Derive() DeriveFn {
	return Secp256k1.Derive()
}

// Generate builds an ML-DSA-65 private key from the 32-byte derived seed.
func (mldsa65Algo) Generate() GenerateFn {
	return func(bz []byte) types.PrivKey {
		privKey, err := mldsa65.GenPrivKeyFromSeed(bz)
		if err != nil {
			// A non-32-byte seed only reaches here from a caller passing
			// untrusted/dummy bytes directly (e.g. ImportPrivKeyHex); the BIP32
			// derivation path always supplies a 32-byte seed. Such callers must
			// guard their input — see keyring.generatePrivKey.
			panic(err)
		}
		return &privKey
	}
}
