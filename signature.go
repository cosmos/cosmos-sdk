package crypto

import (
	"crypto/subtle"
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
		// It is essential that we constant time compare
		// private keys and signatures instead of bytes.Equal,
		// to avoid susceptibility to timing/side channel attacks.
		// See Issue https://github.com/tendermint/go-crypto/issues/43
		return subtle.ConstantTimeCompare(sig[:], otherEd[:]) == 0
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
		// It is essential that we constant time compare
		// private keys and signatures instead of bytes.Equal,
		// to avoid susceptibility to timing/side channel attacks.
		// See Issue https://github.com/tendermint/go-crypto/issues/43
		return subtle.ConstantTimeCompare(sig[:], otherEd[:]) == 0
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
