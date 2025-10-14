package mldsa

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cloudflare/circl/sign/mldsa/mldsa44"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"golang.org/x/crypto/ripemd160"
)

var (
	_ cryptotypes.PrivKey  = &PrivKey{}
	_ codec.AminoMarshaler = &PrivKey{}
)

const (
	keyType     = "mldsa44"
	PrivKeyName = "tendermint/PrivKeyMLDSA"
	PubKeyName  = "tendermint/PubKeyMLDSA"
)

func (privKey *PrivKey) Sign(msg []byte) ([]byte, error) {
	scheme := mldsa44.Scheme()
	privateKey, err := scheme.UnmarshalBinaryPrivateKey(privKey.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
	}
	return privateKey.Sign(rand.Reader, msg, crypto.Hash(0))
}

// Bytes returns the byte representation of the Private Key.
func (privKey *PrivKey) Bytes() []byte {
	return privKey.Key
}

// PubKey performs the point-scalar multiplication from the privKey on the
// generator point to get the pubkey.
func (privKey *PrivKey) PubKey() cryptotypes.PubKey {
	scheme := mldsa44.Scheme()
	privateKey, err := scheme.UnmarshalBinaryPrivateKey(privKey.Key)
	if err != nil {
		return nil
	}
	publicKey := privateKey.Public().(*mldsa44.PublicKey).Bytes()
	return &PubKey{Key: publicKey}
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the
func (privKey *PrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	return privKey.Type() == other.Type() && subtle.ConstantTimeCompare(privKey.Bytes(), other.Bytes()) == 1
}

func (privKey *PrivKey) Type() string {
	return keyType
}

// MarshalAmino overrides Amino binary marshaling.
func (privKey PrivKey) MarshalAmino() ([]byte, error) {
	return privKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (privKey *PrivKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != mldsa44.PrivateKeySize {
		return fmt.Errorf("invalid privkey size")
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

var (
	_ cryptotypes.PubKey   = &PubKey{}
	_ codec.AminoMarshaler = &PubKey{}
)

func (m *PubKey) String() string {
	return fmt.Sprintf("PubKeyMldsa44{%X}", m.Key)
}

func (m *PubKey) Address() cryptotypes.Address {
	sha := sha256.Sum256(m.Key)
	hasherRIPEMD160 := ripemd160.New() //nolint:gosec // keeping around for backwards compatibility
	hasherRIPEMD160.Write(sha[:])      // does not error
	return hasherRIPEMD160.Sum(nil)
}

func (m *PubKey) Bytes() []byte {
	return m.Key
}

func (m *PubKey) VerifySignature(msg, sig []byte) bool {
	if len(sig) != mldsa44.SignatureSize {
		return false
	}
	scheme := mldsa44.Scheme()
	publicKey, err := scheme.UnmarshalBinaryPublicKey(m.Key)
	if err != nil {
		return false
	}
	return scheme.Verify(publicKey, msg, sig, nil)
}

func (m *PubKey) Equals(key cryptotypes.PubKey) bool {
	return m.Type() == key.Type() && bytes.Equal(m.Bytes(), key.Bytes())
}

func (m *PubKey) Type() string {
	return keyType
}

func (m *PubKey) MarshalAmino() ([]byte, error) {
	return m.Key, nil
}

func (m *PubKey) UnmarshalAmino(bytes []byte) error {
	if len(bytes) != mldsa44.PublicKeySize {
		return errorsmod.Wrap(errors.ErrInvalidPubKey, "invalid public key size")
	}
	m.Key = bytes
	return nil
}

func (m *PubKey) MarshalAminoJSON() ([]byte, error) {
	return m.MarshalAmino()
}

func (m *PubKey) UnmarshalAminoJSON(bytes []byte) error {
	return m.UnmarshalAmino(bytes)
}
