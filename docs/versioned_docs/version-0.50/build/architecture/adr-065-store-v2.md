# ADR-065: Store V2

## Changelog

* Feb 14, 2023: Initial Draft (@alexanderbez)

## Status

DRAFT

## Abstract

The storage and state primitives that Cosmos SDK based applications have used have
by and large not changed since the launch of the inaugural Cosmos Hub. The demands
and needs of Cosmos SDK based applications, from both developer and client UX
perspectives, have evolved and outgrown the ecosystem since these primitives
were first introduced.

Over time as these applications have gained significant adoption, many critical
shortcomings and flaws have been exposed in the state and storage primitives of
the Cosmos SDK.

In order to keep up with the evolving demands and needs of both clients and developers,
a major overhaul to these primitives are necessary.

## Context

The Cosmos SDK provides application developers with various storage primitives
for dealing with application state. Specifically, each module contains its own
merkle commitment data structure -- an IAVL tree. In this data structure, a module
can store and retrieve key-value pairs along with Merkle commitments, i.e. proofs,
to those key-value pairs indicating that they do or do not exist in the global
application state. This data structure is the base layer `KVStore`.

In addition, the SDK provides abstractions on top of this Merkle data structure.
Namely, a root multi-store (RMS) is a collection of each module's `KVStore`.
Through the RMS, the application can serve queries and provide proofs to clients
in addition to provide a module access to its own unique `KVStore` though the use
of `StoreKey`, which is an OCAP primitive.

There are further layers of abstraction that sit between the RMS and the underlying
IAVL `KVStore`. A `GasKVStore` is responsible for tracking gas IO consumption for
state machine reads and writes. A `CacheKVStore` is responsible for providing a
way to cache reads and buffer writes to make state transitions atomic, e.g.
transaction execution or governance proposal execution.

There are a few critical drawbacks to these layers of abstraction and the overall
design of storage in the Cosmos SDK:

* Since each module has its own IAVL `KVStore`, commitments are not [atomic](https://github.com/cosmos/cosmos-sdk/issues/14625)
    * Note, we can still allow modules to have their own IAVL `KVStore`, but the
      IAVL library will need to support the ability to pass a DB instance as an
      argument to various IAVL APIs.
* Since IAVL is responsible for both state storage and commitment, running an 
  archive node becomes increasingly expensive as disk space grows exponentially.
* As the size of a network increases, various performance bottlenecks start to
  emerge in many areas such as query performance, network upgrades, state
  migrations, and general application performance.
* Developer UX is poor as it does not allow application developers to experiment
  with different types of approaches to storage and commitments, along with the
  complications of many layers of abstractions referenced above.

See the [Storage Discussion](https://github.com/cosmos/cosmos-sdk/discussions/13545) for more information.

## Alternatives

There was a previous attempt to refactor the storage layer described in [ADR-040](./adr-040-storage-and-smt-state-commitments.md).
However, this approach mainly stems on the short comings of IAVL and various performance
issues around it. While there was a (partial) implementation of [ADR-040](./adr-040-storage-and-smt-state-commitments.md),
it was never adopted for a variety of reasons, such as the reliance on using an
SMT, which was more in a research phase, and some design choices that couldn't
be fully agreed upon, such as the snap-shotting mechanism that would result in
massive state bloat.

## Decision

We propose to build upon some of the great ideas introduced in [ADR-040](./adr-040-storage-and-smt-state-commitments.md),
while being a bit more flexible with the underlying implementations and overall
less intrusive. Specifically, we propose to:

* Separate the concerns of state commitment (**SC**), needed for consensus, and
  state storage (**SS**), needed for state machine and clients.
* Reduce layers of abstractions necessary between the RMS and underlying stores.
* Provide atomic module store commitments by providing a batch database object
  to core IAVL APIs.
* Reduce complexities in the `CacheKVStore` implementation while also improving
  performance<sup>[3]</sup>.

Furthermore, we will keep the IAVL is the backing [commitment](https://cryptography.fandom.com/wiki/Commitment_scheme)
store for the time being. While we might not fully settle on the use of IAVL in
the long term, we do not have strong empirical evidence to suggest a better
alternative. Given that the SDK provides interfaces for stores, it should be sufficient
to change the backing commitment store in the future should evidence arise to
warrant a better alternative. However there is promising work being done to IAVL
that should result in significant performance improvement <sup>[1,2]</sup>.

### Separating SS and SC

By separating SS and SC, it will allow for us to optimize against primary use cases
and access patterns to state. Specifically, The SS layer will be responsible for
direct access to data in the form of (key, value) pairs, whereas the SC layer (IAVL)
will be responsible for committing to data and providing Merkle proofs.

Note, the underlying physical storage database will be the same between both the
SS and SC layers. So to avoid collisions between (key, value) pairs, both layers
will be namespaced.

#### State Commitment (SC)

Given that the existing solution today acts as both SS and SC, we can simply
repurpose it to act solely as the SC layer without any significant changes to
access patterns or behavior. In other words, the entire collection of existing
IAVL-backed module `KVStore`s will act as the SC layer.

However, in order for the SC layer to remain lightweight and not duplicate a
majority of the data held in the SS layer, we encourage node operators to keep
tight pruning strategies.

#### State Storage (SS)

In the RMS, we will expose a *single* `KVStore` backed by the same physical
database that backs the SC layer. This `KVStore` will be explicitly namespaced
to avoid collisions and will act as the primary storage for (key, value) pairs.

While we most likely will continue the use of `cosmos-db`, or some local interface,
to allow for flexibility and iteration over preferred physical storage backends
as research and benchmarking continues. However, we propose to hardcode the use
of RocksDB as the primary physical storage backend.

Since the SS layer will be implemented as a `KVStore`, it will support the
following functionality:

* Range queries
* CRUD operations
* Historical queries and versioning
* Pruning

The RMS will keep track of all buffered writes using a dedicated and internal
`MemoryListener` for each `StoreKey`. For each block height, upon `Commit`, the
SS layer will write all buffered (key, value) pairs under a [RocksDB user-defined timestamp](https://github.com/facebook/rocksdb/wiki/User-defined-Timestamp-%28Experimental%29) column
family using the block height as the timestamp, which is an unsigned integer.
This will allow a client to fetch (key, value) pairs at historical and current
heights along with making iteration and range queries relatively performant as
the timestamp is the key suffix.

Note, we choose not to use a more general approach of allowing any embedded key/value
database, such as LevelDB or PebbleDB, using height key-prefixed keys to
effectively version state because most of these databases use variable length
keys which would effectively make actions likes iteration and range queries less
performant.

Since operators might want pruning strategies to differ in SS compared to SC,
e.g. having a very tight pruning strategy in SC while having a looser pruning
strategy for SS, we propose to introduce an additional pruning configuration,
with parameters that are identical to what exists in the SDK today, and allow
operators to control the pruning strategy of the SS layer independently of the
SC layer.

Note, the SC pruning strategy must be congruent with the operator's state sync
configuration. This is so as to allow state sync snapshots to execute successfully,
otherwise, a snapshot could be triggered on a height that is not available in SC.

#### State Sync

The state sync process should be largely unaffected by the separation of the SC
and SS layers. However, if a node syncs via state sync, the SS layer of the node
will not have the state synced height available, since the IAVL import process is
not setup in way to easily allow direct key/value insertion. A modification of
the IAVL import process would be necessary to facilitate having the state sync
height available.

Note, this is not problematic for the state machine itself because when a query
is made, the RMS will automatically direct the query correctly (see [Queries](#queries)).

#### Queries

To consolidate the query routing between both the SC and SS layers, we propose to
have a notion of a "query router" that is constructed in the RMS. This query router
will be supplied to each `KVStore` implementation. The query router will route
queries to either the SC layer or the SS layer based on a few parameters. If
`prove: true`, then the query must be routed to the SC layer. Otherwise, if the
query height is available in the SS layer, the query will be served from the SS
layer. Otherwise, we fall back on the SC layer.

If no height is provided, the SS layer will assume the latest height. The SS
layer will store a reverse index to lookup `LatestVersion -> timestamp(version)`
which is set on `Commit`.

#### Proofs

Since the SS layer is naturally a storage layer only, without any commitments
to (key, value) pairs, it cannot provide Merkle proofs to clients during queries.

Since the pruning strategy against the SC layer is configured by the operator,
we can therefore have the RMS route the query SC layer if the version exists and
`prove: true`. Otherwise, the query will fall back to the SS layer without a proof.

We could explore the idea of using state snapshots to rebuild an in-memory IAVL
tree in real time against a version closest to the one provided in the query.
However, it is not clear what the performance implications will be of this approach.

### Atomic Commitment

We propose to modify the existing IAVL APIs to accept a batch DB object instead
of relying on an internal batch object in `nodeDB`. Since each underlying IAVL
`KVStore` shares the same DB in the SC layer, this will allow commits to be
atomic.

Specifically, we propose to:

* Remove the `dbm.Batch` field from `nodeDB`
* Update the `SaveVersion` method of the `MutableTree` IAVL type to accept a batch object
* Update the `Commit` method of the `CommitKVStore` interface to accept a batch object
* Create a batch object in the RMS during `Commit` and pass this object to each
  `KVStore`
* Write the database batch after all stores have committed successfully

Note, this will require IAVL to be updated to not rely or assume on any batch
being present during `SaveVersion`.

## Consequences

As a result of a new store V2 package, we should expect to see improved performance
for queries and transactions due to the separation of concerns. We should also
expect to see improved developer UX around experimentation of commitment schemes
and storage backends for further performance, in addition to a reduced amount of
abstraction around KVStores making operations such as caching and state branching
more intuitive.

However, due to the proposed design, there are drawbacks around providing state
proofs for historical queries.

### Backwards Compatibility

This ADR proposes changes to the storage implementation in the Cosmos SDK through
an entirely new package. Interfaces may be borrowed and extended from existing
types that exist in `store`, but no existing implementations or interfaces will
be broken or modified.

### Positive

* Improved performance of independent SS and SC layers
* Reduced layers of abstraction making storage primitives easier to understand
* Atomic commitments for SC
* Redesign of storage types and interfaces will allow for greater experimentation
  such as different physical storage backends and different commitment schemes
  for different application modules

### Negative

* Providing proofs for historical state is challenging

### Neutral

* Keeping IAVL as the primary commitment data structure, although drastic
  performance improvements are being made

## Further Discussions

### Module Storage Control

Many modules store secondary indexes that are typically solely used to support
client queries, but are actually not needed for the state machine's state
transitions. What this means is that these indexes technically have no reason to
exist in the SC layer at all, as they take up unnecessary space. It is worth
exploring what an API would look like to allow modules to indicate what (key, value)
pairs they want to be persisted in the SC layer, implicitly indicating the SS
layer as well, as opposed to just persisting the (key, value) pair only in the
SS layer.

### Historical State Proofs

It is not clear what the importance or demand is within the community of providing
commitment proofs for historical state. While solutions can be devised such as
rebuilding trees on the fly based on state snapshots, it is not clear what the
performance implications are for such solutions.

### Physical DB Backends

This ADR proposes usage of RocksDB to utilize user-defined timestamps as a
versioning mechanism. However, other physical DB backends are available that may
offer alternative ways to implement versioning while also providing performance
improvements over RocksDB. E.g. PebbleDB supports MVCC timestamps as well, but
we'll need to explore how PebbleDB handles compaction and state growth over time.

## References

* [1] https://github.com/cosmos/iavl/pull/676
* [2] https://github.com/cosmos/iavl/pull/664
* [3] https://github.com/cosmos/cosmos-sdk/issues/14990
