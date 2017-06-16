package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/tmlibs/cli"
)

func main() {
	var RootCmd = &cobra.Command{
		Use:   "counter",
		Short: "demo plugin for basecoin",
	}

	RootCmd.AddCommand(
		commands.InitCmd,
		commands.StartCmd,
		commands.TxCmd,
		commands.UnsafeResetAllCmd,
		commands.VersionCmd,
	)

	cmd := cli.PrepareMainCmd(RootCmd, "BC", os.ExpandEnv("$HOME/.basecoin"))
	cmd.Execute()
}
