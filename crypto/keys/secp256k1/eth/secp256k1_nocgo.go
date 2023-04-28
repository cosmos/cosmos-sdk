//go:build !libsecp256k1_sdk
// +build !libsecp256k1_sdk

package eth

import (
	"crypto/ecdsa"
	"errors"

	"github.com/btcsuite/btcd/btcec/v2"
	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	decdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"

	"github.com/cometbft/cometbft/crypto"
)

// Sign creates an ECDSA signature on curve Secp256k1, using SHA256 on the msg.
// The returned signature will be of the form R || S (in lower-S form).
func (privKey *PrivKey) Sign(msg []byte) ([]byte, error) {
	priv := secp256k1.PrivKeyFromBytes(privKey.Key)
	sig := decdsa.SignCompact(priv, crypto.Sha256(msg), false)

	// remove the first byte which is compactSigRecoveryCode
	return sig[1:], nil
}

// VerifyBytes verifies a signature of the form R || S.
// It rejects signatures which are not in lower-S form.
func (pubKey *PubKey) VerifySignature(msg, sigStr []byte) bool {
	if len(sigStr) != 64 {
		return false
	}
	pub, err := secp256k1.ParsePubKey(pubKey.Key)
	if err != nil {
		return false
	}
	// parse the signature, will return error if it is not in lower-S form
	signature, err := signatureFromBytes(sigStr)
	if err != nil {
		return false
	}
	return signature.Verify(crypto.Sha256(msg), pub)
}

// Read Signature struct from R || S. Caller needs to ensure
// that len(sigStr) == 64.
// Rejects malleable signatures (if S value if it is over half order).
func signatureFromBytes(sigStr []byte) (*decdsa.Signature, error) {
	var r secp256k1.ModNScalar
	r.SetByteSlice(sigStr[:32])
	var s secp256k1.ModNScalar
	s.SetByteSlice(sigStr[32:64])
	if s.IsOverHalfOrder() {
		return nil, errors.New("signature is not in lower-S form")
	}

	return decdsa.NewSignature(&r, &s), nil
}

// DecompressPubkey parses a public key in the 33-byte compressed format.
func DecompressPubkey(pubkey []byte) (*ecdsa.PublicKey, error) {
	if len(pubkey) != 33 {
		return nil, errors.New("invalid compressed public key length")
	}
	key, err := btcec.ParsePubKey(pubkey)
	if err != nil {
		return nil, err
	}
	return key.ToECDSA(), nil
}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func CompressPubkey(pubkey *ecdsa.PublicKey) []byte {
	var x, y btcec.FieldVal
	x.SetByteSlice(pubkey.X.Bytes())
	y.SetByteSlice(pubkey.Y.Bytes())
	return btcec.NewPublicKey(&x, &y).SerializeCompressed()
}
