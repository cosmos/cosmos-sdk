package main

import (
	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/basecoin/plugins/counter"
	"github.com/tendermint/basecoin/types"
)

func init() {
	commands.RegisterStartPlugin("counter", func() types.Plugin { return counter.New() })
}
