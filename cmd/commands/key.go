package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/urfave/cli"

	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
)

var (
	KeyCmd = cli.Command{
		Name:        "key",
		Usage:       "Manage keys",
		ArgsUsage:   "",
		Subcommands: []cli.Command{NewKeyCmd},
	}

	NewKeyCmd = cli.Command{
		Name:      "new",
		Usage:     "Create a new private key",
		ArgsUsage: "",
		Action: func(c *cli.Context) error {
			return cmdNewKey(c)
		},
	}
)

func cmdNewKey(c *cli.Context) error {
	key := genKey()
	keyJSON, err := json.MarshalIndent(key, "", "\t")
	if err != nil {
		return err
	}
	fmt.Println(keyJSON)
	return nil
}

//---------------------------------------------
// simple implementation of a key

type Address [20]byte

func (a Address) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%x"`, a[:])), nil
}

func (a *Address) UnmarshalJSON(addrHex []byte) error {
	addr, err := hex.DecodeString(strings.Trim(string(addrHex), `"`))
	if err != nil {
		return err
	}
	copy(a[:], addr)
	return nil
}

type Key struct {
	Address Address         `json:"address"`
	PubKey  crypto.PubKeyS  `json:"pub_key"`
	PrivKey crypto.PrivKeyS `json:"priv_key"`
}

// Implements Signer
func (k *Key) Sign(msg []byte) crypto.Signature {
	return k.PrivKey.Sign(msg)
}

// Generates a new validator with private key.
func genKey() *Key {
	privKey := crypto.GenPrivKeyEd25519()
	addrBytes := privKey.PubKey().Address()
	var addr Address
	copy(addr[:], addrBytes)
	return &Key{
		Address: addr,
		PubKey:  crypto.PubKeyS{privKey.PubKey()},
		PrivKey: crypto.PrivKeyS{privKey},
	}
}

func LoadKey(keyFile string) *Key {
	filePath := path.Join(BasecoinRoot(""), keyFile)
	keyJSONBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		cmn.Exit(err.Error())
	}
	key := new(Key)
	err = json.Unmarshal(keyJSONBytes, key)
	if err != nil {
		cmn.Exit(cmn.Fmt("Error reading key from %v: %v\n", filePath, err))
	}
	return key
}
