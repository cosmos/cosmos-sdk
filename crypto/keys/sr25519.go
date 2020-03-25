package keys

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"io"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"

	schnorrkel "github.com/ChainSafe/go-schnorrkel"
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
)

// Address is the SHA256-20 of the raw pubkey bytes.
func (pubKey PubKeySr25519) Address() crypto.Address {
	return crypto.Address(tmhash.SumTruncated(pubKey[:]))
}

// Bytes marshals the PubKey using amino encoding.
func (pubKey PubKeySr25519) Bytes() []byte {
	if len(pubKey.bytes) != PubKeyEd25519Size {
		panic(
			fmt.Errorf("invalid bytes length: got (%s), expected (%d)", len(pubKey.bytes), PubKeySr25519Size),
		)
	}
	return pubKey.bytes
}

func (pubKey PubKeySr25519) VerifyBytes(msg []byte, sig []byte) bool {
	// make sure we use the same algorithm to sign
	if len(sig) != SignatureSize {
		return false
	}
	var sig64 [SignatureSize]byte
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
	return fmt.Sprintf("PubKeySr25519{%X}", pubKey[:])
}

// Equals - checks that two public keys are the same time
// Runs in constant time based on length of the keys.
func (pubKey PubKeySr25519) Equals(other crypto.PubKey) bool {
	if otherEd, ok := other.(PubKeySr25519); ok {
		return bytes.Equal(pubKey[:], otherEd[:])
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
	miniSecretKey, err := schnorrkel.NewMiniSecretKeyFromRaw(privKey)
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

	return PubKeySr25519{bytes: pubkey.Encode()}
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (privKey PrivKeySr25519) Equals(other crypto.PrivKey) bool {
	if otherEd, ok := other.(PrivKeySr25519); ok {
		return subtle.ConstantTimeCompare(privKey[:], otherEd[:]) == 1
	}
	return false
}

// GenPrivKey generates a new sr25519 private key.
// It uses OS randomness in conjunction with the current global random seed
// in tendermint/libs/common to generate the private key.
func GenPrivKey() PrivKeySr25519 {
	return genPrivKey(crypto.CReader())
}

// genPrivKey generates a new sr25519 private key using the provided reader.
func genPrivKey(rand io.Reader) PrivKeySr25519 {
	var seed []byte

	out := make([]byte, 64)
	_, err := io.ReadFull(rand, out)
	if err != nil {
		panic(err)
	}

	copy(seed[:], out)

	return PrivKeySr25519{bytes: schnorrkel.NewMiniSecretKey(seed).ExpandEd25519().Encode()}
}

// GenPrivKeyFromSecret hashes the secret with SHA2, and uses
// that 32 byte output to create the private key.
// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeyFromSecret(secret []byte) PrivKeySr25519 {
	seed := crypto.Sha256(secret) // Not Ripemd160 because we want 32 bytes.
	var bz [PrivKeySr25519Size]byte
	copy(bz[:], seed)
	privKey, _ := schnorrkel.NewMiniSecretKeyFromRaw(bz)
	return PrivKeySr25519{bytes: privKey.ExpandEd25519().Encode()}
}
