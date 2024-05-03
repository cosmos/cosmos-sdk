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

* [#17294](https://github.com/cosmos/cosmos-sdk/pull/17294) Add snapshot manager Close method.
 
### Improvements

* [#17158](https://github.com/cosmos/cosmos-sdk/pull/17158) Start the goroutine after need to create a snapshot.

### Bug fixes

* [#18651](https://github.com/cosmos/cosmos-sdk/pull/18651) Propagate iavl.MutableTree.Remove errors firstly to the caller instead of returning a synthesized error firstly.
