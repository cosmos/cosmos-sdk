# Basecoin Plugins

In the [previous guide](basecoin-basics.md),
we saw how to use the `basecoin` tool to start a blockchain and send transactions.
We also learned about `Account` and `SendTx`, the basic data types giving us a multi-asset cryptocurrency.
Here, we will demonstrate how to extend the `basecoin` tool to use another transaction type, the `AppTx`,
to send data to a custom plugin. In this case we use a simple plugin that takes a single boolean argument,
and only accept the transaction if the argument is set to `true`.

## Example Plugin

The design of the `basecoin` tool makes it easy to extend for custom functionality.
To see what this looks like, install the `example-plugin` tool:

```
cd $GOPATH/src/github.com/tendermint/basecoin
go install ./docs/guide/src/example-plugin
```

The `example-plugin` tool is just like the `basecoin` tool. 
They both use the same library of commands, including one for signing and broadcasting `SendTx`.
See `example-plugin --help` for details.

A new blockchain can be initialized and started just like with `basecoin`:

```
example-plugin init
example-plugin start
```

The default files are stored in `~/.basecoin-example-plugin`.
In another window, we can send a `SendTx` like we are used to:

```
example-plugin tx send --to 0x1DA7C74F9C219229FD54CC9F7386D5A3839F0090 --amount 1mycoin
```

But the `example-plugin` tool has an additional command, `example-plugin tx example`, 
which crafts an `AppTx` specifically for our example plugin.
This command lets you send a single boolean argument:

```
example-plugin tx example --amount 1mycoin
example-plugin tx example --amount 1mycoin --valid
```

The first transaction is rejected by the plugin because it was not marked as valid, while the second transaction passes. 
We can build plugins that take many arguments of different types, and easily extend the tool to accomodate them.
Of course, we can also expose queries on our plugin: 

```
example-plugin query ExamplePlugin.State
```

Note the `"value":"0101"`. This is the serialized form of the state,
which contains only an integer, the number of valid transactions.
If we send another transaction, and then query again, we will see the value increment:

```
example-plugin tx example --valid --amount 1mycoin
example-plugin query ExamplePlugin.State
```

The value should now be `0102`, because we sent a second valid transaction. 
Notice how the result of the query comes with a proof.
This is a Merkle proof that the state is what we say it is.
In a latter [guide on InterBlockchain Communication](ibc.md),
we'll put this proof to work!


Now, before we implement our own plugin and tooling, it helps to understand the `AppTx` and the design of the plugin system.

## AppTx

The `AppTx` is similar to the `SendTx`, but instead of sending coins from inputs to outputs, 
it sends coins from one input to a plugin, and can also send some data.

```golang
type AppTx struct {
  Gas   int64   `json:"gas"`   
  Fee   Coin    `json:"fee"`   
  Input TxInput `json:"input"`
  Name  string  `json:"type"`  // Name of the plugin
  Data  []byte  `json:"data"`  // Data for the plugin to process
}
```

The `AppTx` enables Basecoin to be extended with arbitrary additional functionality through the use of plugins.
The `Name` field in the `AppTx` refers to the particular plugin which should process the transaction, 
and the `Data` field of the `AppTx` is the data to be forwarded to the plugin for processing.

Note the `AppTx` also has a `Gas` and `Fee`, with the same meaning as for the `SendTx`.
It also includes a single `TxInput`, which specifies the sender of the transaction,
and some coins that can be forwarded to the plugin as well.

## Plugins

A plugin is simply a Go package that implements the `Plugin` interface:

```golang
type Plugin interface {

  // Name of this plugin, should be short.
  Name() string

  // Run a transaction from ABCI DeliverTx
  RunTx(store KVStore, ctx CallContext, txBytes []byte) (res abci.Result)

  // Other ABCI message handlers
  SetOption(store KVStore, key string, value string) (log string)
  InitChain(store KVStore, vals []*abci.Validator)
  BeginBlock(store KVStore, hash []byte, header *abci.Header)
  EndBlock(store KVStore, height uint64) (res abci.ResponseEndBlock)
}

type CallContext struct {
  CallerAddress []byte   // Caller's Address (hash of PubKey)
  CallerAccount *Account // Caller's Account, w/ fee & TxInputs deducted
  Coins         Coins    // The coins that the caller wishes to spend, excluding fees
}
```

The workhorse of the plugin is `RunTx`, which is called when an `AppTx` is processed.
The `Data` from the `AppTx` is passed in as the `txBytes`, 
while the `Input` from the `AppTx` is used to populate the `CallContext`.

Note that `RunTx` also takes a `KVStore` - this is an abstraction for the underlying Merkle tree which stores the account data.
By passing this to the plugin, we enable plugins to update accounts in the Basecoin state directly, 
and also to store arbitrary other information in the state.
In this way, the functionality and state of a Basecoin-derived cryptocurrency can be greatly extended.
One could imagine going so far as to implement the Ethereum Virtual Machine as a plugin!

For details on how to initialize the state using `SetOption`, see the [guide to using the basecoin tool](basecoin-tool.md#genesis).


## Implement your own

To implement your own plugin and tooling, make a copy of `docs/guide/src/example-plugin`, 
and modify the code accordingly. Here, we will briefly describe the design and the changes to be made,
but see the code for more details.

First is the `main.go`, which drives the program. It can be left alone, but you should change any occurences of `example-plugin`
to whatever your plugin tool is going to be called.

Next is the `cmd.go`. This is where we extend the tool with any new commands and flags we need to send transactions to our plugin.
Note the `init()` function, where we register a new transaction subcommand with `RegisterTxSubcommand`, 
and where we load the plugin into the Basecoin app with `RegisterStartPlugin`.

Finally is the `plugin.go`, where we provide an implementation of the `Plugin` interface.
The most important part of the implementation is the `RunTx` method, which determines the meaning of the data
sent along in the `AppTx`. In our example, we define a new transaction type, the `ExamplePluginTx`, which
we expect to be encoded in the `AppTx.Data`, and thus to be decoded in the `RunTx` method, and used to update the plugin state.

For more examples and inspiration, see our [repository of example plugins](https://github.com/tendermint/basecoin-examples).

## Conclusion

In this guide, we demonstrated how to create a new plugin and how to extend the
`basecoin` tool to start a blockchain with the plugin enabled and send transactions to it.
In the next guide, we introduce a [plugin for Inter Blockchain Communication](ibc.md),
which allows us to publish proofs of the state of one blockchain to another,
and thus to transfer tokens and data between them.
