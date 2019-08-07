# Context

## Prerequisites

* [Anatomy of an SDK Application](../basics/app-anatomy.md)
* [Lifecycle of a Transaction](../basics/tx-lifecycle.md)

## Synopsis

This document details the SDK `Context` type.

- [Go Context Package](#go-context-package)
- [SDK Context Type Definition](#context-definition)


## Context Definition

The SDK Context is a custom data structure that uses a Go Context Package [context](https://golang.org/pkg/context) as its base, and has many additional types within its definition that are specific to the SDK and Tendermint. The Context is directly passed between processes as an argument; it serves as a way to cheaply pass a cached version of the [state](./multistore.md) as well as information relevant to the transaction processing.

```go
type Context struct {
	ctx           context.Context
	ms            MultiStore
	header        abci.Header
	chainID       string
	txBytes       []byte
	logger        log.Logger
	voteInfo      []abci.VoteInfo
	gasMeter      GasMeter
	blockGasMeter GasMeter
	checkTx       bool
	minGasPrice   DecCoins
	consParams    *abci.ConsensusParams
	eventManager  *EventManager
}
```

- **Context:** The base type is a Go [Context](https://golang.org/pkg/context).
- **Multistore:** Every context belongs to a particular application and has a reference to its [multistore](./multistore.md). The SDK Context also has `KVStore()` and `TransientStore()` methods to fetch a KVStore or TransientStore from the Multistore using a key.
- **ABCI Header:** The [header](https://tendermint.com/docs/spec/abci/abci.html#header) is an ABCI type. It carries important information about the state of the blockchain, such as block height and proposer of the current block.
- **Chain ID:** The unique identification number of the blockchain a block pertains to.
- **Transaction Bytes:** The `[]byte` representation of a transaction being processed using the context. Every transaction is processed by various parts of the SDK and consensus engine (e.g. Tendermint) throughout its [lifecycle](../basics/tx-lifecycle.md), some of which to not have any understanding of transaction types. Thus, transactions are marshaled into the generic `[]byte` type using some kind of [encoding format](./encoding.md) such as [Amino](./amino.md).
- **Logger:** A [Logger](https://github.com/tendermint/tendermint/blob/master/libs/log/logger.go) from the Tendermint libraries. Learn more about logs [here](https://tendermint.com/docs/tendermint-core/how-to-read-logs.html#how-to-read-logs).
- **VoteInfo:** A list of the ABCI type [`VoteInfo`](https://tendermint.com/docs/spec/abci/abci.html#voteinfo), which includes the name of a validator and a boolean indicating whether they have signed the block.
- **Gas Meters:** Specifically, a `gasMeter` for the transaction currently being processed using the context and a `blockGasMeter` for the entire block it belongs to. Users specify how much in fees they wish to pay for the execution of their transaction; these gas meters keep track of how much [gas](../basics/accounts-fees-gas.md) has been used in the transaction or block so far. If the gas meter runs out, execution halts.
- **CheckTx Mode:** A boolean value indicating whether a transaction should be processed in `CheckTx` or `DeliverTx` mode.
- **Min Gas Price:** The minimum [gas](../basics/accounts-fees-gas.md) price a node is willing to take in order to include a transaction in its block. This price is a local value configured by each node individually.
- **Consensus Params:** The ABCI type [Consensus Parameters](https://tendermint.com/docs/spec/abci/apps.html#consensus-parameters), which enforce certain limits for the blockchain, such as maximum gas for a block.
- **Event Manager:** An event manager with a number of [`Event`](https://github.com/cosmos/cosmos-sdk/blob/master/types/events.go) objects that can be emitted from. Modules developers may define module `events` and `attributes` or use the SDK types, and emit an `EventTypeMessage` whenever a message is handled. For example:
```go
ctx.EventManager().EmitEvent(sdk.NewEvent(
    sdk.EventTypeMessage,
    sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),))
```

## Go Context Package

A basic Context is defined in the [Golang Context Package](https://golang.org/pkg/context). A Context is an immutable data structure that carries request-scoped data across APIs and processes. Contexts are also designed to enable concurrency and to be used in goroutines.

Contexts are intended to be **immutable**. They should never be edited. Instead, the convention is to create a child context from its parent using a `With` function. For example:

``` go
childCtx = parentCtx.WithBlockHeader(header)
```

The [Golang Context Package]((https://golang.org/pkg/context) documentation instructs developers to explicitly pass a context `ctx` as the first argument of a process.

## Cache Wrapping

The Context contains a Multistore, which allows for cache wrapping functionality: a deep copy of the state stored in a Multistore can be made. Processes are free to write changes to the copy of the state, then write the changes back to the original state or disregard them if something goes wrong. The pattern of usage for a Context is as follows:

1. A process receives a Context `ctx` from its parent process, which provides information needed to perform the process.
2. The `ctx` is [**cache wrapped**](./multistore.md), i.e. a copy of the [multistore](./multistore.md) is made so that the process can make changes to the state as it executes, without changing the original `ctx`.
3. The process may read and write from `ctx` as it is executing. It may call a subroutine and passes `ctx` to it as the first argument.
4. When a subroutine returns, it checks if the result is a success or failure. If a failure, nothing needs to be done - the cache wrapped `ctx` is simply discarded. If successful, the changes made to the copy can be committed to the original `ctx`.

For example, here is a codeblock from the [`runTx`](./baseapp.md#runtx-and-runmsgs) function in [`baseapp`](./baseapp.md):

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

1. Prior to calling `runMsgs` on the message(s) in the transaction, it uses `app.cacheTxContext()` to cache-wrap the context and multistore.
2. The cache-wrapped context, `runMsgCtx`, is used in `runMsgs` to return a result.
3. If the process is running in [`checkTxMode`](./baseapp.md#checktx), there is no need to write the changes - the result is returned immediately. 
4. If the process is running in [`deliverTxMode`](./baseapp.md#delivertx) and the result indicates a successful run over all the messages, the cached multistore is written back to the original.

## Next

Read about the next core concept.
