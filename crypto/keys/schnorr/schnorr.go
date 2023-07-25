package schnorr

import (
	errorsmod "cosmossdk.io/errors"
	"crypto/subtle"
	"fmt"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
	"go.dedis.ch/kyber/v4/sign/schnorr"
	"go.dedis.ch/kyber/v4/util/key"
)

// TODO recheck these values
const (
	// PubKeySize is is the size, in bytes, of public keys as used in this package.
	PubKeySize = 32
	// PrivKeySize is the size, in bytes, of private keys as used in this package.
	PrivKeySize = 64
	// SignatureSize Size of an Schnorr signature
	SignatureSize = 64

	keyType = "schnorr"
)

// GenPrivKey generates a new Schnorr private key. These Schnorr keys must not
func GenPrivKey() *PrivKey {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	keyPair := key.NewKeyPair(suite)
	return &PrivKey{Key: keyPair.Private, Suite: suite}
}

// Bytes returns the private key byte format.
func (privKey *PrivKey) Bytes() ([]byte, error) {
	return privKey.Key.MarshalBinary()
}

// Sign produces a signature on the provided message.
func (privKey *PrivKey) Sign(msg []byte) ([]byte, error) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	signedMsg, err := schnorr.Sign(suite, privKey.Key, msg)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Signing message failed: %e", err)
	}
	return signedMsg, err
}

// PubKey gets the corresponding public key from the private key.
func (privKey *PrivKey) PubKey() *PubKey {
	var g kyber.Group = privKey.Suite
	// Gets public key by y = g ^ x definition where x is the private key scalar
	return &PubKey{Key: g.Point().Mul(privKey.Key, nil), Suite: privKey.Suite}
}

// Type returns key type
func (privKey *PrivKey) Type() string {
	return keyType
}

// Equals
// Runs in constant time based on length of the keys.
func (privKey *PrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	if privKey.Type() != other.Type() {
		return false
	}

	privKeyBytes, err := privKey.Bytes()
	if err != nil {
		fmt.Println(err)
		return false
	}

	return subtle.ConstantTimeCompare(privKeyBytes, other.Bytes()) == 1
}

// MarshalAmino overrides Amino binary marshaling.
func (privKey PrivKey) MarshalAmino() ([]byte, error) {
	return privKey.Bytes()
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (privKey PrivKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return privKey.MarshalAmino()
}

// UnmarshalAmino overrides Amino binary marshaling.
func (privKey *PrivKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != PrivKeySize {
		return fmt.Errorf("[ERROR] invalid privkey size.")
	}

	return privKey.Key.UnmarshalBinary(bz)
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (privKey *PrivKey) UnmarshalAminoJSON(bz []byte) error {
	return privKey.UnmarshalAmino(bz)
}

//-------------------------------------

// Address is the SHA256-20 of the raw pubkey bytes.
// TODO ADR-28 addresses
// It doesn't implement ADR-28 addresses and it must not be used
// in SDK except in a tendermint validator context.
func (pubKey *PubKey) Address() crypto.Address {
	pubKeyBytes, err := pubKey.Bytes()
	if err != nil {
		panic(err)
	}

	if len(pubKeyBytes) != PubKeySize {
		panic("pubkey is incorrect size")
	}
	return crypto.Address(tmhash.SumTruncated(pubKeyBytes))
}

// Bytes returns the PubKey byte format.
func (pubKey *PubKey) Bytes() ([]byte, error) {
	return pubKey.Key.MarshalBinary()
}

func (pubKey *PubKey) VerifySignature(msg, sig []byte) bool {
	if len(sig) != SignatureSize {
		return false
	}

	err := schnorr.Verify(pubKey.Suite, pubKey.Key, msg, sig)
	if err != nil {
		fmt.Printf("[ERROR] while verifying signature: %e", err)
	}
	return err == nil
}

// String returns Hex representation of a pub key with their type
func (pubKey *PubKey) String() string {
	return fmt.Sprintf("PubKeyShnorr{%X}", pubKey.Key)
}

func (pubKey *PubKey) Type() string {
	return keyType
}

// TODO Change cryptotypes.PubKey
func (pubKey *PubKey) Equals(other PubKey) bool {
	if pubKey.Type() != other.Type() {
		return false
	}

	pubKeyBytes, err := pubKey.Bytes()
	if err != nil {
		fmt.Printf("[ERROR] There is an error: %e.", err)
		return false
	}
	otherBytes, _ := other.Bytes()
	return subtle.ConstantTimeCompare(pubKeyBytes, otherBytes) == 1
}

// MarshalAmino overrides Amino binary marshaling.
func (pubKey PubKey) MarshalAmino() ([]byte, error) {
	return pubKey.Bytes()
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey PubKey) MarshalAminoJSON() ([]byte, error) {
	return pubKey.MarshalAmino()
}

// UnmarshalAmino overrides Amino binary marshaling.
func (pubKey *PubKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != PubKeySize {
		return errorsmod.Wrap(errors.ErrInvalidPubKey, "invalid pubkey size")
	}

	return pubKey.Key.UnmarshalBinary(bz)
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey *PubKey) UnmarshalAminoJSON(bz []byte) error {
	return pubKey.UnmarshalAmino(bz)
}
