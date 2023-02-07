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

### Features

* [#14897](https://github.com/cosmos/cosmos-sdk/pull/14897) Migrate the Cosmos SDK to CometBFT.
* (x/gov) [#14720](https://github.com/cosmos/cosmos-sdk/pull/14720) Upstream expedited proposals from Osmosis.
* (x/auth) [#14650](https://github.com/cosmos/cosmos-sdk/pull/14650) Add Textual SignModeHandler. It is however **NOT** enabled by default, and should only be used for **TESTING** purposes until `SIGN_MODE_TEXTUAL` is fully released.
* (cli) [#14655](https://github.com/cosmos/cosmos-sdk/pull/14655) Add a new command to list supported algos.
* (x/crisis) [#14588](https://github.com/cosmos/cosmos-sdk/pull/14588) Use CacheContext() in AssertInvariants()
* (client) [#14342](https://github.com/cosmos/cosmos-sdk/pull/14342) Add `simd config` command is now a sub-command, for setting, getting and migrating Cosmos SDK configuration files.
* (query) [#14468](https://github.com/cosmos/cosmos-sdk/pull/14468) Implement pagination for collections.
* (x/bank) [#14045](https://github.com/cosmos/cosmos-sdk/pull/14045) Add CLI command `spendable-balances`, which also accepts the flag `--denom`.
* (x/slashing, x/staking) [#14363](https://github.com/cosmos/cosmos-sdk/pull/14363) Add the infraction a validator commited type as an argument to a `SlashWithInfractionReason` keeper method.
* (client) [#13867](https://github.com/cosmos/cosmos-sdk/pull/13867/) Wire AutoCLI commands with SimApp.
* (x/distribution) [#14322](https://github.com/cosmos/cosmos-sdk/pull/14322) Introduce a new gRPC message handler, `DepositValidatorRewardsPool`, that allows explicit funding of a validator's reward pool.
* (x/evidence) [#13740](https://github.com/cosmos/cosmos-sdk/pull/13740) Add new proto field `hash` of type `string` to `QueryEvidenceRequest` which helps to decode the hash properly while using query API.
* (core) [#13306](https://github.com/cosmos/cosmos-sdk/pull/13306) Add a `FormatCoins` function to in `core/coins` to format sdk Coins following the Value Renderers spec.
* (math) [#13306](https://github.com/cosmos/cosmos-sdk/pull/13306) Add `FormatInt` and `FormatDec` functions in `math` to format integers and decimals following the Value Renderers spec.
* (x/staking) [#13122](https://github.com/cosmos/cosmos-sdk/pull/13122) Add `UnbondingCanComplete` and `PutUnbondingOnHold` to `x/staking` module.
* [#13437](https://github.com/cosmos/cosmos-sdk/pull/13437) Add new flag `--modules-to-export` in `simd export` command to export only selected modules.
* [#13298](https://github.com/cosmos/cosmos-sdk/pull/13298) Add `AddGenesisAccount` helper func in x/auth module which helps adding accounts to genesis state.
* (x/authz) [#12648](https://github.com/cosmos/cosmos-sdk/pull/12648) Add an allow list, an optional list of addresses allowed to receive bank assets via authz MsgSend grant.
* (sdk.Coins) [#12627](https://github.com/cosmos/cosmos-sdk/pull/12627) Make a `Denoms` method on sdk.Coins.
* (testutil) [#12973](https://github.com/cosmos/cosmos-sdk/pull/12973) Add generic `testutil.RandSliceElem` function which selects a random element from the list.
* (client) [#12936](https://github.com/cosmos/cosmos-sdk/pull/12936) Add capability to pre-process transactions before broadcasting from a higher level chain.
* (cli) [#13064](https://github.com/cosmos/cosmos-sdk/pull/13064) Add `debug prefixes` to list supported HRP prefixes via .
* (ledger) [#12935](https://github.com/cosmos/cosmos-sdk/pull/12935) Generalize Ledger integration to allow for different apps or key types that use SECP256k1.
* (x/bank) [#11981](https://github.com/cosmos/cosmos-sdk/pull/11981) Create the `SetSendEnabled` endpoint for managing the bank's SendEnabled settings.
* (x/auth) [#13210](https://github.com/cosmos/cosmos-sdk/pull/13210) Add `Query/AccountInfo` endpoint for simplified access to basic account info.
* (x/consensus) [#12905](https://github.com/cosmos/cosmos-sdk/pull/12905) Create a new `x/consensus` module that is now responsible for maintaining Tendermint consensus parameters instead of `x/param`. Legacy types remain in order to facilitate parameter migration from the deprecated `x/params`. App developers should ensure that they execute `baseapp.MigrateParams` during their chain upgrade. These legacy types will be removed in a future release.
* (x/gov) [#13010](https://github.com/cosmos/cosmos-sdk/pull/13010) Add `cancel-proposal` feature to proposals. Now proposers can cancel the proposal prior to the proposal's voting period end time.
* (client/tx) [#13670](https://github.com/cosmos/cosmos-sdk/pull/13670) Add validation in `BuildUnsignedTx` to prevent simple inclusion of valid mnemonics
* [#13473](https://github.com/cosmos/cosmos-sdk/pull/13473) ADR-038: Go plugin system proposal
* [#14356](https://github.com/cosmos/cosmos-sdk/pull/14356) Add `events.GetAttributes` and `event.GetAttribute` methods to simplify the retrieval of an attribute from event(s).
* [#14472](https://github.com/cosmos/cosmos-sdk/pull/14356) The recommended metadata format for x/gov and x/group proposals now uses an array of strings (instead of a single string) for the `authors` field.
* (mempool) [#14484](https://github.com/cosmos/cosmos-sdk/pull/14484) Add priority nonce mempool option for transaction replacement.
* (client) [#14509](https://github.com/cosmos/cosmos-sdk/pull/#14509) Added `AddKeyringFlags` function.
* (x/gov,cli) [#14718](https://github.com/cosmos/cosmos-sdk/pull/14718) Added `AddGovPropFlagsToCmd` and `ReadGovPropFlags` functions.
* (store) [#14189](https://github.com/cosmos/cosmos-sdk/pull/14189) Add config `iavl-lazy-loading` to enable lazy loading of iavl store, to improve start up time of archive nodes, add method `SetLazyLoading` to `CommitMultiStore` interface.
* (x/bank) [#14894](https://github.com/cosmos/cosmos-sdk/pull/14894) Return a human readable denomination for IBC vouchers when querying bank balances. Added a `ResolveDenom` parameter to `types.QueryAllBalancesRequest` and `--resolve-denom` flag to `GetBalancesCmd()`.
* (x/groups) [#14879](https://github.com/cosmos/cosmos-sdk/pull/14879) Add `Query/Groups` query to get all the groups.

### Improvements

* (store) [#14410](https://github.com/cosmos/cosmos-sdk/pull/14410) `rootmulti.Store.loadVersion` has validation to check if all the module stores' height is correct, it will error if any module store has incorrect height.
* (x/evidence) [#14757](https://github.com/cosmos/cosmos-sdk/pull/14757) Evidence messages do not need to implement a `.Type()` anymore.
* (x/auth/tx) [#14751](https://github.com/cosmos/cosmos-sdk/pull/14751) Remove `.Type()` and `Route()` methods from all msgs and `legacytx.LegacyMsg` interface.
* [#14691](https://github.com/cosmos/cosmos-sdk/pull/14691) Change behavior of `sdk.StringifyEvents` to not flatten events attributes by events type.
    * This change only affects ABCI message logs, and not the actual events.
* [#14692](https://github.com/cosmos/cosmos-sdk/pull/14692) Improve RPC queries error message when app is at height 0.
* [#14609](https://github.com/cosmos/cosmos-sdk/pull/14609) Add RetryForBlocks method to use in tests that require waiting for a transaction to be included in a block.
* [#14017](https://github.com/cosmos/cosmos-sdk/pull/14017) Simplify ADR-028 and `address.Module`.
    * This updates the [ADR-028](https://docs.cosmos.network/main/architecture/adr-028-public-key-addresses) and enhance the `address.Module` API to support module addresses and sub-module addresses in a backward compatible way.
* [#14529](https://github.com/cosmos/cosmos-sdk/pull/14529) Add new property `BondDenom` to `SimulationState` struct.
* (x/group, x/gov) [#14483](https://github.com/cosmos/cosmos-sdk/pull/14483) Add support for `[]string` and `[]int` in `draft-proposal` prompt.
* (protobuf) [#14476](https://github.com/cosmos/cosmos-sdk/pull/14476) Clean up protobuf annotations `{accepts,implements}_interface`.
* (module) [#14415](https://github.com/cosmos/cosmos-sdk/pull/14415) Loosen assertions in SetOrderBeginBlockers() and SetOrderEndBlockers()
* (context)[#14384](https://github.com/cosmos/cosmos-sdk/pull/14384) refactor(context): Pass EventManager to the context as an interface.
* (types) [#14354](https://github.com/cosmos/cosmos-sdk/pull/14354) - improve performance on Context.KVStore and Context.TransientStore by 40%
* (crypto/keyring) [#14151](https://github.com/cosmos/cosmos-sdk/pull/14151) Move keys presentation from `crypto/keyring` to `client/keys`
* (types) [#14163](https://github.com/cosmos/cosmos-sdk/pull/14163) Refactor `(coins Coins) Validate()` to avoid unnecessary map.
* (signing) [#14087](https://github.com/cosmos/cosmos-sdk/pull/14087) Add SignModeHandlerWithContext interface with a new `GetSignBytesWithContext` to get the sign bytes using `context.Context` as an argument to access state.
* (server) [#14062](https://github.com/cosmos/cosmos-sdk/pull/14062) Remove rosetta from server start.
* [13882] (https://github.com/cosmos/cosmos-sdk/pull/13882) Add tx `encode` and `decode` endpoints to amino tx service.
  > Note: These endpoints encodes and decodes only amino txs.
* (x/nft) [#13836](https://github.com/cosmos/cosmos-sdk/pull/13836) Remove the validation for `classID` and `nftID` from the NFT module.
* [#13789](https://github.com/cosmos/cosmos-sdk/pull/13789) Add tx `encode` and `decode` endpoints to tx service.
  > Note: These endpoints will only encode and decode proto messages, Amino encoding and decoding is not supported.
* [#13826](https://github.com/cosmos/cosmos-sdk/pull/13826) Support custom `GasConfig` configuration for applications.
* [#13619](https://github.com/cosmos/cosmos-sdk/pull/13619) Add new function called LogDeferred to report errors in defers. Use the function in x/bank files.
* (tools) [#13603](https://github.com/cosmos/cosmos-sdk/pull/13603) Rename cosmovisor package name to `cosmossdk.io/tools/cosmovisor`. The new tool directory contains Cosmos SDK tools.
* (deps) [#13397](https://github.com/cosmos/cosmos-sdk/pull/13397) Bump Go version minimum requirement to `1.19`.
* [#13070](https://github.com/cosmos/cosmos-sdk/pull/13070) Migrate from `gogo/protobuf` to `cosmos/gogoproto`.
* [#12995](https://github.com/cosmos/cosmos-sdk/pull/12995) Add `FormatTime` and `ParseTimeString` methods.
* [#12952](https://github.com/cosmos/cosmos-sdk/pull/12952) Replace keyring module to Cosmos fork.
* [#12352](https://github.com/cosmos/cosmos-sdk/pull/12352) Move the `RegisterSwaggerAPI` logic into a separate helper function in the server package.
* [#12876](https://github.com/cosmos/cosmos-sdk/pull/12876) Remove proposer-based rewards.
* [#12892](https://github.com/cosmos/cosmos-sdk/pull/12892) `make format` now runs only gofumpt and golangci-lint run ./... --fix, replacing `goimports` `gofmt` and `misspell`
* [#12846](https://github.com/cosmos/cosmos-sdk/pull/12846) Remove `RandomizedParams` from the `AppModuleSimulation` interface which is no longer needed.
* (ci) [#12854](https://github.com/cosmos/cosmos-sdk/pull/12854) Use ghcr.io to host the proto builder image. Update proto builder image to go 1.19
* (x/bank) [#12706](https://github.com/cosmos/cosmos-sdk/pull/12706) Added the `chain-id` flag to the `AddTxFlagsToCmd` API. There is no longer a need to explicitly register this flag on commands whens `AddTxFlagsToCmd` is already called.
* [#12791](https://github.com/cosmos/cosmos-sdk/pull/12791) Bump the math library used in the sdk and replace old usages of sdk.\*
* [#12717](https://github.com/cosmos/cosmos-sdk/pull/12717) Use injected encoding params in simapp.
* [#12702](https://github.com/cosmos/cosmos-sdk/pull/12702) Linting and tidiness, fixed two minor security warnings.
* [#12634](https://github.com/cosmos/cosmos-sdk/pull/12634) Move `sdk.Dec` to math package.
* [#12596](https://github.com/cosmos/cosmos-sdk/pull/12596) Remove all imports of the non-existent gogo/protobuf v1.3.3 to ease downstream use and go workspaces.
* [#12187](https://github.com/cosmos/cosmos-sdk/pull/12187) Add batch operation for x/nft module.
* [#12455](https://github.com/cosmos/cosmos-sdk/pull/12455) Show attempts count in error for signing.
* [#13101](https://github.com/cosmos/cosmos-sdk/pull/13101) Remove weights from `simapp/params` and `testutil/sims`. They are now in their respective modules.
* [#12398](https://github.com/cosmos/cosmos-sdk/issues/12398) Refactor all `x` modules to unit-test via mocks and decouple `simapp`.
* [#13144](https://github.com/cosmos/cosmos-sdk/pull/13144) Add validator distribution info grpc gateway get endpoint.
* [#13168](https://github.com/cosmos/cosmos-sdk/pull/13168) Migrate tendermintdev/proto-builder to ghcr.io. New image `ghcr.io/cosmos/proto-builder:0.8`
* [#13178](https://github.com/cosmos/cosmos-sdk/pull/13178) Add `cosmos.msg.v1.service` protobuf annotation to allow tooling to distinguish between Msg and Query services via reflection.
* [#13236](https://github.com/cosmos/cosmos-sdk/pull/13236) Integrate Filter Logging
* [#13528](https://github.com/cosmos/cosmos-sdk/pull/13528) Update `ValidateMemoDecorator` to only check memo against `MaxMemoCharacters` param when a memo is present.
* [#13781](https://github.com/cosmos/cosmos-sdk/pull/13781) Remove `client/keys.KeysCdc`.
* [#13802](https://github.com/cosmos/cosmos-sdk/pull/13802) Add --output-document flag to the export CLI command to allow writing genesis state to a file.
* [#13794](https://github.com/cosmos/cosmos-sdk/pull/13794) `types/module.Manager` now supports the
`cosmossdk.io/core/appmodule.AppModule` API via the new `NewManagerFromMap` constructor.
* [#14019](https://github.com/cosmos/cosmos-sdk/issues/14019) Remove the interface casting to allow other implementations of a `CommitMultiStore`.
* (x/gov) [#14390](https://github.com/cosmos/cosmos-sdk/pull/14390) Add title, proposer and summary to proposal struct.
* (baseapp) [#14417](https://github.com/cosmos/cosmos-sdk/pull/14417) `SetStreamingService` accepts appOptions, AppCodec and Storekeys needed to set streamers.  
    * Store pacakge no longer has a dependency on baseapp. 
* (store) [#14438](https://github.com/cosmos/cosmos-sdk/pull/14438)  Pass logger from baseapp to store. 
* (store) [#14439](https://github.com/cosmos/cosmos-sdk/pull/14439) Remove global metric gatherer from store. 
    * By default store has a no op metric gatherer, the application developer must set another metric gatherer or us the provided one in `store/metrics`.
* [#14406](https://github.com/cosmos/cosmos-sdk/issues/14406) Migrate usage of types/store.go to store/types/..
* (x/staking) [#14590](https://github.com/cosmos/cosmos-sdk/pull/14590) Return undelegate amount in MsgUndelegateResponse
* (tools) [#14793](https://github.com/cosmos/cosmos-sdk/pull/14793) Dockerfile optimization.

### State Machine Breaking

* (baseapp, x/auth/posthandler) [#13940](https://github.com/cosmos/cosmos-sdk/pull/13940) Update `PostHandler` to receive the `runTx` success boolean.
* (x/group) [#13742](https://github.com/cosmos/cosmos-sdk/pull/13742) Migrate group policy account from module accounts to base account.
* (codec) [#13307](https://github.com/cosmos/cosmos-sdk/pull/13307) Register all modules' `Msg`s with group's ModuleCdc so that Amino sign bytes are correctly generated.
* (codec) [#13196](https://github.com/cosmos/cosmos-sdk/pull/13196) Register all modules' `Msg`s with gov's ModuleCdc so that Amino sign bytes are correctly generated.
* (group) [#13592](https://github.com/cosmos/cosmos-sdk/pull/13592) Fix group types registration with Amino.
* (x/distribution) [#12852](https://github.com/cosmos/cosmos-sdk/pull/12852) Deprecate `CommunityPoolSpendProposal`. Please execute a `MsgCommunityPoolSpend` message via the new v1 `x/gov` module instead. This message can be used to directly fund the `x/gov` module account.
* (x/bank) [#12610](https://github.com/cosmos/cosmos-sdk/pull/12610) `MsgMultiSend` now allows only a single input.
* (x/bank) [#12630](https://github.com/cosmos/cosmos-sdk/pull/12630) Migrate `x/bank` to self-managed parameters and deprecate its usage of `x/params`.
* (x/auth) [#12475](https://github.com/cosmos/cosmos-sdk/pull/12475) Migrate `x/auth` to self-managed parameters and deprecate its usage of `x/params`.
* (x/slashing) [#12399](https://github.com/cosmos/cosmos-sdk/pull/12399) Migrate `x/slashing` to self-managed parameters and deprecate its usage of `x/params`.
* (x/mint) [#12363](https://github.com/cosmos/cosmos-sdk/pull/12363) Migrate `x/mint` to self-managed parameters and deprecate it's usage of `x/params`.
* (x/distribution) [#12434](https://github.com/cosmos/cosmos-sdk/pull/12434) Migrate `x/distribution` to self-managed parameters and deprecate it's usage of `x/params`.
* (x/crisis) [#12445](https://github.com/cosmos/cosmos-sdk/pull/12445) Migrate `x/crisis` to self-managed parameters and deprecate it's usage of `x/params`.
* (x/gov) [#12631](https://github.com/cosmos/cosmos-sdk/pull/12631) Migrate `x/gov` to self-managed parameters and deprecate it's usage of `x/params`.
* (x/staking) [#12409](https://github.com/cosmos/cosmos-sdk/pull/12409) Migrate `x/staking` to self-managed parameters and deprecate it's usage of `x/params`.
* (x/bank) [#11859](https://github.com/cosmos/cosmos-sdk/pull/11859) Move the SendEnabled information out of the Params and into the state store directly.
* (x/gov) [#12771](https://github.com/cosmos/cosmos-sdk/pull/12771) Initial deposit requirement for proposals at submission time.
* (x/staking) [#12967](https://github.com/cosmos/cosmos-sdk/pull/12967) `unbond` now creates only one unbonding delegation entry when multiple unbondings exist at a single height (e.g. through multiple messages in a transaction).
* (x/auth/vesting) [#13502](https://github.com/cosmos/cosmos-sdk/pull/13502) Add Amino Msg registration for `MsgCreatePeriodicVestingAccount`.
* (x/auth)[#13780](https://github.com/cosmos/cosmos-sdk/pull/13780) `id` (type of int64) in `AccountAddressByID` grpc query is now deprecated, update to account-id(type of uint64) to use `AccountAddressByID`.
* (x/group) [#13876](https://github.com/cosmos/cosmos-sdk/pull/13876) Fix group MinExecutionPeriod that is checked on execution now, instead of voting period end.
* (x/feegrant) [#14294](https://github.com/cosmos/cosmos-sdk/pull/14294) Moved the logic of rejecting duplicate grant from `msg_server` to `keeper` method.
* (store) [#14378](https://github.com/cosmos/cosmos-sdk/pull/14378) The `CacheKV` store is thread-safe again, which includes improved iteration and deletion logic. Iteration is on a strictly isolated view now, which is breaking from previous behavior.
* (x/staking) [#14590](https://github.com/cosmos/cosmos-sdk/pull/14590) `MsgUndelegateResponse` now includes undelegated amount. `x/staking` module's `keeper.Undelegate` now returns 3 values (completionTime,undelegateAmount,error)  instead of 2.

### API Breaking Changes

* (x/gov) [#14720](https://github.com/cosmos/cosmos-sdk/pull/14720) Add an expedited field in the gov v1 proposal and `MsgNewMsgProposal`.
* [#14847](https://github.com/cosmos/cosmos-sdk/pull/14847) App and ModuleManager methods `InitGenesis`, `ExportGenesis`, `BeginBlock` and `EndBlock` now also return an error.
* (simulation) [#14728](https://github.com/cosmos/cosmos-sdk/pull/14728) Rename the `ParamChanges` field to `LegacyParamChange` and `Contents` to `LegacyProposalContents` in `simulation.SimulationState`. Additionally it adds a `ProposalMsgs` field to `simulation.SimulationState`.
* (x/upgrade) [#14764](https://github.com/cosmos/cosmos-sdk/pull/14764) The `x/upgrade` module is extracted to have a separate go.mod file which allows it to be a standalone module. 
* (x/gov) [#14782](https://github.com/cosmos/cosmos-sdk/pull/14782) Move the `metadata` argument in `govv1.NewProposal` alongside `title` and `summary`.
* (store) [#14746](https://github.com/cosmos/cosmos-sdk/pull/14746) Extract Store in its own go.mod and rename the package to `cosmossdk.io/store`.
* (simulation) [#14751](https://github.com/cosmos/cosmos-sdk/pull/14751) Remove the `MsgType` field from `simulation.OperationInput` struct.
* (crypto/keyring) [#13734](https://github.com/cosmos/cosmos-sdk/pull/13834) The keyring's `Sign` method now takes a new `signMode` argument. It is only used if the signing key is a Ledger hardware device. You can set it to 0 in all other cases.
* (x/evidence) [14724](https://github.com/cosmos/cosmos-sdk/pull/14724) Extract Evidence in its own go.mod and rename the package to `cosmossdk.io/x/evidence`.
* (x/nft) [#14725](https://github.com/cosmos/cosmos-sdk/pull/14725) Extract NFT in its own go.mod and rename the package to `cosmossdk.io/x/nft`.
* (tx) [#14634](https://github.com/cosmos/cosmos-sdk/pull/14634) Move the `tx` go module to `x/tx`.
* (snapshots) [#14597](https://github.com/cosmos/cosmos-sdk/pull/14597) Move `snapshots` to `store/snapshots`, rename and bump proto package to v1.
* (crypto/keyring) [#14151](https://github.com/cosmos/cosmos-sdk/pull/14151) Move keys presentation from `crypto/keyring` to `client/keys`
* (modules) [#13850](https://github.com/cosmos/cosmos-sdk/pull/13850) and [#14046](https://github.com/cosmos/cosmos-sdk/pull/14046) Remove gogoproto stringer annotations. This removes the custom `String()` methods on all types that were using the annotations.
* (x/auth) [#13850](https://github.com/cosmos/cosmos-sdk/pull/13850/) Remove `MarshalYAML` methods from module (`x/...`) types.
* (x/auth) [#13877](https://github.com/cosmos/cosmos-sdk/pull/13877) Rename `AccountKeeper`'s `GetNextAccountNumber` to `NextAccountNumber`.
* (x/evidence) [#13740](https://github.com/cosmos/cosmos-sdk/pull/13740) The `NewQueryEvidenceRequest` function now takes `hash` as a HEX encoded `string`.
* (server) [#13485](https://github.com/cosmos/cosmos-sdk/pull/13485) The `Application` service now requires the `RegisterNodeService` method to be implemented.
* [#13437](https://github.com/cosmos/cosmos-sdk/pull/13437) Add a list of modules to export argument in `ExportAppStateAndValidators`.
* (x/slashing) [#13427](https://github.com/cosmos/cosmos-sdk/pull/13427) Move `x/slashing/testslashing` to `x/slashing/testutil` for consistency with other modules.
* (x/staking) [#13427](https://github.com/cosmos/cosmos-sdk/pull/13427) Move `x/staking/teststaking` to `x/staking/testutil` for consistency with other modules.
* (simapp) [#13402](https://github.com/cosmos/cosmos-sdk/pull/13402) Move simulation flags to `x/simulation/client/cli`.
* (simapp) [#13402](https://github.com/cosmos/cosmos-sdk/pull/13402) Move simulation helpers functions (`SetupSimulation`, `SimulationOperations`, `CheckExportSimulation`, `PrintStats`, `GetSimulationLog`) to `testutil/sims`.
* (simapp) [#13402](https://github.com/cosmos/cosmos-sdk/pull/13402) Move `testutil/rest` package to `testutil`.
* (types) [#13380](https://github.com/cosmos/cosmos-sdk/pull/13380) Remove deprecated `sdk.NewLevelDB`.
* (simapp) [#13378](https://github.com/cosmos/cosmos-sdk/pull/13378) Move `simapp.App` to `runtime.AppI`.
* (tx) [#12659](https://github.com/cosmos/cosmos-sdk/pull/12659) Remove broadcast mode `block`.
* (db) [#13370](https://github.com/cosmos/cosmos-sdk/pull/13370) Remove storev2alpha1, see also https://github.com/cosmos/cosmos-sdk/pull/13371
* (x/bank) [#12706](https://github.com/cosmos/cosmos-sdk/pull/12706) Removed the `testutil` package from the `x/bank/client` package.
* (simapp) [#12747](https://github.com/cosmos/cosmos-sdk/pull/12747) Remove `simapp.MakeTestEncodingConfig`. Please use `moduletestutil.MakeTestEncodingConfig` (`types/module/testutil`) in tests instead.
* (x/bank) [#12648](https://github.com/cosmos/cosmos-sdk/pull/12648) `NewSendAuthorization` takes a new argument of an optional list of addresses allowed to receive bank assests via authz MsgSend grant. You can pass `nil` for the same behavior as before, i.e. any recipient is allowed.
* (x/bank) [#12593](https://github.com/cosmos/cosmos-sdk/pull/12593) Add `SpendableCoin` method to `BaseViewKeeper`
* (x/slashing) [#12581](https://github.com/cosmos/cosmos-sdk/pull/12581) Remove `x/slashing` legacy querier.
* (types) [#12355](https://github.com/cosmos/cosmos-sdk/pull/12355) Remove the compile-time `types.DBbackend` variable. Removes usage of the same in server/util.go
* (x/gov) [#12368](https://github.com/cosmos/cosmos-sdk/pull/12369) Gov keeper is now passed by reference instead of copy to make post-construction mutation of Hooks and Proposal Handlers possible at a framework level.
* (simapp) [#12270](https://github.com/cosmos/cosmos-sdk/pull/12270) Remove `invCheckPeriod uint` attribute from `SimApp` struct as per migration of `x/crisis` to app wiring
* (simapp) [#12334](https://github.com/cosmos/cosmos-sdk/pull/12334) Move `simapp.ConvertAddrsToValAddrs` and `simapp.CreateTestPubKeys ` to respectively `simtestutil.ConvertAddrsToValAddrs` and `simtestutil.CreateTestPubKeys` (`testutil/sims`)
* (simapp) [#12312](https://github.com/cosmos/cosmos-sdk/pull/12312) Move `simapp.EmptyAppOptions` to `simtestutil.EmptyAppOptions` (`testutil/sims`)
* (simapp) [#12312](https://github.com/cosmos/cosmos-sdk/pull/12312) Remove `skipUpgradeHeights map[int64]bool` and `homePath string` from `NewSimApp` constructor as per migration of `x/upgrade` to app-wiring.
* (testutil) [#12278](https://github.com/cosmos/cosmos-sdk/pull/12278) Move all functions from `simapp/helpers` to `testutil/sims`
* (testutil) [#12233](https://github.com/cosmos/cosmos-sdk/pull/12233) Move `simapp.TestAddr` to `simtestutil.TestAddr` (`testutil/sims`)
* (x/staking) [#12102](https://github.com/cosmos/cosmos-sdk/pull/12102) Staking keeper now is passed by reference instead of copy. Keeper's SetHooks no longer returns keeper. It updates the keeper in place instead.
* (linting) [#12141](https://github.com/cosmos/cosmos-sdk/pull/12141) Fix usability related linting for database. This means removing the infix Prefix from `prefix.NewPrefixWriter` and such so that it is `prefix.NewWriter` and making `db.DBConnection` and such into `db.Connection`
* (x/distribution) [#12434](https://github.com/cosmos/cosmos-sdk/pull/12434) `x/distribution` module `SetParams` keeper method definition is now updated to return `error`.
* (x/staking) [#12409](https://github.com/cosmos/cosmos-sdk/pull/12409) `x/staking` module `SetParams` keeper method definition is now updated to return `error`.
* (x/crisis) [#12445](https://github.com/cosmos/cosmos-sdk/pull/12445) `x/crisis` module `SetConstantFee` keeper method definition is now updated to return `error`.
* (x/gov) [#12631](https://github.com/cosmos/cosmos-sdk/pull/12631) `x/gov` module refactored to use `Params` as single struct instead of `DepositParams`, `TallyParams` & `VotingParams`.
* (x/gov) [#12631](https://github.com/cosmos/cosmos-sdk/pull/12631) Migrate `x/gov` to self-managed parameters and deprecate it's usage of `x/params`.
* (x/bank) [#12630](https://github.com/cosmos/cosmos-sdk/pull/12630) `x/bank` module `SetParams` keeper method definition is now updated to return `error`.
* (x/bank) [#11859](https://github.com/cosmos/cosmos-sdk/pull/11859) Move the SendEnabled information out of the Params and into the state store directly.
  The information can now be accessed using the BankKeeper.
  Setting can be done using MsgSetSendEnabled as a governance proposal.
  A SendEnabled query has been added to both GRPC and CLI.
* (appModule) Remove `Route`, `QuerierRoute` and `LegacyQuerierHandler` from AppModule Interface.
* (x/modules) Remove all LegacyQueries and related code from modules
* (store) [#11825](https://github.com/cosmos/cosmos-sdk/pull/11825) Make extension snapshotter interface safer to use, renamed the util function `WriteExtensionItem` to `WriteExtensionPayload`.
* (x/genutil)[#12956](https://github.com/cosmos/cosmos-sdk/pull/12956) `genutil.AppModuleBasic` has a new attribute: genesis transaction validation function. The existing validation logic is implemented in `genutiltypes.DefaultMessageValidator`. Use `genutil.NewAppModuleBasic` to create a new genutil Module Basic.
* (codec) [#12964](https://github.com/cosmos/cosmos-sdk/pull/12964) `ProtoCodec.MarshalInterface` now returns an error when serializing unregistered types and a subsequent `ProtoCodec.UnmarshalInterface` would fail.
* (x/staking) [#12973](https://github.com/cosmos/cosmos-sdk/pull/12973) Removed `stakingkeeper.RandomValidator`. Use `testutil.RandSliceElem(r, sk.GetAllValidators(ctx))` instead.
* (x/gov) [#13160](https://github.com/cosmos/cosmos-sdk/pull/13160) Remove custom marshaling of proposl and voteoption.
* (types) [#13430](https://github.com/cosmos/cosmos-sdk/pull/13430) Remove unused code `ResponseCheckTx` and `ResponseDeliverTx`
* (store) [#13529](https://github.com/cosmos/cosmos-sdk/pull/13529) Add method `LatestVersion` to `MultiStore` interface, add method `SetQueryMultiStore` to baesapp to support alternative `MultiStore` implementation for query service.
* (pruning) [#13609](https://github.com/cosmos/cosmos-sdk/pull/13609) Move pruning package to be under store package.
* [#13794](https://github.com/cosmos/cosmos-sdk/pull/13794) Most methods on `types/module.AppModule` have been moved to 
extension interfaces. `module.Manager.Modules` is now of type `map[string]interface{}` to support in parallel the new 
`cosmossdk.io/core/appmodule.AppModule` API.
* (signing) [#13701](https://github.com/cosmos/cosmos-sdk/pull/) Add `context.Context` as an argument `x/auth/signing.VerifySignature`.
* (x/group) [#13876](https://github.com/cosmos/cosmos-sdk/pull/13876) Add `GetMinExecutionPeriod` method on DecisionPolicy interface.
* (x/auth)[#13780](https://github.com/cosmos/cosmos-sdk/pull/13780) Querying with `id` (type of int64) in `AccountAddressByID` grpc query now throws error, use account-id(type of uint64) instead.
* (snapshots) [14048](https://github.com/cosmos/cosmos-sdk/pull/14048) Move the Snapshot package to the store package. This is done in an effort group all storage related logic under one package.
* (baseapp) [#14050](https://github.com/cosmos/cosmos-sdk/pull/14050) refactor `ABCIListener` interface to accept go contexts
* (store/streaming)[#14603](https://github.com/cosmos/cosmos-sdk/pull/14603) `StoreDecoderRegistry` moved from store to `types/simulations` this breaks the `AppModuleSimulation` interface. 
* (x/staking) [#14590](https://github.com/cosmos/cosmos-sdk/pull/14590) `MsgUndelegateResponse` now includes undelegated amount. `x/staking` module's `keeper.Undelegate` now returns 3 values (completionTime,undelegateAmount,error)  instead of 2.
* (x/feegrant) [14649](https://github.com/cosmos/cosmos-sdk/pull/14649) Extract Feegrant in its own go.mod and rename the package to `cosmossdk.io/x/feegrant`.
* (x/bank) [#14894](https://github.com/cosmos/cosmos-sdk/pull/14894) Return a human readable denomination for IBC vouchers when querying bank balances. Added a `ResolveDenom` parameter to `types.QueryAllBalancesRequest`.

### Client Breaking Changes

* (grpc-web) [#14652](https://github.com/cosmos/cosmos-sdk/pull/14652) Use same port for gRPC-Web and the API server.

### CLI Breaking Changes

* (x/gov) [#14880](https://github.com/cosmos/cosmos-sdk/pull/14880) Remove `simd tx gov submit-legacy-proposal cancel-software-upgrade` and `software-upgrade` commands. These commands are now in the `x/upgrade` module and using gov v1. Use `tx upgrade software-upgrade` instead.
* (grpc-web) [#14652](https://github.com/cosmos/cosmos-sdk/pull/14652) Remove `grpc-web.address` flag.
* (client) [#14342](https://github.com/cosmos/cosmos-sdk/pull/14342) `simd config` command is now a sub-command. Use `simd config --help` to learn more.
* (x/genutil) [#13535](https://github.com/cosmos/cosmos-sdk/pull/13535) Replace in `simd init`, the `--staking-bond-denom` flag with `--default-denom` which is used for all default denomination in the genesis, instead of only staking.
* (tx) [#12659](https://github.com/cosmos/cosmos-sdk/pull/12659) Remove broadcast mode `block`.
* (genesis) [#14149](https://github.com/cosmos/cosmos-sdk/pull/14149) Add `simd genesis` command, which contains all genesis-related sub-commands.

### Bug Fixes

* (store) [#14931](https://github.com/cosmos/cosmos-sdk/pull/14931) Exclude in-memory KVStores, i.e. `StoreTypeMemory`, from CommitInfo commitments.
* (types/coin) [#14715](https://github.com/cosmos/cosmos-sdk/pull/14715) `sdk.Coins.Add` now returns an empty set of coins `sdk.Coins{}` if both coins set are empty.
    * This is a behavior change, as previously `sdk.Coins.Add` would return `nil` in this case.
* (types/coin) [#14739](https://github.com/cosmos/cosmos-sdk/pull/14739) Deprecate the method `Coin.IsEqual` in favour of  `Coin.Equal`. The difference between the two methods is that the first one results in a panic when denoms are not equal. This panic lead to unexpected behaviour
* (x/bank) [#14538](https://github.com/cosmos/cosmos-sdk/pull/14538) Validate denom in bank balances GRPC queries.
* (baseapp) [#14505](https://github.com/cosmos/cosmos-sdk/pull/14505) PrepareProposal and ProcessProposal now use deliverState for the first block in order to access changes made in InitChain.
* (server) [#14441](https://github.com/cosmos/cosmos-sdk/pull/14441) Fix `--log_format` flag not working.
* (x/upgrade) [#13936](https://github.com/cosmos/cosmos-sdk/pull/13936) Make downgrade verification work again
* (x/group) [#13742](https://github.com/cosmos/cosmos-sdk/pull/13742) Fix `validate-genesis` when group policy accounts exist.
* (x/auth) [#13838](https://github.com/cosmos/cosmos-sdk/pull/13838) Fix calling `String()` when pubkey is set on a `BaseAccount`.
* (rosetta) [#13583](https://github.com/cosmos/cosmos-sdk/pull/13583) Misc fixes for cosmos-rosetta.
* (x/evidence) [#13740](https://github.com/cosmos/cosmos-sdk/pull/13740) Fix evidence query API to decode the hash properly.
* (bank) [#13691](https://github.com/cosmos/cosmos-sdk/issues/13691) Fix unhandled error for vesting account transfers, when total vesting amount exceeds total balance.
* [#13553](https://github.com/cosmos/cosmos-sdk/pull/13553) Ensure all parameter validation for decimal types handles nil decimal values.
* [#13145](https://github.com/cosmos/cosmos-sdk/pull/13145) Fix panic when calling `String()` to a Record struct type.
* [#13116](https://github.com/cosmos/cosmos-sdk/pull/13116) Fix a dead-lock in the `Group-TotalWeight` `x/group` invariant.
* (genutil) [#12140](https://github.com/cosmos/cosmos-sdk/pull/12140) Fix staking's genesis JSON migrate in the `simd migrate v0.46` CLI command.
* (types) [#12154](https://github.com/cosmos/cosmos-sdk/pull/12154) Add `baseAccountGetter` to avoid invalid account error when create vesting account.
* (x/authz) [#12184](https://github.com/cosmos/cosmos-sdk/pull/12184) Fix MsgExec not verifying the validity of nested messages.
* (x/staking) [#12303](https://github.com/cosmos/cosmos-sdk/pull/12303) Use bytes instead of string comparison in delete validator queue
* (store/rootmulti) [#12487](https://github.com/cosmos/cosmos-sdk/pull/12487) Fix non-deterministic map iteration.
* (sdk/dec_coins) [#12903](https://github.com/cosmos/cosmos-sdk/pull/12903) Fix nil `DecCoin` creation when converting `Coins` to `DecCoins`
* (x/gov) [#13051](https://github.com/cosmos/cosmos-sdk/pull/13051) In SubmitPropsal, when a legacy msg fails it's handler call, wrap the error as ErrInvalidProposalContent (instead of ErrNoProposalHandlerExists).
* (x/gov) [#13045](https://github.com/cosmos/cosmos-sdk/pull/13045) Fix gov migrations for v3(0.46).
* (snapshot) [#13400](https://github.com/cosmos/cosmos-sdk/pull/13400) Fix snapshot checksum issue in golang 1.19.
* (server) [#13778](https://github.com/cosmos/cosmos-sdk/pull/13778) Set Cosmos SDK default endpoints to localhost to avoid unknown exposure of endpoints.
* (x/auth) [#13877](https://github.com/cosmos/cosmos-sdk/pull/13877) Handle missing account numbers during `InitGenesis`.
* (cli) [#14509](https://github.com/cosmos/cosmos-sdk/pull/#14509) Added missing options to keyring-backend flag usage
* (cli) [#14919](https://github.com/cosmos/cosmos-sdk/pull/#14919) Fix never assigned error when write validators.

### Deprecated

* (x/evidence) [#13740](https://github.com/cosmos/cosmos-sdk/pull/13740) The `evidence_hash` field of `QueryEvidenceRequest` has been deprecated and now contains a new field `hash` with type `string`.
* (x/bank) [#11859](https://github.com/cosmos/cosmos-sdk/pull/11859) The Params.SendEnabled field is deprecated and unusable.
  The information can now be accessed using the BankKeeper.
  Setting can be done using MsgSetSendEnabled as a governance proposal.
  A SendEnabled query has been added to both GRPC and CLI.
* (x/staking) [#14567](https://github.com/cosmos/cosmos-sdk/pull/14567) The `delegator_address` field of `MsgCreateValidator` has been deprecated.
   The validator address bytes and delegator address bytes refer to the same account while creating validator (defer only in bech32 notation).

## [v0.46.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.8) - 2022-01-23

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
* (simapp) [#13107](https://github.com/cosmos/cosmos-sdk/pull/13107) Call `SetIAVLCacheSize` with the configured value in simapp.
* [#13301](https://github.com/cosmos/cosmos-sdk/pull/13301) Keep the balance query endpoint compatible with legacy blocks
* [#13321](https://github.com/cosmos/cosmos-sdk/pull/13321) Add flag to disable fast node migration and usage.

### Bug Fixes

* (types) [#13265](https://github.com/cosmos/cosmos-sdk/pull/13265) Correctly coalesce coins even with repeated denominations & simplify logic.
* (x/auth) [#13200](https://github.com/cosmos/cosmos-sdk/pull/13200) Fix wrong sequences in `sign-batch`.
* (export) [#13029](https://github.com/cosmos/cosmos-sdk/pull/13029) Fix exporting the blockParams regression.
* [#13046](https://github.com/cosmos/cosmos-sdk/pull/13046) Fix missing return statement in BaseApp.Query.
* (store) [#13336](https://github.com/cosmos/cosmos-sdk/pull/13334) Call streaming listeners for deliver tx event, it was removed accidentally.
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

## [v0.46.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.0) - 2022-07-26

### Features

* (types) [#11985](https://github.com/cosmos/cosmos-sdk/pull/11985) Add a `Priority` field on `sdk.Context`, which represents the CheckTx priority field. It is only used during CheckTx.
* (gRPC) [#11889](https://github.com/cosmos/cosmos-sdk/pull/11889) Support custom read and write gRPC options in `app.toml`. See `max-recv-msg-size` and `max-send-msg-size` respectively.
* (cli) [#11738](https://github.com/cosmos/cosmos-sdk/pull/11738) Add `tx auth multi-sign` as alias of `tx auth multisign` for consistency with `multi-send`.
* (cli) [#11738](https://github.com/cosmos/cosmos-sdk/pull/11738) Add `tx bank multi-send` command for bulk send of coins to multiple accounts.
* (grpc) [#11642](https://github.com/cosmos/cosmos-sdk/pull/11642) Implement `ABCIQuery` in the Tendermint gRPC service, which proxies ABCI `Query` requests directly to the application.
* (x/upgrade) [#11551](https://github.com/cosmos/cosmos-sdk/pull/11551) Update `ScheduleUpgrade` for chains to schedule an automated upgrade on `BeginBlock` without having to go though governance.
* (tx) [#11533](https://github.com/cosmos/cosmos-sdk/pull/11533) Register [`EIP191`](https://eips.ethereum.org/EIPS/eip-191) as an available `SignMode` for chains to use.
* (x/genutil) [#11500](https://github.com/cosmos/cosmos-sdk/pull/11500) Fix GenTx validation and adjust error messages
* [#11430](https://github.com/cosmos/cosmos-sdk/pull/11430) Introduce a new `grpc-only` flag, such that when enabled, will start the node in a query-only mode. Note, gRPC MUST be enabled with this flag.
* (x/bank) [#11417](https://github.com/cosmos/cosmos-sdk/pull/11417) Introduce a new `SpendableBalances` gRPC query that retrieves an account's total (paginated) spendable balances.
* [#11441](https://github.com/cosmos/cosmos-sdk/pull/11441) Added a new method, `IsLTE`, for `types.Coin`. This method is used to check if a `types.Coin` is less than or equal to another `types.Coin`.
* (x/upgrade) [#11116](https://github.com/cosmos/cosmos-sdk/pull/11116) `MsgSoftwareUpgrade` and `MsgCancelUpgrade` have been added to support v1beta2 msgs-based gov proposals.
* [#10977](https://github.com/cosmos/cosmos-sdk/pull/10977) Now every cosmos message protobuf definition must be extended with a `cosmos.msg.v1.signer` option to signal the signer fields in a language agnostic way.
* [#10710](https://github.com/cosmos/cosmos-sdk/pull/10710) Chain-id shouldn't be required for creating a transaction with both --generate-only and --offline flags.
* [#10703](https://github.com/cosmos/cosmos-sdk/pull/10703) Create a new grantee account, if the grantee of an authorization does not exist.
* [#10592](https://github.com/cosmos/cosmos-sdk/pull/10592) Add a `DecApproxEq` function that checks to see if `|d1 - d2| < tol` for some Dec `d1, d2, tol`.
* [#9933](https://github.com/cosmos/cosmos-sdk/pull/9933) Introduces the notion of a Cosmos "Scalar" type, which would just be simple aliases that give human-understandable meaning to the underlying type, both in Go code and in Proto definitions.
* [#9884](https://github.com/cosmos/cosmos-sdk/pull/9884) Provide a new gRPC query handler, `/cosmos/params/v1beta1/subspaces`, that allows the ability to query for all registered subspaces and their respective keys.
* [#9776](https://github.com/cosmos/cosmos-sdk/pull/9776) Add flag `staking-bond-denom` to specify the staking bond denomination value when initializing a new chain.
* [#9533](https://github.com/cosmos/cosmos-sdk/pull/9533) Added a new gRPC method, `DenomOwners`, in `x/bank` to query for all account holders of a specific denomination.
* (bank) [#9618](https://github.com/cosmos/cosmos-sdk/pull/9618) Update bank.Metadata: add URI and URIHash attributes.
* (store) [#8664](https://github.com/cosmos/cosmos-sdk/pull/8664) Implementation of ADR-038 file StreamingService
* [#9837](https://github.com/cosmos/cosmos-sdk/issues/9837) `--generate-only` flag can be used with a keyname from the keyring.
* [#10326](https://github.com/cosmos/cosmos-sdk/pull/10326) `x/authz` add all grants by granter query.
* [#10944](https://github.com/cosmos/cosmos-sdk/pull/10944) `x/authz` add all grants by grantee query
* [#10348](https://github.com/cosmos/cosmos-sdk/pull/10348) Add `fee.{payer,granter}` and `tip` fields to StdSignDoc for signing tipped transactions.
* [#10208](https://github.com/cosmos/cosmos-sdk/pull/10208) Add `TipsTxMiddleware` for transferring tips.
* [#10379](https://github.com/cosmos/cosmos-sdk/pull/10379) Add validation to `x/upgrade` CLI `software-upgrade` command `--plan-info` value.
* [#10507](https://github.com/cosmos/cosmos-sdk/pull/10507) Add antehandler for tx priority.
* [#10311](https://github.com/cosmos/cosmos-sdk/pull/10311) Adds cli to use tips transactions. It adds an `--aux` flag to all CLI tx commands to generate the aux signer data (with optional tip), and a new `tx aux-to-fee` subcommand to let the fee payer gather aux signer data and broadcast the tx
* [#11019](https://github.com/cosmos/cosmos-sdk/pull/11019) Add `MsgCreatePermanentLockedAccount` and CLI method for creating permanent locked account
* [#10947](https://github.com/cosmos/cosmos-sdk/pull/10947) Add `AllowancesByGranter` query to the feegrant module
* [#10407](https://github.com/cosmos/cosmos-sdk/pull/10407) Add validation to `x/upgrade` module's `BeginBlock` to check accidental binary downgrades
* (gov) [#11036](https://github.com/cosmos/cosmos-sdk/pull/11036) Add in-place migrations for 0.43->0.46. Add a `migrate v0.46` CLI command for v0.43->0.46 JSON genesis migration.
* [#11006](https://github.com/cosmos/cosmos-sdk/pull/11006) Add `debug pubkey-raw` command to allow inspecting of pubkeys in legacy bech32 format
* (x/authz) [#10714](https://github.com/cosmos/cosmos-sdk/pull/10714) Add support for pruning expired authorizations
* [#11179](https://github.com/cosmos/cosmos-sdk/pull/11179) Add state rollback command.
* [#11234](https://github.com/cosmos/cosmos-sdk/pull/11234) Add `GRPCClient` field to Client Context. If `GRPCClient` field is set to nil, the `Invoke` method would use ABCI query, otherwise use gprc.
* (authz)[#11060](https://github.com/cosmos/cosmos-sdk/pull/11060) Support grant with no expire time.
* (rosetta) [#11590](https://github.com/cosmos/cosmos-sdk/pull/11590) Add fee suggestion for rosetta and enable offline mode. Also force set events about Fees to Success to pass reconciliation test.
* (types) [#11959](https://github.com/cosmos/cosmos-sdk/pull/11959) Added `sdk.Coins.Find` helper method to find a coin by denom.
* (upgrade) [#12603](https://github.com/cosmos/cosmos-sdk/pull/12603) feat: Move AppModule.BeginBlock and AppModule.EndBlock to extension interfaces
* (telemetry) [#12405](https://github.com/cosmos/cosmos-sdk/pull/12405) Add *query* calls metric to telemetry.
* (query) [#12253](https://github.com/cosmos/cosmos-sdk/pull/12253) Add `GenericFilteredPaginate` to the `query` package to improve UX.

### API Breaking Changes

* (x/auth/ante) [#11985](https://github.com/cosmos/cosmos-sdk/pull/11985) The `MempoolFeeDecorator` has been removed. Instead, the `DeductFeeDecorator` takes a new argument of type `TxFeeChecker`, to define custom fee models. If `nil` is passed to this `TxFeeChecker` argument, then it will default to `checkTxFeeWithValidatorMinGasPrices`, which is the exact same behavior as the old `MempoolFeeDecorator` (i.e. checking fees against validator's own min gas price).
* (x/auth/ante) [#11985](https://github.com/cosmos/cosmos-sdk/pull/11985) The `ExtensionOptionsDecorator` takes an argument of type `ExtensionOptionChecker`. For backwards-compatibility, you can pass `nil`, which defaults to the old behavior of rejecting all tx extensions.
* (crypto/keyring) [#11932](https://github.com/cosmos/cosmos-sdk/pull/11932) Remove `Unsafe*` interfaces from keyring package. Please use interface casting if you wish to access those unsafe functions.
* (types) [#11881](https://github.com/cosmos/cosmos-sdk/issues/11881) Rename `AccAddressFromHex` to `AccAddressFromHexUnsafe`.
* (types) [#11788](https://github.com/cosmos/cosmos-sdk/pull/11788) The `Int` and `Uint` types have been moved to their own dedicated module, `math`. Aliases are kept in the SDK's root `types` package, however, it is encouraged to utilize the new `math` module. As a result, the `Int#ToDec` API has been removed.
* (grpc) [#11642](https://github.com/cosmos/cosmos-sdk/pull/11642) The `RegisterTendermintService` method in the `tmservice` package now requires a `abciQueryFn` query function parameter.
* [#11496](https://github.com/cosmos/cosmos-sdk/pull/11496) Refactor abstractions for snapshot and pruning; snapshot intervals eventually pruned; unit tests.
* (types) [#11689](https://github.com/cosmos/cosmos-sdk/pull/11689) Make `Coins#Sub` and `Coins#SafeSub` consistent with `Coins#Add`.
* (store)[#11152](https://github.com/cosmos/cosmos-sdk/pull/11152) Remove `keep-every` from pruning options.
* [#10950](https://github.com/cosmos/cosmos-sdk/pull/10950) Add `envPrefix` parameter to `cmd.Execute`.
* (x/mint) [#10441](https://github.com/cosmos/cosmos-sdk/pull/10441) The `NewAppModule` function now accepts an inflation calculation function as an argument.
* [#9695](https://github.com/cosmos/cosmos-sdk/pull/9695) Migrate keys from `Info` (serialized as amino) -> `Record` (serialized as proto)
    * Add new `codec.Codec` argument in:
        * `keyring.NewInMemory`
        * `keyring.New`
    * Rename:
        * `SavePubKey` to `SaveOfflineKey`.
        * `NewMultiInfo`, `NewLedgerInfo` to `NewLegacyMultiInfo`, `newLegacyLedgerInfo` respectively. Move them into `legacy_info.go`.
        * `NewOfflineInfo` to `newLegacyOfflineInfo` and move it to `migration_test.go`.
    * Return:
    _`keyring.Record, error` in `SaveOfflineKey`, `SaveLedgerKey`, `SaveMultiSig`, `Key` and `KeyByAddress`.
    _`keyring.Record` instead of `Info` in `NewMnemonic` and `List`.
    * Remove `algo` argument from :
        * `SaveOfflineKey`
    * Take `keyring.Record` instead of `Info` as first argument in:
        * `MkConsKeyOutput`
        * `MkValKeyOutput`
        * `MkAccKeyOutput`
* [#10022](https://github.com/cosmos/cosmos-sdk/pull/10022) `AuthKeeper` interface in `x/auth` now includes a function `HasAccount`.
* [#9759](https://github.com/cosmos/cosmos-sdk/pull/9759) `NewAccountKeeeper` in `x/auth` now takes an additional `bech32Prefix` argument that represents `sdk.Bech32MainPrefix`.
* [#9628](https://github.com/cosmos/cosmos-sdk/pull/9628) Rename `x/{mod}/legacy` to `x/{mod}/migrations`.
* [#9571](https://github.com/cosmos/cosmos-sdk/pull/9571) Implemented error handling for staking hooks, which now return an error on failure.
* [#9427](https://github.com/cosmos/cosmos-sdk/pull/9427) Move simapp `FundAccount` and `FundModuleAccount` to `x/bank/testutil`
* (client/tx) [#9421](https://github.com/cosmos/cosmos-sdk/pull/9421/) `BuildUnsignedTx`, `BuildSimTx`, `PrintUnsignedStdTx` functions are moved to
  the Tx Factory as methods.
* (client/keys) [#9601](https://github.com/cosmos/cosmos-sdk/pull/9601) Added `keys rename` CLI command and `Keyring.Rename` interface method to rename a key in the keyring.
* (x/slashing) [#9458](https://github.com/cosmos/cosmos-sdk/pull/9458) Coins burned from slashing is now returned from Slash function and included in Slash event.
* [#9246](https://github.com/cosmos/cosmos-sdk/pull/9246) The `New` method for the network package now returns an error.
* [#9519](https://github.com/cosmos/cosmos-sdk/pull/9519) `DeleteDeposits` renamed to `DeleteAndBurnDeposits`, `RefundDeposits` renamed to `RefundAndDeleteDeposits`
* (codec) [#9521](https://github.com/cosmos/cosmos-sdk/pull/9521) Removed deprecated `clientCtx.JSONCodec` from `client.Context`.
* (codec) [#9521](https://github.com/cosmos/cosmos-sdk/pull/9521) Rename `EncodingConfig.Marshaler` to `Codec`.
* [#9594](https://github.com/cosmos/cosmos-sdk/pull/9594) `RESTHandlerFn` argument is removed from the `gov/NewProposalHandler`.
* [#9594](https://github.com/cosmos/cosmos-sdk/pull/9594) `types/rest` package moved to `testutil/rest`.
* [#9432](https://github.com/cosmos/cosmos-sdk/pull/9432) `ConsensusParamsKeyTable` moved from `params/keeper` to `params/types`
* [#9576](https://github.com/cosmos/cosmos-sdk/pull/9576) Add debug error message to `sdkerrors.QueryResult` when enabled
* [#9650](https://github.com/cosmos/cosmos-sdk/pull/9650) Removed deprecated message handler implementation from the SDK modules.
* [#10248](https://github.com/cosmos/cosmos-sdk/pull/10248) Remove unused `KeyPowerReduction` variable from x/staking types.
* (x/bank) [#9832](https://github.com/cosmos/cosmos-sdk/pull/9832) `AddressFromBalancesStore` renamed to `AddressAndDenomFromBalancesStore`.
* (tests) [#9938](https://github.com/cosmos/cosmos-sdk/pull/9938) `simapp.Setup` accepts additional `testing.T` argument.
* (baseapp) [#11979](https://github.com/cosmos/cosmos-sdk/pull/11979) Rename baseapp simulation helper methods `baseapp.{Check,Deliver}` to `baseapp.Sim{Check,Deliver}`.
* (x/gov) [#10373](https://github.com/cosmos/cosmos-sdk/pull/10373) Removed gov `keeper.{MustMarshal, MustUnmarshal}`.
* [#10348](https://github.com/cosmos/cosmos-sdk/pull/10348) StdSignBytes takes a new argument of type `*tx.Tip` for signing over tips using LEGACY_AMINO_JSON.
* [#10208](https://github.com/cosmos/cosmos-sdk/pull/10208) The `x/auth/signing.Tx` interface now also includes a new `GetTip() *tx.Tip` method for verifying tipped transactions. The `x/auth/types` expected BankKeeper interface now expects the `SendCoins` method too.
* [#10612](https://github.com/cosmos/cosmos-sdk/pull/10612) `baseapp.NewBaseApp` constructor function doesn't take the `sdk.TxDecoder` anymore. This logic has been moved into the TxDecoderMiddleware.
* [#10692](https://github.com/cosmos/cosmos-sdk/pull/10612) `SignerData` takes 2 new fields, `Address` and `PubKey`, which need to get populated when using SIGN_MODE_DIRECT_AUX.
* [#10748](https://github.com/cosmos/cosmos-sdk/pull/10748) Move legacy `x/gov` api to `v1beta1` directory.
* [#10816](https://github.com/cosmos/cosmos-sdk/pull/10816) Reuse blocked addresses from the bank module. No need to pass them to distribution.
* [#10852](https://github.com/cosmos/cosmos-sdk/pull/10852) Move `x/gov/types` to `x/gov/types/v1beta2`.
* [#10922](https://github.com/cosmos/cosmos-sdk/pull/10922), [/#10957](https://github.com/cosmos/cosmos-sdk/pull/10957) Move key `server.Generate*` functions to testutil and support custom mnemonics in in-process testing network. Moved `TestMnemonic` from `testutil` package to `testdata`.
* (x/bank) [#10771](https://github.com/cosmos/cosmos-sdk/pull/10771) Add safety check on bank module perms to allow module-specific mint restrictions (e.g. only minting a certain denom).
* (x/bank) [#10771](https://github.com/cosmos/cosmos-sdk/pull/10771) Add `bank.BaseKeeper.WithMintCoinsRestriction` function to restrict use of bank `MintCoins` usage.
* [#10868](https://github.com/cosmos/cosmos-sdk/pull/10868), [#10989](https://github.com/cosmos/cosmos-sdk/pull/10989) The Gov keeper accepts now 2 more mandatory arguments, the ServiceMsgRouter and a maximum proposal metadata length.
* [#10868](https://github.com/cosmos/cosmos-sdk/pull/10868), [#10989](https://github.com/cosmos/cosmos-sdk/pull/10989), [#11093](https://github.com/cosmos/cosmos-sdk/pull/11093) The Gov keeper accepts now 2 more mandatory arguments, the ServiceMsgRouter and a gov Config including the max metadata length.
* [#11124](https://github.com/cosmos/cosmos-sdk/pull/11124) Add `GetAllVersions` to application store
* (x/authz) [#10447](https://github.com/cosmos/cosmos-sdk/pull/10447) authz `NewGrant` takes a new argument: block time, to correctly validate expire time.
* [#10961](https://github.com/cosmos/cosmos-sdk/pull/10961) Support third-party modules to add extension snapshots to state-sync.
* [#11274](https://github.com/cosmos/cosmos-sdk/pull/11274) `types/errors.New` now is an alias for `types/errors.Register` and should only be used in initialization code.
* (authz)[#11060](https://github.com/cosmos/cosmos-sdk/pull/11060) `authz.NewMsgGrant` `expiration` is now a pointer. When `nil` is used then no expiration will be set (grant won't expire).
* (x/distribution)[#11457](https://github.com/cosmos/cosmos-sdk/pull/11457) Add amount field to `distr.MsgWithdrawDelegatorRewardResponse` and `distr.MsgWithdrawValidatorCommissionResponse`.
* [#11334](https://github.com/cosmos/cosmos-sdk/pull/11334) Move `x/gov/types/v1beta2` to `x/gov/types/v1`.
* (x/auth/middleware) [#11413](https://github.com/cosmos/cosmos-sdk/pull/11413) Refactor tx middleware to be extensible on tx fee logic. Merged `MempoolFeeMiddleware` and `TxPriorityMiddleware` functionalities into `DeductFeeMiddleware`, make the logic extensible using the `TxFeeChecker` option, the current fee logic is preserved by the default `checkTxFeeWithValidatorMinGasPrices` implementation. Change `RejectExtensionOptionsMiddleware` to `NewExtensionOptionsMiddleware` which is extensible with the `ExtensionOptionChecker` option. Unpack the tx extension options `Any`s to interface `TxExtensionOptionI`.
* (migrations) [#11556](https://github.com/cosmos/cosmos-sdk/pull/11556#issuecomment-1091385011) Remove migration code from 0.42 and below. To use previous migrations, checkout previous versions of the cosmos-sdk.

### Client Breaking Changes

* [#11797](https://github.com/cosmos/cosmos-sdk/pull/11797) Remove all RegisterRESTRoutes (previously deprecated)
* [#11089](https://github.com/cosmos/cosmos-sdk/pull/11089]) interacting with the node through `grpc.Dial` requires clients to pass a codec refer to [doc](docs/docs/run-node/02-interact-node.md).
* [#9594](https://github.com/cosmos/cosmos-sdk/pull/9594) Remove legacy REST API. Please see the [REST Endpoints Migration guide](https://docs.cosmos.network/v0.45/migrations/rest.html) to migrate to the new REST endpoints.
* [#9995](https://github.com/cosmos/cosmos-sdk/pull/9995) Increased gas cost for creating proposals.
* [#11029](https://github.com/cosmos/cosmos-sdk/pull/11029) The deprecated Vote Option field is removed in gov v1beta2 and nil in v1beta1. Use Options instead.
* [#11013](https://github.com/cosmos/cosmos-sdk/pull/11013) The `tx gov submit-proposal` command has changed syntax to support the new Msg-based gov proposals. To access the old CLI command, please use `tx gov submit-legacy-proposal`.
* [#11170](https://github.com/cosmos/cosmos-sdk/issues/11170) Fixes issue related to grpc-gateway of supply by ibc-denom.

### CLI Breaking Changes

* (cli) [#11818](https://github.com/cosmos/cosmos-sdk/pull/11818) CLI transactions preview now respect the chosen `--output` flag format (json or text).
* [#9695](https://github.com/cosmos/cosmos-sdk/pull/9695) `<app> keys migrate` CLI command now takes no arguments.
* [#9246](https://github.com/cosmos/cosmos-sdk/pull/9246) Removed the CLI flag `--setup-config-only` from the `testnet` command and added the subcommand `init-files`.
* [#9780](https://github.com/cosmos/cosmos-sdk/pull/9780) Use sigs.k8s.io for yaml, which might lead to minor YAML output changes
* [#10625](https://github.com/cosmos/cosmos-sdk/pull/10625) Rename `--fee-account` CLI flag to `--fee-granter`
* [#10684](https://github.com/cosmos/cosmos-sdk/pull/10684) Rename `edit-validator` command's `--moniker` flag to `--new-moniker`
* (authz)[#11060](https://github.com/cosmos/cosmos-sdk/pull/11060) Changed the default value of the `--expiration` `tx grant` CLI Flag: was now + 1year, update: null (no expire date).

### Improvements

* (types) [#12201](https://github.com/cosmos/cosmos-sdk/pull/12201) Add `MustAccAddressFromBech32` util function
* [#11696](https://github.com/cosmos/cosmos-sdk/pull/11696) Rename `helpers.GenTx` to `GenSignedMockTx` to avoid confusion with genutil's `GenTxCmd`.
* (x/auth/vesting) [#11652](https://github.com/cosmos/cosmos-sdk/pull/11652) Add util functions for `Period(s)`
* [#11630](https://github.com/cosmos/cosmos-sdk/pull/11630) Add SafeSub method to sdk.Coin.
* [#11511](https://github.com/cosmos/cosmos-sdk/pull/11511) Add api server flags to start command.
* [#11484](https://github.com/cosmos/cosmos-sdk/pull/11484) Implement getter for keyring backend option.
* [#11449](https://github.com/cosmos/cosmos-sdk/pull/11449) Improved error messages when node isn't synced.
* [#11349](https://github.com/cosmos/cosmos-sdk/pull/11349) Add `RegisterAminoMsg` function that checks that a msg name is <40 chars (else this would break ledger nano signing) then registers the concrete msg type with amino, it should be used for registering `sdk.Msg`s with amino instead of `cdc.RegisterConcrete`.
* [#11089](https://github.com/cosmos/cosmos-sdk/pull/11089]) Now cosmos-sdk consumers can upgrade gRPC to its newest versions.
* [#10439](https://github.com/cosmos/cosmos-sdk/pull/10439) Check error for `RegisterQueryHandlerClient` in all modules `RegisterGRPCGatewayRoutes`.
* [#9780](https://github.com/cosmos/cosmos-sdk/pull/9780) Remove gogoproto `moretags` YAML annotations and add `sigs.k8s.io/yaml` for YAML marshalling.
* (x/bank) [#10134](https://github.com/cosmos/cosmos-sdk/pull/10134) Add `HasDenomMetadata` function to bank `Keeper` to check if a client coin denom metadata exists in state.
* (x/bank) [#10022](https://github.com/cosmos/cosmos-sdk/pull/10022) `BankKeeper.SendCoins` now takes less execution time.
* (deps) [#9987](https://github.com/cosmos/cosmos-sdk/pull/9987) Bump Go version minimum requirement to `1.17`
* (cli) [#9856](https://github.com/cosmos/cosmos-sdk/pull/9856) Overwrite `--sequence` and `--account-number` flags with default flag values when used with `offline=false` in `sign-batch` command.
* (rosetta) [#10001](https://github.com/cosmos/cosmos-sdk/issues/10001) Add documentation for rosetta-cli dockerfile and rename folder for the rosetta-ci dockerfile
* [#9699](https://github.com/cosmos/cosmos-sdk/pull/9699) Add `:`, `.`, `-`, and `_` as allowed characters in the default denom regular expression.
* (genesis) [#9697](https://github.com/cosmos/cosmos-sdk/pull/9697) Ensure `InitGenesis` returns with non-empty validator set.
* [#10468](https://github.com/cosmos/cosmos-sdk/pull/10468) Allow futureOps to queue additional operations in simulations
* [#10625](https://github.com/cosmos/cosmos-sdk/pull/10625) Add `--fee-payer` CLI flag
* (cli) [#10683](https://github.com/cosmos/cosmos-sdk/pull/10683) In CLI, allow 1 SIGN_MODE_DIRECT signer in transactions with multiple signers.
* (deps) [#10706](https://github.com/cosmos/cosmos-sdk/issues/10706) Bump rosetta-sdk-go to v0.7.2 and rosetta-cli to v0.7.3
* (types/errors) [#10779](https://github.com/cosmos/cosmos-sdk/pull/10779) Move most functionality in `types/errors` to a standalone `errors` go module, except the `RootCodespace` errors and ABCI response helpers. All functions and types that used to live in `types/errors` are now aliased so this is not a breaking change.
* (gov) [#10854](https://github.com/cosmos/cosmos-sdk/pull/10854) v1beta2's vote doesn't include the deprecate `option VoteOption` anymore. Instead, it only uses `WeightedVoteOption`.
* (types) [#11004](https://github.com/cosmos/cosmos-sdk/pull/11004) Added mutable versions of many of the sdk.Dec types operations. This improves performance when used by avoiding reallocating a new bigint for each operation.
* (x/auth) [#10880](https://github.com/cosmos/cosmos-sdk/pull/10880) Added a new query to the tx query service that returns a block with transactions fully decoded.
* (types) [#11200](https://github.com/cosmos/cosmos-sdk/pull/11200) Added `Min()` and `Max()` operations on sdk.Coins.
* (gov) [#11287](https://github.com/cosmos/cosmos-sdk/pull/11287) Fix error message when no flags are provided while executing `submit-legacy-proposal` transaction.
* (x/auth) [#11482](https://github.com/cosmos/cosmos-sdk/pull/11482) Improve panic message when attempting to register a method handler for a message that does not implement sdk.Msg
* (x/staking) [#11596](https://github.com/cosmos/cosmos-sdk/pull/11596) Add (re)delegation getters
* (errors) [#11960](https://github.com/cosmos/cosmos-sdk/pull/11960) Removed 'redacted' error message from defaultErrEncoder
* (ante) [#12013](https://github.com/cosmos/cosmos-sdk/pull/12013) Index ante events for failed tx.
* [#12668](https://github.com/cosmos/cosmos-sdk/pull/12668) Add `authz_msg_index` event attribute to message events emitted when executing via `MsgExec` through `x/authz`.
* [#12626](https://github.com/cosmos/cosmos-sdk/pull/12626) Upgrade IAVL to v0.19.0 with fast index and error propagation. NOTE: first start will take a while to propagate into new model.
* [#12576](https://github.com/cosmos/cosmos-sdk/pull/12576) Remove dependency on cosmos/keyring and upgrade to 99designs/keyring v1.2.1
* [#12590](https://github.com/cosmos/cosmos-sdk/pull/12590) Allow zero gas in simulation mode.
* [#12453](https://github.com/cosmos/cosmos-sdk/pull/12453) Add `NewInMemoryWithKeyring` function which allows the creation of in memory `keystore` instances with a specified set of existing items.
* [#11390](https://github.com/cosmos/cosmos-sdk/pull/11390) `LatestBlockResponse` & `BlockByHeightResponse` types' `Block` filed has been deprecated and they now contains new field `sdk_block` with `proposer_address` as `string`
* [#12089](https://github.com/cosmos/cosmos-sdk/pull/12089) Mark the `TipDecorator` as beta, don't include it in simapp by default.
* [#12153](https://github.com/cosmos/cosmos-sdk/pull/12153) Add a new `NewSimulationManagerFromAppModules` constructor, to simplify simulation wiring.

### Bug Fixes

* [#11969](https://github.com/cosmos/cosmos-sdk/pull/11969) Fix the panic error in `x/upgrade` when `AppVersion` is not set.
* (tests) [#11940](https://github.com/cosmos/cosmos-sdk/pull/11940) Fix some client tests in the `x/gov` module
* [#11772](https://github.com/cosmos/cosmos-sdk/pull/11772) Limit types.Dec length to avoid overflow.
* [#11724](https://github.com/cosmos/cosmos-sdk/pull/11724) Fix data race issues with api.Server
* [#11693](https://github.com/cosmos/cosmos-sdk/pull/11693) Add validation for gentx cmd.
* [#11645](https://github.com/cosmos/cosmos-sdk/pull/11645) Fix `--home` flag ignored when running help.
* [#11558](https://github.com/cosmos/cosmos-sdk/pull/11558) Fix `--dry-run` not working when using tx command.
* [#11354](https://github.com/cosmos/cosmos-sdk/pull/11355) Added missing pagination flag for `bank q total` query.
* [#11197](https://github.com/cosmos/cosmos-sdk/pull/11197) Signing with multisig now works with multisig address which is not in the keyring.
* (makefile) [#11285](https://github.com/cosmos/cosmos-sdk/pull/11285) Fix lint-fix make target.
* (client) [#11283](https://github.com/cosmos/cosmos-sdk/issues/11283) Support multiple keys for tx simulation and setting automatic gas for txs.
* (store) [#11177](https://github.com/cosmos/cosmos-sdk/pull/11177) Update the prune `everything` strategy to store the last two heights.
* [#10844](https://github.com/cosmos/cosmos-sdk/pull/10844) Automatic recovering non-consistent keyring storage during public key import.
* (store) [#11117](https://github.com/cosmos/cosmos-sdk/pull/11117) Fix data race in store trace component
* (cli) [#11065](https://github.com/cosmos/cosmos-sdk/pull/11065) Ensure the `tendermint-validator-set` query command respects the `-o` output flag.
* (grpc) [#10985](https://github.com/cosmos/cosmos-sdk/pull/10992) The `/cosmos/tx/v1beta1/txs/{hash}` endpoint returns a 404 when a tx does not exist.
* (rosetta) [#10340](https://github.com/cosmos/cosmos-sdk/pull/10340) Use `GenesisChunked(ctx)` instead `Genesis(ctx)` to get genesis block height
* [#9651](https://github.com/cosmos/cosmos-sdk/pull/9651) Change inconsistent limit of `0` to `MaxUint64` on InfiniteGasMeter and add GasRemaining func to GasMeter.
* [#9639](https://github.com/cosmos/cosmos-sdk/pull/9639) Check store keys length before accessing them by making sure that `key` is of length `m+1` (for `key[n:m]`)
* (types) [#9627](https://github.com/cosmos/cosmos-sdk/pull/9627) Fix nil pointer panic on `NewBigIntFromInt`
* (x/genutil) [#9574](https://github.com/cosmos/cosmos-sdk/pull/9575) Actually use the `gentx` client tx flags (like `--keyring-dir`)
* (x/distribution) [#9599](https://github.com/cosmos/cosmos-sdk/pull/9599) Withdraw rewards event now includes a value attribute even if there are 0 rewards (due to situations like 100% commission).
* (x/genutil) [#9638](https://github.com/cosmos/cosmos-sdk/pull/9638) Added missing validator key save when recovering from mnemonic
* [#9762](https://github.com/cosmos/cosmos-sdk/pull/9762) The init command uses the chain-id from the client config if --chain-id is not provided
* [#9980](https://github.com/cosmos/cosmos-sdk/pull/9980) Returning the error when the invalid argument is passed to bank query total supply cli.
* (server) [#10016](https://github.com/cosmos/cosmos-sdk/issues/10016) Fix marshaling of index-events into server config file.
* [#10184](https://github.com/cosmos/cosmos-sdk/pull/10184) Fixed CLI tx commands to no longer explicitly require the chain-id flag as this value can come from a user config.
* (x/upgrade) [#10189](https://github.com/cosmos/cosmos-sdk/issues/10189) Removed potential sources of non-determinism in upgrades
* [#10258](https://github.com/cosmos/cosmos-sdk/issues/10258) Fixes issue related to segmentation fault on mac m1 arm64
* [#10466](https://github.com/cosmos/cosmos-sdk/issues/10466) Fixes error with simulation tests when genesis start time is randomly created after the year 2262
* [#10394](https://github.com/cosmos/cosmos-sdk/issues/10394) Fixes issue related to grpc-gateway of account balance by
  ibc-denom.
* [#10842](https://github.com/cosmos/cosmos-sdk/pull/10842) Fix error when `--generate-only`, `--max-msgs` fags set while executing `WithdrawAllRewards` command.
* [#10897](https://github.com/cosmos/cosmos-sdk/pull/10897) Fix: set a non-zero value on gas overflow.
* [#9790](https://github.com/cosmos/cosmos-sdk/pull/10687) Fix behavior of `DecCoins.MulDecTruncate`.
* [#10990](https://github.com/cosmos/cosmos-sdk/pull/10990) Fixes missing `iavl-cache-size` config parsing in `GetConfig` method.
* (x/authz) [#10447](https://github.com/cosmos/cosmos-sdk/pull/10447) Fix authz `NewGrant` expiration check.
* (x/authz) [#10633](https://github.com/cosmos/cosmos-sdk/pull/10633) Fixed authorization not found error when executing message.
* [#11222](https://github.com/cosmos/cosmos-sdk/pull/11222) reject query with block height in the future
* [#11229](https://github.com/cosmos/cosmos-sdk/pull/11229) Handled the error message of `transaction encountered error` from tendermint.
* (x/authz) [#11252](https://github.com/cosmos/cosmos-sdk/pull/11252) Allow insufficient funds error for authz simulation
* (cli) [#11313](https://github.com/cosmos/cosmos-sdk/pull/11313) Fixes `--gas auto` when executing CLI transactions in `--generate-only` mode
* (cli) [#11337](https://github.com/cosmos/cosmos-sdk/pull/11337) Fixes `show-adress` cli cmd
* (crypto) [#11298](https://github.com/cosmos/cosmos-sdk/pull/11298) Fix cgo secp signature verification and update libscep256k1 library.
* (x/authz) [#11512](https://github.com/cosmos/cosmos-sdk/pull/11512) Fix response of a panic to error, when subtracting balances.
* (rosetta) [#11590](https://github.com/cosmos/cosmos-sdk/pull/11590) `/block` returns an error with nil pointer when a request has both of index and hash and increase timeout for huge genesis.
* (x/feegrant) [#11813](https://github.com/cosmos/cosmos-sdk/pull/11813) Fix pagination total count in `AllowancesByGranter` query.
* (simapp) [#11855](https://github.com/cosmos/cosmos-sdk/pull/11855) Use `sdkmath.Int` instead of `int64` for `SimulationState.InitialStake`.
* (x/capability) [#11737](https://github.com/cosmos/cosmos-sdk/pull/11737) Use a fixed length encoding of `Capability` pointer for `FwdCapabilityKey`
* [#11983](https://github.com/cosmos/cosmos-sdk/pull/11983) (x/feegrant, x/authz) rename grants query commands to `grants-by-grantee`, `grants-by-granter` cmds.
* (testutil/sims) [#12374](https://github.com/cosmos/cosmos-sdk/pull/12374) fix the non-determinstic behavior in simulations caused by `GenSignedMockTx` and check empty coins slice before it is used to create `banktype.MsgSend`.
* [#12448](https://github.com/cosmos/cosmos-sdk/pull/12448) Start telemetry independently from the API server.
* [#12509](https://github.com/cosmos/cosmos-sdk/pull/12509) Fix `Register{Tx,Tendermint}Service` not being called, resulting in some endpoints like the Simulate endpoint not working.
* [#12416](https://github.com/cosmos/cosmos-sdk/pull/12416) Prevent zero gas transactions in the `DeductFeeDecorator` AnteHandler decorator.
* (x/mint) [#12384](https://github.com/cosmos/cosmos-sdk/pull/12384) Ensure `GoalBonded` must be positive when performing `x/mint` parameter validation.
* (x/auth) [#12261](https://github.com/cosmos/cosmos-sdk/pull/12261) Deprecate pagination in GetTxsEventRequest/Response in favor of page and limit to align with tendermint `SignClient.TxSearch`
* (vesting) [#12190](https://github.com/cosmos/cosmos-sdk/pull/12190) Replace https://github.com/cosmos/cosmos-sdk/pull/12190 to use `NewBaseAccountWithAddress` in all vesting account message handlers.
* (linting) [#12132](https://github.com/cosmos/cosmos-sdk/pull/12132) Change sdk.Int to math.Int
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

* (x/gov) [#13576](https://github.com/cosmos/cosmos-sdk/pull/13576) Proposals in voting period are tracked in a separate store.
* (baseapp) [#11985](https://github.com/cosmos/cosmos-sdk/pull/11985) Add a `postHandler` to baseapp. This `postHandler` is like antehandler, but is run *after* the `runMsgs` execution. It is in the same store branch that `runMsgs`, meaning that both `runMsgs` and `postHandler`
* (x/gov) [#11998](https://github.com/cosmos/cosmos-sdk/pull/11998) Tweak the `x/gov` `ModuleAccountInvariant` invariant to ensure deposits are `<=` total module account balance instead of strictly equal.
* (x/upgrade) [#11800](https://github.com/cosmos/cosmos-sdk/pull/11800) Fix `GetLastCompleteUpgrade` to properly return the latest upgrade.
* [#10564](https://github.com/cosmos/cosmos-sdk/pull/10564) Fix bug when updating allowance inside AllowedMsgAllowance
* (x/auth)[#9596](https://github.com/cosmos/cosmos-sdk/pull/9596) Enable creating periodic vesting accounts with a transactions instead of requiring them to be created in genesis.
* (x/bank) [#9611](https://github.com/cosmos/cosmos-sdk/pull/9611) Introduce a new index to act as a reverse index between a denomination and address allowing to query for token holders of a specific denomination. `DenomOwners` is updated to use the new reverse index.
* (x/bank) [#9832](https://github.com/cosmos/cosmos-sdk/pull/9832) Account balance is stored as `sdk.Int` rather than `sdk.Coin`.
* (x/bank) [#9890](https://github.com/cosmos/cosmos-sdk/pull/9890) Remove duplicate denom from denom metadata key.
* (x/upgrade) [#10189](https://github.com/cosmos/cosmos-sdk/issues/10189) Removed potential sources of non-determinism in upgrades
* [#10422](https://github.com/cosmos/cosmos-sdk/pull/10422) and [#10529](https://github.com/cosmos/cosmos-sdk/pull/10529) Add `MinCommissionRate` param to `x/staking` module.
* (x/gov) [#10763](https://github.com/cosmos/cosmos-sdk/pull/10763) modify the fields in `TallyParams` to use `string` instead of `bytes`
* [#10770](https://github.com/cosmos/cosmos-sdk/pull/10770) revert tx when block gas limit exceeded
* (x/gov) [#10868](https://github.com/cosmos/cosmos-sdk/pull/10868) Bump gov to v1. Both v1beta1 and v1beta2 queries and Msgs are accepted.
* [#11011](https://github.com/cosmos/cosmos-sdk/pull/11011) Remove burning of deposits when qourum is not reached on a governance proposal and when the deposit is not fully met.
* [#11019](https://github.com/cosmos/cosmos-sdk/pull/11019) Add `MsgCreatePermanentLockedAccount` and CLI method for creating permanent locked account
* (x/staking) [#10885] (https://github.com/cosmos/cosmos-sdk/pull/10885) Add new `CancelUnbondingDelegation`
  transaction to `x/staking` module. Delegators can now cancel unbonding delegation entry and delegate back to validator.
* (x/feegrant) [#10830](https://github.com/cosmos/cosmos-sdk/pull/10830) Expired allowances will be pruned from state.
* (x/authz,x/feegrant) [#11214](https://github.com/cosmos/cosmos-sdk/pull/11214) Fix Amino JSON encoding of authz and feegrant Msgs to be consistent with other modules.
* (authz)[#11060](https://github.com/cosmos/cosmos-sdk/pull/11060) Support grant with no expire time.

### Deprecated

* (x/upgrade) [#9906](https://github.com/cosmos/cosmos-sdk/pull/9906) Deprecate `UpgradeConsensusState` gRPC query since this functionality is only used for IBC, which now has its own [IBC replacement](https://github.com/cosmos/ibc-go/blob/2c880a22e9f9cc75f62b527ca94aa75ce1106001/proto/ibc/core/client/v1/query.proto#L54)
* (types) [#10948](https://github.com/cosmos/cosmos-sdk/issues/10948) Deprecate the types.DBBackend variable and types.NewLevelDB function. They are replaced by a new entry in `app.toml`: `app-db-backend` and `tendermint/tm-db`s `NewDB` function. If `app-db-backend` is defined, then it is used. Otherwise, if `types.DBBackend` is defined, it is used (until removed: [#11241](https://github.com/cosmos/cosmos-sdk/issues/11241)). Otherwise, Tendermint config's `db-backend` is used.

## v0.45.12 - 2023-01-23

### Improvements

* [#13881](https://github.com/cosmos/cosmos-sdk/pull/13881) Optimize iteration on nested cached KV stores and other operations in general.
* (store) [#11646](https://github.com/cosmos/cosmos-sdk/pull/11646) Add store name in tracekv-emitted store traces
* (deps) Bump Tendermint version to [v0.34.24](https://github.com/tendermint/tendermint/releases/tag/v0.34.24) and use Informal Systems fork.

### API Breaking Changes

* (store) [#13516](https://github.com/cosmos/cosmos-sdk/pull/13516) Update State Streaming APIs:
    * Add method `ListenCommit` to `ABCIListener`
    * Move `ListeningEnabled` and  `AddListener` methods to `CommitMultiStore`
    * Remove `CacheWrapWithListeners` from `CacheWrap` and `CacheWrapper` interfaces
    * Remove listening APIs from the caching layer (it should only listen to the `rootmulti.Store`)
    * Add three new options to file streaming service constructor.
    * Modify `ABCIListener` such that any error from any method will always halt the app via `panic`

### Bug Fixes

* (store) [#13516](https://github.com/cosmos/cosmos-sdk/pull/13516) Fix state listener that was observing writes at wrong time.
* (store) [#12945](https://github.com/cosmos/cosmos-sdk/pull/12945) Fix nil end semantics in store/cachekv/iterator when iterating a dirty cache.

## v0.45.11 - 2022-11-09

### Improvements

* [#13896](https://github.com/cosmos/cosmos-sdk/pull/13896) Queries on pruned height returns error instead of empty values.
* (deps) Bump Tendermint version to [v0.34.23](https://github.com/tendermint/tendermint/releases/tag/v0.34.23).
* (deps) Bump IAVL version to [v0.19.4](https://github.com/cosmos/iavl/releases/tag/v0.19.4).

### Bug Fixes

* [#13673](https://github.com/cosmos/cosmos-sdk/pull/13673) Fix `--dry-run` flag not working when using tx command.

### CLI Breaking Changes

* [#13656](https://github.com/cosmos/cosmos-sdk/pull/13660) Rename `server.FlagIAVLFastNode` to `server.FlagDisableIAVLFastNode` for clarity.

### API Breaking Changes

* [#13673](https://github.com/cosmos/cosmos-sdk/pull/13673) The `GetFromFields` function now takes `Context` as an argument and removes `genOnly`.

## [v0.45.10](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.10) - 2022-10-24

### Features

* (grpc) [#13485](https://github.com/cosmos/cosmos-sdk/pull/13485) Implement a new gRPC query, `/cosmos/base/node/v1beta1/config`, which provides operator configuration. Applications that wish to expose operator minimum gas prices via gRPC should have their application implement the `ApplicationQueryService` interface (see `SimApp#RegisterNodeService` as an example).
* [#13557](https://github.com/cosmos/cosmos-sdk/pull/#13557) - Add `GenSignedMockTx`. This can be used as workaround for #12437 revertion. `v0.46+` contains as well a `GenSignedMockTx` that behaves the same way.
* (x/auth) [#13612](https://github.com/cosmos/cosmos-sdk/pull/13612) Add `Query/ModuleAccountByName` endpoint for accessing the module account info by module name.

### Improvements

* [#13585](https://github.com/cosmos/cosmos-sdk/pull/13585) Bump Tendermint to `v0.34.22`.

### Bug Fixes

* [#13588](https://github.com/cosmos/cosmos-sdk/pull/13588) Fix regression in distrubtion.WithdrawDelegationRewards when rewards are zero.
* [#13564](https://github.com/cosmos/cosmos-sdk/pull/13564) - Fix `make proto-gen`.
* (server) [#13610](https://github.com/cosmos/cosmos-sdk/pull/13610) Read the pruning-keep-every field again.

## [v0.45.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.9) - 2022-10-14

ATTENTION:

This is a security release for the [Dragonberry security advisory](https://forum.cosmos.network/t/ibc-security-advisory-dragonberry/7702).

All users should upgrade immediately.

Users *must* add a replace directive in their go.mod for the new `ics23` package in the SDK:

```go
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```

### Features

* [#13435](https://github.com/cosmos/cosmos-sdk/pull/13435) Extend error context when a simulation fails.

### Improvements

* [#13369](https://github.com/cosmos/cosmos-sdk/pull/13369) Improve UX for `keyring.List` by returning all retrieved keys.
* [#13323](https://github.com/cosmos/cosmos-sdk/pull/13323) Ensure `withdraw_rewards` rewards are emitted from all actions that result in rewards being withdrawn.
* [#13321](https://github.com/cosmos/cosmos-sdk/pull/13321) Add flag to disable fast node migration and usage.
* (store) [#13326](https://github.com/cosmos/cosmos-sdk/pull/13326) Implementation of ADR-038 file StreamingService, backport #8664.
* (store) [#13540](https://github.com/cosmos/cosmos-sdk/pull/13540) Default fastnode migration to false to prevent suprises. Operators must enable it, unless they have it enabled already.

### API Breaking Changes

* (cli) [#13089](https://github.com/cosmos/cosmos-sdk/pull/13089) Fix rollback command don't actually delete multistore versions, added method `RollbackToVersion` to interface `CommitMultiStore` and added method `CommitMultiStore` to `Application` interface.

### Bug Fixes

* Implement dragonberry security patch.
    * For applying the patch please refer to the [RELEASE NOTES](./RELEASE_NOTES.md)
* (store) [#13459](https://github.com/cosmos/cosmos-sdk/pull/13459) Don't let state listener observe the uncommitted writes.

### Notes

Reverted #12437 due to API breaking changes.

## [v0.45.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.8) - 2022-08-25

### Improvements

* [#12981](https://github.com/cosmos/cosmos-sdk/pull/12981) Return proper error when parsing telemetry configuration.
* [#12885](https://github.com/cosmos/cosmos-sdk/pull/12885) Amortize cost of processing cache KV store.
* [#12970](https://github.com/cosmos/cosmos-sdk/pull/12970) Bump Tendermint to `v0.34.21` and IAVL to `v0.19.1`.
* [#12693](https://github.com/cosmos/cosmos-sdk/pull/12693) Make sure the order of each node is consistent when emitting proto events.

### Bug Fixes

* [#13046](https://github.com/cosmos/cosmos-sdk/pull/13046) Fix missing return statement in BaseApp.Query.

## [v0.45.7](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.7) - 2022-08-04

### Features

* (upgrade) [#12603](https://github.com/cosmos/cosmos-sdk/pull/12603) feat: Move AppModule.BeginBlock and AppModule.EndBlock to extension interfaces

### Improvements

* (events) [#12850](https://github.com/cosmos/cosmos-sdk/pull/12850) Add a new `fee_payer` attribute to the `tx` event that is emitted from the `DeductFeeDecorator` AnteHandler decorator.
* (x/params) [#12724](https://github.com/cosmos/cosmos-sdk/pull/12724) Add `GetParamSetIfExists` function to params `Subspace` to prevent panics on breaking changes.
* [#12668](https://github.com/cosmos/cosmos-sdk/pull/12668) Add `authz_msg_index` event attribute to message events emitted when executing via `MsgExec` through `x/authz`.
* [#12697](https://github.com/cosmos/cosmos-sdk/pull/12697) Upgrade IAVL to v0.19.0 with fast index and error propagation. NOTE: first start will take a while to propagate into new model.
    * Note: after upgrading to this version it may take up to 15 minutes to migrate from 0.17 to 0.19. This time is used to create the fast cache introduced into IAVL for performance
* [#12784](https://github.com/cosmos/cosmos-sdk/pull/12784) Upgrade Tendermint to 0.34.20.
* (x/bank) [#12674](https://github.com/cosmos/cosmos-sdk/pull/12674) Add convenience function `CreatePrefixedAccountStoreKey()` to construct key to access account's balance for a given denom.

### Bug Fixes

* (x/mint) [#12384](https://github.com/cosmos/cosmos-sdk/pull/12384) Ensure `GoalBonded` must be positive when performing `x/mint` parameter validation.
* (simapp) [#12437](https://github.com/cosmos/cosmos-sdk/pull/12437) fix the non-determinstic behavior in simulations caused by `GenTx` and check
empty coins slice before it is used to create `banktype.MsgSend`.
* (x/capability) [12818](https://github.com/cosmos/cosmos-sdk/pull/12818) Use fixed length hex for pointer at FwdCapabilityKey.

## [v0.45.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.6) - 2022-06-28

### Improvements

* (simapp) [#12314](https://github.com/cosmos/cosmos-sdk/pull/12314) Increase `DefaultGenTxGas` from `1000000` to `10000000`
* [#12371](https://github.com/cosmos/cosmos-sdk/pull/12371) Update min required Golang version to 1.18.

### Bug Fixes

* [#12317](https://github.com/cosmos/cosmos-sdk/pull/12317) Rename `edit-validator` command's `--moniker` flag to `--new-moniker`
* (x/upgrade) [#12264](https://github.com/cosmos/cosmos-sdk/pull/12264) Fix `GetLastCompleteUpgrade` to properly return the latest upgrade.
* (x/crisis) [#12208](https://github.com/cosmos/cosmos-sdk/pull/12208) Fix progress index of crisis invariant assertion logs.

### Features

* (query) [#12253](https://github.com/cosmos/cosmos-sdk/pull/12253) Add `GenericFilteredPaginate` to the `query` package to improve UX.

## [v0.45.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.5) - 2022-06-09

### Improvements

* (x/feegrant) [#11813](https://github.com/cosmos/cosmos-sdk/pull/11813) Fix pagination total count in `AllowancesByGranter` query.
* (errors) [#12002](https://github.com/cosmos/cosmos-sdk/pull/12002) Removed 'redacted' error message from defaultErrEncoder.
* (ante) [#12017](https://github.com/cosmos/cosmos-sdk/pull/12017) Index ante events for failed tx (backport #12013).
* [#12153](https://github.com/cosmos/cosmos-sdk/pull/12153) Add a new `NewSimulationManagerFromAppModules` constructor, to simplify simulation wiring.

### Bug Fixes

* [#11796](https://github.com/cosmos/cosmos-sdk/pull/11796) Handle EOF error case in `readLineFromBuf`, which allows successful reading of passphrases from STDIN.
* [#11772](https://github.com/cosmos/cosmos-sdk/pull/11772) Limit types.Dec length to avoid overflow.
* [#10947](https://github.com/cosmos/cosmos-sdk/pull/10947) Add `AllowancesByGranter` query to the feegrant module
* [#9639](https://github.com/cosmos/cosmos-sdk/pull/9639) Check store keys length before accessing them by making sure that `key` is of length `m+1` (for `key[n:m]`)
* [#11983](https://github.com/cosmos/cosmos-sdk/pull/11983) (x/feegrant, x/authz) rename grants query commands to `grants-by-grantee`, `grants-by-granter` cmds.

## Improvements

* [#11886](https://github.com/cosmos/cosmos-sdk/pull/11886) Improve error messages

## [v0.45.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.4) - 2022-04-25

### Bug Fixes

* [#11624](https://github.com/cosmos/cosmos-sdk/pull/11624) Handle the error returned from `NewNode` in the `server` package.
* [#11724](https://github.com/cosmos/cosmos-sdk/pull/11724) Fix data race issues with `api.Server`.

### Improvements

* (types) [#12201](https://github.com/cosmos/cosmos-sdk/pull/12201) Add `MustAccAddressFromBech32` util function
* [#11693](https://github.com/cosmos/cosmos-sdk/pull/11693) Add validation for gentx cmd.
* [#11686](https://github.com/cosmos/cosmos-sdk/pull/11686) Update the min required Golang version to `1.17`.
* (x/auth/vesting) [#11652](https://github.com/cosmos/cosmos-sdk/pull/11652) Add util functions for `Period(s)`

## [v0.45.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.3) - 2022-04-12

### Improvements

* [#11562](https://github.com/cosmos/cosmos-sdk/pull/11562) Updated Tendermint to v0.34.19; `unsafe-reset-all` command has been moved to the `tendermint` sub-command.

### Features

* (x/upgrade) [#11551](https://github.com/cosmos/cosmos-sdk/pull/11551) Update `ScheduleUpgrade` for chains to schedule an automated upgrade on `BeginBlock` without having to go though governance.

## [v0.45.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.2) - 2022-04-05

### Features

* (tx) [#11533](https://github.com/cosmos/cosmos-sdk/pull/11533) Register [`EIP191`](https://eips.ethereum.org/EIPS/eip-191) as an available `SignMode` for chains to use.
* [#11430](https://github.com/cosmos/cosmos-sdk/pull/11430) Introduce a new `grpc-only` flag, such that when enabled, will start the node in a query-only mode. Note, gRPC MUST be enabled with this flag.
* (x/bank) [#11417](https://github.com/cosmos/cosmos-sdk/pull/11417) Introduce a new `SpendableBalances` gRPC query that retrieves an account's total (paginated) spendable balances.
* (x/bank) [#10771](https://github.com/cosmos/cosmos-sdk/pull/10771) Add safety check on bank module perms to allow module-specific mint restrictions (e.g. only minting a certain denom).
* (x/bank) [#10771](https://github.com/cosmos/cosmos-sdk/pull/10771) Add `bank.BankKeeper.WithMintCoinsRestriction` function to restrict use of bank `MintCoins` usage. This function is not on the bank `Keeper` interface, so it's not API-breaking, but only additive on the keeper implementation.
* [#10944](https://github.com/cosmos/cosmos-sdk/pull/10944) `x/authz` add all grants by grantee query
* [#11124](https://github.com/cosmos/cosmos-sdk/pull/11124) Add `GetAllVersions` to application store
* (x/auth) [#10880](https://github.com/cosmos/cosmos-sdk/pull/10880) Added a new query to the tx query service that returns a block with transactions fully decoded.
* [#11314](https://github.com/cosmos/cosmos-sdk/pull/11314) Add state rollback command.

### Bug Fixes

* [#11354](https://github.com/cosmos/cosmos-sdk/pull/11355) Added missing pagination flag for `bank q total` query.
* [#11197](https://github.com/cosmos/cosmos-sdk/pull/11197) Signing with multisig now works with multisig address which is not in the keyring.
* (client) [#11283](https://github.com/cosmos/cosmos-sdk/issues/11283) Support multiple keys for tx simulation and setting automatic gas for txs.
* (store) [#11177](https://github.com/cosmos/cosmos-sdk/pull/11177) Update the prune `everything` strategy to store the last two heights.
* (store) [#11117](https://github.com/cosmos/cosmos-sdk/pull/11117) Fix data race in store trace component
* (x/authz) [#11252](https://github.com/cosmos/cosmos-sdk/pull/11252) Allow insufficient funds error for authz simulation
* (crypto) [#11298](https://github.com/cosmos/cosmos-sdk/pull/11298) Fix cgo secp signature verification and update libscep256k1 library.
* (crypto) [#12122](https://github.com/cosmos/cosmos-sdk/pull/12122) Fix keyring migration issue.

### Improvements

* [#9576](https://github.com/cosmos/cosmos-sdk/pull/9576) Add debug error message to query result when enabled
* (types) [#11200](https://github.com/cosmos/cosmos-sdk/pull/11200) Added `Min()` and `Max()` operations on sdk.Coins.
* [#11267](https://github.com/cosmos/cosmos-sdk/pull/11267) Add hooks to allow app modules to add things to state-sync (backport #10961).

## [v0.45.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.1) - 2022-02-03

### Bug Fixes

* (grpc) [#10985](https://github.com/cosmos/cosmos-sdk/pull/10992) The `/cosmos/tx/v1beta1/txs/{hash}` endpoint returns a 404 when a tx does not exist.
* [#10990](https://github.com/cosmos/cosmos-sdk/pull/10990) Fixes missing `iavl-cache-size` config parsing in `GetConfig` method.
* [#11222](https://github.com/cosmos/cosmos-sdk/pull/11222) reject query with block height in the future

### Improvements

* [#10407](https://github.com/cosmos/cosmos-sdk/pull/10407) Added validation to `x/upgrade` module's `BeginBlock` to check accidental binary downgrades
* [#10768](https://github.com/cosmos/cosmos-sdk/pull/10768) Extra logging in in-place store migrations.

## [v0.45.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.0) - 2022-01-18

### State Machine Breaking

* [#10833](https://github.com/cosmos/cosmos-sdk/pull/10833) fix reported tx gas used when block gas limit exceeded.
* (auth) [#10536](https://github.com/cosmos/cosmos-sdk/pull/10536]) Enable `SetSequence` for `ModuleAccount`.
* (store) [#10218](https://github.com/cosmos/cosmos-sdk/pull/10218) Charge gas even when there are no entries while seeking.
* (store) [#10247](https://github.com/cosmos/cosmos-sdk/pull/10247) Charge gas for the key length in gas meter.
* (x/gov) [#10740](https://github.com/cosmos/cosmos-sdk/pull/10740) Increase maximum proposal description size from 5k characters to 10k characters.
* [#10814](https://github.com/cosmos/cosmos-sdk/pull/10814) revert tx when block gas limit exceeded.

### API Breaking Changes

* [#10561](https://github.com/cosmos/cosmos-sdk/pull/10561) The `CommitMultiStore` interface contains a new `SetIAVLCacheSize` method
* [#10922](https://github.com/cosmos/cosmos-sdk/pull/10922), [/#10956](https://github.com/cosmos/cosmos-sdk/pull/10956) Deprecate key `server.Generate*` functions and move them to `testutil` and support custom mnemonics in in-process testing network. Moved `TestMnemonic` from `testutil` package to `testdata`.
* [#11049](https://github.com/cosmos/cosmos-sdk/pull/11049) Add custom tendermint config variables into root command. Allows App developers to set config.toml variables.

### Features

* [#10614](https://github.com/cosmos/cosmos-sdk/pull/10614) Support in-place migration ordering

### Improvements

* [#10486](https://github.com/cosmos/cosmos-sdk/pull/10486) store/cachekv's `Store.Write` conservatively
  looks up keys, but also uses the [map clearing idiom](https://bencher.orijtech.com/perfclinic/mapclearing/)
  to reduce the RAM usage, CPU time usage, and garbage collection pressure from clearing maps,
  instead of allocating new maps.
* (module) [#10711](https://github.com/cosmos/cosmos-sdk/pull/10711) Panic at startup if the app developer forgot to add modules in the `SetOrder{BeginBlocker, EndBlocker, InitGenesis, ExportGenesis}` functions. This means that all modules, even those who have empty implementations for those methods, need to be added to `SetOrder*`.
* (types) [#10076](https://github.com/cosmos/cosmos-sdk/pull/10076) Significantly speedup and lower allocations for `Coins.String()`.
* (auth) [#10022](https://github.com/cosmos/cosmos-sdk/pull/10022) `AuthKeeper` interface in `x/auth` now includes a function `HasAccount`.
* [#10393](https://github.com/cosmos/cosmos-sdk/pull/10393) Add `HasSupply` method to bank keeper to ensure that input denom actually exists on chain.

### Bug Fixes

* (std/codec) [/#10595](https://github.com/cosmos/cosmos-sdk/pull/10595) Add evidence to std/codec to be able to decode evidence in client interactions.
* (types) [#9627](https://github.com/cosmos/cosmos-sdk/pull/9627) Fix nil pointer panic on `NewBigIntFromInt`.
* [#10725](https://github.com/cosmos/cosmos-sdk/pull/10725) populate `ctx.ConsensusParams` for begin/end blockers.
* [#9829](https://github.com/cosmos/cosmos-sdk/pull/9829) Fixed Coin denom sorting not being checked during `Balance.Validate` check. Refactored the Validation logic to use `Coins.Validate` for `Balance.Coins`
* [#10061](https://github.com/cosmos/cosmos-sdk/pull/10061) and [#10515](https://github.com/cosmos/cosmos-sdk/pull/10515) Ensure that `LegacyAminoPubKey` struct correctly unmarshals from JSON

## [v0.44.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.8) - 2022-04-12

### Improvements

* [#11563](https://github.com/cosmos/cosmos-sdk/pull/11563) Updated Tendermint to v0.34.19; `unsafe-reset-all` command has been moved to the `tendermint` sub-command.

## [v0.44.7](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.7) - 2022-04-04

### Features

* (x/bank) [#10771](https://github.com/cosmos/cosmos-sdk/pull/10771) Add safety check on bank module perms to allow module-specific mint restrictions (e.g. only minting a certain denom).
* (x/bank) [#10771](https://github.com/cosmos/cosmos-sdk/pull/10771) Add `bank.BankKeeper.WithMintCoinsRestriction` function to restrict use of bank `MintCoins` usage. This function is not on the bank `Keeper` interface, so it's not API-breaking, but only additive on the keeper implementation.

### Bug Fixes

* [#11354](https://github.com/cosmos/cosmos-sdk/pull/11355) Added missing pagination flag for `bank q total` query.
* (store) [#11177](https://github.com/cosmos/cosmos-sdk/pull/11177) Update the prune `everything` strategy to store the last two heights.
* (store) [#11117](https://github.com/cosmos/cosmos-sdk/pull/11117) Fix data race in store trace component
* (x/authz) [#11252](https://github.com/cosmos/cosmos-sdk/pull/11252) Allow insufficient funds error for authz simulation

### Improvements

* [#9576](https://github.com/cosmos/cosmos-sdk/pull/9576) Add debug error message to query result when enabled

## [v0.44.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.6) - 2022-02-02

### Features

* [#11124](https://github.com/cosmos/cosmos-sdk/pull/11124) Add `GetAllVersions` to application store

### Bug Fixes

* (grpc) [#10985](https://github.com/cosmos/cosmos-sdk/pull/10992) The `/cosmos/tx/v1beta1/txs/{hash}` endpoint returns a 404 when a tx does not exist.
* (std/codec) [/#10595](https://github.com/cosmos/cosmos-sdk/pull/10595) Add evidence to std/codec to be able to decode evidence in client interactions.
* [#10725](https://github.com/cosmos/cosmos-sdk/pull/10725) populate `ctx.ConsensusParams` for begin/end blockers.
* [#10061](https://github.com/cosmos/cosmos-sdk/pull/10061) and [#10515](https://github.com/cosmos/cosmos-sdk/pull/10515) Ensure that `LegacyAminoPubKey` struct correctly unmarshals from JSON

### Improvements

* [#10823](https://github.com/cosmos/cosmos-sdk/pull/10823) updated ambiguous cli description for creating feegrant.

## [v0.44.5-patch](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.5-patch) - 2021-10-14

ATTENTION:

This is a security release for the [Dragonberry security advisory](https://forum.cosmos.network/t/ibc-security-advisory-dragonberry/7702).

All users should upgrade immediately.

Users *must* add a replace directive in their go.mod for the new `ics23` package in the SDK:

```go
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```

## [v0.44.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.5) - 2021-12-02

### Improvements

* (baseapp) [#10631](https://github.com/cosmos/cosmos-sdk/pull/10631) Emit ante events even for the failed txs.
* (store) [#10741](https://github.com/cosmos/cosmos-sdk/pull/10741) Significantly speedup iterator creation after delete heavy workloads. Significantly improves IBC migration times.

### Bug Fixes

* [#10648](https://github.com/cosmos/cosmos-sdk/pull/10648) Upgrade IAVL to 0.17.3 to solve race condition bug in IAVL.

## [v0.44.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.4) - 2021-11-25

### Improvements

* (types) [#10630](https://github.com/cosmos/cosmos-sdk/pull/10630) Add an `Events` field to the `TxResponse` type that captures *all* events emitted by a transaction, unlike `Logs` which only contains events emitted during message execution.
* (x/upgrade) [#10532](https://github.com/cosmos/cosmos-sdk/pull/10532) Add `keeper.DumpUpgradeInfoWithInfoToDisk` to include `Plan.Info` in the upgrade-info file.
* (store) [#10544](https://github.com/cosmos/cosmos-sdk/pull/10544) Use the new IAVL iterator structure which significantly improves iterator performance.

### Bug Fixes

* [#10827](https://github.com/cosmos/cosmos-sdk/pull/10827) Create query `Context` with requested block height
* [#10414](https://github.com/cosmos/cosmos-sdk/pull/10414) Use `sdk.GetConfig().GetFullBIP44Path()` instead `sdk.FullFundraiserPath` to generate key
* (bank) [#10394](https://github.com/cosmos/cosmos-sdk/pull/10394) Fix: query account balance by ibc denom.
* [\10608](https://github.com/cosmos/cosmos-sdk/pull/10608) Change the order of module migration by pushing x/auth to the end. Auth module depends on other modules and should be run last. We have updated the documentation to provide more details how to change module migration order. This is technically a breaking change, but only impacts updates between the upgrades with version change, hence migrating from the previous patch release doesn't cause new migration and doesn't break the state.
* [#10674](https://github.com/cosmos/cosmos-sdk/pull/10674) Fix issue with `Error.Wrap` and `Error.Wrapf` usage with `errors.Is`.

## [v0.44.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.3) - 2021-10-21

### Improvements

* [#10768](https://github.com/cosmos/cosmos-sdk/pull/10768) Added extra logging for tracking in-place store migrations
* [#10262](https://github.com/cosmos/cosmos-sdk/pull/10262) Remove unnecessary logging in `x/feegrant` simulation.
* [#10327](https://github.com/cosmos/cosmos-sdk/pull/10327) Add null guard for possible nil `Amount` in tx fee `Coins`
* [#10339](https://github.com/cosmos/cosmos-sdk/pull/10339) Improve performance of `removeZeroCoins` by only allocating memory when necessary
* [#10045](https://github.com/cosmos/cosmos-sdk/pull/10045) Revert [#8549](https://github.com/cosmos/cosmos-sdk/pull/8549). Do not route grpc queries through Tendermint.
* (deps) [#10375](https://github.com/cosmos/cosmos-sdk/pull/10375) Bump Tendermint to [v0.34.14](https://github.com/tendermint/tendermint/releases/tag/v0.34.14).
* [#10024](https://github.com/cosmos/cosmos-sdk/pull/10024) `store/cachekv` performance improvement by reduced growth factor for iterator ranging by using binary searches to find dirty items when unsorted key count >= 1024.

### Bug Fixes

* (client) [#10226](https://github.com/cosmos/cosmos-sdk/pull/10226) Fix --home flag parsing.
* (rosetta) [#10340](https://github.com/cosmos/cosmos-sdk/pull/10340) Use `GenesisChunked(ctx)` instead `Genesis(ctx)` to get genesis block height

## [v0.44.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.2) - 2021-10-12

Security Release. No breaking changes related to 0.44.x.

## [v0.44.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.1) - 2021-09-29

### Improvements

* (store) [#10040](https://github.com/cosmos/cosmos-sdk/pull/10040) Bump IAVL to v0.17.1 which includes performance improvements on a batch load.
* (types) [#10021](https://github.com/cosmos/cosmos-sdk/pull/10021) Speedup coins.AmountOf(), by removing many intermittent regex calls.
* [#10077](https://github.com/cosmos/cosmos-sdk/pull/10077) Remove telemetry on `GasKV` and `CacheKV` store Get/Set operations, significantly improving their performance.
* (store) [#10026](https://github.com/cosmos/cosmos-sdk/pull/10026) Improve CacheKVStore datastructures / algorithms, to no longer take O(N^2) time when interleaving iterators and insertions.

### Bug Fixes

* [#9969](https://github.com/cosmos/cosmos-sdk/pull/9969) fix: use keyring in config for add-genesis-account cmd.
* (x/genutil) [#10104](https://github.com/cosmos/cosmos-sdk/pull/10104) Ensure the `init` command reads the `--home` flag value correctly.
* (x/feegrant) [#10049](https://github.com/cosmos/cosmos-sdk/issues/10049) Fixed the error message when `period` or `period-limit` flag is not set on a feegrant grant transaction.

### Client Breaking Changes

* [#9879](https://github.com/cosmos/cosmos-sdk/pull/9879) Modify ABCI Queries to use `abci.QueryRequest` Height field if it is non-zero, otherwise continue using context height.

## [v0.44.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.0) - 2021-09-01

### Features

* [#9860](https://github.com/cosmos/cosmos-sdk/pull/9860) Emit transaction fee in ante handler fee decorator. The event type is `tx` and the attribute is `fee`.

### Improvements

* (deps) [#9956](https://github.com/cosmos/cosmos-sdk/pull/9956) Bump Tendermint to [v0.34.12](https://github.com/tendermint/tendermint/releases/tag/v0.34.12).

### Deprecated

* (x/upgrade) [#9906](https://github.com/cosmos/cosmos-sdk/pull/9906) Deprecate `UpgradeConsensusState` gRPC query since this functionality is only used for IBC, which now has its own [IBC replacement](https://github.com/cosmos/ibc-go/blob/2c880a22e9f9cc75f62b527ca94aa75ce1106001/proto/ibc/core/client/v1/query.proto#L54)

### Bug Fixes

* [#9965](https://github.com/cosmos/cosmos-sdk/pull/9965) Fixed `simd version` command output to report the right release tag.
* (x/upgrade) [#10189](https://github.com/cosmos/cosmos-sdk/issues/10189) Removed potential sources of non-determinism in upgrades.

### Client Breaking Changes

* [#10041](https://github.com/cosmos/cosmos-sdk/pull/10041) Remove broadcast & encode legacy REST endpoints. Please see the [REST Endpoints Migration guide](https://docs.cosmos.network/v0.45/migrations/rest.html) to migrate to the new REST endpoints.

## [v0.43.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.43.0) - 2021-08-10

### Features

* [#6711](https://github.com/cosmos/cosmos-sdk/pull/6711) Make integration test suites reusable by apps, tests are exported in each module's `client/testutil` package.
* [#8077](https://github.com/cosmos/cosmos-sdk/pull/8077) Added support for grpc-web, enabling browsers to communicate with a chain's gRPC server
* [#8965](https://github.com/cosmos/cosmos-sdk/pull/8965) cosmos reflection now provides more information on the application such as: deliverable msgs, sdk.Config info etc (still in alpha stage).
* [#8520](https://github.com/cosmos/cosmos-sdk/pull/8520) Add support for permanently locked vesting accounts.
* [#8559](https://github.com/cosmos/cosmos-sdk/pull/8559) Added Protobuf compatible secp256r1 ECDSA signatures.
* [#8786](https://github.com/cosmos/cosmos-sdk/pull/8786) Enabled secp256r1 in x/auth.
* (rosetta) [#8729](https://github.com/cosmos/cosmos-sdk/pull/8729) Data API fully supports balance tracking. Construction API can now construct any message supported by the application.
* [#8754](https://github.com/cosmos/cosmos-sdk/pull/8875) Added support for reverse iteration to pagination.
* (types) [#9079](https://github.com/cosmos/cosmos-sdk/issues/9079) Add `AddAmount`/`SubAmount` methods to `sdk.Coin`.
* [#9088](https://github.com/cosmos/cosmos-sdk/pull/9088) Added implementation to ADR-28 Derived Addresses.
* [#9133](https://github.com/cosmos/cosmos-sdk/pull/9133) Added hooks for governance actions.
* (x/staking) [#9214](https://github.com/cosmos/cosmos-sdk/pull/9214) Added `new_shares` attribute inside `EventTypeDelegate` event.
* [#9382](https://github.com/cosmos/cosmos-sdk/pull/9382) feat: add Dec.Float64() function.
* [#9457](https://github.com/cosmos/cosmos-sdk/pull/9457) Add amino support for x/authz and x/feegrant Msgs.
* [#9498](https://github.com/cosmos/cosmos-sdk/pull/9498) Added `Codec: codec.Codec` attribute to `client/Context` structure.
* [#9540](https://github.com/cosmos/cosmos-sdk/pull/9540) Add output flag for query txs command.
* (errors) [#8845](https://github.com/cosmos/cosmos-sdk/pull/8845) Add `Error.Wrap` handy method
* [#8518](https://github.com/cosmos/cosmos-sdk/pull/8518) Help users of multisig wallets debug signature issues.
* [#9573](https://github.com/cosmos/cosmos-sdk/pull/9573) ADR 040 implementation: New DB interface
* [#9952](https://github.com/cosmos/cosmos-sdk/pull/9952) ADR 040: Implement in-memory DB backend
* [#9848](https://github.com/cosmos/cosmos-sdk/pull/9848) ADR-040: Implement BadgerDB backend
* [#9851](https://github.com/cosmos/cosmos-sdk/pull/9851) ADR-040: Implement RocksDB backend
* [#10308](https://github.com/cosmos/cosmos-sdk/pull/10308) ADR-040: Implement DBConnection.Revert
* [#9892](https://github.com/cosmos/cosmos-sdk/pull/9892) ADR-040: KV Store with decoupled storage and state commitment

### Client Breaking Changes

* [#8363](https://github.com/cosmos/cosmos-sdk/pull/8363) Addresses no longer have a fixed 20-byte length. From the SDK modules' point of view, any 1-255 bytes-long byte array is a valid address.
* (crypto/ed25519) [#8690] Adopt zip1215 ed2559 verification rules.
* [#8849](https://github.com/cosmos/cosmos-sdk/pull/8849) Upgrade module no longer supports time based upgrades.
* [#7477](https://github.com/cosmos/cosmos-sdk/pull/7477) Changed Bech32 Public Key serialization in the client facing functionality (CLI, MsgServer, QueryServer):
    * updated the keyring display structure (it uses protobuf JSON serialization) - the output is more verbose.
    * Renamed `MarshalAny` and `UnmarshalAny` to `MarshalInterface` and `UnmarshalInterface` respectively. These functions must take an interface as parameter (not a concrete type nor `Any` object). Underneath they use `Any` wrapping for correct protobuf serialization.
    * CLI: removed `--text` flag from `show-node-id` command; the text format for public keys is not used any more - instead we use ProtoJSON.
* (store) [#8790](https://github.com/cosmos/cosmos-sdk/pull/8790) Reduce gas costs by 10x for transient store operations.
* [#9139](https://github.com/cosmos/cosmos-sdk/pull/9139) Querying events:
    * via `ServiceMsg` TypeURLs (e.g. `message.action='/cosmos.bank.v1beta1.Msg/Send'`) does not work anymore,
    * via legacy `msg.Type()` (e.g. `message.action='send'`) is being deprecated, new `Msg`s won't emit these events.
    * Please use concrete `Msg` TypeURLs instead (e.g. `message.action='/cosmos.bank.v1beta1.MsgSend'`).
* [#9859](https://github.com/cosmos/cosmos-sdk/pull/9859) The `default` pruning strategy now keeps the last 362880 blocks instead of 100. 362880 equates to roughly enough blocks to cover the entire unbonding period assuming a 21 day unbonding period and 5s block time.
* [#9785](https://github.com/cosmos/cosmos-sdk/issues/9785) Missing coin denomination in logs

### API Breaking Changes

* (keyring) [#8662](https://github.com/cosmos/cosmos-sdk/pull/8662) `NewMnemonic` now receives an additional `passphrase` argument to secure the key generated by the bip39 mnemonic.
* (x/bank) [#8473](https://github.com/cosmos/cosmos-sdk/pull/8473) Bank keeper does not expose unsafe balance changing methods such as `SetBalance`, `SetSupply` etc.
* (x/staking) [#8473](https://github.com/cosmos/cosmos-sdk/pull/8473) On genesis init, if non bonded pool and bonded pool balance, coming from the bank module, does not match what is saved in the staking state, the initialization will panic.
* (x/gov) [#8473](https://github.com/cosmos/cosmos-sdk/pull/8473) On genesis init, if the gov module account balance, coming from bank module state, does not match the one in gov module state, the initialization will panic.
* (x/distribution) [#8473](https://github.com/cosmos/cosmos-sdk/pull/8473) On genesis init, if the distribution module account balance, coming from bank module state, does not match the one in distribution module state, the initialization will panic.
* (client/keys) [#8500](https://github.com/cosmos/cosmos-sdk/pull/8500) `InfoImporter` interface is removed from legacy keybase.
* (x/staking) [#8505](https://github.com/cosmos/cosmos-sdk/pull/8505) `sdk.PowerReduction` has been renamed to `sdk.DefaultPowerReduction`, and most staking functions relying on power reduction take a new function argument, instead of relying on that global variable.
* [#8629](https://github.com/cosmos/cosmos-sdk/pull/8629) Deprecated `SetFullFundraiserPath` from `Config` in favor of `SetPurpose` and `SetCoinType`.
* (x/upgrade) [#8673](https://github.com/cosmos/cosmos-sdk/pull/8673) Remove IBC logic from x/upgrade. Deprecates IBC fields in an Upgrade Plan, an error will be thrown if they are set. IBC upgrade logic moved to 02-client and an IBC UpgradeProposal is added.
* (x/bank) [#8517](https://github.com/cosmos/cosmos-sdk/pull/8517) `SupplyI` interface and `Supply` are removed and uses `sdk.Coins` for supply tracking
* (x/upgrade) [#8743](https://github.com/cosmos/cosmos-sdk/pull/8743) `UpgradeHandler` includes a new argument `VersionMap` which helps facilitate in-place migrations.
* (x/auth) [#8129](https://github.com/cosmos/cosmos-sdk/pull/8828) Updated `SigVerifiableTx.GetPubKeys` method signature to return error.
* (x/upgrade) [\7487](https://github.com/cosmos/cosmos-sdk/pull/8897) Upgrade `Keeper` takes new argument `ProtocolVersionSetter` which implements setting a protocol version on baseapp.
* (baseapp) [\7487](https://github.com/cosmos/cosmos-sdk/pull/8897) BaseApp's fields appVersion and version were swapped to match Tendermint's fields.
* [#8682](https://github.com/cosmos/cosmos-sdk/pull/8682) `ante.NewAnteHandler` updated to receive all positional params as `ante.HandlerOptions` struct. If required fields aren't set, throws error accordingly.
* (x/staking/types) [#7447](https://github.com/cosmos/cosmos-sdk/issues/7447) Remove bech32 PubKey support:
    * `ValidatorI` interface update: `GetConsPubKey` renamed to `TmConsPubKey` (this is to clarify the return type: consensus public key must be a tendermint key); `TmConsPubKey`, `GetConsAddr` methods return error.
    * `Validator` updated according to the `ValidatorI` changes described above.
    * `ToTmValidator` function: added `error` to return values.
    * `Validator.ConsensusPubkey` type changed from `string` to `codectypes.Any`.
    * `MsgCreateValidator.Pubkey` type changed from `string` to `codectypes.Any`.
* (client) [#8926](https://github.com/cosmos/cosmos-sdk/pull/8926) `client/tx.PrepareFactory` has been converted to a private function, as it's only used internally.
* (auth/tx) [#8926](https://github.com/cosmos/cosmos-sdk/pull/8926) The `ProtoTxProvider` interface used as a workaround for transaction simulation has been removed.
* (x/bank) [#8798](https://github.com/cosmos/cosmos-sdk/pull/8798) `GetTotalSupply` is removed in favour of `GetPaginatedTotalSupply`
* (keyring) [#8739](https://github.com/cosmos/cosmos-sdk/pull/8739) Rename InfoImporter -> LegacyInfoImporter.
* (x/bank/types) [#9061](https://github.com/cosmos/cosmos-sdk/pull/9061) `AddressFromBalancesStore` now returns an error for invalid key instead of panic.
* (x/auth) [#9144](https://github.com/cosmos/cosmos-sdk/pull/9144) The `NewTxTimeoutHeightDecorator` antehandler has been converted from a struct to a function.
* (codec) [#9226](https://github.com/cosmos/cosmos-sdk/pull/9226) Rename codec interfaces and methods, to follow a general Go interfaces:
    * `codec.Marshaler` → `codec.Codec` (this defines objects which serialize other objects)
    * `codec.BinaryMarshaler` → `codec.BinaryCodec`
    * `codec.JSONMarshaler` → `codec.JSONCodec`
    * Removed `BinaryBare` suffix from `BinaryCodec` methods (`MarshalBinaryBare`, `UnmarshalBinaryBare`, ...)
    * Removed `Binary` infix from `BinaryCodec` methods (`MarshalBinaryLengthPrefixed`, `UnmarshalBinaryLengthPrefixed`, ...)
* [#9139](https://github.com/cosmos/cosmos-sdk/pull/9139) `ServiceMsg` TypeURLs (e.g. `/cosmos.bank.v1beta1.Msg/Send`) have been removed, as they don't comply to the Probobuf `Any` spec. Please use `Msg` type TypeURLs (e.g. `/cosmos.bank.v1beta1.MsgSend`). This has multiple consequences:
    * The `sdk.ServiceMsg` struct has been removed.
    * `sdk.Msg` now only contains `ValidateBasic` and `GetSigners` methods. The remaining methods `GetSignBytes`, `Route` and `Type` are moved to `legacytx.LegacyMsg`.
    * The `RegisterCustomTypeURL` function and the `cosmos.base.v1beta1.ServiceMsg` interface have been removed from the interface registry.
* (codec) [#9251](https://github.com/cosmos/cosmos-sdk/pull/9251) Rename `clientCtx.JSONMarshaler` to `clientCtx.JSONCodec` as per #9226.
* (x/bank) [#9271](https://github.com/cosmos/cosmos-sdk/pull/9271) SendEnabledCoin(s) renamed to IsSendEnabledCoin(s) to better reflect its functionality.
* (x/bank) [#9550](https://github.com/cosmos/cosmos-sdk/pull/9550) `server.InterceptConfigsPreRunHandler` now takes 2 additional arguments: customAppConfigTemplate and customAppConfig. If you don't need to customize these, simply put `""` and `nil`.
* [#8245](https://github.com/cosmos/cosmos-sdk/pull/8245) Removed `simapp.MakeCodecs` and use `simapp.MakeTestEncodingConfig` instead.
* (x/capability) [#9836](https://github.com/cosmos/cosmos-sdk/pull/9836) Removed `InitializeAndSeal(ctx sdk.Context)` and replaced with `Seal()`. App must add x/capability module to the begin blockers which will assure that the x/capability keeper is properly initialized. The x/capability begin blocker must be run before any other module which uses x/capability.

### State Machine Breaking

* (x/{bank,distrib,gov,slashing,staking}) [#8363](https://github.com/cosmos/cosmos-sdk/issues/8363) Store keys have been modified to allow for variable-length addresses.
* (x/evidence) [#8502](https://github.com/cosmos/cosmos-sdk/pull/8502) `HandleEquivocationEvidence` persists the evidence to state.
* (x/gov) [#7733](https://github.com/cosmos/cosmos-sdk/pull/7733) ADR 037 Implementation: Governance Split Votes, use `MsgWeightedVote` to send a split vote. Sending a regular `MsgVote` will convert the underlying vote option into a weighted vote with weight 1.
* (x/bank) [#8656](https://github.com/cosmos/cosmos-sdk/pull/8656) balance and supply are now correctly tracked via `coin_spent`, `coin_received`, `coinbase` and `burn` events.
* (x/bank) [#8517](https://github.com/cosmos/cosmos-sdk/pull/8517) Supply is now stored and tracked as `sdk.Coins`
* (x/bank) [#9051](https://github.com/cosmos/cosmos-sdk/pull/9051) Supply value is stored as `sdk.Int` rather than `string`.

### CLI Breaking Changes

* [#8880](https://github.com/cosmos/cosmos-sdk/pull/8880) The CLI `simd migrate v0.40 ...` command has been renamed to `simd migrate v0.42`.
* [#8628](https://github.com/cosmos/cosmos-sdk/issues/8628) Commands no longer print outputs using `stderr` by default
* [#9134](https://github.com/cosmos/cosmos-sdk/pull/9134) Renamed the CLI flag `--memo` to `--note`.
* [#9291](https://github.com/cosmos/cosmos-sdk/pull/9291) Migration scripts prior to v0.38 have been removed from the CLI `migrate` command. The oldest supported migration is v0.39->v0.42.
* [#9371](https://github.com/cosmos/cosmos-sdk/pull/9371) Non-zero default fees/Server will error if there's an empty value for min-gas-price in app.toml
* [#9827](https://github.com/cosmos/cosmos-sdk/pull/9827) Ensure input parity of validator public key input between `tx staking create-validator` and `gentx`.
* [#9621](https://github.com/cosmos/cosmos-sdk/pull/9621) Rollback [#9371](https://github.com/cosmos/cosmos-sdk/pull/9371) and log warning if there's an empty value for min-gas-price in app.toml

### Improvements

* (store) [#8012](https://github.com/cosmos/cosmos-sdk/pull/8012) Implementation of ADR-038 WriteListener and listen.KVStore
* (x/bank) [#8614](https://github.com/cosmos/cosmos-sdk/issues/8614) Add `Name` and `Symbol` fields to denom metadata
* (x/auth) [#8522](https://github.com/cosmos/cosmos-sdk/pull/8522) Allow to query all stored accounts
* (crypto/types) [#8600](https://github.com/cosmos/cosmos-sdk/pull/8600) `CompactBitArray`: optimize the `NumTrueBitsBefore` method and add an `Equal` method.
* (x/upgrade) [#8743](https://github.com/cosmos/cosmos-sdk/pull/8743) Add tracking module versions as per ADR-041
* (types) [#8962](https://github.com/cosmos/cosmos-sdk/issues/8962) Add `Abs()` method to `sdk.Int`.
* (x/bank) [#8950](https://github.com/cosmos/cosmos-sdk/pull/8950) Improve efficiency on supply updates.
* (store) [#8811](https://github.com/cosmos/cosmos-sdk/pull/8811) store/cachekv: use typed `types/kv.List` instead of `container/list.List`. The change brings time spent on the time assertion cummulatively to 580ms down from 6.88s.
* (keyring) [#8826](https://github.com/cosmos/cosmos-sdk/pull/8826) add trust to macOS Keychain for calling apps by default, avoiding repeating keychain popups that appears when dealing with keyring (key add, list, ...) operations.
* (makefile) [#7933](https://github.com/cosmos/cosmos-sdk/issues/7933) Use Docker to generate swagger files.
* (crypto/types) [#9196](https://github.com/cosmos/cosmos-sdk/pull/9196) Fix negative index accesses in CompactUnmarshal,GetIndex,SetIndex
* (makefile) [#9192](https://github.com/cosmos/cosmos-sdk/pull/9192) Reuse proto containers in proto related jobs.
* [#9205](https://github.com/cosmos/cosmos-sdk/pull/9205) Improve readability in `abci` handleQueryP2P
* [#9231](https://github.com/cosmos/cosmos-sdk/pull/9231) Remove redundant staking errors.
* [#9314](https://github.com/cosmos/cosmos-sdk/pull/9314) Update Rosetta SDK to upstream's latest release.
* (gRPC-Web) [#9493](https://github.com/cosmos/cosmos-sdk/pull/9493) Add `EnableUnsafeCORS` flag to grpc-web config.
* (x/params) [#9481](https://github.com/cosmos/cosmos-sdk/issues/9481) Speedup simulator for parameter change proposals.
* (x/staking) [#9423](https://github.com/cosmos/cosmos-sdk/pull/9423) Staking delegations now returns empty list instead of rpc error when no records found.
* (x/auth) [#9553](https://github.com/cosmos/cosmos-sdk/pull/9553) The `--multisig` flag now accepts both a name and address.
* [#8549](https://github.com/cosmos/cosmos-sdk/pull/8549) Make gRPC requests go through tendermint Query
* [#8093](https://github.com/cosmos/cosmos-sdk/pull/8093) Limit usage of context.background.
* [#8460](https://github.com/cosmos/cosmos-sdk/pull/8460) Ensure b.ReportAllocs() in all the benchmarks
* [#8461](https://github.com/cosmos/cosmos-sdk/pull/8461) Fix upgrade tx commands not showing up in CLI

### Bug Fixes

* (gRPC) [#8945](https://github.com/cosmos/cosmos-sdk/pull/8945) gRPC reflection now works correctly.
* (keyring) [#8635](https://github.com/cosmos/cosmos-sdk/issues/8635) Remove hardcoded default passphrase value on `NewMnemonic`
* (x/bank) [#8434](https://github.com/cosmos/cosmos-sdk/pull/8434) Fix legacy REST API `GET /bank/total` and `GET /bank/total/{denom}` in swagger
* (x/slashing) [#8427](https://github.com/cosmos/cosmos-sdk/pull/8427) Fix query signing infos command
* (x/bank/types) [#9112](https://github.com/cosmos/cosmos-sdk/pull/9112) fix AddressFromBalancesStore address length overflow
* (x/bank) [#9229](https://github.com/cosmos/cosmos-sdk/pull/9229) Now zero coin balances cannot be added to balances & supply stores. If any denom becomes zero corresponding key gets deleted from store. State migration: [#9664](https://github.com/cosmos/cosmos-sdk/pull/9664).
* [#9363](https://github.com/cosmos/cosmos-sdk/pull/9363) Check store key uniqueness in app wiring.
* [#9460](https://github.com/cosmos/cosmos-sdk/pull/9460) Fix lint error in `MigratePrefixAddress`.
* [#9480](https://github.com/cosmos/cosmos-sdk/pull/9480) Fix added keys when using `--dry-run`.
* (types) [#9511](https://github.com/cosmos/cosmos-sdk/pull/9511) Change `maxBitLen` of `sdk.Int` and `sdk.Dec` to handle max ERC20 value.
* [#9454](https://github.com/cosmos/cosmos-sdk/pull/9454) Fix testnet command with --node-dir-prefix accepts `-` and change `node-dir-prefix token` to `testtoken`.
* (keyring) [#9562](https://github.com/cosmos/cosmos-sdk/pull/9563) fix keyring kwallet backend when using with empty wallet.
* (keyring) [#9583](https://github.com/cosmos/cosmos-sdk/pull/9583) Fix correct population of legacy `Vote.Option` field for votes with 1 VoteOption of weight 1.
* (x/distinction) [#8918](https://github.com/cosmos/cosmos-sdk/pull/8918) Fix module's parameters validation.
* (x/gov/types) [#8586](https://github.com/cosmos/cosmos-sdk/pull/8586) Fix bug caused by NewProposal that unnecessarily creates a Proposal object that’s discarded on any error.
* [#8580](https://github.com/cosmos/cosmos-sdk/pull/8580) Use more cheaper method from the math/big package that provides a way to trivially check if a value is zero with .BitLen() == 0
* [#8567](https://github.com/cosmos/cosmos-sdk/pull/8567) Fix bug by introducing pagination to GetValidatorSetByHeight response
* (x/bank) [#8531](https://github.com/cosmos/cosmos-sdk/pull/8531) Fix bug caused by ignoring errors returned by Balance.GetAddress()
* (server) [#8399](https://github.com/cosmos/cosmos-sdk/pull/8399) fix gRPC-web flag default value
* [#8282](https://github.com/cosmos/cosmos-sdk/pull/8282) fix zero time checks
* (cli) [#9593](https://github.com/cosmos/cosmos-sdk/pull/9593) Check if chain-id is blank before verifying signatures in multisign and error.
* [#9720](https://github.com/cosmos/cosmos-sdk/pull/9720) Feegrant grant cli granter now accepts key name as well as address in general and accepts only address in --generate-only mode
* [#9793](https://github.com/cosmos/cosmos-sdk/pull/9793) Fixed ECDSA/secp256r1 transaction malleability.
* (server) [#9704](https://github.com/cosmos/cosmos-sdk/pull/9704) Start GRPCWebServer in goroutine, avoid blocking other services from starting.
* (bank) [#9687](https://github.com/cosmos/cosmos-sdk/issues/9687) fixes [#9159](https://github.com/cosmos/cosmos-sdk/issues/9159). Added migration to prune balances with zero coins.

### Deprecated

* (grpc) [#8926](https://github.com/cosmos/cosmos-sdk/pull/8926) The `tx` field in `SimulateRequest` has been deprecated, prefer to pass `tx_bytes` instead.
* (sdk types) [#9498](https://github.com/cosmos/cosmos-sdk/pull/9498) `clientContext.JSONCodec` will be removed in the next version. use `clientContext.Codec` instead.

## [v0.42.10](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.10) - 2021-09-28

### Improvements

* (store) [#10026](https://github.com/cosmos/cosmos-sdk/pull/10026) Improve CacheKVStore datastructures / algorithms, to no longer take O(N^2) time when interleaving iterators and insertions.
* (store) [#10040](https://github.com/cosmos/cosmos-sdk/pull/10040) Bump IAVL to v0.17.1 which includes performance improvements on a batch load.
* [#10211](https://github.com/cosmos/cosmos-sdk/pull/10211) Backport of the mechanism to reject redundant IBC transactions from [ibc-go \#235](https://github.com/cosmos/ibc-go/pull/235).

### Bug Fixes

* [#9969](https://github.com/cosmos/cosmos-sdk/pull/9969) fix: use keyring in config for add-genesis-account cmd.

### Client Breaking Changes

* [#9879](https://github.com/cosmos/cosmos-sdk/pull/9879) Modify ABCI Queries to use `abci.QueryRequest` Height field if it is non-zero, otherwise continue using context height.

### API Breaking Changes

* [#10077](https://github.com/cosmos/cosmos-sdk/pull/10077) Remove telemetry on `GasKV` and `CacheKV` store Get/Set operations, significantly improving their performance.

## [v0.42.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.9) - 2021-08-04

### Bug Fixes

* [#9835](https://github.com/cosmos/cosmos-sdk/pull/9835) Moved capability initialization logic to BeginBlocker to fix nondeterminsim issue mentioned in [#9800](https://github.com/cosmos/cosmos-sdk/issues/9800). Applications must now include the capability module in their BeginBlocker order before any module that uses capabilities gets run.
* [#9201](https://github.com/cosmos/cosmos-sdk/pull/9201) Fixed `<app> init --recover` flag.

### API Breaking Changes

* [#9835](https://github.com/cosmos/cosmos-sdk/pull/9835) The `InitializeAndSeal` API has not changed, however it no longer initializes the in-memory state. `InitMemStore` has been introduced to serve this function, which will be called either in `InitChain` or `BeginBlock` (whichever is first after app start). Nodes may run this version on a network running 0.42.x, however, they must update their app.go files to include the capability module in their begin blockers.

### Client Breaking Changes

* [#9781](https://github.com/cosmos/cosmos-sdk/pull/9781) Improve`withdraw-all-rewards` UX when broadcast mode `async` or `async` is used.

## [v0.42.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.8) - 2021-07-30

### Features

* [#9750](https://github.com/cosmos/cosmos-sdk/pull/9750) Emit events for tx signature and sequence, so clients can now query txs by signature (`tx.signature='<base64_sig>'`) or by address and sequence combo (`tx.acc_seq='<addr>/<seq>'`).

### Improvements

* (cli) [#9717](https://github.com/cosmos/cosmos-sdk/pull/9717) Added CLI flag `--output json/text` to `tx` cli commands.

### Bug Fixes

* [#9766](https://github.com/cosmos/cosmos-sdk/pull/9766) Fix hardcoded ledger signing algorithm on `keys add` command.

## [v0.42.7](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.7) - 2021-07-09

### Improvements

* (baseapp) [#9578](https://github.com/cosmos/cosmos-sdk/pull/9578) Return `Baseapp`'s `trace` value for logging error stack traces.

### Bug Fixes

* (x/ibc) [#9640](https://github.com/cosmos/cosmos-sdk/pull/9640) Fix IBC Transfer Ack Success event as it was initially emitting opposite value.
* [#9645](https://github.com/cosmos/cosmos-sdk/pull/9645) Use correct Prometheus format for metric labels.
* [#9299](https://github.com/cosmos/cosmos-sdk/pull/9299) Fix `[appd] keys parse cosmos1...` freezing.
* (keyring) [#9563](https://github.com/cosmos/cosmos-sdk/pull/9563) fix keyring kwallet backend when using with empty wallet.
* (x/capability) [#9392](https://github.com/cosmos/cosmos-sdk/pull/9392) initialization fix, which fixes the consensus error when using statesync.

## [v0.42.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.6) - 2021-06-18

### Improvements

* [#9428](https://github.com/cosmos/cosmos-sdk/pull/9428) Optimize bank InitGenesis. Added `k.initBalances`.
* [#9429](https://github.com/cosmos/cosmos-sdk/pull/9429) Add `cosmos_sdk_version` to node_info
* [#9541](https://github.com/cosmos/cosmos-sdk/pull/9541) Bump tendermint dependency to v0.34.11.

### Bug Fixes

* [#9385](https://github.com/cosmos/cosmos-sdk/pull/9385) Fix IBC `query ibc client header` cli command. Support historical queries for query header/node-state commands.
* [#9401](https://github.com/cosmos/cosmos-sdk/pull/9401) Fixes incorrect export of IBC identifier sequences. Previously, the next identifier sequence for clients/connections/channels was not set during genesis export. This resulted in the next identifiers being generated on the new chain to reuse old identifiers (the sequences began again from 0).
* [#9408](https://github.com/cosmos/cosmos-sdk/pull/9408) Update simapp to use correct default broadcast mode.
* [#9513](https://github.com/cosmos/cosmos-sdk/pull/9513) Fixes testnet CLI command. Testnet now updates the supply in genesis. Previously, when using add-genesis-account and testnet together, inconsistent genesis files would be produced, as only add-genesis-account was updating the supply.
* (x/gov) [#8813](https://github.com/cosmos/cosmos-sdk/pull/8813) fix `GET /cosmos/gov/v1beta1/proposals/{proposal_id}/deposits` to include initial deposit

### Features

* [#9383](https://github.com/cosmos/cosmos-sdk/pull/9383) New CLI command `query ibc-transfer escrow-address <port> <channel id>` to get the escrow address for a channel; can be used to then query balance of escrowed tokens
* (baseapp, types) [#9390](https://github.com/cosmos/cosmos-sdk/pull/9390) Add current block header hash to `Context`
* (store) [#9403](https://github.com/cosmos/cosmos-sdk/pull/9403) Add `RefundGas` function to `GasMeter` interface

## [v0.42.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.5) - 2021-05-18

### Bug Fixes

* [#9514](https://github.com/cosmos/cosmos-sdk/issues/9514) Fix panic when retrieving the `BlockGasMeter` on `(Re)CheckTx` mode.
* [#9235](https://github.com/cosmos/cosmos-sdk/pull/9235) CreateMembershipProof/CreateNonMembershipProof now returns an error
  if input key is empty, or input data contains empty key.
* [#9108](https://github.com/cosmos/cosmos-sdk/pull/9108) Fixed the bug with querying multisig account, which is not showing threshold and public_keys.
* [#9345](https://github.com/cosmos/cosmos-sdk/pull/9345) Fix ARM support.
* [#9040](https://github.com/cosmos/cosmos-sdk/pull/9040) Fix ENV variables binding to CLI flags for client config.

### Features

* [#8953](https://github.com/cosmos/cosmos-sdk/pull/8953) Add the `config` CLI subcommand back to the SDK, which saves client-side configuration in a `client.toml` file.

## [v0.42.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.4) - 2021-04-08

### Client Breaking Changes

* [#9026](https://github.com/cosmos/cosmos-sdk/pull/9026) By default, the `tx sign` and `tx sign-batch` CLI commands use SIGN_MODE_DIRECT to sign transactions for local pubkeys. For multisigs and ledger keys, the default LEGACY_AMINO_JSON is used.

### Bug Fixes

* (gRPC) [#9015](https://github.com/cosmos/cosmos-sdk/pull/9015) Fix invalid status code when accessing gRPC endpoints.
* [#9026](https://github.com/cosmos/cosmos-sdk/pull/9026) Fixed the bug that caused the `gentx` command to fail for Ledger keys.

### Improvements

* [#9081](https://github.com/cosmos/cosmos-sdk/pull/9081) Upgrade Tendermint to v0.34.9 that includes a security issue fix for Tendermint light clients.

## [v0.42.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.3) - 2021-03-24

This release fixes a security vulnerability identified in x/bank.

## [v0.42.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.2) - 2021-03-19

### Improvements

* (grpc) [#8815](https://github.com/cosmos/cosmos-sdk/pull/8815) Add orderBy parameter to `TxsByEvents` endpoint.
* (cli) [#8826](https://github.com/cosmos/cosmos-sdk/pull/8826) Add trust to macOS Keychain for caller app by default.
* (store) [#8811](https://github.com/cosmos/cosmos-sdk/pull/8811) store/cachekv: use typed types/kv.List instead of container/list.List

### Bug Fixes

* (crypto) [#8841](https://github.com/cosmos/cosmos-sdk/pull/8841) Fix legacy multisig amino marshaling, allowing migrations to work between v0.39 and v0.40+.
* (cli tx) [\8873](https://github.com/cosmos/cosmos-sdk/pull/8873) add missing `--output-document` option to `app tx multisign-batch`.

## [v0.42.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.1) - 2021-03-10

This release fixes security vulnerability identified in the simapp.

## [v0.42.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.0) - 2021-03-08

**IMPORTANT**: This release contains an important security fix for all non Cosmos Hub chains running Stargate version of the Cosmos SDK (>0.40). Non-hub chains should not be using any version of the SDK in the v0.40.x or v0.41.x release series. See [#8461](https://github.com/cosmos/cosmos-sdk/pull/8461) for more details.

### Improvements

* (x/ibc) [#8624](https://github.com/cosmos/cosmos-sdk/pull/8624) Emit full header in IBC UpdateClient message.
* (x/crisis) [#8621](https://github.com/cosmos/cosmos-sdk/issues/8621) crisis invariants names now print to loggers.

### Bug fixes

* (x/evidence) [#8461](https://github.com/cosmos/cosmos-sdk/pull/8461) Fix bech32 prefix in evidence validator address conversion
* (x/gov) [#8806](https://github.com/cosmos/cosmos-sdk/issues/8806) Fix q gov proposals command's mishandling of the --status parameter's values.

## [v0.41.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.3) - 2021-03-02

**IMPORTANT**: Due to a bug in the v0.41.x series with how evidence handles validator consensus addresses #8461, SDK based chains that are not using the default bech32 prefix (cosmos, aka all chains except for t
he Cosmos Hub) should not use this release or any release in the v0.41.x series. Please see #8668 for tracking & timeline for the v0.42.0 release, which will include a fix for this issue.

### Features

* [#7787](https://github.com/cosmos/cosmos-sdk/pull/7787) Add multisign-batch command.

### Bug fixes

* [#8730](https://github.com/cosmos/cosmos-sdk/pull/8730) Allow REST endpoint to query txs with multisig addresses.
* [#8680](https://github.com/cosmos/cosmos-sdk/issues/8680) Fix missing timestamp in GetTxsEvent response [#8732](https://github.com/cosmos/cosmos-sdk/pull/8732).
* [#8681](https://github.com/cosmos/cosmos-sdk/issues/8681) Fix missing error message when calling GetTxsEvent [#8732](https://github.com/cosmos/cosmos-sdk/pull/8732)
* (server) [#8641](https://github.com/cosmos/cosmos-sdk/pull/8641) Fix Tendermint and application configuration reading from file
* (client/keys) [#8639](https://github.com/cosmos/cosmos-sdk/pull/8639) Fix keys migrate for mulitisig, offline, and ledger keys. The migrate command now takes a positional old_home_dir argument.

### Improvements

* (store/cachekv), (x/bank/types) [#8719](https://github.com/cosmos/cosmos-sdk/pull/8719) algorithmically fix pathologically slow code
* [#8701](https://github.com/cosmos/cosmos-sdk/pull/8701) Upgrade tendermint v0.34.8.
* [#8714](https://github.com/cosmos/cosmos-sdk/pull/8714) Allow accounts to have a balance of 0 at genesis.

## [v0.41.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.3) - 2021-02-18

### Bug Fixes

* [#8617](https://github.com/cosmos/cosmos-sdk/pull/8617) Fix build failures caused by a small API breakage introduced in tendermint v0.34.7.

## [v0.41.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.2) - 2021-02-18

### Improvements

* Bump tendermint dependency to v0.34.7.

## [v0.41.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.1) - 2021-02-17

### Bug Fixes

* (grpc) [#8549](https://github.com/cosmos/cosmos-sdk/pull/8549) Make gRPC requests go through ABCI and disallow concurrency.
* (x/staking) [#8546](https://github.com/cosmos/cosmos-sdk/pull/8546) Fix caching bug where concurrent calls to GetValidator could cause a node to crash
* (server) [#8481](https://github.com/cosmos/cosmos-sdk/pull/8481) Don't create files when running `{appd} tendermint show-*` subcommands.
* (client/keys) [#8436](https://github.com/cosmos/cosmos-sdk/pull/8436) Fix keybase->keyring keys migration.
* (crypto/hd) [#8607](https://github.com/cosmos/cosmos-sdk/pull/8607) Make DerivePrivateKeyForPath error and not panic on trailing slashes.

### Improvements

* (x/ibc) [#8458](https://github.com/cosmos/cosmos-sdk/pull/8458) Add `packet_connection` attribute to ibc events to enable relayer filtering
* [#8396](https://github.com/cosmos/cosmos-sdk/pull/8396) Add support for ARM platform
* (x/bank) [#8479](https://github.com/cosmos/cosmos-sdk/pull/8479) Aditional client denom metadata validation for `base` and `display` denoms.
* (codec/types) [#8605](https://github.com/cosmos/cosmos-sdk/pull/8605) Avoid unnecessary allocations for NewAnyWithCustomTypeURL on error.

## [v0.41.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.0) - 2021-01-26

### State Machine Breaking

* (x/ibc) [#8266](https://github.com/cosmos/cosmos-sdk/issues/8266) Add amino JSON support for IBC MsgTransfer in order to support Ledger text signing transfer transactions.
* (x/ibc) [#8404](https://github.com/cosmos/cosmos-sdk/pull/8404) Reorder IBC `ChanOpenAck` and `ChanOpenConfirm` handler execution to perform core handler first, followed by application callbacks.

### Bug Fixes

* (simapp) [#8418](https://github.com/cosmos/cosmos-sdk/pull/8418) Add balance coin to supply when adding a new genesis account
* (x/bank) [#8417](https://github.com/cosmos/cosmos-sdk/pull/8417) Validate balances and coin denom metadata on genesis

## [v0.40.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.40.1) - 2021-01-19

### Improvements

* (x/bank) [#8302](https://github.com/cosmos/cosmos-sdk/issues/8302) Add gRPC and CLI queries for client denomination metadata.
* (tendermint) Bump Tendermint version to [v0.34.3](https://github.com/tendermint/tendermint/releases/tag/v0.34.3).

### Bug Fixes

* [#8085](https://github.com/cosmos/cosmos-sdk/pull/8058) fix zero time checks
* [#8280](https://github.com/cosmos/cosmos-sdk/pull/8280) fix GET /upgrade/current query
* (x/auth) [#8287](https://github.com/cosmos/cosmos-sdk/pull/8287) Fix `tx sign --signature-only` to return correct sequence value in signature.
* (build) [\8300](https://github.com/cosmos/cosmos-sdk/pull/8300), [\8301](https://github.com/cosmos/cosmos-sdk/pull/8301) Fix reproducible builds
* (types/errors) [#8355](https://github.com/cosmos/cosmos-sdk/pull/8355) Fix errorWrap `Is` method.
* (x/ibc) [#8341](https://github.com/cosmos/cosmos-sdk/pull/8341) Fix query latest consensus state.
* (proto) [#8350](https://github.com/cosmos/cosmos-sdk/pull/8350), [#8361](https://github.com/cosmos/cosmos-sdk/pull/8361) Update gogo proto deps with v1.3.2 security fixes
* (x/ibc) [#8359](https://github.com/cosmos/cosmos-sdk/pull/8359) Add missing UnpackInterfaces functions to IBC Query Responses. Fixes 'cannot unpack Any' error for IBC types.
* (x/bank) [#8317](https://github.com/cosmos/cosmos-sdk/pull/8317) Fix panic when querying for a not found client denomination metadata.

## [v0.40.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.40.0) - 2021-01-08

v0.40.0, known as the Stargate release of the Cosmos SDK, is one of the largest releases
of the Cosmos SDK since launch. Please read through this changelog and [release notes](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/RELEASE_NOTES.md) to make
sure you are aware of any relevant breaking changes.

### Client Breaking Changes

* **CLI**
    * (client/keys) [#5889](https://github.com/cosmos/cosmos-sdk/pull/5889) remove `keys update` command.
    * (x/auth) [#5844](https://github.com/cosmos/cosmos-sdk/pull/5844) `tx sign` command now returns an error when signing is attempted with offline/multisig keys.
    * (x/auth) [#6108](https://github.com/cosmos/cosmos-sdk/pull/6108) `tx sign` command's `--validate-signatures` flag is migrated into a `tx validate-signatures` standalone command.
    * (x/auth) [#7788](https://github.com/cosmos/cosmos-sdk/pull/7788) Remove `tx auth` subcommands, all auth subcommands exist as `tx <subcommand>`
    * (x/genutil) [#6651](https://github.com/cosmos/cosmos-sdk/pull/6651) The `gentx` command has been improved. No longer are `--from` and `--name` flags required. Instead, a single argument, `name`, is required which refers to the key pair in the Keyring. In addition, an optional
    `--moniker` flag can be provided to override the moniker found in `config.toml`.
    * (x/upgrade) [#7697](https://github.com/cosmos/cosmos-sdk/pull/7697) Rename flag name "--time" to "--upgrade-time", "--info" to "--upgrade-info", to keep it consistent with help message.
* **REST / Queriers**
    * (api) [#6426](https://github.com/cosmos/cosmos-sdk/pull/6426) The ability to start an out-of-process API REST server has now been removed. Instead, the API server is now started in-process along with the application and Tendermint. Configuration options have been added to `app.toml` to enable/disable the API server along with additional HTTP server options.
    * (client) [#7246](https://github.com/cosmos/cosmos-sdk/pull/7246) The rest server endpoint `/swagger-ui/` is replaced by `/swagger/`, and contains swagger documentation for gRPC Gateway routes in addition to legacy REST routes. Swagger API is exposed only if set in `app.toml`.
    * (x/auth) [#5702](https://github.com/cosmos/cosmos-sdk/pull/5702) The `x/auth` querier route has changed from `"acc"` to `"auth"`.
    * (x/bank) [#5572](https://github.com/cosmos/cosmos-sdk/pull/5572) The `/bank/balances/{address}` endpoint now returns all account balances or a single balance by denom when the `denom` query parameter is present.
    * (x/evidence) [#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) Remove CLI and REST handlers for querying `x/evidence` parameters.
    * (x/gov) [#6295](https://github.com/cosmos/cosmos-sdk/pull/6295) Fix typo in querying governance params.
* **General**
    * (baseapp) [#6384](https://github.com/cosmos/cosmos-sdk/pull/6384) The `Result.Data` is now a Protocol Buffer encoded binary blob of type `TxData`. The `TxData` contains `Data` which contains a list of Protocol Buffer encoded message data and the corresponding message type.
    * (client) [#5783](https://github.com/cosmos/cosmos-sdk/issues/5783) Unify all coins representations on JSON client requests for governance proposals.
    * (crypto) [#7419](https://github.com/cosmos/cosmos-sdk/pull/7419) The SDK doesn't use Tendermint's `crypto.PubKey`
    interface anymore, and uses instead it's own `PubKey` interface, defined in `crypto/types`. Replace all instances of
    `crypto.PubKey` by `cryptotypes.Pubkey`.
    * (store/rootmulti) [#6390](https://github.com/cosmos/cosmos-sdk/pull/6390) Proofs of empty stores are no longer supported.
    * (store/types) [#5730](https://github.com/cosmos/cosmos-sdk/pull/5730) store.types.Cp() is removed in favour of types.CopyBytes().
    * (x/auth) [#6054](https://github.com/cosmos/cosmos-sdk/pull/6054) Remove custom JSON marshaling for base accounts as multsigs cannot be bech32 decoded.
    * (x/auth/vesting) [#6859](https://github.com/cosmos/cosmos-sdk/pull/6859) Custom JSON marshaling of vesting accounts was removed. Vesting accounts are now marshaled using their default proto or amino JSON representation.
    * (x/bank) [#5785](https://github.com/cosmos/cosmos-sdk/issues/5785) In x/bank errors, JSON strings coerced to valid UTF-8 bytes at JSON marshalling time
    are now replaced by human-readable expressions. This change can potentially break compatibility with all those client side tools
    that parse log messages.
    * (x/evidence) [#7538](https://github.com/cosmos/cosmos-sdk/pull/7538) The ABCI's `Result.Data` field for
    `MsgSubmitEvidence` responses does not contain the raw evidence's hash, but the protobuf encoded
    `MsgSubmitEvidenceResponse` struct.
    * (x/gov) [#7533](https://github.com/cosmos/cosmos-sdk/pull/7533) The ABCI's `Result.Data` field for
    `MsgSubmitProposal` responses does not contain a raw binary encoding of the `proposalID`, but the protobuf encoded
    `MsgSubmitSubmitProposalResponse` struct.
    * (x/gov) [#6859](https://github.com/cosmos/cosmos-sdk/pull/6859) `ProposalStatus` and `VoteOption` are now JSON serialized using its protobuf name, so expect names like `PROPOSAL_STATUS_DEPOSIT_PERIOD` as opposed to `DepositPeriod`.
    * (x/staking) [#7499](https://github.com/cosmos/cosmos-sdk/pull/7499) `BondStatus` is now a protobuf `enum` instead
    of an `int32`, and JSON serialized using its protobuf name, so expect names like `BOND_STATUS_UNBONDING` as opposed
    to `Unbonding`.
    * (x/staking) [#7556](https://github.com/cosmos/cosmos-sdk/pull/7556) The ABCI's `Result.Data` field for
    `MsgBeginRedelegate` and `MsgUndelegate` responses does not contain custom binary marshaled `completionTime`, but the
    protobuf encoded `MsgBeginRedelegateResponse` and `MsgUndelegateResponse` structs respectively

### API Breaking Changes

* **Baseapp / Client**
    * (AppModule) [#7518](https://github.com/cosmos/cosmos-sdk/pull/7518) [#7584](https://github.com/cosmos/cosmos-sdk/pull/7584) Rename `AppModule.RegisterQueryServices` to `AppModule.RegisterServices`, as this method now registers multiple services (the gRPC query service and the protobuf Msg service). A `Configurator` struct is used to hold the different services.
    * (baseapp) [#5865](https://github.com/cosmos/cosmos-sdk/pull/5865) The `SimulationResponse` returned from tx simulation is now JSON encoded instead of Amino binary.
    * (client) [#6290](https://github.com/cosmos/cosmos-sdk/pull/6290) `CLIContext` is renamed to `Context`. `Context` and all related methods have been moved from package context to client.
    * (client) [#6525](https://github.com/cosmos/cosmos-sdk/pull/6525) Removed support for `indent` in JSON responses. Clients should consider piping to an external tool such as `jq`.
    * (client) [#8107](https://github.com/cosmos/cosmos-sdk/pull/8107) Renamed `PrintOutput` and `PrintOutputLegacy`
    methods of the `context.Client` object to `PrintProto` and `PrintObjectLegacy`.
    * (client/flags) [#6632](https://github.com/cosmos/cosmos-sdk/pull/6632) Remove NewCompletionCmd(), the function is now available in tendermint.
    * (client/input) [#5904](https://github.com/cosmos/cosmos-sdk/pull/5904) Removal of unnecessary `GetCheckPassword`, `PrintPrefixed` functions.
    * (client/keys) [#5889](https://github.com/cosmos/cosmos-sdk/pull/5889) Rename `NewKeyBaseFromDir()` -> `NewLegacyKeyBaseFromDir()`.
    * (client/keys) [#5820](https://github.com/cosmos/cosmos-sdk/pull/5820/) Removed method CloseDB from Keybase interface.
    * (client/rpc) [#6290](https://github.com/cosmos/cosmos-sdk/pull/6290) `client` package and subdirs reorganization.
    * (client/lcd) [#6290](https://github.com/cosmos/cosmos-sdk/pull/6290) `CliCtx` of struct `RestServer` in package client/lcd has been renamed to `ClientCtx`.
    * (codec) [#6330](https://github.com/cosmos/cosmos-sdk/pull/6330) `codec.RegisterCrypto` has been moved to the `crypto/codec` package and the global `codec.Cdc` Amino instance has been deprecated and moved to the `codec/legacy_global` package.
    * (codec) [#8080](https://github.com/cosmos/cosmos-sdk/pull/8080) Updated the `codec.Marshaler` interface
        * Moved `MarshalAny` and `UnmarshalAny` helper functions to `codec.Marshaler` and renamed to `MarshalInterface` and
      `UnmarshalInterface` respectively. These functions must take interface as a parameter (not a concrete type nor `Any`
      object). Underneath they use `Any` wrapping for correct protobuf serialization.
    * (crypto) [#6780](https://github.com/cosmos/cosmos-sdk/issues/6780) Move ledger code to its own package.
    * (crypto/types/multisig) [#6373](https://github.com/cosmos/cosmos-sdk/pull/6373) `multisig.Multisignature` has been renamed to `AminoMultisignature`
    * (codec) `*codec.LegacyAmino` is now a wrapper around Amino which provides backwards compatibility with protobuf `Any`. ALL legacy code should use `*codec.LegacyAmino` instead of `*amino.Codec` directly
    * (crypto) [#5880](https://github.com/cosmos/cosmos-sdk/pull/5880) Merge `crypto/keys/mintkey` into `crypto`.
    * (crypto/hd) [#5904](https://github.com/cosmos/cosmos-sdk/pull/5904) `crypto/keys/hd` moved to `crypto/hd`.
    * (crypto/keyring):
    _ [#5866](https://github.com/cosmos/cosmos-sdk/pull/5866) Rename `crypto/keys/` to `crypto/keyring/`.
    _ [#5904](https://github.com/cosmos/cosmos-sdk/pull/5904) `Keybase` -> `Keyring` interfaces migration. `LegacyKeybase` interface is added in order
    to guarantee limited backward compatibility with the old Keybase interface for the sole purpose of migrating keys across the new keyring backends. `NewLegacy`
    constructor is provided [#5889](https://github.com/cosmos/cosmos-sdk/pull/5889) to allow for smooth migration of keys from the legacy LevelDB based implementation
    to new keyring backends. Plus, the package and the new keyring no longer depends on the sdk.Config singleton. Please consult the [package documentation](https://github.com/cosmos/cosmos-sdk/tree/master/crypto/keyring/doc.go) for more
    information on how to implement the new `Keyring` interface. \* [#5858](https://github.com/cosmos/cosmos-sdk/pull/5858) Make Keyring store keys by name and address's hexbytes representation.
    * (export) [#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) `AppExporter` now returns ABCI consensus parameters to be included in marshaled exported state. These parameters must be returned from the application via the `BaseApp`.
    * (simapp) Deprecating and renaming `MakeEncodingConfig` to `MakeTestEncodingConfig` (both in `simapp` and `simapp/params` packages).
    * (store) [#5803](https://github.com/cosmos/cosmos-sdk/pull/5803) The `store.CommitMultiStore` interface now includes the new `snapshots.Snapshotter` interface as well.
    * (types) [#5579](https://github.com/cosmos/cosmos-sdk/pull/5579) The `keepRecent` field has been removed from the `PruningOptions` type.
    The `PruningOptions` type now only includes fields `KeepEvery` and `SnapshotEvery`, where `KeepEvery`
    determines which committed heights are flushed to disk and `SnapshotEvery` determines which of these
    heights are kept after pruning. The `IsValid` method should be called whenever using these options. Methods
    `SnapshotVersion` and `FlushVersion` accept a version arugment and determine if the version should be
    flushed to disk or kept as a snapshot. Note, `KeepRecent` is automatically inferred from the options
    and provided directly the IAVL store.
    * (types) [#5533](https://github.com/cosmos/cosmos-sdk/pull/5533) Refactored `AppModuleBasic` and `AppModuleGenesis`
    to now accept a `codec.JSONMarshaler` for modular serialization of genesis state.
    * (types/rest) [#5779](https://github.com/cosmos/cosmos-sdk/pull/5779) Drop unused Parse{Int64OrReturnBadRequest,QueryParamBool}() functions.
* **Modules**
    * (modules) [#7243](https://github.com/cosmos/cosmos-sdk/pull/7243) Rename `RegisterCodec` to `RegisterLegacyAminoCodec` and `codec.New()` is now renamed to `codec.NewLegacyAmino()`
    * (modules) [#6564](https://github.com/cosmos/cosmos-sdk/pull/6564) Constant `DefaultParamspace` is removed from all modules, use ModuleName instead.
    * (modules) [#5989](https://github.com/cosmos/cosmos-sdk/pull/5989) `AppModuleBasic.GetTxCmd` now takes a single `CLIContext` parameter.
    * (modules) [#5664](https://github.com/cosmos/cosmos-sdk/pull/5664) Remove amino `Codec` from simulation `StoreDecoder`, which now returns a function closure in order to unmarshal the key-value pairs.
    * (modules) [#5555](https://github.com/cosmos/cosmos-sdk/pull/5555) Move `x/auth/client/utils/` types and functions to `x/auth/client/`.
    * (modules) [#5572](https://github.com/cosmos/cosmos-sdk/pull/5572) Move account balance logic and APIs from `x/auth` to `x/bank`.
    * (modules) [#6326](https://github.com/cosmos/cosmos-sdk/pull/6326) `AppModuleBasic.GetQueryCmd` now takes a single `client.Context` parameter.
    * (modules) [#6336](https://github.com/cosmos/cosmos-sdk/pull/6336) `AppModuleBasic.RegisterQueryService` method was added to support gRPC queries, and `QuerierRoute` and `NewQuerierHandler` were deprecated.
    * (modules) [#6311](https://github.com/cosmos/cosmos-sdk/issues/6311) Remove `alias.go` usage
    * (modules) [#6447](https://github.com/cosmos/cosmos-sdk/issues/6447) Rename `blacklistedAddrs` to `blockedAddrs`.
    * (modules) [#6834](https://github.com/cosmos/cosmos-sdk/issues/6834) Add `RegisterInterfaces` method to `AppModuleBasic` to support registration of protobuf interface types.
    * (modules) [#6734](https://github.com/cosmos/cosmos-sdk/issues/6834) Add `TxEncodingConfig` parameter to `AppModuleBasic.ValidateGenesis` command to support JSON tx decoding in `genutil`.
    * (modules) [#7764](https://github.com/cosmos/cosmos-sdk/pull/7764) Added module initialization options:
        * `server/types.AppExporter` requires extra argument: `AppOptions`.
        * `server.AddCommands` requires extra argument: `addStartFlags types.ModuleInitFlags`
        * `x/crisis.NewAppModule` has a new attribute: `skipGenesisInvariants`. [PR](https://github.com/cosmos/cosmos-sdk/pull/7764)
    * (types) [#6327](https://github.com/cosmos/cosmos-sdk/pull/6327) `sdk.Msg` now inherits `proto.Message`, as a result all `sdk.Msg` types now use pointer semantics.
    * (types) [#7032](https://github.com/cosmos/cosmos-sdk/pull/7032) All types ending with `ID` (e.g. `ProposalID`) now end with `Id` (e.g. `ProposalId`), to match default Protobuf generated format. Also see [#7033](https://github.com/cosmos/cosmos-sdk/pull/7033) for more details.
    * (x/auth) [#6029](https://github.com/cosmos/cosmos-sdk/pull/6029) Module accounts have been moved from `x/supply` to `x/auth`.
    * (x/auth) [#6443](https://github.com/cosmos/cosmos-sdk/issues/6443) Move `FeeTx` and `TxWithMemo` interfaces from `x/auth/ante` to `types`.
    * (x/auth) [#7006](https://github.com/cosmos/cosmos-sdk/pull/7006) All `AccountRetriever` methods now take `client.Context` as a parameter instead of as a struct member.
    * (x/auth) [#6270](https://github.com/cosmos/cosmos-sdk/pull/6270) The passphrase argument has been removed from the signature of the following functions and methods: `BuildAndSign`, ` MakeSignature`, ` SignStdTx`, `TxBuilder.BuildAndSign`, `TxBuilder.Sign`, `TxBuilder.SignStdTx`
    * (x/auth) [#6428](https://github.com/cosmos/cosmos-sdk/issues/6428):
        * `NewAnteHandler` and `NewSigVerificationDecorator` both now take a `SignModeHandler` parameter.
        * `SignatureVerificationGasConsumer` now has the signature: `func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error`.
        * The `SigVerifiableTx` interface now has a `GetSignaturesV2() ([]signing.SignatureV2, error)` method and no longer has the `GetSignBytes` method.
    * (x/auth/tx) [#8106](https://github.com/cosmos/cosmos-sdk/pull/8106) change related to missing append functionality in
    client transaction signing
        * added `overwriteSig` argument to `x/auth/client.SignTx` and `client/tx.Sign` functions.
        * removed `x/auth/tx.go:wrapper.GetSignatures`. The `wrapper` provides `TxBuilder` functionality, and it's a private
      structure. That function was not used at all and it's not exposed through the `TxBuilder` interface.
    * (x/bank) [#7327](https://github.com/cosmos/cosmos-sdk/pull/7327) AddCoins and SubtractCoins no longer return a resultingValue and will only return an error.
    * (x/capability) [#7918](https://github.com/cosmos/cosmos-sdk/pull/7918) Add x/capability safety checks:
        * All outward facing APIs will now check that capability is not nil and name is not empty before performing any state-machine changes
        * `SetIndex` has been renamed to `InitializeIndex`
    * (x/evidence) [#7251](https://github.com/cosmos/cosmos-sdk/pull/7251) New evidence types and light client evidence handling. The module function names changed.
    * (x/evidence) [#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) Remove APIs for getting and setting `x/evidence` parameters. `BaseApp` now uses a `ParamStore` to manage Tendermint consensus parameters which is managed via the `x/params` `Substore` type.
    * (x/gov) [#6147](https://github.com/cosmos/cosmos-sdk/pull/6147) The `Content` field on `Proposal` and `MsgSubmitProposal`
    is now `Any` in concordance with [ADR 019](docs/architecture/adr-019-protobuf-state-encoding.md) and `GetContent` should now
    be used to retrieve the actual proposal `Content`. Also the `NewMsgSubmitProposal` constructor now may return an `error`
    * (x/ibc) [#6374](https://github.com/cosmos/cosmos-sdk/pull/6374) `VerifyMembership` and `VerifyNonMembership` now take a `specs []string` argument to specify the proof format used for verification. Most SDK chains can simply use `commitmenttypes.GetSDKSpecs()` for this argument.
    * (x/params) [#5619](https://github.com/cosmos/cosmos-sdk/pull/5619) The `x/params` keeper now accepts a `codec.Marshaller` instead of
    a reference to an amino codec. Amino is still used for JSON serialization.
    * (x/staking) [#6451](https://github.com/cosmos/cosmos-sdk/pull/6451) `DefaultParamspace` and `ParamKeyTable` in staking module are moved from keeper to types to enforce consistency.
    * (x/staking) [#7419](https://github.com/cosmos/cosmos-sdk/pull/7419) The `TmConsPubKey` method on ValidatorI has been
    removed and replaced instead by `ConsPubKey` (which returns a SDK `cryptotypes.PubKey`) and `TmConsPublicKey` (which
    returns a Tendermint proto PublicKey).
    * (x/staking/types) [#7447](https://github.com/cosmos/cosmos-sdk/issues/7447) Remove bech32 PubKey support:
        * `ValidatorI` interface update. `GetConsPubKey` renamed to `TmConsPubKey` (consensus public key must be a tendermint key). `TmConsPubKey`, `GetConsAddr` methods return error.
        * `Validator` update. Methods changed in `ValidatorI` (as described above) and `ToTmValidator` return error.
        * `Validator.ConsensusPubkey` type changed from `string` to `codectypes.Any`.
        * `MsgCreateValidator.Pubkey` type changed from `string` to `codectypes.Any`.
    * (x/supply) [#6010](https://github.com/cosmos/cosmos-sdk/pull/6010) All `x/supply` types and APIs have been moved to `x/bank`.
    * [#6409](https://github.com/cosmos/cosmos-sdk/pull/6409) Rename all IsEmpty methods to Empty across the codebase and enforce consistency.
    * [#6231](https://github.com/cosmos/cosmos-sdk/pull/6231) Simplify `AppModule` interface, `Route` and `NewHandler` methods become only `Route`
    and returns a new `Route` type.
    * (x/slashing) [#6212](https://github.com/cosmos/cosmos-sdk/pull/6212) Remove `Get*` prefixes from key construction functions
    * (server) [#6079](https://github.com/cosmos/cosmos-sdk/pull/6079) Remove `UpgradeOldPrivValFile` (deprecated in Tendermint Core v0.28).
    * [#5719](https://github.com/cosmos/cosmos-sdk/pull/5719) Bump Go requirement to 1.14+

### State Machine Breaking

* **General**

    * (client) [#7268](https://github.com/cosmos/cosmos-sdk/pull/7268) / [#7147](https://github.com/cosmos/cosmos-sdk/pull/7147) Introduce new protobuf based PubKeys, and migrate PubKey in BaseAccount to use this new protobuf based PubKey format

* **Modules**
    * (modules) [#5572](https://github.com/cosmos/cosmos-sdk/pull/5572) Separate balance from accounts per ADR 004.
    _ Account balances are now persisted and retrieved via the `x/bank` module.
    _ Vesting account interface has been modified to account for changes.
    _ Callers to `NewBaseVestingAccount` are responsible for verifying account balance in relation to
    the original vesting amount.
    _ The `SendKeeper` and `ViewKeeper` interfaces in `x/bank` have been modified to account for changes.
    * (x/auth) [#5533](https://github.com/cosmos/cosmos-sdk/pull/5533) Migrate the `x/auth` module to use Protocol Buffers for state
    serialization instead of Amino.
    _ The `BaseAccount.PubKey` field is now represented as a Bech32 string instead of a `crypto.Pubkey`.
    _ `NewBaseAccountWithAddress` now returns a reference to a `BaseAccount`.
    _ The `x/auth` module now accepts a `Codec` interface which extends the `codec.Marshaler` interface by
    requiring a concrete codec to know how to serialize accounts.
    _ The `AccountRetriever` type now accepts a `Codec` in its constructor in order to know how to
    serialize accounts.
    * (x/bank) [#6518](https://github.com/cosmos/cosmos-sdk/pull/6518) Support for global and per-denomination send enabled flags.
        * Existing send_enabled global flag has been moved into a Params structure as `default_send_enabled`.
        * An array of: `{denom: string, enabled: bool}` is added to bank Params to support per-denomination override of global default value.
    * (x/distribution) [#5610](https://github.com/cosmos/cosmos-sdk/pull/5610) Migrate the `x/distribution` module to use Protocol Buffers for state
    serialization instead of Amino. The exact codec used is `codec.HybridCodec` which utilizes Protobuf for binary encoding and Amino
    for JSON encoding.
    _ `ValidatorHistoricalRewards.ReferenceCount` is now of types `uint32` instead of `uint16`.
    _ `ValidatorSlashEvents` is now a struct with `slashevents`.
    _ `ValidatorOutstandingRewards` is now a struct with `rewards`.
    _ `ValidatorAccumulatedCommission` is now a struct with `commission`. \* The `Keeper` constructor now takes a `codec.Marshaler` instead of a concrete Amino codec. This exact type
    provided is specified by `ModuleCdc`.
    * (x/evidence) [#5634](https://github.com/cosmos/cosmos-sdk/pull/5634) Migrate the `x/evidence` module to use Protocol Buffers for state
    serialization instead of Amino.
    _ The `internal` sub-package has been removed in order to expose the types proto file.
    _ The module now accepts a `Codec` interface which extends the `codec.Marshaler` interface by
    requiring a concrete codec to know how to serialize `Evidence` types. \* The `MsgSubmitEvidence` message has been removed in favor of `MsgSubmitEvidenceBase`. The application-level
    codec must now define the concrete `MsgSubmitEvidence` type which must implement the module's `MsgSubmitEvidence`
    interface.
    * (x/evidence) [#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) Remove parameters from `x/evidence` genesis and module state. The `x/evidence` module now solely uses Tendermint consensus parameters to determine of evidence is valid or not.
    * (x/gov) [#5737](https://github.com/cosmos/cosmos-sdk/pull/5737) Migrate the `x/gov` module to use Protocol
    Buffers for state serialization instead of Amino.
    _ `MsgSubmitProposal` will be removed in favor of the application-level proto-defined `MsgSubmitProposal` which
    implements the `MsgSubmitProposalI` interface. Applications should extend the `NewMsgSubmitProposalBase` type
    to define their own concrete `MsgSubmitProposal` types.
    _ The module now accepts a `Codec` interface which extends the `codec.Marshaler` interface by
    requiring a concrete codec to know how to serialize `Proposal` types.
    * (x/mint) [#5634](https://github.com/cosmos/cosmos-sdk/pull/5634) Migrate the `x/mint` module to use Protocol Buffers for state
    serialization instead of Amino. \* The `internal` sub-package has been removed in order to expose the types proto file.
    * (x/slashing) [#5627](https://github.com/cosmos/cosmos-sdk/pull/5627) Migrate the `x/slashing` module to use Protocol Buffers for state
    serialization instead of Amino. The exact codec used is `codec.HybridCodec` which utilizes Protobuf for binary encoding and Amino
    for JSON encoding. \* The `Keeper` constructor now takes a `codec.Marshaler` instead of a concrete Amino codec. This exact type
    provided is specified by `ModuleCdc`.
    * (x/staking) [#6844](https://github.com/cosmos/cosmos-sdk/pull/6844) Validators are now inserted into the unbonding queue based on their unbonding time and height. The relevant keeper APIs are modified to reflect these changes by now also requiring a height.
    * (x/staking) [#6061](https://github.com/cosmos/cosmos-sdk/pull/6061) Allow a validator to immediately unjail when no signing info is present due to
    falling below their minimum self-delegation and never having been bonded. The validator may immediately unjail once they've met their minimum self-delegation.
    * (x/staking) [#5600](https://github.com/cosmos/cosmos-sdk/pull/5600) Migrate the `x/staking` module to use Protocol Buffers for state
    serialization instead of Amino. The exact codec used is `codec.HybridCodec` which utilizes Protobuf for binary encoding and Amino
    for JSON encoding.
    _ `BondStatus` is now of type `int32` instead of `byte`.
    _ Types of `int16` in the `Params` type are now of type `int32`.
    _ Every reference of `crypto.Pubkey` in context of a `Validator` is now of type string. `GetPubKeyFromBech32` must be used to get the `crypto.Pubkey`.
    _ The `Keeper` constructor now takes a `codec.Marshaler` instead of a concrete Amino codec. This exact type
    provided is specified by `ModuleCdc`.
    * (x/staking) [#7979](https://github.com/cosmos/cosmos-sdk/pull/7979) keeper pubkey storage serialization migration
    from bech32 to protobuf.
    * (x/supply) [#6010](https://github.com/cosmos/cosmos-sdk/pull/6010) Removed the `x/supply` module by merging the existing types and APIs into the `x/bank` module.
    * (x/supply) [#5533](https://github.com/cosmos/cosmos-sdk/pull/5533) Migrate the `x/supply` module to use Protocol Buffers for state
    serialization instead of Amino.
    _ The `internal` sub-package has been removed in order to expose the types proto file.
    _ The `x/supply` module now accepts a `Codec` interface which extends the `codec.Marshaler` interface by
    requiring a concrete codec to know how to serialize `SupplyI` types. \* The `SupplyI` interface has been modified to no longer return `SupplyI` on methods. Instead the
    concrete type's receiver should modify the type.
    * (x/upgrade) [#5659](https://github.com/cosmos/cosmos-sdk/pull/5659) Migrate the `x/upgrade` module to use Protocol
    Buffers for state serialization instead of Amino.
    _ The `internal` sub-package has been removed in order to expose the types proto file.
    _ The `x/upgrade` module now accepts a `codec.Marshaler` interface.

### Features

* **Baseapp / Client / REST**
    * (x/auth) [#6213](https://github.com/cosmos/cosmos-sdk/issues/6213) Introduce new protobuf based path for transaction signing, see [ADR020](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-020-protobuf-transaction-encoding.md) for more details
    * (x/auth) [#6350](https://github.com/cosmos/cosmos-sdk/pull/6350) New sign-batch command to sign StdTx batch files.
    * (baseapp) [#5803](https://github.com/cosmos/cosmos-sdk/pull/5803) Added support for taking state snapshots at regular height intervals, via options `snapshot-interval` and `snapshot-keep-recent`.
    * (baseapp) [#7519](https://github.com/cosmos/cosmos-sdk/pull/7519) Add `ServiceMsgRouter` to BaseApp to handle routing of protobuf service `Msg`s. The two new types defined in ADR 031, `sdk.ServiceMsg` and `sdk.MsgRequest` are introduced with this router.
    * (client) [#5921](https://github.com/cosmos/cosmos-sdk/issues/5921) Introduce new gRPC and gRPC Gateway based APIs for querying app & module data. See [ADR021](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-021-protobuf-query-encoding.md) for more details
    * (cli) [#7485](https://github.com/cosmos/cosmos-sdk/pull/7485) Introduce a new optional `--keyring-dir` flag that allows clients to specify a Keyring directory if it does not reside in the directory specified by `--home`.
    * (cli) [#7221](https://github.com/cosmos/cosmos-sdk/pull/7221) Add the option of emitting amino encoded json from the CLI
    * (codec) [#7519](https://github.com/cosmos/cosmos-sdk/pull/7519) `InterfaceRegistry` now inherits `jsonpb.AnyResolver`, and has a `RegisterCustomTypeURL` method to support ADR 031 packing of `Any`s. `AnyResolver` is now a required parameter to `RejectUnknownFields`.
    * (coin) [#6755](https://github.com/cosmos/cosmos-sdk/pull/6755) Add custom regex validation for `Coin` denom by overwriting `CoinDenomRegex` when using `/types/coin.go`.
    * (config) [#7265](https://github.com/cosmos/cosmos-sdk/pull/7265) Support Tendermint block pruning through a new `min-retain-blocks` configuration that can be set in either `app.toml` or via the CLI. This parameter is used in conjunction with other criteria to determine the height at which Tendermint should prune blocks.
    * (events) [#7121](https://github.com/cosmos/cosmos-sdk/pull/7121) The application now derives what events are indexed by Tendermint via the `index-events` configuration in `app.toml`, which is a list of events taking the form `{eventType}.{attributeKey}`.
    * (tx) [#6089](https://github.com/cosmos/cosmos-sdk/pull/6089) Transactions can now have a `TimeoutHeight` set which allows the transaction to be rejected if it's committed at a height greater than the timeout.
    * (rest) [#6167](https://github.com/cosmos/cosmos-sdk/pull/6167) Support `max-body-bytes` CLI flag for the REST service.
    * (genesis) [#7089](https://github.com/cosmos/cosmos-sdk/pull/7089) The `export` command now adds a `initial_height` field in the exported JSON. Baseapp's `CommitMultiStore` now also has a `SetInitialVersion` setter, so it can set the initial store version inside `InitChain` and start a new chain from a given height.
* **General**
    * (crypto/multisig) [#6241](https://github.com/cosmos/cosmos-sdk/pull/6241) Add Multisig type directly to the repo. Previously this was in tendermint.
    * (codec/types) [#8106](https://github.com/cosmos/cosmos-sdk/pull/8106) Adding `NewAnyWithCustomTypeURL` to correctly
    marshal Messages in TxBuilder.
    * (tests) [#6489](https://github.com/cosmos/cosmos-sdk/pull/6489) Introduce package `testutil`, new in-process testing network framework for use in integration and unit tests.
    * (tx) Add new auth/tx gRPC & gRPC-Gateway endpoints for basic querying & broadcasting support
        * [#7842](https://github.com/cosmos/cosmos-sdk/pull/7842) Add TxsByEvent gRPC endpoint
        * [#7852](https://github.com/cosmos/cosmos-sdk/pull/7852) Add tx broadcast gRPC endpoint
    * (tx) [#7688](https://github.com/cosmos/cosmos-sdk/pull/7688) Add a new Tx gRPC service with methods `Simulate` and `GetTx` (by hash).
    * (store) [#5803](https://github.com/cosmos/cosmos-sdk/pull/5803) Added `rootmulti.Store` methods for taking and restoring snapshots, based on `iavl.Store` export/import.
    * (store) [#6324](https://github.com/cosmos/cosmos-sdk/pull/6324) IAVL store query proofs now return CommitmentOp which wraps an ics23 CommitmentProof
    * (store) [#6390](https://github.com/cosmos/cosmos-sdk/pull/6390) `RootMulti` store query proofs now return `CommitmentOp` which wraps `CommitmentProofs`
        * `store.Query` now only returns chained `ics23.CommitmentProof` wrapped in `merkle.Proof`
        * `ProofRuntime` only decodes and verifies `ics23.CommitmentProof`
* **Modules**
    * (modules) [#5921](https://github.com/cosmos/cosmos-sdk/issues/5921) Introduction of Query gRPC service definitions along with REST annotations for gRPC Gateway for each module
    * (modules) [#7540](https://github.com/cosmos/cosmos-sdk/issues/7540) Protobuf service definitions can now be used for
    packing `Msg`s in transactions as defined in [ADR 031](./docs/architecture/adr-031-msg-service.md). All modules now
    define a `Msg` protobuf service.
    * (x/auth/vesting) [#7209](https://github.com/cosmos/cosmos-sdk/pull/7209) Create new `MsgCreateVestingAccount` message type along with CLI handler that allows for the creation of delayed and continuous vesting types.
    * (x/capability) [#5828](https://github.com/cosmos/cosmos-sdk/pull/5828) Capability module integration as outlined in [ADR 3 - Dynamic Capability Store](https://github.com/cosmos/tree/master/docs/architecture/adr-003-dynamic-capability-store.md).
    * (x/crisis) `x/crisis` has a new function: `AddModuleInitFlags`, which will register optional crisis module flags for the start command.
    * (x/ibc) [#5277](https://github.com/cosmos/cosmos-sdk/pull/5277) `x/ibc` changes from IBC alpha. For more details check the [`x/ibc/core/spec`](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/core/spec) directory, or the ICS specs below:
        * [ICS 002 - Client Semantics](https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics) subpackage
        * [ICS 003 - Connection Semantics](https://github.com/cosmos/ics/blob/master/spec/ics-003-connection-semantics) subpackage
        * [ICS 004 - Channel and Packet Semantics](https://github.com/cosmos/ics/blob/master/spec/ics-004-channel-and-packet-semantics) subpackage
        * [ICS 005 - Port Allocation](https://github.com/cosmos/ics/blob/master/spec/ics-005-port-allocation) subpackage
        * [ICS 006 - Solo Machine Client](https://github.com/cosmos/ics/tree/master/spec/ics-006-solo-machine-client) subpackage
        * [ICS 007 - Tendermint Client](https://github.com/cosmos/ics/blob/master/spec/ics-007-tendermint-client) subpackage
        * [ICS 009 - Loopback Client](https://github.com/cosmos/ics/tree/master/spec/ics-009-loopback-client) subpackage
        * [ICS 020 - Fungible Token Transfer](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer) subpackage
        * [ICS 023 - Vector Commitments](https://github.com/cosmos/ics/tree/master/spec/ics-023-vector-commitments) subpackage
        * [ICS 024 - Host State Machine Requirements](https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements) subpackage
    * (x/ibc) [#6374](https://github.com/cosmos/cosmos-sdk/pull/6374) ICS-23 Verify functions will now accept and verify ics23 CommitmentProofs exclusively
    * (x/params) [#6005](https://github.com/cosmos/cosmos-sdk/pull/6005) Add new CLI command for querying raw x/params parameters by subspace and key.

### Bug Fixes

* **Baseapp / Client / REST**
    * (client) [#5964](https://github.com/cosmos/cosmos-sdk/issues/5964) `--trust-node` is now false by default - for real. Users must ensure it is set to true if they don't want to enable the verifier.
    * (client) [#6402](https://github.com/cosmos/cosmos-sdk/issues/6402) Fix `keys add` `--algo` flag which only worked for Tendermint's `secp256k1` default key signing algorithm.
    * (client) [#7699](https://github.com/cosmos/cosmos-sdk/pull/7699) Fix panic in context when setting invalid nodeURI. `WithNodeURI` does not set the `Client` in the context.
    * (export) [#6510](https://github.com/cosmos/cosmos-sdk/pull/6510/) Field TimeIotaMs now is included in genesis file while exporting.
    * (rest) [#5906](https://github.com/cosmos/cosmos-sdk/pull/5906) Fix an issue that make some REST calls panic when sending invalid or incomplete requests.
    * (crypto) [#7966](https://github.com/cosmos/cosmos-sdk/issues/7966) `Bip44Params` `String()` function now correctly
    returns the absolute HD path by adding the `m/` prefix.
    * (crypto/keyring) [#5844](https://github.com/cosmos/cosmos-sdk/pull/5844) `Keyring.Sign()` methods no longer decode amino signatures when method receivers
    are offline/multisig keys.
    * (store) [#7415](https://github.com/cosmos/cosmos-sdk/pull/7415) Allow new stores to be registered during on-chain upgrades.
* **Modules**
  _ (modules) [#5569](https://github.com/cosmos/cosmos-sdk/issues/5569) `InitGenesis`, for the relevant modules, now ensures module accounts exist.
  _ (x/auth) [#5892](https://github.com/cosmos/cosmos-sdk/pull/5892) Add `RegisterKeyTypeCodec` to register new
  types (eg. keys) to the `auth` module internal amino codec.
  _ (x/bank) [#6536](https://github.com/cosmos/cosmos-sdk/pull/6536) Fix bug in `WriteGeneratedTxResponse` function used by multiple
  REST endpoints. Now it writes a Tx in StdTx format.
  _ (x/genutil) [#5938](https://github.com/cosmos/cosmos-sdk/pull/5938) Fix `InitializeNodeValidatorFiles` error handling.
  _ (x/gentx) [#8183](https://github.com/cosmos/cosmos-sdk/pull/8183) change gentx cmd amount to arg from flag
  _ (x/gov) [#7641](https://github.com/cosmos/cosmos-sdk/pull/7641) Fix tally calculation precision error.
  _ (x/staking) [#6529](https://github.com/cosmos/cosmos-sdk/pull/6529) Export validator addresses (previously was empty).
  _ (x/staking) [#5949](https://github.com/cosmos/cosmos-sdk/pull/5949) Skip staking `HistoricalInfoKey` in simulations as headers are not exported. \* (x/staking) [#6061](https://github.com/cosmos/cosmos-sdk/pull/6061) Allow a validator to immediately unjail when no signing info is present due to
  falling below their minimum self-delegation and never having been bonded. The validator may immediately unjail once they've met their minimum self-delegation.
* **General**
    * (types) [#7038](https://github.com/cosmos/cosmos-sdk/issues/7038) Fix infinite looping of `ApproxRoot` by including a hard-coded maximum iterations limit of 100.
    * (types) [#7084](https://github.com/cosmos/cosmos-sdk/pull/7084) Fix panic when calling `BigInt()` on an uninitialized `Int`.
    * (simulation) [#7129](https://github.com/cosmos/cosmos-sdk/issues/7129) Fix support for custom `Account` and key types on auth's simulation.

### Improvements

* **Baseapp / Client / REST**
    * (baseapp) [#6186](https://github.com/cosmos/cosmos-sdk/issues/6186) Support emitting events during `AnteHandler` execution.
    * (baseapp) [#6053](https://github.com/cosmos/cosmos-sdk/pull/6053) Customizable panic recovery handling added for `app.runTx()` method (as proposed in the [ADR 22](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-022-custom-panic-handling.md)). Adds ability for developers to register custom panic handlers extending standard ones.
    * (client) [#5810](https://github.com/cosmos/cosmos-sdk/pull/5810) Added a new `--offline` flag that allows commands to be executed without an
    internet connection. Previously, `--generate-only` served this purpose in addition to only allowing txs to be generated. Now, `--generate-only` solely
    allows txs to be generated without being broadcasted and disallows Keybase use and `--offline` allows the use of Keybase but does not allow any
    functionality that requires an online connection.
    * (cli) [#7764](https://github.com/cosmos/cosmos-sdk/pull/7764) Update x/banking and x/crisis InitChain to improve node startup time
    * (client) [#5856](https://github.com/cosmos/cosmos-sdk/pull/5856) Added the possibility to set `--offline` flag with config command.
    * (client) [#5895](https://github.com/cosmos/cosmos-sdk/issues/5895) show config options in the config command's help screen.
    * (client/keys) [#8043](https://github.com/cosmos/cosmos-sdk/pull/8043) Add support for export of unarmored private key
    * (client/tx) [#7801](https://github.com/cosmos/cosmos-sdk/pull/7801) Update sign-batch multisig to work online
    * (x/genutil) [#8099](https://github.com/cosmos/cosmos-sdk/pull/8099) `init` now supports a `--recover` flag to recover
    the private validator key from a given mnemonic
* **Modules**
    * (x/auth) [#5702](https://github.com/cosmos/cosmos-sdk/pull/5702) Add parameter querying support for `x/auth`.
    * (x/auth/ante) [#6040](https://github.com/cosmos/cosmos-sdk/pull/6040) `AccountKeeper` interface used for `NewAnteHandler` and handler's decorators to add support of using custom `AccountKeeper` implementations.
    * (x/evidence) [#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) Tendermint Consensus parameters can now be changed via parameter change proposals through `x/gov`.
    * (x/evidence) [#5961](https://github.com/cosmos/cosmos-sdk/issues/5961) Add `StoreDecoder` simulation for evidence module.
    * (x/ibc) [#5948](https://github.com/cosmos/cosmos-sdk/issues/5948) Add `InitGenesis` and `ExportGenesis` functions for `ibc` module.
    * (x/ibc-transfer) [#6871](https://github.com/cosmos/cosmos-sdk/pull/6871) Implement [ADR 001 - Coin Source Tracing](./docs/architecture/adr-001-coin-source-tracing.md).
    * (x/staking) [#6059](https://github.com/cosmos/cosmos-sdk/pull/6059) Updated `HistoricalEntries` parameter default to 100.
    * (x/staking) [#5584](https://github.com/cosmos/cosmos-sdk/pull/5584) Add util function `ToTmValidator` that converts a `staking.Validator` type to `*tmtypes.Validator`.
    * (x/staking) [#6163](https://github.com/cosmos/cosmos-sdk/pull/6163) CLI and REST call to unbonding delegations and delegations now accept
    pagination.
    * (x/staking) [#8178](https://github.com/cosmos/cosmos-sdk/pull/8178) Update default historical header number for stargate
* **General**
    * (crypto) [#7987](https://github.com/cosmos/cosmos-sdk/pull/7987) Fix the inconsistency of CryptoCdc, only use
    `codec/legacy.Cdc`.
    * (logging) [#8072](https://github.com/cosmos/cosmos-sdk/pull/8072) Refactor logging:
    _ Use [zerolog](https://github.com/rs/zerolog) over Tendermint's go-kit logging wrapper.
    _ Introduce Tendermint's `--log_format=plain|json` flag. Using format `json` allows for emitting structured JSON
    logs which can be consumed by an external logging facility (e.g. Loggly). Both formats log to STDERR. \* The existing `--log_level` flag and it's default value now solely relates to the global logging
    level (e.g. `info`, `debug`, etc...) instead of `<module>:<level>`.
    * (rest) [#7649](https://github.com/cosmos/cosmos-sdk/pull/7649) Return an unsigned tx in legacy GET /tx endpoint when signature conversion fails
    * (simulation) [#6002](https://github.com/cosmos/cosmos-sdk/pull/6002) Add randomized consensus params into simulation.
    * (store) [#6481](https://github.com/cosmos/cosmos-sdk/pull/6481) Move `SimpleProofsFromMap` from Tendermint into the SDK.
    * (store) [#6719](https://github.com/cosmos/cosmos-sdk/6754) Add validity checks to stores for nil and empty keys.
    * (SDK) Updated dependencies
        * Updated iavl dependency to v0.15.3
        * Update tendermint to v0.34.1
    * (types) [#7027](https://github.com/cosmos/cosmos-sdk/pull/7027) `Coin(s)` and `DecCoin(s)` updates:
        * Bump denomination max length to 128
        * Allow uppercase letters and numbers in denominations to support [ADR 001](./docs/architecture/adr-001-coin-source-tracing.md)
        * Added `Validate` function that returns a descriptive error
    * (types) [#5581](https://github.com/cosmos/cosmos-sdk/pull/5581) Add convenience functions {,Must}Bech32ifyAddressBytes.
    * (types/module) [#5724](https://github.com/cosmos/cosmos-sdk/issues/5724) The `types/module` package does no longer depend on `x/simulation`.
    * (types) [#5585](https://github.com/cosmos/cosmos-sdk/pull/5585) IBC additions:
        * `Coin` denomination max lenght has been increased to 32.
        * Added `CapabilityKey` alias for `StoreKey` to match IBC spec.
    * (types/rest) [#5900](https://github.com/cosmos/cosmos-sdk/pull/5900) Add Check\*Error function family to spare developers from replicating tons of boilerplate code.
    * (types) [#6128](https://github.com/cosmos/cosmos-sdk/pull/6137) Add `String()` method to `GasMeter`.
    * (types) [#6195](https://github.com/cosmos/cosmos-sdk/pull/6195) Add codespace to broadcast(sync/async) response.
    * (types) [#6897](https://github.com/cosmos/cosmos-sdk/issues/6897) Add KV type from tendermint to `types` directory.
    * (version) [#7848](https://github.com/cosmos/cosmos-sdk/pull/7848) [#7941](https://github.com/cosmos/cosmos-sdk/pull/7941)
    `version --long` output now shows the list of build dependencies and replaced build dependencies.

## [v0.39.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.39.1) - 2020-08-11

### Client Breaking

* (x/auth) [#6861](https://github.com/cosmos/cosmos-sdk/pull/6861) Remove public key Bech32 encoding for all account types for JSON serialization, instead relying on direct Amino encoding. In addition, JSON serialization utilizes Amino instead of the Go stdlib, so integers are treated as strings.

### Improvements

* (client) [#6853](https://github.com/cosmos/cosmos-sdk/pull/6853) Add --unsafe-cors flag.

## [v0.39.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.39.0) - 2020-07-20

### Improvements

* (deps) Bump IAVL version to [v0.14.0](https://github.com/cosmos/iavl/releases/tag/v0.14.0)
* (client) [#5585](https://github.com/cosmos/cosmos-sdk/pull/5585) `CLIContext` additions:
    * Introduce `QueryABCI` that returns the full `abci.ResponseQuery` with inclusion Merkle proofs.
    * Added `prove` flag for Merkle proof verification.
* (x/staking) [#6791)](https://github.com/cosmos/cosmos-sdk/pull/6791) Close {UBDQueue,RedelegationQueu}Iterator once used.

### API Breaking Changes

* (baseapp) [#5837](https://github.com/cosmos/cosmos-sdk/issues/5837) Transaction simulation now returns a `SimulationResponse` which contains the `GasInfo` and `Result` from the execution.

### Client Breaking Changes

* (x/auth) [#6745](https://github.com/cosmos/cosmos-sdk/issues/6745) Remove BaseAccount's custom JSON {,un}marshalling.

### Bug Fixes

* (store) [#6475](https://github.com/cosmos/cosmos-sdk/pull/6475) Revert IAVL pruning functionality introduced in
  [v0.13.0](https://github.com/cosmos/iavl/releases/tag/v0.13.0),
  where the IAVL no longer keeps states in-memory in which it flushes periodically. IAVL now commits and
  flushes every state to disk as it did pre-v0.13.0. The SDK's multi-store will track and ensure the proper
  heights are pruned. The operator can set the pruning options via a `pruning` config via the CLI or
  through `app.toml`. The `pruning` flag exposes `default|everything|nothing|custom` as options --
  see docs for further details. If the operator chooses `custom`, they may provide granular pruning
  options `pruning-keep-recent`, `pruning-keep-every`, and `pruning-interval`. The former two options
  dictate how many recent versions are kept on disk and the offset of what versions are kept after that
  respectively, and the latter defines the height interval in which versions are deleted in a batch.
  **Note, there are some client-facing API breaking changes with regard to IAVL, stores, and pruning settings.**
* (x/distribution) [#6210](https://github.com/cosmos/cosmos-sdk/pull/6210) Register `MsgFundCommunityPool` in distribution amino codec.
* (types) [#5741](https://github.com/cosmos/cosmos-sdk/issues/5741) Prevent `ChainAnteDecorators()` from panicking when empty `AnteDecorator` slice is supplied.
* (baseapp) [#6306](https://github.com/cosmos/cosmos-sdk/issues/6306) Prevent events emitted by the antehandler from being persisted between transactions.
* (client/keys) [#5091](https://github.com/cosmos/cosmos-sdk/issues/5091) `keys parse` does not honor client app's configuration.
* (x/bank) [#6674](https://github.com/cosmos/cosmos-sdk/pull/6674) Create account if recipient does not exist on handing `MsgMultiSend`.
* (x/auth) [#6287](https://github.com/cosmos/cosmos-sdk/pull/6287) Fix nonce stuck when sending multiple transactions from an account in a same block.

## [v0.38.5] - 2020-07-02

### Improvements

* (tendermint) Bump Tendermint version to [v0.33.6](https://github.com/tendermint/tendermint/releases/tag/v0.33.6).

## [v0.38.4] - 2020-05-21

### Bug Fixes

* (x/auth) [#5950](https://github.com/cosmos/cosmos-sdk/pull/5950) Fix `IncrementSequenceDecorator` to use is `IsReCheckTx` instead of `IsCheckTx` to allow account sequence incrementing.

## [v0.38.3] - 2020-04-09

### Improvements

* (tendermint) Bump Tendermint version to [v0.33.3](https://github.com/tendermint/tendermint/releases/tag/v0.33.3).

## [v0.38.2] - 2020-03-25

### Bug Fixes

* (baseapp) [#5718](https://github.com/cosmos/cosmos-sdk/pull/5718) Remove call to `ctx.BlockGasMeter` during failed message validation which resulted in a panic when the tx execution mode was `CheckTx`.
* (x/genutil) [#5775](https://github.com/cosmos/cosmos-sdk/pull/5775) Fix `ExportGenesis` in `x/genutil` to export default genesis state (`[]`) instead of `null`.
* (client) [#5618](https://github.com/cosmos/cosmos-sdk/pull/5618) Fix crash on the client when the verifier is not set.
* (crypto/keys/mintkey) [#5823](https://github.com/cosmos/cosmos-sdk/pull/5823) fix errors handling in `UnarmorPubKeyBytes` (underlying armoring function's return error was not being checked).
* (x/distribution) [#5620](https://github.com/cosmos/cosmos-sdk/pull/5620) Fix nil pointer deref in distribution tax/reward validation helpers.

### Improvements

* (rest) [#5648](https://github.com/cosmos/cosmos-sdk/pull/5648) Enhance /txs usability:
    * Add `tx.minheight` key to filter transaction with an inclusive minimum block height
    * Add `tx.maxheight` key to filter transaction with an inclusive maximum block height
* (crypto/keys) [#5739](https://github.com/cosmos/cosmos-sdk/pull/5739) Print an error message if the password input failed.

## [v0.38.1] - 2020-02-11

### Improvements

* (modules) [#5597](https://github.com/cosmos/cosmos-sdk/pull/5597) Add `amount` event attribute to the `complete_unbonding`
  and `complete_redelegation` events that reflect the total balances of the completed unbondings and redelegations
  respectively.

### Bug Fixes

* (types) [#5579](https://github.com/cosmos/cosmos-sdk/pull/5579) The IAVL `Store#Commit` method has been refactored to
  delete a flushed version if it is not a snapshot version. The root multi-store now keeps track of `commitInfo` instead
  of `types.CommitID`. During `Commit` of the root multi-store, `lastCommitInfo` is updated from the saved state
  and is only flushed to disk if it is a snapshot version. During `Query` of the root multi-store, if the request height
  is the latest height, we'll use the store's `lastCommitInfo`. Otherwise, we fetch `commitInfo` from disk.
* (x/bank) [#5531](https://github.com/cosmos/cosmos-sdk/issues/5531) Added missing amount event to MsgMultiSend, emitted for each output.
* (x/gov) [#5622](https://github.com/cosmos/cosmos-sdk/pull/5622) Track any events emitted from a proposal's handler upon successful execution.

## [v0.38.0] - 2020-01-23

### State Machine Breaking

* (genesis) [#5506](https://github.com/cosmos/cosmos-sdk/pull/5506) The `x/distribution` genesis state
  now includes `params` instead of individual parameters.
* (genesis) [#5017](https://github.com/cosmos/cosmos-sdk/pull/5017) The `x/genaccounts` module has been
  deprecated and all components removed except the `legacy/` package. This requires changes to the
  genesis state. Namely, `accounts` now exist under `app_state.auth.accounts`. The corresponding migration
  logic has been implemented for v0.38 target version. Applications can migrate via:
  `$ {appd} migrate v0.38 genesis.json`.
* (modules) [#5299](https://github.com/cosmos/cosmos-sdk/pull/5299) Handling of `ABCIEvidenceTypeDuplicateVote`
  during `BeginBlock` along with the corresponding parameters (`MaxEvidenceAge`) have moved from the
  `x/slashing` module to the `x/evidence` module.

### API Breaking Changes

* (modules) [#5506](https://github.com/cosmos/cosmos-sdk/pull/5506) Remove individual setters of `x/distribution` parameters. Instead, follow the module spec in getting parameters, setting new value(s) and finally calling `SetParams`.
* (types) [#5495](https://github.com/cosmos/cosmos-sdk/pull/5495) Remove redundant `(Must)Bech32ify*` and `(Must)Get*KeyBech32` functions in favor of `(Must)Bech32ifyPubKey` and `(Must)GetPubKeyFromBech32` respectively, both of which take a `Bech32PubKeyType` (string).
* (types) [#5430](https://github.com/cosmos/cosmos-sdk/pull/5430) `DecCoins#Add` parameter changed from `DecCoins`
  to `...DecCoin`, `Coins#Add` parameter changed from `Coins` to `...Coin`.
* (baseapp/types) [#5421](https://github.com/cosmos/cosmos-sdk/pull/5421) The `Error` interface (`types/errors.go`)
  has been removed in favor of the concrete type defined in `types/errors/` which implements the standard `error` interface.
    * As a result, the `Handler` and `Querier` implementations now return a standard `error`.
  Within `BaseApp`, `runTx` now returns a `(GasInfo, *Result, error)`tuple and`runMsgs`returns a `(_Result, error)`tuple. A reference to a`Result`is now used to indicate success whereas an error signals an invalid message or failed message execution. As a result, the fields`Code`, `Codespace`, `GasWanted`, and `GasUsed`have been removed the`Result`type. The latter two fields are now found in the`GasInfo` type which is always returned regardless of execution outcome.
  _ Note to developers: Since all handlers and queriers must now return a standard `error`, the `types/errors/`
  package contains all the relevant and pre-registered errors that you typically work with. A typical
  error returned will look like `sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "...")`. You can retrieve
  relevant ABCI information from the error via `ABCIInfo`.
* (client) [#5442](https://github.com/cosmos/cosmos-sdk/pull/5442) Remove client/alias.go as it's not necessary and
  components can be imported directly from the packages.
* (store) [#4748](https://github.com/cosmos/cosmos-sdk/pull/4748) The `CommitMultiStore` interface
  now requires a `SetInterBlockCache` method. Applications that do not wish to support this can simply
  have this method perform a no-op.
* (modules) [#4665](https://github.com/cosmos/cosmos-sdk/issues/4665) Refactored `x/gov` module structure and dev-UX:
    * Prepare for module spec integration
    * Update gov keys to use big endian encoding instead of little endian
* (modules) [#5017](https://github.com/cosmos/cosmos-sdk/pull/5017) The `x/genaccounts` module has been deprecated and all components removed except the `legacy/` package.
* [#4486](https://github.com/cosmos/cosmos-sdk/issues/4486) Vesting account types decoupled from the `x/auth` module and now live under `x/auth/vesting`. Applications wishing to use vesting account types must be sure to register types via `RegisterCodec` under the new vesting package.
* [#4486](https://github.com/cosmos/cosmos-sdk/issues/4486) The `NewBaseVestingAccount` constructor returns an error
  if the provided arguments are invalid.
* (x/auth) [#5006](https://github.com/cosmos/cosmos-sdk/pull/5006) Modular `AnteHandler` via composable decorators:
    * The `AnteHandler` interface now returns `(newCtx Context, err error)` instead of `(newCtx Context, result sdk.Result, abort bool)`
    * The `NewAnteHandler` function returns an `AnteHandler` function that returns the new `AnteHandler`
    interface and has been moved into the `auth/ante` directory.
    * `ValidateSigCount`, `ValidateMemo`, `ProcessPubKey`, `EnsureSufficientMempoolFee`, and `GetSignBytes`
    have all been removed as public functions.
    * Invalid Signatures may return `InvalidPubKey` instead of `Unauthorized` error, since the transaction
    will first hit `SetPubKeyDecorator` before the `SigVerificationDecorator` runs.
    * `StdTx#GetSignatures` will return an array of just signature byte slices `[][]byte` instead of
    returning an array of `StdSignature` structs. To replicate the old behavior, use the public field
    `StdTx.Signatures` to get back the array of StdSignatures `[]StdSignature`.
* (modules) [#5299](https://github.com/cosmos/cosmos-sdk/pull/5299) `HandleDoubleSign` along with params `MaxEvidenceAge` and `DoubleSignJailEndTime` have moved from the `x/slashing` module to the `x/evidence` module.
* (keys) [#4941](https://github.com/cosmos/cosmos-sdk/issues/4941) Keybase concrete types constructors such as `NewKeyBaseFromDir` and `NewInMemory` now accept optional parameters of type `KeybaseOption`. These
  optional parameters are also added on the keys sub-commands functions, which are now public, and allows
  these options to be set on the commands or ignored to default to previous behavior.
* [#5547](https://github.com/cosmos/cosmos-sdk/pull/5547) `NewKeyBaseFromHomeFlag` constructor has been removed.
* [#5439](https://github.com/cosmos/cosmos-sdk/pull/5439) Further modularization was done to the `keybase`
  package to make it more suitable for use with different key formats and algorithms:
  _ The `WithKeygenFunc` function added as a `KeybaseOption` which allows a custom bytes to key
  implementation to be defined when keys are created.
  _ The `WithDeriveFunc` function added as a `KeybaseOption` allows custom logic for deriving a key
  from a mnemonic, bip39 password, and HD Path.
  _ BIP44 is no longer build into `keybase.CreateAccount()`. It is however the default when using
  the `client/keys` add command.
  _ `SupportedAlgos` and `SupportedAlgosLedger` functions return a slice of `SigningAlgo`s that are
  supported by the keybase and the ledger integration respectively.
* (simapp) [#5419](https://github.com/cosmos/cosmos-sdk/pull/5419) The `helpers.GenTx()` now accepts a gas argument.
* (baseapp) [#5455](https://github.com/cosmos/cosmos-sdk/issues/5455) A `sdk.Context` is now passed into the `router.Route()` function.

### Client Breaking Changes

* (rest) [#5270](https://github.com/cosmos/cosmos-sdk/issues/5270) All account types now implement custom JSON serialization.
* (rest) [#4783](https://github.com/cosmos/cosmos-sdk/issues/4783) The balance field in the DelegationResponse type is now sdk.Coin instead of sdk.Int
* (x/auth) [#5006](https://github.com/cosmos/cosmos-sdk/pull/5006) The gas required to pass the `AnteHandler` has
  increased significantly due to modular `AnteHandler` support. Increase GasLimit accordingly.
* (rest) [#5336](https://github.com/cosmos/cosmos-sdk/issues/5336) `MsgEditValidator` uses `description` instead of `Description` as a JSON key.
* (keys) [#5097](https://github.com/cosmos/cosmos-sdk/pull/5097) Due to the keybase -> keyring transition, keys need to be migrated. See `keys migrate` command for more info.
* (x/auth) [#5424](https://github.com/cosmos/cosmos-sdk/issues/5424) Drop `decode-tx` command from x/auth/client/cli, duplicate of the `decode` command.

### Features

* (store) [#5435](https://github.com/cosmos/cosmos-sdk/pull/5435) New iterator for paginated requests. Iterator limits DB reads to the range of the requested page.
* (x/evidence) [#5240](https://github.com/cosmos/cosmos-sdk/pull/5240) Initial implementation of the `x/evidence` module.
* (cli) [#5212](https://github.com/cosmos/cosmos-sdk/issues/5212) The `q gov proposals` command now supports pagination.
* (store) [#4724](https://github.com/cosmos/cosmos-sdk/issues/4724) Multistore supports substore migrations upon load. New `rootmulti.Store.LoadLatestVersionAndUpgrade` method in
  `Baseapp` supports `StoreLoader` to enable various upgrade strategies. It no
  longer panics if the store to load contains substores that we didn't explicitly mount.
* [#4972](https://github.com/cosmos/cosmos-sdk/issues/4972) A `TxResponse` with a corresponding code
  and tx hash will be returned for specific Tendermint errors:
  _ `CodeTxInMempoolCache`
  _ `CodeMempoolIsFull` \* `CodeTxTooLarge`
* [#3872](https://github.com/cosmos/cosmos-sdk/issues/3872) Implement a RESTful endpoint and cli command to decode transactions.
* (keys) [#4754](https://github.com/cosmos/cosmos-sdk/pull/4754) Introduce new Keybase implementation that can
  leverage operating systems' built-in functionalities to securely store secrets. MacOS users may encounter
  the following [issue](https://github.com/keybase/go-keychain/issues/47) with the `go-keychain` library. If
  you encounter this issue, you must upgrade your xcode command line tools to version >= `10.2`. You can
  upgrade via: `sudo rm -rf /Library/Developer/CommandLineTools; xcode-select --install`. Verify the
  correct version via: `pkgutil --pkg-info=com.apple.pkg.CLTools_Executables`.
* [#5355](https://github.com/cosmos/cosmos-sdk/pull/5355) Client commands accept a new `--keyring-backend` option through which users can specify which backend should be used
  by the new key store:
  _ `os`: use OS default credentials storage (default).
  _ `file`: use encrypted file-based store.
  _ `kwallet`: use [KDE Wallet](https://utils.kde.org/projects/kwalletmanager/) service.
  _ `pass`: use the [pass](https://www.passwordstore.org/) command line password manager. \* `test`: use password-less key store. *For testing purposes only. Use it at your own risk.*
* (keys) [#5097](https://github.com/cosmos/cosmos-sdk/pull/5097) New `keys migrate` command to assist users migrate their keys
  to the new keyring.
* (keys) [#5366](https://github.com/cosmos/cosmos-sdk/pull/5366) `keys list` now accepts a `--list-names` option to list key names only, whilst the `keys delete`
  command can delete multiple keys by passing their names as arguments. The aforementioned commands can then be piped together, e.g.
  `appcli keys list -n | xargs appcli keys delete`
* (modules) [#4233](https://github.com/cosmos/cosmos-sdk/pull/4233) Add upgrade module that coordinates software upgrades of live chains.
* [#4486](https://github.com/cosmos/cosmos-sdk/issues/4486) Introduce new `PeriodicVestingAccount` vesting account type
  that allows for arbitrary vesting periods.
* (baseapp) [#5196](https://github.com/cosmos/cosmos-sdk/pull/5196) Baseapp has a new `runTxModeReCheck` to allow applications to skip expensive and unnecessary re-checking of transactions.
* (types) [#5196](https://github.com/cosmos/cosmos-sdk/pull/5196) Context has new `IsRecheckTx() bool` and `WithIsReCheckTx(bool) Context` methods to to be used in the `AnteHandler`.
* (x/auth/ante) [#5196](https://github.com/cosmos/cosmos-sdk/pull/5196) AnteDecorators have been updated to avoid unnecessary checks when `ctx.IsReCheckTx() == true`
* (x/auth) [#5006](https://github.com/cosmos/cosmos-sdk/pull/5006) Modular `AnteHandler` via composable decorators:
    * The `AnteDecorator` interface has been introduced to allow users to implement modular `AnteHandler`
    functionality that can be composed together to create a single `AnteHandler` rather than implementing
    a custom `AnteHandler` completely from scratch, where each `AnteDecorator` allows for custom behavior in
    tightly defined and logically isolated manner. These custom `AnteDecorator` can then be chained together
    with default `AnteDecorator` or third-party `AnteDecorator` to create a modularized `AnteHandler`
    which will run each `AnteDecorator` in the order specified in `ChainAnteDecorators`. For details
    on the new architecture, refer to the [ADR](docs/architecture/adr-010-modular-antehandler.md).
    * `ChainAnteDecorators` function has been introduced to take in a list of `AnteDecorators` and chain
    them in sequence and return a single `AnteHandler`:
    _ `SetUpContextDecorator`: Sets `GasMeter` in context and creates defer clause to recover from any
    `OutOfGas` panics in future AnteDecorators and return `OutOfGas` error to `BaseApp`. It MUST be the
    first `AnteDecorator` in the chain for any application that uses gas (or another one that sets the gas meter).
    _ `ValidateBasicDecorator`: Calls tx.ValidateBasic and returns any non-nil error.
    _ `ValidateMemoDecorator`: Validates tx memo with application parameters and returns any non-nil error.
    _ `ConsumeGasTxSizeDecorator`: Consumes gas proportional to the tx size based on application parameters.
    _ `MempoolFeeDecorator`: Checks if fee is above local mempool `minFee` parameter during `CheckTx`.
    _ `DeductFeeDecorator`: Deducts the `FeeAmount` from first signer of the transaction.
    _ `SetPubKeyDecorator`: Sets pubkey of account in any account that does not already have pubkey saved in state machine.
    _ `SigGasConsumeDecorator`: Consume parameter-defined amount of gas for each signature.
    _ `SigVerificationDecorator`: Verify each signature is valid, return if there is an error.
    _ `ValidateSigCountDecorator`: Validate the number of signatures in tx based on app-parameters. \* `IncrementSequenceDecorator`: Increments the account sequence for each signer to prevent replay attacks.
* (cli) [#5223](https://github.com/cosmos/cosmos-sdk/issues/5223) Cosmos Ledger App v2.0.0 is now supported. The changes are backwards compatible and App v1.5.x is still supported.
* (x/staking) [#5380](https://github.com/cosmos/cosmos-sdk/pull/5380) Introduced ability to store historical info entries in staking keeper, allows applications to introspect specified number of past headers and validator sets
    * Introduces new parameter `HistoricalEntries` which allows applications to determine how many recent historical info entries they want to persist in store. Default value is 0.
    * Introduces cli commands and rest routes to query historical information at a given height
* (modules) [#5249](https://github.com/cosmos/cosmos-sdk/pull/5249) Funds are now allowed to be directly sent to the community pool (via the distribution module account).
* (keys) [#4941](https://github.com/cosmos/cosmos-sdk/issues/4941) Introduce keybase option to allow overriding the default private key implementation of a key generated through the `keys add` cli command.
* (keys) [#5439](https://github.com/cosmos/cosmos-sdk/pull/5439) Flags `--algo` and `--hd-path` are added to
  `keys add` command in order to make use of keybase modularized. By default, it uses (0, 0) bip44
  HD path and secp256k1 keys, so is non-breaking.
* (types) [#5447](https://github.com/cosmos/cosmos-sdk/pull/5447) Added `ApproxRoot` function to sdk.Decimal type in order to get the nth root for a decimal number, where n is a positive integer.
    * An `ApproxSqrt` function was also added for convenience around the common case of n=2.

### Improvements

* (iavl) [#5538](https://github.com/cosmos/cosmos-sdk/pull/5538) Remove manual IAVL pruning in favor of IAVL's internal pruning strategy.
* (server) [#4215](https://github.com/cosmos/cosmos-sdk/issues/4215) The `--pruning` flag
  has been moved to the configuration file, to allow easier node configuration.
* (cli) [#5116](https://github.com/cosmos/cosmos-sdk/issues/5116) The `CLIContext` now supports multiple verifiers
  when connecting to multiple chains. The connecting chain's `CLIContext` will have to have the correct
  chain ID and node URI or client set. To use a `CLIContext` with a verifier for another chain:

  ```go
  // main or parent chain (chain as if you're running without IBC)
  mainCtx := context.NewCLIContext()

  // connecting IBC chain
  sideCtx := context.NewCLIContext().
    WithChainID(sideChainID).
    WithNodeURI(sideChainNodeURI) // or .WithClient(...)

  sideCtx = sideCtx.WithVerifier(
    context.CreateVerifier(sideCtx, context.DefaultVerifierCacheSize),
  )
  ```

* (modules) [#5017](https://github.com/cosmos/cosmos-sdk/pull/5017) The `x/auth` package now supports
  generalized genesis accounts through the `GenesisAccount` interface.
* (modules) [#4762](https://github.com/cosmos/cosmos-sdk/issues/4762) Deprecate remove and add permissions in ModuleAccount.
* (modules) [#4760](https://github.com/cosmos/cosmos-sdk/issues/4760) update `x/auth` to match module spec.
* (modules) [#4814](https://github.com/cosmos/cosmos-sdk/issues/4814) Add security contact to Validator description.
* (modules) [#4875](https://github.com/cosmos/cosmos-sdk/issues/4875) refactor integration tests to use SimApp and separate test package
* (sdk) [#4566](https://github.com/cosmos/cosmos-sdk/issues/4566) Export simulation's parameters and app state to JSON in order to reproduce bugs and invariants.
* (sdk) [#4640](https://github.com/cosmos/cosmos-sdk/issues/4640) improve import/export simulation errors by extending `DiffKVStores` to return an array of `KVPairs` that are then compared to check for inconsistencies.
* (sdk) [#4717](https://github.com/cosmos/cosmos-sdk/issues/4717) refactor `x/slashing` to match the new module spec
* (sdk) [#4758](https://github.com/cosmos/cosmos-sdk/issues/4758) update `x/genaccounts` to match module spec
* (simulation) [#4824](https://github.com/cosmos/cosmos-sdk/issues/4824) `PrintAllInvariants` flag will print all failed invariants
* (simulation) [#4490](https://github.com/cosmos/cosmos-sdk/issues/4490) add `InitialBlockHeight` flag to resume a simulation from a given block

    * Support exporting the simulation stats to a given JSON file

* (simulation) [#4847](https://github.com/cosmos/cosmos-sdk/issues/4847), [#4838](https://github.com/cosmos/cosmos-sdk/pull/4838) and [#4869](https://github.com/cosmos/cosmos-sdk/pull/4869) `SimApp` and simulation refactors:
    * Implement `SimulationManager` for executing modules' simulation functionalities in a modularized way
    * Add `RegisterStoreDecoders` to the `SimulationManager` for decoding each module's types
    * Add `GenerateGenesisStates` to the `SimulationManager` to generate a randomized `GenState` for each module
    * Add `RandomizedParams` to the `SimulationManager` that registers each modules' parameters in order to
    simulate `ParamChangeProposal`s' `Content`s
    * Add `WeightedOperations` to the `SimulationManager` that define simulation operations (modules' `Msg`s) with their
    respective weights (i.e chance of being simulated).
    * Add `ProposalContents` to the `SimulationManager` to register each module's governance proposal `Content`s.
* (simulation) [#4893](https://github.com/cosmos/cosmos-sdk/issues/4893) Change `SimApp` keepers to be public and add getter functions for keys and codec
* (simulation) [#4906](https://github.com/cosmos/cosmos-sdk/issues/4906) Add simulation `Config` struct that wraps simulation flags
* (simulation) [#4935](https://github.com/cosmos/cosmos-sdk/issues/4935) Update simulation to reflect a proper `ABCI` application without bypassing `BaseApp` semantics
* (simulation) [#5378](https://github.com/cosmos/cosmos-sdk/pull/5378) Simulation tests refactor:
    * Add `App` interface for general SDK-based app's methods.
    * Refactor and cleanup simulation tests into util functions to simplify their implementation for other SDK apps.
* (store) [#4792](https://github.com/cosmos/cosmos-sdk/issues/4792) panic on non-registered store
* (types) [#4821](https://github.com/cosmos/cosmos-sdk/issues/4821) types/errors package added with support for stacktraces. It is meant as a more feature-rich replacement for sdk.Errors in the mid-term.
* (store) [#1947](https://github.com/cosmos/cosmos-sdk/issues/1947) Implement inter-block (persistent)
  caching through `CommitKVStoreCacheManager`. Any application wishing to utilize an inter-block cache
  must set it in their app via a `BaseApp` option. The `BaseApp` docs have been drastically improved
  to detail this new feature and how state transitions occur.
* (docs/spec) All module specs moved into their respective module dir in x/ (i.e. docs/spec/staking -->> x/staking/spec)
* (docs/) [#5379](https://github.com/cosmos/cosmos-sdk/pull/5379) Major documentation refactor, including:
    * (docs/intro/) Add and improve introduction material for newcomers.
    * (docs/basics/) Add documentation about basic concepts of the cosmos sdk such as the anatomy of an SDK application, the transaction lifecycle or accounts.
    * (docs/core/) Add documentation about core conepts of the cosmos sdk such as `baseapp`, `server`, `store`s, `context` and more.
    * (docs/building-modules/) Add reference documentation on concepts relevant for module developers (`keeper`, `handler`, `messages`, `queries`,...).
    * (docs/interfaces/) Add documentation on building interfaces for the Cosmos SDK.
    * Redesigned user interface that features new dynamically generated sidebar, build-time code embedding from GitHub, new homepage as well as many other improvements.
* (types) [#5428](https://github.com/cosmos/cosmos-sdk/pull/5428) Add `Mod` (modulo) method and `RelativePow` (exponentation) function for `Uint`.
* (modules) [#5506](https://github.com/cosmos/cosmos-sdk/pull/5506) Remove redundancy in `x/distribution`s use of parameters. There
  now exists a single `Params` type with a getter and setter along with a getter for each individual parameter.

### Bug Fixes

* (client) [#5303](https://github.com/cosmos/cosmos-sdk/issues/5303) Fix ignored error in tx generate only mode.
* (cli) [#4763](https://github.com/cosmos/cosmos-sdk/issues/4763) Fix flag `--min-self-delegation` for staking `EditValidator`
* (keys) Fix ledger custom coin type support bug.
* (x/gov) [#5107](https://github.com/cosmos/cosmos-sdk/pull/5107) Sum validator operator's all voting power when tally votes
* (rest) [#5212](https://github.com/cosmos/cosmos-sdk/issues/5212) Fix pagination in the `/gov/proposals` handler.

## [v0.37.14] - 2020-08-12

### Improvements

* (tendermint) Bump Tendermint version to [v0.32.13](https://github.com/tendermint/tendermint/releases/tag/v0.32.13).

## [v0.37.13] - 2020-06-03

### Improvements

* (tendermint) Bump Tendermint version to [v0.32.12](https://github.com/tendermint/tendermint/releases/tag/v0.32.12).
* (cosmos-ledger-go) Bump Cosmos Ledger Wallet library version to [v0.11.1](https://github.com/cosmos/ledger-cosmos-go/releases/tag/v0.11.1).

## [v0.37.12] - 2020-05-05

### Improvements

* (tendermint) Bump Tendermint version to [v0.32.11](https://github.com/tendermint/tendermint/releases/tag/v0.32.11).

## [v0.37.11] - 2020-04-22

### Bug Fixes

* (x/staking) [#6021](https://github.com/cosmos/cosmos-sdk/pull/6021) --trust-node's false default value prevents creation of the genesis transaction.

## [v0.37.10] - 2020-04-22

### Bug Fixes

* (client/context) [#5964](https://github.com/cosmos/cosmos-sdk/issues/5964) Fix incorrect instantiation of tmlite verifier when --trust-node is off.

## [v0.37.9] - 2020-04-09

### Improvements

* (tendermint) Bump Tendermint version to [v0.32.10](https://github.com/tendermint/tendermint/releases/tag/v0.32.10).

## [v0.37.8] - 2020-03-11

### Bug Fixes

* (rest) [#5508](https://github.com/cosmos/cosmos-sdk/pull/5508) Fix `x/distribution` endpoints to properly return height in the response.
* (x/genutil) [#5499](https://github.com/cosmos/cosmos-sdk/pull/) Ensure `DefaultGenesis` returns valid and non-nil default genesis state.
* (x/genutil) [#5775](https://github.com/cosmos/cosmos-sdk/pull/5775) Fix `ExportGenesis` in `x/genutil` to export default genesis state (`[]`) instead of `null`.
* (genesis) [#5086](https://github.com/cosmos/cosmos-sdk/issues/5086) Ensure `gentxs` are always an empty array instead of `nil`.

### Improvements

* (rest) [#5648](https://github.com/cosmos/cosmos-sdk/pull/5648) Enhance /txs usability:
    * Add `tx.minheight` key to filter transaction with an inclusive minimum block height
    * Add `tx.maxheight` key to filter transaction with an inclusive maximum block height

## [v0.37.7] - 2020-02-10

### Improvements

* (modules) [#5597](https://github.com/cosmos/cosmos-sdk/pull/5597) Add `amount` event attribute to the `complete_unbonding`
  and `complete_redelegation` events that reflect the total balances of the completed unbondings and redelegations
  respectively.

### Bug Fixes

* (x/gov) [#5622](https://github.com/cosmos/cosmos-sdk/pull/5622) Track any events emitted from a proposal's handler upon successful execution.
* (x/bank) [#5531](https://github.com/cosmos/cosmos-sdk/issues/5531) Added missing amount event to MsgMultiSend, emitted for each output.

## [v0.37.6] - 2020-01-21

### Improvements

* (tendermint) Bump Tendermint version to [v0.32.9](https://github.com/tendermint/tendermint/releases/tag/v0.32.9)

## [v0.37.5] - 2020-01-07

### Features

* (types) [#5360](https://github.com/cosmos/cosmos-sdk/pull/5360) Implement `SortableDecBytes` which
  allows the `Dec` type be sortable.

### Improvements

* (tendermint) Bump Tendermint version to [v0.32.8](https://github.com/tendermint/tendermint/releases/tag/v0.32.8)
* (cli) [#5482](https://github.com/cosmos/cosmos-sdk/pull/5482) Remove old "tags" nomenclature from the `q txs` command in
  favor of the new events system. Functionality remains unchanged except that `=` is used instead of `:` to be
  consistent with the API's use of event queries.

### Bug Fixes

* (iavl) [#5276](https://github.com/cosmos/cosmos-sdk/issues/5276) Fix potential race condition in `iavlIterator#Close`.
* (baseapp) [#5350](https://github.com/cosmos/cosmos-sdk/issues/5350) Allow a node to restart successfully
  after a `halt-height` or `halt-time` has been triggered.
* (types) [#5395](https://github.com/cosmos/cosmos-sdk/issues/5395) Fix `Uint#LTE`.
* (types) [#5408](https://github.com/cosmos/cosmos-sdk/issues/5408) `NewDecCoins` constructor now sorts the coins.

## [v0.37.4] - 2019-11-04

### Improvements

* (tendermint) Bump Tendermint version to [v0.32.7](https://github.com/tendermint/tendermint/releases/tag/v0.32.7)
* (ledger) [#4716](https://github.com/cosmos/cosmos-sdk/pull/4716) Fix ledger custom coin type support bug.

### Bug Fixes

* (baseapp) [#5200](https://github.com/cosmos/cosmos-sdk/issues/5200) Remove duplicate events from previous messages.

## [v0.37.3] - 2019-10-10

### Bug Fixes

* (genesis) [#5095](https://github.com/cosmos/cosmos-sdk/issues/5095) Fix genesis file migration from v0.34 to
  v0.36/v0.37 not converting validator consensus pubkey to bech32 format.

### Improvements

* (tendermint) Bump Tendermint version to [v0.32.6](https://github.com/tendermint/tendermint/releases/tag/v0.32.6)

## [v0.37.1] - 2019-09-19

### Features

* (cli) [#4973](https://github.com/cosmos/cosmos-sdk/pull/4973) Enable application CPU profiling
  via the `--cpu-profile` flag.
* [#4979](https://github.com/cosmos/cosmos-sdk/issues/4979) Introduce a new `halt-time` config and
  CLI option to the `start` command. When provided, an application will halt during `Commit` when the
  block time is >= the `halt-time`.

### Improvements

* [#4990](https://github.com/cosmos/cosmos-sdk/issues/4990) Add `Events` to the `ABCIMessageLog` to
  provide context and grouping of events based on the messages they correspond to. The `Events` field
  in `TxResponse` is deprecated and will be removed in the next major release.

### Bug Fixes

* [#4979](https://github.com/cosmos/cosmos-sdk/issues/4979) Use `Signal(os.Interrupt)` over
  `os.Exit(0)` during configured halting to allow any `defer` calls to be executed.
* [#5034](https://github.com/cosmos/cosmos-sdk/issues/5034) Binary search in NFT Module wasn't working on larger sets.

## [v0.37.0] - 2019-08-21

### Bug Fixes

* (baseapp) [#4903](https://github.com/cosmos/cosmos-sdk/issues/4903) Various height query fixes:
    * Move height with proof check from `CLIContext` to `BaseApp` as the height
    can automatically be injected there.
    * Update `handleQueryStore` to resemble `handleQueryCustom`
* (simulation) [#4912](https://github.com/cosmos/cosmos-sdk/issues/4912) Fix SimApp ModuleAccountAddrs
  to properly return black listed addresses for bank keeper initialization.
* (cli) [#4919](https://github.com/cosmos/cosmos-sdk/pull/4919) Don't crash CLI
  if user doesn't answer y/n confirmation request.
* (cli) [#4927](https://github.com/cosmos/cosmos-sdk/issues/4927) Fix the `q gov vote`
  command to handle empty (pruned) votes correctly.

### Improvements

* (rest) [#4924](https://github.com/cosmos/cosmos-sdk/pull/4924) Return response
  height even upon error as it may be useful for the downstream caller and have
  `/auth/accounts/{address}` return a 200 with an empty account upon error when
  that error is that the account doesn't exist.

## [v0.36.0] - 2019-08-13

### Breaking Changes

* (rest) [#4837](https://github.com/cosmos/cosmos-sdk/pull/4837) Remove /version and /node_version
  endpoints in favor of refactoring /node_info to also include application version info.
* All REST responses now wrap the original resource/result. The response
  will contain two fields: height and result.
* [#3565](https://github.com/cosmos/cosmos-sdk/issues/3565) Updates to the governance module:
    * Rename JSON field from `proposal_content` to `content`
    * Rename JSON field from `proposal_id` to `id`
    * Disable `ProposalTypeSoftwareUpgrade` temporarily
* [#3775](https://github.com/cosmos/cosmos-sdk/issues/3775) unify sender transaction tag for ease of querying
* [#4255](https://github.com/cosmos/cosmos-sdk/issues/4255) Add supply module that passively tracks the supplies of a chain
    * Renamed `x/distribution` `ModuleName`
    * Genesis JSON and CLI now use `distribution` instead of `distr`
    * Introduce `ModuleAccount` type, which tracks the flow of coins held within a module
    * Replaced `FeeCollectorKeeper` for a `ModuleAccount`
    * Replaced the staking `Pool`, which coins are now held by the `BondedPool` and `NotBonded` module accounts
    * The `NotBonded` module account now only keeps track of the not bonded tokens within staking, instead of the whole chain
    * [#3628](https://github.com/cosmos/cosmos-sdk/issues/3628) Replaced governance's burn and deposit accounts for a `ModuleAccount`
    * Added a `ModuleAccount` for the distribution module
    * Added a `ModuleAccount` for the mint module
    [#4472](https://github.com/cosmos/cosmos-sdk/issues/4472) validation for crisis genesis
* [#3985](https://github.com/cosmos/cosmos-sdk/issues/3985) `ValidatorPowerRank` uses potential consensus power instead of tendermint power
* [#4104](https://github.com/cosmos/cosmos-sdk/issues/4104) Gaia has been moved to its own repository: https://github.com/cosmos/gaia
* [#4104](https://github.com/cosmos/cosmos-sdk/issues/4104) Rename gaiad.toml to app.toml. The internal contents of the application
  config remain unchanged.
* [#4159](https://github.com/cosmos/cosmos-sdk/issues/4159) create the default module patterns and module manager
* [#4230](https://github.com/cosmos/cosmos-sdk/issues/4230) Change the type of ABCIMessageLog#MsgIndex to uint16 for proper serialization.
* [#4250](https://github.com/cosmos/cosmos-sdk/issues/4250) BaseApp.Query() returns app's version string set via BaseApp.SetAppVersion()
  when handling /app/version queries instead of the version string passed as build
  flag at compile time.
* [#4262](https://github.com/cosmos/cosmos-sdk/issues/4262) GoSumHash is no longer returned by the version command.
* [#4263](https://github.com/cosmos/cosmos-sdk/issues/4263) RestServer#Start now takes read and write timeout arguments.
* [#4305](https://github.com/cosmos/cosmos-sdk/issues/4305) `GenerateOrBroadcastMsgs` no longer takes an `offline` parameter.
* [#4342](https://github.com/cosmos/cosmos-sdk/pull/4342) Upgrade go-amino to v0.15.0
* [#4351](https://github.com/cosmos/cosmos-sdk/issues/4351) InitCmd, AddGenesisAccountCmd, and CollectGenTxsCmd take node's and client's default home directories as arguments.
* [#4387](https://github.com/cosmos/cosmos-sdk/issues/4387) Refactor the usage of tags (now called events) to reflect the
  new ABCI events semantics:
    * Move `x/{module}/tags/tags.go` => `x/{module}/types/events.go`
    * Update `docs/specs`
    * Refactor tags in favor of new `Event(s)` type(s)
    * Update `Context` to use new `EventManager`
    * (Begin|End)Blocker no longer return tags, but rather uses new `EventManager`
    * Message handlers no longer return tags, but rather uses new `EventManager`
    Any component (e.g. BeginBlocker, message handler, etc...) wishing to emit an event must do so
    through `ctx.EventManger().EmitEvent(s)`.
    To reset or wipe emitted events: `ctx = ctx.WithEventManager(sdk.NewEventManager())`
    To get all emitted events: `events := ctx.EventManager().Events()`
* [#4437](https://github.com/cosmos/cosmos-sdk/issues/4437) Replace governance module store keys to use `[]byte` instead of `string`.
* [#4451](https://github.com/cosmos/cosmos-sdk/issues/4451) Improve modularization of clients and modules:
    * Module directory structure improved and standardized
    * Aliases autogenerated
    * Auth and bank related commands are now mounted under the respective moduels
    * Client initialization and mounting standardized
* [#4479](https://github.com/cosmos/cosmos-sdk/issues/4479) Remove codec argument redundency in client usage where
  the CLIContext's codec should be used instead.
* [#4488](https://github.com/cosmos/cosmos-sdk/issues/4488) Decouple client tx, REST, and ultil packages from auth. These packages have
  been restructured and retrofitted into the `x/auth` module.
* [#4521](https://github.com/cosmos/cosmos-sdk/issues/4521) Flatten x/bank structure by hiding module internals.
* [#4525](https://github.com/cosmos/cosmos-sdk/issues/4525) Remove --cors flag, the feature is long gone.
* [#4536](https://github.com/cosmos/cosmos-sdk/issues/4536) The `/auth/accounts/{address}` now returns a `height` in the response.
  The account is now nested under `account`.
* [#4543](https://github.com/cosmos/cosmos-sdk/issues/4543) Account getters are no longer part of client.CLIContext() and have now moved
  to reside in the auth-specific AccountRetriever.
* [#4588](https://github.com/cosmos/cosmos-sdk/issues/4588) Context does not depend on x/auth anymore. client/context is stripped out of the following features:
    * GetAccountDecoder()
    * CLIContext.WithAccountDecoder()
    * CLIContext.WithAccountStore()
    x/auth.AccountDecoder is unnecessary and consequently removed.
* [#4602](https://github.com/cosmos/cosmos-sdk/issues/4602) client/input.{Buffer,Override}Stdin() functions are removed. Thanks to cobra's new release they are now redundant.
* [#4633](https://github.com/cosmos/cosmos-sdk/issues/4633) Update old Tx search by tags APIs to use new Events
  nomenclature.
* [#4649](https://github.com/cosmos/cosmos-sdk/issues/4649) Refactor x/crisis as per modules new specs.
* [#3685](https://github.com/cosmos/cosmos-sdk/issues/3685) The default signature verification gas logic (`DefaultSigVerificationGasConsumer`) now specifies explicit key types rather than string pattern matching. This means that zones that depended on string matching to allow other keys will need to write a custom `SignatureVerificationGasConsumer` function.
* [#4663](https://github.com/cosmos/cosmos-sdk/issues/4663) Refactor bank keeper by removing private functions
    * `InputOutputCoins`, `SetCoins`, `SubtractCoins` and `AddCoins` are now part of the `SendKeeper` instead of the `Keeper` interface
* (tendermint) [#4721](https://github.com/cosmos/cosmos-sdk/pull/4721) Upgrade Tendermint to v0.32.1

### Features

* [#4843](https://github.com/cosmos/cosmos-sdk/issues/4843) Add RegisterEvidences function in the codec package to register
  Tendermint evidence types with a given codec.
* (rest) [#3867](https://github.com/cosmos/cosmos-sdk/issues/3867) Allow querying for genesis transaction when height query param is set to zero.
* [#2020](https://github.com/cosmos/cosmos-sdk/issues/2020) New keys export/import command line utilities to export/import private keys in ASCII format
  that rely on Keybase's new underlying ExportPrivKey()/ImportPrivKey() API calls.
* [#3565](https://github.com/cosmos/cosmos-sdk/issues/3565) Implement parameter change proposal support.
  Parameter change proposals can be submitted through the CLI
  or a REST endpoint. See docs for further usage.
* [#3850](https://github.com/cosmos/cosmos-sdk/issues/3850) Add `rewards` and `commission` to distribution tx tags.
* [#3981](https://github.com/cosmos/cosmos-sdk/issues/3981) Add support to gracefully halt a node at a given height
  via the node's `halt-height` config or CLI value.
* [#4144](https://github.com/cosmos/cosmos-sdk/issues/4144) Allow for configurable BIP44 HD path and coin type.
* [#4250](https://github.com/cosmos/cosmos-sdk/issues/4250) New BaseApp.{,Set}AppVersion() methods to get/set app's version string.
* [#4263](https://github.com/cosmos/cosmos-sdk/issues/4263) Add `--read-timeout` and `--write-timeout` args to the `rest-server` command
  to support custom RPC R/W timeouts.
* [#4271](https://github.com/cosmos/cosmos-sdk/issues/4271) Implement Coins#IsAnyGT
* [#4318](https://github.com/cosmos/cosmos-sdk/issues/4318) Support height queries. Queries against nodes that have the queried
  height pruned will return an error.
* [#4409](https://github.com/cosmos/cosmos-sdk/issues/4409) Implement a command that migrates exported state from one version to the next.
  The `migrate` command currently supports migrating from v0.34 to v0.36 by implementing
  necessary types for both versions.
* [#4570](https://github.com/cosmos/cosmos-sdk/issues/4570) Move /bank/balances/{address} REST handler to x/bank/client/rest. The exposed interface is unchanged.
* Community pool spend proposal per Cosmos Hub governance proposal [#7](https://github.com/cosmos/cosmos-sdk/issues/7) "Activate the Community Pool"

### Improvements

* (simulation) PrintAllInvariants flag will print all failed invariants
* (simulation) Add `InitialBlockHeight` flag to resume a simulation from a given block
* (simulation) [#4670](https://github.com/cosmos/cosmos-sdk/issues/4670) Update simulation statistics to JSON format

    * Support exporting the simulation stats to a given JSON file

* [#4775](https://github.com/cosmos/cosmos-sdk/issues/4775) Refactor CI config
* Upgrade IAVL to v0.12.4
* (tendermint) Upgrade Tendermint to v0.32.2
* (modules) [#4751](https://github.com/cosmos/cosmos-sdk/issues/4751) update `x/genutils` to match module spec
* (keys) [#4611](https://github.com/cosmos/cosmos-sdk/issues/4611) store keys in simapp now use a map instead of using individual literal keys
* [#2286](https://github.com/cosmos/cosmos-sdk/issues/2286) Improve performance of CacheKVStore iterator.
* [#3512](https://github.com/cosmos/cosmos-sdk/issues/3512) Implement Logger method on each module's keeper.
* [#3655](https://github.com/cosmos/cosmos-sdk/issues/3655) Improve signature verification failure error message.
* [#3774](https://github.com/cosmos/cosmos-sdk/issues/3774) add category tag to transactions for ease of filtering
* [#3914](https://github.com/cosmos/cosmos-sdk/issues/3914) Implement invariant benchmarks and add target to makefile.
* [#3928](https://github.com/cosmos/cosmos-sdk/issues/3928) remove staking references from types package
* [#3978](https://github.com/cosmos/cosmos-sdk/issues/3978) Return ErrUnknownRequest in message handlers for unknown
  or invalid routed messages.
* [#4190](https://github.com/cosmos/cosmos-sdk/issues/4190) Client responses that return (re)delegation(s) now return balances
  instead of shares.
* [#4194](https://github.com/cosmos/cosmos-sdk/issues/4194) ValidatorSigningInfo now includes the validator's consensus address.
* [#4235](https://github.com/cosmos/cosmos-sdk/issues/4235) Add parameter change proposal messages to simulation.
* [#4235](https://github.com/cosmos/cosmos-sdk/issues/4235) Update the minting module params to implement params.ParamSet so
  individual keys can be set via proposals instead of passing a struct.
* [#4259](https://github.com/cosmos/cosmos-sdk/issues/4259) `Coins` that are `nil` are now JSON encoded as an empty array `[]`.
  Decoding remains unchanged and behavior is left intact.
* [#4305](https://github.com/cosmos/cosmos-sdk/issues/4305) The `--generate-only` CLI flag fully respects offline tx processing.
* [#4379](https://github.com/cosmos/cosmos-sdk/issues/4379) close db write batch.
* [#4384](https://github.com/cosmos/cosmos-sdk/issues/4384)- Allow splitting withdrawal transaction in several chunks
* [#4403](https://github.com/cosmos/cosmos-sdk/issues/4403) Allow for parameter change proposals to supply only desired fields to be updated
  in objects instead of the entire object (only applies to values that are objects).
* [#4415](https://github.com/cosmos/cosmos-sdk/issues/4415) /client refactor, reduce genutil dependancy on staking
* [#4439](https://github.com/cosmos/cosmos-sdk/issues/4439) Implement governance module iterators.
* [#4465](https://github.com/cosmos/cosmos-sdk/issues/4465) Unknown subcommands print relevant error message
* [#4466](https://github.com/cosmos/cosmos-sdk/issues/4466) Commission validation added to validate basic of MsgCreateValidator by changing CommissionMsg to CommissionRates
* [#4501](https://github.com/cosmos/cosmos-sdk/issues/4501) Support height queriers in rest client
* [#4535](https://github.com/cosmos/cosmos-sdk/issues/4535) Improve import-export simulation errors by decoding the `KVPair.Value` into its
  respective type
* [#4536](https://github.com/cosmos/cosmos-sdk/issues/4536) cli context queries return query height and accounts are returned with query height
* [#4553](https://github.com/cosmos/cosmos-sdk/issues/4553) undelegate max entries check first
* [#4556](https://github.com/cosmos/cosmos-sdk/issues/4556) Added IsValid function to Coin
* [#4564](https://github.com/cosmos/cosmos-sdk/issues/4564) client/input.GetConfirmation()'s default is changed to No.
* [#4573](https://github.com/cosmos/cosmos-sdk/issues/4573) Returns height in response for query endpoints.
* [#4580](https://github.com/cosmos/cosmos-sdk/issues/4580) Update `Context#BlockHeight` to properly set the block height via `WithBlockHeader`.
* [#4584](https://github.com/cosmos/cosmos-sdk/issues/4584) Update bank Keeper to use expected keeper interface of the AccountKeeper.
* [#4584](https://github.com/cosmos/cosmos-sdk/issues/4584) Move `Account` and `VestingAccount` interface types to `x/auth/exported`.
* [#4082](https://github.com/cosmos/cosmos-sdk/issues/4082) supply module queriers for CLI and REST endpoints
* [#4601](https://github.com/cosmos/cosmos-sdk/issues/4601) Implement generic pangination helper function to be used in
  REST handlers and queriers.
* [#4629](https://github.com/cosmos/cosmos-sdk/issues/4629) Added warning event that gets emitted if validator misses a block.
* [#4674](https://github.com/cosmos/cosmos-sdk/issues/4674) Export `Simapp` genState generators and util functions by making them public
* [#4706](https://github.com/cosmos/cosmos-sdk/issues/4706) Simplify context
  Replace complex Context construct with a simpler immutible struct.
  Only breaking change is not to support `Value` and `GetValue` as first class calls.
  We do embed ctx.Context() as a raw context.Context instead to be used as you see fit.

  Migration guide:

  ```go
  ctx = ctx.WithValue(contextKeyBadProposal, false)
  ```

  Now becomes:

  ```go
  ctx = ctx.WithContext(context.WithValue(ctx.Context(), contextKeyBadProposal, false))
  ```

  A bit more verbose, but also allows `context.WithTimeout()`, etc and only used
  in one function in this repo, in test code.

* [#3685](https://github.com/cosmos/cosmos-sdk/issues/3685) Add `SetAddressVerifier` and `GetAddressVerifier` to `sdk.Config` to allow SDK users to configure custom address format verification logic (to override the default limitation of 20-byte addresses).
* [#3685](https://github.com/cosmos/cosmos-sdk/issues/3685) Add an additional parameter to NewAnteHandler for a custom `SignatureVerificationGasConsumer` (the default logic is now in `DefaultSigVerificationGasConsumer). This allows SDK users to configure their own logic for which key types are accepted and how those key types consume gas.
* Remove `--print-response` flag as it is no longer used.
* Revert [#2284](https://github.com/cosmos/cosmos-sdk/pull/2284) to allow create_empty_blocks in the config
* (tendermint) [#4718](https://github.com/cosmos/cosmos-sdk/issues/4718) Upgrade tendermint/iavl to v0.12.3

### Bug Fixes

* [#4891](https://github.com/cosmos/cosmos-sdk/issues/4891) Disable querying with proofs enabled when the query height <= 1.
* (rest) [#4858](https://github.com/cosmos/cosmos-sdk/issues/4858) Do not return an error in BroadcastTxCommit when the tx broadcasting
  was successful. This allows the proper REST response to be returned for a
  failed tx during `block` broadcasting mode.
* (store) [#4880](https://github.com/cosmos/cosmos-sdk/pull/4880) Fix error check in
  IAVL `Store#DeleteVersion`.
* (tendermint) [#4879](https://github.com/cosmos/cosmos-sdk/issues/4879) Don't terminate the process immediately after startup when run in standalone mode.
* (simulation) [#4861](https://github.com/cosmos/cosmos-sdk/pull/4861) Fix non-determinism simulation
  by using CLI flags as input and updating Makefile target.
* [#4868](https://github.com/cosmos/cosmos-sdk/issues/4868) Context#CacheContext now sets a new EventManager. This prevents unwanted events
  from being emitted.
* (cli) [#4870](https://github.com/cosmos/cosmos-sdk/issues/4870) Disable the `withdraw-all-rewards` command when `--generate-only` is supplied
* (modules) [#4831](https://github.com/cosmos/cosmos-sdk/issues/4831) Prevent community spend proposal from transferring funds to a module account
* (keys) [#4338](https://github.com/cosmos/cosmos-sdk/issues/4338) fix multisig key output for CLI
* (modules) [#4795](https://github.com/cosmos/cosmos-sdk/issues/4795) restrict module accounts from receiving transactions.
  Allowing this would cause an invariant on the module account coins.
* (modules) [#4823](https://github.com/cosmos/cosmos-sdk/issues/4823) Update the `DefaultUnbondingTime` from 3 days to 3 weeks to be inline with documentation.
* (abci) [#4639](https://github.com/cosmos/cosmos-sdk/issues/4639) Fix `CheckTx` by verifying the message route
* Return height in responses when querying against BaseApp
* [#1351](https://github.com/cosmos/cosmos-sdk/issues/1351) Stable AppHash allows no_empty_blocks
* [#3705](https://github.com/cosmos/cosmos-sdk/issues/3705) Return `[]` instead of `null` when querying delegator rewards.
* [#3966](https://github.com/cosmos/cosmos-sdk/issues/3966) fixed multiple assigns to action tags
  [#3793](https://github.com/cosmos/cosmos-sdk/issues/3793) add delegator tag for MsgCreateValidator and deleted unused moniker and identity tags
* [#4194](https://github.com/cosmos/cosmos-sdk/issues/4194) Fix pagination and results returned from /slashing/signing_infos
* [#4230](https://github.com/cosmos/cosmos-sdk/issues/4230) Properly set and display the message index through the TxResponse.
* [#4234](https://github.com/cosmos/cosmos-sdk/pull/4234) Allow `tx send --generate-only` to
  actually work offline.
* [#4271](https://github.com/cosmos/cosmos-sdk/issues/4271) Fix addGenesisAccount by using Coins#IsAnyGT for vesting amount validation.
* [#4273](https://github.com/cosmos/cosmos-sdk/issues/4273) Fix usage of AppendTags in x/staking/handler.go
* [#4303](https://github.com/cosmos/cosmos-sdk/issues/4303) Fix NewCoins() underlying function for duplicate coins detection.
* [#4307](https://github.com/cosmos/cosmos-sdk/pull/4307) Don't pass height to RPC calls as
  Tendermint will automatically use the latest height.
* [#4362](https://github.com/cosmos/cosmos-sdk/issues/4362) simulation setup bugfix for multisim 7601778
* [#4383](https://github.com/cosmos/cosmos-sdk/issues/4383) - currentStakeRoundUp is now always atleast currentStake + smallest-decimal-precision
* [#4394](https://github.com/cosmos/cosmos-sdk/issues/4394) Fix signature count check to use the TxSigLimit param instead of
  a default.
* [#4455](https://github.com/cosmos/cosmos-sdk/issues/4455) Use `QueryWithData()` to query unbonding delegations.
* [#4493](https://github.com/cosmos/cosmos-sdk/issues/4493) Fix validator-outstanding-rewards command. It now takes as an argument
  a validator address.
* [#4598](https://github.com/cosmos/cosmos-sdk/issues/4598) Fix redelegation and undelegation txs that were not checking for the correct bond denomination.
* [#4619](https://github.com/cosmos/cosmos-sdk/issues/4619) Close iterators in `GetAllMatureValidatorQueue` and `UnbondAllMatureValidatorQueue`
  methods.
* [#4654](https://github.com/cosmos/cosmos-sdk/issues/4654) validator slash event stored by period and height
* [#4681](https://github.com/cosmos/cosmos-sdk/issues/4681) panic on invalid amount on `MintCoins` and `BurnCoins`
    * skip minting if inflation is set to zero
* Sort state JSON during export and initialization

## 0.35.0

### Bug Fixes

* Fix gas consumption bug in `Undelegate` preventing the ability to sync from
  genesis.

## 0.34.10

### Bug Fixes

* Bump Tendermint version to [v0.31.11](https://github.com/tendermint/tendermint/releases/tag/v0.31.11) to address the vulnerability found in the `consensus` package.

## 0.34.9

### Bug Fixes

* Bump Tendermint version to [v0.31.10](https://github.com/tendermint/tendermint/releases/tag/v0.31.10) to address p2p panic errors.

## 0.34.8

### Bug Fixes

* Bump Tendermint version to v0.31.9 to fix the p2p panic error.
* Update gaiareplay's use of an internal Tendermint API

## 0.34.7

### Bug Fixes

#### SDK

* Fix gas consumption bug in `Undelegate` preventing the ability to sync from
  genesis.

## 0.34.6

### Bug Fixes

#### SDK

* Unbonding from a validator is now only considered "complete" after the full
  unbonding period has elapsed regardless of the validator's status.

## 0.34.5

### Bug Fixes

#### SDK

* [#4273](https://github.com/cosmos/cosmos-sdk/issues/4273) Fix usage of `AppendTags` in x/staking/handler.go

### Improvements

### SDK

* [#2286](https://github.com/cosmos/cosmos-sdk/issues/2286) Improve performance of `CacheKVStore` iterator.
* [#3655](https://github.com/cosmos/cosmos-sdk/issues/3655) Improve signature verification failure error message.
* [#4384](https://github.com/cosmos/cosmos-sdk/issues/4384) Allow splitting withdrawal transaction in several chunks.

#### Gaia CLI

* [#4227](https://github.com/cosmos/cosmos-sdk/issues/4227) Support for Ledger App v1.5.
* [#4345](https://github.com/cosmos/cosmos-sdk/pull/4345) Update `ledger-cosmos-go`
  to v0.10.3.

## 0.34.4

### Bug Fixes

#### SDK

* [#4234](https://github.com/cosmos/cosmos-sdk/pull/4234) Allow `tx send --generate-only` to
  actually work offline.

#### Gaia

* [#4219](https://github.com/cosmos/cosmos-sdk/issues/4219) Return an error when an empty mnemonic is provided during key recovery.

### Improvements

#### Gaia

* [#2007](https://github.com/cosmos/cosmos-sdk/issues/2007) Return 200 status code on empty results

### New features

#### SDK

* [#3850](https://github.com/cosmos/cosmos-sdk/issues/3850) Add `rewards` and `commission` to distribution tx tags.

## 0.34.3

### Bug Fixes

#### Gaia

* [#4196](https://github.com/cosmos/cosmos-sdk/pull/4196) Set default invariant
  check period to zero.

## 0.34.2

### Improvements

#### SDK

* [#4135](https://github.com/cosmos/cosmos-sdk/pull/4135) Add further clarification
  to generate only usage.

### Bug Fixes

#### SDK

* [#4135](https://github.com/cosmos/cosmos-sdk/pull/4135) Fix `NewResponseFormatBroadcastTxCommit`
* [#4053](https://github.com/cosmos/cosmos-sdk/issues/4053) Add `--inv-check-period`
  flag to gaiad to set period at which invariants checks will run.
* [#4099](https://github.com/cosmos/cosmos-sdk/issues/4099) Update the /staking/validators endpoint to support
  status and pagination query flags.

## 0.34.1

### Bug Fixes

#### Gaia

* [#4163](https://github.com/cosmos/cosmos-sdk/pull/4163) Fix v0.33.x export script to port gov data correctly.

## 0.34.0

### Breaking Changes

#### Gaia

* [#3463](https://github.com/cosmos/cosmos-sdk/issues/3463) Revert bank module handler fork (re-enables transfers)
* [#3875](https://github.com/cosmos/cosmos-sdk/issues/3875) Replace `async` flag with `--broadcast-mode` flag where the default
  value is `sync`. The `block` mode should not be used. The REST client now
  uses `mode` parameter instead of the `return` parameter.

#### Gaia CLI

* [#3938](https://github.com/cosmos/cosmos-sdk/issues/3938) Remove REST server's SSL support altogether.

#### SDK

* [#3245](https://github.com/cosmos/cosmos-sdk/issues/3245) Rename validator.GetJailed() to validator.IsJailed()
* [#3516](https://github.com/cosmos/cosmos-sdk/issues/3516) Remove concept of shares from staking unbonding and redelegation UX;
  replaced by direct coin amount.

#### Tendermint

* [#4029](https://github.com/cosmos/cosmos-sdk/issues/4029) Upgrade Tendermint to v0.31.3

### New features

#### SDK

* [#2935](https://github.com/cosmos/cosmos-sdk/issues/2935) New module Crisis which can test broken invariant with messages
* [#3813](https://github.com/cosmos/cosmos-sdk/issues/3813) New sdk.NewCoins safe constructor to replace bare sdk.Coins{} declarations.
* [#3858](https://github.com/cosmos/cosmos-sdk/issues/3858) add website, details and identity to gentx cli command
* Implement coin conversion and denomination registration utilities

#### Gaia

* [#2935](https://github.com/cosmos/cosmos-sdk/issues/2935) Optionally assert invariants on a blockly basis using `gaiad --assert-invariants-blockly`
* [#3886](https://github.com/cosmos/cosmos-sdk/issues/3886) Implement minting module querier and CLI/REST clients.

#### Gaia CLI

* [#3937](https://github.com/cosmos/cosmos-sdk/issues/3937) Add command to query community-pool

#### Gaia REST API

* [#3937](https://github.com/cosmos/cosmos-sdk/issues/3937) Add route to fetch community-pool
* [#3949](https://github.com/cosmos/cosmos-sdk/issues/3949) added /slashing/signing_infos to get signing_info for all validators

### Improvements

#### Gaia

* [#3808](https://github.com/cosmos/cosmos-sdk/issues/3808) `gaiad` and `gaiacli` integration tests use ./build/ binaries.
* [#3819](https://github.com/cosmos/cosmos-sdk/issues/3819) Simulation refactor, log output now stored in ~/.gaiad/simulation/
    * Simulation moved to its own module (not a part of mock)
    * Logger type instead of passing function variables everywhere
    * Logger json output (for reloadable simulation running)
    * Cleanup bank simulation messages / remove dup code in bank simulation
    * Simulations saved in `~/.gaiad/simulations/`
    * "Lean" simulation output option to exclude No-ops and !ok functions (`--SimulationLean` flag)
* [#3893](https://github.com/cosmos/cosmos-sdk/issues/3893) Improve `gaiacli tx sign` command
    * Add shorthand flags -a and -s for the account and sequence numbers respectively
    * Mark the account and sequence numbers required during "offline" mode
    * Always do an RPC query for account and sequence number during "online" mode
* [#4018](https://github.com/cosmos/cosmos-sdk/issues/4018) create genesis port script for release v.0.34.0

#### Gaia CLI

* [#3833](https://github.com/cosmos/cosmos-sdk/issues/3833) Modify stake to atom in gaia's doc.
* [#3841](https://github.com/cosmos/cosmos-sdk/issues/3841) Add indent to JSON of `gaiacli keys [add|show|list]`
* [#3859](https://github.com/cosmos/cosmos-sdk/issues/3859) Add newline to echo of `gaiacli keys ...`
* [#3959](https://github.com/cosmos/cosmos-sdk/issues/3959) Improving error messages when signing with ledger devices fails

#### SDK

* [#3238](https://github.com/cosmos/cosmos-sdk/issues/3238) Add block time to tx responses when querying for
  txs by tags or hash.
* [#3752](https://github.com/cosmos/cosmos-sdk/issues/3752) Explanatory docs for minting mechanism (`docs/spec/mint/01_concepts.md`)
* [#3801](https://github.com/cosmos/cosmos-sdk/issues/3801) `baseapp` safety improvements
* [#3820](https://github.com/cosmos/cosmos-sdk/issues/3820) Make Coins.IsAllGT() more robust and consistent.
* [#3828](https://github.com/cosmos/cosmos-sdk/issues/3828) New sdkch tool to maintain changelogs
* [#3864](https://github.com/cosmos/cosmos-sdk/issues/3864) Make Coins.IsAllGTE() more consistent.
* [#3907](https://github.com/cosmos/cosmos-sdk/issues/3907): dep -> go mod migration
    * Drop dep in favor of go modules.
    * Upgrade to Go 1.12.1.
* [#3917](https://github.com/cosmos/cosmos-sdk/issues/3917) Allow arbitrary decreases to validator commission rates.
* [#3937](https://github.com/cosmos/cosmos-sdk/issues/3937) Implement community pool querier.
* [#3940](https://github.com/cosmos/cosmos-sdk/issues/3940) Codespace should be lowercase.
* [#3986](https://github.com/cosmos/cosmos-sdk/issues/3986) Update the Stringer implementation of the Proposal type.
* [#926](https://github.com/cosmos/cosmos-sdk/issues/926) circuit breaker high level explanation
* [#3896](https://github.com/cosmos/cosmos-sdk/issues/3896) Fixed various linters warnings in the context of the gometalinter -> golangci-lint migration
* [#3916](https://github.com/cosmos/cosmos-sdk/issues/3916) Hex encode data in tx responses

### Bug Fixes

#### Gaia

* [#3825](https://github.com/cosmos/cosmos-sdk/issues/3825) Validate genesis before running gentx
* [#3889](https://github.com/cosmos/cosmos-sdk/issues/3889) When `--generate-only` is provided, the Keybase is not used and as a result
  the `--from` value must be a valid Bech32 cosmos address.
* 3974 Fix go env setting in installation.md
* 3996 Change 'make get_tools' to 'make tools' in DOCS_README.md.

#### Gaia CLI

* [#3883](https://github.com/cosmos/cosmos-sdk/issues/3883) Remove Height Flag from CLI Queries
* [#3899](https://github.com/cosmos/cosmos-sdk/issues/3899) Using 'gaiacli config node' breaks ~/config/config.toml

#### SDK

* [#3837](https://github.com/cosmos/cosmos-sdk/issues/3837) Fix `WithdrawValidatorCommission` to properly set the validator's remaining commission.
* [#3870](https://github.com/cosmos/cosmos-sdk/issues/3870) Fix DecCoins#TruncateDecimal to never return zero coins in
  either the truncated coins or the change coins.
* [#3915](https://github.com/cosmos/cosmos-sdk/issues/3915) Remove ';' delimiting support from ParseDecCoins
* [#3977](https://github.com/cosmos/cosmos-sdk/issues/3977) Fix docker image build
* [#4020](https://github.com/cosmos/cosmos-sdk/issues/4020) Fix queryDelegationRewards by returning an error
  when the validator or delegation do not exist.
* [#4050](https://github.com/cosmos/cosmos-sdk/issues/4050) Fix DecCoins APIs
  where rounding or truncation could result in zero decimal coins.
* [#4088](https://github.com/cosmos/cosmos-sdk/issues/4088) Fix `calculateDelegationRewards`
  by accounting for rounding errors when multiplying stake by slashing fractions.

## 0.33.2

### Improvements

#### Tendermint

* Upgrade Tendermint to `v0.31.0-dev0-fix0` which includes critical security fixes.

## 0.33.1

### Bug Fixes

#### Gaia

* [#3999](https://github.com/cosmos/cosmos-sdk/pull/3999) Fix distribution delegation for zero height export bug

## 0.33.0

BREAKING CHANGES

* Gaia REST API

    * [#3641](https://github.com/cosmos/cosmos-sdk/pull/3641) Remove the ability to use a Keybase from the REST API client:
        * `password` and `generate_only` have been removed from the `base_req` object
        * All txs that used to sign or use the Keybase now only generate the tx
        * `keys` routes completely removed
    * [#3692](https://github.com/cosmos/cosmos-sdk/pull/3692) Update tx encoding and broadcasting endpoints:
        * Remove duplicate broadcasting endpoints in favor of POST @ `/txs`
            * The `Tx` field now accepts a `StdTx` and not raw tx bytes
        * Move encoding endpoint to `/txs/encode`

* Gaia

    * [#3787](https://github.com/cosmos/cosmos-sdk/pull/3787) Fork the `x/bank` module into the Gaia application with only a
    modified message handler, where the modified message handler behaves the same as
    the standard `x/bank` message handler except for `MsgMultiSend` that must burn
    exactly 9 atoms and transfer 1 atom, and `MsgSend` is disabled.
    * [#3789](https://github.com/cosmos/cosmos-sdk/pull/3789) Update validator creation flow:
    _ Remove `NewMsgCreateValidatorOnBehalfOf` and corresponding business logic
    _ Ensure the validator address equals the delegator address during
    `MsgCreateValidator#ValidateBasic`

* SDK

    * [#3750](https://github.com/cosmos/cosmos-sdk/issues/3750) Track outstanding rewards per-validator instead of globally,
    and fix the main simulation issue, which was that slashes of
    re-delegations to a validator were not correctly accounted for
    in fee distribution when the redelegation in question had itself
    been slashed (from a fault committed by a different validator)
    in the same BeginBlock. Outstanding rewards are now available
    on a per-validator basis in REST.
    * [#3669](https://github.com/cosmos/cosmos-sdk/pull/3669) Ensure consistency in message naming, codec registration, and JSON
    tags.
    * [#3788](https://github.com/cosmos/cosmos-sdk/pull/3788) Change order of operations for greater accuracy when calculating delegation share token value
    * [#3788](https://github.com/cosmos/cosmos-sdk/pull/3788) DecCoins.Cap -> DecCoins.Intersect
    * [#3666](https://github.com/cosmos/cosmos-sdk/pull/3666) Improve coins denom validation.
    * [#3751](https://github.com/cosmos/cosmos-sdk/pull/3751) Disable (temporarily) support for ED25519 account key pairs.

* Tendermint
    * [#3804] Update to Tendermint `v0.31.0-dev0`

FEATURES

* SDK
    * [#3719](https://github.com/cosmos/cosmos-sdk/issues/3719) DBBackend can now be set at compile time.
    Defaults: goleveldb. Supported: cleveldb.

IMPROVEMENTS

* Gaia REST API

    * Update the `TxResponse` type allowing for the `Logs` result to be JSON decoded automatically.

* Gaia CLI

    * [#3653](https://github.com/cosmos/cosmos-sdk/pull/3653) Prompt user confirmation prior to signing and broadcasting a transaction.
    * [#3670](https://github.com/cosmos/cosmos-sdk/pull/3670) CLI support for showing bech32 addresses in Ledger devices
    * [#3711](https://github.com/cosmos/cosmos-sdk/pull/3711) Update `tx sign` to use `--from` instead of the deprecated `--name`
    CLI flag.
    * [#3738](https://github.com/cosmos/cosmos-sdk/pull/3738) Improve multisig UX:
        * `gaiacli keys show -o json` now includes constituent pubkeys, respective weights and threshold
        * `gaiacli keys show --show-multisig` now displays constituent pubkeys, respective weights and threshold
        * `gaiacli tx sign --validate-signatures` now displays multisig signers with their respective weights
    * [#3730](https://github.com/cosmos/cosmos-sdk/issues/3730) Improve workflow for
    `gaiad gentx` with offline public keys, by outputting stdtx file that needs to be signed.
    * [#3761](https://github.com/cosmos/cosmos-sdk/issues/3761) Querying account related information using custom querier in auth module

* SDK

    * [#3753](https://github.com/cosmos/cosmos-sdk/issues/3753) Remove no-longer-used governance penalty parameter
    * [#3679](https://github.com/cosmos/cosmos-sdk/issues/3679) Consistent operators across Coins, DecCoins, Int, Dec
    replaced: Minus->Sub Plus->Add Div->Quo
    * [#3665](https://github.com/cosmos/cosmos-sdk/pull/3665) Overhaul sdk.Uint type in preparation for Coins Int -> Uint migration.
    * [#3691](https://github.com/cosmos/cosmos-sdk/issues/3691) Cleanup error messages
    * [#3456](https://github.com/cosmos/cosmos-sdk/issues/3456) Integrate in the Int.ToDec() convenience function
    * [#3300](https://github.com/cosmos/cosmos-sdk/pull/3300) Update the spec-spec, spec file reorg, and TOC updates.
    * [#3694](https://github.com/cosmos/cosmos-sdk/pull/3694) Push tagged docker images on docker hub when tag is created.
    * [#3716](https://github.com/cosmos/cosmos-sdk/pull/3716) Update file permissions the client keys directory and contents to `0700`.
    * [#3681](https://github.com/cosmos/cosmos-sdk/issues/3681) Migrate ledger-cosmos-go from ZondaX to Cosmos organization

* Tendermint
    * [#3699](https://github.com/cosmos/cosmos-sdk/pull/3699) Upgrade to Tendermint 0.30.1

BUG FIXES

* Gaia CLI

    * [#3731](https://github.com/cosmos/cosmos-sdk/pull/3731) `keys add --interactive` bip32 passphrase regression fix
    * [#3714](https://github.com/cosmos/cosmos-sdk/issues/3714) Fix USB raw access issues with gaiacli when installed via snap

* Gaia

    * [#3777](https://github.com/cosmso/cosmos-sdk/pull/3777) `gaiad export` no longer panics when the database is empty
    * [#3806](https://github.com/cosmos/cosmos-sdk/pull/3806) Properly return errors from a couple of struct Unmarshal functions

* SDK
    * [#3728](https://github.com/cosmos/cosmos-sdk/issues/3728) Truncate decimal multiplication & division in distribution to ensure
    no more than the collected fees / inflation are distributed
    * [#3727](https://github.com/cosmos/cosmos-sdk/issues/3727) Return on zero-length (including []byte{}) PrefixEndBytes() calls
    * [#3559](https://github.com/cosmos/cosmos-sdk/issues/3559) fix occasional failing due to non-determinism in lcd test TestBonding
    where validator is unexpectedly slashed throwing off test calculations
    * [#3411](https://github.com/cosmos/cosmos-sdk/pull/3411) Include the `RequestInitChain.Time` in the block header init during
    `InitChain`.
    * [#3717](https://github.com/cosmos/cosmos-sdk/pull/3717) Update the vesting specification and implementation to cap deduction from
    `DelegatedVesting` by at most `DelegatedVesting`. This accounts for the case where
    the undelegation amount may exceed the original delegation amount due to
    truncation of undelegation tokens.
    * [#3717](https://github.com/cosmos/cosmos-sdk/pull/3717) Ignore unknown proposers in allocating rewards for proposers, in case
    unbonding period was just 1 block and proposer was already deleted.
    * [#3726](https://github.com/cosmos/cosmos-sdk/pull/3724) Cap(clip) reward to remaining coins in AllocateTokens.

## 0.32.0

BREAKING CHANGES

* Gaia REST API

    * [#3642](https://github.com/cosmos/cosmos-sdk/pull/3642) `GET /tx/{hash}` now returns `404` instead of `500` if the transaction is not found

* SDK
* [#3580](https://github.com/cosmos/cosmos-sdk/issues/3580) Migrate HTTP request/response types and utilities to types/rest.
* [#3592](https://github.com/cosmos/cosmos-sdk/issues/3592) Drop deprecated keybase implementation's New() constructor in
  favor of a new crypto/keys.New(string, string) implementation that
  returns a lazy keybase instance. Remove client.MockKeyBase,
  superseded by crypto/keys.NewInMemory()
* [#3621](https://github.com/cosmos/cosmos-sdk/issues/3621) staking.GenesisState.Bonds -> Delegations

IMPROVEMENTS

* SDK

    * [#3311](https://github.com/cosmos/cosmos-sdk/pull/3311) Reconcile the `DecCoin/s` API with the `Coin/s` API.
    * [#3614](https://github.com/cosmos/cosmos-sdk/pull/3614) Add coin denom length checks to the coins constructors.
    * [#3621](https://github.com/cosmos/cosmos-sdk/issues/3621) remove many inter-module dependancies
    * [#3601](https://github.com/cosmos/cosmos-sdk/pull/3601) JSON-stringify the ABCI log response which includes the log and message
    index.
    * [#3604](https://github.com/cosmos/cosmos-sdk/pull/3604) Improve SDK funds related error messages and allow for unicode in
    JSON ABCI log.
    * [#3620](https://github.com/cosmos/cosmos-sdk/pull/3620) Version command shows build tags
    * [#3638](https://github.com/cosmos/cosmos-sdk/pull/3638) Add Bcrypt benchmarks & justification of security parameter choice
    * [#3648](https://github.com/cosmos/cosmos-sdk/pull/3648) Add JSON struct tags to vesting accounts.

* Tendermint
    * [#3618](https://github.com/cosmos/cosmos-sdk/pull/3618) Upgrade to Tendermint 0.30.03

BUG FIXES

* SDK
    * [#3646](https://github.com/cosmos/cosmos-sdk/issues/3646) `x/mint` now uses total token supply instead of total bonded tokens to calculate inflation

## 0.31.2

BREAKING CHANGES

* SDK
* [#3592](https://github.com/cosmos/cosmos-sdk/issues/3592) Drop deprecated keybase implementation's
  New constructor in favor of a new
  crypto/keys.New(string, string) implementation that
  returns a lazy keybase instance. Remove client.MockKeyBase,
  superseded by crypto/keys.NewInMemory()

IMPROVEMENTS

* SDK

    * [#3604](https://github.com/cosmos/cosmos-sdk/pulls/3604) Improve SDK funds related error messages and allow for unicode in
    JSON ABCI log.

* Tendermint
    * [#3563](https://github.com/cosmos/cosmos-sdk/3563) Update to Tendermint version `0.30.0-rc0`

BUG FIXES

* Gaia

    * [#3585] Fix setting the tx hash in `NewResponseFormatBroadcastTxCommit`.
    * [#3585] Return an empty `TxResponse` when Tendermint returns an empty
    `ResultBroadcastTx`.

* SDK
    * [#3582](https://github.com/cosmos/cosmos-sdk/pull/3582) Running `make test_unit` was failing due to a missing tag
    * [#3617](https://github.com/cosmos/cosmos-sdk/pull/3582) Fix fee comparison when the required fees does not contain any denom
    present in the tx fees.

## 0.31.0

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

    * [#3284](https://github.com/cosmos/cosmos-sdk/issues/3284) Rename the `name`
    field to `from` in the `base_req` body.
    * [#3485](https://github.com/cosmos/cosmos-sdk/pull/3485) Error responses are now JSON objects.
    * [#3477][distribution] endpoint changed "all_delegation_rewards" -> "delegator_total_rewards"

* Gaia CLI (`gaiacli`)

    * [#3399](https://github.com/cosmos/cosmos-sdk/pull/3399) Add `gaiad validate-genesis` command to facilitate checking of genesis files
    * [#1894](https://github.com/cosmos/cosmos-sdk/issues/1894) `version` prints out short info by default. Add `--long` flag. Proper handling of `--format` flag introduced.
    * [#3465](https://github.com/cosmos/cosmos-sdk/issues/3465) `gaiacli rest-server` switched back to insecure mode by default:
        * `--insecure` flag is removed.
        * `--tls` is now used to enable secure layer.
    * [#3451](https://github.com/cosmos/cosmos-sdk/pull/3451) `gaiacli` now returns transactions in plain text including tags.
    * [#3497](https://github.com/cosmos/cosmos-sdk/issues/3497) `gaiad init` now takes moniker as required arguments, not as parameter.
    * [#3501](https://github.com/cosmos/cosmos-sdk/issues/3501) Change validator
    address Bech32 encoding to consensus address in `tendermint-validator-set`.

* Gaia

    * [#3457](https://github.com/cosmos/cosmos-sdk/issues/3457) Changed governance tally validatorGovInfo to use sdk.Int power instead of sdk.Dec
    * [#3495](https://github.com/cosmos/cosmos-sdk/issues/3495) Added Validator Minimum Self Delegation
    * Reintroduce OR semantics for tx fees

* SDK
    * [#2513](https://github.com/cosmos/cosmos-sdk/issues/2513) Tendermint updates are adjusted by 10^-6 relative to staking tokens,
    * [#3487](https://github.com/cosmos/cosmos-sdk/pull/3487) Move HTTP/REST utilities out of client/utils into a new dedicated client/rest package.
    * [#3490](https://github.com/cosmos/cosmos-sdk/issues/3490) ReadRESTReq() returns bool to avoid callers to write error responses twice.
    * [#3502](https://github.com/cosmos/cosmos-sdk/pull/3502) Fixes issue when comparing genesis states
    * [#3514](https://github.com/cosmos/cosmos-sdk/pull/3514) Various clean ups:
        * Replace all GetKeyBase\* functions family in favor of NewKeyBaseFromDir and NewKeyBaseFromHomeFlag.
        * Remove Get prefix from all TxBuilder's getters.
    * [#3522](https://github.com/cosmos/cosmos-sdk/pull/3522) Get rid of double negatives: Coins.IsNotNegative() -> Coins.IsAnyNegative().
    * [#3561](https://github.com/cosmos/cosmos-sdk/issues/3561) Don't unnecessarily store denominations in staking

FEATURES

* Gaia REST API

* [#2358](https://github.com/cosmos/cosmos-sdk/issues/2358) Add distribution module REST interface

* Gaia CLI (`gaiacli`)

    * [#3429](https://github.com/cosmos/cosmos-sdk/issues/3429) Support querying
    for all delegator distribution rewards.
    * [#3449](https://github.com/cosmos/cosmos-sdk/issues/3449) Proof verification now works with absence proofs
    * [#3484](https://github.com/cosmos/cosmos-sdk/issues/3484) Add support
    vesting accounts to the add-genesis-account command.

* Gaia

    * [#3397](https://github.com/cosmos/cosmos-sdk/pull/3397) Implement genesis file sanitization to avoid failures at chain init.
    * [#3428](https://github.com/cosmos/cosmos-sdk/issues/3428) Run the simulation from a particular genesis state loaded from a file

* SDK
    * [#3270](https://github.com/cosmos/cosmos-sdk/issues/3270) [x/staking] limit number of ongoing unbonding delegations /redelegations per pair/trio
    * [#3477][distribution] new query endpoint "delegator_validators"
    * [#3514](https://github.com/cosmos/cosmos-sdk/pull/3514) Provided a lazy loading implementation of Keybase that locks the underlying
    storage only for the time needed to perform the required operation. Also added Keybase reference to TxBuilder struct.
    * [types] [#2580](https://github.com/cosmos/cosmos-sdk/issues/2580) Addresses now Bech32 empty addresses to an empty string

IMPROVEMENTS

* Gaia REST API

    * [#3284](https://github.com/cosmos/cosmos-sdk/issues/3284) Update Gaia Lite
    REST service to support the following:
    _ Automatic account number and sequence population when fields are omitted
    _ Generate only functionality no longer requires access to a local Keybase \* `from` field in the `base_req` body can be a Keybase name or account address
    * [#3423](https://github.com/cosmos/cosmos-sdk/issues/3423) Allow simulation
    (auto gas) to work with generate only.
    * [#3514](https://github.com/cosmos/cosmos-sdk/pull/3514) REST server calls to keybase does not lock the underlying storage anymore.
    * [#3523](https://github.com/cosmos/cosmos-sdk/pull/3523) Added `/tx/encode` endpoint to serialize a JSON tx to base64-encoded Amino.

* Gaia CLI (`gaiacli`)

    * [#3476](https://github.com/cosmos/cosmos-sdk/issues/3476) New `withdraw-all-rewards` command to withdraw all delegations rewards for delegators.
    * [#3497](https://github.com/cosmos/cosmos-sdk/issues/3497) `gaiad gentx` supports `--ip` and `--node-id` flags to override defaults.
    * [#3518](https://github.com/cosmos/cosmos-sdk/issues/3518) Fix flow in
    `keys add` to show the mnemonic by default.
    * [#3517](https://github.com/cosmos/cosmos-sdk/pull/3517) Increased test coverage
    * [#3523](https://github.com/cosmos/cosmos-sdk/pull/3523) Added `tx encode` command to serialize a JSON tx to base64-encoded Amino.

* Gaia

    * [#3418](https://github.com/cosmos/cosmos-sdk/issues/3418) Add vesting account
    genesis validation checks to `GaiaValidateGenesisState`.
    * [#3420](https://github.com/cosmos/cosmos-sdk/issues/3420) Added maximum length to governance proposal descriptions and titles
    * [#3256](https://github.com/cosmos/cosmos-sdk/issues/3256) Add gas consumption
    for tx size in the ante handler.
    * [#3454](https://github.com/cosmos/cosmos-sdk/pull/3454) Add `--jail-whitelist` to `gaiad export` to enable testing of complex exports
    * [#3424](https://github.com/cosmos/cosmos-sdk/issues/3424) Allow generation of gentxs with empty memo field.
    * [#3507](https://github.com/cosmos/cosmos-sdk/issues/3507) General cleanup, removal of unnecessary struct fields, undelegation bugfix, and comment clarification in x/staking and x/slashing

* SDK
    * [#2605] x/params add subkey accessing
    * [#2986](https://github.com/cosmos/cosmos-sdk/pull/2986) Store Refactor
    * [#3435](https://github.com/cosmos/cosmos-sdk/issues/3435) Test that store implementations do not allow nil values
    * [#2509](https://github.com/cosmos/cosmos-sdk/issues/2509) Sanitize all usage of Dec.RoundInt64()
    * [#556](https://github.com/cosmos/cosmos-sdk/issues/556) Increase `BaseApp`
    test coverage.
    * [#3357](https://github.com/cosmos/cosmos-sdk/issues/3357) develop state-transitions.md for staking spec, missing states added to `state.md`
    * [#3552](https://github.com/cosmos/cosmos-sdk/pull/3552) Validate bit length when
    deserializing `Int` types.

BUG FIXES

* Gaia CLI (`gaiacli`)

    * [#3417](https://github.com/cosmos/cosmos-sdk/pull/3417) Fix `q slashing signing-info` panic by ensuring safety of user input and properly returning not found error
    * [#3345](https://github.com/cosmos/cosmos-sdk/issues/3345) Upgrade ledger-cosmos-go dependency to v0.9.3 to pull
    https://github.com/ZondaX/ledger-cosmos-go/commit/ed9aa39ce8df31bad1448c72d3d226bf2cb1a8d1 in order to fix a derivation path issue that causes `gaiacli keys add --recover`
    to malfunction.
    * [#3419](https://github.com/cosmos/cosmos-sdk/pull/3419) Fix `q distr slashes` panic
    * [#3453](https://github.com/cosmos/cosmos-sdk/pull/3453) The `rest-server` command didn't respect persistent flags such as `--chain-id` and `--trust-node` if they were
    passed on the command line.
    * [#3441](https://github.com/cosmos/cosmos-sdk/pull/3431) Improved resource management and connection handling (ledger devices). Fixes issue with DER vs BER signatures.

* Gaia
    * [#3486](https://github.com/cosmos/cosmos-sdk/pull/3486) Use AmountOf in
    vesting accounts instead of zipping/aligning denominations.

## 0.30.0

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

    * [gaia-lite] [#2182] Renamed and merged all redelegations endpoints into `/staking/redelegations`
    * [#3176](https://github.com/cosmos/cosmos-sdk/issues/3176) `tx/sign` endpoint now expects `BaseReq` fields as nested object.
    * [#2222] all endpoints renamed from `/stake` -> `/staking`
    * [#1268] `LooseTokens` -> `NotBondedTokens`
    * [#3289] misc renames:
        * `Validator.UnbondingMinTime` -> `Validator.UnbondingCompletionTime`
        * `Delegation` -> `Value` in `MsgCreateValidator` and `MsgDelegate`
        * `MsgBeginUnbonding` -> `MsgUndelegate`

* Gaia CLI (`gaiacli`)

    * [#810](https://github.com/cosmos/cosmos-sdk/issues/810) Don't fallback to any default values for chain ID.
        * Users need to supply chain ID either via config file or the `--chain-id` flag.
        * Change `chain_id` and `trust_node` in `gaiacli` configuration to `chain-id` and `trust-node` respectively.
    * [#3069](https://github.com/cosmos/cosmos-sdk/pull/3069) `--fee` flag renamed to `--fees` to support multiple coins
    * [#3156](https://github.com/cosmos/cosmos-sdk/pull/3156) Remove unimplemented `gaiacli init` command
    * [#2222] `gaiacli tx stake` -> `gaiacli tx staking`, `gaiacli query stake` -> `gaiacli query staking`
    * [#1894](https://github.com/cosmos/cosmos-sdk/issues/1894) `version` command now shows latest commit, vendor dir hash, and build machine info.
    * [#3320](https://github.com/cosmos/cosmos-sdk/pull/3320) Ensure all `gaiacli query` commands respect the `--output` and `--indent` flags

* Gaia

    * https://github.com/cosmos/cosmos-sdk/issues/2838 - Move store keys to constants
    * [#3162](https://github.com/cosmos/cosmos-sdk/issues/3162) The `--gas` flag now takes `auto` instead of `simulate`
    in order to trigger a simulation of the tx before the actual execution.
    * [#3285](https://github.com/cosmos/cosmos-sdk/pull/3285) New `gaiad tendermint version` to print libs versions
    * [#1894](https://github.com/cosmos/cosmos-sdk/pull/1894) `version` command now shows latest commit, vendor dir hash, and build machine info.
    * [#3249](https://github.com/cosmos/cosmos-sdk/issues/3249) `tendermint`'s `show-validator` and `show-address` `--json` flags removed in favor of `--output-format=json`.

* SDK

    * [distribution] [#3359](https://github.com/cosmos/cosmos-sdk/issues/3359) Always round down when calculating rewards-to-be-withdrawn in F1 fee distribution
    * [#3336](https://github.com/cosmos/cosmos-sdk/issues/3336) Ensure all SDK
    messages have their signature bytes contain canonical fields `value` and `type`.
    * [#3333](https://github.com/cosmos/cosmos-sdk/issues/3333) - F1 storage efficiency improvements - automatic withdrawals when unbonded, historical reward reference counting
    * [staking] [#2513](https://github.com/cosmos/cosmos-sdk/issues/2513) Validator power type from Dec -> Int
    * [staking] [#3233](https://github.com/cosmos/cosmos-sdk/issues/3233) key and value now contain duplicate fields to simplify code
    * [#3064](https://github.com/cosmos/cosmos-sdk/issues/3064) Sanitize `sdk.Coin` denom. Coins denoms are now case insensitive, i.e. 100fooToken equals to 100FOOTOKEN.
    * [#3195](https://github.com/cosmos/cosmos-sdk/issues/3195) Allows custom configuration for syncable strategy
    * [#3242](https://github.com/cosmos/cosmos-sdk/issues/3242) Fix infinite gas
    meter utilization during aborted ante handler executions.
    * [x/distribution] [#3292](https://github.com/cosmos/cosmos-sdk/issues/3292) Enable or disable withdraw addresses with a parameter in the param store
    * [staking] [#2222](https://github.com/cosmos/cosmos-sdk/issues/2222) `/stake` -> `/staking` module rename
    * [staking] [#1268](https://github.com/cosmos/cosmos-sdk/issues/1268) `LooseTokens` -> `NotBondedTokens`
    * [staking] [#1402](https://github.com/cosmos/cosmos-sdk/issues/1402) Redelegation and unbonding-delegation structs changed to include multiple an array of entries
    * [staking] [#3289](https://github.com/cosmos/cosmos-sdk/issues/3289) misc renames:
        * `Validator.UnbondingMinTime` -> `Validator.UnbondingCompletionTime`
        * `Delegation` -> `Value` in `MsgCreateValidator` and `MsgDelegate`
        * `MsgBeginUnbonding` -> `MsgUndelegate`
    * [#3315] Increase decimal precision to 18
    * [#3323](https://github.com/cosmos/cosmos-sdk/issues/3323) Update to Tendermint 0.29.0
    * [#3328](https://github.com/cosmos/cosmos-sdk/issues/3328) [x/gov] Remove redundant action tag

* Tendermint
    * [#3298](https://github.com/cosmos/cosmos-sdk/issues/3298) Upgrade to Tendermint 0.28.0

FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)

    * [#3067](https://github.com/cosmos/cosmos-sdk/issues/3067) Add support for fees on transactions
    * [#3069](https://github.com/cosmos/cosmos-sdk/pull/3069) Add a custom memo on transactions
    * [#3027](https://github.com/cosmos/cosmos-sdk/issues/3027) Implement
    `/gov/proposals/{proposalID}/proposer` to query for a proposal's proposer.

* Gaia CLI (`gaiacli`)

    * [#2399](https://github.com/cosmos/cosmos-sdk/issues/2399) Implement `params` command to query slashing parameters.
    * [#2730](https://github.com/cosmos/cosmos-sdk/issues/2730) Add tx search pagination parameter
    * [#3027](https://github.com/cosmos/cosmos-sdk/issues/3027) Implement
    `query gov proposer [proposal-id]` to query for a proposal's proposer.
    * [#3198](https://github.com/cosmos/cosmos-sdk/issues/3198) New `keys add --multisig` flag to store multisig keys locally.
    * [#3198](https://github.com/cosmos/cosmos-sdk/issues/3198) New `multisign` command to generate multisig signatures.
    * [#3198](https://github.com/cosmos/cosmos-sdk/issues/3198) New `sign --multisig` flag to enable multisig mode.
    * [#2715](https://github.com/cosmos/cosmos-sdk/issues/2715) Reintroduce gaia server's insecure mode.
    * [#3334](https://github.com/cosmos/cosmos-sdk/pull/3334) New `gaiad completion` and `gaiacli completion` to generate Bash/Zsh completion scripts.
    * [#2607](https://github.com/cosmos/cosmos-sdk/issues/2607) Make `gaiacli config` handle the boolean `indent` flag to beautify commands JSON output.

* Gaia

    * [#2182] [x/staking] Added querier for querying a single redelegation
    * [#3305](https://github.com/cosmos/cosmos-sdk/issues/3305) Add support for
    vesting accounts at genesis.
    * [#3198](https://github.com/cosmos/cosmos-sdk/issues/3198) [x/auth] Add multisig transactions support
    * [#3198](https://github.com/cosmos/cosmos-sdk/issues/3198) `add-genesis-account` can take both account addresses and key names

* SDK
    * [#3099](https://github.com/cosmos/cosmos-sdk/issues/3099) Implement F1 fee distribution
    * [#2926](https://github.com/cosmos/cosmos-sdk/issues/2926) Add TxEncoder to client TxBuilder.
    * [#2694](https://github.com/cosmos/cosmos-sdk/issues/2694) Vesting account implementation.
    * [#2996](https://github.com/cosmos/cosmos-sdk/issues/2996) Update the `AccountKeeper` to contain params used in the context of
    the ante handler.
    * [#3179](https://github.com/cosmos/cosmos-sdk/pull/3179) New CodeNoSignatures error code.
    * [#3319](https://github.com/cosmos/cosmos-sdk/issues/3319) [x/distribution] Queriers for all distribution state worth querying; distribution query commands
    * [#3356](https://github.com/cosmos/cosmos-sdk/issues/3356) [x/auth] bech32-ify accounts address in error message.

IMPROVEMENTS

* Gaia REST API

    * [#3176](https://github.com/cosmos/cosmos-sdk/issues/3176) Validate tx/sign endpoint POST body.
    * [#2948](https://github.com/cosmos/cosmos-sdk/issues/2948) Swagger UI now makes requests to light client node

* Gaia CLI (`gaiacli`)

    * [#3224](https://github.com/cosmos/cosmos-sdk/pull/3224) Support adding offline public keys to the keystore

* Gaia

    * [#2186](https://github.com/cosmos/cosmos-sdk/issues/2186) Add Address Interface
    * [#3158](https://github.com/cosmos/cosmos-sdk/pull/3158) Validate slashing genesis
    * [#3172](https://github.com/cosmos/cosmos-sdk/pull/3172) Support minimum fees in a local testnet.
    * [#3250](https://github.com/cosmos/cosmos-sdk/pull/3250) Refactor integration tests and increase coverage
    * [#3248](https://github.com/cosmos/cosmos-sdk/issues/3248) Refactor tx fee
    model:
    _ Validators specify minimum gas prices instead of minimum fees
    _ Clients may provide either fees or gas prices directly
    _ The gas prices of a tx must meet a validator's minimum
    _ `gaiad start` and `gaia.toml` take --minimum-gas-prices flag and minimum-gas-price config key respectively.
    * [#2859](https://github.com/cosmos/cosmos-sdk/issues/2859) Rename `TallyResult` in gov proposals to `FinalTallyResult`
    * [#3286](https://github.com/cosmos/cosmos-sdk/pull/3286) Fix `gaiad gentx` printout of account's addresses, i.e. user bech32 instead of hex.
    * [#3249](https://github.com/cosmos/cosmos-sdk/issues/3249) `--json` flag removed, users should use `--output=json` instead.

* SDK

    * [#3137](https://github.com/cosmos/cosmos-sdk/pull/3137) Add tag documentation
    for each module along with cleaning up a few existing tags in the governance,
    slashing, and staking modules.
    * [#3093](https://github.com/cosmos/cosmos-sdk/issues/3093) Ante handler does no longer read all accounts in one go when processing signatures as signature
    verification may fail before last signature is checked.
    * [staking] [#1402](https://github.com/cosmos/cosmos-sdk/issues/1402) Add for multiple simultaneous redelegations or unbonding-delegations within an unbonding period
    * [staking] [#1268](https://github.com/cosmos/cosmos-sdk/issues/1268) staking spec rewrite

* CI
    * [#2498](https://github.com/cosmos/cosmos-sdk/issues/2498) Added macos CI job to CircleCI
    * [#142](https://github.com/tendermint/devops/issues/142) Increased the number of blocks to be tested during multi-sim
    * [#147](https://github.com/tendermint/devops/issues/142) Added docker image build to CI

BUG FIXES

* Gaia CLI (`gaiacli`)

    * [#3141](https://github.com/cosmos/cosmos-sdk/issues/3141) Fix the bug in GetAccount when `len(res) == 0` and `err == nil`
    * [#810](https://github.com/cosmos/cosmos-sdk/pull/3316) Fix regression in gaiacli config file handling

* Gaia
    * [#3148](https://github.com/cosmos/cosmos-sdk/issues/3148) Fix `gaiad export` by adding a boolean to `NewGaiaApp` determining whether or not to load the latest version
    * [#3181](https://github.com/cosmos/cosmos-sdk/issues/3181) Correctly reset total accum update height and jailed-validator bond height / unbonding height on export-for-zero-height
    * [#3172](https://github.com/cosmos/cosmos-sdk/pull/3172) Fix parsing `gaiad.toml`
    when it already exists.
    * [#3223](https://github.com/cosmos/cosmos-sdk/issues/3223) Fix unset governance proposal queues when importing state from old chain
    * [#3187](https://github.com/cosmos/cosmos-sdk/issues/3187) Fix `gaiad export`
    by resetting each validator's slashing period.

## 0.29.1

BUG FIXES

* SDK
    * [#3207](https://github.com/cosmos/cosmos-sdk/issues/3207) - Fix token printing bug

## 0.29.0

BREAKING CHANGES

* Gaia

    * [#3148](https://github.com/cosmos/cosmos-sdk/issues/3148) Fix `gaiad export` by adding a boolean to `NewGaiaApp` determining whether or not to load the latest version

* SDK
    * [#3163](https://github.com/cosmos/cosmos-sdk/issues/3163) Withdraw commission on self bond removal

## 0.28.1

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
    * [lcd] [#3045](https://github.com/cosmos/cosmos-sdk/pull/3045) Fix quoted json return on GET /keys (keys list)
    * [gaia-lite] [#2191](https://github.com/cosmos/cosmos-sdk/issues/2191) Split `POST /stake/delegators/{delegatorAddr}/delegations` into `POST /stake/delegators/{delegatorAddr}/delegations`, `POST /stake/delegators/{delegatorAddr}/unbonding_delegations` and `POST /stake/delegators/{delegatorAddr}/redelegations`
    * [gaia-lite] [#3056](https://github.com/cosmos/cosmos-sdk/pull/3056) `generate_only` and `simulate` have moved from query arguments to POST requests body.
* Tendermint
    * [tendermint] Now using Tendermint 0.27.3

FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
    * [slashing] [#2399](https://github.com/cosmos/cosmos-sdk/issues/2399) Implement `/slashing/parameters` endpoint to query slashing parameters.
* Gaia CLI (`gaiacli`)
    * [gaiacli] [#2399](https://github.com/cosmos/cosmos-sdk/issues/2399) Implement `params` command to query slashing parameters.
* SDK
    * [client] [#2926](https://github.com/cosmos/cosmos-sdk/issues/2926) Add TxEncoder to client TxBuilder.
* Other
    * Introduced the logjack tool for saving logs w/ rotation

IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)
    * [#2879](https://github.com/cosmos/cosmos-sdk/issues/2879), [#2880](https://github.com/cosmos/cosmos-sdk/issues/2880) Update deposit and vote endpoints to perform a direct txs query
    when a given proposal is inactive and thus having votes and deposits removed
    from state.
* Gaia CLI (`gaiacli`)
    * [#2879](https://github.com/cosmos/cosmos-sdk/issues/2879), [#2880](https://github.com/cosmos/cosmos-sdk/issues/2880) Update deposit and vote CLI commands to perform a direct txs query
    when a given proposal is inactive and thus having votes and deposits removed
    from state.
* Gaia
    * [#3021](https://github.com/cosmos/cosmos-sdk/pull/3021) Add `--gentx-dir` to `gaiad collect-gentxs` to specify a directory from which collect and load gentxs. Add `--output-document` to `gaiad init` to allow one to redirect output to file.

## 0.28.0

BREAKING CHANGES

* Gaia CLI (`gaiacli`)

    * [cli] [#2595](https://github.com/cosmos/cosmos-sdk/issues/2595) Remove `keys new` in favor of `keys add` incorporating existing functionality with addition of key recovery functionality.
    * [cli] [#2987](https://github.com/cosmos/cosmos-sdk/pull/2987) Add shorthand `-a` to `gaiacli keys show` and update docs
    * [cli] [#2971](https://github.com/cosmos/cosmos-sdk/pull/2971) Additional verification when running `gaiad gentx`
    * [cli] [#2734](https://github.com/cosmos/cosmos-sdk/issues/2734) Rewrite `gaiacli config`. It is now a non-interactive config utility.

* Gaia

    * [#128](https://github.com/tendermint/devops/issues/128) Updated CircleCI job to trigger website build on every push to master/develop.
    * [#2994](https://github.com/cosmos/cosmos-sdk/pull/2994) Change wrong-password error message.
    * [#3009](https://github.com/cosmos/cosmos-sdk/issues/3009) Added missing Gaia genesis verification
    * [#128](https://github.com/tendermint/devops/issues/128) Updated CircleCI job to trigger website build on every push to master/develop.
    * [#2994](https://github.com/cosmos/cosmos-sdk/pull/2994) Change wrong-password error message.
    * [#3009](https://github.com/cosmos/cosmos-sdk/issues/3009) Added missing Gaia genesis verification
    * [gas] [#3052](https://github.com/cosmos/cosmos-sdk/issues/3052) Updated gas costs to more reasonable numbers

* SDK
    * [auth] [#2952](https://github.com/cosmos/cosmos-sdk/issues/2952) Signatures are no longer serialized on chain with the account number and sequence number
    * [auth] [#2952](https://github.com/cosmos/cosmos-sdk/issues/2952) Signatures are no longer serialized on chain with the account number and sequence number
    * [stake] [#3055](https://github.com/cosmos/cosmos-sdk/issues/3055) Use address instead of bond height / intratxcounter for deduplication

FEATURES

* Gaia CLI (`gaiacli`)
    * [#2961](https://github.com/cosmos/cosmos-sdk/issues/2961) Add --force flag to gaiacli keys delete command to skip passphrase check and force key deletion unconditionally.

IMPROVEMENTS

* Gaia CLI (`gaiacli`)

    * [#2991](https://github.com/cosmos/cosmos-sdk/issues/2991) Fully validate transaction signatures during `gaiacli tx sign --validate-signatures`

* SDK
    * [#1277](https://github.com/cosmos/cosmos-sdk/issues/1277) Complete bank module specification
    * [#2963](https://github.com/cosmos/cosmos-sdk/issues/2963) Complete auth module specification
    * [#2914](https://github.com/cosmos/cosmos-sdk/issues/2914) No longer withdraw validator rewards on bond/unbond, but rather move
    the rewards to the respective validator's pools.

BUG FIXES

* Gaia CLI (`gaiacli`)

    * [#2921](https://github.com/cosmos/cosmos-sdk/issues/2921) Fix `keys delete` inability to delete offline and ledger keys.

* Gaia

    * [#3003](https://github.com/cosmos/cosmos-sdk/issues/3003) CollectStdTxs() must validate DelegatorAddr against genesis accounts.

* SDK
    * [#2967](https://github.com/cosmos/cosmos-sdk/issues/2967) Change ordering of `mint.BeginBlocker` and `distr.BeginBlocker`, recalculate inflation each block
    * [#3068](https://github.com/cosmos/cosmos-sdk/issues/3068) check for uint64 gas overflow during `Std#ValidateBasic`.
    * [#3071](https://github.com/cosmos/cosmos-sdk/issues/3071) Catch overflow on block gas meter

## 0.27.0

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

    * [gaia-lite] [#2819](https://github.com/cosmos/cosmos-sdk/pull/2819) Txs query param format is now: `/txs?tag=value` (removed '' wrapping the query parameter `value`)

* Gaia CLI (`gaiacli`)

    * [cli] [#2728](https://github.com/cosmos/cosmos-sdk/pull/2728) Seperate `tx` and `query` subcommands by module
    * [cli] [#2727](https://github.com/cosmos/cosmos-sdk/pull/2727) Fix unbonding command flow
    * [cli] [#2786](https://github.com/cosmos/cosmos-sdk/pull/2786) Fix redelegation command flow
    * [cli] [#2829](https://github.com/cosmos/cosmos-sdk/pull/2829) add-genesis-account command now validates state when adding accounts
    * [cli] [#2804](https://github.com/cosmos/cosmos-sdk/issues/2804) Check whether key exists before passing it on to `tx create-validator`.
    * [cli] [#2874](https://github.com/cosmos/cosmos-sdk/pull/2874) `gaiacli tx sign` takes an optional `--output-document` flag to support output redirection.
    * [cli] [#2875](https://github.com/cosmos/cosmos-sdk/pull/2875) Refactor `gaiad gentx` and avoid redirection to `gaiacli tx sign` for tx signing.

* Gaia

    * [mint] [#2825] minting now occurs every block, inflation parameter updates still hourly

* SDK

    * [#2752](https://github.com/cosmos/cosmos-sdk/pull/2752) Don't hardcode bondable denom.
    * [#2701](https://github.com/cosmos/cosmos-sdk/issues/2701) Account numbers and sequence numbers in `auth` are now `uint64` instead of `int64`
    * [#2019](https://github.com/cosmos/cosmos-sdk/issues/2019) Cap total number of signatures. Current per-transaction limit is 7, and if that is exceeded transaction is rejected.
    * [#2801](https://github.com/cosmos/cosmos-sdk/pull/2801) Remove AppInit structure.
    * [#2798](https://github.com/cosmos/cosmos-sdk/issues/2798) Governance API has miss-spelled English word in JSON response ('depositer' -> 'depositor')
    * [#2943](https://github.com/cosmos/cosmos-sdk/pull/2943) Transaction action tags equal the message type. Staking EndBlocker tags are included.

* Tendermint
    * Update to Tendermint 0.27.0

FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)

    * [gov] [#2479](https://github.com/cosmos/cosmos-sdk/issues/2479) Added governance parameter
    query REST endpoints.

* Gaia CLI (`gaiacli`)

    * [gov][cli] [#2479](https://github.com/cosmos/cosmos-sdk/issues/2479) Added governance
    parameter query commands.
    * [stake][cli] [#2027] Add CLI query command for getting all delegations to a specific validator.
    * [#2840](https://github.com/cosmos/cosmos-sdk/pull/2840) Standardize CLI exports from modules

* Gaia

    * [app] [#2791](https://github.com/cosmos/cosmos-sdk/issues/2791) Support export at a specific height, with `gaiad export --height=HEIGHT`.
    * [x/gov] [#2479](https://github.com/cosmos/cosmos-sdk/issues/2479) Implemented querier
    for getting governance parameters.
    * [app] [#2663](https://github.com/cosmos/cosmos-sdk/issues/2663) - Runtime-assertable invariants
    * [app] [#2791](https://github.com/cosmos/cosmos-sdk/issues/2791) Support export at a specific height, with `gaiad export --height=HEIGHT`.
    * [app] [#2812](https://github.com/cosmos/cosmos-sdk/issues/2812) Support export alterations to prepare for restarting at zero-height

* SDK
    * [simulator] [#2682](https://github.com/cosmos/cosmos-sdk/issues/2682) MsgEditValidator now looks at the validator's max rate, thus it now succeeds a significant portion of the time
    * [core] [#2775](https://github.com/cosmos/cosmos-sdk/issues/2775) Add deliverTx maximum block gas limit

IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)

    * [gaia-lite] [#2819](https://github.com/cosmos/cosmos-sdk/pull/2819) Tx search now supports multiple tags as query parameters
    * [#2836](https://github.com/cosmos/cosmos-sdk/pull/2836) Expose LCD router to allow users to register routes there.

* Gaia CLI (`gaiacli`)

    * [#2749](https://github.com/cosmos/cosmos-sdk/pull/2749) Add --chain-id flag to gaiad testnet
    * [#2819](https://github.com/cosmos/cosmos-sdk/pull/2819) Tx search now supports multiple tags as query parameters

* Gaia

    * [#2772](https://github.com/cosmos/cosmos-sdk/issues/2772) Update BaseApp to not persist state when the ante handler fails on DeliverTx.
    * [#2773](https://github.com/cosmos/cosmos-sdk/issues/2773) Require moniker to be provided on `gaiad init`.
    * [#2672](https://github.com/cosmos/cosmos-sdk/issues/2672) [Makefile] Updated for better Windows compatibility and ledger support logic, get_tools was rewritten as a cross-compatible Makefile.
    * [#2766](https://github.com/cosmos/cosmos-sdk/issues/2766) [Makefile] Added goimports tool to get_tools. Get_tools now only builds new versions if binaries are missing.
    * [#110](https://github.com/tendermint/devops/issues/110) Updated CircleCI job to trigger website build when cosmos docs are updated.

* SDK
  & [x/mock/simulation] [#2720] major cleanup, introduction of helper objects, reorganization
* [#2821](https://github.com/cosmos/cosmos-sdk/issues/2821) Codespaces are now strings
* [types] [#2776](https://github.com/cosmos/cosmos-sdk/issues/2776) Improve safety of `Coin` and `Coins` types. Various functions
  and methods will panic when a negative amount is discovered.
* [#2815](https://github.com/cosmos/cosmos-sdk/issues/2815) Gas unit fields changed from `int64` to `uint64`.
* [#2821](https://github.com/cosmos/cosmos-sdk/issues/2821) Codespaces are now strings
* [#2779](https://github.com/cosmos/cosmos-sdk/issues/2779) Introduce `ValidateBasic` to the `Tx` interface and call it in the ante
  handler.
* [#2825](https://github.com/cosmos/cosmos-sdk/issues/2825) More staking and distribution invariants
* [#2912](https://github.com/cosmos/cosmos-sdk/issues/2912) Print commit ID in hex when commit is synced.

* Tendermint
* [#2796](https://github.com/cosmos/cosmos-sdk/issues/2796) Update to go-amino 0.14.1

BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

    * [gaia-lite] [#2868](https://github.com/cosmos/cosmos-sdk/issues/2868) Added handler for governance tally endpoint
    * [#2907](https://github.com/cosmos/cosmos-sdk/issues/2907) Refactor and fix the way Gaia Lite is started.

* Gaia

    * [#2723] Use `cosmosvalcons` Bech32 prefix in `tendermint show-address`
    * [#2742](https://github.com/cosmos/cosmos-sdk/issues/2742) Fix time format of TimeoutCommit override
    * [#2898](https://github.com/cosmos/cosmos-sdk/issues/2898) Remove redundant '$' in docker-compose.yml

* SDK

    * [#2733](https://github.com/cosmos/cosmos-sdk/issues/2733) [x/gov, x/mock/simulation] Fix governance simulation, update x/gov import/export
    * [#2854](https://github.com/cosmos/cosmos-sdk/issues/2854) [x/bank] Remove unused bank.MsgIssue, prevent possible panic
    * [#2884](https://github.com/cosmos/cosmos-sdk/issues/2884) [docs/examples] Fix `basecli version` panic

* Tendermint
    * [#2797](https://github.com/tendermint/tendermint/pull/2797) AddressBook requires addresses to have IDs; Do not crap out immediately after sending pex addrs in seed mode

## 0.26.0

BREAKING CHANGES

* Gaia

    * [gaiad init] [#2602](https://github.com/cosmos/cosmos-sdk/issues/2602) New genesis workflow

* SDK

    * [simulation] [#2665](https://github.com/cosmos/cosmos-sdk/issues/2665) only argument to sdk.Invariant is now app

* Tendermint
    * Upgrade to version 0.26.0

FEATURES

* Gaia CLI (`gaiacli`)

    * [cli] [#2569](https://github.com/cosmos/cosmos-sdk/pull/2569) Add commands to query validator unbondings and redelegations
    * [cli] [#2569](https://github.com/cosmos/cosmos-sdk/pull/2569) Add commands to query validator unbondings and redelegations
    * [cli] [#2524](https://github.com/cosmos/cosmos-sdk/issues/2524) Add support offline mode to `gaiacli tx sign`. Lookups are not performed if the flag `--offline` is on.
    * [cli] [#2558](https://github.com/cosmos/cosmos-sdk/issues/2558) Rename --print-sigs to --validate-signatures. It now performs a complete set of sanity checks and reports to the user. Also added --print-signature-only to print the signature only, not the whole transaction.
    * [cli] [#2704](https://github.com/cosmos/cosmos-sdk/pull/2704) New add-genesis-account convenience command to populate genesis.json with genesis accounts.

* SDK
    * [#1336](https://github.com/cosmos/cosmos-sdk/issues/1336) Mechanism for SDK Users to configure their own Bech32 prefixes instead of using the default cosmos prefixes.

IMPROVEMENTS

* Gaia
* [#2637](https://github.com/cosmos/cosmos-sdk/issues/2637) [x/gov] Switched inactive and active proposal queues to an iterator based queue

* SDK
* [#2573](https://github.com/cosmos/cosmos-sdk/issues/2573) [x/distribution] add accum invariance
* [#2556](https://github.com/cosmos/cosmos-sdk/issues/2556) [x/mock/simulation] Fix debugging output
* [#2396](https://github.com/cosmos/cosmos-sdk/issues/2396) [x/mock/simulation] Change parameters to get more slashes
* [#2617](https://github.com/cosmos/cosmos-sdk/issues/2617) [x/mock/simulation] Randomize all genesis parameters
* [#2669](https://github.com/cosmos/cosmos-sdk/issues/2669) [x/stake] Added invarant check to make sure validator's power aligns with its spot in the power store.
* [#1924](https://github.com/cosmos/cosmos-sdk/issues/1924) [x/mock/simulation] Use a transition matrix for block size
* [#2660](https://github.com/cosmos/cosmos-sdk/issues/2660) [x/mock/simulation] Staking transactions get tested far more frequently
* [#2610](https://github.com/cosmos/cosmos-sdk/issues/2610) [x/stake] Block redelegation to and from the same validator
* [#2652](https://github.com/cosmos/cosmos-sdk/issues/2652) [x/auth] Add benchmark for get and set account
* [#2685](https://github.com/cosmos/cosmos-sdk/issues/2685) [store] Add general merkle absence proof (also for empty substores)
* [#2708](https://github.com/cosmos/cosmos-sdk/issues/2708) [store] Disallow setting nil values

BUG FIXES

* Gaia
* [#2670](https://github.com/cosmos/cosmos-sdk/issues/2670) [x/stake] fixed incorrect `IterateBondedValidators` and split into two functions: `IterateBondedValidators` and `IterateLastBlockConsValidators`
* [#2691](https://github.com/cosmos/cosmos-sdk/issues/2691) Fix local testnet creation by using a single canonical genesis time
* [#2648](https://github.com/cosmos/cosmos-sdk/issues/2648) [gaiad] Fix `gaiad export` / `gaiad import` consistency, test in CI

* SDK
* [#2625](https://github.com/cosmos/cosmos-sdk/issues/2625) [x/gov] fix AppendTag function usage error
* [#2677](https://github.com/cosmos/cosmos-sdk/issues/2677) [x/stake, x/distribution] various staking/distribution fixes as found by the simulator
* [#2674](https://github.com/cosmos/cosmos-sdk/issues/2674) [types] Fix coin.IsLT() impl, coins.IsLT() impl, and renamed coins.Is\* to coins.IsAll\* (see [#2686](https://github.com/cosmos/cosmos-sdk/issues/2686))
* [#2711](https://github.com/cosmos/cosmos-sdk/issues/2711) [x/stake] Add commission data to `MsgCreateValidator` signature bytes.
* Temporarily disable insecure mode for Gaia Lite

## 0.25.0

*October 24th, 2018*.

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

    * [x/stake] Validator.Owner renamed to Validator.Operator
    * [#595](https://github.com/cosmos/cosmos-sdk/issues/595) Connections to the REST server are now secured using Transport Layer Security by default. The --insecure flag is provided to switch back to insecure HTTP.
    * [gaia-lite] [#2258](https://github.com/cosmos/cosmos-sdk/issues/2258) Split `GET stake/delegators/{delegatorAddr}` into `GET stake/delegators/{delegatorAddr}/delegations`, `GET stake/delegators/{delegatorAddr}/unbonding_delegations` and `GET stake/delegators/{delegatorAddr}/redelegations`

* Gaia CLI (`gaiacli`)

    * [x/stake] Validator.Owner renamed to Validator.Operator
    * [cli] unsafe_reset_all, show_validator, and show_node_id have been renamed to unsafe-reset-all, show-validator, and show-node-id
    * [cli] [#1983](https://github.com/cosmos/cosmos-sdk/issues/1983) --print-response now defaults to true in commands that create and send a transaction
    * [cli] [#1983](https://github.com/cosmos/cosmos-sdk/issues/1983) you can now pass --pubkey or --address to gaiacli keys show to return a plaintext representation of the key's address or public key for use with other commands
    * [cli] [#2061](https://github.com/cosmos/cosmos-sdk/issues/2061) changed proposalID in governance REST endpoints to proposal-id
    * [cli] [#2014](https://github.com/cosmos/cosmos-sdk/issues/2014) `gaiacli advanced` no longer exists - to access `ibc`, `rest-server`, and `validator-set` commands use `gaiacli ibc`, `gaiacli rest-server`, and `gaiacli tendermint`, respectively
    * [makefile] `get_vendor_deps` no longer updates lock file it just updates vendor directory. Use `update_vendor_deps` to update the lock file. [#2152](https://github.com/cosmos/cosmos-sdk/pull/2152)
    * [cli] [#2221](https://github.com/cosmos/cosmos-sdk/issues/2221) All commands that
    utilize a validator's operator address must now use the new Bech32 prefix,
    `cosmosvaloper`.
    * [cli] [#2190](https://github.com/cosmos/cosmos-sdk/issues/2190) `gaiacli init --gen-txs` is now `gaiacli init --with-txs` to reduce confusion
    * [cli] [#2073](https://github.com/cosmos/cosmos-sdk/issues/2073) --from can now be either an address or a key name
    * [cli] [#1184](https://github.com/cosmos/cosmos-sdk/issues/1184) Subcommands reorganisation, see [#2390](https://github.com/cosmos/cosmos-sdk/pull/2390) for a comprehensive list of changes.
    * [cli] [#2524](https://github.com/cosmos/cosmos-sdk/issues/2524) Add support offline mode to `gaiacli tx sign`. Lookups are not performed if the flag `--offline` is on.
    * [cli] [#2570](https://github.com/cosmos/cosmos-sdk/pull/2570) Add commands to query deposits on proposals

* Gaia

    * Make the transient store key use a distinct store key. [#2013](https://github.com/cosmos/cosmos-sdk/pull/2013)
    * [x/stake] [#1901](https://github.com/cosmos/cosmos-sdk/issues/1901) Validator type's Owner field renamed to Operator; Validator's GetOwner() renamed accordingly to comply with the SDK's Validator interface.
    * [docs] [#2001](https://github.com/cosmos/cosmos-sdk/pull/2001) Update slashing spec for slashing period
    * [x/stake, x/slashing] [#1305](https://github.com/cosmos/cosmos-sdk/issues/1305) - Rename "revoked" to "jailed"
    * [x/stake] [#1676] Revoked and jailed validators put into the unbonding state
    * [x/stake] [#1877] Redelegations/unbonding-delegation from unbonding validator have reduced time
    * [x/slashing] [#1789](https://github.com/cosmos/cosmos-sdk/issues/1789) Slashing changes for Tendermint validator set offset (NextValSet)
    * [x/stake] [#2040](https://github.com/cosmos/cosmos-sdk/issues/2040) Validator
    operator type has now changed to `sdk.ValAddress`
    * [x/stake] [#2221](https://github.com/cosmos/cosmos-sdk/issues/2221) New
    Bech32 prefixes have been introduced for a validator's consensus address and
    public key: `cosmosvalcons` and `cosmosvalconspub` respectively. Also, existing Bech32 prefixes have been
    renamed for accounts and validator operators:
    _ `cosmosaccaddr` / `cosmosaccpub` => `cosmos` / `cosmospub`
    _ `cosmosvaladdr` / `cosmosvalpub` => `cosmosvaloper` / `cosmosvaloperpub`
    * [x/stake] [#1013] TendermintUpdates now uses transient store
    * [x/stake] [#2435](https://github.com/cosmos/cosmos-sdk/issues/2435) Remove empty bytes from the ValidatorPowerRank store key
    * [x/gov] [#2195](https://github.com/cosmos/cosmos-sdk/issues/2195) Governance uses BFT Time
    * [x/gov] [#2256](https://github.com/cosmos/cosmos-sdk/issues/2256) Removed slashing for governance non-voting validators
    * [simulation] [#2162](https://github.com/cosmos/cosmos-sdk/issues/2162) Added back correct supply invariants
    * [x/slashing] [#2430](https://github.com/cosmos/cosmos-sdk/issues/2430) Simulate more slashes, check if validator is jailed before jailing
    * [x/stake] [#2393](https://github.com/cosmos/cosmos-sdk/issues/2393) Removed `CompleteUnbonding` and `CompleteRedelegation` Msg types, and instead added unbonding/redelegation queues to endblocker
    * [x/mock/simulation] [#2501](https://github.com/cosmos/cosmos-sdk/issues/2501) Simulate transactions & invariants for fee distribution, and fix bugs discovered in the process
        * [x/auth] Simulate random fee payments
        * [cmd/gaia/app] Simulate non-zero inflation
        * [x/stake] Call hooks correctly in several cases related to delegation/validator updates
        * [x/stake] Check full supply invariants, including yet-to-be-withdrawn fees
        * [x/stake] Remove no-longer-in-use store key
        * [x/slashing] Call hooks correctly when a validator is slashed
        * [x/slashing] Truncate withdrawals (unbonding, redelegation) and burn change
        * [x/mock/simulation] Ensure the simulation cannot set a proposer address of nil
        * [x/mock/simulation] Add more event logs on begin block / end block for clarity
        * [x/mock/simulation] Correctly set validator power in abci.RequestBeginBlock
        * [x/minting] Correctly call stake keeper to track inflated supply
        * [x/distribution] Sanity check for nonexistent rewards
        * [x/distribution] Truncate withdrawals and return change to the community pool
        * [x/distribution] Add sanity checks for incorrect accum / total accum relations
        * [x/distribution] Correctly calculate total power using Tendermint updates
        * [x/distribution] Simulate withdrawal transactions
        * [x/distribution] Fix a bug where the fee pool was not correctly tracked on WithdrawDelegatorRewardsAll
    * [x/stake] [#1673](https://github.com/cosmos/cosmos-sdk/issues/1673) Validators are no longer deleted until they can no longer possibly be slashed
    * [#1890](https://github.com/cosmos/cosmos-sdk/issues/1890) Start chain with initial state + sequence of transactions
        * [cli] Rename `gaiad init gentx` to `gaiad gentx`.
        * [cli] Add `--skip-genesis` flag to `gaiad init` to prevent `genesis.json` generation.
        * Drop `GenesisTx` in favor of a signed `StdTx` with only one `MsgCreateValidator` message.
        * [cli] Port `gaiad init` and `gaiad testnet` to work with `StdTx` genesis transactions.
        * [cli] Add `--moniker` flag to `gaiad init` to override moniker when generating `genesis.json` - i.e. it takes effect when running with the `--with-txs` flag, it is ignored otherwise.

* SDK

    * [core] [#2219](https://github.com/cosmos/cosmos-sdk/issues/2219) Update to Tendermint 0.24.0
        * Validator set updates delayed by one block
        * BFT timestamp that can safely be used by applications
        * Fixed maximum block size enforcement
    * [core] [#1807](https://github.com/cosmos/cosmos-sdk/issues/1807) Switch from use of rational to decimal
    * [types] [#1901](https://github.com/cosmos/cosmos-sdk/issues/1901) Validator interface's GetOwner() renamed to GetOperator()
    * [x/slashing] [#2122](https://github.com/cosmos/cosmos-sdk/pull/2122) - Implement slashing period
    * [types] [#2119](https://github.com/cosmos/cosmos-sdk/issues/2119) Parsed error messages and ABCI log errors to make them more human readable.
    * [types] [#2407](https://github.com/cosmos/cosmos-sdk/issues/2407) MulInt method added to big decimal in order to improve efficiency of slashing
    * [simulation] Rename TestAndRunTx to Operation [#2153](https://github.com/cosmos/cosmos-sdk/pull/2153)
    * [simulation] Remove log and testing.TB from Operation and Invariants, in favor of using errors [#2282](https://github.com/cosmos/cosmos-sdk/issues/2282)
    * [simulation] Remove usage of keys and addrs in the types, in favor of simulation.Account [#2384](https://github.com/cosmos/cosmos-sdk/issues/2384)
    * [tools] Removed gocyclo [#2211](https://github.com/cosmos/cosmos-sdk/issues/2211)
    * [baseapp] Remove `SetTxDecoder` in favor of requiring the decoder be set in baseapp initialization. [#1441](https://github.com/cosmos/cosmos-sdk/issues/1441)
    * [baseapp] [#1921](https://github.com/cosmos/cosmos-sdk/issues/1921) Add minimumFees field to BaseApp.
    * [store] Change storeInfo within the root multistore to use tmhash instead of ripemd160 [#2308](https://github.com/cosmos/cosmos-sdk/issues/2308)
    * [codec] [#2324](https://github.com/cosmos/cosmos-sdk/issues/2324) All referrences to wire have been renamed to codec. Additionally, wire.NewCodec is now codec.New().
    * [types] [#2343](https://github.com/cosmos/cosmos-sdk/issues/2343) Make sdk.Msg have a names field, to facilitate automatic tagging.
    * [baseapp] [#2366](https://github.com/cosmos/cosmos-sdk/issues/2366) Automatically add action tags to all messages
    * [x/auth] [#2377](https://github.com/cosmos/cosmos-sdk/issues/2377) auth.StdSignMsg -> txbuilder.StdSignMsg
    * [x/staking] [#2244](https://github.com/cosmos/cosmos-sdk/issues/2244) staking now holds a consensus-address-index instead of a consensus-pubkey-index
    * [x/staking] [#2236](https://github.com/cosmos/cosmos-sdk/issues/2236) more distribution hooks for distribution
    * [x/stake] [#2394](https://github.com/cosmos/cosmos-sdk/issues/2394) Split up UpdateValidator into distinct state transitions applied only in EndBlock
    * [x/slashing] [#2480](https://github.com/cosmos/cosmos-sdk/issues/2480) Fix signing info handling bugs & faulty slashing
    * [x/stake] [#2412](https://github.com/cosmos/cosmos-sdk/issues/2412) Added an unbonding validator queue to EndBlock to automatically update validator.Status when finished Unbonding
    * [x/stake] [#2500](https://github.com/cosmos/cosmos-sdk/issues/2500) Block conflicting redelegations until we add an index
    * [x/params] Global Paramstore refactored
    * [types] [#2506](https://github.com/cosmos/cosmos-sdk/issues/2506) sdk.Dec MarshalJSON now marshals as a normal Decimal, with 10 digits of decimal precision
    * [x/stake] [#2508](https://github.com/cosmos/cosmos-sdk/issues/2508) Utilize Tendermint power for validator power key
    * [x/stake] [#2531](https://github.com/cosmos/cosmos-sdk/issues/2531) Remove all inflation logic
    * [x/mint] [#2531](https://github.com/cosmos/cosmos-sdk/issues/2531) Add minting module and inflation logic
    * [x/auth] [#2540](https://github.com/cosmos/cosmos-sdk/issues/2540) Rename `AccountMapper` to `AccountKeeper`.
    * [types] [#2456](https://github.com/cosmos/cosmos-sdk/issues/2456) Renamed msg.Name() and msg.Type() to msg.Type() and msg.Route() respectively

* Tendermint
    * Update tendermint version from v0.23.0 to v0.25.0, notable changes
        * Mempool now won't build too large blocks, or too computationally expensive blocks
        * Maximum tx sizes and gas are now removed, and are implicitly the blocks maximums
        * ABCI validators no longer send the pubkey. The pubkey is only sent in validator updates
        * Validator set changes are now delayed by one block
        * Block header now includes the next validator sets hash
        * BFT time is implemented
        * Secp256k1 signature format has changed
        * There is now a threshold multisig format
        * See the [tendermint changelog](https://github.com/tendermint/tendermint/blob/master/CHANGELOG.md) for other changes.

FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)

    * [gaia-lite] Endpoints to query staking pool and params
    * [gaia-lite] [#2110](https://github.com/cosmos/cosmos-sdk/issues/2110) Add support for `simulate=true` requests query argument to endpoints that send txs to run simulations of transactions
    * [gaia-lite] [#966](https://github.com/cosmos/cosmos-sdk/issues/966) Add support for `generate_only=true` query argument to generate offline unsigned transactions
    * [gaia-lite] [#1953](https://github.com/cosmos/cosmos-sdk/issues/1953) Add /sign endpoint to sign transactions generated with `generate_only=true`.
    * [gaia-lite] [#1954](https://github.com/cosmos/cosmos-sdk/issues/1954) Add /broadcast endpoint to broadcast transactions signed by the /sign endpoint.
    * [gaia-lite] [#2113](https://github.com/cosmos/cosmos-sdk/issues/2113) Rename `/accounts/{address}/send` to `/bank/accounts/{address}/transfers`, rename `/accounts/{address}` to `/auth/accounts/{address}`, replace `proposal-id` with `proposalId` in all gov endpoints
    * [gaia-lite] [#2478](https://github.com/cosmos/cosmos-sdk/issues/2478) Add query gov proposal's deposits endpoint
    * [gaia-lite] [#2477](https://github.com/cosmos/cosmos-sdk/issues/2477) Add query validator's outgoing redelegations and unbonding delegations endpoints

* Gaia CLI (`gaiacli`)

    * [cli] Cmds to query staking pool and params
    * [gov][cli] [#2062](https://github.com/cosmos/cosmos-sdk/issues/2062) added `--proposal` flag to `submit-proposal` that allows a JSON file containing a proposal to be passed in
    * [#2040](https://github.com/cosmos/cosmos-sdk/issues/2040) Add `--bech` to `gaiacli keys show` and respective REST endpoint to
    provide desired Bech32 prefix encoding
    * [cli] [#2047](https://github.com/cosmos/cosmos-sdk/issues/2047) [#2306](https://github.com/cosmos/cosmos-sdk/pull/2306) Passing --gas=simulate triggers a simulation of the tx before the actual execution.
    The gas estimate obtained via the simulation will be used as gas limit in the actual execution.
    * [cli] [#2047](https://github.com/cosmos/cosmos-sdk/issues/2047) The --gas-adjustment flag can be used to adjust the estimate obtained via the simulation triggered by --gas=simulate.
    * [cli] [#2110](https://github.com/cosmos/cosmos-sdk/issues/2110) Add --dry-run flag to perform a simulation of a transaction without broadcasting it. The --gas flag is ignored as gas would be automatically estimated.
    * [cli] [#2204](https://github.com/cosmos/cosmos-sdk/issues/2204) Support generating and broadcasting messages with multiple signatures via command line:
        * [#966](https://github.com/cosmos/cosmos-sdk/issues/966) Add --generate-only flag to build an unsigned transaction and write it to STDOUT.
        * [#1953](https://github.com/cosmos/cosmos-sdk/issues/1953) New `sign` command to sign transactions generated with the --generate-only flag.
        * [#1954](https://github.com/cosmos/cosmos-sdk/issues/1954) New `broadcast` command to broadcast transactions generated offline and signed with the `sign` command.
    * [cli] [#2220](https://github.com/cosmos/cosmos-sdk/issues/2220) Add `gaiacli config` feature to interactively create CLI config files to reduce the number of required flags
    * [stake][cli] [#1672](https://github.com/cosmos/cosmos-sdk/issues/1672) Introduced
    new commission flags for validator commands `create-validator` and `edit-validator`.
    * [stake][cli] [#1890](https://github.com/cosmos/cosmos-sdk/issues/1890) Add `--genesis-format` flag to `gaiacli tx create-validator` to produce transactions in genesis-friendly format.
    * [cli][#2554](https://github.com/cosmos/cosmos-sdk/issues/2554) Make `gaiacli keys show` multisig ready.

* Gaia

    * [cli] [#2170](https://github.com/cosmos/cosmos-sdk/issues/2170) added ability to show the node's address via `gaiad tendermint show-address`
    * [simulation] [#2313](https://github.com/cosmos/cosmos-sdk/issues/2313) Reworked `make test_sim_gaia_slow` to `make test_sim_gaia_full`, now simulates from multiple starting seeds in parallel
    * [cli] [#1921](https://github.com/cosmos/cosmos-sdk/issues/1921)
        * New configuration file `gaiad.toml` is now created to host Gaia-specific configuration.
        * New --minimum_fees/minimum_fees flag/config option to set a minimum fee.

* SDK
    * [querier] added custom querier functionality, so ABCI query requests can be handled by keepers
    * [simulation] [#1924](https://github.com/cosmos/cosmos-sdk/issues/1924) allow operations to specify future operations
    * [simulation] [#1924](https://github.com/cosmos/cosmos-sdk/issues/1924) Add benchmarking capabilities, with makefile commands "test_sim_gaia_benchmark, test_sim_gaia_profile"
    * [simulation] [#2349](https://github.com/cosmos/cosmos-sdk/issues/2349) Add time-based future scheduled operations to simulator
    * [x/auth] [#2376](https://github.com/cosmos/cosmos-sdk/issues/2376) Remove FeePayer() from StdTx
    * [x/stake] [#1672](https://github.com/cosmos/cosmos-sdk/issues/1672) Implement
    basis for the validator commission model.
    * [x/auth] Support account removal in the account mapper.

IMPROVEMENTS

* [tools] Improved terraform and ansible scripts for infrastructure deployment
* [tools] Added ansible script to enable process core dumps

* Gaia REST API (`gaiacli advanced rest-server`)

    * [x/stake] [#2000](https://github.com/cosmos/cosmos-sdk/issues/2000) Added tests for new staking endpoints
    * [gaia-lite] [#2445](https://github.com/cosmos/cosmos-sdk/issues/2445) Standarized REST error responses
    * [gaia-lite] Added example to Swagger specification for /keys/seed.
    * [x/stake] Refactor REST utils

* Gaia CLI (`gaiacli`)

    * [cli] [#2060](https://github.com/cosmos/cosmos-sdk/issues/2060) removed `--select` from `block` command
    * [cli] [#2128](https://github.com/cosmos/cosmos-sdk/issues/2128) fixed segfault when exporting directly after `gaiad init`
    * [cli] [#1255](https://github.com/cosmos/cosmos-sdk/issues/1255) open KeyBase in read-only mode
    for query-purpose CLI commands
    * [docs] Added commands for querying governance deposits, votes and tally

* Gaia

    * [x/stake] [#2023](https://github.com/cosmos/cosmos-sdk/pull/2023) Terminate iteration loop in `UpdateBondedValidators` and `UpdateBondedValidatorsFull` when the first revoked validator is encountered and perform a sanity check.
    * [x/auth] Signature verification's gas cost now accounts for pubkey type. [#2046](https://github.com/tendermint/tendermint/pull/2046)
    * [x/stake] [x/slashing] Ensure delegation invariants to jailed validators [#1883](https://github.com/cosmos/cosmos-sdk/issues/1883).
    * [x/stake] Improve speed of GetValidator, which was shown to be a performance bottleneck. [#2046](https://github.com/tendermint/tendermint/pull/2200)
    * [x/stake] [#2435](https://github.com/cosmos/cosmos-sdk/issues/2435) Improve memory efficiency of getting the various store keys
    * [genesis] [#2229](https://github.com/cosmos/cosmos-sdk/issues/2229) Ensure that there are no duplicate accounts or validators in the genesis state.
    * [genesis] [#2450](https://github.com/cosmos/cosmos-sdk/issues/2450) Validate staking genesis parameters.
    * Add SDK validation to `config.toml` (namely disabling `create_empty_blocks`) [#1571](https://github.com/cosmos/cosmos-sdk/issues/1571)
    * [#1941](https://github.com/cosmos/cosmos-sdk/issues/1941) Version is now inferred via `git describe --tags`.
    * [x/distribution] [#1671](https://github.com/cosmos/cosmos-sdk/issues/1671) add distribution types and tests

* SDK
    * [tools] Make get_vendor_deps deletes `.vendor-new` directories, in case scratch files are present.
    * [spec] Added simple piggy bank distribution spec
    * [cli] [#1632](https://github.com/cosmos/cosmos-sdk/issues/1632) Add integration tests to ensure `basecoind init && basecoind` start sequences run successfully for both `democoin` and `basecoin` examples.
    * [store] Speedup IAVL iteration, and consequently everything that requires IAVL iteration. [#2143](https://github.com/cosmos/cosmos-sdk/issues/2143)
    * [store] [#1952](https://github.com/cosmos/cosmos-sdk/issues/1952), [#2281](https://github.com/cosmos/cosmos-sdk/issues/2281) Update IAVL dependency to v0.11.0
    * [simulation] Make timestamps randomized [#2153](https://github.com/cosmos/cosmos-sdk/pull/2153)
    * [simulation] Make logs not just pure strings, speeding it up by a large factor at greater block heights [#2282](https://github.com/cosmos/cosmos-sdk/issues/2282)
    * [simulation] Add a concept of weighting the operations [#2303](https://github.com/cosmos/cosmos-sdk/issues/2303)
    * [simulation] Logs get written to file if large, and also get printed on panics [#2285](https://github.com/cosmos/cosmos-sdk/issues/2285)
    * [simulation] Bank simulations now makes testing auth configurable [#2425](https://github.com/cosmos/cosmos-sdk/issues/2425)
    * [gaiad] [#1992](https://github.com/cosmos/cosmos-sdk/issues/1992) Add optional flag to `gaiad testnet` to make config directory of daemon (default `gaiad`) and cli (default `gaiacli`) configurable
    * [x/stake] Add stake `Queriers` for Gaia-lite endpoints. This increases the staking endpoints performance by reusing the staking `keeper` logic for queries. [#2249](https://github.com/cosmos/cosmos-sdk/pull/2149)
    * [store] [#2017](https://github.com/cosmos/cosmos-sdk/issues/2017) Refactor
    gas iterator gas consumption to only consume gas for iterator creation and `Next`
    calls which includes dynamic consumption of value length.
    * [types/decimal] [#2378](https://github.com/cosmos/cosmos-sdk/issues/2378) - Added truncate functionality to decimal
    * [client] [#1184](https://github.com/cosmos/cosmos-sdk/issues/1184) Remove unused `client/tx/sign.go`.
    * [tools] [#2464](https://github.com/cosmos/cosmos-sdk/issues/2464) Lock binary dependencies to a specific version
    * #2573 [x/distribution] add accum invariance

BUG FIXES

* Gaia CLI (`gaiacli`)

    * [cli] [#1997](https://github.com/cosmos/cosmos-sdk/issues/1997) Handle panics gracefully when `gaiacli stake {delegation,unbond}` fail to unmarshal delegation.
    * [cli] [#2265](https://github.com/cosmos/cosmos-sdk/issues/2265) Fix JSON formatting of the `gaiacli send` command.
    * [cli] [#2547](https://github.com/cosmos/cosmos-sdk/issues/2547) Mark --to and --amount as required flags for `gaiacli tx send`.

* Gaia

    * [x/stake] Return correct Tendermint validator update set on `EndBlocker` by not
    including non previously bonded validators that have zero power. [#2189](https://github.com/cosmos/cosmos-sdk/issues/2189)
    * [docs] Fixed light client section links

* SDK
    * [#1988](https://github.com/cosmos/cosmos-sdk/issues/1988) Make us compile on OpenBSD (disable ledger)
    * [#2105](https://github.com/cosmos/cosmos-sdk/issues/2105) Fix DB Iterator leak, which may leak a go routine.
    * [ledger] [#2064](https://github.com/cosmos/cosmos-sdk/issues/2064) Fix inability to sign and send transactions via the LCD by
    loading a Ledger device at runtime.
    * [#2158](https://github.com/cosmos/cosmos-sdk/issues/2158) Fix non-deterministic ordering of validator iteration when slashing in `gov EndBlocker`
    * [simulation] [#1924](https://github.com/cosmos/cosmos-sdk/issues/1924) Make simulation stop on SIGTERM
    * [#2388](https://github.com/cosmos/cosmos-sdk/issues/2388) Remove dependency on deprecated tendermint/tmlibs repository.
    * [#2416](https://github.com/cosmos/cosmos-sdk/issues/2416) Refactored `InitializeTestLCD` to properly include proposing validator in genesis state.
    * #2573 [x/distribution] accum invariance bugfix
    * #2573 [x/slashing] unbonding-delegation slashing invariance bugfix

## 0.24.2

*August 22nd, 2018*.

BUG FIXES

* Tendermint
    * Fix unbounded consensus WAL growth

## 0.24.1

*August 21st, 2018*.

BUG FIXES

* Gaia
    * [x/slashing] Evidence tracking now uses validator address instead of validator pubkey

## 0.24.0

*August 13th, 2018*.

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

    * [x/stake] [#1880](https://github.com/cosmos/cosmos-sdk/issues/1880) More REST-ful endpoints (large refactor)
    * [x/slashing] [#1866](https://github.com/cosmos/cosmos-sdk/issues/1866) `/slashing/signing_info` takes cosmosvalpub instead of cosmosvaladdr
    * use time.Time instead of int64 for time. See Tendermint v0.23.0
    * Signatures are no longer Amino encoded with prefixes (just encoded as raw
    bytes) - see Tendermint v0.23.0

* Gaia CLI (`gaiacli`)

    * [x/stake] change `--keybase-sig` to `--identity`
    * [x/stake] [#1828](https://github.com/cosmos/cosmos-sdk/issues/1828) Force user to specify amount on create-validator command by removing default
    * [x/gov] Change `--proposalID` to `--proposal-id`
    * [x/stake, x/gov] [#1606](https://github.com/cosmos/cosmos-sdk/issues/1606) Use `--from` instead of adhoc flags like `--address-validator`
    and `--proposer` to indicate the sender address.
    * [#1551](https://github.com/cosmos/cosmos-sdk/issues/1551) Remove `--name` completely
    * Genesis/key creation (`gaiad init`) now supports user-provided key passwords

* Gaia

    * [x/stake] Inflation doesn't use rationals in calculation (performance boost)
    * [x/stake] Persist a map from `addr->pubkey` in the state since BeginBlock
    doesn't provide pubkeys.
    * [x/gov] [#1781](https://github.com/cosmos/cosmos-sdk/issues/1781) Added tags sub-package, changed tags to use dash-case
    * [x/gov] [#1688](https://github.com/cosmos/cosmos-sdk/issues/1688) Governance parameters are now stored in globalparams store
    * [x/gov] [#1859](https://github.com/cosmos/cosmos-sdk/issues/1859) Slash validators who do not vote on a proposal
    * [x/gov] [#1914](https://github.com/cosmos/cosmos-sdk/issues/1914) added TallyResult type that gets stored in Proposal after tallying is finished

* SDK

    * [baseapp] Msgs are no longer run on CheckTx, removed `ctx.IsCheckTx()`
    * [baseapp] NewBaseApp constructor takes sdk.TxDecoder as argument instead of wire.Codec
    * [types] sdk.NewCoin takes sdk.Int, sdk.NewInt64Coin takes int64
    * [x/auth] Default TxDecoder can be found in `x/auth` rather than baseapp
    * [client] [#1551](https://github.com/cosmos/cosmos-sdk/issues/1551): Refactored `CoreContext` to `TxContext` and `QueryContext`
        * Removed all tx related fields and logic (building & signing) to separate
      structure `TxContext` in `x/auth/client/context`

* Tendermint
    * v0.22.5 -> See [Tendermint PR](https://github.com/tendermint/tendermint/pull/1966)
        * change all the cryptography imports.
    * v0.23.0 -> See
    [Changelog](https://github.com/tendermint/tendermint/blob/v0.23.0/CHANGELOG.md#0230)
    and [SDK PR](https://github.com/cosmos/cosmos-sdk/pull/1927)
        * BeginBlock no longer includes crypto.Pubkey
        * use time.Time instead of int64 for time.

FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)

    * [x/gov] Can now query governance proposals by ProposalStatus

* Gaia CLI (`gaiacli`)

    * [x/gov] added `query-proposals` command. Can filter by `depositer`, `voter`, and `status`
    * [x/stake] [#2043](https://github.com/cosmos/cosmos-sdk/issues/2043) Added staking query cli cmds for unbonding-delegations and redelegations

* Gaia

    * [networks] Added ansible scripts to upgrade seed nodes on a network

* SDK
    * [x/mock/simulation] Randomized simulation framework
        * Modules specify invariants and operations, preferably in an x/[module]/simulation package
        * Modules can test random combinations of their own operations
        * Applications can integrate operations and invariants from modules together for an integrated simulation
        * Simulates Tendermint's algorithm for validator set updates
        * Simulates validator signing/downtime with a Markov chain, and occaisional double-signatures
        * Includes simulated operations & invariants for staking, slashing, governance, and bank modules
    * [store] [#1481](https://github.com/cosmos/cosmos-sdk/issues/1481) Add transient store
    * [baseapp] Initialize validator set on ResponseInitChain
    * [baseapp] added BaseApp.Seal - ability to seal baseapp parameters once they've been set
    * [cosmos-sdk-cli] New `cosmos-sdk-cli` tool to quickly initialize a new
    SDK-based project
    * [scripts] added log output monitoring to DataDog using Ansible scripts

IMPROVEMENTS

* Gaia

    * [spec] [#967](https://github.com/cosmos/cosmos-sdk/issues/967) Inflation and distribution specs drastically improved
    * [x/gov] [#1773](https://github.com/cosmos/cosmos-sdk/issues/1773) Votes on a proposal can now be queried
    * [x/gov] Initial governance parameters can now be set in the genesis file
    * [x/stake] [#1815](https://github.com/cosmos/cosmos-sdk/issues/1815) Sped up the processing of `EditValidator` txs.
    * [config] [#1930](https://github.com/cosmos/cosmos-sdk/issues/1930) Transactions indexer indexes all tags by default.
    * [ci] [#2057](https://github.com/cosmos/cosmos-sdk/pull/2057) Run `make localnet-start` on every commit and ensure network reaches at least 10 blocks

* SDK
    * [baseapp] [#1587](https://github.com/cosmos/cosmos-sdk/issues/1587) Allow any alphanumeric character in route
    * [baseapp] Allow any alphanumeric character in route
    * [tools] Remove `rm -rf vendor/` from `make get_vendor_deps`
    * [x/auth] Recover ErrorOutOfGas panic in order to set sdk.Result attributes correctly
    * [x/auth] [#2376](https://github.com/cosmos/cosmos-sdk/issues/2376) No longer runs any signature in a multi-msg, if any account/sequence number is wrong.
    * [x/auth] [#2376](https://github.com/cosmos/cosmos-sdk/issues/2376) No longer charge gas for subtracting fees
    * [x/bank] Unit tests are now table-driven
    * [tests] Add tests to example apps in docs
    * [tests] Fixes ansible scripts to work with AWS too
    * [tests] [#1806](https://github.com/cosmos/cosmos-sdk/issues/1806) CLI tests are now behind the build flag 'cli_test', so go test works on a new repo

BUG FIXES

* Gaia CLI (`gaiacli`)

    * [#1766](https://github.com/cosmos/cosmos-sdk/issues/1766) Fixes bad example for keybase identity
    * [x/stake] [#2021](https://github.com/cosmos/cosmos-sdk/issues/2021) Fixed repeated CLI commands in staking

* Gaia
    * [x/stake] [#2077](https://github.com/cosmos/cosmos-sdk/pull/2077) Fixed invalid cliff power comparison
    * [#1804](https://github.com/cosmos/cosmos-sdk/issues/1804) Fixes gen-tx genesis generation logic temporarily until upstream updates
    * [#1799](https://github.com/cosmos/cosmos-sdk/issues/1799) Fix `gaiad export`
    * [#1839](https://github.com/cosmos/cosmos-sdk/issues/1839) Fixed bug where intra-tx counter wasn't set correctly for genesis validators
    * [x/stake] [#1858](https://github.com/cosmos/cosmos-sdk/issues/1858) Fixed bug where the cliff validator was not updated correctly
    * [tests] [#1675](https://github.com/cosmos/cosmos-sdk/issues/1675) Fix non-deterministic `test_cover`
    * [tests] [#1551](https://github.com/cosmos/cosmos-sdk/issues/1551) Fixed invalid LCD test JSON payload in `doIBCTransfer`
    * [basecoin] Fixes coin transaction failure and account query [discussion](https://forum.cosmos.network/t/unmarshalbinarybare-expected-to-read-prefix-bytes-75fbfab8-since-it-is-registered-concrete-but-got-0a141dfa/664/6)
    * [x/gov] [#1757](https://github.com/cosmos/cosmos-sdk/issues/1757) Fix VoteOption conversion to String
    * [x/stake] [#2083] Fix broken invariant of bonded validator power decrease

## 0.23.1

*July 27th, 2018*.

BUG FIXES

* [tendermint] Update to v0.22.8
    * [consensus, blockchain] Register the Evidence interface so it can be
    marshalled/unmarshalled by the blockchain and consensus reactors

## 0.23.0

*July 25th, 2018*.

BREAKING CHANGES

* [x/stake] Fixed the period check for the inflation calculation

IMPROVEMENTS

* [cli] Improve error messages for all txs when the account doesn't exist
* [tendermint] Update to v0.22.6
    * Updates the crypto imports/API (#1966)
* [x/stake] Add revoked to human-readable validator

BUG FIXES

* [tendermint] Update to v0.22.6
    * Fixes some security vulnerabilities reported in the [Bug Bounty](https://hackerone.com/tendermint)
* [#1797](https://github.com/cosmos/cosmos-sdk/issues/1797) Fix off-by-one error in slashing for downtime
* [#1787](https://github.com/cosmos/cosmos-sdk/issues/1787) Fixed bug where Tally fails due to revoked/unbonding validator
* [#1666](https://github.com/cosmos/cosmos-sdk/issues/1666) Add intra-tx counter to the genesis validators

## 0.22.0

*July 16th, 2018*.

BREAKING CHANGES

* [x/gov] Increase VotingPeriod, DepositPeriod, and MinDeposit

IMPROVEMENTS

* [gaiad] Default config updates:
    * `timeout_commit=5000` so blocks only made every 5s
    * `prof_listen_addr=localhost:6060` so profile server is on by default
    * `p2p.send_rate` and `p2p.recv_rate` increases 10x (~5MB/s)

BUG FIXES

* [server] Fix to actually overwrite default tendermint config

## 0.21.1

*July 14th, 2018*.

BUG FIXES

* [build] Added Ledger build support via `LEDGER_ENABLED=true|false`
    * True by default except when cross-compiling

## 0.21.0

*July 13th, 2018*.

BREAKING CHANGES

* [x/stake] Specify DelegatorAddress in MsgCreateValidator
* [x/stake] Remove the use of global shares in the pool
    * Remove the use of `PoolShares` type in `x/stake/validator` type - replace with `Status` `Tokens` fields
* [x/auth] NewAccountMapper takes a constructor instead of a prototype
* [keys] Keybase.Update function now takes in a function to get the newpass, rather than the password itself

FEATURES

* [baseapp] NewBaseApp now takes option functions as parameters

IMPROVEMENTS

* Updated docs folder to accommodate cosmos.network docs project
* [store] Added support for tracing multi-store operations via `--trace-store`
* [store] Pruning strategy configurable with pruning flag on gaiad start

BUG FIXES

* [#1630](https://github.com/cosmos/cosmos-sdk/issues/1630) - redelegation nolonger removes tokens from the delegator liquid account
* [keys] [#1629](https://github.com/cosmos/cosmos-sdk/issues/1629) - updating password no longer asks for a new password when the first entered password was incorrect
* [lcd] importing an account would create a random account
* [server] 'gaiad init' command family now writes provided name as the moniker in `config.toml`
* [build] Added Ledger build support via `LEDGER_ENABLED=true|false`
    * True by default except when cross-compiling

## 0.20.0

*July 10th, 2018*.

BREAKING CHANGES

* msg.GetSignBytes() returns sorted JSON (by key)
* msg.GetSignBytes() field changes
    * `msg_bytes` -> `msgs`
    * `fee_bytes` -> `fee`
* Update Tendermint to v0.22.2
    * Default ports changed from 466xx to 266xx
    * Amino JSON uses type names instead of prefix bytes
    * ED25519 addresses are the first 20-bytes of the SHA256 of the raw 32-byte
    pubkey (Instead of RIPEMD160)
    * go-crypto, abci, tmlibs have been merged into Tendermint
        * The keys sub-module is now in the SDK
    * Various other fixes
* [auth] Signers of a transaction now only sign over their own account and sequence number
* [auth] Removed MsgChangePubKey
* [auth] Removed SetPubKey from account mapper
* [auth] AltBytes renamed to Memo, now a string, max 100 characters, costs a bit of gas
* [types] `GetMsg()` -> `GetMsgs()` as txs wrap many messages
* [types] Removed GetMemo from Tx (it is still on StdTx)
* [types] renamed rational.Evaluate to rational.Round{Int64, Int}
* [types] Renamed `sdk.Address` to `sdk.AccAddress`/`sdk.ValAddress`
* [types] `sdk.AccAddress`/`sdk.ValAddress` natively marshals to Bech32 in String, Sprintf (when used with `%s`), and MarshalJSON
* [keys] Keybase and Ledger support from go-crypto merged into the SDK in the `crypto` folder
* [cli] Rearranged commands under subcommands
* [x/slashing] Update slashing for unbonding period
    * Slash according to power at time of infraction instead of power at
    time of discovery
    * Iterate through unbonding delegations & redelegations which contributed
    to an infraction, slash them proportional to their stake at the time
    * Add REST endpoint to unrevoke a validator previously revoked for downtime
    * Add REST endpoint to retrieve liveness signing information for a validator
* [x/stake] Remove Tick and add EndBlocker
* [x/stake] most index keys nolonger hold a value - inputs are rearranged to form the desired key
* [x/stake] store-value for delegation, validator, ubd, and red do not hold duplicate information contained store-key
* [x/stake] Introduce concept of unbonding for delegations and validators
    * `gaiacli stake unbond` replaced with `gaiacli stake begin-unbonding`
    * Introduced:
        * `gaiacli stake complete-unbonding`
        * `gaiacli stake begin-redelegation`
        * `gaiacli stake complete-redelegation`
* [lcd] Switch key creation output to return bech32
* [lcd] Removed shorthand CLI flags (`a`, `c`, `n`, `o`)
* [gaiad] genesis transactions now use bech32 addresses / pubkeys
* [gov] VoteStatus renamed to ProposalStatus
* [gov] VoteOption, ProposalType, and ProposalStatus all marshal to string form in JSON

DEPRECATED

* [cli] Deprecated `--name` flag in commands that send txs, in favor of `--from`

FEATURES

* [x/gov] Implemented MVP
    * Supported proposal types: just binary (pass/fail) TextProposals for now
    * Proposals need deposits to be votable; deposits are burned if proposal fails
    * Delegators delegate votes to validator by default but can override (for their stake)
* [gaiacli] Ledger support added
    * You can now use a Ledger with `gaiacli --ledger` for all key-related commands
    * Ledger keys can be named and tracked locally in the key DB
* [gaiacli] You can now attach a simple text-only memo to any transaction, with the `--memo` flag
* [gaiacli] added the following flags for commands that post transactions to the chain:
    * async -- send the tx without waiting for a tendermint response
    * json -- return the output in json format for increased readability
    * print-response -- return the tx response. (includes fields like gas cost)
* [lcd] Queried TXs now include the tx hash to identify each tx
* [mockapp] CompleteSetup() no longer takes a testing parameter
* [x/bank] Add benchmarks for signing and delivering a block with a single bank transaction
    * Run with `cd x/bank && go test --bench=.`
* [tools] make get_tools installs tendermint's linter, and gometalinter
* [tools] Switch gometalinter to the stable version
* [tools] Add the following linters
    * misspell
    * gofmt
    * go vet -composites=false
    * unconvert
    * ineffassign
    * errcheck
    * unparam
    * gocyclo
* [tools] Added `make format` command to automate fixing misspell and gofmt errors.
* [server] Default config now creates a profiler at port 6060, and increase p2p send/recv rates
* [types] Switches internal representation of Int/Uint/Rat to use pointers
* [types] Added MinInt and MinUint functions
* [gaiad] `unsafe_reset_all` now resets addrbook.json
* [democoin] add x/oracle, x/assoc
* [tests] created a randomized testing framework.
    * Currently bank has limited functionality in the framework
    * Auth has its invariants checked within the framework
* [tests] Add WaitForNextNBlocksTM helper method
* [keys] New keys now have 24 word recovery keys, for heightened security

* [keys] Add a temporary method for exporting the private key

IMPROVEMENTS

* [x/bank] Now uses go-wire codec instead of 'encoding/json'
* [x/auth] Now uses go-wire codec instead of 'encoding/json'
* revised use of endblock and beginblock
* [stake] module reorganized to include `types` and `keeper` package
* [stake] keeper always loads the store (instead passing around which doesn't really boost efficiency)
* [stake] edit-validator changes now can use the keyword [do-not-modify] to not modify unspecified `--flag` (aka won't set them to `""` value)
* [stake] offload more generic functionality from the handler into the keeper
* [stake] clearer staking logic
* [types] added common tag constants
* [keys] improve error message when deleting non-existent key
* [gaiacli] improve error messages on `send` and `account` commands
* added contributing guidelines
* [docs] Added commands for governance CLI on testnet README

BUG FIXES

* [x/slashing] [#1510](https://github.com/cosmos/cosmos-sdk/issues/1510) Unrevoked validators cannot un-revoke themselves
* [x/stake] [#1513](https://github.com/cosmos/cosmos-sdk/issues/1513) Validators slashed to zero power are unbonded and removed from the store
* [x/stake] [#1567](https://github.com/cosmos/cosmos-sdk/issues/1567) Validators decreased in power but not unbonded are now updated in Tendermint
* [x/stake] error strings lower case
* [x/stake] pool loose tokens now accounts for unbonding and unbonding tokens not associated with any validator
* [x/stake] fix revoke bytes ordering (was putting revoked candidates at the top of the list)
* [x/stake] bond count was counting revoked validators as bonded, fixed
* [gaia] Added self delegation for validators in the genesis creation
* [lcd] tests now don't depend on raw json text
* Retry on HTTP request failure in CLI tests, add option to retry tests in Makefile
* Fixed bug where chain ID wasn't passed properly in x/bank REST handler, removed Viper hack from ante handler
* Fixed bug where `democli account` didn't decode the account data correctly
* [#872](https://github.com/cosmos/cosmos-sdk/issues/872) - recovery phrases no longer all end in `abandon`
* [#887](https://github.com/cosmos/cosmos-sdk/issues/887) - limit the size of rationals that can be passed in from user input
* [#1052](https://github.com/cosmos/cosmos-sdk/issues/1052) - Make all now works
* [#1258](https://github.com/cosmos/cosmos-sdk/issues/1258) - printing big.rat's can no longer overflow int64
* [#1259](https://github.com/cosmos/cosmos-sdk/issues/1259) - fix bug where certain tests that could have a nil pointer in defer
* [#1343](https://github.com/cosmos/cosmos-sdk/issues/1343) - fixed unnecessary parallelism in CI
* [#1353](https://github.com/cosmos/cosmos-sdk/issues/1353) - CLI: Show pool shares fractions in human-readable format
* [#1367](https://github.com/cosmos/cosmos-sdk/issues/1367) - set ChainID in InitChain
* [#1461](https://github.com/cosmos/cosmos-sdk/issues/1461) - CLI tests now no longer reset your local environment data
* [#1505](https://github.com/cosmos/cosmos-sdk/issues/1505) - `gaiacli stake validator` no longer panics if validator doesn't exist
* [#1565](https://github.com/cosmos/cosmos-sdk/issues/1565) - fix cliff validator persisting when validator set shrinks from max
* [#1287](https://github.com/cosmos/cosmos-sdk/issues/1287) - prevent zero power validators at genesis
* [x/stake] fix bug when unbonding/redelegating using `--shares-percent`
* [#1010](https://github.com/cosmos/cosmos-sdk/issues/1010) - two validators can't bond with the same pubkey anymore

## 0.19.0

*June 13, 2018*.

BREAKING CHANGES

* msg.GetSignBytes() now returns bech32-encoded addresses in all cases
* [lcd] REST end-points now include gas
* sdk.Coin now uses sdk.Int, a big.Int wrapper with 256bit range cap

FEATURES

* [x/auth] Added AccountNumbers to BaseAccount and StdTxs to allow for replay protection with account pruning
* [lcd] added an endpoint to query for the SDK version of the connected node

IMPROVEMENTS

* export command now writes current validator set for Tendermint
* [tests] Application module tests now use a mock application
* [gaiacli] Fix error message when account isn't found when running gaiacli account
* [lcd] refactored to eliminate use of global variables, and interdependent tests
* [tests] Added testnet command to gaiad
* [tests] Added localnet targets to Makefile
* [x/stake] More stake tests added to test ByPower index

FIXES

* Fixes consensus fault on testnet - see postmortem [here](https://github.com/cosmos/cosmos-sdk/issues/1197#issuecomment-396823021)
* [x/stake] bonded inflation removed, non-bonded inflation partially implemented
* [lcd] Switch to bech32 for addresses on all human readable inputs and outputs
* [lcd] fixed tx indexing/querying
* [cli] Added `--gas` flag to specify transaction gas limit
* [gaia] Registered slashing message handler
* [x/slashing] Set signInfo.StartHeight correctly for newly bonded validators

FEATURES

* [docs] Reorganize documentation
* [docs] Update staking spec, create WIP spec for slashing, and fees

## 0.18.0

*June 9, 2018*.

BREAKING CHANGES

* [stake] candidate -> validator throughout (details in refactor comment)
* [stake] delegate-bond -> delegation throughout
* [stake] `gaiacli query validator` takes and argument instead of using the `--address-candidate` flag
* [stake] introduce `gaiacli query delegations`
* [stake] staking refactor
    * ValidatorsBonded store now take sorted pubKey-address instead of validator owner-address,
    is sorted like Tendermint by pk's address
    * store names more understandable
    * removed temporary ToKick store, just needs a local map!
    * removed distinction between candidates and validators
        * everything is now a validator
        * only validators with a status == bonded are actively validating/receiving rewards
    * Introduction of Unbonding fields, lowlevel logic throughout (not fully implemented with queue)
    * Introduction of PoolShares type within validators,
    replaces three rational fields (BondedShares, UnbondingShares, UnbondedShares
* [x/auth] move stuff specific to auth anteHandler to the auth module rather than the types folder. This includes:
    * StdTx (and its related stuff i.e. StdSignDoc, etc)
    * StdFee
    * StdSignature
    * Account interface
    * Related to this organization, I also:
* [x/auth] got rid of AccountMapper interface (in favor of the struct already in auth module)
* [x/auth] removed the FeeHandler function from the AnteHandler, Replaced with FeeKeeper
* [x/auth] Removed GetSignatures() from Tx interface (as different Tx styles might use something different than StdSignature)
* [store] Removed SubspaceIterator and ReverseSubspaceIterator from KVStore interface and replaced them with helper functions in /types
* [cli] rearranged commands under subcommands
* [stake] remove Tick and add EndBlocker
* Switch to bech32cosmos on all human readable inputs and outputs

FEATURES

* [x/auth] Added ability to change pubkey to auth module
* [baseapp] baseapp now has settable functions for filtering peers by address/port & public key
* [sdk] Gas consumption is now measured as transactions are executed
    * Transactions which run out of gas stop execution and revert state changes
    * A "simulate" query has been added to determine how much gas a transaction will need
    * Modules can include their own gas costs for execution of particular message types
* [stake] Seperation of fee distribution to a new module
* [stake] Creation of a validator/delegation generics in `/types`
* [stake] Helper Description of the store in x/stake/store.md
* [stake] removed use of caches in the stake keeper
* [stake] Added REST API
* [Makefile] Added terraform/ansible playbooks to easily create remote testnets on Digital Ocean

BUG FIXES

* [stake] staking delegator shares exchange rate now relative to equivalent-bonded-tokens the validator has instead of bonded tokens
  ^ this is important for unbonded validators in the power store!
* [cli] fixed cli-bash tests
* [ci] added cli-bash tests
* [basecoin] updated basecoin for stake and slashing
* [docs] fixed references to old cli commands
* [docs] Downgraded Swagger to v2 for downstream compatibility
* auto-sequencing transactions correctly
* query sequence via account store
* fixed duplicate pub_key in stake.Validator
* Auto-sequencing now works correctly
* [gaiacli] Fix error message when account isn't found when running gaiacli account

## 0.17.5

*June 5, 2018*.

Update to Tendermint v0.19.9 (Fix evidence reactor, mempool deadlock, WAL panic,
memory leak)

## 0.17.4

*May 31, 2018*.

Update to Tendermint v0.19.7 (WAL fixes and more)

## 0.17.3

*May 29, 2018*.

Update to Tendermint v0.19.6 (fix fast-sync halt)

## 0.17.2

*May 20, 2018*.

Update to Tendermint v0.19.5 (reduce WAL use, bound the mempool and some rpcs, improve logging)

## 0.17.1 (May 17, 2018)

Update to Tendermint v0.19.4 (fixes a consensus bug and improves logging)

## 0.17.0 (May 15, 2018)

BREAKING CHANGES

* [stake] MarshalJSON -> MarshalBinaryLengthPrefixed
* Queries against the store must be prefixed with the path "/store"

FEATURES

* [gaiacli] Support queries for candidates, delegator-bonds
* [gaiad] Added `gaiad export` command to export current state to JSON
* [x/bank] Tx tags with sender/recipient for indexing & later retrieval
* [x/stake] Tx tags with delegator/candidate for delegation & unbonding, and candidate info for declare candidate / edit validator

IMPROVEMENTS

* [gaiad] Update for Tendermint v0.19.3 (improve `/dump_consensus_state` and add
  `/consensus_state`)
* [spec/ibc] Added spec!
* [spec/stake] Cleanup structure, include details about slashing and
  auto-unbonding
* [spec/governance] Fixup some names and pseudocode
* NOTE: specs are still a work-in-progress ...

BUG FIXES

* Auto-sequencing now works correctly

## 0.16.0 (May 14th, 2018)

BREAKING CHANGES

* Move module REST/CLI packages to x/[module]/client/rest and x/[module]/client/cli
* Gaia simple-staking bond and unbond functions replaced
* [stake] Delegator bonds now store the height at which they were updated
* All module keepers now require a codespace, see basecoin or democoin for usage
* Many changes to names throughout
    * Type as a prefix naming convention applied (ex. BondMsg -> MsgBond)
    * Removed redundancy in names (ex. stake.StakingKeeper -> stake.Keeper)
* Removed SealedAccountMapper
* gaiad init now requires use of `--name` flag
* Removed Get from Msg interface
* types/rational now extends big.Rat

FEATURES:

* Gaia stake commands include, CreateValidator, EditValidator, Delegate, Unbond
* MountStoreWithDB without providing a custom store works.
* Repo is now lint compliant / GoMetaLinter with tendermint-lint integrated into CI
* Better key output, pubkey go-amino hex bytes now output by default
* gaiad init overhaul
    * Create genesis transactions with `gaiad init gen-tx`
    * New genesis account keys are automatically added to the client keybase (introduce `--client-home` flag)
    * Initialize with genesis txs using `--gen-txs` flag
* Context now has access to the application-configured logger
* Add (non-proof) subspace query helper functions
* Add more staking query functions: candidates, delegator-bonds

BUG FIXES

* Gaia now uses stake, ported from github.com/cosmos/gaia

## 0.15.1 (April 29, 2018)

IMPROVEMENTS:

* Update Tendermint to v0.19.1 (includes many rpc fixes)

## 0.15.0 (April 29, 2018)

NOTE: v0.15.0 is a large breaking change that updates the encoding scheme to use
[Amino](github.com/tendermint/go-amino).

For details on how this changes encoding for public keys and addresses,
see the [docs](https://github.com/tendermint/tendermint/blob/v0.19.1/docs/specification/new-spec/encoding.md#public-key-cryptography).

BREAKING CHANGES

* Remove go-wire, use go-amino
* [store] Add `SubspaceIterator` and `ReverseSubspaceIterator` to `KVStore` interface
* [basecoin] NewBasecoinApp takes a `dbm.DB` and uses namespaced DBs for substores

FEATURES:

* Add CacheContext
* Add auto sequencing to client
* Add FeeHandler to ante handler

BUG FIXES

* MountStoreWithDB without providing a custom store works.

## 0.14.1 (April 9, 2018)

BUG FIXES

* [gaiacli] Fix all commands (just a duplicate of basecli for now)

## 0.14.0 (April 9, 2018)

BREAKING CHANGES:

* [client/builder] Renamed to `client/core` and refactored to use a CoreContext
  struct
* [server] Refactor to improve useability and de-duplicate code
* [types] `Result.ToQuery -> Error.QueryResult`
* [makefile] `make build` and `make install` only build/install `gaiacli` and
  `gaiad`. Use `make build_examples` and `make install_examples` for
  `basecoind/basecli` and `democoind/democli`
* [staking] Various fixes/improvements

FEATURES:

* [democoin] Added Proof-of-Work module

BUG FIXES

* [client] Reuse Tendermint RPC client to avoid excessive open files
* [client] Fix setting log level
* [basecoin] Sort coins in genesis

## 0.13.1 (April 3, 2018)

BUG FIXES

* [x/ibc] Fix CLI and relay for IBC txs
* [x/stake] Various fixes/improvements

## 0.13.0 (April 2, 2018)

BREAKING CHANGES

* [basecoin] Remove cool/sketchy modules -> moved to new `democoin`
* [basecoin] NewBasecoinApp takes a `map[string]dbm.DB` as temporary measure
  to allow mounting multiple stores with their own DB until they can share one
* [x/staking] Renamed to `simplestake`
* [builder] Functions don't take `passphrase` as argument
* [server] GenAppParams returns generated seed and address
* [basecoind] `init` command outputs JSON of everything necessary for testnet
* [basecoind] `basecoin.db -> data/basecoin.db`
* [basecli] `data/keys.db -> keys/keys.db`

FEATURES

* [types] `Coin` supports direct arithmetic operations
* [basecoind] Add `show_validator` and `show_node_id` commands
* [x/stake] Initial merge of full staking module!
* [democoin] New example application to demo custom modules

IMPROVEMENTS

* [makefile] `make install`
* [testing] Use `/tmp` for directories so they don't get left in the repo

BUG FIXES

* [basecoin] Allow app to be restarted
* [makefile] Fix build on Windows
* [basecli] Get confirmation before overriding key with same name

## 0.12.0 (March 27 2018)

BREAKING CHANGES

* Revert to old go-wire for now
* glide -> godep
* [types] ErrBadNonce -> ErrInvalidSequence
* [types] Replace tx.GetFeePayer with FeePayer(tx) - returns the first signer
* [types] NewStdTx takes the Fee
* [types] ParseAccount -> AccountDecoder; ErrTxParse -> ErrTxDecoder
* [x/auth] AnteHandler deducts fees
* [x/bank] Move some errors to `types`
* [x/bank] Remove sequence and signature from Input

FEATURES

* [examples/basecoin] New cool module to demonstrate use of state and custom transactions
* [basecoind] `show_node_id` command
* [lcd] Implement the Light Client Daemon and endpoints
* [types/stdlib] Queue functionality
* [store] Subspace iterator on IAVLTree
* [types] StdSignDoc is the document that gets signed (chainid, msg, sequence, fee)
* [types] CodeInvalidPubKey
* [types] StdFee, and StdTx takes the StdFee
* [specs] Progression of MVPs for IBC
* [x/ibc] Initial shell of IBC functionality (no proofs)
* [x/simplestake] Simple staking module with bonding/unbonding

IMPROVEMENTS

* Lots more tests!
* [client/builder] Helpers for forming and signing transactions
* [types] sdk.Address
* [specs] Staking

BUG FIXES

* [x/auth] Fix setting pubkey on new account
* [x/auth] Require signatures to include the sequences
* [baseapp] Dont panic on nil handler
* [basecoin] Check for empty bytes in account and tx

## 0.11.0 (March 1, 2017)

BREAKING CHANGES

* [examples] dummy -> kvstore
* [examples] Remove gaia
* [examples/basecoin] MakeTxCodec -> MakeCodec
* [types] CommitMultiStore interface has new `GetCommitKVStore(key StoreKey) CommitKVStore` method

FEATURES

* [examples/basecoin] CLI for `basecli` and `basecoind` (!)
* [baseapp] router.AddRoute returns Router

IMPROVEMENTS

* [baseapp] Run msg handlers on CheckTx
* [docs] Add spec for REST API
* [all] More tests!

BUG FIXES

* [baseapp] Fix panic on app restart
* [baseapp] InitChain does not call Commit
* [basecoin] Remove IBCStore because mounting multiple stores is currently broken

## 0.10.0 (February 20, 2017)

BREAKING CHANGES

* [baseapp] NewBaseApp(logger, db)
* [baseapp] NewContext(isCheckTx, header)
* [x/bank] CoinMapper -> CoinKeeper

FEATURES

* [examples/gaia] Mock CLI !
* [baseapp] InitChainer, BeginBlocker, EndBlocker
* [baseapp] MountStoresIAVL

IMPROVEMENTS

* [docs] Various improvements.
* [basecoin] Much simpler :)

BUG FIXES

* [baseapp] initialize and reset msCheck and msDeliver properly

## 0.9.0 (February 13, 2017)

BREAKING CHANGES

* Massive refactor. Basecoin works. Still needs <3

## 0.8.1

* Updates for dependencies

## 0.8.0 (December 18, 2017)

* Updates for dependencies

## 0.7.1 (October 11, 2017)

IMPROVEMENTS:

* server/commands: GetInitCmd takes list of options

## 0.7.0 (October 11, 2017)

BREAKING CHANGES:

* Everything has changed, and it's all about to change again, so don't bother using it yet!

## 0.6.2 (July 27, 2017)

IMPROVEMENTS:

* auto-test all tutorials to detect breaking changes
* move deployment scripts from `/scripts` to `/publish` for clarity

BUG FIXES:

* `basecoin init` ensures the address in genesis.json is valid
* fix bug that certain addresses couldn't receive ibc packets

## 0.6.1 (June 28, 2017)

Make lots of small cli fixes that arose when people were using the tools for
the testnet.

IMPROVEMENTS:

* basecoin
    * `basecoin start` supports all flags that `tendermint node` does, such as
    `--rpc.laddr`, `--p2p.seeds`, and `--p2p.skip_upnp`
    * fully supports `--log_level` and `--trace` for logger configuration
    * merkleeyes no longers spams the logs... unless you want it
        * Example: `basecoin start --log_level="merkleeyes:info,state:info,*:error"`
        * Example: `basecoin start --log_level="merkleeyes:debug,state:info,*:error"`
* basecli
    * `basecli init` is more intelligent and only complains if there really was
    a connected chain, not just random files
    * support `localhost:46657` or `http://localhost:46657` format for nodes,
    not just `tcp://localhost:46657`
    * Add `--genesis` to init to specify chain-id and validator hash
        * Example: `basecli init --node=localhost:46657 --genesis=$HOME/.basecoin/genesis.json`
    * `basecli rpc` has a number of methods to easily accept tendermint rpc, and verifies what it can

BUG FIXES:

* basecli
    * `basecli query account` accepts hex account address with or without `0x`
    prefix
    * gives error message when running commands on an unitialized chain, rather
    than some unintelligable panic

## 0.6.0 (June 22, 2017)

Make the basecli command the only way to use client-side, to enforce best
security practices. Lots of enhancements to get it up to production quality.

BREAKING CHANGES:

* ./cmd/commands -> ./cmd/basecoin/commands
* basecli
    * `basecli proof state get` -> `basecli query key`
    * `basecli proof tx get` -> `basecli query tx`
    * `basecli proof state get --app=account` -> `basecli query account`
    * use `--chain-id` not `--chainid` for consistency
    * update to use `--trace` not `--debug` for stack traces on errors
    * complete overhaul on how tx and query subcommands are added. (see counter or trackomatron for examples)
    * no longer supports counter app (see new countercli)
* basecoin
    * `basecoin init` takes an argument, an address to allocate funds to in the genesis
    * removed key2.json
    * removed all client side functionality from it (use basecli now for proofs)
        * no tx subcommand
        * no query subcommand
        * no account (query) subcommand
        * a few other random ones...
    * enhanced relay subcommand
        * relay start did what relay used to do
        * relay init registers both chains on one another (to set it up so relay start just works)
* docs
    * removed `example-plugin`, put `counter` inside `docs/guide`
* app
    * Implements ABCI handshake by proxying merkleeyes.Info()

IMPROVEMENTS:

* `basecoin init` support `--chain-id`
* intergrates tendermint 0.10.0 (not the rc-2, but the real thing)
* commands return error code (1) on failure for easier script testing
* add `reset_all` to basecli, and never delete keys on `init`
* new shutil based unit tests, with better coverage of the cli actions
* just `make fresh` when things are getting stale ;)

BUG FIXES:

* app: no longer panics on missing app_options in genesis (thanks, anton)
* docs: updated all docs... again
* ibc: fix panic on getting BlockID from commit without 100% precommits (still a TODO)

## 0.5.2 (June 2, 2017)

BUG FIXES:

* fix parsing of the log level from Tendermint config (#97)

## 0.5.1 (May 30, 2017)

BUG FIXES:

* fix ibc demo app to use proper tendermint flags, 0.10.0-rc2 compatibility
* Make sure all cli uses new json.Marshal not wire.JSONBytes

## 0.5.0 (May 27, 2017)

BREAKING CHANGES:

* only those related to the tendermint 0.9 -> 0.10 upgrade

IMPROVEMENTS:

* basecoin cli
    * integrates tendermint 0.10.0 and unifies cli (init, unsafe_reset_all, ...)
    * integrate viper, all command line flags can also be defined in environmental variables or config.toml
* genesis file
    * you can define accounts with either address or pub_key
    * sorts coins for you, so no silent errors if not in alphabetical order
* [light-client](https://github.com/tendermint/light-client) integration
    * no longer must you trust the node you connect to, prove everything!
    * new [basecli command](./cmd/basecli/README.md)
    * integrated [key management](https://github.com/tendermint/go-crypto/blob/master/cmd/README.md), stored encrypted locally
    * tracks validator set changes and proves everything from one initial validator seed
    * `basecli proof state` gets complete proofs for any abci state
    * `basecli proof tx` gets complete proof where a tx was stored in the chain
    * `basecli proxy` exposes tendermint rpc, but only passes through results after doing complete verification

BUG FIXES:

* no more silently ignored error with invalid coin names (eg. "17.22foo coin" used to parse as "17 foo", not warning/error)

## 0.4.1 (April 26, 2017)

BUG FIXES:

* Fix bug in `basecoin unsafe_reset_X` where the `priv_validator.json` was not being reset

## 0.4.0 (April 21, 2017)

BREAKING CHANGES:

* CLI now uses Cobra, which forced changes to some of the flag names and orderings

IMPROVEMENTS:

* `basecoin init` doesn't generate error if already initialized
* Much more testing

## 0.3.1 (March 23, 2017)

IMPROVEMENTS:

* CLI returns exit code 1 and logs error before exiting

## 0.3.0 (March 23, 2017)

BREAKING CHANGES:

* Remove `--data` flag and use `BCHOME` to set the home directory (defaults to `~/.basecoin`)
* Remove `--in-proc` flag and start Tendermint in-process by default (expect Tendermint files in $BCHOME/tendermint).
  To start just the ABCI app/server, use `basecoin start --without-tendermint`.
* Consolidate genesis files so the Basecoin genesis is an object under `app_options` in Tendermint genesis. For instance:

```json
{
  "app_hash": "",
  "chain_id": "foo_bar_chain",
  "genesis_time": "0001-01-01T00:00:00.000Z",
  "validators": [
    {
      "amount": 10,
      "name": "",
      "pub_key": [
        1,
        "7B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
      ]
    }
  ],
  "app_options": {
    "accounts": [
      {
        "pub_key": {
          "type": "ed25519",
          "data": "6880db93598e283a67c4d88fc67a8858aa2de70f713fe94a5109e29c137100c2"
        },
        "coins": [
          {
            "denom": "blank",
            "amount": 12345
          },
          {
            "denom": "ETH",
            "amount": 654321
          }
        ]
      }
    ],
    "plugin_options": ["plugin1/key1", "value1", "plugin1/key2", "value2"]
  }
}
```

Note the array of key-value pairs is now under `app_options.plugin_options` while the `app_options` themselves are well formed.
We also changed `chainID` to `chain_id` and consolidated to have just one of them.

FEATURES:

* Introduce `basecoin init` and `basecoin unsafe_reset_all`

## 0.2.0 (March 6, 2017)

BREAKING CHANGES:

* Update to ABCI v0.4.0 and Tendermint v0.9.0
* Coins are specified on the CLI as `Xcoin`, eg. `5gold`
* `Cost` is now `Fee`

FEATURES:

* CLI for sending transactions and querying the state,
  designed to be easily extensible as plugins are implemented
* Run Basecoin in-process with Tendermint
* Add `/account` path in Query
* IBC plugin for InterBlockchain Communication
* Demo script of IBC between two chains

IMPROVEMENTS:

* Use new Tendermint `/commit` endpoint for crafting IBC transactions
* More unit tests
* Use go-crypto S structs and go-data for more standard JSON
* Demo uses fewer sleeps

BUG FIXES:

* Various little fixes in coin arithmetic
* More commit validation in IBC
* Return results from transactions

## PreHistory

### January 14-18, 2017

* Update to Tendermint v0.8.0
* Cleanup a bit and release blog post

### September 22, 2016

* Basecoin compiles again

<!-- Release links -->

[unreleased]: https://github.com/cosmos/cosmos-sdk/compare/v0.38.2...HEAD
[v0.38.2]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.38.2
[v0.38.1]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.38.1
[v0.38.0]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.38.0
[v0.37.9]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.9
[v0.37.8]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.8
[v0.37.7]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.7
[v0.37.6]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.6
[v0.37.5]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.5
[v0.37.4]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.4
[v0.37.3]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.3
[v0.37.1]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.1
[v0.37.0]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.0
[v0.36.0]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.36.0
