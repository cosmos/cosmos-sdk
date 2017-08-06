package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"

	"github.com/tendermint/basecoin/cmd/basecoin/commands"
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
	return commands.CreateGenesisValidatorFiles(cfg, genesis, cmd.Root().Name())
}

// TODO: better, auto-generate validator...
func getGenesisJSON(chainID string) string {
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
  ]
}`, chainID)
}
