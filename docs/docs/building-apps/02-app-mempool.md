---
sidebar_position: 1
---

# Application mempool

:::note Synopsis
This sections describes how the app side mempool can be used and replaced. 
:::

Since `v0.47` the application has its own mempool to allow much more granular block building than previous versions. This change was enabled by [ABCI 1.0](https://github.com/tendermint/tendermint/blob/main/spec/abci/README.md). Notably it introduces the prepare and process proposal steps of ABCI. 

## Prepare Proposal

Prepare proposal handles construction of the block, meaning that when a proposer is preparing to propose a block it asks the application if the txs it collected from the mempool are the right ones, at which point the application will check its own mempool for txs that it would like to include. Now, reading mempool twice in the previous sentence is confusing, lets break it down. Tendermint has a mempool that handles bradcasting transactions to other nodes in the network, but it does not handle ordering of these transactions. The ordering happens at the application level in its own mempool. Allowing the application to handle ordering enables the application to define how it would like the block constructed. 

Currently, there is a default `PrepareProposal` implementation provided by the application.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/main/baseapp/baseapp.go#L870-L908
```

This default implementation can be overridden by the application developer in favor of a custom implementation in [`app.go`](./01-app-go-v2.md):

```go
prepareOpt := func(app *baseapp.BaseApp) {
	app.SetPrepareProposal(app.DefaultPrepareProposal())
}

baseAppOptions = append(baseAppOptions, prepareOpt)
```

## Process Proposal

Process proposal handles the validation of what is in a block, meaning that after a block has been proposed the other validators have the right to vote no or yes on a block. The validator in the default implementation of `PrepareProposal` runs the transaction in a non execution fashion, it runs the antehandler and gas operations to make sure the transaction is valid. 

Here is the implementation of the default implementation:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-alpha2/baseapp/baseapp.go#L911-L935
```

Like `PrepareProposal` this implementation is the default and can be modified by the application developer in [`app.go`](./01-app-go-v2.md):

```go
processOpt := func(app *baseapp.BaseApp) {
	app.SetProcessProposal(app.DefaultProcessProposal())
}

baseAppOptions = append(baseAppOptions, processOpt)
```

## Mempool

Now that we have walked through the `PrepareProposal` & `ProcessProposal`, we can move on to walking through the mempool. 

There are countless designs that an application developer can write for a mempool, the SDK opted to provide only simple mempool implementations.
Namely, the SDK provides the following mempools:

* [No-op Mempool](#no-op-mempool)
* [Sender Nonce Mempool](#sender-nonce-mempool)
* [Priority Nonce Mempool](#priority-nonce-mempool)

The default SDK is a [No-op Mempool](#no-op-mempool), but it can be replaced by the application developer in [`app.go`](./01-app-go-v2.md):

```go
nonceMempool := mempool.NewSenderNonceMempool()
mempoolOpt   := baseapp.SetMempool(nonceMempool)
baseAppOptions = append(baseAppOptions, mempoolOpt)
```

### No-op Mempool

A no-op mempool is a mempool where transactions are completely discarded and ignored when BaseApp interacts with the mempool.
When this mempool is used, it assumed that an application will rely on Tendermint's transaction ordering defined in `RequestPrepareProposal`,
which is FIFO-ordered by default.

### Sender Nonce Mempool

The nonce mempool is a mempool that keeps transactions from an sorted by nonce in order to avoid the issues with nonces. 
It works by storing the transation in a list sorted by the transaction nonce. When the proposer asks for transactions to be included in a block it randomly selects a sender and gets the first transaction in the list. It repeats this until the mempool is empty or the block is full. 

It is configurable with the following parameters:

#### MaxTxs

It is an integer value that sets the mempool in one of three modes, *bounded*, *unbounded*, or *disabled*.

* **negative**: Disabled, mempool does not insert new transaction and return early.
* **zero**: Unbounded mempool has no transaction limit and will never fail with `ErrMempoolTxMaxCapacity`.
* **positive**: Bounded, it fails with `ErrMempoolTxMaxCapacity` when `maxTx` value is the same as `CountTx()`

#### Seed

Set the seed for the random number generator used to select transactions from the mempool.

### Priority Nonce Mempool

The [priority nonce mempool](https://github.com/cosmos/cosmos-sdk/blob/main/types/mempool/priority_nonce_spec.md) is a mempool implementation that stores txs in a partially ordered set by 2 dimensions:

* priority
* sender-nonce (sequence number)

Internally it uses one priority ordered [skip list](https://pkg.go.dev/github.com/huandu/skiplist) and one skip list per sender ordered by sender-nonce (sequence number). When there are multiple txs from the same sender, they are not always comparable by priority to other sender txs and must be partially ordered by both sender-nonce and priority.

It is configurable with the following parameters:

#### MaxTxs

It is an integer value that sets the mempool in one of three modes, *bounded*, *unbounded*, or *disabled*.

* **negative**: Disabled, mempool does not insert new transaction and return early.
* **zero**: Unbounded mempool has no transaction limit and will never fail with `ErrMempoolTxMaxCapacity`.
* **positive**: Bounded, it fails with `ErrMempoolTxMaxCapacity` when `maxTx` value is the same as `CountTx()`

#### Callback

Allow to set a callback to be called when a transaction is read from the mempool.

More information on the SDK mempool implementation can be found in the [godocs](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/types/mempool).
