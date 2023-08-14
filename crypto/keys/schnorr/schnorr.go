package schnorr

import (
	"crypto/subtle"
	"fmt"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/key"

	errorsmod "cosmossdk.io/errors"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

// TODO -> Check error handling
const (
	PubKeyName  = "tendermint/PubKeySchnorr"
	PrivKeyName = "tendermint/PrivKeySchnorr"
	// PubKeySize is is the size, in bytes, of public keys as used in this package.
	PubKeySize = 32
	// PrivKeySize is the size, in bytes, of private keys as used in this package.
	PrivKeySize = 32
	// SignatureSize Size of an Schnorr signature
	SignatureSize = 64

	keyType = "schnorr"
	curve   = "Ed25519"
)

// GenPrivKey generates a new Schnorr private key.
func GenPrivKey() *PrivKey {
	suite := suites.MustFind(curve)
	keyPair := suite.Scalar().Pick(suite.RandomStream())
	binary, err := keyPair.MarshalBinary()
	if err != nil {
		fmt.Printf("[ERRPR] While generating priv key: %e", err)
	}
	return &PrivKey{Key: binary}
}

func (privKey *PrivKey) GetKeyPair() *key.Pair {
	suite := suites.MustFind(curve)
	keyPair := key.NewKeyPair(suite)
	err := keyPair.Private.UnmarshalBinary(privKey.Bytes())
	if err != nil {
		fmt.Println("failed to unmarshall binary on private key bytes")
		return nil
	}

	return keyPair
}

// Bytes returns the private key byte format.
func (privKey *PrivKey) Bytes() []byte {
	return privKey.Key
}

// Sign produces a signature on the provided message.
func (privKey *PrivKey) Sign(msg []byte) ([]byte, error) {
	suite := suites.MustFind(curve)

	keyPair := privKey.GetKeyPair()

	signedMsg, err := schnorr.Sign(suite, keyPair.Private, msg)
	if err != nil {
		return nil, fmt.Errorf("signing message failed %s", err.Error())
	}
	return signedMsg, err
}

// PubKey gets the corresponding public key from the private key.
func (privKey *PrivKey) PubKey() cryptotypes.PubKey {
	suite := suites.MustFind(curve)
	keyPair := privKey.GetKeyPair()

	// Gets public key by y = g ^ x definition where x is the private key scalar
	publicKey := suite.Point().Mul(keyPair.Private, nil)
	binary, err := publicKey.MarshalBinary()
	if err != nil {
		fmt.Printf("while generating pub key: %s", err.Error())
	}

	return &PubKey{Key: binary}
}

// Type returns key type
func (privKey *PrivKey) Type() string {
	return keyType
}

// Equals compares 2 keys
func (privKey *PrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	if privKey.Type() != other.Type() {
		return false
	}

	privKeyBytes := privKey.Bytes()

	return subtle.ConstantTimeCompare(privKeyBytes, other.Bytes()) == 1
}

// MarshalAmino overrides Amino binary marshaling.
func (privKey PrivKey) MarshalAmino() ([]byte, error) {
	return privKey.Key, nil
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
		return fmt.Errorf("[ERROR] invalid privkey size")
	}

	privKey.Key = bz

	return nil
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
	if len(pubKey.Bytes()) != PubKeySize {
		fmt.Println("pubkey is incorrect size")
		return nil
	}
	return crypto.Address(tmhash.SumTruncated(pubKey.Bytes()))
}

// Bytes returns the PubKey byte format.
func (pubKey *PubKey) Bytes() []byte {
	return pubKey.Key
}

func (pubKey *PubKey) GetKeyPair() *key.Pair {
	suite := suites.MustFind(curve)
	keyPair := key.NewKeyPair(suite)
	_ = keyPair.Public.UnmarshalBinary(pubKey.Key)

	return keyPair
}

func (pubKey *PubKey) VerifySignature(msg, sig []byte) bool {
	if len(sig) != SignatureSize {
		return false
	}

	suite := suites.MustFind(curve)
	keyPair := pubKey.GetKeyPair()

	err := schnorr.Verify(suite, keyPair.Public, msg, sig)
	if err != nil {
		fmt.Println("Schnorr verification failed", err.Error())
	}
	return err == nil
}

func (pubKey *PubKey) Type() string {
	return keyType
}

// TODO Change cryptotypes.PubKey
func (pubKey *PubKey) Equals(other cryptotypes.PubKey) bool {
	if pubKey.Type() != other.Type() {
		return false
	}

	return subtle.ConstantTimeCompare(pubKey.Bytes(), other.Bytes()) == 1
}

// MarshalAmino overrides Amino binary marshaling.
func (pubKey PubKey) MarshalAmino() ([]byte, error) {
	return pubKey.Key, nil
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

	pubKey.Key = bz

	return nil
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey *PubKey) UnmarshalAminoJSON(bz []byte) error {
	return pubKey.UnmarshalAmino(bz)
}
