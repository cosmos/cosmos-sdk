package ed25519

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"io"

	"github.com/tendermint/go-amino"
	"golang.org/x/crypto/ed25519"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// -------------------------------------

var _ crypto.PrivKey = PrivKeyEd25519{}

const (
	PrivKeyAminoName = "tendermint/PrivKeyEd25519"
	PubKeyAminoName  = "tendermint/PubKeyEd25519"
	// Size of an Edwards25519 signature. Namely the size of a compressed
	// Edwards25519 point, and a field element. Both of which are 32 bytes.
	SignatureSize = 64
)

var cdc = amino.NewCodec()

func init() {
	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	cdc.RegisterConcrete(PubKeyEd25519{},
		PubKeyAminoName, nil)

	cdc.RegisterInterface((*crypto.PrivKey)(nil), nil)
	cdc.RegisterConcrete(PrivKeyEd25519{},
		PrivKeyAminoName, nil)
}

// PrivKeyEd25519 implements crypto.PrivKey.
type PrivKeyEd25519 [64]byte

// Bytes marshals the privkey using amino encoding.
func (privKey PrivKeyEd25519) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(privKey)
}

// Sign produces a signature on the provided message.
// This assumes the privkey is wellformed in the golang format.
// The first 32 bytes should be random,
// corresponding to the normal ed25519 private key.
// The latter 32 bytes should be the compressed public key.
// If these conditions aren't met, Sign will panic or produce an
// incorrect signature.
func (privKey PrivKeyEd25519) Sign(msg []byte) ([]byte, error) {
	signatureBytes := ed25519.Sign(privKey[:], msg)
	return signatureBytes, nil
}

// PubKey gets the corresponding public key from the private key.
func (privKey PrivKeyEd25519) PubKey() crypto.PubKey {
	privKeyBytes := [64]byte(privKey)
	initialized := false
	// If the latter 32 bytes of the privkey are all zero, compute the pubkey
	// otherwise privkey is initialized and we can use the cached value inside
	// of the private key.
	for _, v := range privKeyBytes[32:] {
		if v != 0 {
			initialized = true
			break
		}
	}

	if !initialized {
		panic("Expected PrivKeyEd25519 to include concatenated pubkey bytes")
	}

	var pubkeyBytes [PubKeyEd25519Size]byte
	copy(pubkeyBytes[:], privKeyBytes[32:])
	return PubKeyEd25519(pubkeyBytes)
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
	return privKeyEd
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
	return privKeyEd
}

// -------------------------------------

var _ crypto.PubKey = PubKeyEd25519{}

// PubKeyEd25519Size is the number of bytes in an Ed25519 signature.
const PubKeyEd25519Size = 32

// PubKeyEd25519 implements crypto.PubKey for the Ed25519 signature scheme.
type PubKeyEd25519 [PubKeyEd25519Size]byte

// Address is the SHA256-20 of the raw pubkey bytes.
func (pubKey PubKeyEd25519) Address() crypto.Address {
	return tmhash.SumTruncated(pubKey[:])
}

// Bytes marshals the PubKey using amino encoding.
func (pubKey PubKeyEd25519) Bytes() []byte {
	bz, err := cdc.MarshalBinaryBare(pubKey)
	if err != nil {
		panic(err)
	}
	return bz
}

func (pubKey PubKeyEd25519) VerifyBytes(msg []byte, sig []byte) bool {
	// make sure we use the same algorithm to sign
	if len(sig) != SignatureSize {
		return false
	}
	return ed25519.Verify(pubKey[:], msg, sig)
}

func (pubKey PubKeyEd25519) String() string {
	return fmt.Sprintf("PubKeyEd25519{%X}", pubKey[:])
}

func (pubKey PubKeyEd25519) Equals(other crypto.PubKey) bool {
	if otherEd, ok := other.(PubKeyEd25519); ok {
		return bytes.Equal(pubKey[:], otherEd[:])
	}

	return false
}
