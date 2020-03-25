package keys

import (
	"bytes"
	"crypto/subtle"
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"

	schnorrkel "github.com/ChainSafe/go-schnorrkel"
)

const (
	PubKeySr25519Name  = "tendermint/PubKeySr25519"
	PrivKeySr25519Name = "tendermint/PrivKeySr25519"
)

var (
	_ crypto.PubKey  = PubKeySr25519{}
	_ crypto.PrivKey = PrivKeySr25519{}
)

const (
	// PubKeySr25519Size is the number of bytes in an Sr25519 public key.
	PubKeySr25519Size = 32
	// PrivKeySr25519Size is the number of bytes in an Sr25519 private key.
	PrivKeySr25519Size = 32
	// SignatureSr25519Size is the size of an Sr25519 signature. Namely the size of a compressed
	// Sr25519 point, and a field element. Both of which are 32 bytes.
	SignatureSr25519Size = 64
)

// Address is the SHA256-20 of the raw pubkey bytes.
func (pubKey PubKeySr25519) Address() crypto.Address {
	return crypto.Address(tmhash.SumTruncated(pubKey.Bytes()[:]))
}

// Bytes marshals the PubKey using amino encoding.
func (pubKey PubKeySr25519) Bytes() []byte {
	if len(pubKey.bytes) != PubKeyEd25519Size {
		panic(
			fmt.Errorf("invalid bytes length: got (%s), expected (%d)", len(pubKey.bytes), PubKeySr25519Size),
		)
	}
	return pubKey.bytes[:PubKeySr25519Size]
}

func (pubKey PubKeySr25519) VerifyBytes(msg []byte, sig []byte) bool {
	// make sure we use the same algorithm to sign
	if len(sig) != SignatureSr25519Size {
		return false
	}
	var sig64 [SignatureSr25519Size]byte
	copy(sig64[:], sig)

	publicKey := &(schnorrkel.PublicKey{})
	err := publicKey.Decode(pubKey)
	if err != nil {
		return false
	}

	signingContext := schnorrkel.NewSigningContext([]byte{}, msg)

	signature := &(schnorrkel.Signature{})
	err = signature.Decode(sig64)
	if err != nil {
		return false
	}

	return publicKey.Verify(signature, signingContext)
}

func (pubKey PubKeySr25519) String() string {
	return fmt.Sprintf("%s{%X}", PubKeySr25519Name, pubKey.Bytes()[:])
}

// Equals - checks that two public keys are the same time
// Runs in constant time based on length of the keys.
func (pubKey PubKeySr25519) Equals(other crypto.PubKey) bool {
	if otherEd, ok := other.(PubKeySr25519); ok {
		return bytes.Equal(pubKey.bytes[:], otherEd.bytes[:])
	}
	return false
}

// Bytes marshals the privkey using amino encoding.
func (privKey PrivKeySr25519) Bytes() []byte {
	if len(privKey.bytes) != PubKeyEd25519Size {
		panic(
			fmt.Errorf("invalid bytes length: got (%s), expected (%d)", len(privKey.bytes), PrivKeySr25519Size),
		)
	}
	return privKey.bytes
}

// Sign produces a signature on the provided message.
func (privKey PrivKeySr25519) Sign(msg []byte) ([]byte, error) {

	miniSecretKey, err := schnorrkel.NewMiniSecretKeyFromRaw(privKey.Bytes())
	if err != nil {
		return []byte{}, err
	}
	secretKey := miniSecretKey.ExpandEd25519()

	signingContext := schnorrkel.NewSigningContext([]byte{}, msg)

	sig, err := secretKey.Sign(signingContext)
	if err != nil {
		return []byte{}, err
	}

	sigBytes := sig.Encode()
	return sigBytes[:], nil
}

// PubKey gets the corresponding public key from the private key.
func (privKey PrivKeySr25519) PubKey() crypto.PubKey {
	miniSecretKey, err := schnorrkel.NewMiniSecretKeyFromRaw(privKey)
	if err != nil {
		panic(fmt.Errorf("invalid private key: %w", err))
	}
	secretKey := miniSecretKey.ExpandEd25519()

	pubkey, err := secretKey.Public()
	if err != nil {
		panic(fmt.Errorf("could not generate public key: %w", err))
	}

	pubKeySr := pubkey.Encode()
	return PubKeySr25519{bytes: pubKeySr[:]}
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (privKey PrivKeySr25519) Equals(other crypto.PrivKey) bool {
	if otherEd, ok := other.(PrivKeySr25519); ok {
		return subtle.ConstantTimeCompare(privKey.bytes[:], otherEd.bytes[:]) == 1
	}
	return false
}
