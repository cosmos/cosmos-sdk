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

* 

## Alternatives

> TODO: Reference various resources on the design and implementation of the original
> store v2 based on the work from Vulcanize.

## Decision

> This section describes our response to these forces. It is stated in full sentences, with active voice. "We will ..."
> {decision body}

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR.

## References

* https://github.com/cosmos/cosmos-sdk/discussions/13545
* https://github.com/cosmos/cosmos-sdk/issues/12986
