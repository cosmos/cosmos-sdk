<!--
order: 3
-->

# Context

The `context` is a data structure intended to be passed from function to function that carries information about the current state of the application. It provides an access to a branched storage (a safe branch of the entire state) as well as useful objects and information like `gasMeter`, `block height`, `consensus parameters` and more. {synopsis}

## Pre-requisites Readings

- [Anatomy of an SDK Application](../basics/app-anatomy.md) {prereq}
- [Lifecycle of a Transaction](../basics/tx-lifecycle.md) {prereq}

## Context Definition

The SDK `Context` is a custom data structure that contains Go's stdlib [`context`](https://golang.org/pkg/context) as its base, and has many additional types within its definition that are specific to the Cosmos SDK. he `Context` is integral to transaction processing in that it allows modules to easily access their respective [store](./store.md#base-layer-kvstores) in the [`multistore`](./store.md#multistore) and retrieve transactional context such as the block header and gas meter.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/types/context.go#L16-L39

- **Context:** The base type is a Go [Context](https://golang.org/pkg/context), which is explained further in the [Go Context Package](#go-context-package) section below.
- **Multistore:** Every application's `BaseApp` contains a [`CommitMultiStore`](./store.md#multistore) which is provided when a `Context` is created. Calling the `KVStore()` and `TransientStore()` methods allows modules to fetch their respective [`KVStore`](./store.md#base-layer-kvstores) using their unique `StoreKey`.
- **ABCI Header:** The [header](https://tendermint.com/docs/spec/abci/abci.html#header) is an ABCI type. It carries important information about the state of the blockchain, such as block height and proposer of the current block.
- **Chain ID:** The unique identification number of the blockchain a block pertains to.
- **Transaction Bytes:** The `[]byte` representation of a transaction being processed using the context. Every transaction is processed by various parts of the SDK and consensus engine (e.g. Tendermint) throughout its [lifecycle](../basics/tx-lifecycle.md), some of which to not have any understanding of transaction types. Thus, transactions are marshaled into the generic `[]byte` type using some kind of [encoding format](./encoding.md) such as [Amino](./encoding.md).
- **Logger:** A `logger` from the Tendermint libraries. Learn more about logs [here](https://tendermint.com/docs/tendermint-core/how-to-read-logs.html#how-to-read-logs). Modules call this method to create their own unique module-specific logger.
- **VoteInfo:** A list of the ABCI type [`VoteInfo`](https://tendermint.com/docs/spec/abci/abci.html#voteinfo), which includes the name of a validator and a boolean indicating whether they have signed the block.
- **Gas Meters:** Specifically, a [`gasMeter`](../basics/gas-fees.md#main-gas-meter) for the transaction currently being processed using the context and a [`blockGasMeter`](../basics/gas-fees.md#block-gas-meter) for the entire block it belongs to. Users specify how much in fees they wish to pay for the execution of their transaction; these gas meters keep track of how much [gas](../basics/gas-fees.md) has been used in the transaction or block so far. If the gas meter runs out, execution halts.
- **CheckTx Mode:** A boolean value indicating whether a transaction should be processed in `CheckTx` or `DeliverTx` mode.
- **Min Gas Price:** The minimum [gas](../basics/gas-fees.md) price a node is willing to take in order to include a transaction in its block. This price is a local value configured by each node individually, and should therefore **not be used in any functions used in sequences leading to state-transitions**.
- **Consensus Params:** The ABCI type [Consensus Parameters](https://tendermint.com/docs/spec/abci/apps.html#consensus-parameters), which specify certain limits for the blockchain, such as maximum gas for a block.
- **Event Manager:** The event manager allows any caller with access to a `Context` to emit [`Events`](./events.md). Modules may define module specific
  `Events` by defining various `Types` and `Attributes` or use the common definitions found in `types/`. Clients can subscribe or query for these `Events`. These `Events` are collected throughout `DeliverTx`, `BeginBlock`, and `EndBlock` and are returned to Tendermint for indexing. For example:

```go
ctx.EventManager().EmitEvent(sdk.NewEvent(
    sdk.EventTypeMessage,
    sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory)),
)
```

## Go Context Package

A basic `Context` is defined in the [Golang Context Package](https://golang.org/pkg/context). A `Context`
is an immutable data structure that carries request-scoped data across APIs and processes. Contexts
are also designed to enable concurrency and to be used in goroutines.

Contexts are intended to be **immutable**; they should never be edited. Instead, the convention is
to create a child context from its parent using a `With` function. For example:

```go
childCtx = parentCtx.WithBlockHeader(header)
```

The [Golang Context Package](https://golang.org/pkg/context) documentation instructs developers to
explicitly pass a context `ctx` as the first argument of a process.

## Store branching

The `Context` contains a `MultiStore`, which allows for branchinig and caching functionality using `CacheMultiStore` 
(queries in `CacheMultiStore` are cached to avoid future round trips). 
Each `KVStore` is branched in a safe and isolated ephemeral storage. Processes are free to write changes to
the `CacheMultiStore`. If a state-transition sequence is performed without issue, the store branch can 
be committed to the underlying store at the end of the sequence or disregard them if something
goes wrong. The pattern of usage for a Context is as follows:

1. A process receives a Context `ctx` from its parent process, which provides information needed to
   perform the process.
2. The `ctx.ms` is a **branched store**, i.e. a branch of the [multistore](./store.md#multistore) is made so that the process can make changes to the state as it executes, without changing the original`ctx.ms`. This is useful to protect the underlying multistore in case the changes need to be reverted at some point in the execution.
3. The process may read and write from `ctx` as it is executing. It may call a subprocess and pass
   `ctx` to it as needed.
4. When a subprocess returns, it checks if the result is a success or failure. If a failure, nothing
   needs to be done - the branch `ctx` is simply discarded. If successful, the changes made to
   the `CacheMultiStore` can be committed to the original `ctx.ms` via `Write()`.

For example, here is a snippet from the [`runTx`](./baseapp.md#runtx-and-runmsgs) function in
[`baseapp`](./baseapp.md):

```go
runMsgCtx, msCache := app.cacheTxContext(ctx, txBytes)
result = app.runMsgs(runMsgCtx, msgs, mode)
result.GasWanted = gasWanted

if mode != runTxModeDeliver {
  return result
}

if result.IsOK() {
  msCache.Write()
}
```

Here is the process:

1. Prior to calling `runMsgs` on the message(s) in the transaction, it uses `app.cacheTxContext()`
   to branch and cache the context and multistore.
2. `runMsgCtx` - the context with branched store, is used in `runMsgs` to return a result.
3. If the process is running in [`checkTxMode`](./baseapp.md#checktx), there is no need to write the
   changes - the result is returned immediately.
4. If the process is running in [`deliverTxMode`](./baseapp.md#delivertx) and the result indicates
   a successful run over all the messages, the branched multistore is written back to the original.

## Next {hide}

Learn about the [node client](./node.md) {hide}
