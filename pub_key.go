package crypto

import (
	"bytes"

	"github.com/tendermint/ed25519"
	"github.com/tendermint/ed25519/extra25519"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	"golang.org/x/crypto/ripemd160"
)

// PubKey is part of Account and Validator.
type PubKey interface {
	Address() []byte
	KeyString() string
	VerifyBytes(msg []byte, sig Signature) bool
}

// Types of PubKey implementations
const (
	PubKeyTypeEd25519 = byte(0x01)
)

// for wire.readReflect
var _ = wire.RegisterInterface(
	struct{ PubKey }{},
	wire.ConcreteType{PubKeyEd25519{}, PubKeyTypeEd25519},
)

//-------------------------------------

// Implements PubKey
type PubKeyEd25519 [32]byte

// TODO: Slicing the array gives us length prefixing but loses the type byte.
// Revisit if we add more pubkey types.
// For now, we artificially append the type byte in front to give us backwards
// compatibility for when the pubkey wasn't fixed length array
func (pubKey PubKeyEd25519) Address() []byte {
	w, n, err := new(bytes.Buffer), new(int), new(error)
	wire.WriteBinary(pubKey[:], w, n, err)
	if *err != nil {
		PanicCrisis(*err)
	}
	// append type byte
	encodedPubkey := append([]byte{1}, w.Bytes()...)
	hasher := ripemd160.New()
	hasher.Write(encodedPubkey) // does not error
	return hasher.Sum(nil)
}

// TODO: Consider returning a reason for failure, or logging a runtime type mismatch.
func (pubKey PubKeyEd25519) VerifyBytes(msg []byte, sig_ Signature) bool {
	sig, ok := sig_.(SignatureEd25519)
	if !ok {
		return false
	}
	pubKeyBytes := [32]byte(pubKey)
	sigBytes := [64]byte(sig)
	return ed25519.Verify(&pubKeyBytes, msg, &sigBytes)
}

// For use with golang/crypto/nacl/box
// If error, returns nil.
func (pubKey PubKeyEd25519) ToCurve25519() *[32]byte {
	keyCurve25519, pubKeyBytes := new([32]byte), [32]byte(pubKey)
	ok := extra25519.PublicKeyToCurve25519(keyCurve25519, &pubKeyBytes)
	if !ok {
		return nil
	}
	return keyCurve25519
}

func (pubKey PubKeyEd25519) String() string {
	return Fmt("PubKeyEd25519{%X}", pubKey[:])
}

// Must return the full bytes in hex.
// Used for map keying, etc.
func (pubKey PubKeyEd25519) KeyString() string {
	return Fmt("%X", pubKey[:])
}

func (pubKey PubKeyEd25519) Equals(other PubKey) bool {
	if otherEd, ok := other.(PubKeyEd25519); ok {
		return bytes.Equal(pubKey[:], otherEd[:])
	} else {
		return false
	}
}
