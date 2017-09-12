package nano

import (
	"bytes"
	"encoding/hex"

	crypto "github.com/tendermint/go-crypto"
)

// // Implements PrivKey, calling the ledger nano
// type PrivKeyLedger struct{}

// var _ PrivKeyInner = PrivKeyLedger{}

// func (privKey PrivKeyLedger) AssertIsPrivKeyInner() {}

// func (privKey PrivKeyLedger) Bytes() []byte {
// 	return wire.BinaryBytes(PrivKey{privKey})
// }

// func (privKey PrivKeyLedger) Sign(msg []byte) Signature {
// 	privKeyBytes := [64]byte(privKey)
// 	signatureBytes := ed25519.Sign(&privKeyBytes, msg)
// 	return SignatureEd25519(*signatureBytes).Wrap()
// }

// func (privKey PrivKeyLedger) PubKey() PubKey {
// 	privKeyBytes := [64]byte(privKey)
// 	pubBytes := *ed25519.MakePublicKey(&privKeyBytes)
// 	return PubKeyEd25519(pubBytes).Wrap()
// }

// func (privKey PrivKeyLedger) Equals(other PrivKey) bool {
// 	if otherEd, ok := other.Unwrap().(PrivKeyLedger); ok {
// 		return bytes.Equal(privKey[:], otherEd[:])
// 	} else {
// 		return false
// 	}
// }

// MockPrivKeyLedger behaves as the ledger, but stores a pre-packaged call-response
// for use in test cases
type MockPrivKeyLedger struct {
	Msg []byte
	Pub [KeyLength]byte
	Sig [SigLength]byte
}

// NewMockKey returns
func NewMockKey(msg, pubkey, sig string) (pk MockPrivKeyLedger) {
	var err error
	pk.Msg, err = hex.DecodeString(msg)
	if err != nil {
		panic(err)
	}

	bpk, err := hex.DecodeString(pubkey)
	if err != nil {
		panic(err)
	}
	bsig, err := hex.DecodeString(sig)
	if err != nil {
		panic(err)
	}

	copy(pk.Pub[:], bpk)
	copy(pk.Sig[:], bsig)
	return pk
}

var _ crypto.PrivKeyInner = MockPrivKeyLedger{}

// AssertIsPrivKeyInner fulfils PrivKey Interface
func (pk MockPrivKeyLedger) AssertIsPrivKeyInner() {}

// Bytes fulfils PrivKey Interface - not supported
func (pk MockPrivKeyLedger) Bytes() []byte {
	return nil
}

// Sign returns a real SignatureLedger, if the msg matches what we expect
func (pk MockPrivKeyLedger) Sign(msg []byte) crypto.Signature {
	if !bytes.Equal(pk.Msg, msg) {
		panic("Mock key is for different msg")
	}
	return crypto.SignatureEd25519(pk.Sig).Wrap()
}

// PubKey returns a real PubKeyLedger, that will verify this signature
func (pk MockPrivKeyLedger) PubKey() crypto.PubKey {
	return PubKeyLedger{crypto.PubKeyEd25519(pk.Pub)}.Wrap()
}

// Equals compares that two Mocks have the same data
func (pk MockPrivKeyLedger) Equals(other crypto.PrivKey) bool {
	if mock, ok := other.Unwrap().(MockPrivKeyLedger); ok {
		return bytes.Equal(mock.Pub[:], pk.Pub[:]) &&
			bytes.Equal(mock.Sig[:], pk.Sig[:]) &&
			bytes.Equal(mock.Msg, pk.Msg)
	}
	return false
}

////////////////////////////////////////////
// pubkey

// PubKeyLedger works like a normal Ed25519 except a hash before the verify bytes
type PubKeyLedger struct {
	crypto.PubKeyEd25519
}

// VerifyBytes uses the normal Ed25519 algorithm but a sha512 hash beforehand
func (pk PubKeyLedger) VerifyBytes(msg []byte, sig crypto.Signature) bool {
	hmsg := hashMsg(msg)
	return pk.PubKeyEd25519.VerifyBytes(hmsg, sig)
}

// Equals implements PubKey interface
func (pk PubKeyLedger) Equals(other crypto.PubKey) bool {
	if ledger, ok := other.Unwrap().(PubKeyLedger); ok {
		return bytes.Equal(pk.PubKeyEd25519[:], ledger.PubKeyEd25519[:])
	}
	return false
}

/*** registration with go-data ***/

func init() {
	crypto.PrivKeyMapper.
		// RegisterImplementation(PrivKeyLedger{}, "ledger", 0x10).
		RegisterImplementation(MockPrivKeyLedger{}, "mock-ledger", 0x11)

	crypto.PubKeyMapper.
		RegisterImplementation(PubKeyLedger{}, "ledger", 0x10)
}

// // Wrap fulfils interface for PrivKey struct
// func (hi PrivKeyLedger) Wrap() crypto.PrivKey {
// 	return PrivKey{hi}
// }

// Wrap fulfils interface for PrivKey struct
func (pk MockPrivKeyLedger) Wrap() crypto.PrivKey {
	return crypto.PrivKey{pk}
}

// Wrap fulfils interface for PubKey struct
func (pk PubKeyLedger) Wrap() crypto.PubKey {
	return crypto.PubKey{pk}
}
