# SDK Intro

The [Cosmos-SDK](https://github.com/cosmos/cosmos-sdk) is a framework for building multi-asset public Proof-of-Stake (PoS) blockchains, like the Cosmos Hub, as well as permissionned Proof-Of-Authority (PoA) blockchains.

The goal of the Cosmos SDK is to allow developers to easily create custom  blockchains from scratch that can natively interoperate with other blockchains. We envision the SDK as the npm-like framework to build secure blockchain applications on top of [Tendermint](https://github.com/tendermint/tendermint).

It is based on two major principles:

- **Composability:** Anyone can create a module for the Cosmos-SDK, and integrating the already-built modules is as simple as importing them into your blockchain application.

- **Capabilities:** The SDK is inspired by capabilities-based security, and informed by years of wrestling with blockchain state-machines. Most developers will need to access other 3rd party modules when building their own modules. Given that the Cosmos-SDK is an open framework, some of the modules may be malicious, which means there is a need for security principles to reason about inter-module interactions. These principles are based on object-capabilities. In practice, this means that instead of having each module keep an access control list for other modules, each module implements special objects called keepers that can be passed to other modules to grant a pre-defined set of capabilities. For example, if an instance of module A's keepers is passed to module B, the latter will be able to call a restricted set of module A's functions. The capabilities of each keeper are defined by the module's developer, and it's the developer's job to understand and audit the safety of foreign code from 3rd party modules based on the capabilities they are passing into each third party module. For a deeper look at capabilities, jump to [this section](./ocap.md).

## SDK Application Architecture

### State machine 

At its core, a blockchain is a [replicated deterministic state machine](https://en.wikipedia.org/wiki/State_machine_replication). 

A state machine is a computer science concept whereby a machine can have multiple states, but only one at any given time. There is a `state`, which describes the current state of the system, and `transactions`, that trigger state transitions. 

Given a state S and a transaction T, the state machine will return a new state S'. 

```
+--------+                 +--------+
|        |                 |        |
|   S    +---------------->+   S'   |
|        |    apply(T)     |        |
+--------+                 +--------+
```

In practice, the transactions are bundled in blocks to make the process more efficient. Given a state S and a block of transactions B, the state machine will return a new state S'.

```
+--------+                              +--------+
|        |                              |        |
|   S    +----------------------------> |   S'   |
|        |   For each T in B: apply(T)  |        |
+--------+                              +--------+
```

In a blockchain context, the state machine is deterministic. This means that if you start at a given state and replay the same sequence of transactions, you will always end up with the same final state. 

The Cosmos SDK gives you maximum flexibility to define the state of your application, transaction types and state transition functions. The process of building the state-machine with the SDK will be described more in depth in the following sections. But first, let us see how it is replicated using **Tendermint**. 

### Tendermint

As a developer, you just have to define the state machine using the Cosmos-SDK, and [*Tendermint*](https://tendermint.com/docs/introduction/what-is-tendermint.html) will handle replication over the network for you.


```
                ^  +-------------------------------+  ^
                |  |                               |  |   Built with Cosmos SDK
                |  |  State-machine = Application  |  |
                |  |                               |  v
                |  +-------------------------------+
                |  |                               |  ^
Blockchain node |  |           Consensus           |  |
                |  |                               |  |
                |  +-------------------------------+  |   Tendermint Core
                |  |                               |  |
                |  |           Networking          |  |
                |  |                               |  |
                v  +-------------------------------+  v
```


Tendermint is an application-agnostic engine that is responsible for handling the *networking* and *consensus* layers of your blockchain. In practice, this means that Tendermint is responsible for propagating and ordering transaction bytes. Tendermint Core relies on an eponymous Byzantine-Fault-Tolerant (BFT) algorithm to reach consensus on the order of transactions. For more on Tendermint, click [here](https://tendermint.com/docs/introduction/what-is-tendermint.html).

Tendermint consensus algorithm works with a set of special nodes called *Validators*. Validators are responsible for adding blocks of transactions to the blockchain. At any given block, there is a validator set V. A validator in V is chosen by the algorithm to be the proposer of the next block. This block is considered valid if more than two thirds of V signed a *[prevote](https://tendermint.com/docs/spec/consensus/consensus.html#prevote-step-height-h-round-r)* and a *[precommit](https://tendermint.com/docs/spec/consensus/consensus.html#precommit-step-height-h-round-r)* on it, and if all the transactions that it contains are valid. The validator set can be changed by rules written in the state-machine. For a deeper look at the algorithm, click [here](https://tendermint.com/docs/introduction/what-is-tendermint.html#consensus-overview).


The main part of a Cosmos SDK application is a blockchain daemon that is run by each node in the network locally. If less than one third of the *validator set* is byzantine (i.e. malicious), then each node should obtain the same result when querying the state at the same time. 

## ABCI

Tendermint passes transactions from the network to the application through an interface called the [ABCI](https://github.com/tendermint/tendermint/tree/master/abci), which the application must implement. 

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

Note that **Tendermint only handles transaction bytes**. It has no knowledge of what these bytes mean. All Tendermint does is order these transaction bytes deterministically. Tendermint passes the bytes to the application via the ABCI, and expects a return code to inform it if the messages contained in the transactions were successfully processed or not. 

Here are the most important messages of the ABCI:

- `CheckTx`: When a transaction is received by Tendermint Core, it is passed to the application to check if a few basic requirements are met. `CheckTx` is used to protect the mempool of full-nodes against spam. A special handler called the "Ante Handler" is used to execute a series of validation steps such as checking for sufficient fees and validating the signatures. If the check is valid, the transaction is added to the [mempool](https://tendermint.com/docs/spec/reactors/mempool/functionality.html#mempool-functionality) and relayed to peer nodes. Note that transactions are not processed (i.e. no modification of the state occurs) with `CheckTx` since they have not been included in a block yet. 
- `DeliverTx`: When a [valid block](https://tendermint.com/docs/spec/blockchain/blockchain.html#validation) is received by Tendermint Core, each transaction in the given block is passed to the application via `DeliverTx` to be processed. It is during this stage that the state transitions occur. The "Ante Handler" executes again along with the actual handlers for each message in the transaction.
 - `BeginBlock`/`EndBlock`: These messages are executed at the beginning and the end of each block, whether the block contains transaction or not. It is useful to trigger automatic execution of logic. Proceed with caution though, as computationally expensive loops could slow down your blockchain, or even freeze it if the loop is infinite. 

For a more detailed view of the ABCI methods and types, click [here](https://tendermint.com/docs/spec/abci/abci.html#overview).

Any application built on Tendermint needs to implement the ABCI interface in order to communicate with the underlying local Tendermint engine. Fortunately, you do not have to implement the ABCI interface. The Cosmos SDK provides a boilerplate implementation of it in the form of [baseapp](./sdk-design.md#baseapp).

### Next, let us go into the [high-level design principles of the SDK](./sdk-design.md)