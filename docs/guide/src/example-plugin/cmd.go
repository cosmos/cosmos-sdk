package main

import (
	wire "github.com/tendermint/go-wire"
	"github.com/urfave/cli"

	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/basecoin/types"
)

func init() {
	commands.RegisterTxSubcommand(ExamplePluginTxCmd)
	commands.RegisterStartPlugin("example-plugin", func() types.Plugin { return NewExamplePlugin() })
}

var (
	ExampleFlag = cli.BoolFlag{
		Name:  "valid",
		Usage: "Set this to make the transaction valid",
	}

	ExamplePluginTxCmd = cli.Command{
		Name:  "example",
		Usage: "Create, sign, and broadcast a transaction to the example plugin",
		Action: func(c *cli.Context) error {
			return cmdExamplePluginTx(c)
		},
		Flags: append(commands.TxFlags, ExampleFlag),
	}
)

func cmdExamplePluginTx(c *cli.Context) error {
	exampleFlag := c.Bool("valid")
	exampleTx := ExamplePluginTx{exampleFlag}
	return commands.AppTx(c, "example-plugin", wire.BinaryBytes(exampleTx))
}
