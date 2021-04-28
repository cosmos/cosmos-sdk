# ADR 040: Storage and SMT State Commitments

## Changelog

- 2020-01-15: Draft

## Status

DRAFT Not Implemented


## Abstract

Sparse Merke Tree (SMT) is a version of a Merkle Tree with various storage and performance optimizations. This ADR defines a separation of state commitments from data storage and the SDK transition from IAVL to SMT.


## Context

Currently, Cosmos SDK uses IAVL for both state commitments and data storage.

IAVL has effectively become an orphaned project within the Cosmos ecosystem and it's proven to be an inefficient state commitment.
In the current design, IAVL is used for both data storage and as a Merkle Tree for state commitments. IAVL is meant to be a standalone Merkelized key/value database, however it's using a KV DB engine to store all tree nodes. So, each node is stored in a separate record in the KV DB. This causes many inefficiencies and problems:

+ Each object select requires a tree traversal from the root
+ Each edge traversal requires a DB query (nodes are not stored in a memory)
+ Creating snapshots is [expensive](https://github.com/cosmos/cosmos-sdk/issues/7215#issuecomment-684804950). It takes about 30 seconds to export less than 100 MB of state (as of March 2020).
+ Updates in IAVL may trigger tree reorganization and possible O(log(n)) hashes re-computation, which can become a CPU bottleneck.
+ The leaf structure is pretty expensive: it contains the `(key, value)` pair, additional metadata such as height, version. The entire node is hashed, and that hash is used as the key in the underlying database, [ref](https://github.com/cosmos/iavl/blob/master/docs/node/node.md
).

Moreover, the IAVL project lacks support and a maintainer and we already see better and well-established alternatives. Instead of optimizing the IAVL, we are looking into other solutions for both storage and state commitments.


## Decision

We propose to separate the concerns of state commitment (**SC**), needed for consensus, and state storage (**SS**), needed for state machine. Finally we replace IAVL with [LazyLedgers' SMT](https://github.com/lazyledger/smt). LazyLedger SMT is based on Diem (called jellyfish) design [*] - it uses a compute-optimised SMT by replacing subtrees with only default values with a single node (same approach is used by Ethereum2 as  well).

The storage model presented here doesn't deal with data structure nor serialization. It's a Key-Value database, where both key and value are binaries. The storage user is responsible for data serialization.

### Decouple state commitment from storage


Separation of storage and commitment (by the SMT) will allow the optimization of different components according to their usage and access patterns.

SMT will use it's own storage (could use the same database underneath) from the state machine store. For every `(key, value)` pair, the SMT will store `hash(key)` in a path (needed to evenly distribute keys in the tree) and `hash(key, value)` in a leaf (to bind the (key, value) pair stored in the `SS`). Since we don't know a structure of a value (in particular if it contains the key) we hash both the key and the value in the `SC` leaf.

For data access we propose 2 additional KV buckets (namespaces for the key-value pairs, sometimes called [column family](https://github.com/facebook/rocksdb/wiki/Terminology)):
1. B1: `key → value`: the principal object storage, used by a state machine, behind the SDK `KVStore` interface: provides direct access by key and allows prefix iteration (KV DB backend must support it).
2. B2: `hash(key, value) → key`: an index needed to extract a value (through: SMT → B2 → B1) having only a Merkle Path. Recall that SMT will store `hash(key, value)` in it's leafs.
3. we could use more buckets to optimize the app usage if needed.

Above, we propose to use a KV DB. However, for the state machine, we could use an RDBMS, which we discuss below.


### Requirements

State Storage requirements:
+ range queries
+ quick (key, value) access
+ creating a snapshot
+ prunning (garbage collection)

State Commitment requirements:
+ fast updates
+ tree path should be short
+ creating a snapshot
+ pruning (garbage collection)


### LazyLedger SMT for State Commitment

A Sparse Merkle tree is based on the idea of a complete Merkle tree of an intractable size. The assumption here is that as the size of the tree is intractable, there would only be a few leaf nodes with valid data blocks relative to the tree size, rendering the tree as sparse.


### Snapshots for storage sync and versioning

One of the Stargate core features are snapshots and state sync delivered in the `/snapshot` package. This feature is implemented in SDK and requires storage support. Currently IAVL is the only supported backend.

Database snapshot is a view of DB state at a certain time or transaction. It's not a full copy of a database (it would be too big), usually a snapshot mechanism is based on a _copy on write_ and it allows to efficiently deliver DB state at a certain stage.
Some DB engines support snapshotting. Hence, we propose to reuse that functionality for the state sync and versioning (described below). It will the supported DB engines to ones which efficiently implement snapshots. In a final section we will discuss evaluated DBs.

New snapshot will be created in every `EndBlocker`. The `rootmulti.Store` keeps track of the version number and implements the `MultiStore` interface. `MultiStore` encapsulates a `Commiter` interface, which has the `Commit`, `SetPruning`, `GetPruning` functions which will be used for creating and removing snapshots. The `Store.Commit` function increments the version on each call, and checks if it needs to remove old versions. We will need to update the SMT interface to implement the `Commiter` interface.
NOTE: `Commit` must be called exactly once per block. Otherwise we risk going out of sync for the version number and block height.
NOTE: For the SDK storage, we may consider splitting that interface into `Committer` and `PrunningCommiter` - only the multiroot should implement `PrunningCommiter` (cache and prefix store don't need pruning).

Number of historical versions (snapshots) for `abci.Query` and fast sync is part of a node configuration, not a chain configuration (configuration implied by the blockchain consensus). A configuration should allow to specify number of past blocks and number of past blocks modulo some number (eg: 100 past blocks and one snapshot every 100 blocks for past 2000 blocks). Archival nodes can keep all snapshots.

Pruning old snapshots is effectively done by a database. Whenever we update a record in `SC`, SMT will create a new one without removing the old one. Since we are snapshoting each block, we update the mechanism and immediately remove an orphaned from the storage. This is a safe operation - snapshots will keep track of the records which should be available for past versions.

To manage the active snapshots we will either us a DB _max number of snapshots_ option (if available), or will remove snapshots in the `EndBlocker`. The latter option can be done efficiently by identifying snapshots with block height.

#### Accessing old state versions

One of the functional requirements is to access old state. This is done through `abci.Query` structure.  The version is specified by a block height (so we query for an object by a key `K` at block height `H`). The number of old versions supported for `abci.Query` is configurable. Accessing an old state is done by using available snapshots.
`abci.Query` doesn't need old state of `SC`. So, for efficiency, we should keep `SC` and `SS` in different databases (however using the same DB engine).

Moreover, SDK could provide a way to directly access the state. However, a state machines shouldn't do that - since the number of snapshots is configurable, it would lead to nondeterministic execution.

We positively [validated](https://github.com/cosmos/cosmos-sdk/discussions/8297) a snapshot mechanism for querying old state with regards to the database we evaluated.

### State Proofs

For any object stored in State Store (SS), we have corresponding object in `SC`. A proof for object `V` identified by a key `K` is a branch of `SC`, where the path corresponds to the key `hash(K)`, and the leaf is `hash(K, V)`.

### Rollbacks

We need to be able to process transactions and roll-back state updates if a transaction fails. This can be done in the following way: during transaction processing, we keep all state change requests (writes) in a `CacheWrapper` abstraction (as it's done today). Once we finish the block processing, in the `Endblocker`,  we commit a root store - at that time, all changes are written to the SMT and to the `SS` and a snapshot is created.


### Committing to an object without saving it

We identified use-cases, where modules will need to save an object commitment without storing an object itself. Sometimes clients are receiving complex objects, and they have no way to prove a correctness of that object without knowing the storage layout. For those use cases it would be easier to commit to the object without storing it directly.



## Consequences


### Backwards Compatibility

This ADR doesn't introduce any SDK level API changes.

We change the storage layout of the state machine, a storage migration and network upgrade is required to incorporate these changes.

### Positive

+ Decoupling state from state commitment introduce better engineering opportunities for further optimizations and better storage patterns.
+ Performance improvements.
+ Joining SMT based camp which has wider and proven adoption than IAVL. Example projects which decided on SMT: Ethereum2, Diem (Libra), Trillan, Tezos, LazyLedger.

### Negative

+ Storage migration
+ LL SMT doesn't support pruning - we will need to add and test that functionality.

### Neutral

+ Deprecating IAVL, which is one of the core proposals of Cosmos Whitepaper.


## Alternative designs.

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
