package ecdsa

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	//	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// "github.com/cosmos/cosmos-sdk/codec"
//
// "github.com/cosmos/cosmos-sdk/types/errors"

const (
	// PubKeySize is is the size, in bytes, of public keys as used in this package.
	PubKeySize = 32 + 1
	// PrivKeySize is the size, in bytes, of private keys as used in this package.
	// PrivKeySize = 64
	// SeedSize is the size, in bytes, of private key seeds. These are the
	// private key representations used by RFC 8032.
	SeedSize = 32
)

var secp256r1 elliptic.Curve
var curveNames map[elliptic.Curve]string

func init() {
	secp256r1 = elliptic.P256()
	params := secp256r1.Params()
	if params.BitSize/8 != PubKeySize-1 {
		panic(fmt.Sprintf("Wrong PubKeySize=%d, expecting=%d", PubKeySize, params.BitSize/8))
	}

	curveNames = map[elliptic.Curve]string{
		secp256r1: "secp256r1",
	}
}

type ecdsaPK struct {
	*ecdsa.PublicKey

	// cache
	address []byte // skd.AccAddress
}

// TODO:
// var _ cryptotypes.PubKey = &PubKey{}

// String implements PubKey interface
func (pk ecdsaPK) Address() sdk.AccAddress {
	if pk.address == nil {
		pk.address = address.Hash(curveNames[pk.Curve], pk.Bytes())
	}
	return pk.address
}

// String implements PubKey interface
func (pk ecdsaPK) String() string {
	return fmt.Sprintf("%s{%X}", curveNames[pk.Curve], pk.Bytes())
}

// Bytes returns the byte representation of the public key in a compressed representation.
func (pk ecdsaPK) Bytes() []byte {
	return elliptic.MarshalCompressed(pk.Curve, pk.X, pk.Y)
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (pk ecdsaPK) Equal(other crypto.PublicKey) bool {
	pk2, ok := other.(ecdsaPK)
	if !ok {
		return false
	}

	return pk.PublicKey.Equal(pk2.PublicKey)
}

/*
   type PubKey interface {
	proto.Message

	Address() Address
	Bytes() []byte
	VerifySignature(msg []byte, sig []byte) bool
	Equals(PubKey) bool
	Type() string
    }

   // Message is implemented by generated protocol buffer messages.
   type Message interface {
	Reset()
	String() string
	ProtoMessage()
    }
*/
