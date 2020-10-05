# ADR 031

## Changelog

- {date}: {changelog}

## Status

> A decision may be "proposed" if the project stakeholders haven't agreed with it yet, or "accepted" once it is agreed. If a later ADR changes or reverses a decision, it may be marked as "deprecated" or "superseded" with a reference to its replacement.
> {Deprecated|Proposed|Accepted} {Implemented|Not Implemented}

## Context

> This section describes the forces at play, including technological, political, social, and project local. These forces are probably in tension, and should be called out as such. The language in this section is value-neutral. It is simply describing facts.
> {context body}

## Decision

### Use Protobuf Service Definition for Msg's

```proto
package cosmos.gov;

service Msg {
  rpc SubmitProposal(MsgSubmitProposal) returns (MsgSubmitProposalResponse);
}

// Note that for backwards compatibility this uses MsgSubmitProposal as the request type instead of the more canonical MsgSubmitProposalRequest
message MsgSubmitProposal {
  google.protobuf.Any content = 1;
  bytes proposer = 2;
}

message MsgSubmitProposalResponse {
  uint64 proposal_id;
}
```

###

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
