# ADR 008: Evidence Module

## Changelog

- 31-07-2019: Initial draft

## Context

In order to support building highly secure, robust and interoperable blockchain
applications, it is vital for the Cosmos SDK to expose a mechanism in which arbitrary
evidence can be submitted, evaluated and verified resulting in some agreed upon
penalty for any equivocations committed by a validator. Furthermore, such a
mechanism is paramount for any [IBC](https://github.com/cosmos/ics/blob/master/ibc/1_IBC_ARCHITECTURE.md)
protocol implementation in order to support the ability for any equivocations to
be relayed back from a collateralized chain to a primary chain so that the
equivocating validator(s) can be slashed.

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

Proposed

## Consequences

### Positive

- Allows the state machine to process equivocations submitted on-chain and penalize
validators based on agreed upon slashing parameters
- Does not solely rely on Tendermint to submit evidence

### Negative

- No easy way to introduce new evidence types through governance on a live chain
due to the inability to introduce the new evidence type's corresponding handler

### Neutral

## References

- [ICS](https://github.com/cosmos/ics)
- [IBC Architecture](https://github.com/cosmos/ics/blob/master/ibc/1_IBC_ARCHITECTURE.md)
