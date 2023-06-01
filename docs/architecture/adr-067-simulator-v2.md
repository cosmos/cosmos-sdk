# ADR 067: Simulator v2

## Changelog

* June 01, 2023: Initial Draft (@alexanderbez)

## Status

DRAFT

## Abstract

The Cosmos SDK simulator is a tool that allows developers to test the entirety
of their application's state machine through the use of pseudo-randomized "operations",
which represent transactions. The simulator also provides primitives that ensures
there are no non-determinism issues and that the application's state machine can
be successfully exported and imported using randomized state.

The simulator has played an absolutely critical role in the development and testing
of the Cosmos Hub and all the releases of the Cosmos SDK after the launch of the
Cosmos Hub. Since the Hub, the simulator has relatively not changed much, so it's
overdue for a revamp.

## Context

> This section describes the forces at play, including technological, political,
> social, and project local. These forces are probably in tension, and should be
> called out as such. The language in this section is value-neutral. It is simply
> describing facts. It should clearly explain the problem and motivation that the
> proposal aims to resolve.
> {context body}

## Decision

> This section describes our response to these forces. It is stated in full
> sentences, with active voice. "We will ..."
> {decision body}

## Consequences

> This section describes the resulting context, after applying the decision. All
> consequences should be listed here, not just the "positive" ones. A particular
> decision may have positive, negative, and neutral consequences, but all of them
> affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section
> describing these incompatibilities and their severity. The ADR must explain
> how the author proposes to deal with these incompatibilities. ADR submissions
> without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

> {positive consequences}

### Negative

> {negative consequences}

### Neutral

> {neutral consequences}

## Further Discussions

> While an ADR is in the DRAFT or PROPOSED stage, this section should contain a
> summary of issues to be solved in future iterations (usually referencing comments
> from a pull-request discussion).
> 
> Later, this section can optionally list ideas or improvements the author or
> reviewers found during the analysis of this ADR.

## References

* {reference link}
