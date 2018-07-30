# ADR 001: Cosmos SDK Signed Messages

## Changelog

## Context

Having the ability to sign messages off-chain has proven to be a fundamental aspect
of nearly any blockchain. The notion of signing messages off-chain has many 
added benefits such as saving on computational costs and reducing transaction
throughput and overhead. Within the context of the Cosmos SDK, some of the major
applications of signing such data includes, but is not limited to, providing a
cryptographic secure and verifiable means of proving validator identity and
possibly associating it with some other framework or organization. In addition,
having the ability to sign Cosmos SDK messages with a Ledger device.

A standardized protocol for hashing and signing messages that can be implemented
by the Cosmos SDK and any third-party organization is needed the subscribes to
the following:

* A specification of structured data
* A framework for encoding structured data
* A cryptographic secure hashing and signing algorithm
* A framework for supporting extensions and domain separation

This record is only concerned with the rationale and the standardized implementation
of Cosmos SDK signed messages. It does **not** concern itself with the concept of
replay attacks as that will be left up to the higher-level application implementation.
If you view signed messages in the means of authorizing some action or data, then 
such an application would have to either treat this as idempotent or have mechanisms
in place to reject known signed messages.

TODO: Should we bake in replay protection into the protocol?

## Decision

> The proposed implementation is motivated by EIP-712<sup>1</sup> and in general
Ethereum's `eth_sign` method<sup>2</sup>.

### Preliminary

We will a have Cosmos SDK message signing protocol that consists of `TMHASH`, which is
`SHA-256` with all but the first 20 bytes truncated, as the hashing algorithm and
`secp256k1` as the signing algorithm.

Note, our goal here is not to provide context and reasoning about why necessarily
these algorithms were chosen apart from the fact they are the defacto algorithms
used in Tendermint and the Cosmos SDK and that they satisfy our needs for such
algorithms such as having resistance to second pre-image attacks and collision,
as well as being deterministic and uniform.

{decision body}

## Status

Proposed.

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## References

1. https://github.com/ethereum/EIPs/blob/master/EIPS/eip-712.md
2. https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sign