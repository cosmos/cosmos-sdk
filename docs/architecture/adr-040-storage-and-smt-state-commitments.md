# ADR 040: Storage and SMT State Commitments

## Changelog

- 2020-01-15: Draft

## Status

DRAFT Not Implemented

## Abstract

Sparse Merke Tree ([SMT](https://osf.io/8mcnh/)) is a version of a Merkle Tree with various storage and performance optimizations. This ADR defines a separation of state commitments from data storage and the SDK transition from IAVL to SMT.

## Context

Currently, Cosmos SDK uses IAVL for both state [commitments](https://cryptography.fandom.com/wiki/Commitment_scheme) and data storage.

IAVL has effectively become an orphaned project within the Cosmos ecosystem and it's proven to be an inefficient state commitment data structure.
In the current design, IAVL is used for both data storage and as a Merkle Tree for state commitments. IAVL is meant to be a standalone Merkelized key/value database, however it's using a KV DB engine to store all tree nodes. So, each node is stored in a separate record in the KV DB. This causes many inefficiencies and problems:

+ Each object query requires a tree traversal from the root. Subsequent queries for the same object are cached on the SDK level.
+ Each edge traversal requires a DB query.
+ Creating snapshots is [expensive](https://github.com/cosmos/cosmos-sdk/issues/7215#issuecomment-684804950). It takes about 30 seconds to export less than 100 MB of state (as of March 2020).
+ Updates in IAVL may trigger tree reorganization and possible O(log(n)) hashes re-computation, which can become a CPU bottleneck.
+ The node structure is pretty expensive - it contains a standard tree node elements (key, value, left and right element) and additional metadata such as height, version (which is not required by the SDK). The entire node is hashed, and that hash is used as the key in the underlying database, [ref](https://github.com/cosmos/iavl/blob/master/docs/node/node.md
).

Moreover, the IAVL project lacks support and a maintainer and we already see better and well-established alternatives. Instead of optimizing the IAVL, we are looking into other solutions for both storage and state commitments.

## Decision

We propose to separate the concerns of state commitment (**SC**), needed for consensus, and state storage (**SS**), needed for state machine. Finally we replace IAVL with [LazyLedgers' SMT](https://github.com/lazyledger/smt). LazyLedger SMT is based on Diem (called jellyfish) design [*] - it uses a compute-optimised SMT by replacing subtrees with only default values with a single node (same approach is used by Ethereum2) and implements compact proofs.

The storage model presented here doesn't deal with data structure nor serialization. It's a Key-Value database, where both key and value are binaries. The storage user is responsible for data serialization.

### Decouple state commitment from storage

Separation of storage and commitment (by the SMT) will allow the optimization of different components according to their usage and access patterns.

`SS` (SMT) is used to commit to a data and compute merkle proofs. `SC` is used to directly access data. To avoid collisions, both `SS` and `SC` will use a separate storage namespace (they could use the same database underneath). `SC` will store each `(key, value)` pair directly (map key -> value).

SMT is a merkle tree structure: we don't store keys directly. For every `(key, value)` pair, `hash(key)` is stored in a path (we hash a key to evenly distribute keys in the tree) and `hash(key, value)` in a leaf. Since we don't know a structure of a value (in particular if it contains the key) we hash both the key and the value in the `SC` leaf.

For data access we propose 2 additional KV buckets (namespaces for the key-value pairs, sometimes called [column family](https://github.com/facebook/rocksdb/wiki/Terminology)):

1. B1: `key → value`: the principal object storage, used by a state machine, behind the SDK `KVStore` interface: provides direct access by key and allows prefix iteration (KV DB backend must support it).
2. B2: `hash(key, value) → key`: a reverse index to get a key from an SMT path. Recall that SMT will store `(k, v)` as `(hash(k), hash(key, value))`. So, we can get an object value by composing `SMT_path → B2 → B1`.
3. we could use more buckets to optimize the app usage if needed.

We propose to use a KV database for both `SS` and `SC` - each run by its own database instance. This design allows for the separation of `SS` and `SC` into different hardware units, providing support for more complex setup scenarios and improving DB performance (essentially this will create two store shards: one for `SS` and another for `SC`). Moreover, we will be able to configure databases for `SS` and `SC` separately.

### Requirements

State Storage requirements:

+ range queries
+ quick (key, value) access
+ creating a snapshot
+ historical versioning
+ pruning (garbage collection)

State Commitment requirements:

+ fast updates
+ tree path should be short
+ query historical commitment proofs using ICS-23 standard
+ pruning (garbage collection)

### LazyLedger SMT for State Commitment

A Sparse Merkle tree is based on the idea of a complete Merkle tree of an intractable size. The assumption here is that as the size of the tree is intractable, there would only be a few leaf nodes with valid data blocks relative to the tree size, rendering a sparse tree.

### Snapshots for storage sync and state versioning

Below, with simple _snapshot_ we refer to a database snapshot mechanism, not to a _ABCI snapshot sync_. The latter will be referred as _snapshot sync_ (which will directly use DB snapshot as described below).

Database snapshot is a view of DB state at a certain time or transaction. It's not a full copy of a database (it would be too big). Usually a snapshot mechanism is based on a _copy on write_ and it allows DB state to be efficiently delivered at a certain stage.
Some DB engines support snapshotting. Hence, we propose to reuse that functionality for the state sync and versioning (described below). We limit the supported DB engines to ones which efficiently implement snapshots. In a final section we discuss the evaluated DBs.

One of the Stargate core features is a _snapshot sync_ delivered in the `/snapshot` package. It provides a way to trustlessly sync a blockchain without repeating all transactions from the genesis. This feature is implemented in SDK and requires storage support. Currently IAVL is the only supported backend. It works by streaming to a client a snapshot of a `SS` at a certain version together with a header chain.

A new database snapshot will be created in every `EndBlocker` and identified by a block height. The `root` store keeps track of the available snapshots to offer `SS` at a certain version. The `root` store implements the `RootStore` interface described below. In essence, `RootStore` encapsulates a `Committer` interface. `Committer` has a `Commit`, `SetPruning`, `GetPruning` functions which will be used for creating and removing snapshots. The `rootStore.Commit` function creates a new snapshot and increments the version on each call, and checks if it needs to remove old versions. We will need to update the SMT interface to implement the `Committer` interface.
NOTE: `Commit` must be called exactly once per block. Otherwise we risk going out of sync for the version number and block height.
NOTE: For the SDK storage, we may consider splitting that interface into `Committer` and `PruningCommitter` - only the multiroot should implement `PruningCommitter` (cache and prefix store don't need pruning).

Number of historical versions for `abci.RequestQuery` and state sync snapshots is part of a node configuration, not a chain configuration (configuration implied by the blockchain consensus). A configuration should allow to specify number of past blocks and number of past blocks modulo some number (eg: 100 past blocks and one snapshot every 100 blocks for past 2000 blocks). Archival nodes can keep all past versions.

Pruning old snapshots is effectively done by a database. Whenever we update a record in `SC`, SMT won't update nodes - instead it creates new nodes on the update path, without removing the old one. Since we are snapshotting each block, we need to change that mechanism to immediately remove orphaned nodes from the database. This is a safe operation - snapshots will keep track of the records and make it available when accessing past versions.

To manage the active snapshots we will either use a DB _max number of snapshots_ option (if available), or we will remove DB snapshots in the `EndBlocker`. The latter option can be done efficiently by identifying snapshots with block height and calling a store function to remove past versions.

#### Accessing old state versions

One of the functional requirements is to access old state. This is done through `abci.RequestQuery` structure.  The version is specified by a block height (so we query for an object by a key `K` at block height `H`). The number of old versions supported for `abci.RequestQuery` is configurable. Accessing an old state is done by using available snapshots.
`abci.RequestQuery` doesn't need old state of `SC` unless the `prove=true` parameter is set. The SMT merkle proof must be included in the `abci.ResponseQuery` only if both `SC` and `SS` have a snapshot for requested version.

Moreover, the SDK could provide a way to directly access a historical state. However, a state machine shouldn't do that - since the number of snapshots is configurable, it would lead to nondeterministic execution.

We positively [validated](https://github.com/cosmos/cosmos-sdk/discussions/8297) a versioning and snapshot mechanism for querying old state with regards to the database we evaluated.

### State Proofs

For any object stored in State Store (SS), we have corresponding object in `SC`. A proof for object `V` identified by a key `K` is a branch of `SC`, where the path corresponds to the key `hash(K)`, and the leaf is `hash(K, V)`.

### Rollbacks

We need to be able to process transactions and roll-back state updates if a transaction fails. This can be done in the following way: during transaction processing, we keep all state change requests (writes) in a `CacheWrapper` abstraction (as it's done today). Once we finish the block processing, in the `Endblocker`,  we commit a root store - at that time, all changes are written to the SMT and to the `SS` and a snapshot is created.

### Committing to an object without saving it

We identified use-cases, where modules will need to save an object commitment without storing an object itself. Sometimes clients are receiving complex objects, and they have no way to prove a correctness of that object without knowing the storage layout. For those use cases it would be easier to commit to the object without storing it directly.


### Refactor MultiStore

The Stargate `/store` implementation (StoreV1) adds an additional layer in the SDK store construction - the `MultiStore` structure. The multistore exists to support the modularity of the Cosmos SDK - each module is using its own instance of IAVL, but in the current implementation, all instances share the same database.

The latter indicates, however, that the implementation doesn't provide true modularity. Instead it causes problems related to race condition and sync (see: [\#6370](https://github.com/cosmos/cosmos-sdk/issues/6370)).

We propose to reduce the multistore concept from the SDK, and to use a single instance of `SC` and `SS` in a `RootStore` object. To avoid confusion, we should rename the `MultiStore` interface to `RootStore`.

Moreover, to improve usability, we should extend the `KVStore` interface with _prefix store_. This will allow module developers to bind a store to a namespace for module sub-components:

```go
type KVStore interface {
    ... current KVStore

    WithPrefix(prefix string) KVStore
}
```

The `WithPrefix` method will create a proxy object for the parent object and will prepend `prefix` to a key parameter of all key-value operations (similar to the `store/prefix.Store` implementation).
The following invariant must hold:

```
for each OP in [Get, Has, Set, ...]
    store.WithPrefix(prefix).OP(key) == store.OP(prefix + key)
```

The `RootStore` will have the following interface; the methods for configuring tracing and listeners are omitted for brevity.

```go
// Used where read-only access to versions is needed.
type BasicRootStore interface {
    Store
    GetKVStore(StoreKey) KVStore
    CacheRootStore() CacheRootStore
}

// Used as the main app state, replacing CommitMultiStore.
type CommitRootStore interface {
    BasicRootStore
    Committer
    Snapshotter

    GetVersion(uint64) (BasicRootStore, error)
    SetInitialVersion(uint64) error

    ... // Trace and Listen methods
}

// Replaces CacheMultiStore for branched state.
type CacheRootStore interface {
    BasicRootStore
    Write()

    ... // Trace and Listen methods
}

// Example of constructor parameters for the concrete type.
type RootStoreConfig struct {
    Upgrades        *StoreUpgrades
    InitialVersion  uint64

    ReservePrefix(StoreKey, StoreType)
}
```
<!-- TODO: Review whether these types can be further reduced or simplified -->
<!-- TODO: RootStorePersistentCache type -->

In contrast to `MultiStore`, `RootStore` doesn't allow to dynamically mount sub-stores or provide an arbitrary backing DB.

#### Compatibility support

To ease the transition to this new interface for users, we can create a shim which wraps a `CommitMultiStore` but provides a `CommitRootStore` interface, and expose functions to safely create and access the underlying `CommitMultiStore`.

The new `RootStore` and supporting types can be implemented in a `store/v2` package to avoid breaking existing code.

#### Merkle Proofs and IBC

Currently, an IBC (v1.0) Merkle proof path consists of two elements (`["<store-key>", "<record-key>"]`), with each key corresponding to a separate proof. These are each verified according to individual [ICS-23 specs](https://github.com/cosmos/ibc-go/blob/f7051429e1cf833a6f65d51e6c3df1609290a549/modules/core/23-commitment/types/merkle.go#L17), and the result hash of each step is used as the committed value of the next step, until a root commitment hash is obtained.
The root hash of the proof for `"<record-key>"` is hashed with the `"<store-key>"` to validate against the App Hash.

This is not compatible with the `RootStore`, which stores all records in a single Merkle tree structure, and won't produce separate proofs for the store- and record-key. Ideally, the store-key component of the proof could just be omitted, and updated to use a "no-op" spec, so only the record-key is used. However, because the IBC verification code hardcodes the `"ibc"` prefix and applies it to the SDK proof as a separate element of the proof path, this isn't possible without a breaking change. Breaking this behavior would severely impact the Cosmos ecosystem which already widely adopts the IBC module, and updating it across the chains is a time consuming effort.

As a workaround, the `RootStore` will have to maintain two logically separate SMT-based stores: one for IBC state and one for everything else. A simple Merkle map containing these two store keys will act as a `MultiStore`-like index, and the final App hash will be the root hash of this index.

This workaround can be used until the IBC connection code is fully upgraded and supports single-element commitment proofs.

### Optimization: compress module keys

We consider a compression of prefix keys using [Huffman Coding](https://en.wikipedia.org/wiki/Huffman_coding). It will require knowledge of used prefixes (module store keys) a priori. And for the best results, it will need frequency information for each prefix (how often objects are stored in the store under the same prefix key). For Merkle Proofs we can't use prefix compression - so it should only apply for the `SS`. Moreover, the prefix compression should be only applied for the module namespace (the first prefix level). More precisely:
+ each module has it's own namespace;
+ when accessing a module namespace we create a KVStore with embedded prefix;
+ that prefix will be compressed only for accessing and managing `SS`.

Huffman coding assures that in the set of prefix codes there are no two elements where one is a prefix of another:

```
for each k1, k2 \in {store.HuffmanCode(p): p \in StoreModuleKeys}
    assert( !k1.hasPrefix(k2) )
```

NOTE: We need to assure that the codes won't change. Huffman Coding depends on the keys and its frequency - so we would need to generate the codes and then fix the mapping in a static variable.

## Consequences

### Backwards Compatibility

This ADR doesn't introduce any SDK level API changes.

We change the storage layout of the state machine, a storage hard fork and network upgrade is required to incorporate these changes. SMT provides a merkle proof functionality, however it is not compatible with ICS23. Updating the proofs for ICS23 compatibility is required.

### Positive

+ Decoupling state from state commitment introduce better engineering opportunities for further optimizations and better storage patterns.
+ Performance improvements.
+ Joining SMT based camp which has wider and proven adoption than IAVL. Example projects which decided on SMT: Ethereum2, Diem (Libra), Trillan, Tezos, LazyLedger.
+ Multistore removal fixes a longstanding issue with the current MultiStore design.
+ Simplifies merkle proofs - all modules, except IBC, have only one pass for merkle proof.

### Negative

+ Storage migration
+ LL SMT doesn't support pruning - we will need to add and test that functionality.
+ `SS` keys will have an overhead of a key prefix. This doesn't impact `SC` because all keys in `SC` have same size (they are hashed).

### Neutral

+ Deprecating IAVL, which is one of the core proposals of Cosmos Whitepaper.

## Alternative designs

Most of the alternative designs were evaluated in [state commitments and storage report](https://paper.dropbox.com/published/State-commitments-and-storage-review--BDvA1MLwRtOx55KRihJ5xxLbBw-KeEB7eOd11pNrZvVtqUgL3h).

Ethereum research published [Verkle Tire](https://notes.ethereum.org/_N1mutVERDKtqGIEYc-Flw#fnref1) - an idea of combining polynomial commitments with merkle tree in order to reduce the tree height. This concept has a very good potential, but we think it's too early to implement it. The current, SMT based design could be easily updated to the Verkle Tire once other research implement all necessary libraries. The main advantage of the design described in this ADR is the separation of state commitments from the data storage and designing a more powerful interface.

## Further Discussions

### Evaluated KV Databases

We verified existing databases KV databases for evaluating snapshot support. The following databases provide efficient snapshot mechanism: Badger, RocksDB, [Pebble](https://github.com/cockroachdb/pebble). Databases which don't provide such support or are not production ready: boltdb, leveldb, goleveldb, membdb, lmdb.

### RDBMS

Use of RDBMS instead of simple KV store for state. Use of RDBMS will require an SDK API breaking change (`KVStore` interface), will allow better data extraction and indexing solutions. Instead of saving an object as a single blob of bytes, we could save it as record in a table in the state storage layer, and as a `hash(key, protobuf(object))` in the SMT as outlined above. To verify that an object registered in RDBMS is same as the one committed to SMT, one will need to load it from RDBMS, marshal using protobuf, hash and do SMT search.

### Off Chain Store

We were discussing use case where modules can use a support database, which is not automatically committed. Module will responsible for having a sound storage model and can optionally use the feature discussed in __Committing to an object without saving it_ section.

## References

+ [IAVL What's Next?](https://github.com/cosmos/cosmos-sdk/issues/7100)
+ [IAVL overview](https://docs.google.com/document/d/16Z_hW2rSAmoyMENO-RlAhQjAG3mSNKsQueMnKpmcBv0/edit#heading=h.yd2th7x3o1iv) of it's state v0.15
+ [State commitments and storage report](https://paper.dropbox.com/published/State-commitments-and-storage-review--BDvA1MLwRtOx55KRihJ5xxLbBw-KeEB7eOd11pNrZvVtqUgL3h)
+ [LazyLedger SMT](https://github.com/lazyledger/smt)
+ Facebook Diem (Libra) SMT [design](https://developers.diem.com/papers/jellyfish-merkle-tree/2021-01-14.pdf)
+ [Trillian Revocation Transparency](https://github.com/google/trillian/blob/master/docs/papers/RevocationTransparency.pdf), [Trillian Verifiable Data Structures](https://github.com/google/trillian/blob/master/docs/papers/VerifiableDataStructures.pdf).
+ Design and implementation [discussion](https://github.com/cosmos/cosmos-sdk/discussions/8297).
