package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	basecmd "github.com/cosmos/cosmos-sdk/server/commands"
	"github.com/cosmos/gaia/version"
)

// GaiaCmd is the entry point for this binary
var (
	GaiaCmd = &cobra.Command{
		Use:   "gaia",
		Short: "The Cosmos Network delegation-game test",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	lineBreak = &cobra.Command{Run: func(*cobra.Command, []string) {}}
)

func main() {
	// disable sorting
	cobra.EnableCommandSorting = false

	// add commands
	prepareNodeCommands()
	prepareRestServerCommands()
	prepareClientCommands()

	GaiaCmd.AddCommand(
		nodeCmd,
		restServerCmd,
		clientCmd,

		lineBreak,
		version.VersionCmd,
		//auto.AutoCompleteCmd,
	)

	// prepare and add flags
	basecmd.SetUpRoot(GaiaCmd)
	executor := cli.PrepareMainCmd(GaiaCmd, "GA", os.ExpandEnv("$HOME/.cosmos-gaia-cli"))
	executor.Execute()
}
