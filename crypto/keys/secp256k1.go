package keys

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"golang.org/x/crypto/ripemd160" // nolint: staticcheck // necessary for Bitcoin address format

	"github.com/tendermint/tendermint/crypto"
)

const (
	PubKeySecp256k1Name  = "tendermint/PubKeySecp256k1"
	PrivKeySecp256k1Name = "tendermint/PrivKeySecp256k1"
)

var (
	_ crypto.PubKey  = PubKeySecp256K1{}
	_ crypto.PrivKey = PrivKeySecp256K1{}
)

const (
	// PubKeySecp256k1Size is comprised of 32 bytes for one field element
	// (the x-coordinate), plus one byte for the parity of the y-coordinate.
	PubKeySecp256k1Size = 33
	// PrivKeySecp256k1Size is the number of bytes in an Secp256k1 private key.
	PrivKeySecp256k1Size = 32
)

// used to reject malleable signatures
// see:
//  - https://github.com/ethereum/go-ethereum/blob/f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/signature_nocgo.go#L90-L93
//  - https://github.com/ethereum/go-ethereum/blob/f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/crypto.go#L39
var secp256k1halfN = new(big.Int).Rsh(btcec.S256().N, 1)

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

func (pubKey *PubKeySecp256K1) String() string {
	return fmt.Sprintf("%s{%X}", PubKeySecp256k1Name, pubKey.Bytes())
}

func (pubKey PubKeySecp256K1) Equals(other crypto.PubKey) bool {
	otherSecp, ok := other.(PubKeySecp256K1)
	if !ok {
		return false
	}
	return bytes.Equal(pubKey.bytes, otherSecp.bytes)
}

// Sign creates an ECDSA signature on curve Secp256k1, using SHA256 on the msg.
// The returned signature will be of the form R || S (in lower-S form).
func (privKey PubKeySecp256K1) Sign(msg []byte) ([]byte, error) {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKey.Bytes())
	sig, err := priv.Sign(crypto.Sha256(msg))
	if err != nil {
		return nil, err
	}
	sigBytes := serializeSig(sig)
	return sigBytes, nil
}

// VerifyBytes verifies a signature of the form R || S.
// It rejects signatures which are not in lower-S form.
func (pubKey PubKeySecp256K1) VerifyBytes(msg []byte, sigStr []byte) bool {
	if len(sigStr) != 64 {
		return false
	}
	pub, err := btcec.ParsePubKey(pubKey.Bytes(), btcec.S256())
	if err != nil {
		return false
	}
	// parse the signature:
	signature := signatureFromBytes(sigStr)
	// Reject malleable signatures. libsecp256k1 does this check but btcec doesn't.
	// see: https://github.com/ethereum/go-ethereum/blob/f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/signature_nocgo.go#L90-L93
	if signature.S.Cmp(secp256k1halfN) > 0 {
		return false
	}
	return signature.Verify(crypto.Sha256(msg), pub)
}

//-------------------------------------

// Bytes marshalls the private key using amino encoding.
func (privKey PrivKeySecp256K1) Bytes() []byte {
	if len(privKey.bytes) != PubKeySecp256k1Size {
		fmt.Errorf("invalid bytes length: got (%s), expected (%d)", len(privKey.bytes), PrivKeySecp256k1Size)
	}
	return privKey.bytes
}

func (privKey *PrivKeySecp256K1) String() string {
	return fmt.Sprintf("%s{%X}", PrivKeySecp256k1Name, privKey.Bytes()[:])
}

// PubKey performs the point-scalar multiplication from the privKey on the
// generator point to get the pubkey.
func (privKey PrivKeySecp256K1) PubKey() crypto.PubKey {
	_, pubkeyObject := btcec.PrivKeyFromBytes(secp256k1.S256(), privKey.Bytes()[:])
	var pubkeyBytes []byte
	copy(pubkeyBytes, pubkeyObject.SerializeCompressed())
	return PubKeySecp256K1{bytes: pubkeyBytes}
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (privKey PrivKeySecp256K1) Equals(other crypto.PrivKey) bool {
	otherSecp, ok := other.(PrivKeySecp256K1)
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare(privKey.bytes, otherSecp.bytes) == 1
}

// Sign creates an ECDSA signature on curve Secp256k1, using SHA256 on the msg.
func (privKey PrivKeySecp256K1) Sign(msg []byte) ([]byte, error) {
	rsv, err := secp256k1.Sign(crypto.Sha256(msg), privKey.Bytes())
	if err != nil {
		return nil, err
	}
	// we do not need v  in r||s||v:
	rs := rsv[:len(rsv)-1]
	return rs, nil
}

func (pubKey PrivKeySecp256K1) VerifyBytes(msg []byte, sig []byte) bool {
	return secp256k1.VerifySignature(pubKey.Bytes(), crypto.Sha256(msg), sig)
}

//-------------------------------------
// utils

// Read signature struct from R || S.
// CONTRACT: Caller needs to ensure that len(sigStr) == 64.
func signatureFromBytes(sigStr []byte) *btcec.Signature {
	return &btcec.Signature{
		R: new(big.Int).SetBytes(sigStr[:32]),
		S: new(big.Int).SetBytes(sigStr[32:64]),
	}
}

// Serialize signature to R || S.
// R, S are padded to 32 bytes respectively.
func serializeSig(sig *btcec.Signature) []byte {
	rBytes := sig.R.Bytes()
	sBytes := sig.S.Bytes()
	sigBytes := make([]byte, 64)
	// 0 pad the byte arrays from the left if they aren't big enough.
	copy(sigBytes[32-len(rBytes):32], rBytes)
	copy(sigBytes[64-len(sBytes):64], sBytes)
	return sigBytes
}
