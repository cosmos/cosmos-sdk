package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"

	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

// signature holds the r and s values of an ECDSA signature.
type signature struct {
	R, S *big.Int
}

type PubKey struct {
	ecdsa.PublicKey

	// cache
	address tmcrypto.Address
}

// Address gets the address associated with a pubkey. If no address exists, it returns a newly created ADR-28 address
// for ECDSA keys.
// protoName is a concrete proto structure id.
func (pk *PubKey) Address(protoName string) tmcrypto.Address {
	if pk.address == nil {
		pk.address = address.Hash(protoName, pk.Bytes())
	}
	return pk.address
}

// Bytes returns the byte representation of the public key using a compressed form
// specified in section 4.3.6 of ANSI X9.62 with first byte being the curve type.
func (pk *PubKey) Bytes() []byte {
	if pk == nil {
		return nil
	}
	return elliptic.MarshalCompressed(pk.Curve, pk.X, pk.Y)
}

// VerifySignature checks if sig is a valid ECDSA signature for msg.
func (pk *PubKey) VerifySignature(msg []byte, sig []byte) bool {
	s := new(signature)
	if _, err := asn1.Unmarshal(sig, s); err != nil || s == nil {
		return false
	}

	h := sha256.Sum256(msg)
	return ecdsa.Verify(&pk.PublicKey, h[:], s.R, s.S)
}

// String returns a string representation of the public key based on the curveName.
func (pk *PubKey) String(curveName string) string {
	return fmt.Sprintf("%s{%X}", curveName, pk.Bytes())
}

// **** Proto Marshaler ****

// MarshalTo implements proto.Marshaler interface.
func (pk *PubKey) MarshalTo(dAtA []byte) (int, error) {
	bz := pk.Bytes()
	copy(dAtA, bz)
	return len(bz), nil
}

// Unmarshal implements proto.Marshaler interface.
func (pk *PubKey) Unmarshal(bz []byte, curve elliptic.Curve, expectedSize int) error {
	if len(bz) != expectedSize {
		return errors.Wrapf(errors.ErrInvalidPubKey, "wrong ECDSA PK bytes, expecting %d bytes, got %d", expectedSize, len(bz))
	}
	cpk := ecdsa.PublicKey{Curve: curve}
	cpk.X, cpk.Y = elliptic.UnmarshalCompressed(curve, bz)
	if cpk.X == nil || cpk.Y == nil {
		return errors.Wrapf(errors.ErrInvalidPubKey, "wrong ECDSA PK bytes, unknown curve type: %d", bz[0])
	}
	pk.PublicKey = cpk
	return nil
}
