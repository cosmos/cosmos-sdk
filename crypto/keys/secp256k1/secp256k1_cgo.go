// +build libsecp256k1

package secp256k1

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1/internal/secp256k1"
)

// Sign creates an ECDSA signature on curve Secp256k1, using SHA256 on the msg.
func (privKey *PrivKey) Sign(msg []byte) ([]byte, error) {
	rsv, err := secp256k1.Sign(crypto.Sha256(msg), privKey.Key)
	if err != nil {
		return nil, err
	}
	// we do not need v  in r||s||v:
	rs := rsv[:len(rsv)-1]
	return rs, nil
}

// VerifySignature validates the signature.
// The msg will be hashed prior to signature verification.
func (pubKey *PrivKey) VerifySignature(msg []byte, sig []byte) bool {
	return secp256k1.VerifySignature(pubKey.Key, crypto.Sha256(msg), sig)
}
