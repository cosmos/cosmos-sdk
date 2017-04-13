package crypto

import (
	"bytes"
	"fmt"

	. "github.com/tendermint/go-common"
	data "github.com/tendermint/go-data"
	"github.com/tendermint/go-wire"
)

// Signature is a part of Txs and consensus Votes.
type Signature interface {
	Bytes() []byte
	IsZero() bool
	String() string
	Equals(Signature) bool
}

var sigMapper data.Mapper

// register both public key types with go-data (and thus go-wire)
func init() {
	sigMapper = data.NewMapper(SignatureS{}).
		RegisterImplementation(SignatureEd25519{}, NameEd25519, TypeEd25519).
		RegisterImplementation(SignatureSecp256k1{}, NameSecp256k1, TypeSecp256k1)
}

// SignatureS add json serialization to Signature
type SignatureS struct {
	Signature
}

func WrapSignature(sig Signature) SignatureS {
	for ssig, ok := sig.(SignatureS); ok; ssig, ok = sig.(SignatureS) {
		sig = ssig.Signature
	}
	return SignatureS{sig}
}

func (p SignatureS) MarshalJSON() ([]byte, error) {
	return sigMapper.ToJSON(p.Signature)
}

func (p *SignatureS) UnmarshalJSON(data []byte) (err error) {
	parsed, err := sigMapper.FromJSON(data)
	if err == nil && parsed != nil {
		p.Signature = parsed.(Signature)
	}
	return
}

func (p SignatureS) Empty() bool {
	return p.Signature == nil
}

func SignatureFromBytes(sigBytes []byte) (sig Signature, err error) {
	err = wire.ReadBinaryBytes(sigBytes, &sig)
	return
}

//-------------------------------------

// Implements Signature
type SignatureEd25519 [64]byte

func (sig SignatureEd25519) Bytes() []byte {
	return wire.BinaryBytes(struct{ Signature }{sig})
}

func (sig SignatureEd25519) IsZero() bool { return len(sig) == 0 }

func (sig SignatureEd25519) String() string { return fmt.Sprintf("/%X.../", Fingerprint(sig[:])) }

func (sig SignatureEd25519) Equals(other Signature) bool {
	if otherEd, ok := other.(SignatureEd25519); ok {
		return bytes.Equal(sig[:], otherEd[:])
	} else {
		return false
	}
}

func (p SignatureEd25519) MarshalJSON() ([]byte, error) {
	return data.Encoder.Marshal(p[:])
}

func (p *SignatureEd25519) UnmarshalJSON(enc []byte) error {
	var ref []byte
	err := data.Encoder.Unmarshal(&ref, enc)
	copy(p[:], ref)
	return err
}

//-------------------------------------

// Implements Signature
type SignatureSecp256k1 []byte

func (sig SignatureSecp256k1) Bytes() []byte {
	return wire.BinaryBytes(struct{ Signature }{sig})
}

func (sig SignatureSecp256k1) IsZero() bool { return len(sig) == 0 }

func (sig SignatureSecp256k1) String() string { return fmt.Sprintf("/%X.../", Fingerprint(sig[:])) }

func (sig SignatureSecp256k1) Equals(other Signature) bool {
	if otherEd, ok := other.(SignatureSecp256k1); ok {
		return bytes.Equal(sig[:], otherEd[:])
	} else {
		return false
	}
}
func (p SignatureSecp256k1) MarshalJSON() ([]byte, error) {
	return data.Encoder.Marshal(p)
}

func (p *SignatureSecp256k1) UnmarshalJSON(enc []byte) error {
	return data.Encoder.Unmarshal((*[]byte)(p), enc)
}
