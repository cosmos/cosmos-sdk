# Tutorial: How to code a Cosmos-SDK module

In this tutorial, we will learn the basics of coding a Cosmos-SDK module. Before getting into the bulk of it, let us remind some high level concepts about the Cosmos-SDK.

### Tendermint and Cosmos

Blockchains can be divided in three conceptual layers: 

- **Networking:** Responsible for message propagation.
- **Consensus:** Enables validator nodes to agree on the next set of transactions to process (i.e. add blocks to the blockchain).
- **Application:** Responsible for processing transactions, which modify the state.

Before Tendermint, building a blockchain required building all three layers from the ground up. What Tendermint does is providing a generic blockchain engine responsible for Networking and Consensus. With Tendermint, developer can enjoy a high-performance consensus engine and only worry about the application part.

Tendermint connects the blockchain engine (Networking and Consensus Layers) to the Application via a protocol called ABCI. Developers only have to implement a few messages to build an ABCI-application that runs on top of the Tendermint engine. ABCI is language agnostic, meaning that developers can build the application part of their blockchain in any programming language. Building on top of Tendermint also provides the following benefits:

- **Public or private blockchain capable.** Since developers can deploy arbitrary applications on top of Tendermint, it is possible to develop both permissioned and permissionless blockchains on top of it.
- **Performance.** Tendermint is a state of the art blockchain engine. Tendermint Core can have a block time on the order of 1 second and handle thousands of transactions per second.
- **Instant finality.** A property of the Tendermint consensus algorithm is instant finality, meaning that forks are never created, as long as less than a third of the validators are malicious (byzantine). Users can be sure their transactions are finalized as soon as a block is created.
- **Security.** Tendermint consensus is not only fault tolerant, it’s optimally Byzantine fault-tolerant, with accountability. If the blockchain forks, there is a way to determine liability.

But most importantly, Tendermint is natively compatible with the Inter Blockchain Communication Protocol. This means that any Tendermint blockchain, whether public or private, can be natively connected to the Cosmos ecosystem and securely exchange tokens with other blockchains in the ecosystem. Note that benefiting from interoperability via IBC and Cosmos preserves the sovereignty of your Tendermint chain. Non-Tendermint chains can also be connected to Cosmos via IBC adapters or Peg-Zones, but this is out of scope for this document.


### Cosmos-SDK

See [overview.md](overview.md)

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
git checkout -b my_new_module
```

Finally, let us create the repository for our module:

```bash
cd x
mkdir module_tutorial
```

We are all set! 

## Designing the module

### Simple governance module

For this tutorial, we will code a simple governance module. It will allow us to show all the basic notions required to build a functioning module. Note that this is not the module used for the governance of the Cosmos Hub. A much more [advanced governance module](https://github.com/cosmos/cosmos-sdk/tree/develop/x/gov) for the Cosmos-SDK is available. 

All the code for the simple_governance module can be found [here](https://github.com/gamarin2/cosmos-sdk/tree/module_tutorial/examples/basecoin/x/simple_governance). You'll notice that the module and app aren't located at the root level of the repo but in the examples directory. This is just for convenience, you can code your module and app in the base repos. 

Without further talk, let's get into it!

### Requirements

We will start by writting down our module's requirements. We are designing a simple governance module, in which we want:

- Simple text proposals, that any coin holder can submit
- Proposals must be submitted with a deposit in Atoms. If the deposit is superior to a predefined value called `MinDeposit`, the proposal enters the voting period. Otherwise it is rejected.
- Bonded Atom holders can vote on proposal on a 1 bonded Atom 1 vote basis
- Bonded Atom holders can choose between 3 options when casting a vote: `Yes`, `No` and `Abstain`.
- If, at the end of the voting period, there are more `Yes` votes than `No` votes, the proposal is accepted. Otherwise, it is rejected.
- Voting period is 2 weeks

When designing a module, it is good to adopt a certain methodology. Remember that a blockchain application is just a replicated deterministic state-machine. The state is just the representation of the application at a given time. It is up to the application devleper to define what the state will represent, depending on the goal of the application. For example, the state of a simple cryptocurrency application will just be a mapping of addresses to balances. 

The state can be updated according to predefined rules. Given a state and a transaction, the state-machine (i.e. the application) will return a new state. In a blockchain, transactions are bundled in blocks, but the logic is the same. Given a state and a set of transactions (a block), the application returns a new state. A SDK-module is just a subset of the application, but it is based on the same principles. As a result, module developers only have to define a subset of the state and a subset of the transaction types, which trigger state transitions. 

In summary, we have to define:

- A `State`, which represents a subset of the current state of the application.
- `Transactions`, which contain messages that trigger state transitions.


### State

Here, we will define the types we need (excluding transaction types), as well as the stores in the multistore. 

Our voting module is very simple, we only need a single type: `Proposals`. `Proposals` are item to be voted upon. They can be submitted by any users who have to provide a deposit.

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

In terms of store, it is also very simple. We will just create one KVStore in the multistore to store `Proposals`. Each proposal will be indexed by a unique key `proposalID`. 

### Transactions

The title of this section is a bit misleading. Indeed, what you as a module developer have to define is not `transactions`, but `messages`. Both transactions and messages exist in the Cosmos-SDK. A transaction differs from a message in that a message is contained in a transaction.  Transactions wrap around messages and add standard information like signatures and fees. As a module developer, you do not have to worry about transactions, only messages. 

Let us define the messages that we need in order to modify the state. Based on our features, we only need two messages:

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

Now that we have our types defined, we can start actually implementing the module. Let us start by adding the files we will need. Your module folder should look something like that:

```
- types.go

- handler.go
- mapper.go
- amino.go
- errors.go
- commands
```

Let us go into the detail of each of these files.

### types

In this file, you define the custom types for your module. This includes the types from the [State](#State) section and the custom message types for your module defined in the [Transactions](#Transactions) section.

For each new type that it not a message, you can add methods that make sense in the context of your application. In our case, we will implement an `updateTally` function to easily update the tally of a given proposal as vote messages come in.

Messages are a bit different. They implement the `Message` interface defined in the SDK's `types` folder. Here are the methods you need to implement when you define a custom message type:

- `Type()`: This function returns the route name of your module. When messages are processed by the application, they are routed using the string returned by the `Type() method.
- `GetSignBytes()`: Return the byte representation of the message. Is used to sign the message. 
- `GetSigners()`: Return addresses of the signer(s).  
- `ValidateBasic()`: This function is used to discard obviously invalid messages. It is called at the beginning of `runTx()` in the baseapp file. If `ValidateBasic()` does not return `nil`, the app stops running the transaction.
- `Get()`: A basic getter, returns some property of the message. 
- `String()`: Returns a human-readable version of the message

For our simple governance messages, this means:

- `Type()` will return `"simple_gov"`
- For `SubmitProposalMsg`, we need to make sure that the attributes are not empty and that the deposit is both valid and positive. Note that this is only basic validation, we will therefore not check in this method that the sender has sufficient funds to pay for the deposit
- For `VoteMsg`, we check that the address and option are valid and that the proposalID is not negative.
- As for other methods, less customization is required. You can check the code to see a standard way of implementing these.


### keeper

So here we start getting in the bulk of it. Keepers handles read and writes for your module's stores. This is where the notion of capability enters into play.

As module developers, we have to define keepers to interact with our module's store(s) not only from within our module, but also from other modules. When another module wants to access one of our module's store(s), a keeper for this store has to be passed to it at the application level. In practice, it will look like that:

```
keeperA = moduleA.newKeeper(app.moduleAStoreKey)
keeperB = moduleB.newKeeper(app.moduleBStoreKey)

app.Router().
        AddRoute("moduleA", moduleA.NewHandler(keeperA)).
        AddRoute("moduleB", moduleB.NewHandler(keeperB, keeperA))   // Here module B can access one of module A's store via the keeperA instance
```

`KeeperA` grants a set of capabilities to the handler of module B. When developing a module, it is good practice to think about the sensitivity of the different capabilities that can be granted through keepers. For example, maybe some module need to read and write to module A's main store, while others only need to read it. If your module has multiple stores, then some keepers could grant access to all of them, while others would only grant access to specific sub-stores. It is the job of the module developer to make sure it is easy for the application developer to instanciate a keeper with the right capabilities. Of course, the handler of your module will most likely get an unrestricted instance of the module's keeper. 

In our case, we only have one store to access, the `Proposals` store. We will need to set and get values inside this store via our keeper. However, these two actions do not have the same impact in terms of security. While there should no problem  granting read access to our store to other modules, write access is way more sensitive. So ideally application developers should be able to create either a governance mapper that can only get values from the store, or one that can both get and set values. To this end, we will introduce two keepers: `keeper` and `keeperReadOnly`.

Now, when application developers create their application, they will be able to decide which of our module's keeper to use. If another module requires an instance of our module's keeper to be passed to it, the application developer will only pass the version of our mapper that makes sense.

Now let us try to think about which mapper from outside modules our keepers need to have access to. 

Each proposal requires a deposit. This means our module needs to be able to both read and write to the module that handles tokens, which is the bank module. We only accept deposits in `Atoms` so, if possible, the application developer should pass a keeper of the `bank` module that can read and write to Atom balance only, and not other tokens.

We also need to be able to determine the voting power of each voter based on their stake. To this end, we need read access to the store of the `staking` module. However, we don't need write access to this store. We should therefore indicate that in our module, and the application developer should be careful to only pass a read-only keeper of the `staking` module to our module.

### Handler

### Wire

### Errors

### Commands/Rest

### Tests

### App - Bridging it all together

### Testnet 