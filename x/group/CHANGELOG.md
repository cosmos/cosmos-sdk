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

### Improvements

* [#18448](https://github.com/cosmos/cosmos-sdk/pull/18448) Extend group config
* [18286](https://github.com/cosmos/cosmos-sdk/pull/18286) Move prefix store creation down after error checks.

### Features

### API Breaking Changes

* [#19489](https://github.com/cosmos/cosmos-sdk/pull/19489) `appmodule.Environment` is received on the Keeper to get access to different application services.
* [#19410](https://github.com/cosmos/cosmos-sdk/pull/19410) Migrate to Store Service.
