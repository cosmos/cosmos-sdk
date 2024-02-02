<!--
Guiding Principles:
Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.
Usage:
Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:
* (<tag>) [#<issue-number>] Changelog message.
Types of changes (Stanzas):
"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"API Breaking" for breaking exported APIs used by developers building on SDK.
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

### Features

* [#19304](https://github.com/cosmos/cosmos-sdk/pull/19304) Add `MsgSudoExec` for allowing executing any message as a sudo.
* [#19101](https://github.com/cosmos/cosmos-sdk/pull/19101) Add message based params configuration.
* [#18532](https://github.com/cosmos/cosmos-sdk/pull/18532) Add SPAM vote to proposals.
* [#18532](https://github.com/cosmos/cosmos-sdk/pull/18532) Add proposal types to proposals.
* [#18620](https://github.com/cosmos/cosmos-sdk/pull/18620) Add optimistic proposals.
* [#18762](https://github.com/cosmos/cosmos-sdk/pull/18762) Add multiple choice proposals.

### Improvements

* [#18976](https://github.com/cosmos/cosmos-sdk/pull/18976) Log and send an event when a proposal deposit refund or burn has failed.
* [#18856](https://github.com/cosmos/cosmos-sdk/pull/18856) Add `ProposalCancelMaxPeriod` parameter for modifying how long a proposal can be cancelled after it has been submitted.
* [#19167](https://github.com/cosmos/cosmos-sdk/pull/19167) Add `YesQuorum` parameter specifying a minimum of yes vote in the total proposal voting power for the proposal to pass.
* [#18445](https://github.com/cosmos/cosmos-sdk/pull/18445) Extend gov config.
* [#18532](https://github.com/cosmos/cosmos-sdk/pull/18532) Repurpose `govcliutils.NormalizeProposalType` to work for gov v1 proposal types.

### State Machine Breaking

* [#19101](https://github.com/cosmos/cosmos-sdk/pull/19101) Add message based params configuration.
* [#18532](https://github.com/cosmos/cosmos-sdk/pull/18532) Add SPAM vote to proposals.
* [#18532](https://github.com/cosmos/cosmos-sdk/pull/18532) Add proposal types to proposals.
* [#18620](https://github.com/cosmos/cosmos-sdk/pull/18620) Add optimistic proposals.
* [#18762](https://github.com/cosmos/cosmos-sdk/pull/18762) Add multiple choice proposals.
* [#18856](https://github.com/cosmos/cosmos-sdk/pull/18856) Add `ProposalCancelMaxPeriod` parameters.
* [#19167](https://github.com/cosmos/cosmos-sdk/pull/19167) Add `YesQuorum` parameter.

### Client Breaking Changes

* [#19101](https://github.com/cosmos/cosmos-sdk/pull/19101) Querying specific params types was deprecated in gov/v1 and has been removed. gov/v1beta1 rest unchanged.

### API Breaking Changes

* [#18532](https://github.com/cosmos/cosmos-sdk/pull/18532) All functions that were taking an expedited bool parameter now take a `ProposalType` parameter instead.
* [#17496](https://github.com/cosmos/cosmos-sdk/pull/17496) in `x/gov/types/v1beta1/vote.go` `NewVote` was removed, constructing the struct is required for this type.
* [#19101](https://github.com/cosmos/cosmos-sdk/pull/19101) Move `QueryProposalVotesParams` and `QueryVoteParams` from the `types/v1` package to `utils` and remove unused `querier.go` file.

### Deprecated

* [#18532](https://github.com/cosmos/cosmos-sdk/pull/18532) The field `v1.Proposal.Expedited` is deprecated and will be removed in the next release.
