package main

import (
	"os"

	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/cmd/basecoin/commands"
	"github.com/tendermint/tmlibs/cli"
)

func main() {
	rt := commands.RootCmd

	commands.Handler = app.DefaultHandler()

	rt.AddCommand(
		commands.InitCmd,
		commands.StartCmd,
		// commands.RelayCmd,
		commands.UnsafeResetAllCmd,
		commands.VersionCmd,
	)

	cmd := cli.PrepareMainCmd(rt, "BC", os.ExpandEnv("$HOME/.basecoin"))
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
