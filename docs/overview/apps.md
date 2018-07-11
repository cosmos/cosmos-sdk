# Apps in the SDK

The SDK has multiple levels of "application": the ABCI app, the BaseApp, the BasecoinApp, and now your App.

## ABCI App

The basic ABCI interface allowing Tendermint to drive the applications state machine with transaction blocks.

## BaseApp

Implements an ABCI App using a MultiStore for persistence and a Router to handle transactions.
The goal is to provide a secure interface between the store and the extensible state machine
while defining as little about that state machine as possible (staying true to the ABCI).

BaseApp requires stores to be mounted via capabilities keys - handlers can only access
stores they're given the key for. The BaseApp ensures all stores are properly loaded, cached, and committed.
One mounted store is considered the "main" - it holds the latest block header, from which we can find and load the 
most recent state ([TODO](https://github.com/cosmos/cosmos-sdk/issues/522)).

BaseApp distinguishes between two handler types - the `AnteHandler` and the `MsgHandler`.
The former is a global validity check (checking nonces, sigs and sufficient balances to pay fees, 
e.g. things that apply to all transaction from all modules), the later is the full state transition function. 
During CheckTx the state transition function is only applied to the checkTxState and should return
before any expensive state transitions are run (this is up to each developer). It also needs to return the estimated
gas cost. 
During DeliverTx the state transition function is applied to the blockchain state and the transactions
need to be fully executed.

BaseApp is responsible for managing the context passed into handlers - 
it makes the block header available and provides the right stores for CheckTx and DeliverTx.

BaseApp is completely agnostic to serialization formats.

## Basecoin

Basecoin is the first complete application in the stack.
Complete applications require extensions to the core modules of the SDK to actually implement handler functionality.
The native extensions of the SDK, useful for building Cosmos Zones, live under `x`.
Basecoin implements a `BaseApp` state machine using the `x/auth` and `x/bank` extensions,
which define how transaction signers are authenticated and how coins are transferred.
It should also use `x/ibc` and probably a simple staking extension.

Basecoin and the native `x` extensions use go-amino for all serialization needs,
including for transactions and accounts.

## Your Cosmos App

Your Cosmos App is a fork of Basecoin - copy the `examples/basecoin` directory and modify it to your needs.
You might want to:

- add fields to accounts 
- copy and modify handlers 
- add new handlers for new transaction types
- add new stores for better isolation across handlers

The Cosmos Hub takes Basecoin and adds more stores and extensions to handle additional
transaction types and logic, like the advanced staking logic and the governance process.

## Ethermint

Ethermint is a new implementation of `BaseApp` that does not depend on Basecoin.
Instead of `cosmos-sdk/x/` it has its own `ethermint/x` based on `go-ethereum`.

Ethermint uses a Patricia store for its accounts, and an IAVL store for IBC.
It has `x/ante`, which is quite similar to Basecoin's but uses RLP instead of go-amino.
Instead of `x/bank`, it has `x/eth`, which defines the single Ethereum transaction type
and all the semantics of the Ethereum state machine.

Within `x/eth`, transactions sent to particular addresses can be handled in unique ways, 
for instance to handle IBC and staking.
