# BaseApp

## Pre-requisite Reading

- [Anatomy of an SDK application](./app-anatomy.md)
- [Lifecycle of an SDK transaction](./tx-lifecycle.md)

## Synopsis

This document describes `baseapp`, the abstraction that implements most of the common functionalities of an SDK application. 

- [Introduction](#introduction)
- [Type Definition](#type-definition)
- [States](#states)
- [Routing](#routing)
- [ABCI](#abci)
- [CheckTx](#abci-checktx)
- [DeliverTx](#abci-delivertx)
- [Commit](#abbci-commit)
- [Other ABCI Message](#other-abci-message)
    + [Info](#info)
    + [SetOption](#setoption)
    + [Query](#query)
    + [InitChain](#initchain)
    + [BeginBlock](#beginblock)
    + [EndBlock](#endblock)
- [Gas](#gas)


## Introduction

`baseapp` is an abstraction that implements the core of an SDK application, namely:

- The [Application-Blockchain Interface](#abci), for the state-machine to communicate with the underlying consensus engine (e.g. Tendermint). 
- A [Router](#routing), to route [messages](./tx-msgs.md) and [queries](./querier.md) to the appropriate [module](./modules.md).
- Different [states](#states), as the state-machine can have different parallel states updated based on the ABCI message received. 

The goal of `baseapp` is to provide a boilerplate SDK application that developers can easily extend to build their own custom application. Usually, developers will create a custom type for their application, like so:

```go 
type app struct {
    *bam.BaseApp // reference to baseapp
    cdc *codec.Codec

    // list of application store keys

    // list of application keepers

    // module manager
}
```

Extending the application with `baseapp` gives the former access to all of `baseapp`'s methods. This allows developers to compose their custom application with the modules they want, while not having to concern themselves with the hard work of implementing the ABCI, the routing and state management logic. 

## Type Definition

The [`baseapp` type](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/baseapp.go#L45-L91) holds many important parameters for any Cosmos SDK based application. Let us go through the most important components.

*Note: Not all parameters are described, only the most important ones. Refer to the [type definition](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/baseapp.go#L45-L91) for the full list* 

First, the important parameters that are initialized during the initialization of the application:

- A [`CommitMultiStore`](./store.md#commit-multi-store). This is the main store of the application, which holds the canonical state that is committed at the [end of each block](#commit). This store is **not** cached, meaning it is not used to update the application's intermediate (un-committed) states. The `CommitMultiStore` is a multi-store, meaning a store of stores. Each module of the application uses one or multiple `KVStores` in the multi-store to persist their subset of the state. 
- A [database](./store.md#database) `db`, which is used by the `CommitMultiStore` to handle data storage.
- A [router](#messages). The `router` facilitates the routing of [messages](./tx-msgs.md) to the appropriate module for it to be processed.
- A [query router](#queries). The `query router` facilitates the routing of [queries](./querier.md) to the appropriate module for it to be processed.
- A [`txDecoder`](https://godoc.org/github.com/cosmos/cosmos-sdk/types#TxDecoder), used to decode transaction `[]byte` relayed by the underlying Tendermint engine.
- A [`baseKey`], to access the [main store](./store.md#main-store) in the `CommitMultiStore`. The main store is used to persist data related to the core of the application, like consensus parameters.  
- A [`anteHandler`](./accounts-fees.md#antehandler), to handle signature verification and fee paiement when a transaction is received.
- An [`initChainer`](./app-anatomy.md#initchainer), [`beginBlocker` and `endBlocker`](./app-anatomy.md#beginblocker-and-endblocker), which are the functions executed when the application received the [InitChain], [BeginBlock] and [EndBlock] messages from the underlying Tendermint engine. 

Then, parameters used to define [volatile states](#volatile-states) (i.e. cached states):

- `checkState`: This state is updated during [`CheckTx`](#checktx), and reset on [`Commit`](#commit).
- `deliverState`: This state is updated during [`DeliverTx`](#delivertx), and reset on [`Commit`](#commit).

Finally, a few more important parameterd:

- `voteInfos`: This parameter carries the list of validators whose precommit is missing, either because they did not vote or because the proposer did not include their vote. This information is carried by the [context](#context) and can be used by the application for various things like punishing absent validators.
- `minGasPrices`: This parameter defines the minimum [gas prices](./accounts-fees.md#gas) accepted by the node. This is a local parameter, meaning each full-node can set a different `minGasPrices`. It is run by the [`anteHandler`](./accounts-fees.md#antehandler) during `CheckTx`, mainly as a spam protection mechanism. The transaction enters the [mempool](https://tendermint.com/docs/tendermint-core/mempool.html#transaction-ordering) only if the gas price of the transaction is superior to one of the minimum gas price in `minGasPrices` (i.e. if `minGasPrices == 1uatom, 1upho`, the `gas-price` of the transaction must be superior to `1uatom` OR `1upho`).
- `appVersion`: Version of the application. It is set in the [application's constructor function](./baseapp.md#constructor-function).

## States 

`baseapp` handles various parallel states for different purposes. There is the [main state](#main-state), which is the canonical state of the application, and volatile states like [`checkState`](#checkState) and [`deliverState`](#deliverstate), which are used to handle temporary states inbetween updates of the main state.

### Main State

The main state is the canonical state of the application. It is initialized on [`InitChain`](#initchain and updated on [`Commit`](#abci-commit) at the end of each block. 

```
+--------+                              +--------+
|        |                              |        |
|   S    +----------------------------> |   S'   |
|        |   For each T in B: apply(T)  |        |
+--------+                              +--------+
```

The main state is held by `baseapp` in a structure called the [`CommitMultiStore`](./store.md#commit-multi-store). This multi-store is used by developers to instantiate all the stores they need for each of their application's modules.   

### Volatile States

Volatile - or cached - states are used in between [`Commit`s](#commit) to manage temporary states. They are reset to the latest version of the main state after it is committed. There are two main volatile states:

- `checkState`: This cached state is initialized during [`InitChain`](#initchain), updated during [`CheckTx`](#abci-checktx) when an unconfirmed transaction is received, and reset to the [main state](#main-state) on [`Commit`](#abci-commit). 
- `deliverState`: This cached state is initialized during [`BeginBlock`](#beginblock), updated during [`DeliverTx`](#abci-delivertx) when a transaction included in a block is processed, and reset to the [main state](#main-state) on [`Commit`](#abci-commit).

## Routing

When messages and queries are received by the application, they must be routed to the appropriate module in order to be processed. Routing is done via `baseapp`, which holds a `router` for messages, and a `query router` for queries.

### Message Routing

Messages need to be routed after they are extracted from transactions, which are sent from the underlying Tendermint engine via the [`CheckTx`](#checktx) and [`DeliverTx`](#delivertx) ABCI messages. To do so, `baseapp` holds a [`router`](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/router.go) which maps `paths` (`string`) to the appropriate module [`handler`](./handler.md). Usually, the `path` is the name of the module.

The application's `router` is initilalized with all the routes using the application's [module manager](./modules.md#module-manager), which itself is initialized with all the application's modules in the application's [constructor](./app-anatomy.md#app-constructor). 

### Query Routing

Similar to messages, queries need to be routed to the appropriate module's [querier](./querier.md). To do so, `baseapp` holds a [`query router`](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/queryrouter.go), which maps `paths` (`string`) to the appropriate module [`querier`](./querier.md). Usually, the `path` is the name of the module. 

Just like the `router`, the `query router` is initilalized with all the query routes using the application's [module manager](./modules.md#module-manager), which itself is initialized with all the application's modules in the application's [constructor](./app-anatomy.md#app-constructor).

## Main ABCI Messages



### CheckTx

### DeliverTx

## RunTx and RunMsg

## Other ABCI Messages

### Commit

### Info

### SetOption

### Query 

### InitChain

### BeginBlock

### EndBlock