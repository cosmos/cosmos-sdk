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
Namely, a root multi-store is a collection of each module's `KVStore`. Through
the root multi-store, the application can serve queries and provide proofs to
clients in addition to provide a module access to its own unique `KVStore` though
the use of `StoreKey`, which is an OCAP primitive.

There are further layers of abstraction that sit between the root multi-store and
the underlying IAVL `KVStore`. A `GasKVStore` is responsible for tracking gas
IO consumption for state machine reads and writes. A `CacheKVStore` is responsible
for providing a way to cache reads and buffer writes to make state transitions
atomic, e.g. transaction execution or governance proposal execution.

There are a few critical drawbacks to these layers of abstraction and the overall
design of storage in the Cosmos SDK:

* Since each module has its own IAVL `KVStore`, commitments are not atomic ([ref](https://github.com/cosmos/cosmos-sdk/issues/14625))
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

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

> {positive consequences}

### Negative

> {negative consequences}

### Neutral

> {neutral consequences}

## Further Discussions

> While an ADR is in the DRAFT or PROPOSED stage, this section should contain a
> summary of issues to be solved in future iterations (usually referencing comments
> from a pull-request discussion).
> 
> Later, this section can optionally list ideas or improvements the author or
> reviewers found during the analysis of this ADR.

## References

* https://github.com/cosmos/cosmos-sdk/discussions/13545
* https://github.com/cosmos/cosmos-sdk/issues/12986
