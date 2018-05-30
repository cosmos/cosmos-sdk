package crypto

import (
	"bytes"
	"crypto/sha512"
	"fmt"

	"github.com/tendermint/ed25519"
	"github.com/tendermint/ed25519/extra25519"
	"golang.org/x/crypto/ripemd160"
)

var _ PubKey = PubKeyLedgerEd25519{}

// Implements PubKeyInner
type PubKeyLedgerEd25519 [32]byte

func (pubKey PubKeyLedgerEd25519) Address() Address {
	// append type byte
	hasher := ripemd160.New()
	hasher.Write(pubKey.Bytes()) // does not error
	return Address(hasher.Sum(nil))
}

func (pubKey PubKeyLedgerEd25519) Bytes() []byte {
	bz, err := cdc.MarshalBinaryBare(pubKey)
	if err != nil {
		panic(err)
	}
	return bz
}

func (pubKey PubKeyLedgerEd25519) VerifyBytes(msg []byte, sig_ Signature) bool {
	// must verify sha512 hash of msg, no padding, for Ledger compatibility
	sig, ok := sig_.(SignatureEd25519)
	if !ok {
		return false
	}
	pubKeyBytes := [32]byte(pubKey)
	sigBytes := [64]byte(sig)
	h := sha512.New()
	h.Write(msg)
	digest := h.Sum(nil)
	return ed25519.Verify(&pubKeyBytes, digest, &sigBytes)
}

// For use with golang/crypto/nacl/box
// If error, returns nil.
func (pubKey PubKeyLedgerEd25519) ToCurve25519() *[32]byte {
	keyCurve25519, pubKeyBytes := new([32]byte), [32]byte(pubKey)
	ok := extra25519.PublicKeyToCurve25519(keyCurve25519, &pubKeyBytes)
	if !ok {
		return nil
	}
	return keyCurve25519
}

func (pubKey PubKeyLedgerEd25519) String() string {
	return fmt.Sprintf("PubKeyLedgerEd25519{%X}", pubKey[:])
}

func (pubKey PubKeyLedgerEd25519) Equals(other PubKey) bool {
	if otherEd, ok := other.(PubKeyLedgerEd25519); ok {
		return bytes.Equal(pubKey[:], otherEd[:])
	} else {
		return false
	}
}
