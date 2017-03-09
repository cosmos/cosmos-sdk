package main

import (
	"fmt"
	"os"

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
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
