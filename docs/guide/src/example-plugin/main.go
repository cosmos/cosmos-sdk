package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/tmlibs/cli"
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

	cmd := cli.PrepareMainCmd(RootCmd, "BC", os.ExpandEnv("$HOME/.basecoin-example-plugin"))
	cmd.Execute()
}
