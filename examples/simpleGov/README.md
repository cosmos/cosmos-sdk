# Tutorial: How to code a Cosmos-SDK module

In this tutorial, we will learn the basics of coding a Cosmos-SDK module. Before getting into the bulk of it, let us remind some high level concepts about the Cosmos-SDK.

## Tendermint and Cosmos

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


## Introduciton to the Cosmos-SDK

Developing a Tendermint-based blockchain means that you only have to code the application (i.e. the business logic). But that in itself can prove to be rather difficult. This is why the Cosmos-SDK exists.

The Cosmos-SDK is a template framework to build secure blockchain applications on top of Tendermint. It is based on two major principles:

- **Composability:**  The goal of the Cosmos-SDK is to create an ecosystem of modules that allow developers to easily spin up sidechains without having to code every single functionality of their application. Anyone can create a module for the Cosmos-SDK, and using already-built modules in your blockchain is as simple as importing them into your application. For example, the Tendermint team is building a set of basic modules that are needed for the Cosmos Hub, like accounts, staking, IBC, governance. Now if you want to develop a public Tendermint blockchain compatible with Cosmos that has the aforementioned functionalities, you just have to import these already-built modules. As a developer, you only have to create the modules required by your application that do not already exist. As the Cosmos ecosystem develops, we expect the modules ecosystem to gracefully develop, making it easier and easier to develop complex blockchain applications.
- **Capabilities:** Most developers will need to access other modules when building their own modules. The Cosmos-SDK being an open framework, it is likely that some of these modules will be malicious. To address these threats, the Cosmos-SDK is designed to be the foundation of a capabilities-based system. In practice, this means that instead of having each module keep an access control list to give access to other modules, each module implement `mappers` that can be passed to other modules to grant a pre-defined set of capabilities. For example, if an instance of module A's `mapper` is passed to module B, module B will be able to call a restricted set of module A's functions. The *capabilities* of each mapper are defined by the module's developer, and it is the job of the application developer to instanciate and pass mappers from module to module properly. For a deeper look at capabilities, you can read this cool [article](http://habitatchronicles.com/2017/05/what-are-capabilities/)

Now that we have a better understanding of the high level principles of the SDK, let us take a deeper look at how a Cosmos-SDK application is constructed.

*Note: For now the Cosmos-SDK only exists in Golang, which means that module developers can only develop SDK modules in Golang. In the future, we expect that Cosmos-SDK in other programming languages will pop up*

## Reminder on Tendermint and ABCI

Todo

## Architecture of a SDK-app

The Cosmos-SDK gives the basic template for your application architecture. You can find this template [here](https://github.com/cosmos/cosmos-sdk).

In essence, a blockchain application is simply a replicated state machine. There is a state (e.g. for a cryptocurrency, how many coins each account holds), and transactions that trigger state transitions. As the application developer, your job is just define the state, the transactions types and how different transactions modify the state. 

### Modularity

The Cosmos-SDK is a module-based framework. Each module is in itself a little state-machine that can be gracefully combined with other modules to produce a coherent application. In other words, modules define a sub-section of the global state and of the transaction types. Then, it is the job of the root application to route messages to the correct modules depending on their respective types. To understand this process, let us take a look at a simplified standard cycle of the state-machine.

Upon receiving a transaction from the Tendermint Core engine, here is whatthe application does:

1. Decode the transaction and get the message
2. Route the message to the appropriate module using the `Msg.Type()` method
3. Run the transaction in the module. Modify the state if the transaction is valid.
4. Return new state or error message

Steps 1, 2 and 4 are handled by the root application. Step 3 is handled by the appropriate module. 

### SDK Components 

With this in mind, let us go through the important directories of the SDK:

- `baseapp`: This defines the template for a basic application. Basically it implements the ABCI protocol so that your Cosmos-SDK application can communicate with the underlying Tendermint node.
- `client`: Command-Line to interface with the application
- `server`: Rest server to communicate with the node
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

Stores can be cache-wrapped to provide transactions at the persistence level
(and this is well supported for iterators as well). This feature is used to
provide a layer of transactional isolation for transaction processing after the
"AnteHandler" deducts any associated fees for the transaction.  Cache-wrapping
can also be useful when implementing a virtual-machine or scripting environment
for the blockchain.

#### go-amino

The Cosmos-SDK uses
[go-amino](https://github.com/cosmos/cosmos-sdk/blob/96451b55fff107511a65bf930b81fb12bed133a1/examples/basecoin/app/app.go#L97-L111)
extensively to serialize and deserialize Go types into Protobuf3 compatible
bytes.

Go-amino (e.g. over https://github.com/golang/protobuf) uses reflection to
encode/decode any Go object.  This lets the SDK developer focus on defining
data structures in Go without the need to maintain a separate schema for
Proto3.  In addition, Amino extends Proto3 with native support for interfaces
and concrete types.

For example, the Cosmos SDK's `x/auth` package imports the PubKey interface
from `tendermint/go-crypto` , where PubKey implementations include those for
Ed25519 and Secp256k1.  Each auth.BaseAccount has a PubKey.

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

Amino knows what concrete type to decode for each interface value
based on what concretes are registered for the interface.

For example, the "Basecoin" example app knows about Ed25519 and Secp256k1 keys
because they are registered by the app's codec below:

```go
wire.RegisterCrypto(cdc) // Register crypto. 
```

For more information on Go-Amino, see https://github.com/tendermint/go-amino.

#### Keys, Keepers, and Mappers

The Cosmos SDK is designed to enable an ecosystem of libraries that can be
imported together to form a whole application.  To make this ecosystem
more secure, we've developed a design pattern that follows the principle of 
least-authority.

Mappers and Keepers provide access to KV stores via the context.  The only
difference between the two is that a Mapper provides a lower-level API, so
generally a Keeper might hold references to other Keepers and Mappers but not
vice versa.

Mappers and Keepers don't hold any references to any stores directly.  They only
hold a "key" (the `sdk.StoreKey` below):

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

This way, you can hook everything up in your main app.go file and see what
components have access to what stores and other components.

```go
// Define the accountMapper. 
 app.accountMapper = auth.NewAccountMapper( 
    cdc, 
    app.keyAccount,      // target store 
    &types.AppAccount{}, // prototype 
 ) 
```

Later during the execution of a transaction (e.g. via ABCI DeliverTx after a
block commit) the context is passed in as the first argument.  The context
includes references to any relevant KV stores, but you can only access them if
you hold the associated key.

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

Mappers and Keepers cannot hold direct references to stores because the store
is not known at app initialization time.  The store is dynamically created (and
wrapped via CacheKVStore as needed to provide a transactional context) for
every committed transaction (via ABCI DeliverTx) and mempool check transaction
(via ABCI CheckTx). 

#### Tx, Msg, Handler, and AnteHandler

A transaction (Tx interface) is a signed/authenticated message (Msg interface).

Transactions that are discovered by the Tendermint mempool are processed by the
AnteHandler ("ante" just means "before") where the validity of the transaction
is checked and any fees are collected.

Transactions that get committed in a block first get processed through the
AnteHandler, and if the transaction is valid after fees are deducted, they are
processed through the appropriate Handler.

In either case, the transaction bytes must first be parsed.  The default
transaction parser uses Amino.  Most SDK developers will want to use the
standard transaction structure defined in the `x/auth` package (and the
corresponding AnteHandler implementation also provided in `x/auth`):

```go 
 // StdTx is a standard way to wrap a Msg with Fee and Signatures. 
 // NOTE: the first signature is the FeePayer (Signatures must not be nil). 
 type StdTx struct { 
    Msg        sdk.Msg        `json:"msg"` 
    Fee        StdFee         `json:"fee"` 
    Signatures []StdSignature `json:"signatures"` 
 } 
```

Various packages generally define their own message types.  The Basecoin
example app includes multiple message types that are registered in app.go:

```go 
sdk.RegisterWire(cdc)    // Register Msgs 
 bank.RegisterWire(cdc) 
 stake.RegisterWire(cdc) 
 slashing.RegisterWire(cdc) 
 ibc.RegisterWire(cdc)
```

Finally, handlers are added to the router in your app.go file to map messages
to their corresponding handlers. (In the future we will provide more routing
features to enable pattern matching for more flexibility).

```go
 // register message routes 
 app.Router(). 
    AddRoute("auth", auth.NewHandler(app.accountMapper)). 
    AddRoute("bank", bank.NewHandler(app.coinKeeper)). 
    AddRoute("ibc", ibc.NewHandler(app.ibcMapper, app.coinKeeper)). 
    AddRoute("stake", stake.NewHandler(app.stakeKeeper)) 
```

#### EndBlocker

The EndBlocker hook allows us to register callback logic to be performed at the
end of each block.  This lets us process background events, such as processing
validator inflationary atom provisions:

```go
// Process Validator Provisions 
 blockTime := ctx.BlockHeader().Time // XXX assuming in seconds, confirm 
 if pool.InflationLastTime+blockTime >= 3600 { 
    pool.InflationLastTime = blockTime 
    pool = k.processProvisions(ctx) 
 } 
```

By the way, the SDK provides a staking module, which provides all the
bonding/unbonding funcionality for the Cosmos Hub:
https://github.com/cosmos/cosmos-sdk/tree/develop/x/stake (staking module)

#### Start working

So by now you should have realized how easy it is to build a Tendermint blockchain on top of the Cosmos-SDK. You just have to follow these simple steps:

1. Clone the Cosmos-SDK repo
2. Code the modules needed by your application that do not already exist
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
git checkout -b my_new_module
```

Finally, let us create the repository for our module:

```bash
cd x
mkdir module_tutorial
```

We are all set! 

## Designing the app

### Simple governance app

For this tutorial, we will code a simple governance application, and a simple governance module. It will allow us to show all the basic notions required to build a functioning module. Note that this is not the module used for the governance of the Cosmos Hub. A much more [advanced governance module](https://github.com/cosmos/cosmos-sdk/tree/develop/x/gov) for the Cosmos-SDK is available. 

All the code for the simple_governance application can be found [here](https://github.com/gamarin2/cosmos-sdk/tree/module_tutorial/examples/basecoin/x/simple_governance). You'll notice that the module and app aren't located at the root level of the repo but in the examples directory. This is just for convenience, you can code your module and app in the base repos. 

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

In terms of store, we will just create one KVStore in the multistore to store `Proposals`. We will also store the option (`Yes`, `No` or `Abstain`) chosen by each voter on each proposal.

- `Proposals` will be indexed by `<proposalID|'proposal'>`.
- `Options` (`Yes`, `No`, `Abstain`) will be indexed by `<proposalID>|'addresses'|voterAddress`.

Notice the quote mark on `'proposal'` and `'addresses'`. This means that these are constant keywords. So, for example, the option casted by voter with address `0x01` on proposal `0101` will be stored at index `0101|'addresses'|0x01`.

These keywords are used to faciliate range queries. Range queries (TODO: Link to formal spec) allow developer to query a subspace of the store, and return an iterator. They are made possible by the nice properties of the [IAVL+ tree](https://github.com/tendermint/iavl) that is used in the background. In practice, this means that it is possible to store and query a KV pair in O(1) while still being able to iterate over a given subspace of KV pairs. For example, we can query all the addresses that voted on a given proposal, along with their options, by calling `rangeQuery(SimpleGovStore, <proposalID|'addresses'>)`.

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

#### Short intro to keepers

Keepers handles read and writes for your module's stores. This is where the notion of capability enters into play.

As module developers, we have to define keepers to interact with our module's store(s) not only from within our module, but also from other modules. When another module wants to access one of our module's store(s), a keeper for this store has to be passed to it at the application level. In practice, it will look like that:

```
keeperA = moduleA.newKeeper(app.moduleAStoreKey)
keeperB = moduleB.newKeeper(app.moduleBStoreKey)

app.Router().
        AddRoute("moduleA", moduleA.NewHandler(keeperA)).
        AddRoute("moduleB", moduleB.NewHandler(keeperB, keeperA))   // Here module B can access one of module A's store via the keeperA instance
```

`KeeperA` grants a set of capabilities to the handler of module B. When developing a module, it is good practice to think about the sensitivity of the different capabilities that can be granted through keepers. For example, some module may need to read and write to module A's main store, while others only need to read it. If your module has multiple stores, then some keepers could grant access to all of them, while others would only grant access to specific sub-stores. It is the job of the module developer to make sure it is easy for the application developer to instanciate a keeper with the right capabilities. Of course, the handler of your module will most likely get an unrestricted instance of the module's keeper. 

#### Keepers for our app

In our case, we only have one store to access, the `Proposals` store. We will need to set and get values inside this store via our keeper. However, these two actions do not have the same impact in terms of security. While there should no problem  granting read access to our store to other modules, write access is way more sensitive. So ideally application developers should be able to create either a governance mapper that can only get values from the store, or one that can both get and set values. To this end, we will introduce two keepers: `Keeper` and `KeeperRead`. When application developers create their application, they will be able to decide which of our module's keeper to use. 

Now let us try to think about which keepr from **external** modules our keepers need to have access to. 
Each proposal requires a deposit. This means our module needs to be able to both read and write to the module that handles tokens, which is the `bank` module. We only accept deposits in `Atoms` so, if possible, the application developer should pass a keeper of the `bank` module that can read and write to Atom balance only, and not other tokens, to our module's handler.

We also need to be able to determine the voting power of each voter based on their stake. To this end, we need read access to the store of the `staking` module. However, we don't need write access to this store. We should therefore indicate that in our module, and the application developer should be careful to only pass a read-only keeper of the `staking` module to our module's handler.

With all that in mind, we can define the structure of our `Keeper`:

```go
    type Keeper struct {
        ProposalStoreKey    sdk.StoreKey        // Key to our module's store
        Cdc                 *wire.Codec         // Codec to encore/decode structs
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
func NewKeeper(proposalStoreKey sdk.StoreKey, ck bank.CoinKeeper, sm stake.KeeperRead, codespace sdk.CodespaceType) Keeper
```

This function is called from the main `app.go` file to instanciate a new `Keeper`. A similar function exits for `KeeperRead`. 

```go
func NewKeeperRead(proposalStoreKey sdk.StoreKey, ck bank.CoinKeeper, sm stake.KeeperRead, codespace sdk.CodespaceType) KeeperRead
```

Depending on the needs of the application and its modules, either `Keeper`, `KeeperRead`, or both, will be instanciated at application level. 

*Note: Both the `Keeper` type name and `NewKeeper()` function's name are standard names used in every module. It is no requirement to follow this standard, but doing so can facilitate the life of the application developer*

Now, let us describe the methods that we need for our module's `Keeper`. For the full implementation, please refer to `keeper.go`.

- `GetProposal`: Get a `Proposal` given a `proposalID`. Proposals need to be decoded from `byte` before they can be read.
- `SetProposal`: Set a `Proposal` at index `proposalID|'proposal'`. Proposals need to be encoded to `byte` before they can be stored. 
- `NewProposalID`: A function to generate a new unique `proposalID`.
- `GetOption`: Get an `Option` given a `proposalID` and a `voterAddress`.
- `SetOption`: Set an `Option` given a `proposalID` and a `voterAddress`.
- Proposal Queue methods: These methods implement a standard proposal queue to store `Proposals` on a First-In First-Out basis. It is used to tally the votes at the end of the voting period.

The last thing that needs to be done is to override certain methods for the `KeeperRead` type. `KeeperRead` should not have write access to the stores. Therefore, we will override the methods `SetProposal()`, `SetOption()` and `NewProposalID()`, as well as `setProposalQueue()` from the Proposal Queue's methods. For `KeeperRead`, these methods will just throw an error.

### Handler

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



### Wire

### Errors

### App - Bridging it all together

### Commands/Rest

### Testnet 

Nothing to see yet. Come back later! :3 