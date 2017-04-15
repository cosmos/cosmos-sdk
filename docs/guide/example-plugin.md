# Basecoin Example Plugin

In the [previous tutorial](basecoin-basics.md),
we saw how to start a Basecoin blockchain and use the CLI to send transactions.
Here, we will demonstrate how to extend the blockchain and CLI to support a simple plugin.

## Overview

Creating a new plugin and CLI to support it requires a little bit of
boilerplate, but not much.  For convenience, we've implemented an extremely
simple example plugin that can be easily modified.  The example is under
`docs/guide/src/example-plugin`.  To build your own plugin, copy this folder to
a new location and start modifying it there.

Let's take a look at the files in `docs/guide/src/example-plugin`:

```
cmd.go
main.go
plugin.go
```

### main.go

The `main.go` is very simple and does not need to be changed:

```golang
func main() {
	//Initialize example-plugin root command
	var RootCmd = &cobra.Command{
		Use:   "example-plugin",
		Short: "example-plugin usage description",
	}

	//Add the default basecoin commands to the root command
	RootCmd.AddCommand(
		commands.InitCmd,
		commands.StartCmd,
		commands.TxCmd,
		commands.QueryCmd,
		commands.KeyCmd,
		commands.VerifyCmd,
		commands.BlockCmd,
		commands.AccountCmd,
		commands.UnsafeResetAllCmd,
	)

	//Run the root command
	commands.ExecuteWithDebug(RootCmd)
}
```

It creates the CLI, exactly like the `basecoin` one.  However, if we want our
plugin to be active, we need to make sure it is registered with the
application.  In addition, if we want to send transactions to our plugin, we
need to add a new command to the CLI.  This is where the `cmd.go` comes in.

### cmd.go

First we define the new CLI command and associated flag variables.

```golang
var (
	//CLI Flags
	validFlag bool

	//CLI Plugin Commands
	ExamplePluginTxCmd = &cobra.Command{
		Use:   "example",
		Short: "Create, sign, and broadcast a transaction to the example plugin",
		RunE:   examplePluginTxCmd,
	}
)
```

Next within the `init` function we register our plugin's flags and register our
custom plugin command with the root command. This creates a new subcommand
under `tx` (defined below), and ensures the plugin is activated when we start
the app.

```golang
func init() {

	//Set the Plugin Flags
	ExamplePluginTxCmd.Flags().BoolVar(&validFlag, "valid", false, "Set this to make transaction valid")

	//Register a plugin specific CLI command as a subcommand of the tx command
	commands.RegisterTxSubcommand(ExamplePluginTxCmd)

	//Register the example with basecoin at start
	commands.RegisterStartPlugin("example-plugin", func() types.Plugin { return NewExamplePlugin() })
}
```

We now define the actual function which is called by our CLI command. 

```golang
func examplePluginTxCmd(cmd *cobra.Command, args []string) error {
	exampleTx := ExamplePluginTx{validFlag}
	exampleTxBytes := wire.BinaryBytes(exampleTx)
	return commands.AppTx("example-plugin", exampleTxBytes)
}
```

Our function is a simple command with one boolean flag.  However, it actually
inherits the persistent flags from the Basecoin framework. These persistent
flags use pointers to these variables stored in `cmd/commands/tx.go`:

```golang
var (
	//persistent flags
	txNodeFlag  string
	amountFlag  string
	fromFlag    string
	seqFlag     int
	gasFlag     int
	feeFlag     string
	chainIDFlag string
	
	//non-persistent flags
	toFlag      string
	dataFlag    string
	nameFlag    string
)
```

If we now compile and run our program, we can see all the options:

```
cd $GOPATH/src/github.com/tendermint/basecoin
go install ./docs/guide/src/example-plugin
example-plugin tx example --help
```

The output:

```
Create, sign, and broadcast a transaction to the example plugin

Usage:
  example-plugin tx example [flags]

Flags:
      --valid   Set this to make transaction valid

Global Flags:
      --amount string     Coins to send in transaction of the format <amt><coin>,<amt2><coin2>,... (eg: 1btc,2gold,5silver},
      --chain_id string   ID of the chain for replay protection (default "test_chain_id")
      --fee string        Coins for the transaction fee of the format <amt><coin>
      --from string       Path to a private key to sign the transaction (default "key.json")
      --gas int           The amount of gas for the transaction
      --node string       Tendermint RPC address (default "tcp://localhost:46657")
      --sequence int      Sequence number for the account (-1 to autocalculate}, (default -1)
```

Cool, eh?

Before we move on to `plugin.go`, let's look at the `examplePluginTxCmd`
function in `cmd.go`:

```golang
func examplePluginTxCmd(cmd *cobra.Command, args []string) {
	exampleTx := ExamplePluginTx{validFlag}
	exampleTxBytes := wire.BinaryBytes(exampleTx)
	commands.AppTx("example-plugin", exampleTxBytes)
}
```

We read the flag from the CLI library, and then create the example transaction.
Remember that Basecoin itself only knows about two transaction types, `SendTx`
and `AppTx`.  All plugin data must be serialized (ie. encoded as a byte-array)
and sent as data in an `AppTx`. The `commands.AppTx` function does this for us
- it creates an `AppTx` with the corresponding data, signs it, and sends it on
to the blockchain.

### plugin.go

Ok, now we're ready to actually look at the implementation of the plugin in
`plugin.go`.  Note I'll leave out some of the methods as they don't serve any
purpose for this example, but are necessary boilerplate.  Your plugin may have
additional requirements that utilize these other methods.  Here's what's
relevant for us:

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
	
	// Decode txBytes using go-wire. Attempt to write the txBytes to the variable
	// tx, if the txBytes have not been properly encoded from a ExamplePluginTx
	// struct wire will produce an error.
	var tx ExamplePluginTx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	// Perform Transaction Validation
	if !tx.Valid {
		return abci.ErrInternalError.AppendLog("Valid must be true")
	}

	// Load PluginState
	var pluginState ExamplePluginState
	stateBytes := store.Get(ep.StateKey())
	// If the state does not exist, stateBytes will be initialized
	// as an empty byte array with length of zero
	if len(stateBytes) > 0 {
		err = wire.ReadBinaryBytes(stateBytes, &pluginState) //decode using go-wire
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
and then using the `RunTx` method to define how the transaction updates the
state.  Let's break down `RunTx` in parts. First, we deserialize the
transaction:

```golang
var tx ExamplePluginTx
err := wire.ReadBinaryBytes(txBytes, &tx)
if err != nil {
	return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
}
```

The transaction is expected to be serialized according to Tendermint's "wire"
format, as defined in the `github.com/tendermint/go-wire` package.  If it's not
encoded properly, we return an error.


If the transaction deserializes correctly, we can now check if it's valid:

```golang
if !tx.Valid {
	return abci.ErrInternalError.AppendLog("Valid must be true")
}
```

The transaction is valid if the `Valid` field is set, otherwise it's not -
simple as that.  Finally, we can update the state. In this example, the state
simply counts how many valid transactions we've processed. But the state itself
is serialized and kept in some `store`, which is typically a Merkle tree.  So
first we have to load the state from the store and deserialize it:

```golang
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

First, initialize the new blockchain with 

```
basecoin init
```

If you've already run a basecoin blockchain, reset the data with

```
basecoin unsafe_reset_all
```

To start the blockchain with your new plugin, simply run

```
example-plugin start 
```

In another window, we can try sending some transactions:

```
example-plugin tx send --to 0x1DA7C74F9C219229FD54CC9F7386D5A3839F0090 --amount 100mycoin
```

Ok, so that's how we can send a `SendTx` transaction using our `example-plugin` CLI,
but we were already able to do that with the `basecoin` CLI.
With our new CLI, however, we can also send an `ExamplePluginTx`:

```
example-plugin tx example --amount 1mycoin
```

The transaction is invalid! That's because we didn't specify the `--valid` flag:

```
example-plugin tx example --valid --amount 1mycoin
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
example-plugin tx example --valid --amount 1mycoin
example-plugin query ExamplePlugin.State
```

Neat, right? Notice how the result of the query comes with a proof.
This is a Merkle proof that the state is what we say it is.
In a latter [tutorial on InterBlockchain Communication](ibc.md),
we'll put this proof to work!

## Next Steps

In this tutorial we demonstrated how to create a new plugin and how to extend the
basecoin CLI to activate the plugin on the blockchain and to send transactions to it.
Hopefully by now you have some ideas for your own plugin, and feel comfortable implementing them.

In the [next tutorial](more-examples.md), we tour through some other plugin examples,
addin mple-plugin query ExamplePlugin.Statefeatures for minting new coins, voting, and changing the Tendermint validator set.
But first, you may want to learn a bit more about [the design of the plugin system](plugin-design.md)
