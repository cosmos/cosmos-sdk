# Transactions 

## Prerequisites

* [Anatomy of an SDK Application](./app-anatomy.md)

## Synopsis

This document describes how transactions are defined.
1. [Transactions](#transactions)
2. [Transaction Definition](#transaction-definition)
3. [CLI and REST Interfaces](#cli-and-rest-interfaces)
4. [Messages](#messages)
5. [TxBuilder](#txbuilder)
6. [Handlers and Keepers](#handlers-and-keepers)


## Transactions

[**Transactions**](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L36-L43)  are objects that trigger state changes in an application. They are comprised of metadata and [messages](./modules.md#messages) that trigger state changes within a module and are handled by the module's Handler.

When users want to interact with an application and make state changes (e.g. sending coins or changing stored values), they create transactions. Transactions must be signed using the private key associated with the appropriate account(s) as specified in the message definitions, then broadcasted to the network. These transactions must then be included in a block, validated, and approved by the network through the consensus process. To learn about the lifecycle of a transaction, click [here](./tx-lifecycle.md).

## Transaction Definition

Transactions are defined by module developers and application developers. They must implement the [`Tx`](https://github.com/cosmos/cosmos-sdk/blob/73700df8c39d1fe6c3d3a1a635ac03d4bacecf55/types/tx_msg.go#L34-L41) interface and include an encoder and decoder, described here:

* **GetMsgs:** unwraps the transaction and returns a list of its message(s) - one transaction may have one or multiple messages.
* **ValidateBasic:** includes lightweight, [_stateless_](./tx-lifecycle.md#types-of-checks) checks used by ABCI messages [`CheckTx`](./baseapp.md#checktx) and [`DeliverTx`](./baseapp.md#delivertx) to validate transactions. This function is distinct from the `ValidateBasic` functions for messages, which are also for transaction validation but only check messages. For example, when [`runTx`](./baseapp.md#runtx-and-runmsgs) is checking a transaction created from the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/67f6b021180c7ef0bcf25b6597a629aca27766b8/docs/spec/auth) module, it first runs `ValidateBasic` on each message, then runs the `auth` module AnteHandler which calls `ValidateBasic` for the transaction itself.
* **TxEncoder:** Nodes running the consensus engine (e.g. Tendermint Core) are responsible for gossiping transactions and ordering them into blocks, but only handle them in the generic `[]byte` form. Transactions are always marshaled (encoded) before they are relayed to nodes, which compacts them to facilitate gossiping and helps maintain the consensus engine's agnosticism with applications. The Cosmos SDK allows developers to specify any deterministic encoding format for their applications; the default is [`Amino`](./amino.md).
* **TxDecoder:** [ABCI](https://tendermint.com/docs/spec/abci/) calls from the consensus engine to the application, such as `CheckTx` and `DeliverTx`, are used to process transaction data to determine validity and state changes. Since transactions are passed in as `txBytes []byte`, they need to first be unmarshaled (decoded) using `TxDecoder` before any logic is applied.

A transaction is created through one of the possible [interfaces](#interfaces). In the process, an array of [messages](#messages), a [CLIContext](#cli.md) and [`TxBuilder`](#txbuilder) are generated, which are then used to create the transaction itself. The actual state changes triggered by transactions are enabled by the [handlers and keepers](#handlers-and-keepers). The rest of the document will describe each of these components, in this order. 

## CLI and REST Interfaces

The SDK uses several tools for building [interfaces](./interfaces.md) through which users can create transactions. An application developer creates entrypoints to the application by creating a [command-line interface](./interfaces.md#cli) or [REST interface](./interfaces.md#rest), typically found in the application's `/cmd` folder. A module developer specifies possible user interactions by defining HTTP Request Handlers or [commands](https://github.com/spf13/cobra) accessible through command-line, typically found in the module's `/client` folder. When a user interacts with the application's interfaces, they invoke the underlying modules' handlers or command functions, directly creating messages.

## Messages

**Messages** are module-specific objects that trigger state transitions within the scope of the module they belong to. Developers define the messages for each application module by implementing the [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface, and also define a [`Handler`](../building-modules/handler.md) to process them. Messages in a module are typically defined in a `msgs.go` file (though not always), and one handler with multiple functions to handle each of the module's messages is defined in a `handler.go` file.

Note: module messages are not to be confused with [ABCI Messages](https://tendermint.com/docs/spec/abci/abci.html#messages) which define interactions between the Tendermint and application layers.

The [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface has five required functions.

* **Route** returns a string that specifies which module this message is a part of and, thus, which module contains the handler used to implement the actions of this message. Applications that inherit from [baseapp](./baseapp.md) use this function to help deliver transactions by invoking the correct handler(s). For example, `MsgSend` is defined in and handled by the `bank` module; its `Route` function returns the name of the `bank` module.
* **Type** returns a short, human-readable string that describes what the message does. For example, `MsgSend`'s type is `"send"`.
* **ValidateBasic** implements a set of stateless sanity checks for the message and returns an `Error`. For example, the `validateBasic` method for `MsgSend`, the message which sends coins from one account to another, checks that both sender and recipient addresses are valid and the amount is nonnegative.
* **GetSignBytes** returns a `[]byte` representation of the message which is used by the signer to sign it.
* **GetSigners** returns a list of addresses whose corresponding signatures are required for the message to be valid. For example, `MsgSend` requires a signature from the sender of the coins.

While messages contain the information for state transition logic, the transaction's other metadata and relevant information is stored in the `TxBuilder`. 

## TxBuilder

A `TxBuilder` is a transaction [`Context`](https://godoc.org/context), an immutable data structure that can be cloned and updated as the transaction is passed between processes and request handlers. This context is automatically generated when the transaction is created and does not need to be defined by the developer, but does require input from the transaction creator (e.g. using flags through the CLI). The `TxBuilder` contains data such as the `TxEncoder` defined by the developer used to encode messages, the [`KeyBase`](https://github.com/cosmos/cosmos-sdk/blob/85ebf5f72ea21f926b9f371b5dae6af65292c02d/crypto/keys/types.go#L14-L60) used to manage keys, as well as `fees`, `gasPrices`, and other user-provided values. Other user-inputted values and the codec are held in a `CLIContext`. 

Every message in a transaction must be signed by the addresses specified by `GetSigners`. The signing process is handled by the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/67f6b021180c7ef0bcf25b6597a629aca27766b8/docs/spec/auth) module and automatically performed when the transaction is created, unless the user choses to generate and sign separately. The `TxBuilder` contains the `KeyBase` used to manage keys, and thus performs the signing operations. 

## Handlers and Keepers

The final components developers must implement to enable transactions are handlers and keeprs. Each module has a [`Handler`](../building-modules/handler.md) to process module messages and a [`Keeper`](./app-anatomy.md#keeper) to manage stores accessed by messages. Each application has a list of keepers to manage all of the stores needed to encompass the application's state. 
