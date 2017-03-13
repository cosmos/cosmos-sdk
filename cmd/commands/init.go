package commands

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/urfave/cli"

	cmn "github.com/tendermint/go-common"
	tmcfg "github.com/tendermint/tendermint/config/tendermint"
	types "github.com/tendermint/tendermint/types"
)

var InitCmd = cli.Command{
	Name:      "init",
	Usage:     "Initialize a basecoin blockchain",
	ArgsUsage: "",
	Action: func(c *cli.Context) error {
		return cmdInit(c)
	},
	Flags: []cli.Flag{
		ChainIDFlag,
	},
}

func cmdInit(c *cli.Context) error {
	basecoinDir := BasecoinRoot("")
	tmDir := path.Join(basecoinDir, "tendermint")

	// initalize tendermint
	tmConfig := tmcfg.GetConfig(tmDir)

	privValFile := tmConfig.GetString("priv_validator_file")
	if _, err := os.Stat(privValFile); os.IsNotExist(err) {
		privValidator := types.GenPrivValidator()
		privValidator.SetFile(privValFile)
		privValidator.Save()

		genFile := tmConfig.GetString("genesis_file")

		if _, err := os.Stat(genFile); os.IsNotExist(err) {
			genDoc := types.GenesisDoc{
				ChainID: cmn.Fmt("test-chain-%v", cmn.RandStr(6)),
			}
			genDoc.Validators = []types.GenesisValidator{types.GenesisValidator{
				PubKey: privValidator.PubKey,
				Amount: 10,
			}}

			genDoc.SaveAs(genFile)
		}
		log.Notice("Initialized Tendermint", "genesis", tmConfig.GetString("genesis_file"), "priv_validator", tmConfig.GetString("priv_validator_file"))
	} else {
		log.Notice("Already initialized Tendermint", "priv_validator", tmConfig.GetString("priv_validator_file"))
	}

	// initalize basecoin
	genesisFile := path.Join(basecoinDir, "genesis.json")
	key1File := path.Join(basecoinDir, "key.json")
	key2File := path.Join(basecoinDir, "key2.json")

	if err := ioutil.WriteFile(genesisFile, []byte(genesisJSON), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(key1File, []byte(key1JSON), 0400); err != nil {
		return err
	}
	if err := ioutil.WriteFile(key2File, []byte(key2JSON), 0400); err != nil {
		return err
	}

	log.Notice("Initialized Basecoin", "genesis", genesisFile, "key", key1File)

	return nil
}

const genesisJSON = `[
  "base/chainID", "test_chain_id",
  "base/account", {
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
  }
]`

const key1JSON = `{
	"address": "1B1BE55F969F54064628A63B9559E7C21C925165",
	"priv_key": [
		1,
		"C70D6934B4F55F1B7BC33B56B9CA8A2061384AFC19E91E44B40C4BBA182953D1619D3678599971ED29C7529DDD4DA537B97129893598A17C82E3AC9A8BA95279"
	],
	"pub_key": [
		1,
		"619D3678599971ED29C7529DDD4DA537B97129893598A17C82E3AC9A8BA95279"
	]
}`

const key2JSON = `{
	"address": "1DA7C74F9C219229FD54CC9F7386D5A3839F0090",
	"priv_key": [
		1,
		"34BAE9E65CE8245FAD035A0E3EED9401BDE8785FFB3199ACCF8F5B5DDF7486A8352195DA90CB0B90C24295B90AEBA25A5A71BC61BAB2FE2387241D439698B7B8"
	],
	"pub_key": [
		1,
		"352195DA90CB0B90C24295B90AEBA25A5A71BC61BAB2FE2387241D439698B7B8"
	]
}`
