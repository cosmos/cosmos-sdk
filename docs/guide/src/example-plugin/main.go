package main

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/cmd/commands"
)

func main() {

	//Initialize example-plugin root command
	var RootCmd = &cobra.Command{
		Use:   "example-plugin",
		Short: "example-plugin usage description",
	}

	//Add the default basecoin commands to the root command
	RootCmd.AddCommand(
		commands.InitCmd,
		commands.StartCmd,
		commands.TxCmd,
		commands.QueryCmd,
		commands.KeyCmd,
		commands.VerifyCmd,
		commands.BlockCmd,
		commands.AccountCmd,
		commands.UnsafeResetAllCmd,
	)

	//Run the root command
	commands.ExecuteWithDebug(RootCmd)
}
