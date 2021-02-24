package sr25519

import (
	tmsr25519 "github.com/tendermint/tendermint/crypto/sr25519"
)

// Sign produces a signature on the provided message.
func (privKey *PrivKey) Sign(msg []byte) ([]byte, error) {
	return tmsr25519.PrivKey(privKey.Key).Sign(msg)
}

// VerifySignature - verifies a signature
func (pubKey *PubKey) VerifySignature(msg []byte, sigStr []byte) bool {
	return tmsr25519.PubKey(pubKey.Key).VerifySignature(msg, sigStr)
}
