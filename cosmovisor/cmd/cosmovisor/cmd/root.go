package cmd

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
	cverrors "github.com/cosmos/cosmos-sdk/cosmovisor/errors"
)

var rootCmd = &cobra.Command{
	Use:   "cosmovisor",
	Short: "A process manager for Cosmos SDK application binaries.",
	Long:  GetHelpText(),
}

// Execute the CLI application.
func Execute(logger *zerolog.Logger) {
	ctx := context.WithValue(context.Background(), cosmovisor.LoggerKey, logger)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		cverrors.LogErrors(logger, "", err)
		os.Exit(1)
	}
}
