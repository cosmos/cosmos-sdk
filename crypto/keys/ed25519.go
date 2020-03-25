package keys

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"io"

	"golang.org/x/crypto/ed25519"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

const (
	PrivKeyEd25519Name = "tendermint/PrivKeyEd25519"
	PubKeyEd25519Name  = "tendermint/PubKeyEd25519"
)

var (
	_ crypto.PubKey = PubKeyEd25519{}
	_ crypto.PrivKey = PrivKeyEd25519{}
)

const (
	// PubKeyEd25519Size is the number of bytes in an Ed25519 signature.
	PubKeyEd25519Size = 32
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
			fmt.Errorf("invalid bytes length: got (%s), expected (%d)",  len(pubKey.bytes), PubKeyEd25519Size),
		)
	}
	return pubkey.bytes[]
}

func (pubKey PubKeyEd25519) VerifyBytes(msg []byte, sig []byte) bool {
	// make sure we use the same algorithm to sign
	if len(sig) != SignatureEd25519Size {
		return false
	}
	return ed25519.Verify(pubKey[:], msg, sig)
}

func (pubKey PubKeyEd25519) String() string {
	return fmt.Sprintf("%s{%X}", PubKeyEd25519Name, pubKey[:])
}

// nolint: golint
func (pubKey PubKeyEd25519) Equals(other crypto.PubKey) bool {
	if otherEd, ok := other.(PubKeyEd25519); ok {
		return bytes.Equal(pubKey[:], otherEd[:])
	}

	return false
}

//-------------------------------------

// Bytes marshals the privkey using amino encoding.
func (privKey PrivKeyEd25519) Bytes() []byte {
	return privkey.bytes
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

	var pubkeyBytes [PubKeyEd25519Size]byte
	copy(pubkeyBytes[:], privKey.Bytes()[32:])
	return PubKeyEd25519{bytes: pubkeyBytes}
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (privKey PrivKeyEd25519) Equals(other crypto.PrivKey) bool {
	if otherEd, ok := other.(PrivKeyEd25519); ok {
		return subtle.ConstantTimeCompare(privKey[:], otherEd[:]) == 1
	}

	return false
}

// GenPrivKey generates a new ed25519 private key.
// It uses OS randomness in conjunction with the current global random seed
// in tendermint/libs/common to generate the private key.
func GenPrivKey() PrivKeyEd25519 {
	return genPrivKey(crypto.CReader())
}

// genPrivKey generates a new ed25519 private key using the provided reader.
func genPrivKey(rand io.Reader) PrivKeyEd25519 {
	seed := make([]byte, 32)
	_, err := io.ReadFull(rand, seed)
	if err != nil {
		panic(err)
	}

	privKey := ed25519.NewKeyFromSeed(seed)
	var privKeyEd PrivKeyEd25519
	copy(privKeyEd[:], privKey)
	return PrivKeyEd25519{bytes: privKeyEd}
}

// GenPrivKeyFromSecret hashes the secret with SHA2, and uses
// that 32 byte output to create the private key.
// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeyFromSecret(secret []byte) PrivKeyEd25519 {
	seed := crypto.Sha256(secret) // Not Ripemd160 because we want 32 bytes.

	privKey := ed25519.NewKeyFromSeed(seed)
	var privKeyEd PrivKeyEd25519
	copy(privKeyEd[:], privKey)
	return PrivKeyEd25519{bytes: privKeyEd}
}
