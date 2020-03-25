package keys

import (
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"

	schnorrkel "github.com/ChainSafe/go-schnorrkel"
	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/tendermint/tendermint/crypto"
	"golang.org/x/crypto/ed25519"
)

// GenPrivKey generates a new ECDSA private key on the specified curve.
// It uses OS randomness to generate the private key.
func GenPrivKey(curve Curve) (crypto.PrivKey, error) {
	switch curve {
	case ED25519:
		return genPrivKeyEd25519(crypto.CReader()), nil
	case SECP256K1:
		return genPrivKeySecp256k1(crypto.CReader()), nil
	case SR25519:
		return genPrivKeySr25519(crypto.CReader()), nil
	default:
		return nil, fmt.Errorf("invalid key type: %s", curve.String())
	}
}

// GenPrivKeyFromSecret generates a new ECDSA private key on the specified curve
// from a secret.
func GenPrivKeyFromSecret(curve Curve, secret []byte) (crypto.PrivKey, error) {
	switch curve {
	case ED25519:
		return genPrivKeyEd25519FromSecret(secret), nil
	case SECP256K1:
		return genPrivKeySecp256k1FromSecret(secret), nil
	case SR25519:
		return genPrivKeySr25519FromSecret(secret), nil
	default:
		return nil, fmt.Errorf("invalid key type: %s", curve.String())
	}
}

// genPrivKeyEd25519 generates a new ed25519 private key.
// It uses OS randomness in conjunction with the current global random seed
// in tendermint/libs/common to generate the private key.
func genPrivKeyEd25519(rand io.Reader) PrivKeyEd25519 {
	seed := make([]byte, 32)
	_, err := io.ReadFull(rand, seed)
	if err != nil {
		panic(err)
	}

	privKey := ed25519.NewKeyFromSeed(seed)
	var privKeyEd []byte
	copy(privKeyEd[:], privKey)
	return PrivKeyEd25519{bytes: privKeyEd}
}

// genPrivKeySecp256k1 generates a new secp256k1 private key using the provided reader.
func genPrivKeySecp256k1(rand io.Reader) PrivKeySecp256k1 {
	var privKeyBytes []byte
	d := new(big.Int)
	for {
		privKeyBytes = []byte{}
		_, err := io.ReadFull(rand, privKeyBytes[:])
		if err != nil {
			panic(err)
		}

		d.SetBytes(privKeyBytes[:])
		// break if we found a valid point (i.e. > 0 and < N == curverOrder)
		isValidFieldElement := 0 < d.Sign() && d.Cmp(secp256k1.S256().N) < 0
		if isValidFieldElement {
			break
		}
	}

	return PrivKeySecp256k1{bytes: privKeyBytes}
}

// genPrivKeySr25519 generates a new sr25519 private key.
// It uses OS randomness in conjunction with the current global random seed
// in tendermint/libs/common to generate the private key.
func genPrivKeySr25519(rand io.Reader) PrivKeySr25519 {
	var seed [64]byte

	out := make([]byte, 64)
	_, err := io.ReadFull(rand, out)
	if err != nil {
		panic(err)
	}

	copy(seed[:], out)

	privKeySr := schnorrkel.NewMiniSecretKey(seed).ExpandEd25519().Encode()
	return PrivKeySr25519{bytes: privKeySr[:]}
}

// genPrivKeyEd25519FromSecret hashes the secret with SHA2, and uses
// that 32 byte output to create the private key.
// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func genPrivKeyEd25519FromSecret(secret []byte) PrivKeyEd25519 {
	seed := crypto.Sha256(secret) // Not Ripemd160 because we want 32 bytes.

	privKey := ed25519.NewKeyFromSeed(seed)
	var privKeyEd []byte
	copy(privKeyEd[:], privKey)
	return PrivKeyEd25519{bytes: privKeyEd}
}

// genPrivKeySecp256k1FromSecret hashes the secret with SHA2, and uses
// that 32 byte output to create the private key.
//
// It makes sure the private key is a valid field element by setting:
//
// c = sha256(secret)
// k = (c mod (n − 1)) + 1, where n = curve order.
//
// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func genPrivKeySecp256k1FromSecret(secret []byte) PrivKeySecp256k1 {
	one := new(big.Int).SetInt64(1)
	secHash := sha256.Sum256(secret)
	// to guarantee that we have a valid field element, we use the approach of:
	// "Suite B Implementer’s Guide to FIPS 186-3", A.2.1
	// https://apps.nsa.gov/iaarchive/library/ia-guidance/ia-solutions-for-classified/algorithm-guidance/suite-b-implementers-guide-to-fips-186-3-ecdsa.cfm
	// see also https://github.com/golang/go/blob/0380c9ad38843d523d9c9804fe300cb7edd7cd3c/src/crypto/ecdsa/ecdsa.go#L89-L101
	fe := new(big.Int).SetBytes(secHash[:])
	n := new(big.Int).Sub(secp256k1.S256().N, one)
	fe.Mod(fe, n)
	fe.Add(fe, one)

	feB := fe.Bytes()
	var privKey32 []byte
	// copy feB over to fixed 32 byte privKey32 and pad (if necessary)
	copy(privKey32[32-len(feB):32], feB)

	return PrivKeySecp256k1{bytes: privKey32}
}

// genPrivKeySr25519FromSecret hashes the secret with SHA2, and uses
// that 32 byte output to create the private key.
// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func genPrivKeySr25519FromSecret(secret []byte) PrivKeySr25519 {
	seed := crypto.Sha256(secret) // Not Ripemd160 because we want 32 bytes.
	var bz [PrivKeySr25519Size]byte
	copy(bz[:], seed)
	privKey, _ := schnorrkel.NewMiniSecretKeyFromRaw(bz)
	privKeySr := privKey.ExpandEd25519().Encode()
	return PrivKeySr25519{bytes: privKeySr[:]}
}
