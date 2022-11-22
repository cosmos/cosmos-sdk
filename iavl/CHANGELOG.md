# Changelog

## Unreleased

- [#586](https://github.com/cosmos/iavl/pull/586) Remove the `RangeProof` and refactor the ics23_proof to use the internal methods.

## 0.19.4 (October 28, 2022)

 - [#599](https://github.com/cosmos/iavl/pull/599) Populate ImmutableTree creation in copy function with missing field
 - [#589](https://github.com/cosmos/iavl/pull/589) Wrap `tree.addUnsavedRemoval()` with missing `if !tree.skipFastStorageUpgrade` statement

## 0.19.3 (October 8, 2022)

- `ProofInner.Hash()` prevents both right and left from both being set. Only one is allowed to be set. 
## 0.19.2 (October 6, 2022)

- [#547](https://github.com/cosmos/iavl/pull/547) Implement `skipFastStorageUpgrade` in order to skip fast storage upgrade and usage. 
- [#531](https://github.com/cosmos/iavl/pull/531) Upgrade to fast storage in batches. 

## 0.19.1 (August 3, 2022)

### Improvements

- [#525](https://github.com/cosmos/iavl/pull/525) Optimization: use fast unsafe bytes->string conversion.
- [#506](https://github.com/cosmos/iavl/pull/506) Implement cache abstraction.

### Bug Fixes

- [#524](https://github.com/cosmos/iavl/pull/524) Fix: `MutableTree.Get`.

## 0.19.0 (July 6, 2022)

### Breaking Changes

- [#514](https://github.com/cosmos/iavl/pull/514) Downgrade Tendermint to 0.34.x
- [#500](https://github.com/cosmos/iavl/pull/500) Return errors instead of panicking.

### Improvements

- [#514](https://github.com/cosmos/iavl/pull/514) Use Go v1.18

## 0.18.0 (March 10, 2022)

### Breaking Changes

- Bumped Tendermint to 0.35.1

### Improvements

- [\#468](https://github.com/cosmos/iavl/pull/468) Fast storage optimization for queries and iterations
- [\#452](https://github.com/cosmos/iavl/pull/452) Optimization: remove unnecessary (\*bytes.Buffer).Reset right after creating buffer.
- [\#445](https://github.com/cosmos/iavl/pull/445) Bump github.com/tendermint/tendermint to v0.35.0
- [\#453](https://github.com/cosmos/iavl/pull/453),[\#456](https://github.com/cosmos/iavl/pull/456) Optimization: buffer reuse
- [\#474](https://github.com/cosmos/iavl/pull/474) bump github.com/confio/ics23 to v0.7
- [\#475](https://github.com/cosmos/iavl/pull/475) Use go v1.17

## 0.17.3 (December 1, 2021)

### Bug Fixes

- [\#448](https://github.com/cosmos/iavl/pull/448) Change RMutex to Mutex in `VersionExists()`.

## 0.17.2 (November 13, 2021)

### Improvements

- [\#440](https://github.com/cosmos/iavl/pull/440) Introduce Cosmos SDK iterator type directly into IAVL. Improves the iterator performance by 40%.

## 0.17.1 (September 15, 2021)

### Bug Fixes

- [\#432](https://github.com/cosmos/iavl/pull/432) Fix race condition related to Cosmos SDK and nodeDB usage.

## 0.17.0 (August 31, 2021)

### Improvements

- Various performance improvements. Credits: Orijtech.
- Updating dependencies

### CLI Breaking Changes

- [\#396](https://github.com/cosmos/iavl/pull/396) Add "prefix" arg to iaviewer.

## 0.16.0 (May 04, 2021)

### Breaking Changes

- [\#355](https://github.com/cosmos/iavl/pull/355) `Get` in `iavlServer` no longer returns an error if the requested key does not exist. `GetResponse` now contains a `NotFound` boolean to indicate that a key does not exist, and the returned index will be that of the next occupied key.

### Improvements

- [\#355](https://github.com/cosmos/iavl/pull/355) Add support for `GetByIndex` to `iavlServer` and RPC interface.

### Bug Fixes

- [\#385](https://github.com/cosmos/iavl/pull/385) Fix `GetVersioned` - now it works with `LazyLoadVersion`.
- [\#374](https://github.com/cosmos/iavl/pull/374) Fix large genesis file commit.

## 0.15.3 (December 21, 2020)

Special thanks to external contributors on this release: @odeke-em

### Improvements

- [\#352](https://github.com/cosmos/iavl/pull/352) Reuse buffer to improve performance of `GetMembershipProof()` and `GetNonMembershipProof()`.

## 0.15.2 (December 14, 2020)

Special thanks to external contributors on this release: @odeke-em

### Bug Fixes

- [\#347](https://github.com/cosmos/iavl/pull/347) Fix another integer overflow in `decodeBytes()` that can cause panics for certain inputs. The `ValueOp` and `AbsenceOp` proof decoders are vulnerable to this via malicious inputs since 0.15.0.

- [\#349](https://github.com/cosmos/iavl/pull/349) Fix spurious blank lines in `PathToLeaf.String()`.

## 0.15.1 (December 13, 2020)

Special thanks to external contributors on this release: @odeke-em

### Bug Fixes

- [\#340](https://github.com/cosmos/iavl/pull/340) Fix integer overflow in `decodeBytes()` that can cause panics on 64-bit systems and out-of-memory issues on 32-bit systems. The `ValueOp` and `AbsenceOp` proof decoders are vulnerable to this via malicious inputs. The bug was introduced in 0.15.0.

## 0.15.0 (November 23, 2020)

The IAVL project has moved from https://github.com/tendermint/iavl to
https://github.com/cosmos/iavl. This changes the module import path, which is now
`github.com/cosmos/iavl`.

Users upgrading from 0.13 should read important upgrade information in the 0.14.0 release below.

### Breaking Changes

- [\#285](https://github.com/cosmos/iavl/pull/285) The module path has changed from
  `github.com/tendermint/iavl` to `github.com/cosmos/iavl`.

- [\#304](https://github.com/cosmos/iavl/pull/304) Empty trees now return hashes rather than `nil`
  from e.g. `Hash()`, `WorkingHash()`, and `SaveVersion()`, for conformance with RFC-6962.

- [\#317](https://github.com/cosmos/iavl/pull/317) `LoadVersion()` and `LazyLoadVersion()` now
  error if called with a positive version number on an empty tree.

### Improvements

- [\#296](https://github.com/cosmos/iavl/pull/296) Add `iavlserver`, a gRPC/REST API server.

- [\#276](https://github.com/cosmos/iavl/pull/276/files) Introduced
  `ImmutableTree.GetMembershipProof()` and `GetNonMembershipProof()` to return ics23 ExistenceProof
  and NonExistenceProof respectively.

- [\#265](https://github.com/cosmos/iavl/pull/265) Encoding of tree nodes and proofs is now done
  using the Go stdlib and Protobuf instead of Amino. The binary encoding is identical.

### Bug Fixes

- [\#309](https://github.com/cosmos/iavl/pull/309) Allow `SaveVersion()` for old, empty versions as
  long as the new version is identical.

## 0.14.3 (November 23, 2020)

Special thanks to external contributors on this release: @klim0v

### Bug Fixes

- [\#324](https://github.com/cosmos/iavl/pull/324) Fix `DeleteVersions` not properly removing
  orphans, and add `DeleteVersionsRange` to delete a range.

## 0.14.2 (October 12, 2020)

### Bug Fixes

- [\#318](https://github.com/cosmos/iavl/pull/318) Fix constant overflow when compiling for 32bit machines.

## 0.14.1 (October 9, 2020)

### Improvements

- [\#299](https://github.com/cosmos/iavl/pull/299) Added `Options.InitialVersion` to specify the
  initial version for new IAVL trees.

- [\#312](https://github.com/cosmos/iavl/pull/312) Added `MutableTree.SetInitialVersion()` to
  set the initial version after tree initialization.

### Bug Fixes

- [\#288](https://github.com/cosmos/iavl/pull/288) Fix panics when generating proofs for keys that are all `0xFF`.

## 0.14.0 (July 2, 2020)

**Important information:** the pruning functionality introduced with IAVL 0.13.0 via the options
`KeepEvery` and `KeepRecent` has problems with data corruption, performance, and memory usage. For
these reasons, this functionality has now been removed. All 0.13 users are urged to upgrade, and to
not change their pruning settings while on 0.13.

Make sure to follow these instructions when upgrading, to avoid data corruption:

- If using `KeepEvery: 1` (the default) then upgrading to 0.14 is safe.

- Otherwise, upgrade after saving a multiple of `KeepEvery` - for example, with `KeepEvery: 1000`
  stop 0.13 after saving e.g. version `7000` to disk. A later version must never have been saved
  to the tree. Upgrading to 0.14 is then safe.

- Otherwise, consider using the `Repair013Orphans()` function to repair faulty data in databases
  last written to by 0.13. This must be done before opening the database with IAVL 0.14, and a
  database backup should be taken first. Upgrading to 0.14 is then safe.

- Otherwise, after upgrading to 0.14, do not delete the last version saved to disk by 0.13 - this
  contains incorrect data that may cause data corruption when deleted, making the database
  unusable. For example, with `KeepEvery: 1000` then stopping 0.13 at version `7364` (saving
  `7000` to disk) and upgrading to 0.14 means version `7000` must never be deleted.

  It may be possible to delete it if the exact same sequence of changes have been written to the
  newer versions as before the upgrade, and all versions between `7000` and `7364` are deleted
  first, but thorough testing and backups are recommended if attempting this.

Users wishing to prune historical versions can do so via `MutableTree.DeleteVersion()`.

### Breaking Changes

- [\#274](https://github.com/cosmos/iavl/pull/274) Remove pruning options `KeepEvery` and
  `KeepRecent` (see warning above) and the `recentDB` parameter to `NewMutableTreeWithOpts()`.

### Improvements

- [\#282](https://github.com/cosmos/iavl/pull/282) Add `Repair013Orphans()` to repair faulty
  orphans in a database last written to by IAVL 0.13.x

- [\#271](https://github.com/cosmos/iavl/pull/271) Add `MutableTree.DeleteVersions()` for deleting
  multiple versions.

- [\#235](https://github.com/cosmos/iavl/pull/235) Reduce `ImmutableTree.Export()` buffer size from
  64 to 32 nodes.

### Bug Fixes

- [\#281](https://github.com/cosmos/iavl/pull/281) Remove unnecessary Protobuf dependencies.

- [\#275](https://github.com/cosmos/iavl/pull/275) Fix data corruption with
  `LoadVersionForOverwriting`.

## 0.13.3 (April 5, 2020)

### Bug Fixes

- [import][\#230](https://github.com/tendermint/iavl/pull/230) Set correct version when committing an empty import.

## 0.13.2 (March 18, 2020)

### Improvements

- [\#213] Added `ImmutableTree.Export()` and `MutableTree.Import()` to export tree contents at a specific version and import it to recreate an identical tree.

## 0.13.1 (March 13, 2020)

### Improvements

- [dep][\#220](https://github.com/tendermint/iavl/pull/220) Update tm-db to 0.5.0, which includes a new B-tree based MemDB used by IAVL for non-persisted versions.

### Bug Fixes

- [nodedb][\#219](https://github.com/tendermint/iavl/pull/219) Fix a concurrent database access issue when deleting orphans.

## 0.13.0 (January 16, 2020)

Special thanks to external contributors on this release:
@rickyyangz, @mattkanwisher

### BREAKING CHANGES

- [pruning][\#158](https://github.com/tendermint/iavl/pull/158) NodeDB constructor must provide `keepRecent` and `keepEvery` fields to define PruningStrategy. All Save functionality must specify whether they should flushToDisk as well using `flushToDisk` boolean argument. All Delete functionality must specify whether object should be deleted from memory only using the `memOnly` boolean argument.
- [dep][\#194](https://github.com/tendermint/iavl/pull/194) Update tm-db to 0.4.0 this includes interface breaking to return errors.

### IMPROVEMENTS

### Bug Fix

- [orphans][#177](https://github.com/tendermint/iavl/pull/177) Collect all orphans after remove (@rickyyangz)

## 0.12.4 (July 31, 2019)

### IMPROVEMENTS

- [\#46](https://github.com/tendermint/iavl/issues/46) Removed all instances of cmn (tendermint/tendermint/libs/common)

## 0.12.3 (July 12, 2019)

Special thanks to external contributors on this release:
@ethanfrey

IMPROVEMENTS

- Implement LazyLoadVersion (@alexanderbez)
  LazyLoadVersion attempts to lazy load only the specified target version
  without loading previous roots/versions. - see [goDoc](https://godoc.org/github.com/tendermint/iavl#MutableTree.LazyLoadVersion)
- Move to go.mod (@Liamsi)
- `iaviewer` command to visualize IAVL database from leveldb (@ethanfrey)

## 0.12.2 (March 13, 2019)

IMPROVEMENTS

- Use Tendermint v0.30.2 and close batch after write (related pull request in Tendermint: https://github.com/tendermint/tendermint/pull/3397)

## 0.12.1 (February 12, 2019)

IMPROVEMENTS

- Use Tendermint v0.30

## 0.12.0 (November 26, 2018)

BREAKING CHANGES

- Uses new Tendermint ReverseIterator API. See https://github.com/tendermint/tendermint/pull/2913

## 0.11.1 (October 29, 2018)

IMPROVEMENTS

- Uses GoAmino v0.14

## 0.11.0 (September 7, 2018)

BREAKING CHANGES

- Changed internal database key format to store int64 key components in a full 8-byte fixed width ([#107])
- Removed some architecture dependent methods (e.g., use `Get` instead of `Get64` etc) ([#96])

IMPROVEMENTS

- Database key format avoids use of fmt.Sprintf fmt.Sscanf leading to ~10% speedup in benchmark BenchmarkTreeLoadAndDelete ([#107], thanks to [@silasdavis])

[#107]: https://github.com/tendermint/iavl/pull/107
[@silasdavis]: https://github.com/silasdavis
[#96]: https://github.com/tendermint/iavl/pull/96

## 0.10.0

BREAKING CHANGES

- refactored API for clean separation of [mutable][1] and [immutable][2] tree (#92, #88);
  with possibility to:
  - load read-only snapshots at previous versions on demand
  - load mutable trees at the most recently saved tree

[1]: https://github.com/tendermint/iavl/blob/9e62436856efa94c1223043be36ebda01ae0b6fc/mutable_tree.go#L14-L21
[2]: https://github.com/tendermint/iavl/blob/9e62436856efa94c1223043be36ebda01ae0b6fc/immutable_tree.go#L10-L17

BUG FIXES

- remove memory leaks (#92)

IMPROVEMENTS

- Change tendermint dep to ^v0.22.0 (#91)

## 0.10.0 (July 11, 2018)

BREAKING CHANGES

- getRangeProof and Get\[Versioned\]\[Range\]WithProof return nil proof/error if tree is empty.

## 0.9.2 (July 3, 2018)

IMPROVEMENTS

- some minor changes: mainly lints, updated parts of documentation, unexported some helpers (#80)

## 0.9.1 (July 1, 2018)

IMPROVEMENTS

- RangeProof.ComputeRootHash() to compute root rather than provide as in Verify(hash)
- RangeProof.Verify\*() first require .Verify(root), which memoizes

## 0.9.0 (July 1, 2018)

BREAKING CHANGES

- RangeProof.VerifyItem doesn't require an index.
- Only return values in range when getting proof.
- Return keys as well.

BUG FIXES

- traversal bugs in traverseRange.

## 0.8.2

- Swap `tmlibs` for `tendermint/libs`
- Remove `sha256truncated` in favour of `tendermint/crypto/tmhash` - same hash
  function but technically a breaking change to the API, though unlikely to effect anyone.

NOTE this means IAVL is now dependent on Tendermint Core for the libs (since it
makes heavy use of the `db` package). Ideally, that dependency would be
abstracted away, and/or this repo will be merged into the Cosmos-SDK, which is
currently is primary consumer. Once it achieves greater stability, we could
consider breaking it out into it's own repo again.

## 0.8.1

_July 1st, 2018_

BUG FIXES

- fix bug in iterator going outside its range

## 0.8.0 (June 24, 2018)

BREAKING CHANGES

- Nodes are encoded using proto3/amino style integers and byte slices (ie. varints and
  varint prefixed byte slices)
- Unified RangeProof
- Proofs are encoded using Amino
- Hash function changed from RIPEMD160 to the first 20 bytes of SHA256 output

## 0.7.0 (March 21, 2018)

BREAKING CHANGES

- LoadVersion and Load return the loaded version number
  - NOTE: this behaviour was lost previously and we failed to document in changelog,
    but now it's back :)

## 0.6.1 (March 2, 2018)

IMPROVEMENT

- Remove spurious print statement from LoadVersion

## 0.6.0 (March 2, 2018)

BREAKING CHANGES

- NewTree order of arguments swapped
- int -> int64, uint64 -> int64
- NewNode takes a version
- Node serialization format changed so version is written right after size
- SaveVersion takes no args (auto increments)
- tree.Get -> tree.Get64
- nodeDB.SaveBranch does not take a callback
- orphaningTree.SaveVersion -> SaveAs
- proofInnerNode includes Version
- ReadKeyXxxProof consolidated into ReadKeyProof
- KeyAbsentProof doesn't include Version
- KeyRangeProof.Version -> Versions

FEATURES

- Implement chunking algorithm to serialize entire tree

## 0.5.0 (October 27, 2017)

First versioned release!
(Originally accidentally released as v0.2.0)
