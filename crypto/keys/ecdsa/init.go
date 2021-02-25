// ECDSA package implements Cosmos-SDK compatible ECDSA public and private key. The keys
// can be protobuf serialized and packed in Any.
// Currently supported keys are:
// + secp256r1
package ecdsa

import (
	"crypto/elliptic"
	"fmt"
	math_bits "math/bits"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

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

// Protobuf Bytes size - this computation is based on gogotypes.BytesValue.Sizee implementation
var sovPubKeySize = 1 + PubKeySize + sovKeys(PubKeySize)
var sovPrivKeySize = 1 + PrivKeySize + sovKeys(PrivKeySize)

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

// RegisterInterfaces adds ecdsa PubKey to pubkey registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*cryptotypes.PubKey)(nil), &ecdsaPK{})
}

func sovKeys(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
