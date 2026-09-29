// Package mldsa65 wraps the cometbft ML-DSA-65 key implementation so it
// satisfies the cosmos-sdk crypto/types interfaces and can be registered with
// the SDK codecs, keyring, and multisig machinery.
//
// The on-wire encoding (PubKey / PrivKey proto messages, defined in
// proto/cosmos/crypto/mldsa65/keys.proto) is a single `bytes key` field
// holding the FIPS 204 packed key.
package mldsa65

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/crypto"
	mldsa "github.com/cometbft/cometbft/crypto/mldsa65"
	"github.com/cometbft/cometbft/crypto/tmhash"

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

// NewPrivateKeyFromBytes validates and wraps the given private key bytes.
func NewPrivateKeyFromBytes(bz []byte) (PrivKey, error) {
	sk, err := mldsa.NewPrivKeyFromBytes(bz)
	if err != nil {
		return PrivKey{}, err
	}
	return PrivKey{Key: sk.Bytes()}, nil
}

// GenPrivKey generates a fresh ML-DSA-65 private key using OS randomness.
func GenPrivKey() (PrivKey, error) {
	sk, err := mldsa.GenPrivKey()
	if err != nil {
		return PrivKey{}, err
	}
	return PrivKey{Key: sk.Bytes()}, nil
}

// GenPrivKeyFromSeed deterministically derives an ML-DSA-65 private key from a
// 32-byte seed (mldsa.SeedSize). The same seed always yields the same key,
// which is what makes mnemonic-based account recovery possible.
func GenPrivKeyFromSeed(seed []byte) (PrivKey, error) {
	sk, err := mldsa.GenPrivKeyFromSeed(seed)
	if err != nil {
		return PrivKey{}, err
	}
	return PrivKey{Key: sk.Bytes()}, nil
}

// Bytes returns the serialized private key bytes.
func (privKey PrivKey) Bytes() []byte {
	return privKey.Key
}

// PubKey returns the corresponding public key. Returns nil if the underlying
// private key bytes cannot be parsed.
func (privKey PrivKey) PubKey() cryptotypes.PubKey {
	sk, err := mldsa.NewPrivKeyFromBytes(privKey.Key)
	if err != nil {
		return nil
	}
	pk, ok := sk.PubKey().(mldsa.PubKey)
	if !ok {
		return nil
	}
	return &PubKey{Key: pk.Bytes()}
}

// Equals returns true if the other key is also ML-DSA-65 and the bytes match.
func (privKey PrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	return privKey.Type() == other.Type() && bytes.Equal(privKey.Bytes(), other.Bytes())
}

// Type returns the algorithm identifier.
func (PrivKey) Type() string {
	return mldsa.KeyType
}

// Sign produces a deterministic ML-DSA-65 signature over msg.
func (privKey PrivKey) Sign(msg []byte) ([]byte, error) {
	sk, err := mldsa.NewPrivKeyFromBytes(privKey.Key)
	if err != nil {
		return nil, err
	}
	return sk.Sign(msg)
}

// MarshalAmino overrides Amino binary marshaling.
func (privKey PrivKey) MarshalAmino() ([]byte, error) {
	return privKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (privKey *PrivKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != mldsa.PrivKeySize {
		return errors.New("invalid mldsa65 privkey size")
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

// Address returns the account/validator address: SHA256(pubkey) truncated to 20
// bytes, matching the convention used by ed25519 keys. This scheme is valid for
// both CometBFT validator addresses and SDK account addresses; it MUST NOT change
// once accounts exist, as that would be a breaking change to derived addresses.
func (pubKey PubKey) Address() crypto.Address {
	if len(pubKey.Key) != mldsa.PubKeySize {
		panic(fmt.Sprintf("length of pubkey is incorrect, got: %d expected: %d", len(pubKey.Key), mldsa.PubKeySize))
	}
	// FIPS 204 packed encoding round-trips, so hashing the stored bytes matches
	// what re-parsing then hashing would produce, without the expensive unpack.
	// Mirrors CometBFT's mldsa65.PubKey.Address and the ed25519/secp256k1 keys.
	return crypto.Address(tmhash.SumTruncated(pubKey.Key))
}

// VerifySignature verifies the given signature against msg.
func (pubKey PubKey) VerifySignature(msg, sig []byte) bool {
	if len(sig) != mldsa.SignatureSize {
		return false
	}
	pk, err := mldsa.NewPubKeyFromBytes(pubKey.Key)
	if err != nil {
		return false
	}
	return pk.VerifySignature(msg, sig)
}

// Bytes returns the serialized public key bytes.
func (pubKey PubKey) Bytes() []byte {
	return pubKey.Key
}

// Type returns the algorithm identifier.
func (PubKey) Type() string {
	return mldsa.KeyType
}

// Equals returns true if the other key is also ML-DSA-65 and the bytes match.
func (pubKey PubKey) Equals(other cryptotypes.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}

// String returns the hex representation of the public key.
func (pubKey PubKey) String() string {
	return fmt.Sprintf("PubKeyMlDsa65{%X}", pubKey.Key)
}

// MarshalAmino overrides Amino binary marshaling.
func (pubKey PubKey) MarshalAmino() ([]byte, error) {
	return pubKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (pubKey *PubKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != mldsa.PubKeySize {
		return errors.New("invalid mldsa65 pubkey size")
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
