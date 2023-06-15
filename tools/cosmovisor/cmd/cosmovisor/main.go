package main

import (
	"context"
	"os"

	"cosmossdk.io/log"
)

func main() {
	logger := log.NewLogger(os.Stdout).With(log.ModuleKey, "cosmovisor")
	ctx := context.WithValue(context.Background(), log.ContextKey, logger)

	if err := NewRootCmd().ExecuteContext(ctx); err != nil {
		if errMulti, ok := err.(interface{ Unwrap() []error }); ok {
			err := errMulti.Unwrap()
			for _, e := range err {
				logger.Error("", "error", e)
			}
		} else {
			logger.Error("", "error", err)
		}

		os.Exit(1)
	}
}
