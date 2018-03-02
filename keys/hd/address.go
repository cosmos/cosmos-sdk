package hd

// XXX This package doesn't work with our address scheme,
// XXX and it probably doesn't work for our other pubkey types.
// XXX Fix it up to be more general but compatible.

import (
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"math/big"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

/*

  This file implements BIP32 HD wallets.
  Note it only works for SECP256k1 keys.
  It also includes some Bitcoin specific utility functions.

*/

// ComputeBTCAddress returns the BTC address using the pubKeyHex and chainCodeHex
// for the given path and index.
func ComputeBTCAddress(pubKeyHex string, chainCodeHex string, path string, index int32) string {
	pubKeyBytes := DerivePublicKeyForPath(
		HexDecode(pubKeyHex),
		HexDecode(chainCodeHex),
		fmt.Sprintf("%v/%v", path, index),
	)
	return BTCAddrFromPubKeyBytes(pubKeyBytes)
}

// ComputePrivateKey returns the private key using the master mprivHex and chainCodeHex
// for the given path and index.
func ComputePrivateKey(mprivHex string, chainHex string, path string, index int32) string {
	privKeyBytes := DerivePrivateKeyForPath(
		HexDecode(mprivHex),
		HexDecode(chainHex),
		fmt.Sprintf("%v/%v", path, index),
	)
	return HexEncode(privKeyBytes)
}

// ComputeBTCAddressForPrivKey returns the Bitcoin address for the given privKey.
func ComputeBTCAddressForPrivKey(privKey string) string {
	pubKeyBytes := PubKeyBytesFromPrivKeyBytes(HexDecode(privKey), true)
	return BTCAddrFromPubKeyBytes(pubKeyBytes)
}

// SignBTCMessage signs a "Bitcoin Signed Message".
func SignBTCMessage(privKey string, message string, compress bool) string {
	prefixBytes := []byte("Bitcoin Signed Message:\n")
	messageBytes := []byte(message)
	bytes := []byte{}
	bytes = append(bytes, byte(len(prefixBytes)))
	bytes = append(bytes, prefixBytes...)
	bytes = append(bytes, byte(len(messageBytes)))
	bytes = append(bytes, messageBytes...)
	privKeyBytes := HexDecode(privKey)
	x, y := btcec.S256().ScalarBaseMult(privKeyBytes)
	ecdsaPubKey := ecdsa.PublicKey{
		Curve: btcec.S256(),
		X:     x,
		Y:     y,
	}
	ecdsaPrivKey := &btcec.PrivateKey{
		PublicKey: ecdsaPubKey,
		D:         new(big.Int).SetBytes(privKeyBytes),
	}
	sigbytes, err := btcec.SignCompact(btcec.S256(), ecdsaPrivKey, CalcHash256(bytes), compress)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(sigbytes)
}

// ComputeMastersFromSeed returns the master public key, master secret, and chain code in hex.
func ComputeMastersFromSeed(seed string) (string, string, string) {
	key, data := []byte("Bitcoin seed"), []byte(seed)
	secret, chain := I64(key, data)
	pubKeyBytes := PubKeyBytesFromPrivKeyBytes(secret, true)
	return HexEncode(pubKeyBytes), HexEncode(secret), HexEncode(chain)
}

// ComputeWIF returns the privKey in Wallet Import Format.
func ComputeWIF(privKey string, compress bool) string {
	return WIFFromPrivKeyBytes(HexDecode(privKey), compress)
}

// ComputeBTCTxId returns the bitcoin transaction ID.
func ComputeBTCTxId(rawTxHex string) string {
	return HexEncode(ReverseBytes(CalcHash256(HexDecode(rawTxHex))))
}

/*
func printKeyInfo(privKeyBytes []byte, pubKeyBytes []byte, chain []byte) {
	if pubKeyBytes == nil {
		pubKeyBytes = PubKeyBytesFromPrivKeyBytes(privKeyBytes, true)
	}
	addr := AddrFromPubKeyBytes(pubKeyBytes)
	log.Println("\nprikey:\t%v\npubKeyBytes:\t%v\naddr:\t%v\nchain:\t%v",
		HexEncode(privKeyBytes),
		HexEncode(pubKeyBytes),
		addr,
		HexEncode(chain))
}
*/

//-------------------------------------------------------------------

// DerivePrivateKeyForPath derives the private key by following the path from privKeyBytes,
// using the given chainCode.
func DerivePrivateKeyForPath(privKeyBytes []byte, chainCode []byte, path string) []byte {
	data := privKeyBytes
	parts := strings.Split(path, "/")
	for _, part := range parts {
		prime := part[len(part)-1:] == "'"
		// prime == private derivation. Otherwise public.
		if prime {
			part = part[:len(part)-1]
		}
		i, err := strconv.Atoi(part)
		if err != nil {
			panic(err)
		}
		if i < 0 {
			panic(errors.New("index too large."))
		}
		data, chainCode = DerivePrivateKey(data, chainCode, uint32(i), prime)
		//printKeyInfo(data, nil, chain)
	}
	return data
}

// DerivePublicKeyForPath derives the public key by following the path from pubKeyBytes
// using the given chainCode.
func DerivePublicKeyForPath(pubKeyBytes []byte, chainCode []byte, path string) []byte {
	data := pubKeyBytes
	parts := strings.Split(path, "/")
	for _, part := range parts {
		prime := part[len(part)-1:] == "'"
		if prime {
			panic(errors.New("cannot do a prime derivation from public key"))
		}
		i, err := strconv.Atoi(part)
		if err != nil {
			panic(err)
		}
		if i < 0 {
			panic(errors.New("index too large."))
		}
		data, chainCode = DerivePublicKey(data, chainCode, uint32(i))
		//printKeyInfo(nil, data, chainCode)
	}
	return data
}

// DerivePrivateKey derives the private key with index and chainCode.
// If prime is true, the derivation is 'hardened'.
// It returns the new private key and new chain code.
func DerivePrivateKey(privKeyBytes []byte, chainCode []byte, index uint32, prime bool) ([]byte, []byte) {
	var data []byte
	if prime {
		index = index | 0x80000000
		data = append([]byte{byte(0)}, privKeyBytes...)
	} else {
		public := PubKeyBytesFromPrivKeyBytes(privKeyBytes, true)
		data = public
	}
	data = append(data, uint32ToBytes(index)...)
	data2, chainCode2 := I64(chainCode, data)
	x := addScalars(privKeyBytes, data2)
	return x, chainCode2
}

// DerivePublicKey derives the public key with index and chainCode.
// It returns the new public key and new chain code.
func DerivePublicKey(pubKeyBytes []byte, chainCode []byte, index uint32) ([]byte, []byte) {
	data := []byte{}
	data = append(data, pubKeyBytes...)
	data = append(data, uint32ToBytes(index)...)
	data2, chainCode2 := I64(chainCode, data)
	data2p := PubKeyBytesFromPrivKeyBytes(data2, true)
	return addPoints(pubKeyBytes, data2p), chainCode2
}

// eliptic curve pubkey addition
func addPoints(a []byte, b []byte) []byte {
	ap, err := btcec.ParsePubKey(a, btcec.S256())
	if err != nil {
		panic(err)
	}
	bp, err := btcec.ParsePubKey(b, btcec.S256())
	if err != nil {
		panic(err)
	}
	sumX, sumY := btcec.S256().Add(ap.X, ap.Y, bp.X, bp.Y)
	sum := &btcec.PublicKey{
		Curve: btcec.S256(),
		X:     sumX,
		Y:     sumY,
	}
	return sum.SerializeCompressed()
}

// modular big endian addition
func addScalars(a []byte, b []byte) []byte {
	aInt := new(big.Int).SetBytes(a)
	bInt := new(big.Int).SetBytes(b)
	sInt := new(big.Int).Add(aInt, bInt)
	x := sInt.Mod(sInt, btcec.S256().N).Bytes()
	x2 := [32]byte{}
	copy(x2[32-len(x):], x)
	return x2[:]
}

func uint32ToBytes(i uint32) []byte {
	b := [4]byte{}
	binary.BigEndian.PutUint32(b[:], i)
	return b[:]
}

//-------------------------------------------------------------------

// HexEncode encodes b in hex.
func HexEncode(b []byte) string {
	return hex.EncodeToString(b)
}

// HexDecode hex decodes the str. If str is not valid hex
// it will return an empty byte slice.
func HexDecode(str string) []byte {
	b, _ := hex.DecodeString(str)
	return b
}

// I64 returns the two halfs of the SHA512 HMAC of key and data.
func I64(key []byte, data []byte) ([]byte, []byte) {
	mac := hmac.New(sha512.New, key)
	mac.Write(data)
	I := mac.Sum(nil)
	return I[:32], I[32:]
}

//-------------------------------------------------------------------

const (
	btcPrefixPubKeyHash = byte(0x00)
	btcPrefixPrivKey    = byte(0x80)
)

// BTCAddrFromPubKeyBytes returns a B58 encoded Bitcoin mainnet address.
func BTCAddrFromPubKeyBytes(pubKeyBytes []byte) string {
	versionPrefix := btcPrefixPubKeyHash // TODO Make const or configurable
	h160 := CalcHash160(pubKeyBytes)
	h160 = append([]byte{versionPrefix}, h160...)
	checksum := CalcHash256(h160)
	b := append(h160, checksum[:4]...)
	return base58.Encode(b)
}

// BTCAddrBytesFromPubKeyBytes returns a hex Bitcoin mainnet address and its checksum.
func BTCAddrBytesFromPubKeyBytes(pubKeyBytes []byte) (addrBytes []byte, checksum []byte) {
	versionPrefix := btcPrefixPubKeyHash // TODO Make const or configurable
	h160 := CalcHash160(pubKeyBytes)
	_h160 := append([]byte{versionPrefix}, h160...)
	checksum = CalcHash256(_h160)[:4]
	return h160, checksum
}

// WIFFromPrivKeyBytes returns the privKeyBytes in Wallet Import Format.
func WIFFromPrivKeyBytes(privKeyBytes []byte, compress bool) string {
	versionPrefix := btcPrefixPrivKey // TODO Make const or configurable
	bytes := append([]byte{versionPrefix}, privKeyBytes...)
	if compress {
		bytes = append(bytes, byte(1))
	}
	checksum := CalcHash256(bytes)
	bytes = append(bytes, checksum[:4]...)
	return base58.Encode(bytes)
}

// PubKeyBytesFromPrivKeyBytes returns the optionally compressed public key bytes.
func PubKeyBytesFromPrivKeyBytes(privKeyBytes []byte, compress bool) (pubKeyBytes []byte) {
	x, y := btcec.S256().ScalarBaseMult(privKeyBytes)
	pub := &btcec.PublicKey{
		Curve: btcec.S256(),
		X:     x,
		Y:     y,
	}

	if compress {
		return pub.SerializeCompressed()
	}
	return pub.SerializeUncompressed()
}

//--------------------------------------------------------------

// CalcHash returns the hash of data using hasher.
func CalcHash(data []byte, hasher hash.Hash) []byte {
	hasher.Write(data)
	return hasher.Sum(nil)
}

// CalcHash160 returns the ripemd160(sha256(data)).
func CalcHash160(data []byte) []byte {
	return CalcHash(CalcHash(data, sha256.New()), ripemd160.New())
}

// CalcHash256 returns the sha256(sha256(data)).
func CalcHash256(data []byte) []byte {
	return CalcHash(CalcHash(data, sha256.New()), sha256.New())
}

// CalcSha512 returns the sha512(data).
func CalcSha512(data []byte) []byte {
	return CalcHash(data, sha512.New())
}

// ReverseBytes returns the buf in the opposite order
func ReverseBytes(buf []byte) []byte {
	var res []byte
	if len(buf) == 0 {
		return res
	}

	// Walk till mid-way, swapping bytes from each end:
	// b[i] and b[len-i-1]
	blen := len(buf)
	res = make([]byte, blen)
	mid := blen / 2
	for left := 0; left <= mid; left++ {
		right := blen - left - 1
		res[left] = buf[right]
		res[right] = buf[left]
	}
	return res
}
