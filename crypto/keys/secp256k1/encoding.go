package secp256k1

import (
	"bytes"
	"fmt"

	amino "github.com/tendermint/go-amino"

	crypto "github.com/cosmos/cosmos-sdk/crypto/keys"
)

const (
	PrivKeyAminoName = "tendermint/PrivKeySecp256k1"
	PubKeyAminoName  = "tendermint/PubKeySecp256k1"
)

var cdc = amino.NewCodec()

func init() {
	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	cdc.RegisterConcrete(PubKey{},
		PubKeyAminoName, nil)

	cdc.RegisterInterface((*crypto.PrivKey)(nil), nil)
	cdc.RegisterConcrete(PrivKey{},
		PrivKeyAminoName, nil)
}

var (
	prefixPrivKeySecp256k1      = []byte{0xE1, 0xB0, 0xF7, 0x9B}
	lengthPrivKeySecp256k1 byte = 0x20

	prefixPubKeySecp256k1      = []byte{0xEB, 0x5A, 0xE9, 0x87}
	lengthPubKeySecp256k1 byte = 0x21
)

// MarshalBinary attempts to marshal a PrivKeySecp256k1 type that is backwards
// compatible with Amino.
//
// NOTE: Amino will not delegate MarshalBinaryBare calls to types that implement
// it. For now, clients must call MarshalBinary directly on the type to get the
// custom compatible encoding.
func (privKey PrivKey) AminoMarshal() ([]byte, error) {
	lbz := []byte{lengthPrivKeySecp256k1}
	p := len(prefixPrivKeySecp256k1)
	l := len(lbz)
	bz := make([]byte, p+l+len(privKey[:]))

	copy(bz[:p], prefixPrivKeySecp256k1)
	copy(bz[p:p+l], lbz)
	copy(bz[p+l:], privKey[:])

	return bz, nil
}

// UnmarshalBinary attempts to unmarshal provided amino compatbile bytes into a
// PrivKeySecp256k1 reference. An error is returned if the encoding is invalid.
//
// NOTE: Amino will not delegate UnmarshalBinaryBare calls to types that implement
// it. For now, clients must call UnmarshalBinary directly on the type to get the
// custom compatible decoding.
func (privKey *PrivKey) AminoUnmarshal(bz []byte) error {
	lbz := []byte{lengthPrivKeySecp256k1}
	p := len(prefixPrivKeySecp256k1)
	l := len(lbz)

	if !bytes.Equal(bz[:p], prefixPrivKeySecp256k1) {
		return fmt.Errorf("invalid prefix; expected: %X, got: %X", prefixPrivKeySecp256k1, bz[:p])
	}
	if !bytes.Equal(bz[p:p+l], lbz) {
		return fmt.Errorf("invalid encoding length; expected: %X, got: %X", lbz, bz[p:p+l])
	}
	if len(bz[p+l:]) != int(lengthPrivKeySecp256k1) {
		return fmt.Errorf("invalid key length; expected: %d, got: %d", int(lengthPrivKeySecp256k1), len(bz[p+l:]))
	}

	*privKey = bz[p+l:]
	return nil
}

// MarshalBinary attempts to marshal a PubKeySecp256k1 type that is backwards
// compatible with Amino.
//
// NOTE: Amino will not delegate MarshalBinaryBare calls to types that implement
// it. For now, clients must call MarshalBinary directly on the type to get the
// custom compatible encoding.
func (pubKey PubKey) AminoMarshal() ([]byte, error) {
	lbz := []byte{lengthPubKeySecp256k1}
	p := len(prefixPubKeySecp256k1)
	l := len(lbz)
	bz := make([]byte, p+l+len(pubKey[:]))

	copy(bz[:p], prefixPubKeySecp256k1)
	copy(bz[p:p+l], lbz)
	copy(bz[p+l:], pubKey[:])

	return bz, nil
}

// UnmarshalBinary attempts to unmarshal provided amino compatbile bytes into a
// PubKeySecp256k1 reference. An error is returned if the encoding is invalid.
//
// NOTE: Amino will not delegate UnmarshalBinaryBare calls to types that implement
// it. For now, clients must call UnmarshalBinary directly on the type to get the
// custom compatible decoding.
func (pubKey *PubKey) AminoUnmarshal(bz []byte) error {
	lbz := []byte{lengthPubKeySecp256k1}
	p := len(prefixPubKeySecp256k1)
	l := len(lbz)

	if !bytes.Equal(bz[:p], prefixPubKeySecp256k1) {
		return fmt.Errorf("invalid prefix; expected: %X, got: %X", prefixPubKeySecp256k1, bz[:p])
	}
	if !bytes.Equal(bz[p:p+l], lbz) {
		return fmt.Errorf("invalid encoding length; expected: %X, got: %X", lbz, bz[p:p+l])
	}
	if len(bz[p+l:]) != int(lengthPubKeySecp256k1) {
		return fmt.Errorf("invalid key length; expected: %d, got: %d", int(lengthPubKeySecp256k1), len(bz[p+l:]))
	}

	*pubKey = bz[p+l:]
	return nil
}
