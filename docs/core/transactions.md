# Transactions

## Prerequisites

* [Anatomy of an SDK Application](./app-anatomy.md)

## Synopsis

This document describes how various components are defined to enable transactions. It also describes how transactions are generated.

1. [Transactions](#transactions)
2. [Transaction Definition](#transaction-definition)
3. [CLI and REST Interfaces](#cli-and-rest-interfaces)
4. [Messages](#messages)
5. [Transaction Generation](#transaction-generation)
6. [Handlers](#handlers)


## Transactions

[**Transactions**](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L36-L43)  are objects that wrap messages and transaction data, created to enable a user to interact with an application. Specifically, they are comprised of metadata held in [contexts](./contexts) and [messages](../modules.md#messages) that trigger state changes within a module and are handled by the module's Handler.

When users want to interact with an application and make state changes (e.g. sending coins), they create transactions. Each of a transaction's messages must be signed using the private key associated with the appropriate account(s), and then the transaction is broadcasted to the network. A transaction must then be included in a block, validated, and approved by the network through the consensus process. To read more about the lifecycle of a transaction, click [here](../basics/tx-lifecycle.md).

## Transaction Definition

Transactions messages are defined by module developers. The transaction objects themselves are SDK types that implement the [`Tx`](https://github.com/cosmos/cosmos-sdk/blob/73700df8c39d1fe6c3d3a1a635ac03d4bacecf55/types/tx_msg.go#L34-L41) interface and include an encoder and decoder:

* **GetMsgs:** unwraps the transaction and returns a list of its message(s) - one transaction may have one or multiple messages.
* **ValidateBasic:** includes lightweight, [*stateless*](../basics/tx-lifecycle.md#types-of-checks) checks used by ABCI messages [`CheckTx`](../basics/baseapp.md#checktx) and [`DeliverTx`](../basics/baseapp.md#delivertx) to validate transactions. The `StdTx` `ValidateBasic` function checks that its transactions are signed by the correct number of signers and that the fees do not exceed what the user's maximum. Note that this function is distinct from the `ValidateBasic` functions for *messages*, which are also for transaction validation but only check messages. For example, when [`runTx`](../basics/baseapp.md#runtx-and-runmsgs) is checking a transaction created from the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/67f6b021180c7ef0bcf25b6597a629aca27766b8/docs/spec/auth) module, it first runs `ValidateBasic` on each message, then runs the `auth` module AnteHandler which calls `ValidateBasic` for the transaction itself.
* **TxEncoder:** Nodes running the consensus engine (e.g. Tendermint Core) are responsible for gossiping transactions and ordering them into blocks, but only handle them in the generic `[]byte` form. Transactions are always marshaled (encoded) before they are relayed to nodes, which compacts them to facilitate gossiping and helps maintain the consensus engine's separation from from application logic. The Cosmos SDK allows developers to specify any deterministic encoding format for their applications; the default is [`Amino`](./amino.md).
* **TxDecoder:** [ABCI](https://tendermint.com/docs/spec/abci/) calls from the consensus engine to the application, such as `CheckTx` and `DeliverTx`, are used to process transaction data to determine validity and state changes. Since transactions are passed in as `txBytes []byte`, they need to first be unmarshaled (decoded) using `TxDecoder` before any logic is applied.

A transaction is created through one of the possible [interfaces](#interfaces). In the process, two contexts and an array of [messages](#messages) are created, which are then used to [generate](#transaction-generation) the transaction itself. The actual state changes triggered by transactions are enabled by the [handlers](#handlers). The rest of the document will describe each of these components, in this order.

## CLI and REST Interfaces

The SDK uses several tools for building [interfaces](./interfaces.md) through which users can create transactions through the command-line or HTTP Requests. Application developers create entrypoints to the application by creating a [command-line interface](./interfaces.md#cli) or [REST interface](./interfaces.md#rest), typically found in the application's `/cmd` folder. These interfaces allow users to interact with the application through command-line or through HTTP requests.

In order for module messages to be utilized in transactions created through these interfaces, module developers must also specify possible user [interactions](../modules/interfaces.md), typically in the module's `/client` folder. For the [command-line interface](../modules/interfaces.md#cli), module developers create subcommands to add as children to the application top-level transaction command `TxCmd`. For [HTTP requests](../modules/interfaces.md#rest), module developers specify acceptable request types, register REST routes, and create HTTP Request Handlers.

When users interact with the application's interfaces, they invoke the underlying modules' handlers or command functions, directly creating messages.

## Messages

**Messages** are module-specific objects that trigger state transitions within the scope of the module they belong to. Module developers define the messages for their module by implementing the [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface, and also define a [`Handler`](../building-modules/handler.md) to process them. Messages in a module are typically defined in a `msgs.go` file (though not always), and one handler with multiple functions to handle each of the module's messages is defined in a `handler.go` file.

Note: module messages are not to be confused with [ABCI Messages](https://tendermint.com/docs/spec/abci/abci.html#messages) which define interactions between the Tendermint and application layers. While ABCI messages such as CheckTx and DeliverTx contain Transactions, which contain module Messages, they are not to be confused with the module level messages themselves

The [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface has five required functions.

* **Route** returns a string that specifies which module this message is a part of and, thus, which module contains the handler used to implement the actions of this message. [Baseapp](./baseapp.md) uses this function to deliver transactions by invoking the correct handler(s). For example, `MsgSend` is defined in and handled by the `bank` module; its `Route` function returns the name of the `bank` module so `Baseapp` understands to use that module's handler.
* **Type** returns a short, human-readable string that describes what the message does. For example, `MsgSend`'s type is `"send"`.
* **ValidateBasic** implements a set of stateless sanity checks for the message and returns an `Error`. For example, the `validateBasic` method for `MsgSend`, the message which sends coins from one account to another, checks that both sender and recipient addresses are valid and the amount is nonnegative.
* **GetSignBytes** returns a `[]byte` representation of the message which is used by the signer to sign it.
* **GetSigners** returns a list of addresses whose corresponding signatures are required for the message to be valid. For example, `MsgSend` requires a signature from the sender of the coins.

While messages contain the information for state transition logic, a transaction's other metadata and relevant information are stored in the `TxBuilder` and `CLIContext`.

## Transaction Generation

[`Contexts`](https://godoc.org/context) are immutable objects that contain all the information needed to process a request. In the process of creating a transaction, two contexts are created: the [`CLIContext`](../interfaces/query-lifecycle.md#clicontext) and `TxBuilder`. Both are automatically generated and do not need to be defined by application developers, but do require input from the transaction creator (e.g. using flags through the CLI).

The [`TxBuilder`](https://github.com/cosmos/cosmos-sdk/blob/73700df8c39d1fe6c3d3a1a635ac03d4bacecf55/x/auth/types/txbuilder.go) contains data closely related with the processing of transactions:

* `TxEncoder` defined by the developer for this type of transaction. Used to encode messages before being processed by nodes running Tendermint.
* `Keybase` that manages the user's keys and is used to perform signing operations.
* `AccountNumber` from which this transaction originated.
* `Sequence`, the number of transactions that the user has sent out, used to prevent replay attacks.
* `Gas` option chosen by the users for how to calculate how much gas they will need to pay. A common option is "auto" which generates an automatic estimate.
* `GasAdjustment` to adjust the estimate of gas by a scalar value, used to avoid underestimating the amount of gas required.
* `SimulateAndExecute` option to simply simulate the transaction execution without broadcasting.
* `ChainID` representing which blockchain this transaction pertains to.
* `Memo` to send with the transaction.
* `Fees`, the maximum amount the user is willing to pay in fees. Alternative to specifying gas prices.
* `GasPrices`, the amount per unit of gas the user is willing to pay in fees. Alternative to specifying fees.

The `CLIContext` is initialized using the application's `codec` and data more closely related to the user interaction with the interface, holding data such as the output to the user and the broadcast mode. Read more about `CLIContext` [here](../interfaces/query-lifecycle.md#clicontext).

Every message in a transaction must be signed by the addresses specified by `GetSigners`. The signing process must be handled by a module, and the most widely used one is the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/67f6b021180c7ef0bcf25b6597a629aca27766b8/docs/spec/auth) module. Signing is automatically performed when the transaction is created, unless the user choses to generate and sign separately. The `TxBuilder` (namely, the `KeyBase`) is used to perform the signing operations, and the `CLIContext` is used to broadcast transactions.

## Handlers

The final components developers must implement to enable transactions are handlers and keeprs. Each module has a Handler to process all of the module's message types. To read more about handlers, visit the documentation for building modules [here](../building-modules/handler.md).
