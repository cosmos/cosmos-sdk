package crypto

import (
	"fmt"

	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
)

// Signature is a part of Txs and consensus Votes.
type Signature interface {
	Bytes() []byte
	IsZero() bool
	String() string
}

// Types of Signature implementations
const (
	SignatureTypeEd25519   = byte(0x01)
	SignatureTypeSecp256k1 = byte(0x02)
)

// for wire.readReflect
var _ = wire.RegisterInterface(
	struct{ Signature }{},
	wire.ConcreteType{SignatureEd25519{}, SignatureTypeEd25519},
	wire.ConcreteType{SignatureSecp256k1{}, SignatureTypeSecp256k1},
)

//-------------------------------------

// Implements Signature
type SignatureEd25519 [64]byte

func (sig SignatureEd25519) Bytes() []byte {
	return wire.BinaryBytes(struct{ Signature }{sig})
}

func (sig SignatureEd25519) IsZero() bool { return len(sig) == 0 }

func (sig SignatureEd25519) String() string { return fmt.Sprintf("/%X.../", Fingerprint(sig[:])) }

//-------------------------------------

// Implements Signature
type SignatureSecp256k1 []byte

func (sig SignatureSecp256k1) Bytes() []byte {
	return wire.BinaryBytes(struct{ Signature }{sig})
}

func (sig SignatureSecp256k1) IsZero() bool { return len(sig) == 0 }

func (sig SignatureSecp256k1) String() string { return fmt.Sprintf("/%X.../", Fingerprint(sig[:])) }
