package secp256r1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// "github.com/cosmos/cosmos-sdk/codec"
//
// "github.com/cosmos/cosmos-sdk/types/errors"

const (
	keyType     = "secp256r1"
	PrivKeyName = "cosmos/PrivKeySecp256r1"
	PubKeyName  = "cosmos/PubKeySecp256r1"
	// PubKeySize is is the size, in bytes, of public keys as used in this package.
	// PubKeySize = 32
	// PrivKeySize is the size, in bytes, of private keys as used in this package.
	// PrivKeySize = 64
	// SeedSize is the size, in bytes, of private key seeds. These are the
	// private key representations used by RFC 8032.
	SeedSize = 32
)

var curve elliptic.Curve

func init() {
	curve = elliptic.P256()
	// params := curve.Params()
	// if params.BitSize/8 != PubKeySize {
	// 	panic(fmt.Sprintf("Wrong PubKeySize=%d, expecting=%d", PubKeySize, params.BitSize/8))
	// }
}

// type curve

type ecdsaPK struct {
	ecdsa.PublicKey
	//	typ
}

// var _ cryptotypes.PubKey = &PubKey{}

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

// Bytes returns the byte representation of the Private Key.
// func (privKey *PrivKey) Bytes() []byte {
// 	return privKey.Key
// }

func (pk *ecdsaPK) String() string {
	return fmt.Sprintf("secp256r1{%X}", pubKey.Key)
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (pk *ecdsaPK) Equals(other cryptotypes.PubKey) bool {

}
