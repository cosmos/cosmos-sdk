package main

import (
	"github.com/cosmos/cosmos-sdk/cosmovisor"
	"github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor/cmd"
)

var logger = cosmovisor.NewLogger()

func main() {
	cmd.Execute(logger)
}
