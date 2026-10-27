// Package secp256k1eth wraps the cometbft secp256k1eth key implementation so
// it satisfies the cosmos-sdk crypto/types interfaces and can be registered
// with the SDK codecs, enabling secp256k1eth consensus/validator keys.
//
// The on-wire encoding (PubKey / PrivKey proto messages, defined in
// proto/cosmos/crypto/secp256k1eth/keys.proto) is a single `bytes key` field
// holding the compressed SEC1 public key or the 32-byte private scalar.
package secp256k1eth

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/crypto"
	cmtsecp256k1eth "github.com/cometbft/cometbft/crypto/secp256k1eth"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// ===============================================================================================
// Private Key
// ===============================================================================================

var (
	_ cryptotypes.PrivKey  = &PrivKey{}
	_ codec.AminoMarshaler = &PrivKey{}
)

// NewPrivateKeyFromBytes validates and wraps the given 32-byte private scalar.
func NewPrivateKeyFromBytes(bz []byte) (PrivKey, error) {
	if len(bz) != cmtsecp256k1eth.PrivKeySize {
		return PrivKey{}, fmt.Errorf(
			"secp256k1eth: invalid private key size, expected %d bytes, got %d",
			cmtsecp256k1eth.PrivKeySize, len(bz),
		)
	}
	key := make([]byte, cmtsecp256k1eth.PrivKeySize)
	copy(key, bz)
	return PrivKey{Key: key}, nil
}

// GenPrivKey generates a fresh secp256k1eth private key using OS randomness.
func GenPrivKey() PrivKey {
	return PrivKey{Key: cmtsecp256k1eth.GenPrivKey()}
}

// GenPrivKeyFromSecret deterministically derives a key from secret bytes,
// matching cometbft's GenPrivKeySecp256k1Eth. It exists for reproducible
// test/e2e keys — it is NOT HD/mnemonic derivation.
func GenPrivKeyFromSecret(secret []byte) PrivKey {
	return PrivKey{Key: cmtsecp256k1eth.GenPrivKeySecp256k1Eth(secret)}
}

// Bytes returns the raw 32-byte private scalar.
func (privKey PrivKey) Bytes() []byte {
	return privKey.Key
}

// PubKey returns the corresponding compressed SEC1 public key.
func (privKey PrivKey) PubKey() cryptotypes.PubKey {
	return &PubKey{Key: cmtsecp256k1eth.PrivKey(privKey.Key).PubKey().Bytes()}
}

// Equals returns true if the other key is also secp256k1eth and the bytes match.
func (privKey PrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	return privKey.Type() == other.Type() && bytes.Equal(privKey.Bytes(), other.Bytes())
}

// Type returns the algorithm identifier.
func (PrivKey) Type() string {
	return cmtsecp256k1eth.KeyType
}

// Sign produces a 65-byte go-ethereum signature [R || S || V] (V in {0,1})
// over the legacy Keccak-256 hash of msg, in canonical lower-S form.
func (privKey PrivKey) Sign(msg []byte) ([]byte, error) {
	return cmtsecp256k1eth.PrivKey(privKey.Key).Sign(msg)
}

// MarshalAmino overrides Amino binary marshaling.
func (privKey PrivKey) MarshalAmino() ([]byte, error) {
	return privKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (privKey *PrivKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != cmtsecp256k1eth.PrivKeySize {
		return errors.New("invalid secp256k1eth privkey size")
	}
	privKey.Key = bz
	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (privKey PrivKey) MarshalAminoJSON() ([]byte, error) {
	return privKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (privKey *PrivKey) UnmarshalAminoJSON(bz []byte) error {
	return privKey.UnmarshalAmino(bz)
}

// ===============================================================================================
// Public Key
// ===============================================================================================

var (
	_ cryptotypes.PubKey   = &PubKey{}
	_ codec.AminoMarshaler = &PubKey{}
)

// NewPubKeyFromBytes validates the bytes are a compressed secp256k1 point and
// wraps them.
func NewPubKeyFromBytes(bz []byte) (PubKey, error) {
	pk, err := cmtsecp256k1eth.NewPubKeyFromBytes(bz)
	if err != nil {
		return PubKey{}, err
	}
	return PubKey{Key: pk.Bytes()}, nil
}

// Address returns the 20-byte Ethereum address
// Keccak256(uncompressedPubKey[1:])[12:], matching cometbft's secp256k1eth
// validator address. This is deliberately NOT ADR-28 account derivation: the
// key is for validator/consensus use, and slashing/evidence look validators up
// by the address cometbft reports.
func (pubKey PubKey) Address() crypto.Address {
	pk, err := cmtsecp256k1eth.NewPubKeyFromBytes(pubKey.Key)
	if err != nil {
		panic(fmt.Sprintf("secp256k1eth pubkey: %v", err))
	}
	return pk.Address()
}

// VerifySignature verifies a 65-byte [R || S || V] go-ethereum signature over
// the legacy Keccak-256 hash of msg, rejecting malleable (non-lower-S) forms.
func (pubKey PubKey) VerifySignature(msg, sig []byte) bool {
	return cmtsecp256k1eth.PubKey(pubKey.Key).VerifySignature(msg, sig)
}

// Bytes returns the compressed SEC1 public key bytes.
func (pubKey PubKey) Bytes() []byte {
	return pubKey.Key
}

// Type returns the algorithm identifier.
func (PubKey) Type() string {
	return cmtsecp256k1eth.KeyType
}

// Equals returns true if the other key is also secp256k1eth and the bytes match.
func (pubKey PubKey) Equals(other cryptotypes.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}

// String returns the hex representation of the public key.
func (pubKey PubKey) String() string {
	return fmt.Sprintf("PubKeySecp256k1eth{%X}", pubKey.Key)
}

// MarshalAmino overrides Amino binary marshaling.
func (pubKey PubKey) MarshalAmino() ([]byte, error) {
	return pubKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (pubKey *PubKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != cmtsecp256k1eth.PubKeySize {
		return errors.New("invalid secp256k1eth pubkey size")
	}
	pubKey.Key = bz
	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey PubKey) MarshalAminoJSON() ([]byte, error) {
	return pubKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey *PubKey) UnmarshalAminoJSON(bz []byte) error {
	return pubKey.UnmarshalAmino(bz)
}
