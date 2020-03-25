package keys

import (
	"bytes"
	"crypto/subtle"
	"fmt"

	"golang.org/x/crypto/ed25519"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

const (
	PubKeyEd25519Name  = "tendermint/PubKeyEd25519"
	PrivKeyEd25519Name = "tendermint/PrivKeyEd25519"
)

var (
	_ crypto.PubKey  = PubKeyEd25519{}
	_ crypto.PrivKey = PrivKeyEd25519{}
)

const (
	// PubKeyEd25519Size is the number of bytes in an Ed25519 public key.
	PubKeyEd25519Size = 32
	// PrivKeyEd25519Size is the number of bytes in an Ed25519 private key.
	PrivKeyEd25519Size = 64
	// SignatureEd25519Size of an Edwards25519 signature. Namely the size of a compressed
	// Edwards25519 point, and a field element. Both of which are 32 bytes.
	SignatureEd25519Size = 64
)

//-------------------------------------

// Address is the SHA256-20 of the raw pubkey bytes.
func (pubKey PubKeyEd25519) Address() crypto.Address {
	return crypto.Address(tmhash.SumTruncated(pubKey.Bytes()))
}

// Bytes marshals the PubKey using amino encoding.
func (pubKey PubKeyEd25519) Bytes() []byte {
	if len(pubKey.bytes) != PubKeyEd25519Size {
		panic(
			fmt.Errorf("invalid bytes length: got (%s), expected (%d)", len(pubKey.bytes), PubKeyEd25519Size),
		)
	}
	return pubKey.bytes[:]
}

func (pubKey PubKeyEd25519) VerifyBytes(msg []byte, sig []byte) bool {
	// make sure we use the same algorithm to sign
	if len(sig) != SignatureEd25519Size {
		return false
	}
	return ed25519.Verify(pubKey.Bytes()[:], msg, sig)
}

func (pubKey PubKeyEd25519) String() string {
	return fmt.Sprintf("%s{%X}", PubKeyEd25519Name, pubKey.Bytes()[:])
}

// nolint: golint
func (pubKey PubKeyEd25519) Equals(other crypto.PubKey) bool {
	if otherEd, ok := other.(PubKeyEd25519); ok {
		return bytes.Equal(pubKey.bytes, otherEd.bytes)
	}

	return false
}

//-------------------------------------

// Bytes marshals the privkey using amino encoding.
func (privKey PrivKeyEd25519) Bytes() []byte {
	return privKey.bytes
}

// Sign produces a signature on the provided message.
// This assumes the privkey is wellformed in the golang format.
// The first 32 bytes should be random,
// corresponding to the normal ed25519 private key.
// The latter 32 bytes should be the compressed public key.
// If these conditions aren't met, Sign will panic or produce an
// incorrect signature.
func (privKey PrivKeyEd25519) Sign(msg []byte) ([]byte, error) {
	signatureBytes := ed25519.Sign(privKey.Bytes(), msg)
	return signatureBytes, nil
}

// PubKey gets the corresponding public key from the private key.
func (privKey PrivKeyEd25519) PubKey() crypto.PubKey {
	initialized := false
	// If the latter 32 bytes of the privkey are all zero, compute the pubkey
	// otherwise privkey is initialized and we can use the cached value inside
	// of the private key.
	for _, v := range privKey.Bytes()[32:] {
		if v != 0 {
			initialized = true
			break
		}
	}

	if !initialized {
		panic("expected PrivKeyEd25519 to include concatenated pubkey bytes")
	}

	var pubkeyBytes []byte
	copy(pubkeyBytes[:32], privKey.Bytes()[32:])
	return PubKeyEd25519{bytes: pubkeyBytes}
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (privKey PrivKeyEd25519) Equals(other crypto.PrivKey) bool {
	if otherEd, ok := other.(PrivKeyEd25519); ok {
		return subtle.ConstantTimeCompare(privKey.bytes[:], otherEd.bytes[:]) == 1
	}

	return false
}
