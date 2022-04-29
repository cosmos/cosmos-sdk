package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	cverrors "github.com/cosmos/cosmos-sdk/cosmovisor/errors"
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
		cverrors.LogErrors(logger, "", err)
		os.Exit(1)
	}
}
