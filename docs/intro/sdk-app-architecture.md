# SDK Application Architecture

## Parts of a SDK blockchain

A blockchain application is just a [replicated deterministic state machine](https://en.wikipedia.org/wiki/State_machine_replication). As a developer, you just have to define the state machine (i.e. what the state, a starting state and messages that trigger state transitions), and [*Tendermint*](https://tendermint.com/docs/introduction/introduction.html) will handle replication over the network for you.

>Tendermint is an application-agnostic engine that is responsible for handling the *networking* and *consensus* layers of your blockchain. In practice, this means that Tendermint is reponsible for propagating and ordering transaction bytes. Tendermint Core relies on an eponymous Byzantine-Fault-Tolerant (BFT) algorithm to reach consensus on the order of transactions. For more on Tendermint, click [here](https://tendermint.com/docs/introduction/introduction.html).

Tendermint passes transactions from the network to the application through an interface called the [ABCI](https://github.com/tendermint/tendermint/tree/master/abci). If you look at the architecture of the blockchain node you are building, it looks like the following:

```
+---------------------+
|                     |
|     Application     |
|                     |
+--------+---+--------+
         ^   |
         |   | ABCI
         |   v
+--------+---+--------+
|                     |
|                     |
|     Tendermint      |
|                     |
|                     |
+---------------------+
```

Fortunately, you do not have to implement the ABCI interface. The Cosmos SDK provides a boilerplate implementation of it in the form of [baseapp](#baseapp).

## BaseApp

Implements an ABCI App using a [MultiStore](../reference/store) for persistence and a Router to handle transactions.
The goal is to provide a secure interface between the store and the extensible state machine while defining as little about that state machine as possible (staying true to the ABCI).

For more on `baseapp`, please go to the [baseapp reference](../reference/baseapp.md).

## Modules

The power of the SDK lies in its modularity. SDK blockchains are built out of customizable and interoperable modules. These modules are contained in the `x/` folder.

In addition to the already existing modules in `x/`, that anyone can use in their app, the SDK lets you build your own custom modules. In other words, building a SDK blockchain consists in importing some modules and building others (the ones you need that do not exist yet!).

Some core modules include:

- `x/auth`: Used to manage accounts and signatures.
- `x/bank`: Used to enable tokens and token transfers.
- `x/staking` + `x/slashing`: Used to build Proof-Of-Stake blockchains.

## Basecoin

Basecoin is the first complete application in the stack. Complete applications require extensions to the core modules of the SDK to actually implement handler functionality.

Basecoin implements a `BaseApp` state machine using the `x/auth` and `x/bank` modules, which define how transaction signers are authenticated and how coins are transferred. It should also use `x/ibc` and probably a simple staking extension.

Basecoin and the native `x/` extensions use go-amino for all serialization needs, including for transactions and accounts.

## SDK tutorial

If you want to learn more about how to build an SDK application and get a deepeer understanding of the concepts presented above in the process, please check out the [SDK Application Tutorial](https://github.com/cosmos/sdk-application-tutorial).


