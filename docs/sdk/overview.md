# Cosmos SDK Overview

The Cosmos-SDK is a framework for building Tendermint ABCI applications in
Golang. It is designed to allow developers to easily create custom interoperable
blockchain applications within the Cosmos Network.

To achieve its goals of flexibility and security, the SDK makes extensive use of
the [object-capability
model](https://en.wikipedia.org/wiki/Object-capability_model)
and the [principle of least
privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege).

For an introduction to object-capabilities, see this [article](http://habitatchronicles.com/2017/05/what-are-capabilities/).

## Languages

The Cosmos-SDK is currently written in [Golang](https://golang.org/), though the
framework could be implemented similarly in other languages.
Contact us for information about funding an implementation in another language.

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

## Object-Capability Model

When thinking about security, it's good to start with a specific threat
model. Our threat model is the following:

> We assume that a thriving ecosystem of Cosmos-SDK modules that are easy to compose into a blockchain application will contain faulty or malicious modules.

The Cosmos-SDK is designed to address this threat by being the
foundation of an object capability system.

> The structural properties of object capability systems favor
> modularity in code design and ensure reliable encapsulation in
> code implementation.
>
> These structural properties facilitate the analysis of some
> security properties of an object-capability program or operating
> system. Some of these — in particular, information flow properties
> — can be analyzed at the level of object references and
> connectivity, independent of any knowledge or analysis of the code
> that determines the behavior of the objects.
>
> As a consequence, these security properties can be established
> and maintained in the presence of new objects that contain unknown
> and possibly malicious code.
>
> These structural properties stem from the two rules governing
> access to existing objects:
>
> 1.  An object A can send a message to B only if object A holds a
>     reference to B.
> 2.  An object A can obtain a reference to C only
>     if object A receives a message containing a reference to C. As a
>     consequence of these two rules, an object can obtain a reference
>     to another object only through a preexisting chain of references.
>     In short, "Only connectivity begets connectivity."

See the [wikipedia
article](https://en.wikipedia.org/wiki/Object-capability_model) on the Object Capability Model for more
information.

Strictly speaking, Golang does not implement object capabilities
completely, because of several issues:

- pervasive ability to import primitive modules (e.g. "unsafe", "os")
- pervasive ability to override module vars <https://github.com/golang/go/issues/23161>
- data-race vulnerability where 2+ goroutines can create illegal interface values

The first is easy to catch by auditing imports and using a proper
dependency version control system like Dep. The second and third are
unfortunate but it can be audited with some cost.

Perhaps [Go2 will implement the object capability
model](https://github.com/golang/go/issues/23157).

### What does it look like?

Only reveal what is necessary to get the work done.

For example, the following code snippet violates the object capabilities
principle:

```go
type AppAccount struct {...}
var account := &AppAccount{
    Address: pub.Address(),
    Coins: sdk.Coins{{"ATM", 100}},
}
var sumValue := externalModule.ComputeSumValue(account)
```

The method `ComputeSumValue` implies a pure function, yet the implied
capability of accepting a pointer value is the capability to modify that
value. The preferred method signature should take a copy instead.

```go
var sumValue := externalModule.ComputeSumValue(*account)
```

In the Cosmos SDK, you can see the application of this principle in the
basecoin examples folder.

```go
// File: cosmos-sdk/examples/basecoin/app/init_handlers.go
package app

import (
    "github.com/cosmos/cosmos-sdk/x/bank"
    "github.com/cosmos/cosmos-sdk/x/sketchy"
)

func (app *BasecoinApp) initRouterHandlers() {

    // All handlers must be added here.
    // The order matters.
    app.router.AddRoute("bank", bank.NewHandler(app.accountMapper))
    app.router.AddRoute("sketchy", sketchy.NewHandler())
}
```

In the Basecoin example, the sketchy handler isn't provided an account
mapper, which does provide the bank handler with the capability (in
conjunction with the context of a transaction run).

## Application Architecture

The SDK has multiple levels of "application": the ABCI app, the BaseApp, the BasecoinApp, and now your App.

### ABCI App

The basic ABCI interface allowing Tendermint to drive the applications state machine with transaction blocks.

### BaseApp

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

### Basecoin

Basecoin is the first complete application in the stack. Complete applications require extensions to the core modules of the SDK to actually implement handler functionality.

The native extensions of the SDK, useful for building Cosmos Zones, live under `x`.
Basecoin implements a `BaseApp` state machine using the `x/auth` and `x/bank` extensions,
which define how transaction signers are authenticated and how coins are transferred.
It should also use `x/ibc` and probably a simple staking extension.

Basecoin and the native `x` extensions use go-amino for all serialization needs,
including for transactions and accounts.

### Your Cosmos App

Your Cosmos App is a fork of Basecoin - copy the `examples/basecoin` directory and modify it to your needs.
You may want to:

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
