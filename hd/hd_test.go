package crypto

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tyler-smith/go-bip32"

	"github.com/tendermint/go-crypto"
)

type addrData struct {
	Mnemonic string
	Seed     string
	Priv     string
	Pub      string
	Addr     string
}

// NOTE: atom fundraiser address
var hdPath string = "m/44'/118'/0'/0/0"
var hdToAddrTable []addrData

/*{
	{
		Mnemonic: "spawn essence sudden gown library fire chalk edge start museum glimpse sea",
		Priv:     "ab20a81c1b9002538e2269e1f1302d519901633d40408313211598899bc00fc6",
		Pub:      "03eb89fb1c4582eed592e09c31c4665d3956154ea66fd269933d3f036e879abfe6",
		Addr:     "f7d613738f0a665ec320306d14f5d62a850ff714",
	},
}*/

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

		seed := bip39.NewSeed(d.Mnemonic, "")

		fmt.Println(i, d.Mnemonic)

		priv, pub := tylerSmith(seed)
		//priv, pub := btcsuite(seed)

		fmt.Printf("\t%X %X\n", seedB, seed)
		fmt.Printf("\t%X %X\n", privB, priv)
		fmt.Printf("\t%X %X\n", pubB, pub)
		assert.Equal(t, priv, privB, "Expected priv keys to match")
		assert.Equal(t, pub, pubB, "Expected pub keys to match")

		var pubT crypto.PubKeySecp256k1
		copy(pubT[:], pub)
		addr := pubT.Address()
		assert.Equal(t, addr, addrB, "Expected addresses to match")

		/*		if i%10 == 0 {
				fmt.Printf("ADDR %d: %s %X %X\n", i, d.Mnemonic, addr, addrB)
			}*/
	}
}

func ifExit(err error, n int) {
	if err != nil {
		fmt.Println(n, err)
		os.Exit(1)
	}
}

func btcsuite(seed []byte) ([]byte, []byte) {
	masterKey, _ := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
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
	return priv, pub
}

// return priv and pub
func tylerSmith(seed []byte) ([]byte, []byte) {
	masterKey, _ := bip32.NewMasterKey(seed)
	fh := bip32.FirstHardenedChild
	k, _ := masterKey.NewChildKey(fh + 44)
	k, _ = k.NewChildKey(fh + 118)
	k, _ = k.NewChildKey(fh + 0)
	k, _ = k.NewChildKey(0)
	k, _ = k.NewChildKey(0)

	priv := k.Key
	pub := k.PublicKey().Key
	return priv, pub
}
