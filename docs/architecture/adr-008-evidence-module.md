# ADR 008: Evidence Module

## Changelog

- 31-07-2019: Initial draft

## Context

In order to support building highly secure, robust and interoperable blockchain
applications, it is vital for the Cosmos SDK to expose a mechanism in which arbitrary
evidence can be submitted, evaluated and verified resulting in some agreed upon
penalty for any equivocations committed by a validator. Furthermore, such a
mechanism is paramount for any IBC protocol implementation in order to support the
ability for any equivocations to be relayed back from a collateralized chain to
a primary chain so that the equivocating validator(s) can be slashed.

## Decision

We will implement an evidence module in the Cosmos SDK supporting the following
functionality:

- Provide developers with the abstractions and interfaces necessary to define:
  - Custom evidence messages and types along with their slashing penalties
  - Evidence message handlers to determine the validity of a submitted equivocation
- Support the ability through governance to modify slashing penalties of any
evidence type
- Querier implementation to support querying params, evidence types, params, and
all submitted valid equivocations  

## Status

> A decision may be "proposed" if the project stakeholders haven't agreed with it yet, or "accepted" once it is agreed. If a later ADR changes or reverses a decision, it may be marked as "deprecated" or "superseded" with a reference to its replacement.
> {Deprecated|Proposed|Accepted}

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## References

- {reference link}
