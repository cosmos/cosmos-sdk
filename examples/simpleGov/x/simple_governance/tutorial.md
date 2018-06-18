# Tutorial: How to code a Cosmos-SDK application

In this tutorial, we will learn the basics of coding a Cosmos-SDK application. We will start by an introduction to Tendermint and the Cosmos ecosystem, followed by a high-level overview of the Tendermint software and the Cosmos-SDK framework. After that, we will delve into the code and build a simple application on top of the Cosmos-SDK.

## Tendermint and Cosmos

Blockchains can be divided in three conceptual layers:

- **Networking:** Responsible for propagating transactions.
- **Consensus:** Enables validator nodes to agree on the next set of transactions to process (i.e. add blocks of transactions to the blockchain).
- **Application:** Responsible for updating the state given a set of transactions, i.e. processing transactions

The *networking* layer makes sure that each node receives transactions. The *consensus* layer makes sure that each node agrees on the same transactions to modify their local state. As for the *application* layer, it processes transactions. Given a transaction and a state, the application will return a new state. In Bitcoin for example, the state is a list of balances for each account (in reality, it's a list of UTXO, short for Unspent Transaction Output, but let's call them balances for the sake of simplicity), and transactions modify the state by changing these balances. In the case of Ethereum, the application is a virtual machine. Each transaction goes through this virtual machine and modifies the state according to the specific smart contract that is called within it.

Before Tendermint, building a blockchain required building all three layers from the ground up. It was such a tedious task that most developers preferred forking the Bitcoin codebase, thereby being constrained by the limitations of the Bitcoin protocol. Then, Ethereum came in and greatly simplified the development of decentralised applications by providing a Virtual-Machine blockchain on which anyone could deploy custom logic in the form of Smart Contracts. But it did not simplify the development of blockchains themselves, as Go-Ethereum remained a very monolithic tech stack that is difficult to hard-fork from, much like Bitcoin. That is where Tendermint came in.

The goal of Tendermint is to provide the *networking* and *consensus* layers of a blockchain as a generic engine on which arbitrary applications can be built. With Tendermint, developers only have to worry about the *application* layer of their blockchain, thereby saving them hundreds of hours of development work. Note that Tendermint also designates the name of the byzantine fault tolerant consensus algorithm used within the Tendermint Core engine.

Tendermint connects the blockchain engine (*networking* and *consensus* layers) to the Application via a protocol called the [ABCI](https://github.com/tendermint/abci), short for Application-Blockchain Inteface. Developers only have to implement a few messages to build an ABCI-application that runs on top of the Tendermint engine. ABCI is language agnostic, meaning that developers can build the application part of their blockchain in any programming language. Building on top of Tendermint also provides the following benefits:

- **Public or private blockchain capable.** Since developers can deploy arbitrary applications on top of Tendermint, it is possible to develop both permissioned and permissionless blockchains on top of it.
- **Performance.** Tendermint is a state of the art blockchain engine. Tendermint Core can have a block time on the order of 1 second and handle thousands of transactions per second.
- **Instant finality.** A property of the Tendermint consensus algorithm is instant finality, meaning that forks are never created, as long as less than a third of the validators are malicious (byzantine). Users can be sure their transactions are finalized as soon as a block is created.
- **Security.** Tendermint consensus is not only fault tolerant, it’s optimally Byzantine fault-tolerant, with accountability. If the blockchain forks, there is a way to determine liability.
- **Light-client support**. Tendermint provides built-in light-clients.

But most importantly, Tendermint is natively compatible with the [Inter-Blockchain Communication Protocol](https://github.com/cosmos/cosmos-sdk/tree/develop/docs/spec/ibc) (IBC). This means that any Tendermint blockchain, whether public or private, can be natively connected to the Cosmos ecosystem and securely exchange tokens with other blockchains in the ecosystem. Note that benefiting from interoperability via IBC and Cosmos preserves the sovereignty of your Tendermint chain. Non-Tendermint chains can also be connected to Cosmos via IBC adapters or Peg-Zones, but this is out of scope for this document.

For a more detailed overview of the Cosmos ecosystem, you can read [this article](https://blog.cosmos.network/understanding-the-value-proposition-of-cosmos-ecaef63350d).


## Introduction to the Cosmos-SDK

Developing a Tendermint-based blockchain means that you only have to code the application (i.e. the state-machine). But that in itself can prove to be rather difficult. This is why the Cosmos-SDK exists.

The [Cosmos-SDK](https://github.com/cosmos/cosmos-sdk) is a platform for building multi-asset Proof-of-Stake (PoS) blockchains, like the Cosmos Hub, as well as Proof-Of-Authority (PoA) blockchains.

The goal of the Cosmos-SDK is to allow developers to easily create custom interoperable blockchain applications within the Cosmos Network without having to recreate common blockchain functionality, thus abstracting away the complexity of building a Tendermint ABCI application. We envision the SDK as the npm-like framework to build secure blockchain applications on top of Tendermint.

In terms of its design, the SDK optimizes flexibility and security. The framework is designed around a modular execution stack which allows applications to mix and match elements as desired. In addition, all modules are sandboxed for greater application security.

It is based on two major principles:

- **Composability:** Anyone can create a module for the Cosmos-SDK and integrating the already-built modules is as simple as importing them into your blockchain application.

- **Capabilities:** The SDK is inspired by capabilities-based security, and informed by years of wrestling with blockchain state machines. Most developers will need to access other 3rd party modules when building their own modules. Given that the Cosmos-SDK is an open framework and that we assume that some those modules may be malicious, we designed the SDK using object-capabilities (ocaps) based principles. In practice, this means that instead of having each module keep an access control list for other modules, each module implements keepers that can be passed to other modules to grant a pre-defined set of capabilities. For example, if an instance of module A's keepers is passed to module B, the latter will be able to call a restricted set of module A's functions. The capabilities of each keeper are defined by the module's developer, and it's their job to understand and audit the safety of foreign code from 3rd party modules based on the capabilities they are passing into each 3rd party module. For a deeper look at capabilities, you can read this article.

*Note: For now the Cosmos-SDK only exists in Golang, which means that developers can only develop SDK modules in Golang. In the future, we expect that the SDK will be implemented in other programming languages. Funding opportunities supported by the Tendermint team may be available eventually.*

## Reminder on Tendermint and ABCI

Cosmos-SDK is a framework to develop the *application* layer of the blockchain. This application can be plugged on any consensus engine (*consensus* + *networking* layers) that supports a simple protocol called the ABCI, short for [Application-Blockchain Interface](https://github.com/tendermint/abci).

Tendermint is the default consensus engine on which the Cosmos-SDK is built. It is important to have a good understanding of the respective responsibilities of both the *Application* and the *Consensus Engine*.

Responsibilities of the *Consensus Engine*:
- Propagate transactions
- Agree on the order of valid transactions

Reponsibilities of the *Application*:
- Generate Transactions
- Check if transactions are valid
- Process Transactions (includes state changes)

It is worth underlining that the *Consensus Engine* has knowledge of a given validator set for each block, but that it is the responsiblity of the *Application* to trigger validator set changes. This is the reason why it is possible to build both **public and private chains** with the Cosmos-SDK and Tendermint. A chain will be public or private depending on the rules, defined at application level, that govern validator set changes.

The ABCI establishes the connection between the *Consensus Engine* and the *Application*. Essentially, it boils down to two messages:

- `CheckTx`: Ask the application if the transaction is valid. When a node receives a transaction, it will run `CheckTx` on it. If the transaction is valid, it is added to the mempool.
- `DeliverTx`: Ask the application to process the transaction. Returns a new state.

Let us give a high-level overview of  how the *Consensus Engine* and the *Application* interract with each other.

- At all times, when the consensus engine of a validator node receives a transaction, it passes it to the application via `CheckTx` to check its validity. If it is valid, the transaction is added to the mempool.
- Let us say we are at block N. There is a validator set V. A proposer is selected from V by the *Consensus Engine* to propose the next block. The proposer gathers valid transaction from its mempool and forms a block. Then, the block is gossiped to other validators to be signed. The block becomes block N+1 once 2/3+ of V have signed a *precommit* on it (For a more detailed explanation of the consensus algorithm, click [here](https://github.com/tendermint/tendermint/wiki/Byzantine-Consensus-Algorithm)).
- When block N+1 is signed by 2/3+ of V, it is gossipped to full-nodes. When full-nodes receive the block, they confirm its validity. A block is valid if it it holds valid signatures from more than 2/3 of V and if all the transactions in the block are valid. To check the validity of transactions, the *Consensus Engine* transfers them to the application via `DeliverTx`. After each transaction, `DeliverTx` returns a new state if the transaction was valid. At the end of the block, a final state is committed. Of course, this means that the order of transaction within a block matters.

## Architecture of a SDK-app

The Cosmos-SDK gives the basic template for an application architecture. You can find this template [here](https://github.com/cosmos/cosmos-sdk).

In essence, a blockchain application is simply a replicated state machine. There is a state (e.g. for a cryptocurrency, how many coins each account holds), and transactions that trigger state transitions. As the application developer, your job is just define the state, the transactions types and how different transactions modify the state.

### Modularity

The Cosmos-SDK is a module-based framework. Each module is in itself a little state-machine that can be gracefully combined with other modules to produce a coherent application. In other words, modules define a sub-section of the global state and of the transaction types. Then, it is the job of the root application to route transactions to the correct modules depending on their respective types. To understand this process, let us take a look at a simplified standard cycle of the state-machine.

Upon receiving a transaction from the Tendermint Core engine, here is what the *Application* does:

1. Decode the transaction and get the message
2. Route the message to the appropriate module using the `Msg.Type()` method
3. Run the transaction in the module. Modify the state if the transaction is valid.
4. Return new state or error message

Steps 1, 2 and 4 are handled by the root application. Step 3 is handled by the appropriate module.

### SDK Components

With this in mind, let us go through the important directories of the SDK:

- `baseapp`: This defines the template for a basic application. Basically it implements the ABCI protocol so that your Cosmos-SDK application can communicate with the underlying Tendermint node.
- `client`: Command-Line Interface to interact with the application
- `server`: REST server to communicate with the node
- `examples`: Contains example on how to build a working application based on `baseapp` and modules
- `store`: Contains code for the multistore. The multistore is basically your state. Each module can create any number of KVStores from the multistore. Be careful to properly handle access rights to each store with appropriate `keepers`.
- `types`: Common types required in any SDK-based application.
- `x`: This is where modules live. You will find all the already-built modules in this directory. To use any of these modules, you just need to properly import them in your application. We will see how in the [App - Bridging it all together] section.

### Introductory Coderun

#### KVStore

The KVStore provides the basic persistence layer for your SDK application.

```go
type KVStore interface {
    Store

    // Get returns nil iff key doesn't exist. Panics on nil key.
    Get(key []byte) []byte

    // Has checks if a key exists. Panics on nil key.
    Has(key []byte) bool

    // Set sets the key. Panics on nil key.
    Set(key, value []byte)

    // Delete deletes the key. Panics on nil key.
    Delete(key []byte)

    // Iterator over a domain of keys in ascending order. End is exclusive.
    // Start must be less than end, or the Iterator is invalid.
    // CONTRACT: No writes may happen within a domain while an iterator exists over it.
    Iterator(start, end []byte) Iterator

    // Iterator over a domain of keys in descending order. End is exclusive.
    // Start must be greater than end, or the Iterator is invalid.
    // CONTRACT: No writes may happen within a domain while an iterator exists over it.
    ReverseIterator(start, end []byte) Iterator

    // TODO Not yet implemented.
    // CreateSubKVStore(key *storeKey) (KVStore, error)

    // TODO Not yet implemented.
    // GetSubKVStore(key *storeKey) KVStore
 }
```

You can mount multiple KVStores onto your application, e.g. one for staking, one for accounts, one for IBC, and so on.

```go
 app.MountStoresIAVL(app.keyMain, app.keyAccount, app.keyIBC, app.keyStake, app.keySlashing)
```

The implementation of a KVStore is responsible for providing any Merkle proofs for each query, if requested.

```go
 func (st *iavlStore) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
```

Stores can be cache-wrapped to provide transactions at the persistence level (and this is well supported for iterators as well). This feature is used to provide a layer of transactional isolation for transaction processing after the "AnteHandler" deducts any associated fees for the transaction. Cache-wrapping can also be useful when implementing a virtual-machine or scripting environment for the blockchain.

#### go-amino

The Cosmos-SDK uses [go-amino](https://github.com/cosmos/cosmos-sdk/blob/96451b55fff107511a65bf930b81fb12bed133a1/examples/basecoin/app/app.go#L97-L111) extensively to serialize and deserialize Go types into Protobuf3 compatible bytes.

Go-amino (e.g. over https://github.com/golang/protobuf) uses reflection to encode/decode any Go object.  This lets the SDK developer focus on defining data structures in Go without the need to maintain a separate schema for Proto3. In addition, Amino extends Proto3 with native support for interfaces and concrete types.

For example, the Cosmos SDK's `x/auth` package imports the PubKey interface from `tendermint/go-crypto` , where PubKey implementations include those for _Ed25519_ and _Secp256k1_.  Each `auth.BaseAccount` has a PubKey.

```go
 // BaseAccount - base account structure.
 // Extend this by embedding this in your AppAccount.
 // See the examples/basecoin/types/account.go for an example.
 type BaseAccount struct {
    Address  sdk.Address   `json:"address"`
    Coins    sdk.Coins     `json:"coins"`
    PubKey   crypto.PubKey `json:"public_key"`
    Sequence int64         `json:"sequence"`
 }
```

Amino knows what concrete type to decode for each interface value based on what concretes are registered for the interface.

For example, the `Basecoin` example app knows about _Ed25519_ and _Secp256k1_ keys because they are registered by the app's `codec` below:

```go
wire.RegisterCrypto(cdc) // Register crypto.
```

For more information on Go-Amino, see https://github.com/tendermint/go-amino.

#### Keys, Keepers, and Mappers

The Cosmos SDK is designed to enable an ecosystem of libraries that can be imported together to form a whole application. To make this ecosystem more secure, we've developed a design pattern that follows the principle of least-authority.

`Mappers` and `Keepers` provide access to KV stores via the context. The only difference between the two is that a `Mapper` provides a lower-level API, so generally a `Keeper` might hold references to other Keepers and `Mappers` but not vice versa.

`Mappers` and `Keepers` don't hold any references to any stores directly.  They only hold a _key_ (the `sdk.StoreKey` below):

```go
type AccountMapper struct {

    // The (unexposed) key used to access the store from the Context.
    key sdk.StoreKey

    // The prototypical Account concrete type.
    proto Account

    // The wire codec for binary encoding/decoding of accounts.
    cdc *wire.Codec
 }
```

This way, you can hook everything up in your main `app.go` file and see what components have access to what stores and other components.

```go
// Define the accountMapper.
 app.accountMapper = auth.NewAccountMapper(
    cdc,
    app.keyAccount,      // target store
    &types.AppAccount{}, // prototype
 )
```

Later during the execution of a transaction (e.g. via ABCI `DeliverTx` after a block commit) the context is passed in as the first argument.  The context includes references to any relevant KV stores, but you can only access them if you hold the associated key.

```go
 // Implements sdk.AccountMapper.
 func (am AccountMapper) GetAccount(ctx sdk.Context, addr sdk.Address) Account {
    store := ctx.KVStore(am.key)
    bz := store.Get(addr)
    if bz == nil {
        return nil
    }
    acc := am.decodeAccount(bz)
    return acc
 }
```

`Mappers` and `Keepers` cannot hold direct references to stores because the store is not known at app initialization time.  The store is dynamically created (and wrapped via `CacheKVStore` as needed to provide a transactional context) for every committed transaction (via ABCI `DeliverTx`) and mempool check transaction (via ABCI `CheckTx`).

#### Tx, Msg, Handler, and AnteHandler

A transaction (`Tx` interface) is a signed/authenticated message (`Msg` interface).

Transactions that are discovered by the Tendermint mempool are processed by the `AnteHandler` (_ante_ just means before) where the validity of the transaction is checked and any fees are collected.

Transactions that get committed in a block first get processed through the `AnteHandler`, and if the transaction is valid after fees are deducted, they are processed through the appropriate Handler.

In either case, the transaction bytes must first be parsed. The default transaction parser uses Amino. Most SDK developers will want to use the standard transaction structure defined in the `x/auth` package (and the corresponding `AnteHandler` implementation also provided in `x/auth`):

```go
 // StdTx is a standard way to wrap a Msg with Fee and Signatures.
 // NOTE: the first signature is the FeePayer (Signatures must not be nil).
 type StdTx struct {
    Msg        sdk.Msg        `json:"msg"`
    Fee        StdFee         `json:"fee"`
    Signatures []StdSignature `json:"signatures"`
 }
```

Various packages generally define their own message types.  The `Basecoin` example app includes multiple message types that are registered in `app.go`:

```go
sdk.RegisterWire(cdc)    // Register Msgs
 bank.RegisterWire(cdc)
 stake.RegisterWire(cdc)
 slashing.RegisterWire(cdc)
 ibc.RegisterWire(cdc)
```

Finally, handlers are added to the router in your `app.go` file to map messages to their corresponding handlers. (In the future we will provide more routing features to enable pattern matching for more flexibility).

```go
 // register message routes
 app.Router().
    AddRoute("auth", auth.NewHandler(app.accountMapper)).
    AddRoute("bank", bank.NewHandler(app.coinKeeper)).
    AddRoute("ibc", ibc.NewHandler(app.ibcMapper, app.coinKeeper)).
    AddRoute("stake", stake.NewHandler(app.stakeKeeper))
```

#### EndBlocker

The `EndBlocker` hook allows us to register callback logic to be performed at the end of each block.  This lets us process background events, such as processing validator inflationary Atom provisions:

```go
// Process Validator Provisions
 blockTime := ctx.BlockHeader().Time // XXX assuming in seconds, confirm
 if pool.InflationLastTime+blockTime >= 3600 {
    pool.InflationLastTime = blockTime
    pool = k.processProvisions(ctx)
 }
```

By the way, the SDK provides a [staking module](https://github.com/cosmos/cosmos-sdk/tree/develop/x/stake), which provides all the bonding/unbonding funcionality for the Cosmos Hub.

#### Start working

To get started, you just have to follow these simple steps:

1. Clone the [Cosmos-SDK](https://github.com/cosmos/cosmos-sdk/tree/develop) repo
2. Code the modules needed by your application that do not already exist.
3. Create your app directory. In the app main file, import the module you need and instantiate the different stores.
4. Launch your blockchain.

Easy as pie! With the introduction over, let us delve into practice and learn how to code a SDK application with an example.

## Setup

### Prerequisites

- Have [go](https://golang.org/dl/) and [git](https://git-scm.com/downloads) installed
- Don't forget to set your `PATH` and `GOPATH`

### Setup work environment

Go to the [Cosmos-SDK repo](https://githum.com/cosmos/cosmos-sdk) and fork it. Then open a terminal and:

```bash
cd $GOPATH/src/github.com/your_username
git clone github.com/your_username/cosmos-sdk
cd cosmos-sdk
```

Now we'll add the origin Cosmos-SDK as upstream in case some cool feature or module gets merged:

```bash
git remote add upstream github.com/cosmos/cosmos-sdk
git fetch upstream
git rebase upstream/master
```

We will also create a branch dedicated to our module:

```bash
git checkout -b my_new_application
```

We are all set!

## Designing the app

### Simple governance application

For this tutorial, we will code a **simple governance application**, accompagnied by a **simple governance module**. It will allow us to explain most of the basic notions required to build a functioning application. Note that this is not the governance module used for the Cosmos Hub. A much more [advanced governance module](https://github.com/cosmos/cosmos-sdk/tree/develop/x/gov) will be used instead.

All the code for the `simple_governance` application can be found [here](https://github.com/gamarin2/cosmos-sdk/tree/module_tutorial/examples/simpleGov/x/simple_governance). You'll notice that the module and app aren't located at the root level of the repo but in the examples directory. This is just for convenience, you can code your module and application directly in the root directory.

Without further talk, let's get into it!

### Requirements

We will start by writting down our module's requirements. We are designing a simple governance module, in which we want:

- Simple text proposals, that any coin holder can submit.
- Proposals must be submitted with a deposit in Atoms. If the deposit is superior to a predefined value called `MinDeposit`, the proposal enters the voting period. Otherwise it is rejected.
- Bonded Atom holders can vote on proposal on a 1 bonded Atom 1 vote basis
- Bonded Atom holders can choose between 3 options when casting a vote: `Yes`, `No` and `Abstain`.
- If, at the end of the voting period, there are more `Yes` votes than `No` votes, the proposal is accepted. Otherwise, it is rejected.
- Voting period is 2 weeks

When designing a module, it is good to adopt a certain methodology. Remember that a blockchain application is just a replicated state-machine. The state is the representation of the application at a given time. It is up to the application developer to define what the state represents, depending on the goal of the application. For example, the state of a simple cryptocurrency application will be a mapping of addresses to balances.

The state can be updated according to predefined rules. Given a state and a transaction, the state-machine (i.e. the application) will return a new state. In a blockchain application, transactions are bundled in blocks, but the logic is the same. Given a state and a set of transactions (a block), the application returns a new state. A SDK-module is just a subset of the application, but it is based on the same principles. As a result, module developers only have to define a subset of the state and a subset of the transaction types, which trigger state transitions.

In summary, we have to define:

- A `State`, which represents a subset of the current state of the application.
- `Transactions`, which contain messages that trigger state transitions.

### State

Here, we will define the types we need (excluding transaction types), as well as the stores in the multistore.

Our voting module is very simple, we only need a single type: `Proposal`. `Proposals` are item to be voted upon. They can be submitted by any user. A deposit has to be provided.

```go
type Proposal struct {
    Title           string          // Title of the proposal
    Description     string          // Description of the proposal
    Submitter       sdk.Address     // Address of the submitter. Needed to refund deposit if proposal is accepted.
    SubmitBlock     int64           // Block at which proposal is submitted. Also the block at which voting period begins.
    State           string          // State can be either "Open", "Accepted" or "Rejected"

    YesVotes        int64           // Total number of Yes votes
    NoVotes         int64           // Total number of No votes
    AbstainVotes    int64           // Total number of Abstain votes
}
```

In terms of store, we will just create one KVStore in the multistore to store `Proposals`. We will also store the vote (`Yes`, `No` or `Abstain`) chosen by each voter on each proposal.

- `Proposals` will be indexed by `'proposals'|<proposalID>`.
- `Votes` (`Yes`, `No`, `Abstain`) will be indexed by `'proposals'|<proposalID>|'votes'|<voterAddress>`.

Notice the quote mark on `'proposals'` and `'votes'`. They indicate that these are constant keywords. So, for example, the option casted by voter with address `0x01` on proposal `0101` will be stored at index `'proposals'|0101|'votes'|0x01`.

These keywords are used to faciliate range queries. Range queries (TODO: Link to formal spec) allow developer to query a subspace of the store, and return an iterator. They are made possible by the nice properties of the [IAVL+ tree](https://github.com/tendermint/iavl) that is used in the background. In practice, this means that it is possible to store and query a Key-Value pair in O(1), while still being able to iterate over a given subspace of Key-Value pairs. For example, we can query all the addresses that voted on a given proposal, along with their votes, by calling `rangeQuery(SimpleGovStore, <proposalID|'addresses'>)`.

### Transactions

The title of this section is a bit misleading. Indeed, what you, as a module developer, have to define are not `Transactions`, but `Messages`. Both transactions and messages exist in the Cosmos-SDK, but a transaction differs from a message in that a message is contained in a transaction. Transactions wrap around messages and add standard information like signatures and fees. As a module developer, you do not have to worry about transactions, only messages.

Let us define the messages we need in order to modify the state. Based on our features, we only need two messages:

- `SubmitProposalMsg`: to submit proposals
- `VoteMsg`: to vote on proposals

```go
type SubmitProposalMsg struct {
    Title           string      // Title of the proposal
    Description     string      // Description of the proposal
    Deposit         sdk.Coins   // Deposit paid by submitter. Must be > MinDeposit to enter voting period
    Submitter       sdk.Address // Address of the submitter
}
```

```go
type VoteMsg struct {
    ProposalID  int64           // ID of the proposal
    Option      string          // Option chosen by voter
    Voter       sdk.Address     // Address of the voter
}
```

## Implementation

Now that we have our types defined, we can start actually implementing the module.

First, let us go into the module's folder and create a folder for our module.

```bash
cd x/
mkdir simple_governance
cd simple_governance
```

Let us start by adding the files we will need. Your module's folder should look something like that:

```
x
└─── simple_governance
      ├─── client
      │     ├───  cli
      │     │     └─── simple_governance.go
      │     └─── rest
      │           └─── simple_governance.go
      ├─── errors.go
      ├─── handler_test.go
      ├─── handler.go
      ├─── keeper_keys.go
      ├─── keeper_test.go
      ├─── keeper.go
      ├─── test_common.go
      ├─── types_test.go
      ├─── types.go
      └─── wire.go
```

Let us go into the detail of each of these files.

### Types

In this file, we define the custom types for our module. This includes the types from the [State](#State) section and the custom message types defined in the [Transactions](#Transactions) section.

For each new type that is not a message, it is possible to add methods that make sense in the context of the application. In our case, we will implement an `updateTally` function to easily update the tally of a given proposal as vote messages come in.

Messages are a bit different. They implement the `Message` interface defined in the SDK's `types` folder. Here are the methods you need to implement when you define a custom message type:

- `Type()`: This function returns the name of our module's route. When messages are processed by the application, they are routed using the string returned by the `Type()` method.
- `GetSignBytes()`: Returns the byte representation of the message. It is used to sign the message.
- `GetSigners()`: Returns address(es) of the signer(s).
- `ValidateBasic()`: This function is used to discard obviously invalid messages. It is called at the beginning of `runTx()` in the baseapp file. If `ValidateBasic()` does not return `nil`, the app stops running the transaction.
- `Get()`: A basic getter, returns some property of the message.
- `String()`: Returns a human-readable version of the message

For our simple governance messages, this means:

- `Type()` will return `"simpleGov"`
- For `SubmitProposalMsg`, we need to make sure that the attributes are not empty and that the deposit is both valid and positive. Note that this is only basic validation, we will therefore not check in this method that the sender has sufficient funds to pay for the deposit
- For `VoteMsg`, we check that the address and option are valid and that the proposalID is not negative.
- As for other methods, less customization is required. You can check the code to see a standard way of implementing these.

### Keeper

#### Short intro to keepers

`Keepers` handles read and writes for modules' stores. This is where the notion of capability enters into play.

As module developers, we have to define keepers to interact with our module's store(s) not only from within our module, but also from other modules. When another module wants to access one of our module's store(s), a keeper for this store has to be passed to it at the application level. In practice, it will look like that:

```go
// in app.go

// instanciate keepers
keeperA = moduleA.newKeeper(app.moduleAStoreKey)
keeperB = moduleB.newKeeper(app.moduleBStoreKey)

// pass instance of keeperA to handler of module B
app.Router().
        AddRoute("moduleA", moduleA.NewHandler(keeperA)).
        AddRoute("moduleB", moduleB.NewHandler(keeperB, keeperA))   // Here module B can access one of module A's store via the keeperA instance
```

`KeeperA` grants a set of capabilities to the handler of module B. When developing a module, it is good practice to think about the sensitivity of the different capabilities that can be granted through keepers. For example, some module may need to read and write to module A's main store, while others only need to read it. If a module has multiple stores, then some keepers could grant access to all of them, while others would only grant access to specific sub-stores. It is the job of the module developer to make sure it is easy for  application developers to instanciate a keeper with the right capabilities. Of course, the handler of a module will most likely get an unrestricted instance of that module's keeper.

#### Keepers for our app

In our case, we only have one store to access, the `SimpleGov` store. We will need to set and get values inside this store via our keeper. However, these two actions do not have the same impact in terms of security. While there should no problem in granting read access to our store to other modules, write access is way more sensitive. So ideally application developers should be able to create either a governance mapper that can only get values from the store, or one that can both get and set values. To this end, we will introduce two keepers: `Keeper` and `KeeperRead`. When application developers create their application, they will be able to decide which of our module's keeper to use.

Now let us try to think about which keeper from **external** modules our module's keepers need access to.
Each proposal requires a deposit. This means our module needs to be able to both read and write to the module that handles tokens, which is the `bank` module. We also need to be able to determine the voting power of each voter based on their stake. To this end, we need read access to the store of the `staking` module. However, we don't need write access to this store. We should therefore indicate that in our module, and the application developer should be careful to only pass a read-only keeper of the `staking` module to our module's handler.

With all that in mind, we can define the structure of our `Keeper`:

```go
    type Keeper struct {
        SimpleGov    sdk.StoreKey        // Key to our module's store
        cdc                 *wire.Codec         // Codec to encore/decode structs
        ck                  bank.Keeper         // Needed to handle deposits. This module onlyl requires read/writes to Atom balance
        sm                  stake.Keeper        // Needed to compute voting power. This module only needs read access to the staking store.
        codespace           sdk.CodespaceType   // Reserves space for error codes
    }
```

And the structure of our `KeeperRead`:

```go
type KeeperRead struct {
    Keeper
}
```

`KeeperRead` will inherit all methods from `Keeper`, except those that we override. These will be the methods that perform writes to the store.

#### Functions and Methods

The first function we have to create is the constructor.

```go
func NewKeeper(SimpleGov sdk.StoreKey, ck bank.Keeper, sm stake.Keeper, codespace sdk.CodespaceType) Keeper
```

This function is called from the main `app.go` file to instanciate a new `Keeper`. A similar function exits for `KeeperRead`.

```go
func NewKeeperRead(SimpleGov sdk.StoreKey, ck bank.Keeper, sm stake.Keeper, codespace sdk.CodespaceType) KeeperRead
```

Depending on the needs of the application and its modules, either `Keeper`, `KeeperRead`, or both, will be instanciated at application level.

*Note: Both the `Keeper` type name and `NewKeeper()` function's name are standard names used in every module. It is no requirement to follow this standard, but doing so can facilitate the life of application developers*

Now, let us describe the methods we need for our module's `Keeper`. For the full implementation, please refer to `keeper.go`.

- `GetProposal`: Get a `Proposal` given a `proposalID`. Proposals need to be decoded from `byte` before they can be read.
- `SetProposal`: Set a `Proposal` at index `'proposals'|<proposalID>`. Proposals need to be encoded to `byte` before they can be stored.
- `NewProposalID`: A function to generate a new unique `proposalID`.
- `GetVote`: Get a vote `Option` given a `proposalID` and a `voterAddress`.
- `SetVote`: Set a vote `Option` given a `proposalID` and a `voterAddress`.
- Proposal Queue methods: These methods implement a standard proposal queue to store `Proposals` on a First-In First-Out basis. It is used to tally the votes at the end of the voting period.

The last thing that needs to be done is to override certain methods for the `KeeperRead` type. `KeeperRead` should not have write access to the stores. Therefore, we will override the methods `SetProposal()`, `SetVote()` and `NewProposalID()`, as well as `setProposalQueue()` from the Proposal Queue's methods. For `KeeperRead`, these methods will just throw an error.

*Note: If you look at the code, you'll notice that the context `ctx` is a parameter of many of the methods. The context `ctx` provides useful information on the current state such as the current block height and allows the keeper `k` to access the `KVStore`. You can check all the methods of `ctx` [here](https://github.com/cosmos/cosmos-sdk/blob/develop/types/context.go#L144-L168)*.

### Handler

#### Constructor and core handlers

Handlers implement the core logic of the state-machine. When a transaction is routed from the app to the module, it is run by the `handler` function.

In practice, one `handler` will be implemented for each message of the module. In our case, we have two message types. We will therefore need two `handler` functions. We will also need a constructor function to route the message to the correct `handler`:

```go
func NewHandler(k Keeper) sdk.Handler {
    return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
        switch msg := msg.(type) {
        case SubmitProposalMsg:
            return handleSubmitProposalMsg(ctx, k, msg)
        case VoteMsg:
            return handleVoteMsg(ctx, k, msg)
        default:
            errMsg := "Unrecognized gov Msg type: " + reflect.TypeOf(msg).Name()
            return sdk.ErrUnknownRequest(errMsg).Result()
        }
    }
}
```

The messages are routed to the appropriate `handler` depending on their type. For our simple governance module, we only have two `handlers`, that correspond to our two message types. They have similar signatures:

```go
func handleSubmitProposalMsg(ctx sdk.Context, k Keeper, msg SubmitProposalMsg) sdk.Result
```

Let us take a look at the parameters of this function:

- The context `ctx` to access the stores.
- The keeper `k` allows the handler to read and write from the different stores, including the module's store (`SimpleGovernance` in our case) and all the stores from other modules that the keeper `k` has been granted an access to (`stake` and `bank` in our case).
- The message `msg` that holds all the information provided by the sender of the transaction.

The function returns a `Result` that is returned to the application. It contains several useful information such as the amount of `Gas` for this transaction and wether the message was succesfully processed or not. At this point, we exit the boundaries of our simple governance module and go back to root application level. The `Result` will differ from application to application. You can check the `sdk.Result` type directly [here](https://github.com/cosmos/cosmos-sdk/blob/develop/types/result.go) for more info.

#### BeginBlocker and EndBlocker

Contrary to most Smart-Contracts platform, it is possible to perform automatic (i.e. not triggered by a transaction sent by an end-user) execution of logic in Cosmos-SDK applications.

This automatic execution of code takes place in the `BeginBlock` and `EndBlock` functions that are called at the beginning and at the end of every block. They are powerful tools, but it is important for application developers to be careful with them. For example, it is crutial that developers control the amount of computing that happens in these functions, as expensive computation could delay the block time, and never-ending loop freeze the chain altogether.

`BeginBlock` and `EndBlock` are composable functions, meaning that each module can implement its own `BeginBlock` and `EndBlock` logic. When needed, `BeginBlock` and `EndBlock` logic is implemented in the module's `handler`. Here is the standard way to proceed for `EndBlock` (`BeginBlock` follows the exact same pattern):

```go
func NewEndBlocker(k Keeper) sdk.EndBlocker {
    return func(ctx sdk.Context, req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
        err := checkProposal(ctx, k)
        if err != nil {
            panic(err)
        }
        return
    }
}
```

Do not forget that each module need to declare its `BeginBlock` and `EndBlock` constructors at application level. See the [Application - Bridging it all together](#application_-_bridging_it_all_together).

For the purpose of our simple governance application, we will use `EndBlock` to automatically tally the results of the vote. Here are the different steps that will be performed:

1. Get the oldest proposal from the `ProposalProcessingQueue`
2. Check if the `CurrentBlock` is the block at which the voting period for this proposal ends. If Yes, go to 3.. If no, exit.
3. Check if proposal is accepted or rejected. Update the proposal status.
4. Pop the proposal from the `ProposalProcessingQueue` and go back to 1.

Let us perform a quick safety analysis on this process.
- The loop will not run forever because the number of proposals in `ProposalProcessingQueue` is finite
- The computation should not be too expensive because tallying of individual proposals is not expensive and the number of proposals is expected be relatively low. That is because proposals require a `Deposit` to be accepted. `MinDeposit` should be high enough so that we don't have too many `Proposals` in the queue.
- In the eventuality that the application becomes so successful that the `ProposalProcessingQueue` ends up containing so many proposals that the blockchain starts slowing down, the module should be modified to mitigate the situation. One clever way of doing it is to cap the number of iteration per individual `EndBlock` at `MaxIteration`. This way, tallying will be spread over many blocks if the number of proposals is too important and block time should remain stable. This would require to modify the current check `if (CurrentBlock == Proposal.SubmitBlock + VotingPeriod)` to `if (CurrentBlock > Proposal.SubmitBlock + VotingPeriod) AND (Proposal.Status == ProposalStatusActive)`.

### Wire

The `wire.go` file allows developers to register the concrete message types of their module into the codec. In our case, we have two messages to declare:

```go
func RegisterWire(cdc *wire.Codec) {
    cdc.RegisterConcrete(SubmitProposalMsg{}, "simple_governance/SubmitProposalMsg", nil)
    cdc.RegisterConcrete(VoteMsg{}, "simple_governance/VoteMsg", nil)
}
```
Don't forget to call this function in `app.go` (see [Application - Bridging it all together](#application_-_bridging_it_all_together) for more).

### Errors

The `error.go` file allows us to define custom error messages for our module.  Declaring errors should be relatively similar in all modules. You can look in the [error.go](./error.go) file of our simple governance module for a concrete example. The code is self-explanatory.

Note that the errors of our module inherit from the `sdk.Error` interface and therefore possess the method `Result()`. This method is useful when there is an error in the `handler` and an error has to be returned in place of an actual result.

### Command-Line Interface and Rest API

Each module can define a set of commands for the Command-Line Interface and endpoints for the REST API. Let us create a `client` repository to define the commands and endpoints for our simple governance module.

```bash
mkdir client
cd client
mkdir cli
mkdir rest
```

#### Command-Line Interface (CLI)

Go in the `cli` folder and create a `simple_governance.go` file. This is where we will define the commands for our module.

The CLI builds on top of [Cobra](https://github.com/spf13/cobra). Here is the schema to build a command on top of Cobra:

```go
    // Declare flags
    const(
        Flag = "flag"
        ...
    )

    // Main command function. One function for each command.
    func Command(codec *wire.Codec) *cobra.Command {
        // Create the command to return
        command := &cobra.Command{
            Use: "actual command",
            Short: "Short description",
            Run: func(cmd *cobra.Command, args []string) error {
                // Actual function to run when command is used
            },
        }

        // Add flags to the command
        command.Flags().<Type>(FlagNameConstant, <example_value>, "<Description>")

        return command
    }
```

For a detailed implementation of the commands of the simple governance module, click [here](../client/cli/simple_governance.go).

#### Rest API

The Rest Server, also called [Light-Client Daemon (LCD)](https://github.com/cosmos/cosmos-sdk/tree/master/client/lcd), provides support for **HTTP queries**.

________________________________________________________

USER INTERFACE <=======> REST SERVER <=======> FULL-NODE

________________________________________________________

It allows end-users that do not want to run full-nodes themselves to interract with the chain. The LCD can be configured to perform **Light-Client verification** via the flag `--trust-node`, which can be set to `true` or `false`.

- If *light-client verification* is enabled, the Rest Server acts as a light-client and needs to be run on the end-user's machine. It allows them to interract with the chain in a trustless way without having to store the whole chain locally.

- If *light-client verification* is disabled, the Rest Server acts as a simple relayer for HTTP calls. In this setting, the Rest server needs not be run on the end-user's machine. Instead, it will probably be run by the same entity that operates the full-node the server connects to. This mode is useful if end-users trust the full-node operator and do not want to store anything locally.

Now, let us define endpoints that will be available for users to query through HTTP requests. These endpoints will be defined in a `simple_governance.go` file stored in the `rest` folder.

| Method | URL                             | Description                                                 |
|--------|---------------------------------|-------------------------------------------------------------|
| GET    | /proposals                      | Range query to get all submitted proposals                  |
| POST   | /proposals                      | Submit a new proposal                                       |
| GET    | /proposals/{id}                 | Returns a proposal given its ID                             |
| GET    | /proposals/{id}/votes           | Range query to get all the votes casted on a given proposal |
| POST   | /proposals/{id}/votes           | Cast a vote on a given proposal                             |
| GET    | /proposals/{id}/votes/{address} | Returns the vote of a given address on a given proposal     |

It is the job of module developers to provide sensible endpoints so that front-end developers and service providers can properly interact with it.

As for the actual in-code implementation of the endpoints for our simple governance module, you can take a look at [this file](../client/rest/simple_governance.go). Additionaly, here is a [link](https://hackernoon.com/restful-api-designing-guidelines-the-best-practices-60e1d954e7c9) for REST APIs best practices.

### Application - Bridging it all together

Now that we have built all the pieces that we need, it is time to integrate them into the application. Let us exit the `/x` director go back at the root of the SDK directory.

Then, let us create an `app` folder.

```bash
// At root level of directory
mkdir app
cd app
```

We are ready to create our simple governance application!

#### App structure

*Note: You can check the full file (with comments!) [here](link)*

First, create an `app.go` file. This is the main file that defines your application. In it, you will declare all the modules you need, their keepers, handlers, stores, etc. Let us take a look at each section of this file to see how the application is constructed.

First, we need to define the name of our application.

```go
const (
    appName = "SimpleGovApp"
)
```

Then, let us define the structure of our application.

```go
// Extended ABCI application
type SimpleGovApp struct {
    *bam.BaseApp
    cdc *wire.Codec

    // keys to access the substores
    capKeyMainStore      *sdk.KVStoreKey
    capKeyAccountStore   *sdk.KVStoreKey
    capKeyStakingStore   *sdk.KVStoreKey
    capKeySimpleGovStore *sdk.KVStoreKey

    // keepers
    feeCollectionKeeper auth.FeeCollectionKeeper
    coinKeeper          bank.Keeper
    stakeKeeper         simplestake.Keeper
    simpleGovKeeper     simpleGov.Keeper

    // Manage getting and setting accounts
    accountMapper auth.AccountMapper
}
```

- Each application builds on top of the `BaseApp` template, hence the pointer.
- `cdc` is the codec used in our application.
- Then come the keys to the stores we need in our application. For our simple governance app, we need 3 stores + the main store.
- Then come the keepers and mappers.

Let us do a quick reminder so that it is  clear why we need these stores and keepers. Our application is primarily based on the `simple_governance` module. However, we have established in section [Keepers for our app](#keepers_for_our_app) that our module needs access to two other modules: the `bank` module and the `stake` module. We also need the `auth` module for basic account functionalities. Finally, we need access to the main multistore to declare the stores of each of the module we use.

#### App commands

We will need to add the newly created commands to our application. Create a `cmd` folder inside your root  directory:

```bash
// At root level of directory
mkdir cmd
cd cmd
mkdir simplegovd
mkdir simplegovcli
```
Now within each of the `simplegov...` folders create a `main.go` file:

```bash
touch main.go
```

`simplegovd` is the folder that stores the command for running the server daemon, whereas `simplegovcli` defines the commands of your application.

##### CLI

To interact with out application we'll have to add the commands from the `simple_governance` module to our `simpleGov` application, as well as the pre-built SDK commands:

```go
//  cmd/simplegovcli/main.go
...
	rootCmd.AddCommand(
		client.GetCommands(
			simplegovcmd.GetCmdQueryProposal("proposals", cdc),
			simplegovcmd.GetCmdQueryProposals("proposals", cdc),
			simplegovcmd.GetCmdQueryProposalVotes("proposals", cdc),
			simplegovcmd.GetCmdQueryProposalVote("proposals", cdc),
		)...)
	rootCmd.AddCommand(
		client.PostCommands(
			simplegovcmd.PostCmdPropose(cdc),
			simplegovcmd.PostCmdVote(cdc),
		)...)
...
```

##### Daemon server

The `simplegovd` command will run the daemon server as a background process. First, create some util functions:

```go
//  cmd/simplegovd/main.go
// SimpleGovAppInit initial parameters
var SimpleGovAppInit = server.AppInit{
	AppGenState: SimpleGovAppGenState,
	AppGenTx:    server.SimpleAppGenTx,
}

// SimpleGovAppGenState sets up the app_state and appends the simpleGov app state
func SimpleGovAppGenState(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {
	appState, err = server.SimpleAppGenState(cdc, appGenTxs)
	if err != nil {
		return
	}
	return
}

func newApp(logger log.Logger, db dbm.DB) abci.Application {
	return app.NewSimpleGovApp(logger, db)
}

func exportAppState(logger log.Logger, db dbm.DB) (json.RawMessage, error) {
	dapp := app.NewSimpleGovApp(logger, db)
	return dapp.ExportAppStateJSON()
}
```

Now let's define the command for the daemon server within the `main()` function:

```go
//  cmd/simplegovd/main.go
func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "simplegovd",
		Short:             "Simple Governance Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, SimpleGovAppInit,
		server.ConstructAppCreator(newApp, "simplegov"),
		server.ConstructAppExporter(exportAppState, "simplegov"))

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.simplegovd")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)
	executor.Execute()
}
```

##### Makefile

The [Makefile](https://en.wikipedia.org/wiki/Makefile) compiles the Go program by defining a set of rules with targets and recipes. We'll need to add our application commands to it:

```
// Makefile
build_examples:
ifeq ($(OS),Windows_NT)
	...
	go build $(BUILD_FLAGS) -o build/simplegovd.exe ./examples/simpleGov/cmd/simplegovd
	go build $(BUILD_FLAGS) -o build/simplegovcli.exe ./examples/simpleGov/cmd/simplegovcli
else
	...
	go build $(BUILD_FLAGS) -o build/simplegovd ./examples/simpleGov/cmd/simplegovd
	go build $(BUILD_FLAGS) -o build/simplegovcli ./examples/simpleGov/cmd/simplegovcli
endif
...
install_examples:
    ...
	go install $(BUILD_FLAGS) ./examples/simpleGov/cmd/simplegovd
	go install $(BUILD_FLAGS) ./examples/simpleGov/cmd/simplegovcli
```

#### App constructor

Now, we need to define the constructor for our application.

```go
func NewSimpleGovApp(logger log.Logger, db dbm.DB) *SimpleGovApp
```

In this function, we will:

- Create the codec

```go
var cdc = MakeCodec()
```

- Instantiate our application. This includes creating the keys to access each of the substores.

```go
// Create your application object.
    var app = &SimpleGovApp{
        BaseApp:              bam.NewBaseApp(appName, cdc, logger, db),
        cdc:                  cdc,
        capKeyMainStore:      sdk.NewKVStoreKey("main"),
        capKeyAccountStore:   sdk.NewKVStoreKey("acc"),
        capKeyStakingStore:   sdk.NewKVStoreKey("stake"),
        capKeySimpleGovStore: sdk.NewKVStoreKey("simpleGov"),
    }
```

- Instantiate the keepers. Note that keepers generally need access to other module's keepers. In this case, make sure you only pass an instance of the keeper for the functionality that is needed. If a keeper only needs to read in another module's store, a read-only keeper should be passed to it.

```go
app.coinKeeper = bank.NewKeeper(app.accountMapper)
app.stakeKeeper = simplestake.NewKeeper(app.capKeyStakingStore, app.coinKeeper,app.RegisterCodespace(simplestake.DefaultCodespace))
app.simpleGovKeeper = simpleGov.NewKeeper(app.capKeySimpleGovStore, app.coinKeeper, app.stakeKeeper, app.RegisterCodespace(simpleGov.DefaultCodespace))
```

- Declare the handlers.

```go
app.Router().
        AddRoute("bank", bank.NewHandler(app.coinKeeper)).
        AddRoute("simplestake", simplestake.NewHandler(app.stakeKeeper)).
        AddRoute("simpleGov", simpleGov.NewHandler(app.simpleGovKeeper))
```

- Initialize the application.

```go
// Initialize BaseApp.
    app.MountStoresIAVL(app.capKeyMainStore, app.capKeyAccountStore, app.capKeySimpleGovStore, app.capKeyStakingStore)
    app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper))
    err := app.LoadLatestVersion(app.capKeyMainStore)
    if err != nil {
        cmn.Exit(err.Error())
    }
    return app
```

#### App codec

Finally, we need to define the `MakeCodec()` function and register the concrete types and interface from the various modules.

```go
func MakeCodec() *wire.Codec {
    var cdc = wire.NewCodec()
    wire.RegisterCrypto(cdc) // Register crypto.
    sdk.RegisterWire(cdc)    // Register Msgs
    bank.RegisterWire(cdc)
    simplestake.RegisterWire(cdc)
    simpleGov.RegisterWire(cdc)

    // Register AppAccount
    cdc.RegisterInterface((*auth.Account)(nil), nil)
    cdc.RegisterConcrete(&types.AppAccount{}, "simpleGov/Account", nil)
    return cdc
}
```

### Running the app

It's time to finally test our implementatio

#### Installation

Once you have finallized your application, install it using `go get`. The following commands will install the pre-built modules and examples of the SDK as well as your `simpleGov` application:

```bash
go get github.com/<your_username>/cosmos-sdk
cd $GOPATH/src/github.com/<your_username>/cosmos-sdk
make get_vendor_deps
make install
make install_examples
```

Check that the app is correctly installed by typing:

```bash
simplegovcli -h
simplegovd -h
```

#### Submit a proposal

Uuse the CLI to create a new proposal:

```bash
simplegovcli propose --title="Voting Period update" --description="Should we change the proposal voting period to 3 weeks?" --deposit=300Atoms
```

Get the details of your newly created proposal:

```bash
simplegovcli proposal 1
```

You can also check all the existing open proposals:

```bash
simplegovcli proposals --active=true
```

#### Cast a vote to an existing proposal

Let's cast a vote on the created proposal:

```bash
simplegovcli vote --proposal-id=1 --option="No"
```

Get the value of the option from your casted vote :

```bash
simplegovcli proposal-vote 1 <your_address>
```

You can also check all the casted votes of a proposal:

```bash
simplegovcli proposals-votes 1
```

### Testnet

WIP

### Conclusion

Congratulations ! You have succesfully created your first application and module with the Cosmos-SDK. If you have any question regarding this tutorial or about development on the SDK, please reach out us through our official communication channels:

- [Cosmos-SDK Riot Channel](https://riot.im/app/#/room/#cosmos-sdk:matrix.org)
- [Telegram](https://t.me/cosmosproject)
