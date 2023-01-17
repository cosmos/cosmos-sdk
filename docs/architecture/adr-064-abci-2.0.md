# ADR 64: ABCI 2.0 Integration (Phase II)

## Changelog

* 2023-01-17: Initial Draft (@alexanderbez)

## Status

PROPOSED

## Abstract

This ADR outlines the continuation of the efforts to implement ABCI++ in the Cosmos
SDK outlined in [ADR 060: ABCI 1.0 (Phase I)](adr-060-abci-1.0.md).

Specifically, this ADR outlines implementation and design of `VoteExtensions` and
`FinalizeBlock`.

## Context

> This section describes the forces at play, including technological, political, social, and project local. These forces are probably in tension, and should be called out as such. The language in this section is value-neutral. It is simply describing facts. It should clearly explain the problem and motivation that the proposal aims to resolve.
> {context body}

## Alternatives

> This section describes alternative designs to the chosen design. This section is important and if an adr does not have any alternatives then it should be considered that the ADR was not thought through. 

## Decision

> This section describes our response to these forces. It is stated in full sentences, with active voice. "We will ..."
> {decision body}

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

* [ADR 060: ABCI 1.0 (Phase I)](adr-060-abci-1.0.md)
