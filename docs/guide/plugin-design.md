# Basecoin Plugins

Basecoin implements a simple cryptocurrency, which is useful in and of itself,
but is far more useful if it can support additional functionality.
Here we describe how that functionality can be achieved through a plugin system.


## AppTx

In addition to the `SendTx`, Basecoin also defines another transaction type, the `AppTx`:

```
type AppTx struct {
  Gas   int64   `json:"gas"`   
  Fee   Coin    `json:"fee"`   
  Input TxInput `json:"input"`
  Name  string  `json:"type"`  // Name of the plugin
  Data  []byte  `json:"data"`  // Data for the plugin to process
}
```

The `AppTx` enables Basecoin to be extended with arbitrary additional functionality through the use of plugins.
The `Name` field in the `AppTx` refers to the particular plugin which should process the transasaction, 
and the `Data` field of the `AppTx` is the data to be forwarded to the plugin for processing.

Note the `AppTx` also has a `Gas` and `Fee`, with the same meaning as for the `SendTx`.
It also includes a single `TxInput`, which specifies the sender of the transaction,
and some coins that can be forwarded to the plugin as well.

## Plugins

A plugin is simply a Go package that implements the `Plugin` interface:

```
type Plugin interface {

  // Name of this plugin, should be short.
  Name() string

  // Run a transaction from ABCI DeliverTx
  RunTx(store KVStore, ctx CallContext, txBytes []byte) (res abci.Result)

  // Other ABCI message handlers
  SetOption(store KVStore, key string, value string) (log string)
  InitChain(store KVStore, vals []*abci.Validator)
  BeginBlock(store KVStore, height uint64)
  EndBlock(store KVStore, height uint64) []*abci.Validator
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
In this way, the functionality and state of a Basecoin-derrived cryptocurrency can be greatly extended.
One could imagine going so far as to implement the Ethereum Virtual Machine as a plugin!


## Next steps

1. Examples of [Basecoin plugins](more-examples.md)
1. Learn how to use [InterBlockchain Communication (IBC)](ibc.md)
1. [Deploy testnets](deployment.md) running your basecoin application.
