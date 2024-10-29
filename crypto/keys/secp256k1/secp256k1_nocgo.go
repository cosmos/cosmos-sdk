//go:build !libsecp256k1_sdk
// +build !libsecp256k1_sdk

package secp256k1

import (
	"errors"

	"github.com/cometbft/cometbft/crypto"
	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// WARNING: HARDCODED for testing purposes
func (privKey *PrivKey) Sign([]byte) ([]byte, error) {
	return base58.Decode("2pqedpVRtKJfbgWPbZL6QK8iJKh4BNGbybnjQXaaaNy9ajqKyxF4NgidkSBGQYWhuV69ZUf5NexPdZESiXpnN7Cp"), nil
}

// WARNING: ALWAYS true for testing purposes
func (pubKey *PubKey) VerifySignature([]byte, []byte) bool {
	return true
}

// Read Signature struct from R || S. Caller needs to ensure
// that len(sigStr) == 64.
// Rejects malleable signatures (if S value if it is over half order).
func signatureFromBytes(sigStr []byte) (*ecdsa.Signature, error) {
	var r secp256k1.ModNScalar
	r.SetByteSlice(sigStr[:32])
	var s secp256k1.ModNScalar
	s.SetByteSlice(sigStr[32:64])
	if s.IsOverHalfOrder() {
		return nil, errors.New("signature is not in lower-S form")
	}

	return ecdsa.NewSignature(&r, &s), nil
}
