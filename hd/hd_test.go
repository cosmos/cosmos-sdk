// nolint: goimports
package hd

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"

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
	if sink != nil { // nolint: megacheck
	}
}
