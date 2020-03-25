package keys

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"

	secp256k1 "github.com/btcsuite/btcd/btcec"
	"golang.org/x/crypto/ripemd160" // nolint: staticcheck // necessary for Bitcoin address format

	"github.com/tendermint/tendermint/crypto"
)

const (
	PubKeySecp256k1Name  = "tendermint/PubKeySecp256k1"
	PrivKeySecp256k1Name = "tendermint/PrivKeySecp256k1"
)

var (
	_ crypto.PubKey  = PubKeySecp256K1{}
	_ crypto.PrivKey = PubKeySecp256K1{}
)

const (
	// PubKeySecp256k1Size is comprised of 32 bytes for one field element
	// (the x-coordinate), plus one byte for the parity of the y-coordinate.
	PubKeySecp256k1Size = 33
	// PrivKeySecp256k1Size is the number of bytes in an Secp256k1 private key.
	PrivKeySecp256k1Size = 32
)

// Address returns a Bitcoin style addresses: RIPEMD160(SHA256(pubkey))
func (pubKey PubKeySecp256K1) Address() crypto.Address {
	hasherSHA256 := sha256.New()
	_, err := hasherSHA256.Write(pubKey.Bytes())
	if err != nil {
		panic(err)
	}
	sha := hasherSHA256.Sum(nil)

	hasherRIPEMD160 := ripemd160.New()
	_, err = hasherRIPEMD160.Write(sha)
	if err != nil {
		panic(err)
	}

	return crypto.Address(hasherRIPEMD160.Sum(nil))
}

// Bytes returns the pubkey bytes.
func (pubKey PubKeySecp256K1) Bytes() []byte {
	if len(pubKey.bytes) != PubKeySecp256k1Size {
		fmt.Errorf("invalid bytes length: got (%s), expected (%d)", len(pubKey.bytes), PubKeySecp256k1Size)
	}
	return pubKey.bytes
}

func (pubKey PubKeySecp256K1) String() string {
	return fmt.Sprintf("%s{%X}", PubKeySecp256k1Name, pubKey.Bytes())
}

func (pubKey PubKeySecp256K1) Equals(other crypto.PubKey) bool {
	if otherSecp, ok := other.(PubKeySecp256K1); ok {
		return bytes.Equal(pubKey.bytes, otherSecp.bytes)
	}
	return false
}

//-------------------------------------

// Bytes marshalls the private key using amino encoding.
func (privKey PrivKeySecp256K1) Bytes() []byte {
	if len(privKey.bytes) != PubKeySecp256k1Size {
		fmt.Errorf("invalid bytes length: got (%s), expected (%d)", len(privKey.bytes), PrivKeySecp256k1Size)
	}
	return privKey.bytes
}

// PubKey performs the point-scalar multiplication from the privKey on the
// generator point to get the pubkey.
func (privKey PrivKeySecp256K1) PubKey() crypto.PubKey {
	_, pubkeyObject := secp256k1.PrivKeyFromBytes(secp256k1.S256(), privKey.Bytes()[:])
	var pubkeyBytes []byte
	copy(pubkeyBytes, pubkeyObject.SerializeCompressed())
	return PubKeySecp256K1{bytes: pubkeyBytes}
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (privKey PrivKeySecp256K1) Equals(other crypto.PrivKey) bool {
	if otherSecp, ok := other.(PrivKeySecp256K1); ok {
		return subtle.ConstantTimeCompare(privKey.bytes, otherSecp.bytes) == 1
	}
	return false
}
