package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
	"github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor/cmd"
	cverrors "github.com/cosmos/cosmos-sdk/cosmovisor/errors"
)

func main() {
	cosmovisor.SetupLogging()
	if err := cmd.RunCosmovisorCommand(os.Args[1:]); err != nil {
		cverrors.LogErrors(cosmovisor.Logger, "", err)
		os.Exit(1)
	}
}
