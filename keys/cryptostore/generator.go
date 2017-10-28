package cryptostore

import (
	"github.com/pkg/errors"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/nano"
)

var (
	// GenEd25519 produces Ed25519 private keys
	GenEd25519 Generator = GenFunc(genEd25519)
	// GenSecp256k1 produces Secp256k1 private keys
	GenSecp256k1 Generator = GenFunc(genSecp256)
	// GenLedgerEd25519 used Ed25519 keys stored on nano ledger s with cosmos app
	GenLedgerEd25519 Generator = GenFunc(genLedgerEd25519)
)

// Generator determines the type of private key the keystore creates
type Generator interface {
	Generate(secret []byte) (crypto.PrivKey, error)
}

// GenFunc is a helper to transform a function into a Generator
type GenFunc func(secret []byte) (crypto.PrivKey, error)

func (f GenFunc) Generate(secret []byte) (crypto.PrivKey, error) {
	return f(secret)
}

func genEd25519(secret []byte) (crypto.PrivKey, error) {
	key := crypto.GenPrivKeyEd25519FromSecret(secret).Wrap()
	return key, nil
}

func genSecp256(secret []byte) (crypto.PrivKey, error) {
	key := crypto.GenPrivKeySecp256k1FromSecret(secret).Wrap()
	return key, nil
}

// secret is completely ignored for the ledger...
// just for interface compatibility
func genLedgerEd25519(secret []byte) (crypto.PrivKey, error) {
	return nano.NewPrivKeyLedgerEd25519Ed25519()
}

type genInvalidByte struct {
	typ byte
}

func (g genInvalidByte) Generate(secret []byte) (crypto.PrivKey, error) {
	err := errors.Errorf("Cannot generate keys for algorithm: %X", g.typ)
	return crypto.PrivKey{}, err
}

type genInvalidAlgo struct {
	algo string
}

func (g genInvalidAlgo) Generate(secret []byte) (crypto.PrivKey, error) {
	err := errors.Errorf("Cannot generate keys for algorithm: %s", g.algo)
	return crypto.PrivKey{}, err
}

func getGenerator(algo string) Generator {
	switch algo {
	case crypto.NameEd25519:
		return GenEd25519
	case crypto.NameSecp256k1:
		return GenSecp256k1
	case nano.NameLedgerEd25519:
		return GenLedgerEd25519
	default:
		return genInvalidAlgo{algo}
	}
}

func getGeneratorByType(typ byte) Generator {
	switch typ {
	case crypto.TypeEd25519:
		return GenEd25519
	case crypto.TypeSecp256k1:
		return GenSecp256k1
	case nano.TypeLedgerEd25519:
		return GenLedgerEd25519
	default:
		return genInvalidByte{typ}
	}
}
