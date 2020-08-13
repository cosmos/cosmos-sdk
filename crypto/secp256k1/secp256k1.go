package secp256k1

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"io"
	"math/big"

	secp256k1 "github.com/btcsuite/btcd/btcec"
	"golang.org/x/crypto/ripemd160" // nolint: staticcheck // necessary for Bitcoin address format

	amino "github.com/tendermint/go-amino"

	"github.com/tendermint/tendermint/crypto"
)

//-------------------------------------
const (
	PrivKeyAminoName = "tendermint/PrivKeySecp256k1"
	PubKeyAminoName  = "tendermint/PubKeySecp256k1"
)

var cdc = amino.NewCodec()

func init() {
	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	cdc.RegisterConcrete(PubKeySecp256k1{},
		PubKeyAminoName, nil)

	cdc.RegisterInterface((*crypto.PrivKey)(nil), nil)
	cdc.RegisterConcrete(PrivKeySecp256k1{},
		PrivKeyAminoName, nil)
}

//-------------------------------------

var _ crypto.PrivKey = PrivKeySecp256k1{}

// PrivKeySecp256k1 implements PrivKey.
type PrivKeySecp256k1 [32]byte

// Bytes marshalls the private key using amino encoding.
func (privKey PrivKeySecp256k1) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(privKey)
}

// PubKey performs the point-scalar multiplication from the privKey on the
// generator point to get the pubkey.
func (privKey PrivKeySecp256k1) PubKey() crypto.PubKey {
	_, pubkeyObject := secp256k1.PrivKeyFromBytes(secp256k1.S256(), privKey[:])
	var pubkeyBytes PubKeySecp256k1
	copy(pubkeyBytes[:], pubkeyObject.SerializeCompressed())
	return pubkeyBytes
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (privKey PrivKeySecp256k1) Equals(other crypto.PrivKey) bool {
	if otherSecp, ok := other.(PrivKeySecp256k1); ok {
		return subtle.ConstantTimeCompare(privKey[:], otherSecp[:]) == 1
	}
	return false
}

// GenPrivKey generates a new ECDSA private key on curve secp256k1 private key.
// It uses OS randomness to generate the private key.
func GenPrivKey() PrivKeySecp256k1 {
	return genPrivKey(crypto.CReader())
}

// genPrivKey generates a new secp256k1 private key using the provided reader.
func genPrivKey(rand io.Reader) PrivKeySecp256k1 {
	var privKeyBytes [32]byte
	d := new(big.Int)
	for {
		privKeyBytes = [32]byte{}
		_, err := io.ReadFull(rand, privKeyBytes[:])
		if err != nil {
			panic(err)
		}

		d.SetBytes(privKeyBytes[:])
		// break if we found a valid point (i.e. > 0 and < N == curverOrder)
		isValidFieldElement := 0 < d.Sign() && d.Cmp(secp256k1.S256().N) < 0
		if isValidFieldElement {
			break
		}
	}

	return PrivKeySecp256k1(privKeyBytes)
}

var one = new(big.Int).SetInt64(1)

// GenPrivKeySecp256k1 hashes the secret with SHA2, and uses
// that 32 byte output to create the private key.
//
// It makes sure the private key is a valid field element by setting:
//
// c = sha256(secret)
// k = (c mod (n − 1)) + 1, where n = curve order.
//
// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeySecp256k1(secret []byte) PrivKeySecp256k1 {
	secHash := sha256.Sum256(secret)
	// to guarantee that we have a valid field element, we use the approach of:
	// "Suite B Implementer’s Guide to FIPS 186-3", A.2.1
	// https://apps.nsa.gov/iaarchive/library/ia-guidance/ia-solutions-for-classified/algorithm-guidance/suite-b-implementers-guide-to-fips-186-3-ecdsa.cfm
	// see also https://github.com/golang/go/blob/0380c9ad38843d523d9c9804fe300cb7edd7cd3c/src/crypto/ecdsa/ecdsa.go#L89-L101
	fe := new(big.Int).SetBytes(secHash[:])
	n := new(big.Int).Sub(secp256k1.S256().N, one)
	fe.Mod(fe, n)
	fe.Add(fe, one)

	feB := fe.Bytes()
	var privKey32 [32]byte
	// copy feB over to fixed 32 byte privKey32 and pad (if necessary)
	copy(privKey32[32-len(feB):32], feB)

	return PrivKeySecp256k1(privKey32)
}

//-------------------------------------

var _ crypto.PubKey = PubKeySecp256k1{}

// PubKeySecp256k1Size is comprised of 32 bytes for one field element
// (the x-coordinate), plus one byte for the parity of the y-coordinate.
const PubKeySecp256k1Size = 33

// PubKeySecp256k1 implements crypto.PubKey.
// It is the compressed form of the pubkey. The first byte depends is a 0x02 byte
// if the y-coordinate is the lexicographically largest of the two associated with
// the x-coordinate. Otherwise the first byte is a 0x03.
// This prefix is followed with the x-coordinate.
type PubKeySecp256k1 [PubKeySecp256k1Size]byte

// Address returns a Bitcoin style addresses: RIPEMD160(SHA256(pubkey))
func (pubKey PubKeySecp256k1) Address() crypto.Address {
	hasherSHA256 := sha256.New()
	hasherSHA256.Write(pubKey[:]) // does not error
	sha := hasherSHA256.Sum(nil)

	hasherRIPEMD160 := ripemd160.New()
	hasherRIPEMD160.Write(sha) // does not error
	return crypto.Address(hasherRIPEMD160.Sum(nil))
}

// Bytes returns the pubkey marshalled with amino encoding.
func (pubKey PubKeySecp256k1) Bytes() []byte {
	bz, err := cdc.MarshalBinaryBare(pubKey)
	if err != nil {
		panic(err)
	}
	return bz
}

func (pubKey PubKeySecp256k1) String() string {
	return fmt.Sprintf("PubKeySecp256k1{%X}", pubKey[:])
}

func (pubKey PubKeySecp256k1) Equals(other crypto.PubKey) bool {
	if otherSecp, ok := other.(PubKeySecp256k1); ok {
		return bytes.Equal(pubKey[:], otherSecp[:])
	}
	return false
}
