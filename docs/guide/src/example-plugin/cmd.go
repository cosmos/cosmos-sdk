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

	// Create a transaction object with flag results
	// This object is responsible for passing on custom plugin information
	exampleTx := ExamplePluginTx{exampleFlag}

	// The custom plugin object is passed to the plugin in the form of
	// a byte array. This is achieved serializing the object using go-wire.
	// Once received in the plugin, these exampleTxBytes are decoded back
	// into the original object struct ExamplePluginTx
	exampleTxBytes := wire.BinaryBytes(exampleTx)

	// Send the transaction and return any errors.
	// Here exampleTxBytes will be passed on to the plugin through the
	// following series of function calls:
	//  - commands.AppTx as data (cmd/commands/tx.go)
	//  - commands.broadcastTx as tx.Data (cmd/commands/tx.go)
	//    - after being broadcast the Tendermint transaction
	//      will be run through app.CheckTx, and if successful DeliverTx,
	//      let's assume app.CheckTx passes
	//  - app.DeliverTx serialized within txBytes as tx.Data (app/app.go)
	//  - state.ExecTx as tx.Data (state/execution.go)
	//  - plugin.RunTx as txBytes (docs/guide/src/example-plugin/plugin.go)
	return commands.AppTx(c, "example-plugin", exampleTxBytes)
}
