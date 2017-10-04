package hd

import (
	"bytes"
	//"crypto/hmac"
	//"crypto/sha512"
	//"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"

	//"github.com/btcsuite/btcd/chaincfg"
	//"github.com/btcsuite/btcutil/hdkeychain"
	//"github.com/mndrix/btcutil"
	//"github.com/tyler-smith/go-bip32"

	"github.com/tendermint/go-crypto"
)

type addrData struct {
	Mnemonic string
	Master   string
	Seed     string
	Priv     string
	Pub      string
	Addr     string
}

// NOTE: atom fundraiser address
// var hdPath string = "m/44'/118'/0'/0/0"
var hdToAddrTable []addrData

func init() {

	b, err := ioutil.ReadFile("test.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = json.Unmarshal(b, &hdToAddrTable)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestHDToAddr(t *testing.T) {

	for i, d := range hdToAddrTable {
		privB, _ := hex.DecodeString(d.Priv)
		pubB, _ := hex.DecodeString(d.Pub)
		addrB, _ := hex.DecodeString(d.Addr)
		seedB, _ := hex.DecodeString(d.Seed)
		masterB, _ := hex.DecodeString(d.Master)

		seed := bip39.NewSeed(d.Mnemonic, "")

		fmt.Println("================================")
		fmt.Println("ROUND:", i, "MNEMONIC:", d.Mnemonic)

		// master, priv, pub := tylerSmith(seed)
		// master, priv, pub := btcsuite(seed)
		master, priv, pub := gocrypto(seed)

		fmt.Printf("\tNODEJS GOLANG\n")
		fmt.Printf("SEED \t%X %X\n", seedB, seed)
		fmt.Printf("MSTR \t%X %X\n", masterB, master)
		fmt.Printf("PRIV \t%X %X\n", privB, priv)
		fmt.Printf("PUB  \t%X %X\n", pubB, pub)
		_, _ = priv, privB

		assert.Equal(t, master, masterB, fmt.Sprintf("Expected masters to match for %d", i))
		assert.Equal(t, priv, privB, "Expected priv keys to match")
		assert.Equal(t, pub, pubB, fmt.Sprintf("Expected pub keys to match for %d", i))

		var pubT crypto.PubKeySecp256k1
		copy(pubT[:], pub)
		addr := pubT.Address()
		fmt.Printf("ADDR  \t%X %X\n", addrB, addr)
		assert.Equal(t, addr, addrB, fmt.Sprintf("Expected addresses to match %d", i))

	}
}

func TestReverseBytes(t *testing.T) {
	tests := [...]struct {
		v    []byte
		want []byte
	}{
		{[]byte(""), []byte("")},
		{nil, nil},
		{[]byte("Tendermint"), []byte("tnimredneT")},
		{[]byte("T"), []byte("T")},
		{[]byte("Te"), []byte("eT")},
	}

	for i, tt := range tests {
		got := ReverseBytes(tt.v)
		if !bytes.Equal(got, tt.want) {
			t.Errorf("#%d:\ngot= (%x)\nwant=(%x)", i, got, tt.want)
		}
	}
}

/*
func ifExit(err error, n int) {
	if err != nil {
		fmt.Println(n, err)
		os.Exit(1)
	}
}
*/

func gocrypto(seed []byte) ([]byte, []byte, []byte) {

	_, priv, ch, _ := ComputeMastersFromSeed(string(seed))

	privBytes := DerivePrivateKeyForPath(
		HexDecode(priv),
		HexDecode(ch),
		"44'/118'/0'/0/0",
	)

	pubBytes := PubKeyBytesFromPrivKeyBytes(privBytes, true)

	return HexDecode(priv), privBytes, pubBytes
}

/*
func btcsuite(seed []byte) ([]byte, []byte, []byte) {
	fmt.Println("HD")
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		hmac := hmac.New(sha512.New, []byte("Bitcoin seed"))
		hmac.Write([]byte(seed))
		intermediary := hmac.Sum(nil)

		curve := btcutil.Secp256k1()
		curveParams := curve.Params()

		// Split it into our key and chain code
		keyBytes := intermediary[:32]
		fmt.Printf("\t%X\n", keyBytes)
		fmt.Printf("\t%X\n", curveParams.N.Bytes())
		keyInt, _ := binary.ReadVarint(bytes.NewBuffer(keyBytes))
		fmt.Printf("\t%d\n", keyInt)
	}
	fh := hdkeychain.HardenedKeyStart
	k, err := masterKey.Child(uint32(fh + 44))
	ifExit(err, 44)
	k, err = k.Child(uint32(fh + 118))
	ifExit(err, 118)
	k, err = k.Child(uint32(fh + 0))
	ifExit(err, 1)
	k, err = k.Child(uint32(0))
	ifExit(err, 2)
	k, err = k.Child(uint32(0))
	ifExit(err, 3)
	ecpriv, err := k.ECPrivKey()
	ifExit(err, 10)
	ecpub, err := k.ECPubKey()
	ifExit(err, 11)

	priv := ecpriv.Serialize()
	pub := ecpub.SerializeCompressed()
	mkey, _ := masterKey.ECPrivKey()
	return mkey.Serialize(), priv, pub
}

// return priv and pub
func tylerSmith(seed []byte) ([]byte, []byte, []byte) {
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		hmac := hmac.New(sha512.New, []byte("Bitcoin seed"))
		hmac.Write([]byte(seed))
		intermediary := hmac.Sum(nil)

		curve := btcutil.Secp256k1()
		curveParams := curve.Params()

		// Split it into our key and chain code
		keyBytes := intermediary[:32]
		fmt.Printf("\t%X\n", keyBytes)
		fmt.Printf("\t%X\n", curveParams.N.Bytes())
		keyInt, _ := binary.ReadVarint(bytes.NewBuffer(keyBytes))
		fmt.Printf("\t%d\n", keyInt)

	}
	ifExit(err, 0)
	fh := bip32.FirstHardenedChild
	k, err := masterKey.NewChildKey(fh + 44)
	ifExit(err, 44)
	k, err = k.NewChildKey(fh + 118)
	ifExit(err, 118)
	k, err = k.NewChildKey(fh + 0)
	ifExit(err, 1)
	k, err = k.NewChildKey(0)
	ifExit(err, 2)
	k, err = k.NewChildKey(0)
	ifExit(err, 3)

	priv := k.Key
	pub := k.PublicKey().Key
	return masterKey.Key, priv, pub
}
*/

// Benchmarks
var revBytesCases = [][]byte{
	nil,
	[]byte(""),

	[]byte("12"),

	// 16byte case
	[]byte("abcdefghijklmnop"),

	// 32byte case
	[]byte("abcdefghijklmnopqrstuvwxyz123456"),

	// 64byte case
	[]byte("abcdefghijklmnopqrstuvwxyz123456abcdefghijklmnopqrstuvwxyz123456"),
}

func BenchmarkReverseBytes(b *testing.B) {
	var sink []byte
	for i := 0; i < b.N; i++ {
		for _, tt := range revBytesCases {
			sink = ReverseBytes(tt)
		}
	}
	b.ReportAllocs()

	// sink is necessary to ensure if the compiler tries
	// to smart, that it won't optimize away the benchmarks.
	if sink != nil {
		_ = sink
	}
}
