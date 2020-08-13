// +build libsecp256k1

package secp256k1

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1/internal/secp256k1"
)

// Sign creates an ECDSA signature on curve Secp256k1, using SHA256 on the msg.
func (privKey PrivKeySecp256k1) Sign(msg []byte) ([]byte, error) {
	rsv, err := secp256k1.Sign(crypto.Sha256(msg), privKey[:])
	if err != nil {
		return nil, err
	}
	// we do not need v  in r||s||v:
	rs := rsv[:len(rsv)-1]
	return rs, nil
}

func (pubKey PubKeySecp256k1) VerifyBytes(msg []byte, sig []byte) bool {
	return secp256k1.VerifySignature(pubKey[:], crypto.Sha256(msg), sig)
}
