# Multistore

## Prerequisites


## Synopsis
Every blockchain application stores **internal state**, which includes all of the data relevant to the application and is updated through transactions. Depending on the application, the internal state may include account balances, ownership of items, governance logic, etc. This doc describes the different `Stores` defined by the SDK for the purpose of storing internal state.

## Multistore Design
### Design Requirements
To build a data structure for a blockchain application, there are several necessary functionalities. In a distributed computation model, state changes are executed by nodes locally before and after consensus (ultimately they must agree on the same state), and executions may sometimes fail due to errors, insufficient gas or other. Thus, the store should be able to make changes but also revert back to a previous state if needed.

Also specific to blockchain applications is the idea of committing: after blocks have been finalized, nodes not only execute the state changes of that block but also commit them. Committing finalizes the state changes and requires that, in the future, nodes should be able to prove that they arrived at this particular state.

Following with the SDK's modular design, applications may use multiple modules to handle various functionalities, but those modules should not have access to portions of state that don't pertain to them. Also important for building using the Cosmos SDK is implementing ABCI - in particular, to enable Queries from the consensus engine.

Let's build up to a Multistore from the bottom up.

### Data Structure
The most basic function is to store data. Every **base store** has an underlying data structure that actually stores the data, typically structured as a tree or database. One example of a base store is [`IAVL`](https://github.com/cosmos/cosmos-sdk/blob/master/store/iavl/store.go) which will be described in the Implementation section below. Some stores, like [`CacheKVStore`](https://github.com/cosmos/cosmos-sdk/blob/9a16e2675f392b083dd1074ff92ff1f9fbda750d/store/cachekv/store.go)s, [`TraceKV`](https://github.com/cosmos/cosmos-sdk/blob/9a16e2675f392b083dd1074ff92ff1f9fbda750d/store/tracekv/store.go)s and [`GasKV`](https://github.com/cosmos/cosmos-sdk/blob/9a16e2675f392b083dd1074ff92ff1f9fbda750d/store/gaskv/store.go)s are **wrapper stores** that don't define their own data structures but instead wrap an underlying KVStore or Multistore and add extra functionality such as (respectively) caching, tracing, and tracking gas.

### Store and KVStore
At the next step is a [Store](https://github.com/cosmos/cosmos-sdk/blob/9a16e2675f392b083dd1074ff92ff1f9fbda750d/store/types/store.go#L12-L15) which has a data structure, a [StoreType](https://github.com/cosmos/cosmos-sdk/blob/36dcd7b7ad94cf59a8471506e10b937507d1dfa5/store/types/store.go#L201-L209) (DB, IAVL, or Transient), and is a [CacheWrapper](https://github.com/cosmos/cosmos-sdk/blob/36dcd7b7ad94cf59a8471506e10b937507d1dfa5/store/types/store.go#L157-L178). CacheWrapping is described later; for now, `Store` is just a basic unit of data storage with knowledge of its underlying data structure.

In most cases, instead of Store, the interface implemented is a [KVStore](https://github.com/cosmos/cosmos-sdk/blob/5344e8d768f306c29eb5451177499bfe540a80e9/store/types/store.go#L103-L133) which is similar to a Store but structured as an iterable key-value mapping (similar to a Python dictionary). The functions `Get`, `Set`, `Has`, and `Delete` all require a Key to be passed in to gain access to the Value. An Iterator should also be implemented.

### Access Management
Now, we get to the actual Multistore interface, which is a type that can have multiple Stores.
```go
type MultiStore interface {
	Store
	CacheMultiStore() CacheMultiStore
	GetStore(StoreKey) Store
	GetKVStore(StoreKey) KVStore
	TracingEnabled() bool
	SetTracer(w io.Writer) MultiStore
	SetTracingContext(TraceContext) MultiStore
}
```
One of the key functions of the Multistore is to control access and provide some level of security. To enable access management, most applications have an `app.go` or main module that holds the application's internal state in a Multistore and controls access to each portion of it. Multistores may also have multiple substores, which themselves can be Multistores or some other flavor of Stores. The SDK has an object-capability model for security: to gain access to a store or substore itself, a module calls a Getter function (i.e. `GetStore` or `GetKVStore`) with the corresponding `StoreKey`. To clarify, the StoreKey is not the same as a key in a key-value dictionary; it is more like a secret key needed to "unlock" the store. These keys are secret and cannot be forged. These functions make it convenient to access stores and substores but offer security by restricting access to only modules that have the correct key.

### Revert Capability
This functionality is enabled by the [CacheWrapper](https://github.com/cosmos/cosmos-sdk/blob/36dcd7b7ad94cf59a8471506e10b937507d1dfa5/store/types/store.go#L157-L178) part of `Store`. As a `CacheWrapper`, a store can `CacheWrap` which means it can create a deep copy of the underlying data structure: calling CacheWrap on a KVStore returns a [CacheKVStore](https://github.com/cosmos/cosmos-sdk/blob/9a16e2675f392b083dd1074ff92ff1f9fbda750d/store/cachekv/store.go). Changes can be made to this copy without affecting the original store, and you can later `Write` or sync those changes to the original store. The function `CacheMultiStore` creates this copy. If a transaction makes a few state changes, then errors or runs out of gas, this Cache Wrapping functionality allows the state to be reverted back.

### Tracing
If `TracingEnabled`, a Multistore also enables tracing of operations on the store: a `traceWriter` is a [`Writer`](https://golang.org/pkg/io/#Writer) that logs each change to the store, and `traceContext` is the state that is being traced. Both are initialized using their respective Setters, `SetTracer` and `SetTracingContext`. This tracer allows us to track changes; it is useful for debugging.

### Commit
To enable committing, Multistores should also implement [CommitStore](https://github.com/cosmos/cosmos-sdk/blob/36dcd7b7ad94cf59a8471506e10b937507d1dfa5/store/types/store.go#L17-L28). CommitStore  has a `commit` function, which allows state changes to persist afterward and outputs a [CommitID](https://github.com/cosmos/cosmos-sdk/blob/36dcd7b7ad94cf59a8471506e10b937507d1dfa5/store/types/store.go#L180-L197). The CommitID can be used to prove committed state changes in the future; in practice, this typically includes a Merkle Root.

### Query
Multistores should implement [Queryable](https://github.com/cosmos/cosmos-sdk/blob/36dcd7b7ad94cf59a8471506e10b937507d1dfa5/store/types/store.go#L30-L36) which enables **ABCI queries**. Implementing ABCI is important for the application to communicate with the consensus engine (i.e. Tendermint). More on the ABCI can be found in [this doc](https://tendermint.com/docs/spec/abci/).

## Implementation
There are a few existing [stores](https://github.com/cosmos/cosmos-sdk/tree/9a16e2675f392b083dd1074ff92ff1f9fbda750d/store) in the SDK that implement the Multistore interface, each with varying functionalities and purposes.

### IAVL
One example is [IAVL](https://github.com/cosmos/cosmos-sdk/blob/master/store/iavl/store.go), which implements the `KVStore` and  `CommitStore` interfaces, enables ABCI queries, and is structured as a self-balancing Merkle Tree. Information on the underlying IAVL tree can be found [here](https://github.com/tendermint/iavl). This doc describes how the IAVL Store implements the Multistore interfaces.

To implement the `Store`, it returns `StoreTypeIAVL` as the StoreType and creates a new store as a copy of the current one in order to Cache Wrap. To implement `KVStore`, it uses the underlying IAVL tree's `Set`, `Get`, `Has` and `Remove`, and defines an [IAVLIterator](https://github.com/cosmos/cosmos-sdk/blob/f4a96fd6b65ff24d0ccfe55536a2c3d6abe3d3fa/store/iavl/store.go#L256-L283) used to iterate through the data structure.

To implement a `Committer` for CommitStore, it obtains the current `hash` (which is the merkle root) and `version` of the underlying IAVL tree by calling its [`SaveVersion`](https://github.com/tendermint/iavl/blob/de0740903a67b624d887f9055d4c60175dcfa758/mutable_tree.go#L320-L364) function and returns them as a CommitID. `LastCommitID` retrieves the previous `CommitID` by simply by getting the tree's currently saved hash and version instead of calculating new ones.

The IAVL Store also defines a [`Query`](https://github.com/cosmos/cosmos-sdk/blob/f4a96fd6b65ff24d0ccfe55536a2c3d6abe3d3fa/store/iavl/store.go#L178-L252) function to implement the ABCI interface. The interaction starts with a [`Request`](https://github.com/tendermint/tendermint/blob/4514842a631059a4148026ebce0e46fdf628f875/abci/types/types.pb.go) sent to the Store and ends with the Store outputting a [`Response`](https://github.com/tendermint/tendermint/blob/4514842a631059a4148026ebce0e46fdf628f875/abci/types/types.pb.go). `Query` first sets which height it will query the tree for, then gets the data appropriate for what the Request is asking for. It could ask for the value corresponding to a certain `Key`, as well as the Merkle Proof for it, or it may ask for `subspace`, which is the list of all `KVPairs` with a certain prefix.
