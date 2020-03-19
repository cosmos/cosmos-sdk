package sr25519

import (
	"bytes"
	"fmt"

	amino "github.com/tendermint/go-amino"

	crypto "github.com/cosmos/cosmos-sdk/crypto/keys"
)

var _ crypto.PrivKey = PrivKey{}

const (
	PrivKeyAminoName = "tendermint/PrivKeySr25519"
	PubKeyAminoName  = "tendermint/PubKeySr25519"

	// SignatureSize is the size of an Edwards25519 signature. Namely the size of a compressed
	// Sr25519 point, and a field element. Both of which are 32 bytes.
	SignatureSize = 64
)

var (
	prefixPrivKeySr25519      = []byte{0x2F, 0x82, 0xD7, 0x8B}
	lengthPrivKeySr25519 byte = 0x20

	prefixPubKeySr25519      = []byte{0x0D, 0xFB, 0x10, 0x05}
	lengthPubKeySr25519 byte = 0x20
)

func RegisterCodec(c *amino.Codec) {
	c.RegisterInterface((*crypto.PubKey)(nil), nil)
	c.RegisterConcrete(PubKey{}, PubKeyAminoName, nil)

	c.RegisterInterface((*crypto.PrivKey)(nil), nil)
	c.RegisterConcrete(PrivKey{}, PrivKeyAminoName, nil)
}

// Marshal attempts to marshal a PrivKeySr25519 type that is backwards
// compatible with Amino.
func (privKey PrivKey) AminoMarshal() ([]byte, error) {
	lbz := []byte{lengthPrivKeySr25519}
	p := len(prefixPrivKeySr25519)
	l := len(lbz)
	bz := make([]byte, p+l+len(privKey))

	copy(bz[:p], prefixPrivKeySr25519)
	copy(bz[p:p+l], lbz)
	copy(bz[p+l:], privKey[:])

	return bz, nil
}

// Unmarshal attempts to unmarshal provided amino compatbile bytes into a
// PrivKeySr25519 reference. An error is returned if the encoding is invalid.
func (privKey *PrivKey) AminoUnmarshal(bz []byte) error {
	lbz := []byte{lengthPrivKeySr25519}
	p := len(prefixPrivKeySr25519)
	l := len(lbz)

	if !bytes.Equal(bz[:p], prefixPrivKeySr25519) {
		return fmt.Errorf("invalid prefix; expected: %X, got: %X", prefixPrivKeySr25519, bz[:p])
	}
	if !bytes.Equal(bz[p:p+l], lbz) {
		return fmt.Errorf("invalid encoding length; expected: %X, got: %X", lbz, bz[p:p+l])
	}
	if len(bz[p+l:]) != int(lengthPrivKeySr25519) {
		return fmt.Errorf("invalid key length; expected: %d, got: %d", int(lengthPrivKeySr25519), len(bz[p+l:]))
	}

	*privKey = bz[p+l:]
	return nil
}

// MarshalBinary attempts to marshal a PubKeySr25519 type that is backwards
// compatible with Amino.
//
// NOTE: Amino will not delegate MarshalBinaryBare calls to types that implement
// it. For now, clients must call MarshalBinary directly on the type to get the
// custom compatible encoding.
func (pubKey PubKey) AminoMarshal() ([]byte, error) {
	lbz := []byte{lengthPubKeySr25519}
	p := len(prefixPubKeySr25519)
	l := len(lbz)
	bz := make([]byte, p+l+len(pubKey[:]))

	copy(bz[:p], prefixPubKeySr25519)
	copy(bz[p:p+l], lbz)
	copy(bz[p+l:], pubKey[:])

	return bz, nil
}

// UnmarshalBinary attempts to unmarshal provided amino compatbile bytes into a
// PubKeySr25519 reference. An error is returned if the encoding is invalid.
//
// NOTE: Amino will not delegate UnmarshalBinaryBare calls to types that implement
// it. For now, clients must call UnmarshalBinary directly on the type to get the
// custom compatible decoding.
func (pubKey *PubKey) AminoUnmarshal(bz []byte) error {
	lbz := []byte{lengthPubKeySr25519}
	p := len(prefixPubKeySr25519)
	l := len(lbz)

	if !bytes.Equal(bz[:p], prefixPubKeySr25519) {
		return fmt.Errorf("invalid prefix; expected: %X, got: %X", prefixPubKeySr25519, bz[:p])
	}
	if !bytes.Equal(bz[p:p+l], lbz) {
		return fmt.Errorf("invalid encoding length; expected: %X, got: %X", lbz, bz[p:p+l])
	}
	if len(bz[p+l:]) != int(lengthPubKeySr25519) {
		return fmt.Errorf("invalid key length; expected: %d, got: %d", int(lengthPubKeySr25519), len(bz[p+l:]))
	}

	*pubKey = bz[p+l:]
	return nil
}
