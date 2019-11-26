# ADR 16: Historical Header Module

## Changelog

- 26 November 2019: First version

## Context

In order for the Cosmos SDK to implement the [IBC specification](https://github.com/cosmos/ics), modules within the SDK must have the ability to introspect recent consensus states (validator sets & commitment roots) as proofs of these values on other chains must be checked during the handshakes.


> This section describes the forces at play, including technological, political, social, and project local. These forces are probably in tension, and should be called out as such. The language in this section is value-neutral. It is simply describing facts.
> {context body}

## Decision

> This section describes our response to these forces. It is stated in full sentences, with active voice. "We will ..."
> {decision body}

## Status

> A decision may be "proposed" if the project stakeholders haven't agreed with it yet, or "accepted" once it is agreed. If a later ADR changes or reverses a decision, it may be marked as "deprecated" or "superseded" with a reference to its replacement.
> {Deprecated|Proposed|Accepted}

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Positive

- Easy retrieval of headers & state roots for recent past heights by modules anywhere in the SDK

### Negative

- Additional storage usage, as opposed to querying from Tendermint (at present this is not possible)

### Neutral

(none known)

## References

- [ICS 2: "Consensus state introspection"](https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements#consensus-state-introspection)
