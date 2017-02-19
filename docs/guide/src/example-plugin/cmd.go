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

	// Create a transaction using the flag.
	// The tx passes on custom information to the plugin
	exampleTx := ExamplePluginTx{exampleFlag}

	// The tx is passed to the plugin in the form of
	// a byte array. This is achieved by serializing the object using go-wire.
	// Once received in the plugin, these exampleTxBytes are decoded back
	// into the original ExamplePluginTx struct
	exampleTxBytes := wire.BinaryBytes(exampleTx)

	// Send the transaction and return any errors.
	// Here exampleTxBytes is packaged in the `tx.Data` field of an AppTx,
	// and passed on to the plugin through the following sequence:
	//  - passed as `data` to `commands.AppTx` (cmd/commands/tx.go)
	//  - set as the `tx.Data` field of an AppTx, which is then passed to commands.broadcastTx (cmd/commands/tx.go)
	//  - the tx is broadcast to Tendermint, which runs it through app.CheckTx (app/app.go)
	//  - after passing CheckTx, it will eventually be included in a block and run through app.DeliverTx (app/app.go)
	//  - DeliverTx receives txBytes, which is the serialization of the full AppTx (app/app.go)
	//  - Once deserialized, the tx is passed to `state.ExecTx` (state/execution.go)
	//  - If the tx passes various checks, the `tx.Data` is forwarded as `txBytes` to `plugin.RunTx` (docs/guide/src/example-plugin/plugin.go)
	//  - Finally, it deserialized back to the ExamplePluginTx
	return commands.AppTx(c, "example-plugin", exampleTxBytes)
}
