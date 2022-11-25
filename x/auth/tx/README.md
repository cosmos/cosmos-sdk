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

This document specifies the `x/auth/tx` package of the Cosmos SDK.

This package represents the Cosmos SDK implementation of the `client.TxConfig`, `client.TxBuilder`, `client.TxEncoder` and `client.TxDecoder` interfaces.

## Contents

* [Transactions](#transactions)
    * [`TxConfig`](#txconfig)
    * [`TxBuilder`](#txbuilder)
    * [`TxEncoder`/ `TxDecoder`](#txencoder-txdecoder)
* [Client](#client)
    * [CLI](#cli)
        * [Query](#query)
        * [Utils](#utils)
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

The `x/auth/tx` module provides a CLI command to query any transaction, given its hash, transaction sequence or signature.

Without any argument, the command will query the transaction using the transaction hash.

```sh
simd query tx DFE87B78A630C0EFDF76C80CD24C997E252792E0317502AE1A02B9809F0D8685
```

When querying a transaction from an account given its sequence, use the `--type=acc_seq` flag:

```sh
simd query tx --type=acc_seq cosmos1u69uyr6v9qwe6zaaeaqly2h6wnedac0xpxq325/1
```

When querying a transaction given its signature, use the `--type=signature` flag:

```sh
simd q tx --type=signature Ofjvgrqi8twZfqVDmYIhqwRLQjZZ40XbxEamk/veH3gQpRF0hL2PH4ejRaDzAX+2WChnaWNQJQ41ekToIi5Wqw==
```

#### Utils

The `x/auth/tx` module provides a convinient CLI command for decoding and encoding transactions.


```sh
simd tx decode [protobuf-byte-string]
```

```sh
simd tx encode [file-json]
```

Example:

```sh
simd tx bank send $ALICE $BOB 1000stake --generate-only > tx_unsigned.json
simd tx encode tx_unsigned.json > tx_unsigned.bin
simd tx decode $(cat tx_unsigned.bin)
```

### gRPC

A user can query the `x/auth/tx` module using gRPC endpoints.

#### `TxDecode`

#### `TxEncode`

#### `TxEncodeAmino`

#### `TxDecodeAmino`
