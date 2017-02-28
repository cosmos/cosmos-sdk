package cryptostore

import crypto "github.com/tendermint/go-crypto"

var (
	// GenEd25519 produces Ed25519 private keys
	GenEd25519 Generator = GenFunc(genEd25519)
	// GenSecp256k1 produces Secp256k1 private keys
	GenSecp256k1 Generator = GenFunc(genSecp256)
)

// Generator determines the type of private key the keystore creates
type Generator interface {
	Generate() crypto.PrivKey
}

// GenFunc is a helper to transform a function into a Generator
type GenFunc func() crypto.PrivKey

func (f GenFunc) Generate() crypto.PrivKey {
	return f()
}

func genEd25519() crypto.PrivKey {
	return crypto.GenPrivKeyEd25519()
}

func genSecp256() crypto.PrivKey {
	return crypto.GenPrivKeySecp256k1()
}
