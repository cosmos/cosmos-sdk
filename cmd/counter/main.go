package main

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/cmd/commands"
)

func main() {

	var RootCmd = &cobra.Command{
		Use:   "counter",
		Short: "demo plugin for basecoin",
	}

	RootCmd.AddCommand(
		commands.StartCmd,
		commands.TxCmd,
		commands.QueryCmd,
		commands.KeyCmd,
		commands.VerifyCmd,
		commands.BlockCmd,
		commands.AccountCmd,
		commands.QuickVersionCmd("0.1.0"),
	)

	commands.ExecuteWithDebug(RootCmd)
}
