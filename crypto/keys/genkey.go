package keys

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"
)

// GenPrivKey generates a new ECDSA private key on the specified curve.
// It uses OS randomness to generate the private key.
func GenPrivKey(curve Curve) (crypto.PrivKey, error) {
	switch curve {
	case ED25519:
		return ed25519.GenPrivKey(), nil
	case SECP256K1:
		return secp256k1.GenPrivKey(), nil
	case SR25519:
		return sr25519.GenPrivKey(), nil
	default:
		return nil, fmt.Errorf("invalid key type: %s", curve.String())
	}
}

// GenPrivKeyFromSecret generates a new ECDSA private key on the specified curve
// from a secret.
func GenPrivKeyFromSecret(curve Curve, secret []byte) (crypto.PrivKey, error) {
	switch curve {
	case ED25519:
		return ed25519.GenPrivKeyFromSecret(secret), nil
	case SECP256K1:
		return nil, nil //TODO this is not supported currently
	case SR25519:
		return sr25519.GenPrivKeyFromSecret(secret), nil
	default:
		return nil, fmt.Errorf("invalid key type: %s", curve.String())
	}
}
