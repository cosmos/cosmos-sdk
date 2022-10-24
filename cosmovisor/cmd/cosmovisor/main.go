package main

import (
	"context"
	"os"

	"github.com/pointnetwork/cosmos-point-sdk/cosmovisor"
	cverrors "github.com/pointnetwork/cosmos-point-sdk/cosmovisor/errors"
)

func main() {
	logger := cosmovisor.NewLogger()
	ctx := context.WithValue(context.Background(), cosmovisor.LoggerKey, logger)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		cverrors.LogErrors(logger, "", err)
		os.Exit(1)
	}
}
