package main

import (
	"os"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/stateviewer"
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
