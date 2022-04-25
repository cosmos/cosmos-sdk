package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor/cmd"
	cverrors "github.com/cosmos/cosmos-sdk/cosmovisor/errors"
	"github.com/cosmos/cosmos-sdk/cosmovisor/logging"
)

func main() {
	logger := logging.NewLogger()
	app := cmd.NewCosmovisor(logger)
	if err := app.RunCosmovisorCommand(os.Args[1:]); err != nil {
		cverrors.LogErrors(logger, "", err)
		os.Exit(1)
	}
}
