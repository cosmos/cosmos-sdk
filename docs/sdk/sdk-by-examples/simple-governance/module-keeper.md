## Keeper

**File: [`x/simple_governance/keeper.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/x/simple_governance/keeper.go)**

### Short intro to keepers

`Keepers` are a module abstraction that handle reading/writing to the module store. This is a practical implementation of the **Object Capability Model** for Cosmos. 


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

### Store for our app

Before we delve into the keeper itself, let us see what objects we need to store in our governance sub-store, and how to index them.

- `Proposals` will be indexed by `'proposals'|<proposalID>`.
- `Votes` (`Yes`, `No`, `Abstain`) will be indexed by `'proposals'|<proposalID>|'votes'|<voterAddress>`.

Notice the quote mark on `'proposals'` and `'votes'`. They indicate that these are constant keywords. So, for example, the option casted by voter with address `0x01` on proposal `0101` will be stored at index `'proposals'|0101|'votes'|0x01`.

These keywords are used to faciliate range queries. Range queries (TODO: Link to formal spec) allow developer to query a subspace of the store, and return an iterator. They are made possible by the nice properties of the [IAVL+ tree](https://github.com/tendermint/iavl) that is used in the background. In practice, this means that it is possible to store and query a Key-Value pair in O(1), while still being able to iterate over a given subspace of Key-Value pairs. For example, we can query all the addresses that voted on a given proposal, along with their votes, by calling `rangeQuery(SimpleGovStore, <proposalID|'addresses'>)`.

### Keepers for our app

In our case, we only have one store to access, the `SimpleGov` store. We will need to set and get values inside this store via our keeper. However, these two actions do not have the same impact in terms of security. While there should no problem in granting read access to our store to other modules, write access is way more sensitive. So ideally application developers should be able to create either a governance mapper that can only get values from the store, or one that can both get and set values. To this end, we will introduce two keepers: `Keeper` and `KeeperRead`. When application developers create their application, they will be able to decide which of our module's keeper to use.

Now, let us try to think about which keeper from **external** modules our module's keepers need access to.
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

### Functions and Methods

The first function we have to create is the constructor.

```go
func NewKeeper(SimpleGov sdk.StoreKey, ck bank.Keeper, sm stake.Keeper, codespace sdk.CodespaceType) Keeper
```

This function is called from the main [`app.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/app/app.go) file to instanciate a new `Keeper`. A similar function exits for `KeeperRead`.

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