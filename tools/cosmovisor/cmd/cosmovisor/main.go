package main

import (
	"context"
	"os"

	"cosmossdk.io/log"
	cverrors "cosmossdk.io/tools/cosmovisor/errors"
)

func main() {
	logger := log.NewZeroLogger(log.ModuleKey, "cosmovisor")
	ctx := context.WithValue(context.Background(), log.ContextKey, logger)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		cverrors.LogErrors(logger, "", err)
		os.Exit(1)
	}
}
