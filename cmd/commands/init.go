package commands

import (
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
		RunE:  initCmd,
	}
)

// setupFile aborts on error... or should we return it??
// returns 1 iff it set a file, otherwise 0 (so we can add them)
func setupFile(path, data string, perm os.FileMode) (int, error) {
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		return 0, nil
	}
	err = ioutil.WriteFile(path, []byte(data), perm)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func initCmd(cmd *cobra.Command, args []string) error {
	rootDir := BasecoinRoot("")

	cmn.EnsureDir(rootDir, 0777)

	// initalize basecoin
	genesisFile := path.Join(rootDir, "genesis.json")
	privValFile := path.Join(rootDir, "priv_validator.json")
	key1File := path.Join(rootDir, "key.json")
	key2File := path.Join(rootDir, "key2.json")

	mod1, err := setupFile(genesisFile, GenesisJSON, 0644)
	if err != nil {
		return err
	}
	mod2, err := setupFile(privValFile, PrivValJSON, 0400)
	if err != nil {
		return err
	}
	mod3, err := setupFile(key1File, Key1JSON, 0400)
	if err != nil {
		return err
	}
	mod4, err := setupFile(key2File, Key2JSON, 0400)
	if err != nil {
		return err
	}

	if (mod1 + mod2 + mod3 + mod4) > 0 {
		log.Notice("Initialized Basecoin", "genesis", genesisFile, "key", key1File)
	} else {
		log.Notice("Already initialized", "priv_validator", privValFile)
	}

	return nil
}

var PrivValJSON = `{
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

var GenesisJSON = `{
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

var Key1JSON = `{
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

var Key2JSON = `{
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
