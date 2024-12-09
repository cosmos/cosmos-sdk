# ADR-065: Store V2

## Changelog

* Feb 14, 2023: Initial Draft (@alexanderbez)
* Dec 21, 2023: Updates after implementation (@alexanderbez)

## Status

ACCEPTED

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

* Reduce layers of abstractions necessary between the RMS and underlying stores.
* Remove unnecessary store types and implementations such as `CacheKVStore`.
* Remove the branching logic from the store package.
* Ensure the `RootStore` interface remains as lightweight as possible.
* Allow application developers to easily swap out SC backends.

Furthermore, we will keep IAVL as the default [SC](https://cryptography.fandom.com/wiki/Commitment_scheme)
backend for the time being. While we might not fully settle on the use of IAVL in
the long term, we do not have strong empirical evidence to suggest a better
alternative. Given that the SDK provides interfaces for stores, it should be sufficient
to change the backing commitment store in the future should evidence arise to
warrant a better alternative. However there is promising work being done to IAVL
that should result in significant performance improvement <sup>[1,2]</sup>.

Note, we will provide applications with the ability to use IAVL v1,  IAVL v2 and MemIAVL as
either SC backend, with the latter showing extremely promising performance improvements
over IAVL v0 and v1, at the cost of a state migration.


### State Commitment (SC)

A foremost design goal is that SC backends should be easily swappable, i.e. not
necessarily IAVL.  To this end, the scope of SC has been reduced, it must only:

* Provide a stateful root app hash for height h resulting from applying a batch
  of key-value set/deletes to height h-1.
* Fulfill (though not necessarily provide) historical proofs for all heights < h.
* Provide an API for snapshot create/restore to fulfill state sync requests.

An SC implementation may choose not to provide historical proofs past height h - n (n can be 0)
due to the time and space constraints, but since store v2 defines an API for historical
proofs there should be at least one configuration of a given SC backend which
supports this.


#### State Sync

The state sync process should be largely unaffected by the separation of the SC
and SS layers. However, if a node syncs via state sync, the SS layer of the node
will not have the state synced height available, since the IAVL import process is
not setup in way to easily allow direct key/value insertion.

We propose a simple `SnapshotManager` that consumes and produces snapshots. SC 
backends will be responsible for providing a snapshot of the state at a given 
height and both SS and SC consume snapshots to restore state.

#### RootStore

We will define a `RootStore` interface and default implementation that will be
the primary interface for the application to interact with. The `RootStore` will
be responsible for housing SS and SC backends. Specifically, a `RootStore` will
provide the following functionality:

* Manage commitment of state 
* Provide modules access to state
* Query delegation (i.e. get a value for a <key, height> tuple)
* Providing commitment proofs

#### Store Keys

Naturally, if a single SC tree is used in all RootStore implementations, then the
notion of a store key becomes entirely useless. However, we cannot dictate or
predicate how all applications will implement their RooStore (if they choose to).

Since an app can choose to have multiple SC trees, we need to keep the notion of
store keys. Unlike store v1, we represent store keys as simple strings as opposed
to concrete types to provide OCAP functionality. The store key strings act to
solely provide key prefixing/namespacing functionality for modules.

#### Proofs

Providing a `CommitmentOp` type, will be the responsibility of the SC backend. Retrieving proofs will be done through the a `RootStore`, which will internally route the request to the SC backend.

#### Commitment

Before ABCI 2.0, specifically before `FinalizeBlock` was introduced, the flow of state
commitment in BaseApp was defined by writes being written to the `RootMultiStore`
and then a single Commit call on the `RootMultiStore` during the ABCI Commit method.

With the advent of ABCI 2.0, the commitment flow has now changed to `WorkingHash` being
called during `FinalizeBlock` and then Commit being called on ABCI Commit. Note,
`WorkingHash` does not actually commit state to disk, but rather computes an
uncommitted work-in-progress hash, which is returned in `FinalizeBlock`. Then,
during the ABCI Commit phase, the state is finally flushed to disk.

In store v2, we must respect this flow. Thus, a caller is expected to call `WorkingHash`
during `FinalizeBlock`, which takes the latest changeset in the `RootStore`,
writes that to the SC tree in a single batch and returns a hash. Finally, during
the ABCI Commit phase, we call `Commit` on the `RootStore` which commits the SC
tree and flushes the changeset to the SS backend.

## Consequences

As a result of a new store V2 package, we should expect to see improved performance
for queries and transactions due to the separation of concerns. We should also
expect to see improved developer UX around experimentation of commitment schemes
and storage backends for further performance, in addition to a reduced amount of
abstraction around KVStores making operations such as caching and state branching
more intuitive.

### Backwards Compatibility

This ADR proposes changes to the storage implementation in the Cosmos SDK through
an entirely new package. Interfaces may be borrowed and extended from existing
types that exist in `store`, but no existing implementations or interfaces will
be broken or modified.

### Positive

* Improved performance of SC layers
* Reduced layers of abstraction making storage primitives easier to understand
* Redesign of storage types and interfaces will allow for greater experimentation
  such as different physical storage backends and different commitment schemes
  for different application modules

### Negative

### Neutral

* Removal of OCAP-based store keys in favor of simple strings for state retrieval
  and name-spacing. We consider this neutral as removal of OCAP functionality can
  be seen as a negative, however, we're simply moving the OCAP functionality upstream
  to the KVStore service. The SS and SC layers shouldn't have to concern themselves
  with OCAP responsibilities.
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

## References

* [1] https://github.com/cosmos/iavl/pull/676
* [2] https://github.com/cosmos/iavl/pull/664
* [3] https://github.com/cosmos/cosmos-sdk/issues/14990
* [4] https://docs.google.com/document/d/e/2PACX-1vSCFfXZm2vsRsACOPoxGqysMaUg7jY833LwR3YyjA1S3FNHfXRiJor-qLjzx833TavLXLPSIcFZJhyh/pub
