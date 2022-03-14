# ADR-052: State Commitment Listening

## Changelog

- March 14th, 2022: Initial Draft

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
data structure from a CosmosSDK blockchain in order to retain the provability of state we need additional features
for listening to all the updates at the state commitment layer.

### Eventual state
The complete state as it exists at a given height `n`: the entire SMT and all the values it references.

### Differential state
The subset of the state at `n` that includes only the nodes and values that were updated during the transition from
`n-1` to `n`- aka a "statediff".

Cumulatively, all the statediffs from chain genesis to the current block `n` will materialize the
eventual state at `n` as well as all historical eventual states below `n`. Additionally, these difference sets can be
(relatively quickly) iterated to produce useful indexes and relations around and between tree nodes/values and other
Tendermint/Cosmos data structures.

## Decision

There are at least four potential approaches for extracting the entire state + state commitment data structure from a
CosmosSDK blockchain to an external destination. All four are discussed below with their pros and cons as well as
an explanation of the work that remains to be done to support each of these approaches.

### Replay of KVPairs
One way the entire Merkle data structure can be extracted to an external system would be to replay all the KVPairs
listened to using the features implemented in ADR-038. If we listen to every KVPair emitted from every module in this
manner, and then reinsert them into an external SMT the SMT can be recapitulated in full.

Pros:
* No additional changes needed within the SDK.

Cons:
* Requires reinserting every KVPair into an external SMT and re-materializing the entire state commitment data structure,
duplicating work that already occurs within the Cosmos blockchain.
* KVPairs must be replayed in the correct order.
* If at any point any single KVPair is missed a significant number of the external SMT nodes materialized in this manner (including the root)
would not match the canonical nodes, and it would be extremely difficult to identify what KVPair is missing in order to repair
the SMT.
* If we want to build a path index around these external SMT nodes, we require an additional service or another
implementation of the SMT that enables the mapping of this path information during or after external SMT node materialization.
* Does not provide a means to horizontally scale the extraction and processing of historical state, to extract historical
state with this approach we must sync a node from genesis and replay all KVPairs in order which cannot be readily parallelized.

We believe this approach is inadequate from both a performance perspective due to the need to replay every KVPair
insert/update/delete and re-materialize SMT nodes that have already been materialized in the SMT that backs the blockchain
and also from a feature perspective as it does not provide a direct path forward for generating a path index around all
state commitment nodes. Additionally, it does not provide a parallelizeable means of extracting historical state
commitment data.

### Database versioned snapshot iteration
We can, in theory, leverage the versioned snapshot architecture of the databases (Badger and RocksDB) underpinning the
SMT to extract the updated nodes at a specific block. The SDK uses Badger and RocksDB transactions as
the `MapStore` interface that the SMT implementation writes to. These transactions create a database with versioned snapshots,
these snapshots will contain (reference) all the SMT nodes (`hash(node) => node` kv mappings) that exist in the SMT at that
height. These snapshots may also include all the other data written to disk, e.g. the B1 and B2 buckets maintained at the
SDK layer and the `hash(key) => value` mapping maintained by the SMT.

If we can devise an efficient way of extracting only the node information from the snapshot
(e.g. stipulate a prefix for this keyspace) then we could extract all the nodes at height `n` and `n-1`, find their intersection,
and remove this intersection from the nodes at `n` to produce the difference set at `n`.

Pros:
* No additional changes needed within the SDK.
* Provides a means to generate statediff objects at any arbitrary height.
* Can horizontally scale processing of historical data using this approach by creating snapshots of the database,
fs overlays of these snapshots, and iterating the state in parallel across separate processes (across separate block ranges).
* Requires fewer DB round-trips compared to difference iteration approach (below).

Cons:
* In order to generate the difference set, we need access to state at both `n` and `n-1`.
* Depends on the database implementation underpinning the SMT to support iterateable versioned snapshots.
* Need to update the SMT implementation to introduce keyspace prefixing for the `hash(node) => node` bucket.
* Performance is likely prohibitive. The sets of nodes at `n` and `n-1` will be *very* large, loading these sets
and finding their intersection will be expensive. Time complexity of finding intersection: `O(x+y)` where `x`
is number of nodes at `n` and `y`is number of nodes at `n-1`. Memory intensive- ideally we would load the entire node
sets at `n` and `n-1` into memory in order to find their intersection.
* We are missing the path context provided by generating this difference set during a tree difference iteration- if we
want to build a path index around the SMT nodes, we require an additional process.

This approach is insufficient for any scenario where we want to extract and maintain path information for the SMT
nodes, as this context is missing when iterating the flat database keyspace. Time complexity is linear with respect
to the number of nodes in the SMT. For these reasons, we propose [SMT difference iteration](#smt-difference-iteration)
for the extraction of historical state commitment data.

### SMT difference iteration
Another approach for extracting the entire state commitment data structure to an external system would be to implement
an SMT node difference iterator. This approach would be agnostic towards the backing database.
The difference iterator is a tree iterator that simultaneously iterates the trees at height `n` and `n-1` in
order to traverse only the nodes which differ between the two trees. In this manner, it can produce and emits a "statediff"
objects for the SMT at height `n`.

At a high level this approach has three steps:
1. Implement basic iterator for SMT.
2. Implement difference iterator for SMT, uses above.
3. Implement standalone statediffing service that uses the above difference iterator to generate and process statediffs.

Pros:
* Underlying DB agnostic.
* Does not require any changes to the underlying SMT implementation.
* More performant than replaying KVPairs to maintain an external materialization of the SMT.
* Time complexity is linear with respect to the number of nodes in the *differential state* of the SMT, not the entire
SMT.
* Provides a means to generate statediff objects at any arbitrary height.
* Can horizontally scale processing of historical data using this approach by creating snapshots of the database,
fs overlays of these snapshots, and iterating the state in parallel across separate processes (across separate block ranges).
* Difference iterator is aware of the path of the SMT nodes during iteration. This allows it to generate an index around
SMT node path, this path association is useful for efficient generation
of proofs (enables us to select all the nodes along a path from the root to a specific child with a single query, rather
than having to iterate down the tree using multiple database lookups).
* Additional middleware/hooks can be plugged into the difference iterator and/or statediffing service in order to generate useful
indexes or associate additional metadata with the nodes (e.g. IPLD Middleware, to-be proposed in a following ADR, could associate
the appropriate multicodec types with the processed nodes and values).

Cons:
* Requires the implementation of an SMT node difference iterator and a statediffing services that use it.
* In order to generate the difference set, we need access to state at both `n` and `n-1`.
* Because neither Badger nor RocksDB support concurrent read access from multiple system processes, if using this approach
to process real-time data at the head of the chain we would need to run this difference iteration from within the context
of the CosmosSDK blockchain system process.
* Requires a lot of round-trips to the DB (to iterate the difference set).

This approach enables mapping of node path to the nodes we extract and enables the parallelizable processing of historical state
in a manner that scales linearly with respect to the number of nodes in a difference set. We propose this approach for
historical state commitment processing while using the below approach for real-time processing at the head of the chain.

### SMT cache/commit cycles with node flushing capabilities
We can implement a cache/commit wrapper around the existing SMT implementation that allows us to flush updated nodes
(and values) from the SMT at the end of every block to produce the difference set for that block.

To do so we need to:
1. Update the SMT interface with a new method, `Commit() error`.
2. Create a cache/commit wrapper for the SMT.
3. Create an SMT constructor that allows us to pass a listening channel to the cache/commit wrapper.
4. Tie this listening channel into the SDK in a capacity analogous to the plugin-based `StreamingService` introduced in ADR-038.

Pros:
* Most performant approach. Not only is it the most performant approach for extracting all SMT nodes from a CosmosSDK
blockchain in real-time- because it requires no additional tree iteration or node materialization- but additionally the
cache/commit feature necessary to realize this approach may reduce the number of disk operations the SMT must perform in its
role in the SDK and provide a general performance improvement to the SMT based `MultiStore`.
* Additional middleware can be wrapped around the channels used to flush the cached SMT nodes in order to generate useful
indexes or associate additional metadata with the nodes (e.g. IPLD Middleware, to-be proposed in a following ADR, could associate
the appropriate multicodec types with the streamed nodes and values).

Cons:
* Requires changes to the underlying SMT interface.
* Requires changes to the utilization pattern of the SMT in the SDK.
* The cache/commit feature introduces additional memory overhead and complexity for the SMT.
* The flush feature introduces additional overhead and complexity for the SMT.
* If we wish to associate node path with the SMT nodes while they are streamed out in this capacity, additional implementation
complexity and overhead is introduced into the SMT (we can no longer get away with using a simple wrapper around the existing SMT
implementation).
* Does not provide a means to horizontally scale the extraction and processing of historical state, to extract historical
state with this approach we must sync a node from genesis and listen to all SMT node emissions.

We believe this is the best approach for extracting the full SMT data structure at the head of the chain in real-time. In
combination with [SMT difference iteration](#smt-difference-iteration) for historical data

## Consequences

### SMT difference iteration
No direct impact on the SDK, as this approach will be introduced as an entirely optional auxiliary service and does not
require changing the SMT interface or implementation.

#### Backwards Compatibility
Does not impact backwards compatibility as no changes are affected to the SDK directly.

#### Positive
Services capable of extracting the entire historical SMT state to an external destination. This enables an external
system to provide proofs for all Cosmos state.

#### Negative

#### Neutral

### SMT cache/commit cycles with node flushing capabilities
Requires updating the SMT interface used by the SDK, and updating the SDK to use this updated interface. A single
`Commit() error` method needs to be added, but none of the existing methods are altered. Since the `MultiStore` already
operates in a cache/commit cycle, tying this commit interface into the SDK will not be very intrusive. Similarly, we can
reuse the existing plugin-based `StreamingService` framework introduced in ADR-038 for wiring the SMT cache/commit listener
into the SDK.

#### Backwards Compatibility
Updating the SMT interface to support a flushable cache/commit cycle at the SMT level does not break backwards compatibility
since the rest of the SMT interface is unchanged (aka we can still use the existing SMT access pattern).

#### Positive
Services capable of streaming the entire SMT state in realtime to an external destination. This enables an external
system to provide proofs for all Cosmos state.

#### Negative

#### Neutral

## Further Discussions

TODO

## Test Cases [optional]

TODO

## References

* POC SMT cache/commit listener wrapper: https://github.com/vulcanize/smt/blob/cache_listener_wrap/cache_listening_mapstore.go
