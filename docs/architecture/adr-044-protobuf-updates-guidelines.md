# ADR 044: Guidelines for Updating Protobuf Definitions

## Changelog

- 28.06.2021: Initial Draft

## Status

Draft

## Abstract

This ADR provides guidelines and recommended practices when updating Protobuf definitions. These guidelines are targeted at module developers.

## Context

The SDK maintains a set of [Protobuf definitions](https://github.com/cosmos/cosmos-sdk/tree/master/proto/cosmos). When making changes to these Protobuf definitions, the SDK currently only follows [Buf's](https://docs.buf.build/) recommendations. We noticed however that Buf's recommendations might still result in breaking changes in the SDK in some cases. For example:

- Adding fields to `Msg`s. Adding fields is a not a Protubuf spec-breaking operation. However, when adding new fields to `Msg`s, the unknown field rejection will throw an error when sending the new `Msg` to an older node.
- Marking fields as `reserved`. Protobuf proposes the `reserved` keyword for removing fields without the need to bump the package version. However, by doing so, client backwards compatability is broken as Protobuf doesn't generate anything for `reserved` fields. See [#9446](https://github.com/cosmos/cosmos-sdk/issues/9446) for more details on this issue.

Morever, module developers often face other questions around Protobuf definitions such as "Can I rename a field?" or "Can I deprecate a field?" This ADR aims to answer all these questions by providing clear guidelines on allowed updates for Protobuf definitions.

## Decision

We decide to keep [Buf's](https://docs.buf.build/) recommendations with the following exceptions (TODO document what they mean, and why we use them):

- UNARY_RPC
- COMMENT_FIELD
- SERVICE_SUFFIX
- PACKAGE_VERSION_SUFFIX
- RPC_REQUEST_STANDARD_NAME

On top of Buf's recommendations we add the following guidelines that are specific to the SDK.

### Updating Protobuf Definition Withouth Bumping Version

#### 1. `Msg`s SHALL NOT have new fields.

TODO because of unknown fields rejection.
TODO mention nested structs too (TODO how about anys?)
TODO queries can have new fields

#### 2. Fields MAY be marked as `deprecated`, and nodes MAY implement a protocol-breaking change for handling these fields.

Protobuf supports the [`deprecated` field option](https://developers.google.com/protocol-buffers/docs/proto#options), and this option MAY be used on any fields, including `Msg`s. If a node handles a Protobuf message with a non-empty deprecated field, the node MAY change its behavior upon processing it, even in a protocol-breaking way. When possible, the node SHOULD handle backwards compatibility.

As an example, the SDK v0.42 to v0.43 update contained two Protobuf-breaking changes, listed below. Instead of bumping the package versions from `v1beta1` to `v1`, the SDK team decided to follow this guideline, by reverting the breaking changes, marking those changes as deprecated, and modifying the node implementation when processing messages with deprecated fields. More specifically:

- the SDK recently removed support for [time-based software upgrades](https://github.com/cosmos/cosmos-sdk/pull/8849). As such, the `time` field has been marked as deprecated in `cosmos.upgrade.v1beta1.Plan`. Moreover, the node will reject any proposal containing an upgrade Plan whose `time` field is non-empty.
- the SDK now supports [governance split votes](./adr-037-gov-split-vote.md). When querying for votes, the returned `cosmos.gov.v1beta1.Vote` message has its `option` field (used for 1 vote option) deprecated in favor of its `options` field (allowing multiple vote options). Whenever possible, the SDK still populates the deprecated `option` field, that is, if and only if the `len(options) == 1` and `options[0].Weight == 1.0`.

#### 3. Fields SHALL NOT be renamed.

TODO as this will break client compatibility.

### Bumping Protobuf Package Version

TODO, needs architecture review. Some topics:

- When bumping versions, should the SDK support both versions?
  - i.e. v1beta1 -> v1, should we have two folders in the SDK, and handlers for both versions?

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

- {reference link}
