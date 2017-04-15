package main

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/cmd/commands"
)

func main() {
	var RootCmd = &cobra.Command{
		Use:   "basecoin",
		Short: "A cryptocurrency framework in Golang based on Tendermint-Core",
	}

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
		commands.VersionCmd,
	)

	commands.ExecuteWithDebug(RootCmd)
}
