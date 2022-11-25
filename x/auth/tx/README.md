---
sidebar_position: 1
---

# `x/auth/tx`

:::note

### Pre-requisite Readings

* [Transactions](https://docs.cosmos.network/main/core/transactions#transaction-generation)
* [Encoding](https://docs.cosmos.network/main/core/encoding#transaction-encoding)

:::

## Abstract

This document specifies the `x/auth/tx` module of the Cosmos SDK.

`x/auth/tx` is not an usual module that implements the `AppModule` interface with `BeginBlocker` and `EndBlocker` methods.
Instead, it represents the Cosmos SDK implementation of implements the `client.TxConfig`, `client.TxBuilder`, `client.TxEncoder` and `client.TxDecoder` interfaces.

## Contents

* [Transactions](#transactions)
    * [`TxConfig`](#txconfig)
    * [`TxBuilder`](#txbuilder)
    * [`TxEncoder`/ `TxDecoder`](#txencoder-txdecoder)
* [Client](#client)
    * [CLI](#cli)
    * [gRPC](#grpc)
        * [`TxDecode`](#txdecode)
        * [`TxEncode`](#txencode)
        * [`TxEncodeAmino`](#txencodeamino)
        * [`TxDecodeAmino`](#txdecodeamino)

## Transactions

### `TxConfig`

`client.TxConfig` defines an interface a client can utilize to generate an application-defined concrete transaction type.
The interface defines a set of methods for creating a `client.TxBuilder`.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-alpha1/client/tx_config.go#L25-L31
```

The default implementation of `client.TxConfig` is instantiated by `NewTxConfig` in `x/auth/tx` module.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-alpha1/x/auth/tx/config.go#L24-L30
```

### `TxBuilder`

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-alpha1/client/tx_config.go#L33-L50
```

The [`client.TxBuilder`](https://docs.cosmos.network/main/core/transactions#transaction-generation) interface is as well implemented by `x/auth/tx`.
A `client.TxBuilder` can be accessed with `TxConfig.NewTxBuilder()`.  

### `TxEncoder`/ `TxDecoder`

More information about `TxEncoder` and `TxDecoder` can be found [here](https://docs.cosmos.network/main/core/encoding#transaction-encoding).

## Client

### CLI

#### Query

```go
```

### gRPC

A user can query the `x/auth/tx` module using gRPC endpoints.

#### `TxDecode`

#### `TxEncode`

#### `TxEncodeAmino`

#### `TxDecodeAmino`
