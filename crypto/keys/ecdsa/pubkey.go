package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"

	gogotypes "github.com/gogo/protobuf/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

// signature holds the r and s values of an ECDSA signature.
type signature struct {
	R, S *big.Int
}

type ecdsaPK struct {
	ecdsa.PublicKey

	// cache
	address tmcrypto.Address
}

// String implements PubKey interface.
func (pk *ecdsaPK) Address() tmcrypto.Address {
	if pk.address == nil {
		pk.address = address.Hash(curveNames[pk.Curve], pk.Bytes())
	}
	return pk.address
}

// Bytes returns the byte representation of the public key using a compressed form
// specified in section 4.3.6 of ANSI X9.62 with first byte being the curve type.
func (pk *ecdsaPK) Bytes() []byte {
	if pk == nil {
		return nil
	}
	compressed := make([]byte, PubKeySize)
	compressed[0] = curveTypes[pk.Curve]
	compressed[1] = byte(pk.Y.Bit(0)) | 2
	pk.X.FillBytes(compressed[2:])
	return compressed
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (pk *ecdsaPK) Equals(other cryptotypes.PubKey) bool {
	pk2, ok := other.(*ecdsaPK)
	if !ok {
		return false
	}

	return pk.PublicKey.Equal(&pk2.PublicKey)
}

// VerifySignature implements skd.PubKey interface.
func (pk *ecdsaPK) VerifySignature(msg []byte, sig []byte) bool {
	s := new(signature)
	if _, err := asn1.Unmarshal(sig, s); err != nil || s == nil {
		return false
	}

	h := sha256.Sum256(msg)
	return ecdsa.Verify(&pk.PublicKey, h[:], s.R, s.S)
}

// Type returns key type name. Implements sdk.PubKey interface.
func (pk *ecdsaPK) Type() string {
	return curveNames[pk.Curve]
}

// **** proto.Message ****

// Reset implements proto.Message interface.
func (pk *ecdsaPK) Reset() {
	*pk = ecdsaPK{}
}

// ProtoMessage implements proto.Message interface.
func (*ecdsaPK) ProtoMessage() {}

// String implements proto.Message interface
func (pk *ecdsaPK) String() string {
	return fmt.Sprintf("%s{%X}", curveNames[pk.Curve], pk.Bytes())
}

// **** Proto Marshaler ****

// Marshal implements ProtoMarshaler interface.
func (pk *ecdsaPK) Marshal() ([]byte, error) {
	bv := gogotypes.BytesValue{Value: pk.Bytes()}
	return bv.Marshal()
}

// MarshalTo implements ProtoMarshaler interface.
func (pk *ecdsaPK) MarshalTo(data []byte) (int, error) {
	bv := gogotypes.BytesValue{Value: pk.Bytes()}
	return bv.MarshalTo(data)
}

// MarshalToSizedBuffer implements ProtoMarshaler interface.
func (pk *ecdsaPK) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	bv := gogotypes.BytesValue{Value: pk.Bytes()}
	return bv.MarshalToSizedBuffer(dAtA)
}

// Unmarshal implements ProtoMarshaler interface.
func (pk *ecdsaPK) Unmarshal(b []byte) error {
	bv := gogotypes.BytesValue{}
	err := bv.Unmarshal(b)
	if err != nil {
		return err
	}
	return pk.UnmarshalAmino(bv.Value)
}

// Size implements ProtoMarshaler interface.
func (pk *ecdsaPK) Size() int {
	if pk == nil {
		return 0
	}
	return sovPubKeySize
}

// **** Amino Marshaler ****

// MarshalAmino overrides Amino binary marshalling.
func (pk *ecdsaPK) MarshalAmino() ([]byte, error) {
	return pk.Bytes(), nil
}

// UnmarshalAmino overrides Amino binary marshalling.
func (pk *ecdsaPK) UnmarshalAmino(bz []byte) error {
	if len(bz) != PubKeySize {
		return errors.Wrapf(errors.ErrInvalidPubKey, "wrong ECDSA PK bytes, expecting %d bytes", PubKeySize)
	}
	curve, ok := curveTypesRev[bz[0]]
	if !ok {
		return errors.Wrapf(errors.ErrInvalidPubKey, "wrong ECDSA PK bytes, unknown curve type: %d", bz[0])
	}
	cpk := ecdsa.PublicKey{Curve: curve}
	cpk.X, cpk.Y = elliptic.UnmarshalCompressed(curve, bz[1:])
	if cpk.X == nil || cpk.Y == nil {
		return errors.Wrapf(errors.ErrInvalidPubKey, "wrong ECDSA PK bytes, unknown curve type: %d", bz[0])
	}
	if pk == nil {
		*pk = ecdsaPK{cpk, nil}
	} else {
		pk.PublicKey = cpk
	}
	return nil
}

// MarshalAminoJSON overrides Amino JSON marshalling.
func (pk *ecdsaPK) MarshalAminoJSON() ([]byte, error) {
	return pk.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshalling.
func (pk *ecdsaPK) UnmarshalAminoJSON(bz []byte) error {
	return pk.UnmarshalAmino(bz)
}
