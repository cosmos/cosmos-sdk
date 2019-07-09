# Transactions and Messages

## Prerequisites

* [Anatomy of an SDK Application](./app-anatomy.md)

## Synopsis

This doc describes how transactions are defined.
1. [Transactions](#transactions)
2. [Transaction Definition](#transaction-definition)
3. [Messages](#messages)
4. [Handlers and Keepers](#handlers-and-keepers)
5. [CLI and REST Interfaces](#cli-and-rest-interfaces)

## Transactions

[**Transactions**](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L36-L43)  trigger state changes. They are comprised of metadata and [messages](./modules.md#messages) that instruct modules to perform various actions.

When users interact with an application and make state changes (e.g. sending coins or changing stored values), transactions are created. Transactions must be signed by the appropriate account holders as specified in the transaction definitions, then broadcast to the network. These transactions must then be included in a block, validated, and approved by the network through the consensus process. To learn about the lifecycle of a transaction, click [here](./tx-lifecycle.md).

## Transaction Definition

Transactions are defined by module developers and application developers. They must implement the [`Tx`](https://github.com/cosmos/cosmos-sdk/blob/73700df8c39d1fe6c3d3a1a635ac03d4bacecf55/types/tx_msg.go#L34-L41) interface and include an encoder and decoder.

* **GetMsgs:** unwraps the transaction and returns a list of its message(s) - one transaction may have one or multiple messages.
* **ValidateBasic:** includes lightweight, stateless checks on the validity of a transaction.
* **TxEncoder:** Nodes running the consensus engine (e.g. Tendermint Core) process transactions in the `[]byte` form, so transactions are always encoded before being handled by Tendermint processes. The Cosmos SDK allows developers to specify any deterministic encoding format for their applications; the default is [`Amino`](./amino.md).
* **TxDecoder:** [ABCI](https://tendermint.com/docs/spec/abci/) calls such as `CheckTx` and `DeliverTx` can be used to process transaction data at the application layer; they are passed in as `txBytes []byte` an first unmarshaled (decoded) using `TxDecoder` before any logic is applied.

## Messages

**Messages** describe possible actions within a module. Developers define the specific messages for each application module by implementing the [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface, and also define a [`Handler`](./handlers.md) that performs the specified actions. All the messages in a module are typically defined in a `msgs.go` file, and one handler with multiple functions to handle each of the module's messages is defined in a `handler.go` file.

Note: module messages are not to be confused with [ABCI Messages](https://tendermint.com/docs/spec/abci/abci.html#messages) which define interactions between the Tendermint and application layers.

The [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface has five required functions.

* **Route** returns a string that specifies which module this message is a part of and, thus, which module contains the handler used to implement the actions of this message. Applications that inherit from [baseapp](./baseapp.md) use this function to help deliver transactions by invoking the correct handler(s). For example, `MsgSend` is defined in and handled by the `bank` module; its `Route` function returns the name of the `bank` module.
* **Type** returns a short, human-readable string that describes what the message does. For example, `MsgSend`'s type is `"send"`.
* **ValidateBasic** implements a set of stateless sanity checks for the message and returns an `Error`. For example, the `validateBasic` method for `MsgSend`, the message which sends coins from one account to another, checks that both sender and recipient addresses are valid and the amount is nonnegative.
* **GetSignBytes** returns a `[]byte` representation of the message.
* **GetSigners** returns a list of addresses whose corresponding signatures are required for the message to be valid. For example, `MsgSend` requires a signature from the sender.

## Handlers and Keepers

Each module has a [`Handler`](./app-anatomy.md#handler) to process module messages and a [`Keeper`](./app-anatomy.md#keeper) to manage stores accessed by messages.

## Interfaces

The module developer can specify possible user interactions by defining HTTP Request Handlers or Cobra [commands](https://github.com/spf13/cobra) accessible through command-line, typically found in the module's `/client` folder. An application developer creates entrypoints to the application by creating a [command-line interface](./interfaces.md#cli) or [REST interface](./interfaces.md#rest), typically found in the application's `/cmd` folder.
