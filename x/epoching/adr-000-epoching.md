# ADR 000: Modular AnteHandler

## Changelog

- 2020 Dec 18: Initial draft

## Status

PROPOSED

## Abstract

Updating validator power on every end of epochs instead of everyblock. Epoch could be once per day.

## Context

Currently delegation and validator set update is done on every block and it would be better to do it on epochs.

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


## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.


## References

- {reference link}
