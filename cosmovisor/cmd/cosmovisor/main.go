package main

import (
	"context"
	"os"

	"cosmossdk.io/cosmovisor"
	cverrors "gcosmossdk.io/cosmovisor/errors"
)

func main() {
	logger := cosmovisor.NewLogger()
	ctx := context.WithValue(context.Background(), cosmovisor.LoggerKey, logger)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		cverrors.LogErrors(logger, "", err)
		os.Exit(1)
	}
}
