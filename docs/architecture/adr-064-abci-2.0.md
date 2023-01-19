# ADR 64: ABCI 2.0 Integration (Phase II)

## Changelog

* 2023-01-17: Initial Draft (@alexanderbez)

## Status

PROPOSED

## Abstract

This ADR outlines the continuation of the efforts to implement ABCI++ in the Cosmos
SDK outlined in [ADR 060: ABCI 1.0 (Phase I)](adr-060-abci-1.0.md).

Specifically, this ADR outlines the design and implementation of ABCI 2.0, which
includes `ExtendVote`, `VerifyVoteExtension` and `FinalizeBlock`.

## Context

ABCI 2.0 continues the promised updates from ABCI++, specifically three additional
ABCI methods that the application can implement in order to gain further control,
insight and customization of the consensus process, unlocking many novel use-cases
that previously not possible. We describe these three new methods below:

### `ExtendVote`

This method allows each validator process to extend the pre-commit phase of the
Tendermint consensus process. Specifically, it allows the application to perform
custom business logic that extends the pre-commit vote and supply additional data
as part of the vote.

The data, called vote extension, will be broadcast and received together with the
vote it is extending, and will be made available to the application in the next
height. Specifically, the proposer of the next block will receive the vote extensions
in `RequestPrepareProposal.local_last_commit.votes`.

If the application does not have vote extension information to provide,
it returns a 0-length byte array as its vote extension.

**NOTE**: 

* Although each validator process submits its own vote extension, ONLY the *proposer*
  of the *next* block will receive all the vote extensions included as part of the
  pre-commit phase of the previous block. This means only the proposer will
  implicitly have access to all the vote extensions, via `RequestPrepareProposal`,
  and that not all vote extensions may be included, since a validator does not
  have to wait for all pre-commits.
* The pre-commit vote is signed independently from the vote extension.

### `VerifyVoteExtension`

This method allows validators to validate the vote extension data attached to
each pre-commit message it receives. If the validation fails, the whole pre-commit
message will be deemed invalid and ignored by Tendermint.

Tendermint uses `VerifyVoteExtension` when validating a pre-commit vote. Specifically,
for a pre-commit, Tendermint will:

* Reject the message if it doesn't contain a signed vote AND a signed vote extension
* Reject the message if the vote's signature OR the vote extension's signature fails to verify
* Reject the message if `VerifyVoteExtension` was rejected

Otherwise, Tendermint will accept the pre-commit message.

Note, this has important consequences on liveness, i.e., if vote extensions repeatedly
cannot be verified by correct validators, Tendermint may not be able to finalize
a block even if sufficiently many (+2/3) validators send pre-commit votes for
that block. Thus, `VerifyVoteExtension` should be used with special care.

Tendermint recommends that an application that detects an invalid vote extension
SHOULD accept it in `ResponseVerifyVoteExtension` and ignore it in its own logic.

### `FinalizeBlock`

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

## References

* [ADR 060: ABCI 1.0 (Phase I)](adr-060-abci-1.0.md)
