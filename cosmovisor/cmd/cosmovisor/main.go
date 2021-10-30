package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
	"github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor/cmd"
)

func main() {
	cosmovisor.SetupLogging()
	if err := cmd.RunCosmovisorCommand(os.Args[1:]); err != nil {
		cosmovisor.Logger.Error().Err(err).Msg("")
		os.Exit(1)
	}
}
