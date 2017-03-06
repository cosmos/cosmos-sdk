# Basecoin Example Plugin

In the [previous tutorial](basecoin-basics.md),
we saw how to start a Basecoin blockchain and use the CLI to send transactions.
Here, we will demonstrate how to extend the blockchain and CLI to support a simple plugin.

## Overview

Creating a new plugin and CLI to support it requires a little bit of boilerplate, but not much.
For convenience, we've implemented an extremely simple example plugin that can be easily modified.
The example is under `docs/guide/src/example-plugin`.
To build your own plugin, copy this folder to a new location and start modifying it there.

Let's take a look at the files in `docs/guide/src/example-plugin`:

```
cmd.go
main.go
plugin.go
```

The `main.go` is very simple and does not need to be changed:

```golang
func main() {
	app := cli.NewApp()
	app.Name = "example-plugin"
	app.Usage = "example-plugin [command] [args...]"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		commands.StartCmd,
		commands.TxCmd,
		commands.KeyCmd,
		commands.QueryCmd,
		commands.AccountCmd,
	}
	app.Run(os.Args)
}
```

It creates the CLI, exactly like the `basecoin` one.
However, if we want our plugin to be active,
we need to make sure it is registered with the application.
In addition, if we want to send transactions to our plugin,
we need to add a new command to the CLI.
This is where the `cmd.go` comes in.

## Commands

First, we register the plugin:


```golang
func init() {
	commands.RegisterTxSubcommand(ExamplePluginTxCmd)
	commands.RegisterStartPlugin("example-plugin", func() types.Plugin { return NewExamplePlugin() })
}
```

This creates a new subcommand under `tx` (defined below),
and ensures the plugin is activated when we start the app.
Now we actually define the new command:

```golang
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
```

It's a simple command with one flag, which is just a boolean.
However, it actually inherits more flags from the Basecoin framework:

```golang
Flags: append(commands.TxFlags, ExampleFlag),
```

The `commands.TxFlags` is defined in `cmd/commands/tx.go`:

```golang
var TxFlags = []cli.Flag{
	NodeFlag,
	ChainIDFlag,

	FromFlag,

	AmountFlag,
	CoinFlag,
	GasFlag,
	FeeFlag,
	SeqFlag,
}
```

It adds all the default flags for a Basecoin transaction.

If we now compile and run our program, we can see all the options:

```
cd $GOPATH/src/github.com/tendermint/basecoin
go install ./docs/guide/src/example-plugin
example-plugin tx example --help
```

The output:

```
NAME:
   example-plugin tx example - Create, sign, and broadcast a transaction to the example plugin

USAGE:
   example-plugin tx example [command options] [arguments...]

OPTIONS:
   --node value      Tendermint RPC address (default: "tcp://localhost:46657")
   --chain_id value  ID of the chain for replay protection (default: "test_chain_id")
   --from value      Path to a private key to sign the transaction (default: "key.json")
   --amount value    Coins to send in transaction of the format <amt><coin>,<amt2><coin2>,... (eg: 1btc,2gold,5silver)
   --gas value       The amount of gas for the transaction (default: 0)
   --fee value       Coins for the transaction fee of the format <amt><coin>
   --sequence value  Sequence number for the account (default: 0)
   --valid           Set this to make the transaction valid
```

Cool, eh?

Before we move on to `plugin.go`, let's look at the `cmdExamplePluginTx` function in `cmd.go`:

```golang
func cmdExamplePluginTx(c *cli.Context) error {
	exampleFlag := c.Bool("valid")
	exampleTx := ExamplePluginTx{exampleFlag}
	return commands.AppTx(c, "example-plugin", wire.BinaryBytes(exampleTx))
}
```

We read the flag from the CLI library, and then create the example transaction.
Remember that Basecoin itself only knows about two transaction types, `SendTx` and `AppTx`.
All plugin data must be serialized (ie. encoded as a byte-array)
and sent as data in an `AppTx`. The `commands.AppTx` function does this for us -
it creates an `AppTx` with the corresponding data, signs it, and sends it on to the blockchain.

## RunTx

Ok, now we're ready to actually look at the implementation of the plugin in `plugin.go`.
Note I'll leave out some of the methods as they don't serve any purpose for this example,
but are necessary boilerplate.
Your plugin may have additional requirements that utilize these other methods.
Here's what's relevant for us:

```golang
type ExamplePluginState struct {
	Counter int
}

type ExamplePluginTx struct {
	Valid bool
}

type ExamplePlugin struct {
	name string
}

func (ep *ExamplePlugin) Name() string {
	return ep.name
}

func (ep *ExamplePlugin) StateKey() []byte {
	return []byte("ExamplePlugin.State")
}

func NewExamplePlugin() *ExamplePlugin {
	return &ExamplePlugin{
		name: "example-plugin",
	}
}

func (ep *ExamplePlugin) SetOption(store types.KVStore, key string, value string) (log string) {
	return ""
}

func (ep *ExamplePlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {

	// Decode tx
	var tx ExamplePluginTx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	// Validate tx
	if !tx.Valid {
		return abci.ErrInternalError.AppendLog("Valid must be true")
	}

	// Load PluginState
	var pluginState ExamplePluginState
	stateBytes := store.Get(ep.StateKey())
	if len(stateBytes) > 0 {
		err = wire.ReadBinaryBytes(stateBytes, &pluginState)
		if err != nil {
			return abci.ErrInternalError.AppendLog("Error decoding state: " + err.Error())
		}
	}

	//App Logic
	pluginState.Counter += 1

	// Save PluginState
	store.Set(ep.StateKey(), wire.BinaryBytes(pluginState))

	return abci.OK
}
```

All we're doing here is defining a state and transaction type for our plugin,
and then using the `RunTx` method to define how the transaction updates the state.
Let's break down `RunTx` in parts. First, we deserialize the transaction:


```golang
// Decode tx
var tx ExamplePluginTx
err := wire.ReadBinaryBytes(txBytes, &tx)
if err != nil {
	return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
}
```

The transaction is expected to be serialized according to Tendermint's "wire" format,
as defined in the `github.com/tendermint/go-wire` package.
If it's not encoded properly, we return an error.


If the transaction deserializes correctly, we can now check if it's valid:

```golang
// Validate tx
if !tx.Valid {
	return abci.ErrInternalError.AppendLog("Valid must be true")
}
```

The transaction is valid if the `Valid` field is set, otherwise it's not - simple as that.
Finally, we can update the state. In this example, the state simply counts how many valid transactions
we've processed. But the state itself is serialized and kept in some `store`, which is typically a Merkle tree.
So first we have to load the state from the store and deserialize it:

```golang
// Load PluginState
var pluginState ExamplePluginState
stateBytes := store.Get(ep.StateKey())
if len(stateBytes) > 0 {
	err = wire.ReadBinaryBytes(stateBytes, &pluginState)
	if err != nil {
		return abci.ErrInternalError.AppendLog("Error decoding state: " + err.Error())
	}
}
```

Note the state is stored under `ep.StateKey()`, which is defined above as `ExamplePlugin.State`. 
Also note, that we do nothing if there is no existing state data.  Is that a bug? No, we just make 
use of Go's variable initialization, that `pluginState` will contain a `Counter` value of 0. 
If your app needs more initialization than empty variables, then do this logic here in an `else` block.

Finally, we can update the state's `Counter`, and save the state back to the store:

```golang
//App Logic
pluginState.Counter += 1

// Save PluginState
store.Set(ep.StateKey(), wire.BinaryBytes(pluginState))

return abci.OK
```

And that's it! Now that we have a simple plugin, let's see how to run it.

## Running your plugin

In the [previous tutorial](basecoin-basics.md),
we used a pre-generated `genesis.json` and `priv_validator.json` for the application.
This time, let's make our own.

First, let's create a new directory and change into it:

```
mkdir example-data
cd example-data
```

Now, let's create a new private key:

```
example-plugin key new > key.json
```

Here's what my `key.json looks like:

```json
{
	"address": "15F591CA434CFCCBDEC1D206F3ED3EBA207BFE7D",
	"priv_key": [
		1,
		"737C629667A9EAADBB8E7CF792D5A8F63AA4BB51E06457DDD7FDCC6D7412AAAD43AA6C88034F9EB8D2717CA4BBFCBA745EFF19B13EFCD6F339EDBAAAFCD2F7B3"
	],
	"pub_key": [
		1,
		"43AA6C88034F9EB8D2717CA4BBFCBA745EFF19B13EFCD6F339EDBAAAFCD2F7B3"
	]
}
```

Now we can make a `genesis.json` file and add an account with out public key:

```json
[
  "base/chainID", "example-chain",
  "base/account", {
    "pub_key": [1, "43AA6C88034F9EB8D2717CA4BBFCBA745EFF19B13EFCD6F339EDBAAAFCD2F7B3"],
    "coins": [
	{
	  "denom": "gold",
	  "amount": 1000000000,
	}
    ]
  }
]
```

Here we've granted ourselves `1000000000` units of the `gold` token.

Before we can start the blockchain, we must initialize and/or reset the Tendermint state for a new blockchain:

```
tendermint init
tendermint unsafe_reset_all
```

Great, now we're ready to go.
To start the blockchain, simply run

```
example-plugin start --in-proc
```

In another window, we can try sending some transactions:

```
example-plugin tx send --to 0x1B1BE55F969F54064628A63B9559E7C21C925165 --amount 100gold --chain_id example-chain
```

Note the `--chain_id` flag. In the [previous tutorial](basecoin-basics.md),
we didn't include it because we were using the default chain ID ("test_chain_id").
Now that we're using a custom chain, we need to specify the chain explicitly on the command line.

Ok, so that's how we can send a `SendTx` transaction using our `example-plugin` CLI,
but we were already able to do that with the `basecoin` CLI.
With our new CLI, however, we can also send an `ExamplePluginTx`:

```
example-plugin tx example --amount 1gold --chain_id example-chain
```

The transaction is invalid! That's because we didn't specify the `--valid` flag:

```
example-plugin tx example --valid --amount 1gold --chain_id example-chain
```

Tada! We successfuly created, signed, broadcast, and processed our custom transaction type.

## Query

Now that we've sent a transaction to update the state, let's query for the state.
Recall that the state is stored under the key `ExamplePlugin.State`:


```
example-plugin query ExamplePlugin.State
```

Note the `"value":"0101"` piece. This is the serialized form of the state,
which contains only an integer.
If we send another transaction, and then query again, we'll see the value increment:

```
example-plugin tx example --valid --amount 1gold --chain_id example-chain
example-plugin query ExamplePlugin.State
```

Neat, right? Notice how the result of the query comes with a proof.
This is a Merkle proof that the state is what we say it is.
In a latter [tutorial on Interblockchain Communication](ibc.md),
we'll put this proof to work!

## Next Steps

In this tutorial we demonstrated how to create a new plugin and how to extend the
basecoin CLI to activate the plugin on the blockchain and to send transactions to it.
Hopefully by now you have some ideas for your own plugin, and feel comfortable implementing them.

In the [next tutorial](more-examples.md), we tour through some other plugin examples,
adding features for minting new coins, voting, and changin the Tendermint validator set.
But first, you may want to learn a bit more about [the design of the plugin system](plugin-design.md)
