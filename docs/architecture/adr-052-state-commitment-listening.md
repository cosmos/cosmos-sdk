# ADR-051: State Commitment Listening

## Changelog

- Feb 18th, 2022: Initial Draft

## Status

DRAFT

## Abstract

This ADR describes the features necessary to enable real-time streaming of updated state commitment (SMT) nodes to an
external data consumer.

## Context

ADR-038 introduced features that enable listening to state changes of individual KVStores, by emitting key value pairs
as they are updated. For many applications this is a sufficient or even preferable method for listening to state changes,
as it allows for the selective listening of updated KVPairs in specific modules and does not impose the additional overhead
necessary to stream all the updates at the state commitment layer. But, for applications that wish to extract the entire Merkle
data structure from a CosmosSDK blockchain in order to retain the provability of state changes we need additional features
for listening to all the updates at the state commitment layer.

### Eventual state
The complete state as it exists at a given height `n`: the entire SMT and all the values it references.

### Differential state
The subset of the state at `n` that includes only the nodes and values that were updated during the transition from `n-1` to `n`.

Cumulatively, all the statediffs from chain genesis to the current block `n` will materialize the
entire state at `n` as well as all historical states below `n`. Additionally, these difference sets can be
(relatively quickly) iterated to produce useful indexes and relations around and between tree nodes and other
Tendermint/Cosmos data structures.

### Block witness
The subset of the state at `n-1` that includes on the nodes and values required to perform the state transition to `n`.

## Decision

There are at least four potential approaches for extracting the entire state + state commitment data structure from a
CosmosSDK blockchain to an external destination. All three of these will be discussed below with their pros and cons as
a further explanation of the work that remains to be done for each of these approaches.

### Replay of KVPairs
One way the entire Merkle data structure can be extracted to an external system would be to replay all the KVPairs
listened to using the features implemented in ADR-038. If we listen to every KVPair emitted from every module in this
manner, and then reinsert them into an external SMT the SMT can be recapitulated in full.

Pros:
* No additional changes needed within the SDK

Cons:
* Requires reinserting every KVPair into an external SMT and re-materializing the entire state commitment data structure,
duplicating work that already occurs within the Cosmos blockchain.
* If at any point any single KVPair is missed a significant number of the external SMT nodes materialized in this manner (including the root)
would not match the canonical nodes, and it would be extremely difficult to identify what KVPair is missing in order to repair
the SMT.
* If we want to build a path index around these external SMT nodes, we require an additional service or another
implementation of the SMT that enables the mapping of this path information during or after external SMT node materialization
  (e.g. an implementation of SMT that uses Postgres as its backing MapStore and generates useful indexes/relations around
the SMT nodes during or after they are inserted into Postgres)
* Does not provide a means to horizontally scale the extraction and processing of historical data (only way to retrieve

We believe this approach is inadequate from both a performance perspective due to the need to replay every KVPair
insert/update/delete and re-materialize SMT nodes that have already been materialized in the SMT that backs the blockchain
and also from a feature perspective as it does not provide a direct path forward for generating useful indexes and relations
around the state commitment data structure. Additionally, it does not provide a means of extracting historical state
commitment data.

### Database versioned snapshot iteration
We can potentially leverage the versioned snapshot architecture of the databases (Badger and RocksDB) underpinning the
SMT to efficiently extract (only) the updated nodes at a specific block. The SDK uses Badger and RocksDB transactions as
the `MapStore` interface that the SMT implementation writes to. These transactions create a database with versioned snapshots,
these snapshots will contain (reference) all the SMT nodes (`hash(node) => node)` kv mappings) that exist in the SMT at that
height. These snapshots may also include all the other data written to disk, e.g. the B1 and B2 buckets maintained at the
SDK layer and the `hash(key) => value` mapping maintained by the SMT.

If we can devise an efficient way of extracting only the node information from the snapshot
(e.g. stipulate a prefix for this keyspace) then we could extract the nodes at height `n` and `n-1`, find their intersection,
and remove this intersection from the nodes at `n` to produce the difference set.

Pros:
* No additional work within the SDK.
* Provides a means to generate statediff objects at any arbitrary height, enabling the processing of historical state
commitment data.
* Can horizontally scale processing of historical data using this approach.

Cons:
* Depends on the database implementation underpinning the SMT to support iterateable versioned snapshots.
* Need to update the SMT implementation to introduce keyspace prefixing for the `hash(node) => node` bucket.
* Unclear at this stage how performant this will actually be, the sets of nodes at `n` and `n-1` will be *very* large
and finding their intersection will be expensive.
* We are missing the path context provided by generating this difference set during a tree difference iteration.

This approach is insufficient for any scenario where we want to extract and maintain path information for the SMT
nodes, as this context is missing when iterating the flat database keyspace.

### SMT difference iteration
Another approach for extracting the entire state commitment data structure to an external system would be to implement
an SMT node difference iterator. This approach would be agnostic towards the backing database.
The difference iterator is a tree iterator that simultaneously iterates the trees at height `n` and `n-1` in
order to traverse only the nodes which differ between the two trees. In this manner, it produces and emits a "statediff"
objects for the SMT at height `n`.

Pros:
* Underlying DB agnostic
* Does not require any changes to the underlying SMT implementation.
* More performant than replaying KVPairs to maintain an external materialization of the SMT.
* Provides a means to generate statediff objects at any arbitrary height, enabling the processing of historical state
commitment data.
* Can horizontally scale processing of historical data using this approach.
* Difference iterator is aware of the path of the SMT nodes during iteration. This allows it to generate an index around SMT node path, this path association is useful for efficient generation
of proofs (enables us to select all the nodes along a path from the root to a specific child with a single query, rather
than having to iterate down the tree using multiple database lookups).
* Additional middleware/hooks can be plugged into the difference iterator and/or statediffing service in order to generate useful
indexes or associate additional metadata with the nodes (e.g. IPLD Middleware, to-be proposed in a following ADR, could associate
the appropriate multicodec types with the processed nodes and values).

Cons:
* Requires the implementation of an SMT node difference iterator and statediffing services that use it.
* This node difference iterator needs additional (optional) features implemented within it in order to produce useful
indexes/relations around the statediff nodes during the initial iteration. Alternatively, the statediff object could be
re-iterated in a secondary system to generate these indexes, but requiring a second iteration of the statediff nodes
has significant performance overhead.
* In order to generate the difference set, we need access to state at both A and A-1.
* Because neither BadgerDB nor RocksDB support concurrent read access from multiple system processes, this difference
iteration will have to occur from within the context of the CosmosSDK blockchain process (for real-time data at head,
for historical data we can create snapshots of the database and iterate them in separate processes).

This approach enables mapping of path to the nodes we extract and we believe would be sufficient from a performance perspective and because it provides the ability to
process historical data this SMT node difference iterator will be a necessary tool for some use cases. For this reason,
we propose this approach for historical state processing while using the below approach for real-time processing at
the head of the chain.

### SMT cache/commit cycles with node flushing capabilities

Pros:
* Most performant approach. Not only is it the most performant approach for extracting all SMT nodes from a CosmosSDK
blockchain in real-time, because it requires no additional tree iteration or node materialization, but additionally the
cache/commit feature necessary to realize this approach would reduce the number of disk operations the SMT must perform in its
role in the SDK.
* Additional middleware can be wrapped around the channels used to flush the cached SMT nodes in order to generate useful
indexes or associate additional metadata with the nodes (e.g. IPLD Middleware, to-be proposed in a following ADR, could associate
the appropriate multicodec types with the streamed nodes and values).

Cons:
* Requires changes to the underlying SMT implementation.
* Requires changes to the utilization pattern of the SMT in the SDK.
* The cache/commit feature introduces additional memory overhead and complexity for the SMT.
* The flush feature introduces additional overhead and complexity for the SMT.
* If we wish to associate node path with the SMT nodes while they are streamed out in this capacity, additional implementation
complexity and overhead is introduced into the SMT.

We believe this is the best approach for extracing the full SMT data structure at the head of the chain in real-time.

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

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.

## References

* {reference link}
