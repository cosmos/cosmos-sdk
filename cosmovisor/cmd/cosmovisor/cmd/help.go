package cmd

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(helpCmd)
}

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Print the help command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(GetHelpText())
	},
}

// GetHelpText creates the help text multi-line string.
func GetHelpText() string {
	return fmt.Sprintf(`Cosmosvisor - A process manager for Cosmos SDK application binaries.

Cosmovisor is a wrapper for a Cosmos SDK based App (set using the required %s env variable).
It starts the App by passing all provided arguments and monitors the %s/data/upgrade-info.json
file to perform an update. The upgrade-info.json file is created by the App x/upgrade module
when the blockchain height reaches an approved upgrade proposal. The file includes data from
the proposal. Cosmovisor interprets that data to perform an update: switch a current binary
and restart the App.

Configuration of Cosmovisor is done through environment variables, which are
documented in: https://github.com/cosmos/cosmos-sdk/tree/main/cosmovisor/README.md

To get help for the configured binary:
  cosmovisor run help

Available Commands:
  help     This help message
  run      Runs app passing all subsequent parameters
  version  Prints version of cosmovisor and the associated app.
`, cosmovisor.EnvName, cosmovisor.EnvHome)
}
