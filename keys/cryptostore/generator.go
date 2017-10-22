package cryptostore

import (
	"github.com/pkg/errors"

	crypto "github.com/tendermint/go-crypto"
)

var (
	// GenEd25519 produces Ed25519 private keys
	GenEd25519 Generator = GenFunc(genEd25519)
	// GenSecp256k1 produces Secp256k1 private keys
	GenSecp256k1 Generator = GenFunc(genSecp256)
)

// Generator determines the type of private key the keystore creates
type Generator interface {
	Generate(secret []byte) crypto.PrivKey
}

// GenFunc is a helper to transform a function into a Generator
type GenFunc func(secret []byte) crypto.PrivKey

func (f GenFunc) Generate(secret []byte) crypto.PrivKey {
	return f(secret)
}

func genEd25519(secret []byte) crypto.PrivKey {
	return crypto.GenPrivKeyEd25519FromSecret(secret).Wrap()
}

func genSecp256(secret []byte) crypto.PrivKey {
	return crypto.GenPrivKeySecp256k1FromSecret(secret).Wrap()
}

func getGenerator(algo string) (Generator, error) {
	switch algo {
	case crypto.NameEd25519:
		return GenEd25519, nil
	case crypto.NameSecp256k1:
		return GenSecp256k1, nil
	default:
		return nil, errors.Errorf("Cannot generate keys for algorithm: %s", algo)
	}
}

func getGeneratorByType(typ byte) (Generator, error) {
	switch typ {
	case crypto.TypeEd25519:
		return GenEd25519, nil
	case crypto.TypeSecp256k1:
		return GenSecp256k1, nil
	default:
		return nil, errors.Errorf("Cannot generate keys for algorithm: %X", typ)
	}
}
