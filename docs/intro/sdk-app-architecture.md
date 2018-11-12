# SDK Application Architecture

## Parts of a SDK blockchain



## Directory Structure

The SDK is laid out in the following directories:

- `baseapp`: Defines the template for a basic [ABCI](https://cosmos.network/whitepaper#abci) application so that your Cosmos-SDK application can communicate with the underlying Tendermint node.
- `client`: CLI and REST server tooling for interacting with SDK application.
- `examples`: Examples of how to build working standalone applications.
- `server`: The full node server for running an SDK application on top of
  Tendermint.
- `store`: The database of the SDK - a Merkle multistore supporting multiple types of underling Merkle key-value stores.
- `types`: Common types in SDK applications.
- `x`: Extensions to the core, where all messages and handlers are defined.

## BaseApp

Implements an ABCI App using a MultiStore for persistence and a Router to handle transactions.
The goal is to provide a secure interface between the store and the extensible state machine
while defining as little about that state machine as possible (staying true to the ABCI).

BaseApp requires stores to be mounted via capabilities keys - handlers can only access
stores they're given the key for. The BaseApp ensures all stores are properly loaded, cached, and committed.
One mounted store is considered the "main" - it holds the latest block header, from which we can find and load the most recent state.

BaseApp distinguishes between two handler types - the `AnteHandler` and the `MsgHandler`.
The former is a global validity check (checking nonces, sigs and sufficient balances to pay fees,
e.g. things that apply to all transaction from all modules), the later is the full state transition function.
During CheckTx the state transition function is only applied to the checkTxState and should return
before any expensive state transitions are run (this is up to each developer). It also needs to return the estimated
gas cost.

During DeliverTx the state transition function is applied to the blockchain state and the transactions
need to be fully executed.

BaseApp is responsible for managing the context passed into handlers - it makes the block header available and provides the right stores for CheckTx and DeliverTx. BaseApp is completely agnostic to serialization formats.

## Basecoin

Basecoin is the first complete application in the stack. Complete applications require extensions to the core modules of the SDK to actually implement handler functionality.

The native extensions of the SDK, useful for building Cosmos Zones, live under `x`.
Basecoin implements a `BaseApp` state machine using the `x/auth` and `x/bank` extensions,
which define how transaction signers are authenticated and how coins are transferred. It should also use `x/ibc` and probably a simple staking extension.

Basecoin and the native `x` extensions use go-amino for all serialization needs, including for transactions and accounts.



