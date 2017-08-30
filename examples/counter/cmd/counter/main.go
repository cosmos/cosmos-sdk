package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	client "github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/cmd/basecoin/commands"
	"github.com/cosmos/cosmos-sdk/examples/counter/plugins/counter"
)

func main() {
	var RootCmd = &cobra.Command{
		Use:   "counter",
		Short: "demo plugin for basecoin",
	}

	// TODO: register the counter here
	commands.Handler = counter.NewHandler("mycoin")

	RootCmd.AddCommand(
		commands.InitCmd,
		commands.StartCmd,
		commands.UnsafeResetAllCmd,
		client.VersionCmd,
	)

	cmd := cli.PrepareMainCmd(RootCmd, "CT", os.ExpandEnv("$HOME/.counter"))
	cmd.Execute()
}
