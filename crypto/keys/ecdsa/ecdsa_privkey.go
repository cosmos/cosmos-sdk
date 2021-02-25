package ecdsa

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	gogotypes "github.com/gogo/protobuf/types"
)

type ecdsaSK struct {
	ecdsa.PrivateKey
}

// GenSecp256r1 generates a new secp256r1 private key. It uses operating system randomness.
func GenSecp256r1() (cryptotypes.PrivKey, error) {
	key, err := ecdsa.GenerateKey(secp256r1, rand.Reader)
	return &ecdsaSK{*key}, err
}

// PubKey implements Cosmos-SDK PrivKey interface.
func (sk *ecdsaSK) PubKey() cryptotypes.PubKey {
	return &ecdsaPK{sk.PublicKey, nil}
}

// Bytes serialize the private key with first byte being the curve type.
func (sk *ecdsaSK) Bytes() []byte {
	if sk == nil {
		return nil
	}
	bz := make([]byte, PrivKeySize)
	bz[0] = curveTypes[sk.Curve]
	sk.D.FillBytes(bz[1:])
	return bz
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (sk *ecdsaSK) Equals(other cryptotypes.LedgerPrivKey) bool {
	sk2, ok := other.(*ecdsaSK)
	if !ok {
		return false
	}
	return sk.PrivateKey.Equal(&sk2.PrivateKey)
}

// Type returns key type name. Implements sdk.PrivKey interface.
func (sk *ecdsaSK) Type() string {
	return curveNames[sk.Curve]
}

// Sign hashes and signs the message usign ECDSA. Implements sdk.PrivKey interface.
func (sk *ecdsaSK) Sign(msg []byte) ([]byte, error) {
	digest := sha256.Sum256(msg)
	return sk.PrivateKey.Sign(rand.Reader, digest[:], nil)
}

// **** proto.Message ****

// Reset implements proto.Message interface.
func (sk *ecdsaSK) Reset() {
	sk.D = new(big.Int)
	sk.PublicKey = ecdsa.PublicKey{}
}

// ProtoMessage implements proto.Message interface.
func (*ecdsaSK) ProtoMessage() {}

// String implements proto.Message interface.
func (sk *ecdsaSK) String() string {
	return curveNames[sk.Curve] + "{-}"
}

// **** Proto Marshaler ****

// Marshal implements ProtoMarshaler interface.
func (sk *ecdsaSK) Marshal() ([]byte, error) {
	bv := gogotypes.BytesValue{Value: sk.Bytes()}
	return bv.Marshal()
}

// MarshalTo implements ProtoMarshaler interface.
func (sk *ecdsaSK) MarshalTo(data []byte) (int, error) {
	bv := gogotypes.BytesValue{Value: sk.Bytes()}
	return bv.MarshalTo(data)
}

// MarshalToSizedBuffer implements ProtoMarshaler interface.
func (sk *ecdsaSK) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	bv := gogotypes.BytesValue{Value: sk.Bytes()}
	return bv.MarshalToSizedBuffer(dAtA)
}

// Unmarshal implements ProtoMarshaler interface.
func (sk *ecdsaSK) Unmarshal(b []byte) error {
	bv := gogotypes.BytesValue{}
	err := bv.Unmarshal(b)
	if err != nil {
		return err
	}
	bz := bv.Value
	if len(bz) != PrivKeySize {
		return fmt.Errorf("wrong ECDSA SK bytes, expecting %d bytes", PrivKeySize)
	}
	curve, ok := curveTypesRev[bz[0]]
	if !ok {
		return fmt.Errorf("wrong ECDSA PK bytes, unknown curve type: %d", bz[0])
	}

	if sk == nil {
		sk = new(ecdsaSK)
	}
	sk.Curve = curve
	sk.D = new(big.Int).SetBytes(bz[1:])
	sk.X, sk.Y = curve.ScalarBaseMult(bz[1:])
	return nil
}

// Size implements ProtoMarshaler interface.
func (sk *ecdsaSK) Size() int {
	if sk == nil {
		return 0
	}
	return sovPrivKeySize
}
