package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var logger *zerolog.Logger

var rootCmd = &cobra.Command{
	Use:   "cosmovisor",
	Short: "A process manager for Cosmos SDK application binaries.",
	Long:  GetHelpText(),
}

// Execute the CLI application.
func Execute(log *zerolog.Logger) {
	logger = log

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
