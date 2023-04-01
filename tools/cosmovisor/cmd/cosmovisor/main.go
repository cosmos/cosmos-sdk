package main

import (
	"context"
	"os"

	"cosmossdk.io/log"
	cverrors "cosmossdk.io/tools/cosmovisor/errors"
)

func main() {
	logger := log.NewLogger(os.Stdout).With(log.ModuleKey, "cosmovisor")
	ctx := context.WithValue(context.Background(), log.ContextKey, logger)

	if err := NewRootCmd().ExecuteContext(ctx); err != nil {
		cverrors.LogErrors(logger, "", err)
		os.Exit(1)
	}
}
