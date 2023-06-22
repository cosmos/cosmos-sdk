package main

import (
	"context"
	"os"
)

func main() {
	// error logger used only during configuration phase
	cfg, _ := getConfigForInitCmd()
	logger := cfg.Logger(os.Stderr)

	if err := NewRootCmd().ExecuteContext(context.Background()); err != nil {
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
