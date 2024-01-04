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

* [#18532](https://github.com/cosmos/cosmos-sdk/pull/18532) Add SPAM vote to proposals.
* [#18532](https://github.com/cosmos/cosmos-sdk/pull/18532) Add proposal types to proposals.
* [#18620](https://github.com/cosmos/cosmos-sdk/pull/18620) Add optimistic proposals.

### Improvements

* [#18856](https://github.com/cosmos/cosmos-sdk/pull/18856) Add `ProposalCancelMaxPeriod` for modifying how long a proposal can be cancelled after it has been submitted.
* [#18445](https://github.com/cosmos/cosmos-sdk/pull/18445) Extend gov config.
* [#18532](https://github.com/cosmos/cosmos-sdk/pull/18532) Repurpose `govcliutils.NormalizeProposalType` to work for gov v1 proposal types.

### API Breaking Changes

* [#18532](https://github.com/cosmos/cosmos-sdk/pull/18532) All functions that were taking an expedited bool parameter now take a `ProposalType` parameter instead.
* [#17496](https://github.com/cosmos/cosmos-sdk/pull/17496) in `x/gov/types/v1beta1/vote.go` `NewVote` was removed, constructing the struct is required for this type.

### Deprecated

* [#18532](https://github.com/cosmos/cosmos-sdk/pull/18532) The field `v1.Proposal.Expedited` is deprecated and will be removed in the next release.
