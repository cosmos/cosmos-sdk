package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"

	cmn "github.com/tendermint/go-common"
)

//commands
var (
	InitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a basecoin blockchain",
		Run:   initCmd,
	}
)

//flags
//var chainIDFlag string

//func init() {
//
//	//register flags
//	flags := []Flag2Register{
//		{&chainIDFlag, "chain_id", "test_chain_id", "ID of the chain for replay protection"},
//	}
//	RegisterFlags(InitCmd, flags)
//
//}

func initCmd(cmd *cobra.Command, args []string) {
	rootDir := BasecoinRoot("")

	cmn.EnsureDir(rootDir, 0777)

	// initalize basecoin
	genesisFile := path.Join(rootDir, "genesis.json")
	privValFile := path.Join(rootDir, "priv_validator.json")
	key1File := path.Join(rootDir, "key.json")
	key2File := path.Join(rootDir, "key2.json")

	if _, err := os.Stat(privValFile); os.IsNotExist(err) {
		err := ioutil.WriteFile(genesisFile, []byte(genesisJSON), 0644)
		if err != nil {
			cmn.Exit(fmt.Sprintf("%+v\n", err))
		}

		err = ioutil.WriteFile(privValFile, []byte(privValJSON), 0400)
		if err != nil {
			cmn.Exit(fmt.Sprintf("%+v\n", err))
		}

		err = ioutil.WriteFile(key1File, []byte(key1JSON), 0400)
		if err != nil {
			cmn.Exit(fmt.Sprintf("%+v\n", err))
		}

		err = ioutil.WriteFile(key2File, []byte(key2JSON), 0400)
		if err != nil {
			cmn.Exit(fmt.Sprintf("%+v\n", err))
		}

		log.Notice("Initialized Basecoin", "genesis", genesisFile, "key", key1File)
	} else {
		log.Notice("Already initialized", "priv_validator", privValFile)
	}
}

const privValJSON = `{
	"address": "7A956FADD20D3A5B2375042B2959F8AB172A058F",
	"last_height": 0,
	"last_round": 0,
	"last_signature": null,
	"last_signbytes": "",
	"last_step": 0,
	"priv_key": [
		1,
		"D07ABE82A8B15559A983B2DB5D4842B2B6E4D6AF58B080005662F424F17D68C17B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
	],
	"pub_key": [
		1,
		"7B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
	]
}`

const genesisJSON = `{
  "app_hash": "",
  "chain_id": "test_chain_id",
  "genesis_time": "0001-01-01T00:00:00.000Z",
  "validators": [
    {
      "amount": 10,
      "name": "",
      "pub_key": [
	1,
	"7B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
      ]
    }
  ],
  "app_options": {
    "accounts": [{
      "pub_key": {
        "type": "ed25519",
        "data": "619D3678599971ED29C7529DDD4DA537B97129893598A17C82E3AC9A8BA95279"
      },
      "coins": [
        {
          "denom": "mycoin",
          "amount": 9007199254740992
        }
      ]
    }]
  }
}`

const key1JSON = `{
	"address": "1B1BE55F969F54064628A63B9559E7C21C925165",
	"priv_key": {
		"type": "ed25519",
		"data": "C70D6934B4F55F1B7BC33B56B9CA8A2061384AFC19E91E44B40C4BBA182953D1619D3678599971ED29C7529DDD4DA537B97129893598A17C82E3AC9A8BA95279"
	},
	"pub_key": {
		"type": "ed25519",
		"data": "619D3678599971ED29C7529DDD4DA537B97129893598A17C82E3AC9A8BA95279"
	}
}`

const key2JSON = `{
	"address": "1DA7C74F9C219229FD54CC9F7386D5A3839F0090",
	"priv_key": {
		"type": "ed25519",
		"data": "34BAE9E65CE8245FAD035A0E3EED9401BDE8785FFB3199ACCF8F5B5DDF7486A8352195DA90CB0B90C24295B90AEBA25A5A71BC61BAB2FE2387241D439698B7B8"
	},
	"pub_key": {
		"type": "ed25519",
		"data": "352195DA90CB0B90C24295B90AEBA25A5A71BC61BAB2FE2387241D439698B7B8"
	}
}`
