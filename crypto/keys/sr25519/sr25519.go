package sr25519

import (
	"bytes"
	"crypto/subtle"
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	tmsr25519 "github.com/tendermint/tendermint/crypto/sr25519"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var _ cryptotypes.PrivKey = &PrivKey{}
var _ codec.AminoMarshaler = &PrivKey{}

const (
	PrivKeySize = 32
	keyType     = "sr25519"
	PrivKeyName = "tendermint/PrivKeySr25519"
	PubKeyName  = "tendermint/PubKeySr25519"
)

// Bytes returns the byte representation of the Private Key.
func (privKey *PrivKey) Bytes() []byte {
	return privKey.Key
}

// PubKey gets the corresponding public key from the private key.
func (privKey *PrivKey) PubKey() cryptotypes.PubKey {
	return &PubKey{Key: tmsr25519.PrivKey(privKey.Key).PubKey().Bytes()}
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the
func (privKey *PrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	return privKey.Type() == other.Type() && subtle.ConstantTimeCompare(privKey.Bytes(), other.Bytes()) == 1
}

func (privKey *PrivKey) Type() string {
	return keyType
}

// MarshalAmino overrides Amino binary marshalling.
func (privKey PrivKey) MarshalAmino() ([]byte, error) {
	return privKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshalling.
func (privKey *PrivKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != PrivKeySize {
		return fmt.Errorf("invalid privkey size")
	}
	privKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshalling.
func (privKey PrivKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return privKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshalling.
func (privKey *PrivKey) UnmarshalAminoJSON(bz []byte) error {
	return privKey.UnmarshalAmino(bz)
}

// GenPrivKey generates a new sr25519 private key.
// It uses OS randomness in conjunction with the current global random seed
// in tendermint/libs/common to generate the private key.
func GenPrivKey() *PrivKey {
	return &PrivKey{Key: tmsr25519.GenPrivKey().Bytes()}
}

// GenPrivKeyFromSecret hashes the secret with SHA2, and uses
// that 32 byte output to create the private key.
// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeyFromSecret(secret []byte) *PrivKey {
	return &PrivKey{Key: tmsr25519.GenPrivKeyFromSecret(secret).Bytes()}
}

//-------------------------------------

var _ cryptotypes.PubKey = &PubKey{}
var _ codec.AminoMarshaler = &PubKey{}

// PubKeySize is comprised of 32 bytes for one field element
// (the x-coordinate), plus one byte for the parity of the y-coordinate.
const PubKeySize = 33

// Address returns a Bitcoin style addresses: RIPEMD160(SHA256(pubkey))
func (pubKey *PubKey) Address() crypto.Address {
	return tmsr25519.PubKey(pubKey.Key).Address()
}

// Bytes returns the pubkey byte format.
func (pubKey *PubKey) Bytes() []byte {
	return pubKey.Key
}

func (pubKey *PubKey) String() string {
	return fmt.Sprintf("PubKeySr25519{%X}", pubKey.Key)
}

func (pubKey *PubKey) Type() string {
	return keyType
}

func (pubKey *PubKey) Equals(other cryptotypes.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}

// MarshalAmino overrides Amino binary marshalling.
func (pubKey PubKey) MarshalAmino() ([]byte, error) {
	return pubKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshalling.
func (pubKey *PubKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != PubKeySize {
		return errors.Wrap(errors.ErrInvalidPubKey, "invalid pubkey size")
	}
	pubKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshalling.
func (pubKey PubKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return pubKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshalling.
func (pubKey *PubKey) UnmarshalAminoJSON(bz []byte) error {
	return pubKey.UnmarshalAmino(bz)
}
