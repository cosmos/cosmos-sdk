package crypto

import (
	"bytes"
	"fmt"

	"github.com/tendermint/go-wire"
	data "github.com/tendermint/go-wire/data"
	. "github.com/tendermint/tmlibs/common"
)

func SignatureFromBytes(sigBytes []byte) (sig Signature, err error) {
	err = wire.ReadBinaryBytes(sigBytes, &sig)
	return
}

//----------------------------------------

// DO NOT USE THIS INTERFACE.
// You probably want to use Signature.
// +gen wrapper:"Signature,Impl[SignatureEd25519,SignatureSecp256k1],ed25519,secp256k1"
type SignatureInner interface {
	AssertIsSignatureInner()
	Bytes() []byte
	IsZero() bool
	Equals(Signature) bool
	Wrap() Signature
}

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

func SignatureEd25519FromBytes(data []byte) Signature {
	var sig SignatureEd25519
	copy(sig[:], data)
	return sig.Wrap()
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
	if otherSecp, ok := other.Unwrap().(SignatureSecp256k1); ok {
		return bytes.Equal(sig[:], otherSecp[:])
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
