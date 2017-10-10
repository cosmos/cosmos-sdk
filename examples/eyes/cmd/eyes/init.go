package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"

	"github.com/cosmos/cosmos-sdk/server/commands"
)

// InitCmd - node initialization command
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize eyes abci server",
	RunE:  initCmd,
}

//nolint - flags
var (
	FlagChainID = "chain-id" //TODO group with other flags or remove? is this already a flag here?
)

func init() {
	InitCmd.Flags().String(FlagChainID, "eyes_test_id", "Chain ID")
}

func initCmd(cmd *cobra.Command, args []string) error {
	// this will ensure that config.toml is there if not yet created, and create dir
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return err
	}

	genesis := getGenesisJSON(viper.GetString(commands.FlagChainID))
	return commands.CreateGenesisValidatorFiles(cfg, genesis, PrivValJSON, cmd.Root().Name())
}

// PrivValJSON - validator private key file contents in json
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

// TODO: better, auto-generate validator...
func getGenesisJSON(chainID string) string {
	return fmt.Sprintf(`{
  "app_hash": "",
  "chain_id": "%s",
  "genesis_time": "0001-01-01T00:00:00.000Z",
  "validators": [
    {
      "power": 10,
      "name": "",
      "pub_key": {
        "type": "ed25519",
        "data": "7B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
      }
    }
  ]
}`, chainID)
}
