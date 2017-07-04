package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/cmd/basecoin/commands"
	"github.com/tendermint/basecoin/docs/guide/counter/plugins/counter"
	"github.com/tendermint/basecoin/types"
)

func main() {
	var RootCmd = &cobra.Command{
		Use:   "counter",
		Short: "demo plugin for basecoin",
	}

	// TODO: register the counter here
	commands.Handler = app.DefaultHandler()

	RootCmd.AddCommand(
		commands.InitCmd,
		commands.StartCmd,
		commands.UnsafeResetAllCmd,
		commands.VersionCmd,
	)

	commands.RegisterStartPlugin("counter", func() types.Plugin { return counter.New() })
	cmd := cli.PrepareMainCmd(RootCmd, "CT", os.ExpandEnv("$HOME/.counter"))
	cmd.Execute()
}
