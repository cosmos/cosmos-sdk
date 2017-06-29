package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
)

//commands
var (
	InitCmd = &cobra.Command{
		Use:   "init [address]",
		Short: "Initialize a basecoin blockchain",
		RunE:  initCmd,
	}
)

//flags
var (
	chainIDFlag string
)

func init() {
	flags := []Flag2Register{
		{&chainIDFlag, "chain-id", "test_chain_id", "Chain ID"},
	}
	RegisterFlags(InitCmd, flags)
}

// returns 1 iff it set a file, otherwise 0 (so we can add them)
func setupFile(path, data string, perm os.FileMode) (int, error) {
	_, err := os.Stat(path)
	if !os.IsNotExist(err) { //note, os.IsExist(err) != !os.IsNotExist(err)
		return 0, nil
	}
	err = ioutil.WriteFile(path, []byte(data), perm)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func initCmd(cmd *cobra.Command, args []string) error {
	// this will ensure that config.toml is there if not yet created, and create dir
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return err
	}

	if len(args) != 1 {
		return fmt.Errorf("`init` takes one argument, a basecoin account address. Generate one using `basecli keys new mykey`")
	}
	userAddr := args[0]

	// initalize basecoin
	genesisFile := cfg.GenesisFile()
	privValFile := cfg.PrivValidatorFile()
	keyFile := path.Join(cfg.RootDir, "key.json")

	mod1, err := setupFile(genesisFile, GetGenesisJSON(chainIDFlag, userAddr), 0644)
	if err != nil {
		return err
	}
	mod2, err := setupFile(privValFile, PrivValJSON, 0400)
	if err != nil {
		return err
	}
	mod3, err := setupFile(keyFile, KeyJSON, 0400)
	if err != nil {
		return err
	}

	if (mod1 + mod2 + mod3) > 0 {
		msg := fmt.Sprintf("Initialized %s", cmd.Root().Name())
		logger.Info(msg, "genesis", genesisFile, "priv_validator", privValFile)
	} else {
		logger.Info("Already initialized", "priv_validator", privValFile)
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
  "priv_key": {
    "type": "ed25519",
    "data": "D07ABE82A8B15559A983B2DB5D4842B2B6E4D6AF58B080005662F424F17D68C17B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
  },
  "pub_key": {
    "type": "ed25519",
    "data": "7B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
  }
}`

// GetGenesisJSON returns a new tendermint genesis with Basecoin app_options
// that grant a large amount of "mycoin" to a single address
// TODO: A better UX for generating genesis files
func GetGenesisJSON(chainID, addr string) string {
	return fmt.Sprintf(`{
  "app_hash": "",
  "chain_id": "%s",
  "genesis_time": "0001-01-01T00:00:00.000Z",
  "validators": [
    {
      "amount": 10,
      "name": "",
      "pub_key": {
        "type": "ed25519",
        "data": "7B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
      }
    }
  ],
  "app_options": {
    "accounts": [{
      "address": "%s",
      "coins": [
        {
          "denom": "mycoin",
          "amount": 9007199254740992
        }
      ]
    }]
  }
}`, chainID, addr)
}

// TODO: remove this once not needed for relay
var KeyJSON = `{
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
