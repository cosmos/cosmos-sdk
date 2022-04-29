package main

import (
	"github.com/cosmos/cosmos-sdk/cosmovisor"
	"github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor/cmd"
)

func main() {
	cmd.Execute(cosmovisor.NewLogger())
}
