//go:build libsecp256k1_sdk
// +build libsecp256k1_sdk

package eth

import (
	fmt "fmt"
	io "io"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1/internal/secp256k1"
	"golang.org/x/crypto/sha3"
)

// Sign signs the provided message using the ECDSA private key. It returns an error if the
// Sign creates a recoverable ECDSA signature on the `secp256k1` curve over the
// provided hash of the message. The produced signature is 65 bytes
// where the last byte contains the recovery ID.
// Sign creates a recoverable ECDSA signature on the secp256k1 curve over the
// provided hash of the message. The produced signature is 65 bytes
// where the last byte contains the recovery ID.
func (privKey PrivKey) Sign(digestBz []byte) ([]byte, error) {
	if len(digestBz) != DigestLength {
		digestBz = Keccak256(digestBz)
	}

	if len(digestBz) != DigestLength {
		return nil, fmt.Errorf("hash is required to be exactly %d bytes (%d)", DigestLength, len(digestBz))
	}

	return secp256k1.Sign(digestBz, privKey.Key)
}

// VerifySignature verifies that the ECDSA public key created a given signature over
// the provided message. The signature should be in [R || S] format.
func (pubKey PubKey) VerifySignature(msg, sig []byte) bool {
	// This is a little hacky, but in order to work around the fact that the Cosmos-SDK typically
	// does not hash messages, we have to accept an unhashed message and hash it.
	// NOTE: this function will not work correctly if a msg of length 32 is provided, that is actually
	// the hash of the message that was signed.
	if len(msg) != DigestLength {
		msg = Keccak256(msg)
	}

	// The signature length must be correct.
	if len(sig) == EthSignatureLength {
		// remove recovery ID (V) if contained in the signature
		sig = sig[:len(sig)-1]
	}

	// the signature needs to be in [R || S] format when provided to VerifySignature
	return secp256k1.VerifySignature(pubKey.Key, msg, sig)
}
