# ADR-066: MEV module

## Changelog

* 2023-03-27 — rename x/builder to x/mev
* 2023-03-22 — initial submission

## Status

PROPOSED

## Abstract

This ADR proposes a new Cosmos SDK module — the **MEV module** — which offers a framework for chains to capture and redistribute MEV.

The proposal ensures the sovereignty of chains to define approved builders, enforce preferences on block production, and specify exactly how profits are distributed. It does so by auctioning partial space in each block to an open market of builders, which maximizes aggregate value and network utility.

## Context

### Motivation

Block construction matters.

There is enormous value in choosing the transactions that go into in a block, and the order of those transactions. That value — commonly known as MEV — is large enough that it will be captured one way or another. In Ethereum, it’s captured with the de facto standard mechanism of Flashbots. In Cosmos, it’s captured with a variety of unique, custom, and incompatible mechanisms, due to the lack of a standard.

Several parties, including Mekatek, provide MEV solutions that are delivered as forks of CometBFT. Other solutions are tightly coupled to specific chains. These are all good first steps, but we’ve learned that they aren’t enough. In order to be effective, we need a solution that’s generalized (works on all Cosmos networks, not just specific ones) and more deeply integrated with the protocol (to support necessary features like evidence).

### Proposal

We propose a standard, opt-in mechanism for Cosmos-based chains to capture MEV. It includes a new Cosmos SDK **MEV module** and a **builder API** specification.

The proposal...

- Allows chains to set and enforce rules (i.e. preferences) on block construction
- Allows chains to define exactly how MEV profits are distributed
- Maximizes MEV profits with an open market of off-chain builders
- Provides sealed auctions that keep searcher bundles and builder bids private
- Makes proposers and builders accountable via verifiable evidence of commitments
- Fully supports future cross-chain MEV

## Alternatives

There are several approaches for MEV capture currently in use.

### MEV-Boost

Probably most famously, MEV-Boost from Flashbots is the de facto standard for MEV capture on Ethereum. It implements what’s known as proposer-builder separation, which (broadly) means the entire block is constructed by the builder, and proposed by the proposer without inspection or modification.

We believe MEV-Boost is probably a good approach for Ethereum, but not the right fit for Cosmos. Important properties, like sealed-bid auctions and censorship resistance, are tightly coupled to Ethereum-specific capabilities, like commit-reveal schemes based on threshold encryption and attestations. Also, it doesn’t provide a way to express block construction preferences that govern the behavior of builders.

### Off-chain APIs

Some products offer MEV capture in Cosmos networks via an off-chain API.

Products like Skip Select and Mekatek Zenith offer a custom patch of CometBFT that runs a spot auction for top-of-block transactions by querying a remote builder through an HTTP API. Validators opt-in by building their full node with the custom patch, and propose blocks that include builder bids normally.

We believe that this approach was a useful first step for Cosmos, but isn’t viable long-term. Custom patching incurs significant operational overhead and runtime risk for validators, and makes it difficult (or impossible) for chains to maintain sovereignty over block construction.  Preferences, accountability, evidence, and other important requirements for fair MEV capture require some amount of protocol integration.

### On-chain searchers

Some products claim to capture MEV themselves, acting as a sort of fully-autonomous and on-chain searcher. Skip’s Protorev and Astroport’s Redpoint are two examples of this approach. They look at each block before it is proposed, and try to capture a specific form of MEV, based on heuristics that are (typically) hard-coded and application-specific. They then modify the block based on the findings, and (typically) distribute profits back into the network.

We believe this approach is too restrictive to be viable. It’s only able to identify a small category of MEV, which means third-party builders are still incentivized to try and capture the rest. An open market of builders is necessary for an MEV capture solution to be effective.

### POB

POB is a recent proposal from Skip and Osmosis. Like this proposal, it implements a top-of-block auction as a new SDK module. Unlike this proposal, there is no mechanism for accountability between proposers and builders, and bids are submitted as normal gossiped transactions. The “push” model for bids has many interesting properties, but without something like threshold encryption, it means that bids are public, which makes auctions vulnerable to bid sniping, latency races, and similar problems. And to the best of our knowledge, it relies on features only present in ABCI 2.0.

We believe that POB has a lot of good properties. (It’s very similar to this proposal!) However, we consider verifiable accountability between proposers and builders, and sealed auctions of private bids, to be table-stakes features.

## Decision

We propose a new MEV module, which allows chains to auction partial block segments to off-chain builders.

It’s opt-in for both chains and validators. When enabled, it maintains the autonomy and authority of chains to set rules on block construction and profit distribution. It maximizes that profit, by creating an open and competitive market for third-party builders. And it defends against malicious behavior, like MEV stealing and censorship, with verifiable proofs of mutual builder and proposer commitments.

### Cosmos SDK module

The MEV module is implemented as a Cosmos SDK module. Chains can opt-in by wiring it up, like any other module. There are two main actors: proposers and builders.

### Proposers

Proposers are validators, when it’s their turn to build a block. Validators must explicitly opt-in to the MEV module by sending a registration transaction on-chain. This creates a MEV module specific identity for the validator, with a chain-of-trust from the validator’s operator (i.e. staking) address.

```bash
simd tx mev register-proposer \
  ~/.simapp/config/builder_module_proposer_key.json \
  --from=cosmosvaloper1rf57ygfmfex8es8yhsde707x7x98p4l6nxmufs
```

Registration produces a new key pair for the validator, which is stored on-disk rather than in the key ring. The MEV module needs access to that private key to e.g. sign messages to builders, and modules don’t have access to the key ring.

### PrepareProposal: Block Construction

The core logic of the MEV module is implemented in its PrepareProposal function.

PrepareProposal is an ABCI 1.0 method that’s called by Tendermint to produce the next block. By default, PrepareProposal returns a block built from the mempool transactions known to the proposer. We extend that default behavior, and run what is essentially a **spot auction** for part of the block space.

The proposer first solicits bids for its block space from builders, by making an HTTP request to the bid endpoint of registered and approved builders. That request contains information about the block that’s up for auction, including preferences governing how blocks are constructed, which bids must satisfy.(A preference is a rule defined by the chain, defined by a unique ID, and enforced by a validation function.)

Builders respond with a bid, containing a **hash** of the transactions they’d like to include in the block, and a promised payment to the network (via the MEV module) if the bid is accepted.

The proposer passes valid bids to a function that selects a winner. That function can be specified by the application; the default behavior is to select the highest-paying valid bid. The proposer then calls the winning builder’s commit endpoint, sending a verifiable commitment to include the winning bid in the block.

The builder responds with the complete set of transactions for the bid, plus a **segment commitment transaction**, which contains verifiable metadata about the promises made by the proposer and builder to each other.

The block is constructed in three parts. First, the **prefix** region is always at the top of the block, and contains transactions returned from calling a special function which applications may optionally provide; by default, the prefix region is empty. Next, the **segment** region of the block contains transactions from the winning bid. Finally, the **suffix** region of the block contains any transactions in the proposer’s mempool that haven’t already been included.

Any non-recoverable error during PrepareProposal immediately “fails fast” and returns a default block built from the mempool transactions known to the proposer. PrepareProposal is subject to CometBFT’s consensus timeouts, so we enforce a strict timeout on the overall execution, which is << the consensus timeouts. We also enforce timeouts on individual calls to builders.

### ProcessProposal: Block Validation

ProcessProposal is another ABCI 1.0 method, called by the proposer and validators during consensus, to validate (and potentially reject) proposed blocks.

The MEV module defines a ProcessProposal method which looks for a commitment transaction. If it finds one, it verifies that the outer block satisfies the commitments in that transaction. Some checks are as follows.

- The commitment’s prefix region and hash match the block data
- The commitment’s segment region and hash match the block data
- The commitment transaction is directly after the final segment region transaction
- The commitment’s preferences are satisfied by the segment transactions in the block
- The builder account can make the promised payment

Only the first commitment transaction in a block is evaluated. It is invalid for a commitment transaction to exist in the prefix or segment regions of the block.

### Builders

Builders register on-chain similarly to validators. An example registration flow follows.

```bash
# Create a new wallet and identity for this builder.
simd keys add builder-wallet

# Fund that account, for gas fees, and bid payments.
simd tx bank send \
  some-other-wallet \
  $(simd keys show -a builder-wallet) \
  1000stake

# Register the builder.
simd tx mev register-builder \
  --from=$(simd keys show -a builder-wallet) \
  --moniker="Turbo Builder" \
  --security-contact="security@turbobuilder.xyz" \
  --api="v0:https://api.turbobuilder.xyz/v0" # TODO: not the final syntax

```

A builder is uniquely defined by its wallet address. A registered builder is one-to-many with versioned API URLs, which should serve the corresponding builder API. Builder API v0 is defined below.

Registered builders must be explicitly approved in order to receive bid requests from proposers. Builders are approved via a MEV module consensus param called `AllowedBuilderAddresses`. The default (empty) value is treated as “default allow” and all registered builders will participate in auctions. The `MaxBuildersPerAuction` consensus param restricts the maximum number of builders that can participate in each auction. If more than this number of builders are registered, the builders for a given auction are chosen via a PRNG seeded with the block height. The default (zero) value means that the proposer will not run an auction at all, and will propose the default block containing only local mempool transactions.

These parameters are modified by submitting (and approving) a governance proposal containing a `MsgUpdateParams` with the new values.

### Builder API v0

The builder API specifies how proposers communicate with builders.

The API is defined as a set of HTTP endpoints, which receive JSON requests, and return JSON responses. HTTP and JSON were chosen deliberately, as they represent the “least common denominator” for communication over the internet. This allows builders to be implemented without constraints on language (e.g. Go), and without requiring compile-time dependencies on third-party schema definitions (e.g. protocol buffers). We consider these properties important to ensure a robust builder marketplace.

The API only supports requests from proposers to builders. Requests from any other client are invalid. Builders have no obligation to respond to any client except proposers.

Endpoints are expressed as paths, appended to the end of the registered URL for each builder. For example, a builder URL `https://example.com/cosmos/builder/v0` would receive bid requests to `https://example.com/cosmos/builder/v0/bid`.

The API is versioned, and the version applies to the entire set of endpoints, e.g. `/v0/{bid,commit}`, `/v1/{bid,commit,...}`.

Proposers only accept replies from builders with a response code of `200 OK`, which specify `content-type: application/json`, and which are correctly signed.

#### POST /bid

**Bid Request**

When a proposer wants to propose a block, it solicits bids from builders by making a `POST /bid` request to each registered builder, providing the following JSON request body.

```json
{
	"chain_id": "my-chain-id",
	"height": "123456",
	"payment_denom": "ustake",
	"preference_ids": ["preference-identifier-1"],
	"prefix_txs": [
		"Zm9vCg==",
		"YmFyCg==",
	],
	"max_bytes": 1234567,
	"max_gas": 6543210,
	"signature": "MWM2MmYzNDVmZDNiNTMzY2ZhNzc4MmQ1MzE5NTEwZWM4MmUyY2I1ODlkNTA1YmI4OGRlNTdkODU4MTA4MWNlMQ=="
}
```

| FIELD | TYPE | DESCRIPTION |
| --- | --- | --- |
| chain_id | string | The chain ID. |
| height | string | The height of the proposed block. |
| payment_denom | string | The denomination which is acceptable for bid payments. |
| preference_ids | array of string | Preference IDs as defined by the chain which must be satisfied by bids. |
| prefix_txs | array of string | The serialized bytes of the transactions in the prefix region of the block, encoded as base64 strings. |
| max_bytes | number | The maximum number of bytes the all the segment txs must have cumulatively. |
| max_gas | number | The maximum gas all the segment txs must consume cumulatively. |
| signature | string | A signature over the SignBytes of the request, made by the proposer’s private key, encoded as a base64 string. |

**Bid Response**

Builders validate the request signature by issuing a `query builder proposer-for-height` with the provided height, to a trusted full node (or equivalent) for the network. It returns the registration data for the proposer for the given height, which includes the proposer’s public key. The builder then verifies the request signature with that public key.

Once the request has been validated, builders compute a bid for the auction, containing a set of transactions that satisfy any preferences in the request, and a payment (in the specified denomination) they will provide if the bid is accepted.

Builders then return a bid response, including the promised payment, and a count and hash of the bid transactions. This represents a commitment to the proposer: if the bid wins the auction, the builder must provide the full set of transactions.

If the builder doesn’t want to bid on an auction, or cannot satisfy the preferences in the bid request, they should return a valid bid response, with an empty set of transactions, and a payment of 0 denom.

```json
{
	"chain_id": "my-chain-id",
	"height": "123456",
	"preference_ids": ["preference-identifier-1"],
	"prefix_hash": "rsBwZF/lPuOzdjBZN2E08FjMM3JHyXit0Xi2zN+wAZ8=",
	"payment_promise": "1000stake",
	"segment_length": 3,
	"segment_hash": "NPyqHMlwWu0PpABvi10qhC1bRKLUzQVXlmO0dnI3Xcg=",
	"segment_bytes": 1234000,
	"segment_gas": 65432000,
	"signature": "MWM2MmYzNDVmZDNiNTMzY2ZhNzc4MmQ1MzE5NTEwZWM4MmUyY2I1ODlkNTA1YmI4OGRlNTdkODU4MTA4MWNlMQ=="
}
```

| FIELD | TYPE | DESCRIPTION |
| --- | --- | --- |
| chain_id | string | The chain ID from the request. |
| height | string | The height from the request. |
| preference_ids | array of string | The same preference IDs from the request, which the bid has satisfied. |
| prefix_hash | string | A SHA256 hash of the prefix transactions from the request, in order, encoded as a base64 string. |
| payment_promise | string | The payment that the builder promises to make to the MEV module (and transitively the network) if this bid is accepted, as a string of the form <amount><denom> e.g. 1000stake. |
| segment_length | number | The number of transactions in the bid. |
| segment_hash | string | A SHA256 hash of the serialized transactions of the bid, in order, encoded as a base64 string. |
| segment_bytes | number | The number of bytes the transactions sum to. |
| segment_gas | number | The gas the transactions consume. |
| signature | string | A signature over the SignBytes of the response, made by the builder’s private key, encoded as a base64 string. |

The proposer will verify the signature of the entire bid response against the registered public key of the corresponding builder. The proposer will also verify the hashes in the response against the corresponding data in the request.

Auctions are uniquely identified by chain ID, height, and proposer address. Bids are one-to-one with auctions, and are immutable: once a builder bids on a given auction, it must return exactly the same bid to all (valid) requests for that auction. Builders may delete bids once the chain height advances beyond the auction height.

Bids are binding commitments. When a builder makes a bid, it must store the transactions for that bid, and be able to return them to any valid requestor.

#### POST /commit

**Commit Request**

When a proposer selects a winning bid, it makes a `POST /commit` request to the winning builder, providing a verifiable commitment to the builder that it will include the bid in the block. This commitment can be used by validators to ensure a proposed block matches the expectations of the builder. And because it is signed, it can also be used by builders as evidence against malicious proposers who commit to a bid that they then modify, ignore, or otherwise censor.

| FIELD | TYPE | DESCRIPTION |
| --- | --- | --- |
| proposer_address | string | The proposer address, encoded as a Bech32 string. |
| builder_address | string | The builder address from the winning bid, encoded as a Bech32 string. |
| chain_id | string | The chain ID. |
| height | string | The height of the proposed block. |
| preference_ids | array of string | The validated honored preferences from the winning bid. |
| prefix_offset | number | The offset in the block of the beginning of the prefix region, typically 0. |
| prefix_length | number | The length of the prefix region of the block. |
| prefix_hash | string | The verified prefix_hash from the winning bid.  |
| segment_offset | number | The offset in the block of the beginning of the segment region (the transactions from the winning bid), typically prefix_offset + prefix_length. |
| segment_length | number | The segment_length from the winning bid. |
| segment_size | number | The segment_size from the winning bid. |
| segment_gas | number | The segment_gas from the winning bid. |
| segment_hash | string | The segment_hash from the winning bid. |
| payment_promise | string | The payment_promise from the winning bid. |
| signature | string | A signature over the SignBytes of the object, made by the proposer’s private key, encoded as a base64 string. |

```json
{
	"proposer_address": "cosmos1rf57ygfmfex8es8yhsde707x7x98p4l6nxmufs",
	"builder_address": "cosmos1ef67ygfajeu8es8yhsde837x7x98p4l6nxmuaxx",
	"chain_id": "my-chain-id",
	"height": "123456",
	"preference_ids": ["preference-identifier-1"],
	"prefix_offset": 0,
	"prefix_length": 2,
	"prefix_hash": "rsBwZF/lPuOzdjBZN2E08FjMM3JHyXit0Xi2zN+wAZ8=",
	"segment_offset": 2,
	"segment_length": 3,
	"segment_hash": "NPyqHMlwWu0PpABvi10qhC1bRKLUzQVXlmO0dnI3Xcg=",
	"segment_bytes": 1234000,
	"segment_gas": 65432000,
	"payment_promise": "1000stake",
	"signature": "ZWI5YzQ1MWFiMjFhNzNkNDc0ZWI4ZjU4OTZjZTA1NGU4NzA3NDhiYTRlZDRhZDg0NDNhMjFkYmQwYTE3ZTAyZA==",
}
```

**Commit Response**

The builder authenticates the commit request, same as it does for bids. It also verifies the metadata in the commit request matches the metadata from the corresponding bid response, which it made earlier.

Once the commit request is validated, the builder retrieves the segment transactions for the bid, and produces a segment commitment transaction with a single MsgCommitSegment message. That message represents verifiable evidence of the commitments made by the proposer and the builder to each other.

Most MsgCommitSegment fields can be copied directly from the commit request, including the proposer signature. To produce the builder signature, the builder signs the SignBytes of the commit request with its own private key.

The builder returns a commit response to the proposer, which is signed in the same way as a bid response.

```json
{
	"chain_id": "my-chain-id",
	"height": "123456",
	"segment_txs": [
		"c2VnbWVudC10eC0xCg==",
		"YW5vdGhlci1zZWdtZW50LXRyYW5zYWN0aW9uCg=="
	],
	"segment_commitment_tx": "c2VnbWVudC1jb21taXRtZW50LXR4Cg==",
	"signature": "M2E1Y2JlMTgzMDNlYzc2ODUyN2Y3NDhhOGYxNjJiN2FjYzIxZTQ3Y2FlZGZlZWUxZTIwNDFiYjhjY2Y5NGJjOA=="
}
```

| FIELD | TYPE | DESCRIPTION |
| --- | --- | --- |
| chain_id | string | The chain ID. |
| height | string | The height of the proposed block. |
| segment_txs | array of string | The serialized bytes of each transaction in the bid, encoded as base64 strings.  |
| segment_commitment_tx | string | The serialized bytes of the mutually signed segment commitment transaction, encoded as a base64 string. |
| signature | string | A signature over the SignBytes of the response, made by the builder’s private key, encoded as a base64 string. |

The proposer validates this response, and includes the segment transactions, followed by the segment commitment transaction, in the block.

### Hashes

Hashes in the builder API are defined to be SHA256.

If a transaction is encoded as a base64 string, it must first be decoded to its raw bytes before it can be hashed. Arrays are hashed by passing the raw bytes of each element to the hash function in sequence.

### Signatures

Every type defined by the builder API is signed by the sender, using the sender’s registered key. That’s the key returned by e.g. `simd query builder show-builder` or `show-proposer` as appropriate.

The MEV module automatically signs outgoing requests, and verifies the signature of incoming responses. Builders must verify the signatures of incoming requests, and sign outgoing responses.

Signatures are made over a specific serialization of a value, commonly known as SignBytes. SignBytes must be deterministic and bijective in order for signatures to be valid. Most common serialization formats don’t provide these properties, and so cannot be used. In particular, Protocol Buffers is unsuitable as it [explicitly states that its output is unstable and non-deterministic](https://protobuf.dev/programming-guides/encoding/#implications).

We define the (deterministic and bijective) SignBytes for API types as follows.

- A compact, single-line JSON serialization of the value
    - All leading, trailing, and interior whitespace removed
    - No trailing newline
- Object keys sorted in lexicographical order
- Any top-level `signature` keys removed
- All other keys explicitly provided with their values
    - String values in UTF-8 with standard escaping only (i.e. no HTML escaping)
    - Number values as integers, no floats or decimal points
    - Bytes as base64 encoded strings
    - Arrays as normal
- No `null` values
    - Empty or missing string values as `""`
    - Zero number values as `0`
    - Empty or null bytes values as `""`
    - Empty or null array values as `[]`

The module implementation will include API request and response types, with methods for generating SignBytes, producing signatures, and verifying those signatures. It will also include a comprehensive set of test cases, which can be used as a reference for other implementations.

### Payment

Bids include a promised payment, which will be transferred from the builder’s account to the x/mev module account if the bid wins the auction. Bids with invalid payments (e.g. insufficient funds) are considered invalid, and can result in consequences for the builder.

At the end of every block, payments are distributed according to an application-defined distribution function. By default, payments are transferred to the x/distribution module account and distributed to validators, delegators, etc. in the same way as fees.

### Preferences: Good MEV, Bad MEV

MEV is often described as being either good or bad. A common example of bad MEV is sandwich attacks, which are a specific form of front-running. Chains must be able to define what they consider to be bad MEV, and those rules, commonly known as **preferences**, will be enforced.

Sandwiching and front-running are not well-defined. They typically rely on application-specific types (e.g. pools), and detecting them is heuristic (i.e. approximate) and potentially computationally expensive. Preferences are application-specific: what works for one chain may not work for another.

We allow applications to define whatever preferences they like, as functions that validate a set of transactions and reject anything containing e.g. bad MEV. Those preferences are defined on-chain, enforced by the proposer as part of block construction, and verified by validators as part of consensus.

### Evidence

Evidence facilitates accountability between proposers and builders. It’s modeled as specific message types, and submitted by sending a transaction on chain. It contains details about things that happened in the past, and enough metadata (signatures) to allow those claims to be independently verified.

Proposers get evidence of certain kinds of builder misbehavior, like if a builder includes a promised payment in a bid that’s larger than the balance of its account. Builders get evidence of certain kinds of proposer misbehavior, like if a proposer commits to a bid but then doesn’t actually include it in the block.

Evidence doesn’t imply guilt. Consider a builder which receives a commitment from a proposer, but returns an error instead of the bid transactions. To the proposer, this might be evidence of misbehavior, but it could also be due to a network error.

The proposer can’t include a bid it doesn’t have, so it would produce a block without the transactions it committed to. When the builder saw that block on chain, they could submit the commitment they received as evidence against the proposer, and that claim would be valid, even though the builder was the one acting maliciously.

Evidence establishes specific, and typically narrow, facts. We don’t try to do any form of automatic punishment from evidence claims. Applications should use evidence to inform a open, social, governance-based, etc. decision-making process.

### Query Interface

The **ProposerForHeight** query method takes a height, and returns the MEV module proposer for that height, if the proposer is registered. That record includes a public key, which builders should use to verify the signature of received requests.

## Consequences

### Backwards Compatibility

This is all new functionality, so there is no impact on backwards compatibility.

### Positive

Cosmos chains will be able to rely on a standardized way to opt-in to off-chain block building. The proposal allows networks to define and enforce what kinds of transactions are permitted, and how revenue from builders is distributed. This will foster the development of a builder market which is competitive, transparent, and fair to all participants.

### Negative

Off-chain block building is additional complexity for the chains that opt-in to it.

## Future Work

We are mostly done with an initial implementation of the module, and a reference implementation of a builder. When the ADR is approved, we will follow up with those artifacts. The submitters fully intend to own and maintain the module over time, so there should be no significant extra work for the SDK maintainers.

## Test Cases

Both the prototype implementation of the module, and the reference implementation of the builder API, will contain comprehensive test suites.

## References

- [Interchain MEV Group](https://groups.google.com/g/interchain-mev-group)
- [Builders as Citizens (forum post)](https://forum.cosmos.network/t/builders-as-citizens-standardizing-mev-auctions-in-cosmos/10004)
- [What is MEV (Ethereum)](https://ethereum.org/en/developers/docs/mev/)
- [The importance of block space](https://www.generalist.com/briefing/blockspace)
- [Block space as a market](https://www.aniccaresearch.tech/blog/consensus-capital-markets)
- [Profit-sharing validators](https://www.recvc.com/mev-2-0-the-rise-of-mpsvs/)
- [A Formalization of MEV (Flashbots)](https://arxiv.org/pdf/2112.01472.pdf)
- [Transaction fee design (Ethereum)](https://timroughgarden.org/papers/eip1559.pdf)
- [ARC-50: Astroport/Redpoint MEV](https://forum.astroport.fi/t/arc-50-astroport-redpoint-mev-round-2/1022/2)
