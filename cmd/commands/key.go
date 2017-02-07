package commands

import (
	"fmt"
	"io/ioutil"

	"github.com/urfave/cli"

	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
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
	keyJSON := wire.JSONBytesPretty(key)
	fmt.Println(string(keyJSON))
	return nil
}

//---------------------------------------------
// simple implementation of a key

type Key struct {
	Address []byte         `json:"address"`
	PubKey  crypto.PubKey  `json:"pub_key"`
	PrivKey crypto.PrivKey `json:"priv_key"`
}

// Implements Signer
func (k *Key) Sign(msg []byte) crypto.Signature {
	return k.PrivKey.Sign(msg)
}

// Generates a new validator with private key.
func genKey() *Key {
	privKey := crypto.GenPrivKeyEd25519()
	return &Key{
		Address: privKey.PubKey().Address(),
		PubKey:  privKey.PubKey(),
		PrivKey: privKey,
	}
}

func LoadKey(filePath string) *Key {
	keyJSONBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		cmn.Exit(err.Error())
	}
	key := wire.ReadJSON(&Key{}, keyJSONBytes, &err).(*Key)
	if err != nil {
		cmn.Exit(cmn.Fmt("Error reading PrivValidator from %v: %v\n", filePath, err))
	}
	return key
}
