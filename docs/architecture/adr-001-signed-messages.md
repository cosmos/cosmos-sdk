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

## Decision

> This section describes our response to these forces. It is stated in full sentences, with active voice. "We will ..."

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

* {reference link}
