package main

import (
	"os"

	"cosmossdk.io/tools/stateviewer"
	"github.com/spf13/cobra"
)

func main() {
	if err := Commands().Execute(); err != nil {
		os.Exit(1)
	}
}

// Commands contains all the state-viewer commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stateviewer",
		Short: "Utilities viewing Cosmos SDK application state",
	}

	cmd.AddCommand(
		stateviewer.ViewCommand(),
		stateviewer.RawViewCommand(),
		stateviewer.StatsCmd(),
		stateviewer.VersionCmd(),
	)

	return cmd
}
