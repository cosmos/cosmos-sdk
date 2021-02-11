package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"

	tmcrypto "github.com/tendermint/tendermint/crypto"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

// "github.com/cosmos/cosmos-sdk/codec"
//
// "github.com/cosmos/cosmos-sdk/types/errors"

const (
	// PubKeySize is is the size, in bytes, of public keys as used in this package.
	PubKeySize = 32 + 1 + 1
	// PrivKeySize is the size, in bytes, of private keys as used in this package.
	PrivKeySize = 32 + 1
)

var secp256r1 elliptic.Curve
var curveNames map[elliptic.Curve]string
var curveTypes map[elliptic.Curve]byte
var curveTypesRev map[byte]elliptic.Curve

func init() {
	secp256r1 = elliptic.P256()
	// PubKeySize is ceil of field bit size + 1 for the sign + 1 for the type
	expected := (secp256r1.Params().BitSize+7)/8 + 2
	if expected != PubKeySize {
		panic(fmt.Sprintf("Wrong PubKeySize=%d, expecting=%d", PubKeySize, expected))
	}

	curveNames = map[elliptic.Curve]string{
		secp256r1: "secp256r1",
	}
	curveTypes = map[elliptic.Curve]byte{
		// 0 reserved
		secp256r1: 1,
	}
	curveTypesRev = map[byte]elliptic.Curve{}
	for c, b := range curveTypes {
		curveTypesRev[b] = c
	}
}

// signature holds the r and s values of an ECDSA signature
type signature struct {
	R, S *big.Int
}

type ecdsaPK struct {
	ecdsa.PublicKey

	// cache
	address tmcrypto.Address
}

var _ cryptotypes.PubKey = &ecdsaPK{}

// String implements PubKey interface
func (pk ecdsaPK) Address() tmcrypto.Address {
	if pk.address == nil {
		pk.address = address.Hash(curveNames[pk.Curve], pk.Bytes())
	}
	return pk.address
}

// String implements PubKey interface
func (pk ecdsaPK) String() string {
	return fmt.Sprintf("%s{%X}", curveNames[pk.Curve], pk.Bytes())
}

// Bytes returns the byte representation of the public key using a compressed form
// specified in section 4.3.6 of ANSI X9.62 with first byte being the curve type.
func (pk ecdsaPK) Bytes() []byte {
	compressed := make([]byte, PubKeySize)
	compressed[0] = curveTypes[pk.Curve]
	compressed[1] = byte(pk.Y.Bit(0)) | 2
	pk.X.FillBytes(compressed[2:])
	return compressed
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (pk ecdsaPK) Equals(other cryptotypes.PubKey) bool {
	pk2, ok := other.(ecdsaPK)
	if !ok {
		return false
	}

	return pk.PublicKey.Equal(&pk2.PublicKey)
}

// VerifySignature implements skd.PubKey interface
func (pk ecdsaPK) VerifySignature(msg []byte, sig []byte) bool {
	s := new(signature)
	if _, err := asn1.Unmarshal(sig, s); err != nil || s == nil {
		return false
	}

	h := sha256.Sum256(msg)
	return ecdsa.Verify(&pk.PublicKey, h[:], s.R, s.S)
}

// Type returns key type name. Implements sdk.PubKey interface
func (pk ecdsaPK) Type() string {
	return curveNames[pk.Curve]
}

// **** proto.Message ****

func (pk ecdsaPK) Reset()     {} // TODO: maybe we need to have this?
func (ecdsaPK) ProtoMessage() {}

/*
ProtoMarshaler interface {
	Marshal() ([]byte, error)
	MarshalTo(data []byte) (n int, err error)
	MarshalToSizedBuffer(dAtA []byte) (int, error)
	Size() int
	Unmarshal(data []byte) error
}
*/

// **** Amino Marshaler ****

// MarshalAmino overrides Amino binary marshalling.
func (pk ecdsaPK) MarshalAmino() ([]byte, error) {
	return pk.Bytes(), nil
}

// UnmarshalAmino overrides Amino binary marshalling.
func (pk *ecdsaPK) UnmarshalAmino(bz []byte) error {
	if len(bz) != PubKeySize {
		return errors.Wrap(errors.ErrInvalidPubKey, "invalid pubkey size")
	}
	curve, ok := curveTypesRev[bz[0]]
	if !ok {
		return errors.Wrap(errors.ErrInvalidPubKey, "invalid curve type")
	}
	x, y := elliptic.UnmarshalCompressed(curve, bz[1:])
	if x == nil || y == nil {
		return errors.Wrap(errors.ErrInvalidPubKey, "invalid pubkey bytes")
	}
	pk.PublicKey.Curve = curve
	pk.PublicKey.X, pk.PublicKey.Y = x, y
	return nil
}

// MarshalAminoJSON overrides Amino JSON marshalling.
func (pk ecdsaPK) MarshalAminoJSON() ([]byte, error) {
	return pk.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshalling.
func (pk *ecdsaPK) UnmarshalAminoJSON(bz []byte) error {
	return pk.UnmarshalAmino(bz)
}
