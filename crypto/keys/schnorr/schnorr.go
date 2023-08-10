package schnorr

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"go.dedis.ch/kyber/v3"
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

type Suite interface {
	kyber.Group
	kyber.Encoding
	kyber.XOFFactory
}

type basicSig struct {
	C kyber.Scalar // challenge
	R kyber.Scalar // response
}

// GenPrivKey generates a new Schnorr private key.
func GenPrivKey() *PrivKey {
	suite := suites.MustFind(curve)
	keyPair := suite.Scgit alar().Pick(suite.RandomStream())
	binary, err := keyPair.MarshalBinary()
	if err != nil {
		fmt.Printf("[ERRPR] While generating priv key: %e", err)
	}
	return &PrivKey{Key: binary}
}

func (privKey *PrivKey) GetKeyPair() *key.Pair {
	suite := suites.MustFind(curve)
	keyPair := key.NewKeyPair(suite)
	_ = keyPair.Private.UnmarshalBinary(privKey.Key)

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

	// Create random secret v and public point commitment T
	v := suite.Scalar().Pick(suite.RandomStream())
	publicPoint := suite.Point().Mul(v, nil)

	// Create challenge c based on message and T
	c, err := hashSchnorr(suite, msg, publicPoint)
	if err != nil {
		return nil, fmt.Errorf("generating schnorr hash: %s", err.Error())
	}

	// Compute response r = v - x*c
	r := suite.Scalar()
	r.Mul(keyPair.Private, c).Sub(v, r)

	// Return verifiable signature {c, r}
	// Verifier will be able to compute v = r + x*c
	// And check that hashElgamal for T and the message == c
	buf := bytes.Buffer{}
	sig := basicSig{c, r}
	err = suite.Write(&buf, &sig)
	if err != nil {
		return nil, fmt.Errorf("signing message failed %s", err.Error())
	}

	return buf.Bytes(), err
}

// Returns a secret that depends on a message msg and a point P
func hashSchnorr(suite Suite, msg []byte, p kyber.Point) (kyber.Scalar, error) {
	pb, _ := p.MarshalBinary()
	c := suite.XOF(pb)
	_, err := c.Write(msg)
	return suite.Scalar().Pick(c), err
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
		panic("pubkey is incorrect size")
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

	buf := bytes.NewBuffer(sig)
	newBasicSig := basicSig{}
	if err := suite.Read(buf, &newBasicSig); err != nil {
		return false
	}
	r, c := newBasicSig.R, newBasicSig.C

	// base**(r + x*c) == T
	var P, T kyber.Point
	P, T = suite.Point(), suite.Point()
	T.Add(T.Mul(r, nil), P.Mul(c, pubKey.GetKeyPair().Public))

	// Verify that the hash based on the message and T
	// matches the challange c from the signature
	c, err := hashSchnorr(suite, msg, T)
	if err != nil {
		return false
	}

	if !c.Equal(newBasicSig.C) {
		return false
	}
	return true
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
