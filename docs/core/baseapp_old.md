# BaseApp

## Pre-requisite Reading

- [Anatomy of an SDK application](./app-anatomy.md)
- [Lifecycle of an SDK transaction](./tx-lifecycle.md)

## Synopsis

This document describes `baseapp`, the abstraction that implements most of the common functionalities of an SDK application. 

- [Introduction](#introduction)
- [Type Definition](#type-definition)
- [States and Modes](#states-and-modes)
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
- Different [states and modes](#states-and-modes), as the state-machine can be in different modes depending on the ABCI message it processes. 

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

The [`baseapp` type](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/baseapp.go#L45-L91) holds many important parameters for any Cosmos SDK based application. Let us go through the most important components:

- A [`CommitMultiStore`](./store.md#commit-multi-store). This is the main store of the application, which holds the canonical state that is committed at the [end of each block](#endblock). This store is **not** cached, meaning it is not used to update the application's intermediate (un-committed) states. The `CommitMultiStore` is a multi-store, meaning a store of stores. Each module of the application uses one or multiple `KVStores` in the multi-store to persist their subset of the state. 
- A [database](./store.md#database) `db`, which is used by the `CommitMultiStore` to handle data storage.
- 

## States and Modes

## Routing

## ABCI

## ABCI - CheckTx

## ABCI - DeliverTx

## ABCI - Commit

## Other ABCI Messages

### Info

### SetOption

### Query 

### InitChain

### BeginBlock

### EndBlock

## Gas 