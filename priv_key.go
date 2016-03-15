package crypto

import (
	"github.com/tendermint/ed25519"
	"github.com/tendermint/ed25519/extra25519"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
)

// PrivKey is part of PrivAccount and state.PrivValidator.
type PrivKey interface {
	Bytes() []byte
	Sign(msg []byte) Signature
	PubKey() PubKey
}

// Types of PrivKey implementations
const (
	PrivKeyTypeEd25519 = byte(0x01)
)

// for wire.readReflect
var _ = wire.RegisterInterface(
	struct{ PrivKey }{},
	wire.ConcreteType{PrivKeyEd25519{}, PrivKeyTypeEd25519},
)

func PrivKeyFromBytes(privKeyBytes []byte) (privKey PrivKey, err error) {
	err = wire.ReadBinaryBytes(privKeyBytes, &privKey)
	return
}

//-------------------------------------

// Implements PrivKey
type PrivKeyEd25519 [64]byte

func (privKey PrivKeyEd25519) Bytes() []byte {
	return wire.BinaryBytes(struct{ PrivKey }{privKey})
}

func (key PrivKeyEd25519) Sign(msg []byte) Signature {
	privKeyBytes := [64]byte(key)
	signatureBytes := ed25519.Sign(&privKeyBytes, msg)
	return SignatureEd25519(*signatureBytes)
}

func (privKey PrivKeyEd25519) PubKey() PubKey {
	privKeyBytes := [64]byte(privKey)
	return PubKeyEd25519(*ed25519.MakePublicKey(&privKeyBytes))
}

func (privKey PrivKeyEd25519) ToCurve25519() *[32]byte {
	keyCurve25519 := new([32]byte)
	privKeyBytes := [64]byte(privKey)
	extra25519.PrivateKeyToCurve25519(keyCurve25519, &privKeyBytes)
	return keyCurve25519
}

func (privKey PrivKeyEd25519) String() string {
	return Fmt("PrivKeyEd25519{*****}")
}

// Deterministically generates new priv-key bytes from key.
func (key PrivKeyEd25519) Generate(index int) PrivKeyEd25519 {
	newBytes := wire.BinarySha256(struct {
		PrivKey [64]byte
		Index   int
	}{key, index})
	var newKey [64]byte
	copy(newKey[:], newBytes)
	return PrivKeyEd25519(newKey)
}

func GenPrivKeyEd25519() PrivKeyEd25519 {
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], CRandBytes(32))
	ed25519.MakePublicKey(privKeyBytes)
	return PrivKeyEd25519(*privKeyBytes)
}

// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeyEd25519FromSecret(secret []byte) PrivKeyEd25519 {
	privKey32 := Sha256(secret) // Not Ripemd160 because we want 32 bytes.
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], privKey32)
	ed25519.MakePublicKey(privKeyBytes)
	return PrivKeyEd25519(*privKeyBytes)
}
