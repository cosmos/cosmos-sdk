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

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking Protobuf, gRPC and REST routes used by end-users.
"CLI Breaking" for breaking CLI commands.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

### Improvements

* (x/gov) [#17387](https://github.com/cosmos/cosmos-sdk/pull/17387) Add `MsgSubmitProposal` `SetMsgs` method. 
* (x/gov) [#17354](https://github.com/cosmos/cosmos-sdk/issues/17354) Emit `VoterAddr` in `proposal_vote` event.
* (x/genutil) [#17296](https://github.com/cosmos/cosmos-sdk/pull/17296) Add `MigrateHandler` to allow reuse migrate genesis related function.
    * In v0.46, v0.47 this function is additive to the `genesis migrate` command. However in v0.50+, adding custom migrations to the `genesis migrate` command is directly possible.

## [v0.46.14](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.14) - 2023-07-17

### Features

* (sims) [#16656](https://github.com/cosmos/cosmos-sdk/pull/16656) Add custom max gas for block for sim config with unlimited as default.

### Improvements

* (cli) [#16856](https://github.com/cosmos/cosmos-sdk/pull/16856) Improve `simd prune` UX by using the app default home directory and set pruning method as first variable argument (defaults to default). `pruning.PruningCmd` rest unchanged for API compability, use `pruning.Cmd` instead.
* (deps) [#16553](https://github.com/cosmos/cosmos-sdk/pull/16553) Bump CometBFT to [v0.34.29](https://github.com/cometbft/cometbft/blob/v0.34.29/CHANGELOG.md#v03429).

### Bug Fixes

* (x/auth) [#16994](https://github.com/cosmos/cosmos-sdk/pull/16994) Fix regression where querying transactions events with `<=` or `>=` would not work.
* (x/auth) [#16554](https://github.com/cosmos/cosmos-sdk/pull/16554) `ModuleAccount.Validate` now reports a nil `.BaseAccount` instead of panicking.
* [#16588](https://github.com/cosmos/cosmos-sdk/pull/16588) Propogate snapshotter failures to the caller, (it would create an empty snapshot silently before).
* (types) [#15433](https://github.com/cosmos/cosmos-sdk/pull/15433) Allow disabling of account address caches (for printing bech32 account addresses).

## [v0.46.13](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.13) - 2023-06-08

### Features 

* (snapshots) [#16060](https://github.com/cosmos/cosmos-sdk/pull/16060) Support saving and restoring snapshot locally.
* (baseapp) [#16290](https://github.com/cosmos/cosmos-sdk/pull/16290) Add circuit breaker setter in baseapp.
* (x/group) [#16191](https://github.com/cosmos/cosmos-sdk/pull/16191) Add EventProposalPruned event to group module whenever a proposal is pruned.

### Improvements

* (deps) [#15973](https://github.com/cosmos/cosmos-sdk/pull/15973) Bump CometBFT to [v0.34.28](https://github.com/cometbft/cometbft/blob/v0.34.28/CHANGELOG.md#v03428).
* (store) [#15683](https://github.com/cosmos/cosmos-sdk/pull/15683) `rootmulti.Store.CacheMultiStoreWithVersion` now can handle loading archival states that don't persist any of the module stores the current state has.
* (simapp) [#15903](https://github.com/cosmos/cosmos-sdk/pull/15903) Add `AppStateFnWithExtendedCbs` with moduleStateCb callback function to allow access moduleState. Note, this function is present in `simtestutil` from `v0.47.2+`.
* (gov) [#15979](https://github.com/cosmos/cosmos-sdk/pull/15979) Improve gov error message when failing to convert v1 proposal to v1beta1.
* (server) [#16061](https://github.com/cosmos/cosmos-sdk/pull/16061) Add Comet bootstrap command.
* (store) [#16067](https://github.com/cosmos/cosmos-sdk/pull/16067) Add local snapshots management commands.
* (baseapp) [#16193](https://github.com/cosmos/cosmos-sdk/pull/16193) Add `Close` method to `BaseApp` for custom app to cleanup resource in graceful shutdown.

### Bug Fixes

* Fix [barberry](https://forum.cosmos.network/t/cosmos-sdk-security-advisory-barberry/10825) security vulnerability.
* (cli) [#16312](https://github.com/cosmos/cosmos-sdk/pull/16312) Allow any addresses in `client.ValidatePromptAddress`.
* (store/iavl) [#15717](https://github.com/cosmos/cosmos-sdk/pull/15717) Upstream error on empty version (this change was present on all version but v0.46).

## [v0.46.12](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.12) - 2023-04-04

### Features

* (x/groups) [#14879](https://github.com/cosmos/cosmos-sdk/pull/14879) Add `Query/Groups` query to get all the groups.

### Improvements

* (simapp) [#15305](https://github.com/cosmos/cosmos-sdk/pull/15305) Add `AppStateFnWithExtendedCb` with callback function to extend rawState and `AppStateRandomizedFnWithState` with extra genesisState argument which is the genesis state of the app.
* (x/distribution) [#15462](https://github.com/cosmos/cosmos-sdk/pull/15462) Add delegator address to the event for withdrawing delegation rewards
* [#14019](https://github.com/cosmos/cosmos-sdk/issues/14019) Remove the interface casting to allow other implementations of a `CommitMultiStore`.

### Bug Fixes

* (x/auth/vesting) [#15383](https://github.com/cosmos/cosmos-sdk/pull/15383) Add extra checks when creating a periodic vesting account.
* (x/gov) [#13051](https://github.com/cosmos/cosmos-sdk/pull/13051) In SubmitPropsal, when a legacy msg fails it's handler call, wrap the error as ErrInvalidProposalContent (instead of ErrNoProposalHandlerExists).

## [v0.46.11](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.11) - 2023-03-03

### Improvements

* (deps) Migrate to [CometBFT](https://github.com/cometbft/cometbft). Follow the instructions in the [release notes](./RELEASE_NOTES.md).
* (store) [#15152](https://github.com/cosmos/cosmos-sdk/pull/15152) Remove unmaintained and experimental `store/v2alpha1`.
* (store) [#14410](https://github.com/cosmos/cosmos-sdk/pull/14410) `rootmulti.Store.loadVersion` has validation to check if all the module stores' height is correct, it will error if any module store has incorrect height.

### Bug Fixes

* [#15243](https://github.com/cosmos/cosmos-sdk/pull/15243) `LatestBlockResponse` & `BlockByHeightResponse` types' field `sdk_block` was incorrectly cast `proposer_address` bytes to validator operator address, now to consensus address.

## [v0.46.10](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.10) - 2023-02-16

### Improvements

* (cli) [#14953](https://github.com/cosmos/cosmos-sdk/pull/14953) Enable profiling block replay during abci handshake with `--cpu-profile`.

## [v0.46.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.9) - 2023-02-07

### Improvements

* (deps) [#14846](https://github.com/cosmos/cosmos-sdk/pull/14846) Bump btcd.
* (deps) Bump Tendermint version to [v0.34.26](https://github.com/informalsystems/tendermint/releases/tag/v0.34.26).
* (store) [#14189](https://github.com/cosmos/cosmos-sdk/pull/14189) Add config `iavl-lazy-loading` to enable lazy loading of iavl store, to improve start up time of archive nodes, add method `SetLazyLoading` to `CommitMultiStore` interface.
    * A new field has been added to the app.toml. This alllows nodes with larger databases to startup quicker 

    ```toml
    # IAVLLazyLoading enable/disable the lazy loading of iavl store.
    # Default is false.
    iavl-lazy-loading = ""  
  ```

### Bug Fixes

* (cli) [#14919](https://github.com/cosmos/cosmos-sdk/pull/#14919) Fix never assigned error when write validators.
* (store) [#14798](https://github.com/cosmos/cosmos-sdk/pull/14798) Copy btree to avoid the problem of modify while iteration.
* (cli) [#14799](https://github.com/cosmos/cosmos-sdk/pull/14799) Fix Evidence CLI query flag parsing (backport #13458)

## [v0.46.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.8) - 2023-01-23

### Improvements

* [#13881](https://github.com/cosmos/cosmos-sdk/pull/13881) Optimize iteration on nested cached KV stores and other operations in general.
* (x/gov) [#14347](https://github.com/cosmos/cosmos-sdk/pull/14347) Support `v1.Proposal` message in `v1beta1.Proposal.Content`.
* (deps) Use Informal System fork of Tendermint version to [v0.34.24](https://github.com/informalsystems/tendermint/releases/tag/v0.34.24).

### Bug Fixes

* (x/group) [#14526](https://github.com/cosmos/cosmos-sdk/pull/14526) Fix wrong address set in `EventUpdateGroupPolicy`.
* (ante) [#14448](https://github.com/cosmos/cosmos-sdk/pull/14448) Return anteEvents when postHandler fail.

### API Breaking

* (x/gov) [#14422](https://github.com/cosmos/cosmos-sdk/pull/14422) Remove `Migrate_V046_6_To_V046_7` function which shouldn't be used for chains which already migrated to 0.46.

## [v0.46.7](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.7) - 2022-12-13

### Features

* (client) [#14051](https://github.com/cosmos/cosmos-sdk/pull/14051) Add `--grpc` client option.

### Improvements

* (deps) Bump Tendermint version to [v0.34.24](https://github.com/tendermint/tendermint/releases/tag/v0.34.24).
* [#13651](https://github.com/cosmos/cosmos-sdk/pull/13651) Update `server/config/config.GetConfig` function.
* [#14175](https://github.com/cosmos/cosmos-sdk/pull/14175) Add `server.DefaultBaseappOptions(appopts)` function to reduce boiler plate in root.go. 

### State Machine Breaking

* (x/gov) [#14214](https://github.com/cosmos/cosmos-sdk/pull/14214) Fix gov v0.46 migration to v1 votes.
    * Also provide a helper function `govv046.Migrate_V0466_To_V0467` for migrating a chain already on v0.46 with versions <=v0.46.6 to the latest v0.46.7 correct state.
* (x/group) [#14071](https://github.com/cosmos/cosmos-sdk/pull/14071) Don't re-tally proposal after voting period end if they have been marked as ACCEPTED or REJECTED.

### API Breaking Changes

* (store) [#13516](https://github.com/cosmos/cosmos-sdk/pull/13516) Update State Streaming APIs:
    * Add method `ListenCommit` to `ABCIListener`
    * Move `ListeningEnabled` and  `AddListener` methods to `CommitMultiStore`
    * Remove `CacheWrapWithListeners` from `CacheWrap` and `CacheWrapper` interfaces
    * Remove listening APIs from the caching layer (it should only listen to the `rootmulti.Store`)
    * Add three new options to file streaming service constructor.
    * Modify `ABCIListener` such that any error from any method will always halt the app via `panic`
* (store) [#13529](https://github.com/cosmos/cosmos-sdk/pull/13529) Add method `LatestVersion` to `MultiStore` interface, add method `SetQueryMultiStore` to baesapp to support alternative `MultiStore` implementation for query service.

### Bug Fixes

* (baseapp) [#13983](https://github.com/cosmos/cosmos-sdk/pull/13983) Don't emit duplicate ante-handler events when a post-handler is defined.
* (baseapp) [#14049](https://github.com/cosmos/cosmos-sdk/pull/14049) Fix state sync when interval is zero.
* (store) [#13516](https://github.com/cosmos/cosmos-sdk/pull/13516) Fix state listener that was observing writes at wrong time.

## [v0.46.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.6) - 2022-11-18

### Improvements

* (config) [#13894](https://github.com/cosmos/cosmos-sdk/pull/13894) Support state streaming configuration in `app.toml` template and default configuration.

### Bug Fixes

* (x/gov) [#13918](https://github.com/cosmos/cosmos-sdk/pull/13918) Fix propagation of message errors when executing a proposal.

## [v0.46.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.5) - 2022-11-17

### Features

* (x/bank) [#13891](https://github.com/cosmos/cosmos-sdk/pull/13891) Provide a helper function `Migrate_V0464_To_V0465` for migrating a chain **already on v0.46 with versions <=v0.46.4** to the latest v0.46.5 correct state.

### Improvements

* [#13826](https://github.com/cosmos/cosmos-sdk/pull/13826) Support custom `GasConfig` configuration for applications.
* (deps) Bump Tendermint version to [v0.34.23](https://github.com/tendermint/tendermint/releases/tag/v0.34.23).

### State Machine Breaking

* (x/group) [#13876](https://github.com/cosmos/cosmos-sdk/pull/13876) Fix group MinExecutionPeriod that is checked on execution now, instead of voting period end.

### API Breaking Changes

* (x/group) [#13876](https://github.com/cosmos/cosmos-sdk/pull/13876) Add `GetMinExecutionPeriod` method on DecisionPolicy interface.

### Bug Fixes

* (x/group) [#13869](https://github.com/cosmos/cosmos-sdk/pull/13869) Group members weight must be positive and a finite number.
* (x/bank) [#13821](https://github.com/cosmos/cosmos-sdk/pull/13821) Fix bank store migration of coin metadata.
* (x/group) [#13808](https://github.com/cosmos/cosmos-sdk/pull/13808) Fix propagation of message events to the current context in `EndBlocker`.
* (x/gov) [#13728](https://github.com/cosmos/cosmos-sdk/pull/13728) Fix propagation of message events to the current context in `EndBlocker`.
* (store) [#13803](https://github.com/cosmos/cosmos-sdk/pull/13803) Add an error log if IAVL set operation failed.
* [#13861](https://github.com/cosmos/cosmos-sdk/pull/13861) Allow `_` characters in tx event queries, i.e. `GetTxsEvent`.

## [v0.46.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.4) - 2022-11-01

### Features

* (x/auth) [#13612](https://github.com/cosmos/cosmos-sdk/pull/13612) Add `Query/ModuleAccountByName` endpoint for accessing the module account info by module name.

### Improvements

* (deps) Bump IAVL version to [v0.19.4](https://github.com/cosmos/iavl/releases/tag/v0.19.4).

### Bug Fixes

* (x/auth/tx) [#12474](https://github.com/cosmos/cosmos-sdk/pull/12474) Remove condition in GetTxsEvent that disallowed multiple equal signs, which would break event queries with base64 strings (i.e. query by signature).
* (store) [#13530](https://github.com/cosmos/cosmos-sdk/pull/13530) Fix app-hash mismatch if upgrade migration commit is interrupted.

### CLI Breaking Changes

* [#13656](https://github.com/cosmos/cosmos-sdk/pull/13659) Rename `server.FlagIAVLFastNode` to `server.FlagDisableIAVLFastNode` for clarity.

### API Breaking Changes

* (context) [#13063](https://github.com/cosmos/cosmos-sdk/pull/13063) Update `Context#CacheContext` to automatically emit all events on the parent context's `EventManager`.

## [v0.46.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.3) - 2022-10-20

ATTENTION:

This is a security release for the [Dragonberry security advisory](https://forum.cosmos.network/t/ibc-security-advisory-dragonberry/7702). 

All users should upgrade immediately.

Users *must* add a replace directive in their go.mod for the new `ics23` package in the SDK:

```go
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```

### Features

* [#13435](https://github.com/cosmos/cosmos-sdk/pull/13435) Extend error context when a simulation fails.
* (grpc) [#13485](https://github.com/cosmos/cosmos-sdk/pull/13485) Implement a new gRPC query, `/cosmos/base/node/v1beta1/config`, which provides operator configuration.
* [#13577](https://github.com/cosmos/cosmos-sdk/pull/13577) Added `ApplicationQueryService` interface (the related method is added directly to the `Application` interface and `ApplicationQueryService` is removed in the future version). Applications implementing `ApplicationQueryService` enabling registration of module external gRPC services. When implemented the SDK will automatically register chain information query service introduced in [#13485](https://github.com/cosmos/cosmos-sdk/pull/13485).
* (cli) [#13147](https://github.com/cosmos/cosmos-sdk/pull/13147) Add the `--append` flag to the `sign-batch` CLI cmd to combine the messages and sign those txs which are created with `--generate-only`.
* (cli) [#13454](https://github.com/cosmos/cosmos-sdk/pull/13454) `sign-batch` CLI can now read multiple transaction files.

### Improvements

* [#13586](https://github.com/cosmos/cosmos-sdk/pull/13586) Bump Tendermint to `v0.34.22`.
* (auth) [#13460](https://github.com/cosmos/cosmos-sdk/pull/13460) The `q auth address-by-id` CLI command has been renamed to `q auth address-by-acc-num` to be more explicit. However, the old `address-by-id` version is still kept as an alias, for backwards compatibility.
* [#13433](https://github.com/cosmos/cosmos-sdk/pull/13433) Remove dead code in cacheMergeIterator `Domain()`.

### Bug Fixes

* Implement dragonberry security patch.
    * For applying the patch please refer to the [RELEASE NOTES](./RELEASE_NOTES.md)
* (store) [#13459](https://github.com/cosmos/cosmos-sdk/pull/13459) Don't let state listener observe the uncommitted writes.
* [#12548](https://github.com/cosmos/cosmos-sdk/pull/12548) Prevent signing from wrong key while using multisig.

### API Breaking Changes

* (server) [#13485](https://github.com/cosmos/cosmos-sdk/pull/13485) The `Application` service now requires the `RegisterNodeService` method to be implemented.

## [v0.46.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.2) - 2022-10-03

### API Breaking Changes

* (cli) [#13089](https://github.com/cosmos/cosmos-sdk/pull/13089) Fix rollback command don't actually delete multistore versions, added method `RollbackToVersion` to interface `CommitMultiStore` and added method `CommitMultiStore` to `Application` interface.
* (cli) [#13089](https://github.com/cosmos/cosmos-sdk/pull/13089) `NewRollbackCmd` now takes an `appCreator types.AppCreator`.

### Features

* (baseapp) [#12168](https://github.com/cosmos/cosmos-sdk/pull/12168) Add `SetMsgServiceRouter` to `BaseApp`.
* (cli) [#13207](https://github.com/cosmos/cosmos-sdk/pull/13207) Reduce user's password prompts when calling keyring `List()` function.
* (cli) [#13353](https://github.com/cosmos/cosmos-sdk/pull/13353) Add `tx group draft-proposal` command for generating group proposal JSONs (skeleton).
* (cli) [#13304](https://github.com/cosmos/cosmos-sdk/pull/13304) Add `tx gov draft-proposal` command for generating proposal JSONs (skeleton).
* (x/authz) [#13047](https://github.com/cosmos/cosmos-sdk/pull/13047) Add a GetAuthorization function to the keeper.
* (cli) [#12742](https://github.com/cosmos/cosmos-sdk/pull/12742) Add the `prune` CLI cmd to manually prune app store history versions based on the pruning options.

### Improvements

* [#13323](https://github.com/cosmos/cosmos-sdk/pull/13323) Ensure `withdraw_rewards` rewards are emitted from all actions that result in rewards being withdrawn.
* [#13233](https://github.com/cosmos/cosmos-sdk/pull/13233) Add `--append` to `add-genesis-account` sub-command to append new tokens after an account is already created.
* (x/group) [#13214](https://github.com/cosmos/cosmos-sdk/pull/13214) Add `withdraw-proposal` command to group module's CLI transaction commands.
* (x/auth) [#13048](https://github.com/cosmos/cosmos-sdk/pull/13048) Add handling of AccountNumberStoreKeyPrefix to the simulation decoder.
* (simapp) [#13108](https://github.com/cosmos/cosmos-sdk/pull/13108) Call `SetIAVLCacheSize` with the configured value in simapp.
* [#13318](https://github.com/cosmos/cosmos-sdk/pull/13318) Keep the balance query endpoint compatible with legacy blocks.
* [#13321](https://github.com/cosmos/cosmos-sdk/pull/13321) Add flag to disable fast node migration and usage. 

### Bug Fixes

* (types) [#13265](https://github.com/cosmos/cosmos-sdk/pull/13265) Correctly coalesce coins even with repeated denominations & simplify logic.
* (x/auth) [#13200](https://github.com/cosmos/cosmos-sdk/pull/13200) Fix wrong sequences in `sign-batch`.
* (export) [#13029](https://github.com/cosmos/cosmos-sdk/pull/13029) Fix exporting the blockParams regression.
* [#13046](https://github.com/cosmos/cosmos-sdk/pull/13046) Fix missing return statement in BaseApp.Query.
* (store) [#13336](https://github.com/cosmos/cosmos-sdk/pull/13336) Call streaming listeners for deliver tx event, it was removed accidentally, backport #13334.
* (grpc) [#13417](https://github.com/cosmos/cosmos-sdk/pull/13417) fix grpc query panic that could crash the node (backport #13352).
* (grpc) [#13418](https://github.com/cosmos/cosmos-sdk/pull/13418) Add close for grpc only mode.

## [v0.46.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.1) - 2022-08-24

### Improvements

* [#12953](https://github.com/cosmos/cosmos-sdk/pull/12953) Change the default priority mechanism to be based on gas price.
* [#12981](https://github.com/cosmos/cosmos-sdk/pull/12981) Return proper error when parsing telemetry configuration.
* [#12969](https://github.com/cosmos/cosmos-sdk/pull/12969) Bump Tendermint to `v0.34.21` and IAVL to `v0.19.1`.
* [#12885](https://github.com/cosmos/cosmos-sdk/pull/12885) Amortize cost of processing cache KV store.
* (events) [#12850](https://github.com/cosmos/cosmos-sdk/pull/12850) Add a new `fee_payer` attribute to the `tx` event that is emitted from the `DeductFeeDecorator` AnteHandler decorator.
* (x/params) [#12615](https://github.com/cosmos/cosmos-sdk/pull/12615) Add `GetParamSetIfExists` function to params `Subspace` to prevent panics on breaking changes.
* (x/bank) [#12674](https://github.com/cosmos/cosmos-sdk/pull/12674) Add convenience function `CreatePrefixedAccountStoreKey()` to construct key to access account's balance for a given denom.
* [#12877](https://github.com/cosmos/cosmos-sdk/pull/12877) Bumped cosmossdk.io/math to v1.0.0-beta.3
* [#12693](https://github.com/cosmos/cosmos-sdk/pull/12693) Make sure the order of each node is consistent when emitting proto events.

### Bug Fixes

* (x/group) [#12888](https://github.com/cosmos/cosmos-sdk/pull/12888) Fix event propagation to the current context of `x/group` message execution `[]sdk.Result`.
* (x/upgrade) [#12906](https://github.com/cosmos/cosmos-sdk/pull/12906) Fix upgrade failure by moving downgrade verification logic after store migration.
* (store) [#12945](https://github.com/cosmos/cosmos-sdk/pull/12945) Fix nil end semantics in store/cachekv/iterator when iterating a dirty cache.

## [v0.46.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.0) - 2022-07-26

### Features

* (types) [#11985](https://github.com/cosmos/cosmos-sdk/pull/11985) Add a `Priority` field on `sdk.Context`, which represents the CheckTx priority field. It is only used during CheckTx.
* (gRPC) [#11889](https://github.com/cosmos/cosmos-sdk/pull/11889) Support custom read and write gRPC options in `app.toml`. See `max-recv-msg-size` and `max-send-msg-size` respectively.
* (cli) [\#11738](https://github.com/cosmos/cosmos-sdk/pull/11738) Add `tx auth multi-sign` as alias of `tx auth multisign` for consistency with `multi-send`.
* (cli) [\#11738](https://github.com/cosmos/cosmos-sdk/pull/11738) Add `tx bank multi-send` command for bulk send of coins to multiple accounts.
* (grpc) [\#11642](https://github.com/cosmos/cosmos-sdk/pull/11642) Implement `ABCIQuery` in the Tendermint gRPC service, which proxies ABCI `Query` requests directly to the application.
* (x/upgrade) [\#11551](https://github.com/cosmos/cosmos-sdk/pull/11551) Update `ScheduleUpgrade` for chains to schedule an automated upgrade on `BeginBlock` without having to go though governance.
* (cli) [\#11548](https://github.com/cosmos/cosmos-sdk/pull/11548) Add Tendermint's `inspect` command to the `tendermint` sub-command.
* (tx) [#\11533](https://github.com/cosmos/cosmos-sdk/pull/11533) Register [`EIP191`](https://eips.ethereum.org/EIPS/eip-191) as an available `SignMode` for chains to use.
* (x/genutil) [\#11500](https://github.com/cosmos/cosmos-sdk/pull/11500) Fix GenTx validation and adjust error messages
* [\#11430](https://github.com/cosmos/cosmos-sdk/pull/11430) Introduce a new `grpc-only` flag, such that when enabled, will start the node in a query-only mode. Note, gRPC MUST be enabled with this flag.
* (x/bank) [\#11417](https://github.com/cosmos/cosmos-sdk/pull/11417) Introduce a new `SpendableBalances` gRPC query that retrieves an account's total (paginated) spendable balances.
* [\#11441](https://github.com/cosmos/cosmos-sdk/pull/11441) Added a new method, `IsLTE`, for `types.Coin`. This method is used to check if a `types.Coin` is less than or equal to another `types.Coin`.
* (x/upgrade) [\#11116](https://github.com/cosmos/cosmos-sdk/pull/11116) `MsgSoftwareUpgrade` and `MsgCancelUpgrade` have been added to support v1beta2 msgs-based gov proposals.
* [\#11308](https://github.com/cosmos/cosmos-sdk/pull/11308) Added a mandatory metadata field to Vote in x/gov v1beta2.
* [\#10977](https://github.com/cosmos/cosmos-sdk/pull/10977) Now every cosmos message protobuf definition must be extended with a ``cosmos.msg.v1.signer`` option to signal the signer fields in a language agnostic way.
* [\#10710](https://github.com/cosmos/cosmos-sdk/pull/10710) Chain-id shouldn't be required for creating a transaction with both --generate-only and --offline flags.
* [\#10703](https://github.com/cosmos/cosmos-sdk/pull/10703) Create a new grantee account, if the grantee of an authorization does not exist.
* [\#10592](https://github.com/cosmos/cosmos-sdk/pull/10592) Add a `DecApproxEq` function that checks to see if `|d1 - d2| < tol` for some Dec `d1, d2, tol`.
* [\#9933](https://github.com/cosmos/cosmos-sdk/pull/9933) Introduces the notion of a Cosmos "Scalar" type, which would just be simple aliases that give human-understandable meaning to the underlying type, both in Go code and in Proto definitions.
* [\#9884](https://github.com/cosmos/cosmos-sdk/pull/9884) Provide a new gRPC query handler, `/cosmos/params/v1beta1/subspaces`, that allows the ability to query for all registered subspaces and their respective keys.
* [\#9776](https://github.com/cosmos/cosmos-sdk/pull/9776) Add flag `staking-bond-denom` to specify the staking bond denomination value when initializing a new chain.
* [\#9533](https://github.com/cosmos/cosmos-sdk/pull/9533) Added a new gRPC method, `DenomOwners`, in `x/bank` to query for all account holders of a specific denomination.
* (bank) [\#9618](https://github.com/cosmos/cosmos-sdk/pull/9618) Update bank.Metadata: add URI and URIHash attributes.
* (store) [\#8664](https://github.com/cosmos/cosmos-sdk/pull/8664) Implementation of ADR-038 file StreamingService
* [\#9837](https://github.com/cosmos/cosmos-sdk/issues/9837) `--generate-only` flag can be used with a keyname from the keyring.
* [\#10326](https://github.com/cosmos/cosmos-sdk/pull/10326) `x/authz` add all grants by granter query.
* [\#10944](https://github.com/cosmos/cosmos-sdk/pull/10944) `x/authz` add all grants by grantee query
* [\#10348](https://github.com/cosmos/cosmos-sdk/pull/10348) Add `fee.{payer,granter}` and `tip` fields to StdSignDoc for signing tipped transactions.
* [\#10208](https://github.com/cosmos/cosmos-sdk/pull/10208) Add `TipsTxMiddleware` for transferring tips.
* [\#10379](https://github.com/cosmos/cosmos-sdk/pull/10379) Add validation to `x/upgrade` CLI `software-upgrade` command `--plan-info` value.
* [\#10507](https://github.com/cosmos/cosmos-sdk/pull/10507) Add middleware for tx priority.
* [\#10311](https://github.com/cosmos/cosmos-sdk/pull/10311) Adds cli to use tips transactions. It adds an `--aux` flag to all CLI tx commands to generate the aux signer data (with optional tip), and a new `tx aux-to-fee` subcommand to let the fee payer gather aux signer data and broadcast the tx
* [\#10430](https://github.com/cosmos/cosmos-sdk/pull/10430) ADR-040: Add store/v2 `MultiStore` implementation
* [\#11019](https://github.com/cosmos/cosmos-sdk/pull/11019) Add `MsgCreatePermanentLockedAccount` and CLI method for creating permanent locked account
* [\#10947](https://github.com/cosmos/cosmos-sdk/pull/10947) Add `AllowancesByGranter` query to the feegrant module
* [\#10407](https://github.com/cosmos/cosmos-sdk/pull/10407) Add validation to `x/upgrade` module's `BeginBlock` to check accidental binary downgrades
* (gov) [\#11036](https://github.com/cosmos/cosmos-sdk/pull/11036) Add in-place migrations for 0.43->0.46. Add a `migrate v0.46` CLI command for v0.43->0.46 JSON genesis migration.
* [\#11006](https://github.com/cosmos/cosmos-sdk/pull/11006) Add `debug pubkey-raw` command to allow inspecting of pubkeys in legacy bech32 format
* (x/authz) [\#10714](https://github.com/cosmos/cosmos-sdk/pull/10714) Add support for pruning expired authorizations
* [\#10015](https://github.com/cosmos/cosmos-sdk/pull/10015) ADR-040: ICS-23 proofs for SMT store
* [\#11240](https://github.com/cosmos/cosmos-sdk/pull/11240) Replace various modules `ModuleCdc` with the global `legacy.Cdc`
* [#11179](https://github.com/cosmos/cosmos-sdk/pull/11179) Add state rollback command.
* [\#10794](https://github.com/cosmos/cosmos-sdk/pull/10794) ADR-040: Add State Sync to V2 Store
* [\#11234](https://github.com/cosmos/cosmos-sdk/pull/11234) Add `GRPCClient` field to Client Context. If `GRPCClient` field is set to nil, the `Invoke` method would use ABCI query, otherwise use gprc.
* [\#10962](https://github.com/cosmos/cosmos-sdk/pull/10962) ADR-040: Add state migration from iavl (v1Store) to smt (v2Store)
* (types) [\#10948](https://github.com/cosmos/cosmos-sdk/issues/10948) Add `app-db-backend` to the `app.toml` config to replace the compile-time `types.DBbackend` variable.
* (authz)[\#11060](https://github.com/cosmos/cosmos-sdk/pull/11060) Support grant with no expire time.
* (rosetta) [\#11590](https://github.com/cosmos/cosmos-sdk/pull/11590) Add fee suggestion for rosetta and enable offline mode. Also force set events about Fees to Success to pass reconciliation test.
* (types) [\#11959](https://github.com/cosmos/cosmos-sdk/pull/11959) Added `sdk.Coins.Find` helper method to find a coin by denom.
* (upgrade) [#12603](https://github.com/cosmos/cosmos-sdk/pull/12603) feat: Move AppModule.BeginBlock and AppModule.EndBlock to extension interfaces
* (telemetry) [#12405](https://github.com/cosmos/cosmos-sdk/pull/12405) Add *query* calls metric to telemetry.
* (cli) [#12028](https://github.com/cosmos/cosmos-sdk/pull/12028) Add the `tendermint key-migrate` to perform Tendermint v0.35 DB key migration.
* (query) [#12253](https://github.com/cosmos/cosmos-sdk/pull/12253) Add `GenericFilteredPaginate` to the `query` package to improve UX.

### API Breaking Changes

* (x/auth/ante) [#11985](https://github.com/cosmos/cosmos-sdk/pull/11985) The `MempoolFeeDecorator` has been removed. Instead, the `DeductFeeDecorator` takes a new argument of type `TxFeeChecker`, to define custom fee models. If `nil` is passed to this `TxFeeChecker` argument, then it will default to `checkTxFeeWithValidatorMinGasPrices`, which is the exact same behavior as the old `MempoolFeeDecorator` (i.e. checking fees against validator's own min gas price).
* (x/auth/ante) [#11985](https://github.com/cosmos/cosmos-sdk/pull/11985) The `ExtensionOptionsDecorator` takes an argument of type `ExtensionOptionChecker`. For backwards-compatibility, you can pass `nil`, which defaults to the old behavior of rejecting all tx extensions.
* (crypto/keyring) [#11932](https://github.com/cosmos/cosmos-sdk/pull/11932) Remove `Unsafe*` interfaces from keyring package. Please use interface casting if you wish to access those unsafe functions.
* (types) [#11881](https://github.com/cosmos/cosmos-sdk/issues/11881) Rename `AccAddressFromHex` to `AccAddressFromHexUnsafe`.
* (types) [#11788](https://github.com/cosmos/cosmos-sdk/pull/11788) The `Int` and `Uint` types have been moved to their own dedicated module, `math`. Aliases are kept in the SDK's root `types` package, however, it is encouraged to utilize the new `math` module. As a result, the `Int#ToDec` API has been removed.
* (grpc) [\#11642](https://github.com/cosmos/cosmos-sdk/pull/11642) The `RegisterTendermintService` method in the `tmservice` package now requires a `abciQueryFn` query function parameter.
* [\#11496](https://github.com/cosmos/cosmos-sdk/pull/11496) Refactor abstractions for snapshot and pruning; snapshot intervals eventually pruned; unit tests.
* (types) [\#11689](https://github.com/cosmos/cosmos-sdk/pull/11689) Make `Coins#Sub` and `Coins#SafeSub` consistent with `Coins#Add`.
* (store)[\#11152](https://github.com/cosmos/cosmos-sdk/pull/11152) Remove `keep-every` from pruning options.
* [\#10950](https://github.com/cosmos/cosmos-sdk/pull/10950) Add `envPrefix` parameter to `cmd.Execute`.
* (x/mint) [\#10441](https://github.com/cosmos/cosmos-sdk/pull/10441) The `NewAppModule` function now accepts an inflation calculation function as an argument.
* [\#10295](https://github.com/cosmos/cosmos-sdk/pull/10295) Remove store type aliases from /types
* [\#9695](https://github.com/cosmos/cosmos-sdk/pull/9695) Migrate keys from `Info` (serialized as amino) -> `Record` (serialized as proto)
    * Add new `codec.Codec` argument in:
        * `keyring.NewInMemory`
        * `keyring.New`
    * Rename:
        * `SavePubKey` to `SaveOfflineKey`.
        * `NewMultiInfo`, `NewLedgerInfo`  to `NewLegacyMultiInfo`, `newLegacyLedgerInfo`  respectively.  Move them into `legacy_info.go`.
        * `NewOfflineInfo` to `newLegacyOfflineInfo` and move it to `migration_test.go`.
    * Return:
    *`keyring.Record, error` in `SaveOfflineKey`, `SaveLedgerKey`, `SaveMultiSig`, `Key` and `KeyByAddress`.
    *`keyring.Record` instead of `Info` in `NewMnemonic` and `List`.
    * Remove `algo` argument from :
        * `SaveOfflineKey`
    * Take `keyring.Record` instead of `Info` as first argument in:
        * `MkConsKeyOutput`
        * `MkValKeyOutput`
        * `MkAccKeyOutput`
* [\#10022](https://github.com/cosmos/cosmos-sdk/pull/10022) `AuthKeeper` interface in `x/auth` now includes a function `HasAccount`.
* [\#9759](https://github.com/cosmos/cosmos-sdk/pull/9759) `NewAccountKeeeper` in `x/auth` now takes an additional `bech32Prefix` argument that represents `sdk.Bech32MainPrefix`.
* [\#9628](https://github.com/cosmos/cosmos-sdk/pull/9628) Rename `x/{mod}/legacy` to `x/{mod}/migrations`.
* [\#9571](https://github.com/cosmos/cosmos-sdk/pull/9571) Implemented error handling for staking hooks, which now return an error on failure.
* [\#9427](https://github.com/cosmos/cosmos-sdk/pull/9427) Move simapp `FundAccount` and `FundModuleAccount` to `x/bank/testutil`
* (client/tx) [\#9421](https://github.com/cosmos/cosmos-sdk/pull/9421/) `BuildUnsignedTx`, `BuildSimTx`, `PrintUnsignedStdTx` functions are moved to
  the Tx Factory as methods.
* (client/keys) [\#9407](https://github.com/cosmos/cosmos-sdk/pull/9601) Added `keys rename` CLI command and `Keyring.Rename` interface method to rename a key in the keyring.
* (x/slashing) [\#9458](https://github.com/cosmos/cosmos-sdk/pull/9458) Coins burned from slashing is now returned from Slash function and included in Slash event.
* [\#9246](https://github.com/cosmos/cosmos-sdk/pull/9246) The `New` method for the network package now returns an error.
* [\#9519](https://github.com/cosmos/cosmos-sdk/pull/9519) `DeleteDeposits` renamed to `DeleteAndBurnDeposits`, `RefundDeposits` renamed to `RefundAndDeleteDeposits`
* (codec) [\#9521](https://github.com/cosmos/cosmos-sdk/pull/9521) Removed deprecated `clientCtx.JSONCodec` from `client.Context`.
* (codec) [\#9521](https://github.com/cosmos/cosmos-sdk/pull/9521) Rename `EncodingConfig.Marshaler` to `Codec`.
* [\#9594](https://github.com/cosmos/cosmos-sdk/pull/9594) `RESTHandlerFn` argument is removed from the `gov/NewProposalHandler`.
* [\#9594](https://github.com/cosmos/cosmos-sdk/pull/9594) `types/rest` package moved to `testutil/rest`.
* [\#9432](https://github.com/cosmos/cosmos-sdk/pull/9432) `ConsensusParamsKeyTable` moved from `params/keeper` to `params/types`
* [\#9576](https://github.com/cosmos/cosmos-sdk/pull/9576) Add debug error message to `sdkerrors.QueryResult` when enabled
* [\#9650](https://github.com/cosmos/cosmos-sdk/pull/9650) Removed deprecated message handler implementation from the SDK modules.
* [\#10248](https://github.com/cosmos/cosmos-sdk/pull/10248) Remove unused `KeyPowerReduction` variable from x/staking types.
* (x/bank) [\#9832](https://github.com/cosmos/cosmos-sdk/pull/9832) `AddressFromBalancesStore` renamed to `AddressAndDenomFromBalancesStore`.
* (tests) [\#9938](https://github.com/cosmos/cosmos-sdk/pull/9938) `simapp.Setup` accepts additional `testing.T` argument.
* (baseapp) [\#11979](https://github.com/cosmos/cosmos-sdk/pull/11979) Rename baseapp simulation helper methods `baseapp.{Check,Deliver}` to `baseapp.Sim{Check,Deliver}`.
* (x/gov) [\#10373](https://github.com/cosmos/cosmos-sdk/pull/10373) Removed gov `keeper.{MustMarshal, MustUnmarshal}`.
* [\#10348](https://github.com/cosmos/cosmos-sdk/pull/10348) StdSignBytes takes a new argument of type `*tx.Tip` for signing over tips using LEGACY_AMINO_JSON.
* [\#10208](https://github.com/cosmos/cosmos-sdk/pull/10208) The `x/auth/signing.Tx` interface now also includes a new `GetTip() *tx.Tip` method for verifying tipped transactions. The `x/auth/types` expected BankKeeper interface now expects the `SendCoins` method too.
* [\#10612](https://github.com/cosmos/cosmos-sdk/pull/10612) `baseapp.NewBaseApp` constructor function doesn't take the `sdk.TxDecoder` anymore. This logic has been moved into the TxDecoderMiddleware.
* [\#10692](https://github.com/cosmos/cosmos-sdk/pull/10612) `SignerData` takes 2 new fields, `Address` and `PubKey`, which need to get populated when using SIGN_MODE_DIRECT_AUX.
* [\#10748](https://github.com/cosmos/cosmos-sdk/pull/10748) Move legacy `x/gov` api to `v1beta1` directory.
* [\#10816](https://github.com/cosmos/cosmos-sdk/pull/10816) Reuse blocked addresses from the bank module. No need to pass them to distribution.
* [\#10852](https://github.com/cosmos/cosmos-sdk/pull/10852) Move `x/gov/types` to `x/gov/types/v1beta2`.
* [\#10922](https://github.com/cosmos/cosmos-sdk/pull/10922), [/#10957](https://github.com/cosmos/cosmos-sdk/pull/10957) Move key `server.Generate*` functions to testutil and support custom mnemonics in in-process testing network. Moved `TestMnemonic` from `testutil` package to `testdata`.
* (x/bank) [\#10771](https://github.com/cosmos/cosmos-sdk/pull/10771) Add safety check on bank module perms to allow module-specific mint restrictions (e.g. only minting a certain denom).
* (x/bank) [\#10771](https://github.com/cosmos/cosmos-sdk/pull/10771) Add `bank.BaseKeeper.WithMintCoinsRestriction` function to restrict use of bank `MintCoins` usage.
* [\#10868](https://github.com/cosmos/cosmos-sdk/pull/10868), [\#10989](https://github.com/cosmos/cosmos-sdk/pull/10989) The Gov keeper accepts now 2 more mandatory arguments, the ServiceMsgRouter and a maximum proposal metadata length.
* [\#10868](https://github.com/cosmos/cosmos-sdk/pull/10868), [\#10989](https://github.com/cosmos/cosmos-sdk/pull/10989), [\#11093](https://github.com/cosmos/cosmos-sdk/pull/11093) The Gov keeper accepts now 2 more mandatory arguments, the ServiceMsgRouter and a gov Config including the max metadata length.
* [\#11124](https://github.com/cosmos/cosmos-sdk/pull/11124) Add `GetAllVersions` to application store
* (x/authz) [\#10447](https://github.com/cosmos/cosmos-sdk/pull/10447) authz `NewGrant` takes a new argument: block time, to correctly validate expire time.
* [\#10961](https://github.com/cosmos/cosmos-sdk/pull/10961) Support third-party modules to add extension snapshots to state-sync.
* [\#11274](https://github.com/cosmos/cosmos-sdk/pull/11274) `types/errors.New` now is an alias for `types/errors.Register` and should only be used in initialization code.
* (authz)[\#11060](https://github.com/cosmos/cosmos-sdk/pull/11060) `authz.NewMsgGrant` `expiration` is now a pointer. When `nil` is used then no expiration will be set (grant won't expire).
* (x/distribution)[\#11457](https://github.com/cosmos/cosmos-sdk/pull/11457) Add amount field to `distr.MsgWithdrawDelegatorRewardResponse` and `distr.MsgWithdrawValidatorCommissionResponse`.
* [\#11334](https://github.com/cosmos/cosmos-sdk/pull/11334) Move `x/gov/types/v1beta2` to `x/gov/types/v1`.
* (x/auth/middleware) [#11413](https://github.com/cosmos/cosmos-sdk/pull/11413) Refactor tx middleware to be extensible on tx fee logic. Merged `MempoolFeeMiddleware` and `TxPriorityMiddleware` functionalities into `DeductFeeMiddleware`, make the logic extensible using the `TxFeeChecker` option, the current fee logic is preserved by the default `checkTxFeeWithValidatorMinGasPrices` implementation. Change `RejectExtensionOptionsMiddleware` to `NewExtensionOptionsMiddleware` which is extensible with the `ExtensionOptionChecker` option. Unpack the tx extension options `Any`s to interface `TxExtensionOptionI`.
* (migrations) [#11556](https://github.com/cosmos/cosmos-sdk/pull/11556#issuecomment-1091385011) Remove migration code from 0.42 and below. To use previous migrations, checkout previous versions of the cosmos-sdk.

### Client Breaking Changes

* [\#11797](https://github.com/cosmos/cosmos-sdk/pull/11797) Remove all RegisterRESTRoutes (previously deprecated)
* [\#11089](https://github.com/cosmos/cosmos-sdk/pull/11089]) interacting with the node through `grpc.Dial` requires clients to pass a codec refer to [doc](docs/run-node/interact-node.md).
* [\#9594](https://github.com/cosmos/cosmos-sdk/pull/9594) Remove legacy REST API. Please see the [REST Endpoints Migration guide](https://docs.cosmos.network/v0.45/migrations/rest.html) to migrate to the new REST endpoints.
* [\#9995](https://github.com/cosmos/cosmos-sdk/pull/9995) Increased gas cost for creating proposals.
* [\#11029](https://github.com/cosmos/cosmos-sdk/pull/11029) The deprecated Vote Option field is removed in gov v1beta2 and nil in v1beta1. Use Options instead.
* [\#11013](https://github.com/cosmos/cosmos-sdk/pull/11013) The `tx gov submit-proposal` command has changed syntax to support the new Msg-based gov proposals. To access the old CLI command, please use `tx gov submit-legacy-proposal`.
* [\#11170](https://github.com/cosmos/cosmos-sdk/issues/11170) Fixes issue related to grpc-gateway of supply by ibc-denom.

### CLI Breaking Changes

* (cli) [\#11818](https://github.com/cosmos/cosmos-sdk/pull/11818) CLI transactions preview now respect the chosen `--output` flag format (json or text).
* [\#9695](https://github.com/cosmos/cosmos-sdk/pull/9695) `<app> keys migrate` CLI command now takes no arguments.
* [\#9246](https://github.com/cosmos/cosmos-sdk/pull/9246) Removed the CLI flag `--setup-config-only` from the `testnet` command and added the subcommand `init-files`.
* [\#9780](https://github.com/cosmos/cosmos-sdk/pull/9780) Use sigs.k8s.io for yaml, which might lead to minor YAML output changes
* [\#10625](https://github.com/cosmos/cosmos-sdk/pull/10625) Rename `--fee-account` CLI flag to `--fee-granter`
* [\#10684](https://github.com/cosmos/cosmos-sdk/pull/10684) Rename `edit-validator` command's `--moniker` flag to `--new-moniker`
* (authz)[\#11060](https://github.com/cosmos/cosmos-sdk/pull/11060) Changed the default value of the `--expiration` `tx grant` CLI Flag: was now + 1year, update: null (no expire date).

### Improvements

* [\#12576](https://github.com/cosmos/cosmos-sdk/pull/12576) Remove dependency on cosmos/keyring and upgrade to 99designs/keyring v1.2.1
* [\#11693](https://github.com/cosmos/cosmos-sdk/pull/11693) Add validation for gentx cmd.
* (x/auth/vesting) [\#11652](https://github.com/cosmos/cosmos-sdk/pull/11652) Add util functions for `Period(s)`
* [\#11630](https://github.com/cosmos/cosmos-sdk/pull/11630) Add SafeSub method to sdk.Coin.
* [\#11511](https://github.com/cosmos/cosmos-sdk/pull/11511) Add api server flags to start command.
* [\#11484](https://github.com/cosmos/cosmos-sdk/pull/11484) Implement getter for keyring backend option.
* [\#11449](https://github.com/cosmos/cosmos-sdk/pull/11449) Improved error messages when node isn't synced.
* [\#11349](https://github.com/cosmos/cosmos-sdk/pull/11349) Add `RegisterAminoMsg` function that checks that a msg name is <40 chars (else this would break ledger nano signing) then registers the concrete msg type with amino, it should be used for registering `sdk.Msg`s with amino instead of `cdc.RegisterConcrete`.
* [\#11089](https://github.com/cosmos/cosmos-sdk/pull/11089]) Now cosmos-sdk consumers can upgrade gRPC to its newest versions.
* [\#10439](https://github.com/cosmos/cosmos-sdk/pull/10439) Check error for `RegisterQueryHandlerClient` in all modules `RegisterGRPCGatewayRoutes`.
* [\#9780](https://github.com/cosmos/cosmos-sdk/pull/9780) Remove gogoproto `moretags` YAML annotations and add `sigs.k8s.io/yaml` for YAML marshalling.
* (x/bank) [\#10134](https://github.com/cosmos/cosmos-sdk/pull/10134) Add `HasDenomMetadata` function to bank `Keeper` to check if a client coin denom metadata exists in state.
* (x/bank) [\#10022](https://github.com/cosmos/cosmos-sdk/pull/10022) `BankKeeper.SendCoins` now takes less execution time.
* (deps) [\#9987](https://github.com/cosmos/cosmos-sdk/pull/9987) Bump Go version minimum requirement to `1.17`
* (cli) [\#9856](https://github.com/cosmos/cosmos-sdk/pull/9856) Overwrite `--sequence` and `--account-number` flags with default flag values when used with `offline=false` in `sign-batch` command.
* (rosetta) [\#10001](https://github.com/cosmos/cosmos-sdk/issues/10001) Add documentation for rosetta-cli dockerfile and rename folder for the rosetta-ci dockerfile
* [\#9699](https://github.com/cosmos/cosmos-sdk/pull/9699) Add `:`, `.`, `-`, and `_` as allowed characters in the default denom regular expression.
* (genesis) [\#9697](https://github.com/cosmos/cosmos-sdk/pull/9697) Ensure `InitGenesis` returns with non-empty validator set.
* [\#10341](https://github.com/cosmos/cosmos-sdk/pull/10341) Move from `io/ioutil` to `io` and `os` packages.
* [\#10468](https://github.com/cosmos/cosmos-sdk/pull/10468) Allow futureOps to queue additional operations in simulations
* [\#10625](https://github.com/cosmos/cosmos-sdk/pull/10625) Add `--fee-payer` CLI flag
* (cli) [\#10683](https://github.com/cosmos/cosmos-sdk/pull/10683) In CLI, allow 1 SIGN_MODE_DIRECT signer in transactions with multiple signers.
* (deps) [\#10210](https://github.com/cosmos/cosmos-sdk/pull/10210) Bump Tendermint to [v0.35.0](https://github.com/tendermint/tendermint/releases/tag/v0.35.0).
* (deps) [\#10706](https://github.com/cosmos/cosmos-sdk/issues/10706) Bump rosetta-sdk-go to v0.7.2 and rosetta-cli to v0.7.3
* (types/errors) [\#10779](https://github.com/cosmos/cosmos-sdk/pull/10779) Move most functionality in `types/errors` to a standalone `errors` go module, except the `RootCodespace` errors and ABCI response helpers. All functions and types that used to live in `types/errors` are now aliased so this is not a breaking change.
* (gov) [\#10854](https://github.com/cosmos/cosmos-sdk/pull/10854) v1beta2's vote doesn't include the deprecate `option VoteOption` anymore. Instead, it only uses `WeightedVoteOption`.
* (types) [\#11004](https://github.com/cosmos/cosmos-sdk/pull/11004) Added mutable versions of many of the sdk.Dec types operations.  This improves performance when used by avoiding reallocating a new bigint for each operation.
* (x/auth) [\#10880](https://github.com/cosmos/cosmos-sdk/pull/10880) Added a new query to the tx query service that returns a block with transactions fully decoded.
* (types) [\#11200](https://github.com/cosmos/cosmos-sdk/pull/11200) Added `Min()` and `Max()` operations on sdk.Coins.
* (gov) [\#11287](https://github.com/cosmos/cosmos-sdk/pull/11287) Fix error message when no flags are provided while executing `submit-legacy-proposal` transaction.
* (x/auth) [\#11482](https://github.com/cosmos/cosmos-sdk/pull/11482) Improve panic message when attempting to register a method handler for a message that does not implement sdk.Msg
* (x/staking) [\#11596](https://github.com/cosmos/cosmos-sdk/pull/11596) Add (re)delegation getters
* (errors) [\#11960](https://github.com/cosmos/cosmos-sdk/pull/11960) Removed 'redacted' error message from defaultErrEncoder
* (ante) [#12013](https://github.com/cosmos/cosmos-sdk/pull/12013) Index ante events for failed tx.
* [#12668](https://github.com/cosmos/cosmos-sdk/pull/12668) Add `authz_msg_index` event attribute to message events emitted when executing via `MsgExec` through `x/authz`.
* [#12626](https://github.com/cosmos/cosmos-sdk/pull/12626) Upgrade IAVL to v0.19.0 with fast index and error propagation. NOTE: first start will take a while to propagate into new model.
* [#12649](https://github.com/cosmos/cosmos-sdk/pull/12649) Bump tendermint to v0.34.20.
* [#12576](https://github.com/cosmos/cosmos-sdk/pull/12576) Remove dependency on cosmos/keyring and upgrade to 99designs/keyring v1.2.1
* [#12589](https://github.com/cosmos/cosmos-sdk/pull/12589) Allow zero gas in simulation mode.
* [#12453](https://github.com/cosmos/cosmos-sdk/pull/12453) Add `NewInMemoryWithKeyring` function which allows the creation of in memory `keystore` instances with a specified set of existing items.
* [#11390](https://github.com/cosmos/cosmos-sdk/pull/11390) `LatestBlockResponse` & `BlockByHeightResponse` types' `Block` filed has been deprecated and they now contains new field `sdk_block` with `proposer_address` as `string`
* (deps) Downgrade to Tendermint [v0.34.20-rc0](https://github.com/tendermint/tendermint/releases/tag/v0.34.20-rc0).
* [#12089](https://github.com/cosmos/cosmos-sdk/pull/12089) Mark the `TipDecorator` as beta, don't include it in simapp by default.
* [#12153](https://github.com/cosmos/cosmos-sdk/pull/12153) Add a new `NewSimulationManagerFromAppModules` constructor, to simplify simulation wiring.

### Bug Fixes

* [#11969](https://github.com/cosmos/cosmos-sdk/pull/11969) Fix the panic error in `x/upgrade` when `AppVersion` is not set.
* (tests) [\#11940](https://github.com/cosmos/cosmos-sdk/pull/11940) Fix some client tests in the `x/gov` module
* [\#11772](https://github.com/cosmos/cosmos-sdk/pull/11772) Limit types.Dec length to avoid overflow.
* [\#11724](https://github.com/cosmos/cosmos-sdk/pull/11724) Fix data race issues with api.Server
* [\#11693](https://github.com/cosmos/cosmos-sdk/pull/11693) Add validation for gentx cmd.
* [\#11645](https://github.com/cosmos/cosmos-sdk/pull/11645) Fix `--home` flag ignored when running help.
* [\#11558](https://github.com/cosmos/cosmos-sdk/pull/11558) Fix `--dry-run` not working when using tx command.
* [\#11354](https://github.com/cosmos/cosmos-sdk/pull/11355) Added missing pagination flag for `bank q total` query.
* [\#11197](https://github.com/cosmos/cosmos-sdk/pull/11197) Signing with multisig now works with multisig address which is not in the keyring.
* (makefile) [\#11285](https://github.com/cosmos/cosmos-sdk/pull/11285) Fix lint-fix make target.
* (client) [\#11283](https://github.com/cosmos/cosmos-sdk/issues/11283) Support multiple keys for tx simulation and setting automatic gas for txs.
* (store) [\#11177](https://github.com/cosmos/cosmos-sdk/pull/11177) Update the prune `everything` strategy to store the last two heights.
* [\#10844](https://github.com/cosmos/cosmos-sdk/pull/10844) Automatic recovering non-consistent keyring storage during public key import.
* (store) [\#11117](https://github.com/cosmos/cosmos-sdk/pull/11117) Fix data race in store trace component
* (cli) [\#11065](https://github.com/cosmos/cosmos-sdk/pull/11065) Ensure the `tendermint-validator-set` query command respects the `-o` output flag.
* (grpc) [\#10985](https://github.com/cosmos/cosmos-sdk/pull/10992) The `/cosmos/tx/v1beta1/txs/{hash}` endpoint returns a 404 when a tx does not exist.
* (rosetta) [\#10340](https://github.com/cosmos/cosmos-sdk/pull/10340) Use `GenesisChunked(ctx)` instead `Genesis(ctx)` to get genesis block height
* [#10180](https://github.com/cosmos/cosmos-sdk/issues/10180) Documentation: make references to Cosmos SDK consistent
* [\#9651](https://github.com/cosmos/cosmos-sdk/pull/9651) Change inconsistent limit of `0` to `MaxUint64` on InfiniteGasMeter and add GasRemaining func to GasMeter.
* [\#9639](https://github.com/cosmos/cosmos-sdk/pull/9639) Check store keys length before accessing them by making sure that `key` is of length `m+1` (for `key[n:m]`)
* (types) [\#9627](https://github.com/cosmos/cosmos-sdk/pull/9627) Fix nil pointer panic on `NewBigIntFromInt`
* (x/genutil) [\#9574](https://github.com/cosmos/cosmos-sdk/pull/9575) Actually use the `gentx` client tx flags (like `--keyring-dir`)
* (x/distribution) [\#9599](https://github.com/cosmos/cosmos-sdk/pull/9599) Withdraw rewards event now includes a value attribute even if there are 0 rewards (due to situations like 100% commission).
* (x/genutil) [\#9638](https://github.com/cosmos/cosmos-sdk/pull/9638) Added missing validator key save when recovering from mnemonic
* [\#9762](https://github.com/cosmos/cosmos-sdk/pull/9762) The init command uses the chain-id from the client config if --chain-id is not provided
* [\#9854](https://github.com/cosmos/cosmos-sdk/pull/9854) Fixed the `make proto-gen` to get dynamic container name based on project name for the cosmos based sdks.
* [\#9980](https://github.com/cosmos/cosmos-sdk/pull/9980) Returning the error when the invalid argument is passed to bank query total supply cli.
* (server) [#10016](https://github.com/cosmos/cosmos-sdk/issues/10016) Fix marshaling of index-events into server config file.
* [\#10184](https://github.com/cosmos/cosmos-sdk/pull/10184) Fixed CLI tx commands to no longer explicitly require the chain-id flag as this value can come from a user config.
* [\#10239](https://github.com/cosmos/cosmos-sdk/pull/10239) Fixed x/bank/044 migrateDenomMetadata.
* (x/upgrade) [\#10189](https://github.com/cosmos/cosmos-sdk/issues/10189) Removed potential sources of non-determinism in upgrades
* [\#10258](https://github.com/cosmos/cosmos-sdk/issues/10258) Fixes issue related to segmentation fault on mac m1 arm64
* [\#10466](https://github.com/cosmos/cosmos-sdk/issues/10466) Fixes error with simulation tests when genesis start time is randomly created after the year 2262
* [\#10394](https://github.com/cosmos/cosmos-sdk/issues/10394) Fixes issue related to grpc-gateway of account balance by
  ibc-denom.
* [\#10593](https://github.com/cosmos/cosmos-sdk/pull/10593) Update swagger-ui to v4.1.0 to fix xss vulnerability.
* [\#10842](https://github.com/cosmos/cosmos-sdk/pull/10842) Fix error when `--generate-only`, `--max-msgs` fags set while executing `WithdrawAllRewards` command.
* [\#10897](https://github.com/cosmos/cosmos-sdk/pull/10897) Fix: set a non-zero value on gas overflow.
* [#9790](https://github.com/cosmos/cosmos-sdk/pull/10687) Fix behavior of `DecCoins.MulDecTruncate`.
* [\#10990](https://github.com/cosmos/cosmos-sdk/pull/10990) Fixes missing `iavl-cache-size` config parsing in `GetConfig` method.
* (crypto) [#11027] Remove dependency on Tendermint core for xsalsa20symmetric.
* (x/authz) [\#10447](https://github.com/cosmos/cosmos-sdk/pull/10447) Fix authz `NewGrant` expiration check.
* (x/authz) [\#10633](https://github.com/cosmos/cosmos-sdk/pull/10633) Fixed authorization not found error when executing message.
* [#11222](https://github.com/cosmos/cosmos-sdk/pull/11222) reject query with block height in the future
* [#11229](https://github.com/cosmos/cosmos-sdk/pull/11229) Handled the error message of `transaction encountered error` from tendermint.
* (x/authz) [\#11252](https://github.com/cosmos/cosmos-sdk/pull/11252) Allow insufficient funds error for authz simulation
* (cli) [\#11313](https://github.com/cosmos/cosmos-sdk/pull/11313) Fixes `--gas auto` when executing CLI transactions in `--generate-only` mode
* (cli) [\#11337](https://github.com/cosmos/cosmos-sdk/pull/11337) Fixes `show-adress` cli cmd
* (crypto) [\#11298](https://github.com/cosmos/cosmos-sdk/pull/11298) Fix cgo secp signature verification and update libscep256k1 library.
* (x/authz) [\#11512](https://github.com/cosmos/cosmos-sdk/pull/11512) Fix response of a panic to error, when subtracting balances.
* (rosetta) [\#11590](https://github.com/cosmos/cosmos-sdk/pull/11590) `/block` returns an error with nil pointer when a request has both of index and hash and increase timeout for huge genesis.
* (x/feegrant) [\#11813](https://github.com/cosmos/cosmos-sdk/pull/11813) Fix pagination total count in `AllowancesByGranter` query.
* (simapp) [\#11855](https://github.com/cosmos/cosmos-sdk/pull/11855) Use `sdkmath.Int` instead of `int64` for `SimulationState.InitialStake`.
* (x/capability) [\#11737](https://github.com/cosmos/cosmos-sdk/pull/11737) Use a fixed length encoding of `Capability` pointer for `FwdCapabilityKey`
* [\#11983](https://github.com/cosmos/cosmos-sdk/pull/11983) (x/feegrant, x/authz) rename grants query commands to `grants-by-grantee`, `grants-by-granter` cmds.
* (protos) [#12701](https://github.com/cosmos/cosmos-sdk/pull/12701) Fix tendermint and ics23 versions used in Makefile.  Run "make proto-gen".
* (testutil/sims) [#12374](https://github.com/cosmos/cosmos-sdk/pull/12374) fix the non-determinstic behavior in simulations caused by `GenSignedMockTx` and check empty coins slice before it is used to create `banktype.MsgSend`.
* [#12448](https://github.com/cosmos/cosmos-sdk/pull/12448) Start telemetry independently from the API server.
* [#12509](https://github.com/cosmos/cosmos-sdk/pull/12509) Fix `Register{Tx,Tendermint}Service` not being called, resulting in some endpoints like the Simulate endpoint not working.
* [#12416](https://github.com/cosmos/cosmos-sdk/pull/12416) Prevent zero gas transactions in the `DeductFeeDecorator` AnteHandler decorator.
* (x/mint) [#12384](https://github.com/cosmos/cosmos-sdk/pull/12384) Ensure `GoalBonded` must be positive when performing `x/mint` parameter validation.
* (x/auth) [#12261](https://github.com/cosmos/cosmos-sdk/pull/12261) Deprecate pagination in GetTxsEventRequest/Response in favor of page and limit to align with tendermint `SignClient.TxSearch`
* (vesting) [#12190](https://github.com/cosmos/cosmos-sdk/pull/12190) Replace https://github.com/cosmos/cosmos-sdk/pull/12190 to use `NewBaseAccountWithAddress` in all vesting account message handlers.
* (linting) [#12135](https://github.com/cosmos/cosmos-sdk/pull/12135/) Fix variable naming issues per enabled linters.  Run gofumpt to ensure easy reviews of ongoing linting work. 
* (linting) [#12132](https://github.com/cosmos/cosmos-sdk/pull/12132) Change sdk.Int to math.Int, run `gofumpt -w -l .`, and `golangci-lint run ./... --fix`
* (cli) [#12127](https://github.com/cosmos/cosmos-sdk/pull/12127) Fix the CLI not always taking into account `--fee-payer` and `--fee-granter` flags.
* (migrations) [#12028](https://github.com/cosmos/cosmos-sdk/pull/12028) Fix v0.45->v0.46 in-place store migrations.
* (baseapp) [#12089](https://github.com/cosmos/cosmos-sdk/pull/12089) Include antehandler and runMsgs events in SimulateTx.
* (cli) [#12095](https://github.com/cosmos/cosmos-sdk/pull/12095) Fix running a tx with --dry-run returns an error
* (x/auth) [#12108](https://github.com/cosmos/cosmos-sdk/pull/12108) Fix GetBlockWithTxs error when querying block with 0 tx
* (genutil) [#12140](https://github.com/cosmos/cosmos-sdk/pull/12140) Fix staking's genesis JSON migrate in the `simd migrate v0.46` CLI command.
* (types) [#12154](https://github.com/cosmos/cosmos-sdk/pull/12154) Add `baseAccountGetter` to avoid invalid account error when create vesting account.
* (x/crisis) [#12208](https://github.com/cosmos/cosmos-sdk/pull/12208) Fix progress index of crisis invariant assertion logs.
* (types) [#12229](https://github.com/cosmos/cosmos-sdk/pull/12229) Increase sdk.Dec maxApproxRootIterations to 300

### State Machine Breaking

* (baseapp) [\#11985](https://github.com/cosmos/cosmos-sdk/pull/11985) Add a `postHandler` to baseapp. This `postHandler` is like antehandler, but is run *after* the `runMsgs` execution. It is in the same store branch that `runMsgs`, meaning that both `runMsgs` and `postHandler`
* (x/gov) [#11998](https://github.com/cosmos/cosmos-sdk/pull/11998) Tweak the `x/gov` `ModuleAccountInvariant` invariant to ensure deposits are `<=` total module account balance instead of strictly equal.
* (x/upgrade) [\#11800](https://github.com/cosmos/cosmos-sdk/pull/11800) Fix `GetLastCompleteUpgrade` to properly return the latest upgrade.
* [\#10564](https://github.com/cosmos/cosmos-sdk/pull/10564) Fix bug when updating allowance inside AllowedMsgAllowance
* (x/auth)[\#9596](https://github.com/cosmos/cosmos-sdk/pull/9596) Enable creating periodic vesting accounts with a transactions instead of requiring them to be created in genesis.
* (x/bank) [\#9611](https://github.com/cosmos/cosmos-sdk/pull/9611) Introduce a new index to act as a reverse index between a denomination and address allowing to query for
  token holders of a specific denomination. `DenomOwners` is updated to use the new reverse index.
* (x/bank) [\#9832](https://github.com/cosmos/cosmos-sdk/pull/9832) Account balance is stored as `sdk.Int` rather than `sdk.Coin`.
* (x/bank) [\#9890](https://github.com/cosmos/cosmos-sdk/pull/9890) Remove duplicate denom from denom metadata key.
* (x/upgrade) [\#10189](https://github.com/cosmos/cosmos-sdk/issues/10189) Removed potential sources of non-determinism in upgrades
* [\#10422](https://github.com/cosmos/cosmos-sdk/pull/10422) and [\#10529](https://github.com/cosmos/cosmos-sdk/pull/10529) Add `MinCommissionRate` param to `x/staking` module.
* (x/gov) [#10763](https://github.com/cosmos/cosmos-sdk/pull/10763) modify the fields in `TallyParams` to use `string` instead of `bytes`
* [#10770](https://github.com/cosmos/cosmos-sdk/pull/10770) revert tx when block gas limit exceeded
* (x/gov) [\#10868](https://github.com/cosmos/cosmos-sdk/pull/10868) Bump gov to v1beta2. Both v1beta1 and v1beta2 queries and Msgs are accepted.
* [\#11011](https://github.com/cosmos/cosmos-sdk/pull/11011) Remove burning of deposits when qourum is not reached on a governance proposal and when the deposit is not fully met.
* [\#11019](https://github.com/cosmos/cosmos-sdk/pull/11019) Add `MsgCreatePermanentLockedAccount` and CLI method for creating permanent locked account
* (x/staking) [\#10885] (https://github.com/cosmos/cosmos-sdk/pull/10885) Add new `CancelUnbondingDelegation`
  transaction to `x/staking` module. Delegators can now cancel unbonding delegation entry and delegate back to validator.
* (x/feegrant) [\#10830](https://github.com/cosmos/cosmos-sdk/pull/10830) Expired allowances will be pruned from state.
* (x/authz,x/feegrant) [\#11214](https://github.com/cosmos/cosmos-sdk/pull/11214) Fix Amino JSON encoding of authz and feegrant Msgs to be consistent with other modules.
* (authz)[\#11060](https://github.com/cosmos/cosmos-sdk/pull/11060) Support grant with no expire time.
* (x/gov) [\#10868](https://github.com/cosmos/cosmos-sdk/pull/10868) Bump gov to v1. 

### Deprecated

* (x/upgrade) [#9906](https://github.com/cosmos/cosmos-sdk/pull/9906) Deprecate `UpgradeConsensusState` gRPC query since this functionality is only used for IBC, which now has its own [IBC replacement](https://github.com/cosmos/ibc-go/blob/2c880a22e9f9cc75f62b527ca94aa75ce1106001/proto/ibc/core/client/v1/query.proto#L54)
* (types) [#10948](https://github.com/cosmos/cosmos-sdk/issues/10948) Deprecate the types.DBBackend variable and types.NewLevelDB function. They are replaced by a new entry in `app.toml`: `app-db-backend` and `tendermint/tm-db`s `NewDB` function. If `app-db-backend` is defined, then it is used. Otherwise, if `types.DBBackend` is defined, it is used (until removed: [#11241](https://github.com/cosmos/cosmos-sdk/issues/11241)). Otherwise, Tendermint config's `db-backend` is used.

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/cosmos-sdk/blob/main/CHANGELOG.md#v0460---2022-07-26).
