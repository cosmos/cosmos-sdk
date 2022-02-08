# ADR-050: Arbitrary Protobuf IPLD Support Scheme

## Changelog

- Feb 14th, 2022: Initial Draft

## Status

DRAFT

## Abstract
This ADR describes a generic IPLD content addressing scheme for the arbitrary protobuf types stored in SMT referenced state storage.

## Context

Because all values stored in the state of a ComsosSDK chain are protobuf encoded, we are presented with a unique
opportunity to provide extremely rich IPLD content-typing information for the arbitrary types stored in
(referenced from) the SMT in a generic fashion.

### SDK context

The SDK stores values in state storage in a protobuf encoded format. Each module defines their own custom types in
.proto definitions and these types are registered with a ProtoCodec that manages the marshalling and unmarshalling of these
types to and from the binary format that is persisted to disk. As such, an indefinite/unbounded number of protobuf types
need to be supported within the Cosmos ecosystem at large.

Rather than needing to register new content types and implement custom codecs for every Cosmos protobuf type we
wish to support as IPLD it would be useful to have a generic means of supporting arbitrary protobuf types. This would
open the doors for some interesting features and tools.
For example: a universal (and richly typing) IPLD block explorer for all/any blockchains in the Cosmos ecosystem that
doesn't require custom integrations for every type it explores and represents.

### IPLD context

[IPLD](https://ipld.io/docs/) stands for InterPlanetary Linked Data.
IPLD is a self-describing data model for arbitrary hash-linked data.
IPLD enables a universal namespace for all Merkle data structures and the representation, access, and exploration of these
data structures using the same set of tools. IPLD is the data model for IPFS, Filecoin, and a growing number of other
distributed protocols.

Some core concepts of IPLD are described below.

### IPLD Objects

IPLD objects are abstract representations of nodes in some hash-linked data structure. In this representation,
the hash-links that exist in the native representation of the underlying data structure
(e.g. the SHA-2-256 hashes of an SMT leaf node) are transformed into CIDs- hash-links that describe the
content they reference.

### IPLD Blocks

IPLD blocks are the binary encoding of an IPLD object that is persisted on disk.
They contain (or are) the binary content that is hashed
and referenced by CID. For blockchains, this binary encoding of an IPLD object will correspond to the consensus encoding
of the represented object. E.g. for Ethereum headers, the IPLD block is the RLP encoding of the header.

For Tendermint-Cosmos blockchains, the IPLD blocks are the consensus binary encodings
of the Merkle/SMT nodes of the various Merkle trees and the values they reference.

### CIDs

CIDs are self-describing content-addressed identifiers for IPLD blocks. They are composed of a content-hash of an
IPLD block prefixed with bytes that identify the hashing algorithm applied on that block to produce that 
hash (multihash-prefix), a byte prefix that identifies the content type of the IPLD 
block (multicodec-content-type), a prefix that identifies the version of the CID itself (multicodec-version),
and a prefix that identifies the base encoding of the CID (multibase).

`<cidv1> ::= <multibase><multicodec-version><multicodec-content-type><multihash-preifx><content-hash>`

### IPLD Codecs

IPLD blocks are converted to and from in-memory IPLD objects using IPLD codecs, which marshal/unmarshal an IPLD object
to/from the binary IPLD block encoding. Just as `multihash-content-address` maps to a specific hash function or
algorithm, `multicodec-content-type` maps to a specific codec for encoding and decoding the content. These mappings are
maintained in a codec registry. These codecs contain custom logic that resolve native hash-links to CIDs.

## Decision

We need to define three new IPLD `multicodec-content-type`s and implement their codecs
in order to support rich content-typing of arbitrary protobuf types.

### New IPLD content types

The three new content types are described below, codecs for these types will need to be implemented. For Go, these
implementations will work with [go-ipld-prime](https://github.com/ipld/go-ipld-prime) and will be added to the existing
codecs found in https://github.com/vulcanize/go-codec-dagcosmos. Eventually we may also wish to implement these codecs
in JS/TS (as https://github.com/vulcanize/ts-dag-eth is to https://github.com/vulcanize/go-codec-dageth).

### Self-Describing Protobuf Multicodec-Content-Type

Define a new `multicodec-content-type` for a
[self-describing protobuf message](https://developers.google.com/protocol-buffers/docs/techniques?authuser=2#self-description):
`self-describing-protobuf`. This `multicodec-content-type` will be used to create CID references to the protobuf encodings of such
self-describing messages.

CIDv1 for a `self-describing-protobuf` block:

`<multibase><multicodec-version><self-describing-protobuf-multicodec-content-type><SHA-2-256-mh-prefix><SHA-2-256(protobuf-encoded-self-describing-message)>`

### Typed-Protobuf Multicodec-Content-Type

Define a new `multicodec-content-type`- `typed-protobuf`- that specifies that the first 32 bytes of the hash-linked IPLD
block are an SHA-2-256 hash-link to a self-describing protobuf message (to a `self-describing-protobuf` IPLD block, as described above)
that represents the contents of the .proto file that compiles into the protobuf message that the remaining bytes of the
hash-linked IPLD object can be unmarshalled into.

In this `multicodec-content-type` specification we also stipulate that the content-hash referencing the IPLD block
only includes as digest input the protobuf encoded bytes for the storage value whereas the `self-describing-protobuf` hash-link prefix
is excluded.
This is a significant deviation from previous IPLD codecs, as it means the content-hash is not a hash of *all* of the
content it references, but is necessary to maintain the native consensus hashes and hash-to-content mappings.

In other-words, the `content-hash` in a `typed-protobuf` CID will be
`hash(<referenced-protobuf-encoded-value>)` (when using SHA-2-256 multihash type, this matches the hash we see in the Cosmos SMT)
instead of `hash(<self-describing-protobuf-hash><referenced-protobuf-encoded-value>)`.

Otherwise `content-hash` would not match the hashes we see natively in the Tendermint+Cosmos Merkle DAG, and we
would not be able to directly derive CIDs from them.

CIDv1 for a `typed-protobuf` block:

`<multibase><multicodec-version><typed-protobuf-multicodec-content-type><SHA-2-256-mh-prefix><SHA-2-256(referenced-protobuf-encoded-value)>`

Another major deviation this necessitates is the requirement of an IPLD retrieval, unmarshalling, and protobuf
compilation step in order to fully unmarshal the `referenced-protobuf-encoded-value` stored in a `typed-protobuf` block.

The algorithm will look like:

1. Fetch the `typed-protobuf` block binary using a `typed-protobuf` CID
2. Decode the `self-describing-protobuf` CID from the block's hash-link prefix
3. Use that `self-describing-protobuf` CID to fetch the `self-describing-protobuf` block binary
4. Decode that binary into a self-describing protobuf message
5. Use that self-describing protobuf message to create and compile the canonical proto message for the `referenced-protobuf-encoded-value`
6. Decode the `referenced-protobuf-encoded-value` binary into the proto message type compiled in step 5

### Protobuf-SMT Multicodec-Content-Type

Define a new `multicodec-content-type`- `protobuf-smt`- for SMT nodes wherein the IPLD object representation of a leaf nodes converts
the canonical `content-hash` (i.e. `SHA-2-256(referenced-protobuf-encoded-value)`) into a `typed-protobuf` CID, as
described above. Intermediate node representation is no different from standard SMT representation. The codec can
make use of the existing byte prefix to differentiate between a leaf and intermediate node, so we do not need a separate
codec for intermediate and leaf nodes.

### Additional work
#### IPLD aware protobuf compiler

For simple protobuf definitions which have no external dependencies on other protobuf definitions, this work will not
be necessary- a single standalone self-describing protobuf message will be enough to generate and compile the protobuf
types necessary for unpacking an arbitrary protobuf value. On the other hand, if we have more complex protobuf
definitions with external dependencies (that we cannot inline) we need some way of resolving these dependencies within
the context of the IPLD codec performing the object unmarshalling. To this end, we propose to create an IPLD
aware/compatible protobuf compiler that is capable of rendering and resolving dependency trees as IPLD.

This could be written as an existing protobuf compiler plugin that allows .proto files to import other proto packages
using `self-describing-protobuf` CIDs as import identifiers. In this way, protobuf dependency trees could be represented
as IPLD DAGs and IPFS could be used as a hash-linked registry for all protobuf types.

Further specification and discussion of this is, perhaps, outside the content of the SDK.

#### IPLD middleware for the SDK

In order to leverage this model for a Cosmos blockchain, features need to be introduced into the SDK (or as auxiliary
services) for

1. Generating the self-describing messages from the .proto definitions of state storage types in modules
2. Publishing or exposing these as `self-describing-protobuf` IPLD blocks
3. Mapping state storage objects- as they are streamed out using the features in ADR-038 and ADR-XXX- to their
respective `self-describing-protobuf` 

The above process is very similar, at a high level, to the ORM work that has already been done in the SDK.

These steps will be discussed further in a subsequent ADR (referred to as ADR-YYY for now).

## Consequences

There are no direct consequences on the other components of the SDK, as everything discussed here is entirely optional.
In fact, at this stage, everything is theoretical and only exists as an abstract data model with no integration into
the SDK. This model does not impose or require any changes on/to Cosmos blockchain state, it is an abstract representation
of that state which can be materialized in external systems (such as IPFS).

Nonetheless, there are consequences for defining and attempting to standardize a new abstract data model for Cosmos state.
The approval of this model should only occur once it has been determined to a satisfactory degree that it is the best
available model for representing arbitrary Cosmos state as IPLD.

In order to introduce generic support for arbitrary protobuf types in state storage, the approach proposed here deviates
from previous IPLD codecs and content-hash-to-content standards. For better, or worse, this could set new precedents for
IPLD that need to be considered within the context of the greater IPLD ecosystem. We propose that this work be seen as 
extensions to the standard CID and IPLD concepts and believe that these types of deviations would provide improved
flexibility/generalizability of the IPLD model in other contexts as well.

### Backwards Compatibility

No backwards incompatibilities.

### Positive

Define a generic IPLD model for arbitrary Cosmos state. This model will enable universal IPLD integration for any and
all "canonical" Cosmos blockchain (e.g. if they don't use SMT or don't require Protobuf encoding of state values, 
this falls apart), improving interoperability with various protocols that leverage IPLD. The concrete implementation of
the tools to (optionally) integrate and leverage this model within the SDK will be proposed and discussed in a later ADR.

### Negative

Code and documentation bloat/bandwidth.

Because of the deviations we make from the current historical precedence for IPLD, we suspect upstreaming registration
and support of these codecs into Protocol Labs repositories will be a complicated process.

### Neutral

Nothing comes to mind.

## Further Discussions

We need to complete the proposals for ADR-XXX (SMT Node State Streaming/Listening Features) and ADR-YYY (IPLD Middleware)
to provide all the necessary context for this ADR.

## Test Cases

None in the SDK, there will be encoding/decoding tests for the IPLD codecs in the codec repo linked below.

## References

* Existing IPLD codec implementations for Tendermint and Cosmos data structures: https://github.com/vulcanize/go-codec-dagcosmos
* Existing IPLD Spec/Schemas for Tendermint and Cosmos data structures: https://github.com/vulcanize/ipld/tree/cosmos_specs/specs/codecs/dag-cosmos
* Tendermint and Cosmos IPLD Schemas discussion: https://github.com/cosmos/cosmos-sdk/discussions/9505 
* First mention of supporting the arbitrary protobuf types in IPLD: https://github.com/cosmos/cosmos-sdk/issues/7097#issuecomment-742752603
* Issue describing various options (including this one) for supporting Cosmos protobuf types as IPLD: https://github.com/vulcanize/go-codec-dagcosmos/issues/23
