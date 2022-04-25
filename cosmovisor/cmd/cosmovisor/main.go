package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor/cmd"
	cverrors "github.com/cosmos/cosmos-sdk/cosmovisor/errors"
	"github.com/cosmos/cosmos-sdk/cosmovisor/logging"
)

func main() {
	logger := logging.NewLogger()
	if err := cmd.RunCosmovisorCommand(logger, os.Args[1:]); err != nil {
		cverrors.LogErrors(logger, "", err)
		os.Exit(1)
	}
}
