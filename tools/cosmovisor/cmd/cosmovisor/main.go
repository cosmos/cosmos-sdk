package main

import (
	"context"
	"os"
	"time"

	"cosmossdk.io/log"

	cverrors "cosmossdk.io/tools/cosmovisor/errors"
)

func main() {
	// error logger used only during configuration phase
	logger := log.NewLogger(
		os.Stderr,
		log.ColorOption(false),
		log.TimeFormatOption(time.Kitchen),
	).With(log.ModuleKey, "cosmovisor")

	if err := NewRootCmd().ExecuteContext(context.Background()); err != nil {
		cverrors.LogErrors(logger, "", err)
		os.Exit(1)
	}
}
