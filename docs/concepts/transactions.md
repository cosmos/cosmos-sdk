# Transactions 

## Prerequisites

* [Anatomy of an SDK Application](./app-anatomy.md)

## Synopsis

This document describes how transactions are defined.
1. [Transactions](#transactions)
2. [Transaction Definition](#transaction-definition)
3. [Messages](#messages)
4. [Handlers and Keepers](#handlers-and-keepers)
5. [CLI and REST Interfaces](#cli-and-rest-interfaces)

## Transactions

[**Transactions**](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L36-L43)  are objects that trigger state changes in an application. They are comprised of metadata and [messages](./modules.md#messages) that trigger state changes within a module and are handled by the module's Handler.

When users want to interact with an application and make state changes (e.g. sending coins or changing stored values), they create transactions. Transactions must be signed using the private key associated with the appropriate account(s) as specified in the message definitions, then broadcasted to the network. These transactions must then be included in a block, validated, and approved by the network through the consensus process. To learn about the lifecycle of a transaction, click [here](./tx-lifecycle.md).

## Transaction Definition

Transactions are defined by module developers and application developers. They must implement the [`Tx`](https://github.com/cosmos/cosmos-sdk/blob/73700df8c39d1fe6c3d3a1a635ac03d4bacecf55/types/tx_msg.go#L34-L41) interface and include an encoder and decoder, described here:

* **GetMsgs:** unwraps the transaction and returns a list of its message(s) - one transaction may have one or multiple messages.
* **ValidateBasic:** includes lightweight, [_stateless_](./tx-lifecycle.md#types-of-checks) used by ABCI messages [`CheckTx`](./baseapp.md#checktx) and [`DeliverTx`](./baseapp.md#delivertx) to validate transactions. This function is distinct from the `ValidateBasic` functions for messages, which are also for transaction validation but only check messages. For example, when [`runTx`](./baseapp.md#runtx-and-runmsgs) is checking a transaction created from the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/67f6b021180c7ef0bcf25b6597a629aca27766b8/docs/spec/auth) module, it first runs `ValidateBasic` on each message, then runs the `auth` module AnteHandler which calls `ValidateBasic` for the transaction itself.
* **TxEncoder:** Nodes running the consensus engine (e.g. Tendermint Core) are responsible for gossiping transactions and ordering them into blocks, but only handle them in the generic `[]byte` form. Transactions are always marshaled (encoded) before they are relayed to nodes, which compacts them to facilitate gossiping and helps maintain the consensus engine's agnosticism with applications. The Cosmos SDK allows developers to specify any deterministic encoding format for their applications; the default is [`Amino`](./amino.md).
* **TxDecoder:** [ABCI](https://tendermint.com/docs/spec/abci/) calls from the consensus engine to the application, such as `CheckTx` and `DeliverTx`, are used to process transaction data to determine validity and state changes. Since transactions are passed in as `txBytes []byte`, they need to first be unmarshaled (decoded) using `TxDecoder` before any logic is applied.

TODO: A transaction is typically [created....](https://github.com/cosmos/cosmos-sdk/blob/1a7f31f7c8de1feded6b7c0df45c71a5f40c61df/x/auth/client/utils/tx.go) 

## Messages

**Messages** are objects created by the end-user in order to trigger state transitions. Developers define the specific messages for each application module by implementing the [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface, and also define a [`Handler`](../building-modules/handler.md) to process them. Messages in a module are typically defined in a `msgs.go` file (though not always), and one handler with multiple functions to handle each of the module's messages is defined in a `handler.go` file.

Note: module messages are not to be confused with [ABCI Messages](https://tendermint.com/docs/spec/abci/abci.html#messages) which define interactions between the Tendermint and application layers.

The [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface has five required functions.

* **Route** returns a string that specifies which module this message is a part of and, thus, which module contains the handler used to implement the actions of this message. Applications that inherit from [baseapp](./baseapp.md) use this function to help deliver transactions by invoking the correct handler(s). For example, `MsgSend` is defined in and handled by the `bank` module; its `Route` function returns the name of the `bank` module.
* **Type** returns a short, human-readable string that describes what the message does. For example, `MsgSend`'s type is `"send"`.
* **ValidateBasic** implements a set of stateless sanity checks for the message and returns an `Error`. For example, the `validateBasic` method for `MsgSend`, the message which sends coins from one account to another, checks that both sender and recipient addresses are valid and the amount is nonnegative.
* **GetSignBytes** returns a `[]byte` representation of the message which is used by the signer to sign it.
* **GetSigners** returns a list of addresses whose corresponding signatures are required for the message to be valid. For example, `MsgSend` requires a signature from the sender of the coins.

## Handlers and Keepers

Each module has a [`Handler`](./app-anatomy.md#handler) to process module messages and a [`Keeper`](./app-anatomy.md#keeper) to manage stores accessed by messages.

## TxBuilder

TODO

## Interfaces

The module developer can specify possible user interactions by defining HTTP Request Handlers or Cobra [commands](https://github.com/spf13/cobra) accessible through command-line, typically found in the module's `/client` folder. An application developer creates entrypoints to the application by creating a [command-line interface](./interfaces.md#cli) or [REST interface](./interfaces.md#rest), typically found in the application's `/cmd` folder.
