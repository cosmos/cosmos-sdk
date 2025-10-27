//go:build ((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64)) && bls12381

package bls12_381

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/bls12381"
	"github.com/cometbft/cometbft/crypto/tmhash"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
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
	secretKey, err := bls12381.NewPrivateKeyFromBytes(bz)
	if err != nil {
		return PrivKey{}, err
	}
	return PrivKey{
		Key: secretKey.Bytes(),
	}, nil
}

// GenPrivKey generates a new key.
func GenPrivKey() (PrivKey, error) {
	secretKey, err := bls12381.GenPrivKey()
	return PrivKey{
		Key: secretKey.Bytes(),
	}, err
}

// Bytes returns the byte representation of the Key.
func (privKey PrivKey) Bytes() []byte {
	return privKey.Key
}

// PubKey returns the private key's public key. If the privkey is not valid
// it returns a nil value.
func (privKey PrivKey) PubKey() cryptotypes.PubKey {
	secretKey, err := bls12381.NewPrivateKeyFromBytes(privKey.Key)
	if err != nil {
		return nil
	}

	return &PubKey{
		Key: secretKey.PubKey().Bytes(),
	}
}

// Equals returns true if two keys are equal and false otherwise.
func (privKey PrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	return privKey.Type() == other.Type() && bytes.Equal(privKey.Bytes(), other.Bytes())
}

// Type returns the type.
func (PrivKey) Type() string {
	return bls12381.KeyType
}

// Sign signs the given byte array. If msg is larger than
// MaxMsgLen, SHA256 sum will be signed instead of the raw bytes.
func (privKey PrivKey) Sign(msg []byte) ([]byte, error) {
	secretKey, err := bls12381.NewPrivateKeyFromBytes(privKey.Key)
	if err != nil {
		return nil, err
	}

	return secretKey.Sign(msg)
}

// MarshalAmino overrides Amino binary marshaling.
func (privKey PrivKey) MarshalAmino() ([]byte, error) {
	return privKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (privKey *PrivKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != bls12381.PrivKeySize {
		return errors.New("invalid privkey size")
	}
	privKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (privKey PrivKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return privKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (privKey *PrivKey) UnmarshalAminoJSON(bz []byte) error {
	return privKey.UnmarshalAmino(bz)
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
	pk, _ := bls12381.NewPublicKeyFromBytes(pubKey.Key)
	if len(pk.Bytes()) != bls12381.PubKeySize {
		panic("pubkey is incorrect size")
	}
	return crypto.Address(tmhash.SumTruncated(pubKey.Key))
}

// VerifySignature verifies the given signature.
func (pubKey PubKey) VerifySignature(msg, sig []byte) bool {
	if len(sig) != bls12381.SignatureLength {
		return false
	}

	pubK, err := bls12381.NewPublicKeyFromBytes(pubKey.Key)
	if err != nil { // invalid pubkey
		return false
	}

	return pubK.VerifySignature(msg, sig)
}

// Bytes returns the byte format.
func (pubKey PubKey) Bytes() []byte {
	return pubKey.Key
}

// Type returns the key's type.
func (PubKey) Type() string {
	return bls12381.KeyType
}

// Equals returns true if the other's type is the same and their bytes are deeply equal.
func (pubKey PubKey) Equals(other cryptotypes.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}

// String returns Hex representation of a pubkey with it's type
func (pubKey PubKey) String() string {
	return fmt.Sprintf("PubKeyBLS12_381{%X}", pubKey.Key)
}
