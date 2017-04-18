package crypto

import (
	"bytes"
	"fmt"

	. "github.com/tendermint/tmlibs/common"
	data "github.com/tendermint/go-wire/data"
	"github.com/tendermint/go-wire"
)

func SignatureFromBytes(sigBytes []byte) (sig Signature, err error) {
	err = wire.ReadBinaryBytes(sigBytes, &sig)
	return
}

//----------------------------------------

type Signature struct {
	SignatureInner `json:"unwrap"`
}

// DO NOT USE THIS INTERFACE.
// You probably want to use Signature.
type SignatureInner interface {
	AssertIsSignatureInner()
	Bytes() []byte
	IsZero() bool
	String() string
	Equals(Signature) bool
	Wrap() Signature
}

func (sig Signature) MarshalJSON() ([]byte, error) {
	return sigMapper.ToJSON(sig.SignatureInner)
}

func (sig *Signature) UnmarshalJSON(data []byte) (err error) {
	parsed, err := sigMapper.FromJSON(data)
	if err == nil && parsed != nil {
		sig.SignatureInner = parsed.(SignatureInner)
	}
	return
}

// Unwrap recovers the concrete interface safely (regardless of levels of embeds)
func (sig Signature) Unwrap() SignatureInner {
	pk := sig.SignatureInner
	for wrap, ok := pk.(Signature); ok; wrap, ok = pk.(Signature) {
		pk = wrap.SignatureInner
	}
	return pk
}

func (sig Signature) Empty() bool {
	return sig.SignatureInner == nil
}

var sigMapper = data.NewMapper(Signature{}).
	RegisterImplementation(SignatureEd25519{}, NameEd25519, TypeEd25519).
	RegisterImplementation(SignatureSecp256k1{}, NameSecp256k1, TypeSecp256k1)

//-------------------------------------

var _ SignatureInner = SignatureEd25519{}

// Implements Signature
type SignatureEd25519 [64]byte

func (sig SignatureEd25519) AssertIsSignatureInner() {}

func (sig SignatureEd25519) Bytes() []byte {
	return wire.BinaryBytes(Signature{sig})
}

func (sig SignatureEd25519) IsZero() bool { return len(sig) == 0 }

func (sig SignatureEd25519) String() string { return fmt.Sprintf("/%X.../", Fingerprint(sig[:])) }

func (sig SignatureEd25519) Equals(other Signature) bool {
	if otherEd, ok := other.Unwrap().(SignatureEd25519); ok {
		return bytes.Equal(sig[:], otherEd[:])
	} else {
		return false
	}
}

func (sig SignatureEd25519) MarshalJSON() ([]byte, error) {
	return data.Encoder.Marshal(sig[:])
}

func (sig *SignatureEd25519) UnmarshalJSON(enc []byte) error {
	var ref []byte
	err := data.Encoder.Unmarshal(&ref, enc)
	copy(sig[:], ref)
	return err
}

func (sig SignatureEd25519) Wrap() Signature {
	return Signature{sig}
}

//-------------------------------------

var _ SignatureInner = SignatureSecp256k1{}

// Implements Signature
type SignatureSecp256k1 []byte

func (sig SignatureSecp256k1) AssertIsSignatureInner() {}

func (sig SignatureSecp256k1) Bytes() []byte {
	return wire.BinaryBytes(Signature{sig})
}

func (sig SignatureSecp256k1) IsZero() bool { return len(sig) == 0 }

func (sig SignatureSecp256k1) String() string { return fmt.Sprintf("/%X.../", Fingerprint(sig[:])) }

func (sig SignatureSecp256k1) Equals(other Signature) bool {
	if otherEd, ok := other.Unwrap().(SignatureSecp256k1); ok {
		return bytes.Equal(sig[:], otherEd[:])
	} else {
		return false
	}
}
func (sig SignatureSecp256k1) MarshalJSON() ([]byte, error) {
	return data.Encoder.Marshal(sig)
}

func (sig *SignatureSecp256k1) UnmarshalJSON(enc []byte) error {
	return data.Encoder.Unmarshal((*[]byte)(sig), enc)
}

func (sig SignatureSecp256k1) Wrap() Signature {
	return Signature{sig}
}
