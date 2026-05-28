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

### API Breaking

### Features

### Improvements

### Deprecated

### Bug Fixes

## v2.0.0 (April 10, 2026)

### API Breaking

* [#25470](https://github.com/cosmos/cosmos-sdk/pull/25470) Refactor store interfaces to support generic value types (object stores):
    * Replace `BasicKVStore`, `KVStore`, and `Iterator` interfaces/types with generic `GBasicKVStore[V]`, `GKVStore[V]`, and `GIterator[V]`. The old names are retained as type aliases (e.g. `KVStore = GKVStore[[]byte]`).
    * Remove `Iterator` as a direct alias to `dbm.Iterator`. It is now `GIterator[[]byte]`, a distinct interface defined in the store package. Code that type-asserts to `dbm.Iterator` will break.
    * Remove `CacheWrap()` and `CacheWrapWithTrace()` method declarations from the `CacheWrap` interface. `CacheWrap` now embeds `CacheWrapper` to obtain `CacheWrap()`.
    * Add `GetObjKVStore(StoreKey) ObjKVStore` to the `MultiStore` interface.
    * Add generic store variants across `cachekv`, `gaskv`, `prefix`, `transient`, and `mem` packages (`GStore[V]`, `NewGStore`, `NewObjStore`).
* [#26037](https://github.com/cosmos/cosmos-sdk/pull/26037) Remove `GetCommitStore` and `GetCommitKVStore` from the `CommitMultiStore` interface. Remove top-level `store.CommitStore` and `store.CommitKVStore` type aliases from `store/reexport.go`.
* [#26060](https://github.com/cosmos/cosmos-sdk/pull/26060) Remove non-functional `StoreMetrics`. This metric interface never worked, so this simply removes dead code.
* [#26061](https://github.com/cosmos/cosmos-sdk/pull/26061) Remove tracing from store interfaces and implementations:
    * Remove `SetTracer`, `SetTracingContext`, and `TracingEnabled` from `MultiStore` interface.
    * Remove `CacheWrapWithTrace` from `CacheWrapper` interface.
    * Remove `traceWriter` and `traceContext` parameters from `cachemulti.NewStore`, `cachemulti.NewFromKVStore`, and `cachemulti.NewFromParent`.
    * Remove `store/tracekv` package entirely.
    * Remove `TraceContext` type `store/types`.

### Features

* [#25470](https://github.com/cosmos/cosmos-sdk/pull/25470) Add object KV stores and refactor the base store to be generic across the value parameter:
    * Add object store types: `ObjKVStore`, `ObjBasicKVStore`, `ObjIterator`, `ObjectStoreKey`, `StoreTypeObject`.
    * Add generic store types: `GBasicKVStore[V]`, `GKVStore[V]`, `GIterator[V]`.
    * Add `cachemulti.NewFromParent` constructor for lazy cache multistore construction from a parent store function.
* [#25647](https://github.com/cosmos/cosmos-sdk/pull/25647) Add `EarliestVersion() int64` to the `CommitMultiStore` interface and `GetEarliestVersion(db)` helper.

### Bug Fixes

* [#20425](https://github.com/cosmos/cosmos-sdk/pull/20425) Fix nil pointer panic when querying historical state where a new store does not exist.
* [#24583](https://github.com/cosmos/cosmos-sdk/pull/24583) Fix pruning height calculation to correctly handle in-flight snapshots. Adds `SnapshotAnnouncer` interface and `AnnounceSnapshotHeight` to track snapshots in progress and prevent premature pruning of their heights.

## v1.1.2 (March 31, 2025)

### Bug Fixes

* [#24090](https://github.com/cosmos/cosmos-sdk/pull/24090) Running the `prune` command now disables async pruning.

## v1.1.1 (September 06, 2024)

### Improvements

* [#21574](https://github.com/cosmos/cosmos-sdk/pull/21574) Upgrade IVL to IAVL 1.2.0.


## v1.1.0 (March 20, 2024)

### Improvements

* [#19770](https://github.com/cosmos/cosmos-sdk/pull/19770) Upgrade IAVL to IAVL v1.1.1.

## v1.0.2 (January 10, 2024)

### Bug Fixes

* [#18897](https://github.com/cosmos/cosmos-sdk/pull/18897) Replace panic in pruning to avoid consensus halting. 

## v1.0.1 (November 28, 2023)

### Bug Fixes

* [#18563](https://github.com/cosmos/cosmos-sdk/pull/18563) `LastCommitID().Hash` will always return `sha256([]byte{})` if the store is empty.

## v1.0.0 (October 31, 2023)

### Features

* [#17294](https://github.com/cosmos/cosmos-sdk/pull/17294) Add snapshot manager Close method.
* [#15568](https://github.com/cosmos/cosmos-sdk/pull/15568) Migrate the `iavl` to the new key format.
    * Remove `DeleteVersion`, `DeleteVersions`, `LazyLoadVersionForOverwriting` from `iavl` tree API.
    * Add `DeleteVersionsTo` and `SaveChangeSet`, since it will keep versions sequentially like `fromVersion` to `toVersion`.
    * Refactor the pruning manager to use `DeleteVersionsTo`.
* [#15712](https://github.com/cosmos/cosmos-sdk/pull/15712) Add `WorkingHash` function to the store interface  to get the current app hash before commit.
* [#14645](https://github.com/cosmos/cosmos-sdk/pull/14645) Add limit to the length of key and value.
* [#15683](https://github.com/cosmos/cosmos-sdk/pull/15683) `rootmulti.Store.CacheMultiStoreWithVersion` now can handle loading archival states that don't persist any of the module stores the current state has.
* [#16060](https://github.com/cosmos/cosmos-sdk/pull/16060) Support saving restoring snapshot locally.
* [#14746](https://github.com/cosmos/cosmos-sdk/pull/14746) The `store` module is extracted to have a separate go.mod file which allows it be a standalone module.
* [#14410](https://github.com/cosmos/cosmos-sdk/pull/14410) `rootmulti.Store.loadVersion` has validation to check if all the module stores' height is correct, it will error if any module store has incorrect height.

### Improvements

* [#17158](https://github.com/cosmos/cosmos-sdk/pull/17158) Start the goroutine after need to create a snapshot.

### API Breaking Changes

* [#16321](https://github.com/cosmos/cosmos-sdk/pull/16321) QueryInterface defines its own request and response types instead of relying on comet/abci & returns an error
