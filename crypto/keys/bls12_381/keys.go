package bls12_381

import (
	"bytes"
	"crypto/sha256"
	fmt "fmt"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"

	bls12381 "github.com/cosmos/crypto/curves/bls12381"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

const (
	PrivKeyName = "tendermint/PrivKeyEd25519"
	PubKeyName  = "tendermint/PubKeyEd25519"
	// PubKeySize is the size, in bytes, of public keys as used in this package.
	PubKeySize = 32
	// PrivKeySize is the size, in bytes, of private keys as used in this package.
	PrivKeySize = 64
	// SignatureSize the size of an Edwards25519 signature. Namely the size of a compressed
	// Edwards25519 point, and a field element. Both of which are 32 bytes.
	SignatureSize = 64
	// SeedSize is the size, in bytes, of private key seeds. These are the
	// private key representations used by RFC 8032.
	SeedSize = 32

	keyType = "bls12381"
)

// ===============================================================================================
// Private Key
// ===============================================================================================

// PrivKey is a wrapper around the Ethereum BLS12-381 private key type. This
// wrapper conforms to crypto.Pubkey to allow for the use of the Ethereum
// BLS12-381 private key type.

var (
	_ cryptotypes.PrivKey  = &PrivKey{}
	_ codec.AminoMarshaler = &PrivKey{}
)

// NewPrivateKeyFromBytes build a new key from the given bytes.
func NewPrivateKeyFromBytes(bz []byte) (PrivKey, error) {
	secretKey, err := bls12381.SecretKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}
	return secretKey.Marshal(), nil
}

// GenPrivKey generates a new key.
func GenPrivKey() (PrivKey, error) {
	secretKey, err := bls12381.RandKey()
	return PrivKey(secretKey.Marshal()), err
}

// Bytes returns the byte representation of the Key.
func (privKey PrivKey) Bytes() []byte {
	return privKey
}

// PubKey returns the private key's public key. If the privkey is not valid
// it returns a nil value.
func (privKey PrivKey) PubKey() cryptotypes.PubKey {
	secretKey, err := bls12381.SecretKeyFromBytes(privKey)
	if err != nil {
		return nil
	}

	return PubKey(secretKey.PublicKey().Marshal())
}

// Equals returns true if two keys are equal and false otherwise.
func (privKey PrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	return privKey.Type() == other.Type() && bytes.Equal(privKey.Bytes(), other.Bytes())
}

// Type returns the type.
func (PrivKey) Type() string {
	return KeyType
}

// Sign signs the given byte array. If msg is larger than
// MaxMsgLen, SHA256 sum will be signed instead of the raw bytes.
func (privKey PrivKey) Sign(msg []byte) ([]byte, error) {
	secretKey, err := bls12381.SecretKeyFromBytes(privKey)
	if err != nil {
		return nil, err
	}

	if len(msg) > MaxMsgLen {
		hash := sha256.Sum256(msg)
		sig := secretKey.Sign(hash[:])
		return sig.Marshal(), nil
	}
	sig := secretKey.Sign(msg)
	return sig.Marshal(), nil
}

// ===============================================================================================
// Public Key
// ===============================================================================================

// Pubkey is a wrapper around the Ethereum BLS12-381 public key type. This
// wrapper conforms to crypto.Pubkey to allow for the use of the Ethereum
// BLS12-381 public key type.

var _ cryptotypes.PubKey = &PubKey{}

// Address returns the address of the key.
//
// The function will panic if the public key is invalid.
func (pubKey PubKey) Address() crypto.Address {
	pk, _ := bls12381.PublicKeyFromBytes(pubKey)
	if len(pk.Marshal()) != PubKeySize {
		panic("pubkey is incorrect size")
	}
	return crypto.Address(tmhash.SumTruncated(pubKey))
}

// VerifySignature verifies the given signature.
func (pubKey PubKey) VerifySignature(msg, sig []byte) bool {
	if len(sig) != SignatureLength {
		return false
	}

	pubK, err := bls12381.PublicKeyFromBytes(pubKey)
	if err != nil { // invalid pubkey
		return false
	}

	if len(msg) > MaxMsgLen {
		hash := sha256.Sum256(msg)
		msg = hash[:]
	}

	ok, err := bls12381.VerifySignature(sig, [MaxMsgLen]byte(msg[:MaxMsgLen]), pubK)
	if err != nil { // bad signature
		return false
	}

	return ok
}

// Bytes returns the byte format.
func (pubKey PubKey) Bytes() []byte {
	return pubKey
}

// Type returns the key's type.
func (PubKey) Type() string {
	return KeyType
}

// Equals returns true if the other's type is the same and their bytes are deeply equal.
func (pubKey PubKey) Equals(other cryptotypes.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}

// String returns Hex representation of a pubkey with it's type
func (pubKey PubKey) String() string {
	return fmt.Sprintf("PubKeyBLS12_381{%X}", pubKey.Key)
}
