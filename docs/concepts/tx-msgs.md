# Transactions and Messages

## Prerequisites

* [Baseapp](./baseapp.md)

## Synopsis

This doc describes how transactions and messages are defined. 
1. [Transactions](#transactions)
2. [Messages](#messages)

## Transactions

[**Transactions**](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L36-L43) are comprised of one or multiple [messages](#messages) and trigger state changes. 

TODO: Role of tx

To learn about the lifecycle of a transaction, read [this doc](./tx-lifecycle.md).

### Transaction Definition

TODO: what should be implemented, role of tx
link handler, keeper, antehandler

CLI transaction subcommands are defined in the `tx.go` file. 

### Encoding and Decoding

All nodes running Tendermint only handle transactions in `[]byte` form. Transactions are always encoded (i.e. using [`Amino`](./amino.md), a protocol buffer which serializes application data structures into compact `[]byte`s for efficient transmission) prior to consensus engine processing, and decoded for application processing. Thus, ABCI calls `CheckTx` and `DeliverTx` to an application provide transactions in the form `txBytes []byte` which are first unmarshaled into the application's specific transaction data structure using `app.txDecoder`. 

TODO: specifying encoding

## Messages

**Messages** describe possible actions within a module. Developers define the specific messages for each application module by implementing the [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface, and also define a [`Handler`](./handlers.md) that performs the specified actions. All the messages in a module are typically defined in a `msgs.go` file, and one handler with multiple functions to handle each of the module's messages is defined in a `handler.go` file.

Note: module messages are not to be confused with [ABCI Messages](https://tendermint.com/docs/spec/abci/abci.html#messages) which define interactions between the Tendermint and application layers.

### Message Interface

The [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface has five required functions.

* **Route** returns a string that specifies which module this message is a part of and, thus, which module contains the handler used to implement the actions of this message. Applications that inherit from [baseapp](./baseapp.md) use this function to help deliver transactions by invoking the correct handler(s). For example, `MsgSend` is defined in and handled by the `bank` module; its `Route` function returns the name of the `bank` module.
* **Type** returns a short, human-readable string that describes what the message does. For example, `MsgSend`'s type is `"send"`. 
* **ValidateBasic** implements a set of stateless sanity checks for the message and returns an `Error`. For example, the `validateBasic` method for `MsgSend`, the message which sends coins from one account to another, checks that both sender and recipient addresses are valid and the amount is nonnegative. 
* **GetSignBytes** returns a `[]byte` representation of the message. 
* **GetSigners** returns a list of addresses whose corresponding signatures are required for the message to be valid. For example, `MsgSend` requires a signature from the sender.

