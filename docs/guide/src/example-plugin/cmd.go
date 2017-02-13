package main

import (
	wire "github.com/tendermint/go-wire"
	"github.com/urfave/cli"

	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/basecoin/types"
)

//Called during CLI initialization
func init() {

	//Register a plugin specific CLI command as a subcommand of the tx command
	commands.RegisterTxSubcommand(ExamplePluginTxCmd)

	//Register the example with basecoin at start
	commands.RegisterStartPlugin("example-plugin", func() types.Plugin { return NewExamplePlugin() })
}

var (
	//CLI Flags
	ExampleFlag = cli.BoolFlag{
		Name:  "valid",
		Usage: "Set this to make the transaction valid",
	}

	//CLI Plugin Commands
	ExamplePluginTxCmd = cli.Command{
		Name:  "example",
		Usage: "Create, sign, and broadcast a transaction to the example plugin",
		Action: func(c *cli.Context) error {
			return cmdExamplePluginTx(c)
		},
		Flags: append(commands.TxFlags, ExampleFlag),
	}
)

//Send a transaction
func cmdExamplePluginTx(c *cli.Context) error {
	//Retrieve any flag results
	exampleFlag := c.Bool("valid")

	//Create a transaction object with flag results
	exampleTx := ExamplePluginTx{exampleFlag}

	//Encode transaction bytes
	exampleTxBytes := wire.BinaryBytes(exampleTx)

	//Send the transaction and return any errors
	return commands.AppTx(c, "example-plugin", exampleTxBytes)
}
