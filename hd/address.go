package hd

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
	"github.com/tendermint/go-crypto"
	"golang.org/x/crypto/ripemd160"
)

func ComputeAddress(pubKeyHex string, chainHex string, path string, index int32) string {
	pubKeyBytes := DerivePublicKeyForPath(
		HexDecode(pubKeyHex),
		HexDecode(chainHex),
		fmt.Sprintf("%v/%v", path, index),
	)
	return AddrFromPubKeyBytes(pubKeyBytes)
}

func ComputePrivateKey(mprivHex string, chainHex string, path string, index int32) string {
	privKeyBytes := DerivePrivateKeyForPath(
		HexDecode(mprivHex),
		HexDecode(chainHex),
		fmt.Sprintf("%v/%v", path, index),
	)
	return HexEncode(privKeyBytes)
}

func ComputeAddressForPrivKey(privKey string) string {
	pubKeyBytes := PubKeyBytesFromPrivKeyBytes(HexDecode(privKey), true)
	return AddrFromPubKeyBytes(pubKeyBytes)
}

func SignMessage(privKey string, message string, compress bool) string {
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
	sigbytes, err := btcec.SignCompact(btcec.S256(), ecdsaPrivKey, crypto.Sha256(crypto.Sha256(bytes)), compress)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(sigbytes)
}

// returns MPK, Chain, and master secret in hex.
func ComputeMastersFromSeed(seed string) (string, string, string, string) {
	secret, chain := I64([]byte("Bitcoin seed"), []byte(seed))
	pubKeyBytes := PubKeyBytesFromPrivKeyBytes(secret, true)
	return HexEncode(pubKeyBytes), HexEncode(secret), HexEncode(chain), HexEncode(secret)
}

func ComputeWIF(privKey string, compress bool) string {
	return WIFFromPrivKeyBytes(HexDecode(privKey), compress)
}

func ComputeTxId(rawTxHex string) string {
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

func DerivePrivateKeyForPath(privKeyBytes []byte, chain []byte, path string) []byte {
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
		data, chain = DerivePrivateKey(data, chain, uint32(i), prime)
		//printKeyInfo(data, nil, chain)
	}
	return data
}

func DerivePublicKeyForPath(pubKeyBytes []byte, chain []byte, path string) []byte {
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
		data, chain = DerivePublicKey(data, chain, uint32(i))
		//printKeyInfo(nil, data, chain)
	}
	return data
}

func DerivePrivateKey(privKeyBytes []byte, chain []byte, i uint32, prime bool) ([]byte, []byte) {
	var data []byte
	if prime {
		i = i | 0x80000000
		data = append([]byte{byte(0)}, privKeyBytes...)
	} else {
		public := PubKeyBytesFromPrivKeyBytes(privKeyBytes, true)
		data = public
	}
	data = append(data, uint32ToBytes(i)...)
	data2, chain2 := I64(chain, data)
	x := addScalars(privKeyBytes, data2)
	return x, chain2
}

func DerivePublicKey(pubKeyBytes []byte, chain []byte, i uint32) ([]byte, []byte) {
	data := []byte{}
	data = append(data, pubKeyBytes...)
	data = append(data, uint32ToBytes(i)...)
	data2, chain2 := I64(chain, data)
	data2p := PubKeyBytesFromPrivKeyBytes(data2, true)
	return addPoints(pubKeyBytes, data2p), chain2
}

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

func HexEncode(b []byte) string {
	return hex.EncodeToString(b)
}

func HexDecode(str string) []byte {
	b, _ := hex.DecodeString(str)
	return b
}

func I64(key []byte, data []byte) ([]byte, []byte) {
	mac := hmac.New(sha512.New, key)
	mac.Write(data)
	I := mac.Sum(nil)
	return I[:32], I[32:]
}

// This returns a Bitcoin-like address.
func AddrFromPubKeyBytes(pubKeyBytes []byte) string {
	prefix := byte(0x00) // TODO Make const or configurable
	h160 := CalcHash160(pubKeyBytes)
	h160 = append([]byte{prefix}, h160...)
	checksum := CalcHash256(h160)
	b := append(h160, checksum[:4]...)
	return base58.Encode(b)
}

func AddrBytesFromPubKeyBytes(pubKeyBytes []byte) (addrBytes []byte, checksum []byte) {
	prefix := byte(0x00) // TODO Make const or configurable
	h160 := CalcHash160(pubKeyBytes)
	_h160 := append([]byte{prefix}, h160...)
	checksum = CalcHash256(_h160)[:4]
	return h160, checksum
}

func WIFFromPrivKeyBytes(privKeyBytes []byte, compress bool) string {
	prefix := byte(0x80) // TODO Make const or configurable
	bytes := append([]byte{prefix}, privKeyBytes...)
	if compress {
		bytes = append(bytes, byte(1))
	}
	checksum := CalcHash256(bytes)
	bytes = append(bytes, checksum[:4]...)
	return base58.Encode(bytes)
}

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

// Calculate the hash of hasher over buf.
func CalcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// calculate hash160 which is ripemd160(sha256(data))
func CalcHash160(buf []byte) []byte {
	return CalcHash(CalcHash(buf, sha256.New()), ripemd160.New())
}

// calculate hash256 which is sha256(sha256(data))
func CalcHash256(buf []byte) []byte {
	return CalcHash(CalcHash(buf, sha256.New()), sha256.New())
}

// calculate sha512(data)
func CalcSha512(buf []byte) []byte {
	return CalcHash(buf, sha512.New())
}

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
