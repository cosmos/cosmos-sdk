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
appropriate stanza (see below). Each entry is required to include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The tag should consist of where the change is being made ex. (x/staking), (store)
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

### Bug Fixes

* (x/authz) [#24638](https://github.com/cosmos/cosmos-sdk/pull/24638) Fixed a minor bug where the grant key was cast as a string and dumped directly into the error message leading to an error string possibly containing invalid UTF-8.

### Deprecated

* (x/nft) [#24575](https://github.com/cosmos/cosmos-sdk/pull/24575) Deprecate the `x/nft` module in the Cosmos SDK repository.  This module will not be maintained to the extent that our core modules will and will be kept in a [legacy repo](https://github.com/cosmos/cosmos-legacy).
* (x/group) [#24571](https://github.com/cosmos/cosmos-sdk/pull/24571) Deprecate the `x/group` module in the Cosmos SDK repository.  This module will not be maintained to the extent that our core modules will and will be kept in a [legacy repo](https://github.com/cosmos/cosmos-legacy).

### Bug Fixes

* (client, client/rpc, x/auth/tx) [#24551](https://github.com/cosmos/cosmos-sdk/pull/24551) Handle cancellation properly when supplying context to client methods.

## [v0.53.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.53.0) - 2025-04-29

### Features

* (simsx) [#24062](https://github.com/cosmos/cosmos-sdk/pull/24062) [#24145](https://github.com/cosmos/cosmos-sdk/pull/24145) Add new simsx framework on top of simulations for better module dev experience.
* (baseapp) [#24069](https://github.com/cosmos/cosmos-sdk/pull/24069) Create CheckTxHandler to allow extending the logic of CheckTx.
* (types) [#24093](https://github.com/cosmos/cosmos-sdk/pull/24093) Added a new method, `IsGT`, for `types.Coin`. This method is used to check if a `types.Coin` is greater than another `types.Coin`.
* (client/keys) [#24071](https://github.com/cosmos/cosmos-sdk/pull/24071) Add support for importing hex key using standard input.
* (types) [#23780](https://github.com/cosmos/cosmos-sdk/pull/23780) Add a ValueCodec for the math.Uint type that can be used in collections maps.
* (perf)[#24045](https://github.com/cosmos/cosmos-sdk/pull/24045) Sims: Replace runsim command with Go stdlib testing. CLI: `Commit` default true, `Lean`, `SimulateEveryOperation`, `PrintAllInvariants`, `DBBackend` params removed
* (crypto/keyring) [#24040](https://github.com/cosmos/cosmos-sdk/pull/24040) Expose the db keyring used in the keystore.
* (types) [#23919](https://github.com/cosmos/cosmos-sdk/pull/23919) Add MustValAddressFromBech32 function.
* (all) [#23708](https://github.com/cosmos/cosmos-sdk/pull/23708) Add unordered transaction support.
  * Adds a `--timeout-timestamp` flag that allows users to specify a block time at which the unordered transactions should expire from the mempool.
* (x/epochs) [#23815](https://github.com/cosmos/cosmos-sdk/pull/23815) Upstream `x/epochs` from Osmosis
* (client) [#23811](https://github.com/cosmos/cosmos-sdk/pull/23811) Add auto cli for node service.
* (genutil) [#24018](https://github.com/cosmos/cosmos-sdk/pull/24018) Allow manually setting the consensus key type in genesis
* (client) [#18557](https://github.com/cosmos/cosmos-sdk/pull/18557) Add `--qrcode` flag to `keys show` command to support displaying keys address QR code.
* (x/auth) [#24030](https://github.com/cosmos/cosmos-sdk/pull/24030) Allow usage of ed25519 keys for transaction signing.
* (baseapp) [#24163](https://github.com/cosmos/cosmos-sdk/pull/24163) Add `StreamingManager` to baseapp to extend the abci listeners.
* (x/protocolpool) [#23933](https://github.com/cosmos/cosmos-sdk/pull/23933) Add x/protocolpool module.
  * x/distribution can now utilize an externally managed community pool. NOTE: this will make the message handlers for FundCommunityPool and CommunityPoolSpend error, as well as the query handler for CommunityPool.
* (client) [#18101](https://github.com/cosmos/cosmos-sdk/pull/18101) Add a `keyring-default-keyname` in `client.toml` for specifying a default key name, and skip the need to use the `--from` flag when signing transactions.
* (x/gov) [#24355](https://github.com/cosmos/cosmos-sdk/pull/24355) Allow users to set a custom CalculateVoteResultsAndVotingPower function to be used in govkeeper.Tally.
* (x/mint) [#24436](https://github.com/cosmos/cosmos-sdk/pull/24436) Allow users to set a custom minting function used in the `x/mint` begin blocker.
  * The `InflationCalculationFn` argument to `mint.NewAppModule()` is now ignored and must be nil.  To set a custom `InflationCalculationFn` on the default minter, use `mintkeeper.WithMintFn(mintkeeper.DefaultMintFn(customInflationFn))`.
* (api) [#24428](https://github.com/cosmos/cosmos-sdk/pull/24428) Add block height to response headers

### Improvements

* (x/feegrant) [24461](https://github.com/cosmos/cosmos-sdk/pull/24461) Use collections for `FeeAllowance`, `FeeAllowanceQueue`.
* (client) [#24561](https://github.com/cosmos/cosmos-sdk/pull/24561) TimeoutTimestamp flag has been changed to TimeoutDuration, which now sets the timeout timestamp of unordered transactions to the current time + duration passed.
* (telemetry) [#24541](https://github.com/cosmos/cosmos-sdk/pull/24541) Telemetry now includes a pre_blocker metric key. x/upgrade should migrate to this key in v0.54.0.
* (x/auth) [#24541](https://github.com/cosmos/cosmos-sdk/pull/24541) x/auth's PreBlocker now emits telemetry under the pre_blocker metric key.
* (x/bank) [#24431](https://github.com/cosmos/cosmos-sdk/pull/24431) Reduce the number of `ValidateDenom` calls in `bank.SendCoins` and `Coin`.
  * The `AmountOf()` method on`sdk.Coins` no longer will `panic` if given an invalid denom and will instead return a zero value.
* (x/staking) [#24391](https://github.com/cosmos/cosmos-sdk/pull/24391) Replace panics with error results; more verbose error messages
* (x/staking) [#24354](https://github.com/cosmos/cosmos-sdk/pull/24354) Optimize validator endblock by reducing bech32 conversions, resulting in significant performance improvement
* (client/keys) [#18950](https://github.com/cosmos/cosmos-sdk/pull/18950) Improve `<appd> keys add`, `<appd> keys import` and `<appd> keys rename` by checking name validation.
* (client/keys) [#18703](https://github.com/cosmos/cosmos-sdk/pull/18703) Improve `<appd> keys add` and `<appd> keys show` by checking whether there are duplicate keys in the multisig case.
* (client/keys) [#18745](https://github.com/cosmos/cosmos-sdk/pull/18745) Improve `<appd> keys export` and `<appd> keys mnemonic` by adding --yes option to skip interactive confirmation.
* (x/bank) [#24106](https://github.com/cosmos/cosmos-sdk/pull/24106) `SendCoins` now checks for `SendRestrictions` before instead of after deducting coins using `subUnlockedCoins`.
* (crypto/ledger) [#24036](https://github.com/cosmos/cosmos-sdk/pull/24036) Improve error message when deriving paths using index > 100
* (gRPC) [#23844](https://github.com/cosmos/cosmos-sdk/pull/23844) Add debug log prints for each gRPC request.
* (gRPC) [#24073](https://github.com/cosmos/cosmos-sdk/pull/24073) Adds error handling for out-of-gas panics in grpc query handlers.
* (server) [#24072](https://github.com/cosmos/cosmos-sdk/pull/24072) Return BlockHeader by shallow copy in server Context.
* (x/bank) [#24053](https://github.com/cosmos/cosmos-sdk/pull/24053) Resolve a foot-gun by swapping send restrictions check in `InputOutputCoins` before coin deduction.
* (codec/types) [#24336](https://github.com/cosmos/cosmos-sdk/pull/24336) Most types definitions were moved to `github.com/cosmos/gogoproto/types/any` with aliases to these left in `codec/types` so that there should be no breakage to existing code. This allows protobuf generated code to optionally reference the SDK's custom `Any` type without a direct dependency on the SDK. This can be done by changing the `protoc` `M` parameter for `any.proto` to `Mgoogle/protobuf/any.proto=github.com/cosmos/gogoproto/types/any`.

### Bug Fixes

* (x/gov)[#24460](https://github.com/cosmos/cosmos-sdk/pull/24460) Do not call Remove during Walk in defaultCalculateVoteResultsAndVotingPower.
* (baseapp) [24261](https://github.com/cosmos/cosmos-sdk/pull/24261) Fix post handler error always results in code 1
* (server) [#24068](https://github.com/cosmos/cosmos-sdk/pull/24068) Allow align block header with skip check header in grpc server.
* (x/gov) [#24044](https://github.com/cosmos/cosmos-sdk/pull/24044) Fix some places in which we call Remove inside a Walk (x/gov).
* (baseapp) [#24042](https://github.com/cosmos/cosmos-sdk/pull/24042) Fixed a data race inside BaseApp.getContext, found by end-to-end (e2e) tests.
* (client/server) [#24059](https://github.com/cosmos/cosmos-sdk/pull/24059) Consistently set viper prefix in client and server. It defaults for the binary name for both client and server.
* (client/keys) [#24041](https://github.com/cosmos/cosmos-sdk/pull/24041) `keys delete` won't terminate when a key is not found, but will log the error.
* (baseapp) [#24027](https://github.com/cosmos/cosmos-sdk/pull/24027) Ensure that `BaseApp.Init` checks that the commit multistore is set to protect against nil dereferences.
* (x/group) [GHSA-47ww-ff84-4jrg](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-47ww-ff84-4jrg) Fix x/group can halt when erroring in EndBlocker
* (x/distribution) [#23934](https://github.com/cosmos/cosmos-sdk/pull/23934) Fix vulnerability in `incrementReferenceCount` in distribution.
* (baseapp) [#23879](https://github.com/cosmos/cosmos-sdk/pull/23879) Ensure finalize block response is not empty in the defer check of FinalizeBlock to avoid panic by nil pointer.
* (query) [#23883](https://github.com/cosmos/cosmos-sdk/pull/23883) Fix NPE in query pagination.
* (client) [#23860](https://github.com/cosmos/cosmos-sdk/pull/23860) Add missing `unordered` field for legacy amino signing of tx body.
* (x/bank) [#23836](https://github.com/cosmos/cosmos-sdk/pull/23836) Fix `DenomMetadata` rpc allow value with slashes.
* (query) [87d3a43](https://github.com/cosmos/cosmos-sdk/commit/87d3a432af95f4cf96aa02351ed5fcc51cca6e7b) Fix collection filtered pagination.
* (sims) [#23952](https://github.com/cosmos/cosmos-sdk/pull/23952) Use liveness matrix for validator sign status in sims
* (baseapp) [#24055](https://github.com/cosmos/cosmos-sdk/pull/24055) Align block header when query with latest height.
* (baseapp) [#24074](https://github.com/cosmos/cosmos-sdk/pull/24074) Use CometBFT's ComputeProtoSizeForTxs in defaultTxSelector.SelectTxForProposal for consistency.
* (cli) [#24090](https://github.com/cosmos/cosmos-sdk/pull/24090) Prune cmd should disable async pruning.
* (x/auth) [#19239](https://github.com/cosmos/cosmos-sdk/pull/19239) Sets from flag in multi-sign command to avoid no key name provided error.
* (x/auth) [#23741](https://github.com/cosmos/cosmos-sdk/pull/23741) Support legacy global AccountNumber for legacy compatibility.
* (baseapp) [#24526](https://github.com/cosmos/cosmos-sdk/pull/24526) Fix incorrect retention height when `commitHeight` equals `minRetainBlocks`.
* (x/protocolpool) [#24594](https://github.com/cosmos/cosmos-sdk/pull/24594) Fix NPE when initializing module via depinject.
* (x/epochs) [#24610](https://github.com/cosmos/cosmos-sdk/pull/24610) Fix semantics of `CurrentEpochStartHeight` being set before epoch has started.

## [v0.50.13](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.13) - 2025-03-12

### Bug Fixes

* [GHSA-47ww-ff84-4jrg](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-47ww-ff84-4jrg) Fix x/group can halt when erroring in EndBlocker

## [v0.50.12](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.12) - 2025-02-20

### Bug Fixes

* [GHSA-x5vx-95h7-rv4p](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-x5vx-95h7-rv4p) Fix Group module can halt chain when handling a malicious proposal.

## [v0.50.11](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.11) - 2024-12-16

### Features

* (crypto/keyring) [#21653](https://github.com/cosmos/cosmos-sdk/pull/21653) New Linux-only backend that adds Linux kernel's `keyctl` support.

### Improvements

* (server) [#21941](https://github.com/cosmos/cosmos-sdk/pull/21941) Regenerate addrbook.json for in place testnet.

### Bug Fixes

* Fix [ABS-0043/ABS-0044](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-8wcc-m6j2-qxvm) Limit recursion depth for unknown field detection and unpack any
* (server) [#22564](https://github.com/cosmos/cosmos-sdk/pull/22564) Fix fallback genesis path in server
* (x/group) [#22425](https://github.com/cosmos/cosmos-sdk/pull/22425) Proper address rendering in error
* (sims) [#21906](https://github.com/cosmos/cosmos-sdk/pull/21906) Skip sims test when running dry on validators
* (cli) [#21919](https://github.com/cosmos/cosmos-sdk/pull/21919) Query address-by-acc-num by account_id instead of id.
* (x/group) [#22229](https://github.com/cosmos/cosmos-sdk/pull/22229) Accept `1` and `try` in CLI for group proposal exec.

## [v0.50.10](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.10) - 2024-09-20

### Features

* (cli) [#20779](https://github.com/cosmos/cosmos-sdk/pull/20779) Added `module-hash-by-height` command to query and retrieve module hashes at a specified blockchain height, enhancing debugging capabilities.
* (cli) [#21372](https://github.com/cosmos/cosmos-sdk/pull/21372) Added a `bulk-add-genesis-account` genesis command to add many genesis accounts at once.
* (types/collections) [#21724](https://github.com/cosmos/cosmos-sdk/pull/21724) Added `LegacyDec` collection value.

### Improvements

* (x/bank) [#21460](https://github.com/cosmos/cosmos-sdk/pull/21460) Added `Sender` attribute in `MsgMultiSend` event.
* (genutil) [#21701](https://github.com/cosmos/cosmos-sdk/pull/21701) Improved error messages for genesis validation.
* (testutil/integration) [#21816](https://github.com/cosmos/cosmos-sdk/pull/21816) Allow to pass baseapp options in `NewIntegrationApp`.

### Bug Fixes

* (runtime) [#21769](https://github.com/cosmos/cosmos-sdk/pull/21769) Fix baseapp options ordering to avoid overwriting options set by modules.
* (x/consensus) [#21493](https://github.com/cosmos/cosmos-sdk/pull/21493) Fix regression that prevented to upgrade to > v0.50.7 without consensus version params.
* (baseapp) [#21256](https://github.com/cosmos/cosmos-sdk/pull/21256) Halt height will not commit the block indicated, meaning that if halt-height is set to 10, only blocks until 9 (included) will be committed. This is to go back to the original behavior before a change was introduced in v0.50.0.
* (baseapp) [#21444](https://github.com/cosmos/cosmos-sdk/pull/21444) Follow-up, Return PreBlocker events in FinalizeBlockResponse.
* (baseapp) [#21413](https://github.com/cosmos/cosmos-sdk/pull/21413) Fix data race in sdk mempool.

## [v0.50.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.9) - 2024-08-07

## Bug Fixes

* (baseapp) [#21159](https://github.com/cosmos/cosmos-sdk/pull/21159) Return PreBlocker events in FinalizeBlockResponse.
* [#20939](https://github.com/cosmos/cosmos-sdk/pull/20939) Fix collection reverse iterator to include `pagination.key` in the result.
* (client/grpc) [#20969](https://github.com/cosmos/cosmos-sdk/pull/20969) Fix `node.NewQueryServer` method not setting `cfg`.
* (testutil/integration) [#21006](https://github.com/cosmos/cosmos-sdk/pull/21006) Fix `NewIntegrationApp` method not writing default genesis to state.
* (runtime) [#21080](https://github.com/cosmos/cosmos-sdk/pull/21080) Fix `app.yaml` / `app.json` incompatibility with `depinject v1.0.0`.

## [v0.50.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.8) - 2024-07-15

## Features

* (client) [#20690](https://github.com/cosmos/cosmos-sdk/pull/20690) Import mnemonic from file

## Improvements

* (x/authz,x/feegrant) [#20590](https://github.com/cosmos/cosmos-sdk/pull/20590) Provide updated keeper in depinject for authz and feegrant modules.
* [#20631](https://github.com/cosmos/cosmos-sdk/pull/20631) Fix json parsing in the wait-tx command.
* (x/auth) [#20438](https://github.com/cosmos/cosmos-sdk/pull/20438) Add `--skip-signature-verification` flag to multisign command to allow nested multisigs.

## Bug Fixes

* (simulation) [#17911](https://github.com/cosmos/cosmos-sdk/pull/17911) Fix all problems with executing command `make test-sim-custom-genesis-fast` for simulation test.
* (simulation) [#18196](https://github.com/cosmos/cosmos-sdk/pull/18196) Fix the problem of `validator set is empty after InitGenesis` in simulation test.

## [v0.50.7](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.7) - 2024-06-04

### Improvements

* (debug) [#20328](https://github.com/cosmos/cosmos-sdk/pull/20328) Add consensus address for debug cmd.
* (runtime) [#20264](https://github.com/cosmos/cosmos-sdk/pull/20264) Expose grpc query router via depinject.
* (x/consensus) [#20381](https://github.com/cosmos/cosmos-sdk/pull/20381) Use Comet utility for consensus module consensus param updates.
* (client) [#20356](https://github.com/cosmos/cosmos-sdk/pull/20356) Overwrite client context when available in `SetCmdClientContext`.

### Bug Fixes

* (baseapp) [#20346](https://github.com/cosmos/cosmos-sdk/pull/20346) Correctly assign `execModeSimulate` to context for `simulateTx`.
* (baseapp) [#20144](https://github.com/cosmos/cosmos-sdk/pull/20144) Remove txs from mempool when AnteHandler fails in recheck.
* (baseapp) [#20107](https://github.com/cosmos/cosmos-sdk/pull/20107) Avoid header height overwrite block height.
* (cli) [#20020](https://github.com/cosmos/cosmos-sdk/pull/20020) Make bootstrap-state command support both new and legacy genesis format.
* (testutil/sims) [#20151](https://github.com/cosmos/cosmos-sdk/pull/20151) Set all signatures and don't overwrite the previous one in `GenSignedMockTx`.

## [v0.50.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.6) - 2024-04-22

### Features

* (types) [#19759](https://github.com/cosmos/cosmos-sdk/pull/19759) Align SignerExtractionAdapter in PriorityNonceMempool Remove.
* (client) [#19870](https://github.com/cosmos/cosmos-sdk/pull/19870) Add new query command `wait-tx`. Alias `event-query-tx-for` to `wait-tx` for backward compatibility.

### Improvements

* (telemetry) [#19903](https://github.com/cosmos/cosmos-sdk/pull/19903) Conditionally emit metrics based on enablement.
    * **Introduction of `Now` Function**: Added a new function called `Now` to the telemetry package. It returns the current system time if telemetry is enabled, or a zero time if telemetry is not enabled.
    * **Atomic Global Variable**: Implemented an atomic global variable to manage the state of telemetry's enablement. This ensures thread safety for the telemetry state.
    * **Conditional Telemetry Emission**: All telemetry functions have been updated to emit metrics only when telemetry is enabled. They perform a check with `isTelemetryEnabled()` and return early if telemetry is disabled, minimizing unnecessary operations and overhead.
* (deps) [#19810](https://github.com/cosmos/cosmos-sdk/pull/19810) Upgrade prometheus version and fix API breaking change due to prometheus bump.
* (deps) [#19810](https://github.com/cosmos/cosmos-sdk/pull/19810) Bump `cosmossdk.io/store` to v1.1.0.
* (server) [#19884](https://github.com/cosmos/cosmos-sdk/pull/19884) Add start customizability to start command options.
* (x/gov) [#19853](https://github.com/cosmos/cosmos-sdk/pull/19853) Emit `depositor` in `EventTypeProposalDeposit`.
* (x/gov) [#19844](https://github.com/cosmos/cosmos-sdk/pull/19844) Emit the proposer of governance proposals.
* (baseapp) [#19616](https://github.com/cosmos/cosmos-sdk/pull/19616) Don't share gas meter in tx execution.

## Bug Fixes

* (x/authz) [#20114](https://github.com/cosmos/cosmos-sdk/pull/20114) Follow up of [GHSA-4j93-fm92-rp4m](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-4j93-fm92-rp4m) for `x/authz`.
* (crypto) [#19691](https://github.com/cosmos/cosmos-sdk/pull/19745) Fix tx sign doesn't throw an error when incorrect Ledger is used.
* (baseapp) [#19970](https://github.com/cosmos/cosmos-sdk/pull/19970) Fix default config values to use no-op mempool as default.
* (crypto) [#20027](https://github.com/cosmos/cosmos-sdk/pull/20027) secp256r1 keys now implement gogoproto's customtype interface.
* (x/bank) [#20028](https://github.com/cosmos/cosmos-sdk/pull/20028) Align query with multi denoms for send-enabled.

## [v0.50.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.5) - 2024-03-12

### Features

* (baseapp) [#19626](https://github.com/cosmos/cosmos-sdk/pull/19626) Add `DisableBlockGasMeter` option to `BaseApp`, which removes the block gas meter during transaction execution.

### Improvements

* (x/distribution) [#19707](https://github.com/cosmos/cosmos-sdk/pull/19707) Add autocli config for `DelegationTotalRewards` for CLI consistency with `q rewards` commands in previous versions.
* (x/auth) [#19651](https://github.com/cosmos/cosmos-sdk/pull/19651) Allow empty public keys in `GetSignBytesAdapter`.

### Bug Fixes

* (x/gov) [#19725](https://github.com/cosmos/cosmos-sdk/pull/19725) Fetch a failed proposal tally from proposal.FinalTallyResult in the gprc query.
* (types) [#19709](https://github.com/cosmos/cosmos-sdk/pull/19709) Fix skip staking genesis export when using `CoreAppModuleAdaptor` / `CoreAppModuleBasicAdaptor` for it.
* (x/auth) [#19549](https://github.com/cosmos/cosmos-sdk/pull/19549) Accept custom get signers when injecting `x/auth/tx`.
* (x/staking) Fix a possible bypass of delegator slashing: [GHSA-86h5-xcpx-cfqc](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-86h5-xcpx-cfqc)
* (baseapp) Fix a bug in `baseapp.ValidateVoteExtensions` helper ([GHSA-95rx-m9m5-m94v](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-95rx-m9m5-m94v)). The helper has been fixed and for avoiding API breaking changes `currentHeight` and `chainID` arguments are ignored. Those arguments are removed from the helper in v0.51+.

## [v0.50.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.4) - 2024-02-19

### Features

* (server) [#19280](https://github.com/cosmos/cosmos-sdk/pull/19280) Adds in-place testnet CLI command.

### Improvements

* (client) [#19393](https://github.com/cosmos/cosmos-sdk/pull/19393/) Add `ReadDefaultValuesFromDefaultClientConfig` to populate the default values from the default client config in client.Context without creating a app folder.

### Bug Fixes

* (x/auth/vesting) [GHSA-4j93-fm92-rp4m](#bug-fixes) Add `BlockedAddr` check in `CreatePeriodicVestingAccount`.
* (baseapp) [#19338](https://github.com/cosmos/cosmos-sdk/pull/19338) Set HeaderInfo in context when calling `setState`.
* (baseapp): [#19200](https://github.com/cosmos/cosmos-sdk/pull/19200) Ensure that sdk side ve math matches cometbft.
* [#19106](https://github.com/cosmos/cosmos-sdk/pull/19106) Allow empty public keys when setting signatures. Public keys aren't needed for every transaction.
* (baseapp) [#19198](https://github.com/cosmos/cosmos-sdk/pull/19198) Remove usage of pointers in logs in all optimistic execution goroutines.
* (baseapp) [#19177](https://github.com/cosmos/cosmos-sdk/pull/19177) Fix baseapp `DefaultProposalHandler` same-sender non-sequential sequence.
* (crypto) [#19371](https://github.com/cosmos/cosmos-sdk/pull/19371) Avoid CLI redundant log in stdout, log to stderr instead.

## [v0.50.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.3) - 2024-01-15

### Features

* (types) [#18991](https://github.com/cosmos/cosmos-sdk/pull/18991) Add SignerExtractionAdapter to PriorityNonceMempool/Config and provide Default implementation matching existing behavior.
* (gRPC) [#19043](https://github.com/cosmos/cosmos-sdk/pull/19043) Add `halt_height` to the gRPC `/cosmos/base/node/v1beta1/config` request.

### Improvements

* (x/bank) [#18956](https://github.com/cosmos/cosmos-sdk/pull/18956) Introduced a new `DenomOwnersByQuery` query method for `DenomOwners`, which accepts the denom value as a query string parameter, resolving issues with denoms containing slashes.
* (x/gov) [#18707](https://github.com/cosmos/cosmos-sdk/pull/18707) Improve genesis validation.
* (x/auth/tx) [#18772](https://github.com/cosmos/cosmos-sdk/pull/18772) Remove misleading gas wanted from tx simulation failure log.
* (client/tx) [#18852](https://github.com/cosmos/cosmos-sdk/pull/18852) Add `WithFromName` to tx factory.
* (types) [#18888](https://github.com/cosmos/cosmos-sdk/pull/18888) Speedup DecCoin.Sort() if len(coins) <= 1
* (types) [#18875](https://github.com/cosmos/cosmos-sdk/pull/18875) Speedup coins.Sort() if len(coins) <= 1
* (baseapp) [#18915](https://github.com/cosmos/cosmos-sdk/pull/18915) Add a new `ExecModeVerifyVoteExtension` exec mode and ensure it's populated in the `Context` during `VerifyVoteExtension` execution.
* (testutil) [#18930](https://github.com/cosmos/cosmos-sdk/pull/18930) Add NodeURI for clientCtx.

### Bug Fixes

* (baseapp) [#19058](https://github.com/cosmos/cosmos-sdk/pull/19058) Fix baseapp posthandler branch would fail if the `runMsgs` had returned an error.
* (baseapp) [#18609](https://github.com/cosmos/cosmos-sdk/issues/18609) Fixed accounting in the block gas meter after module's beginBlock and before DeliverTx, ensuring transaction processing always starts with the expected zeroed out block gas meter.
* (baseapp) [#18895](https://github.com/cosmos/cosmos-sdk/pull/18895) Fix de-duplicating vote extensions during validation in ValidateVoteExtensions.

## [v0.50.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.2) - 2023-12-11

### Features

* (debug) [#18219](https://github.com/cosmos/cosmos-sdk/pull/18219) Add debug commands for application codec types.
* (client/keys) [#17639](https://github.com/cosmos/cosmos-sdk/pull/17639) Allows using and saving public keys encoded as base64.
* (server) [#17094](https://github.com/cosmos/cosmos-sdk/pull/17094) Add a `shutdown-grace` flag for waiting a given time before exit.

### Improvements

* (telemetry) [#18646] (https://github.com/cosmos/cosmos-sdk/pull/18646) Enable statsd and dogstatsd telemetry sinks.
* (server) [#18478](https://github.com/cosmos/cosmos-sdk/pull/18478) Add command flag to disable colored logs.
* (x/gov) [#18025](https://github.com/cosmos/cosmos-sdk/pull/18025) Improve `<appd> q gov proposer` by querying directly a proposal instead of tx events. It is an alias of `q gov proposal` as the proposer is a field of the proposal.
* (version) [#18063](https://github.com/cosmos/cosmos-sdk/pull/18063) Allow to define extra info to be displayed in `<appd> version --long` command.
* (codec/unknownproto)[#18541](https://github.com/cosmos/cosmos-sdk/pull/18541) Remove the use of "protoc-gen-gogo/descriptor" in favour of using the official protobuf descriptorpb types inside unknownproto.

### Bug Fixes

* (x/auth) [#18564](https://github.com/cosmos/cosmos-sdk/pull/18564) Fix total fees calculation when batch signing.
* (server) [#18537](https://github.com/cosmos/cosmos-sdk/pull/18537) Fix panic when defining minimum gas config as `100stake;100uatom`. Use a `,` delimiter instead of `;`. Fixes the server config getter to use the correct delimiter.
* [#18531](https://github.com/cosmos/cosmos-sdk/pull/18531) Baseapp's `GetConsensusParams` returns an empty struct instead of panicking if no params are found.
* (client/tx) [#18472](https://github.com/cosmos/cosmos-sdk/pull/18472) Utilizes the correct Pubkey when simulating a transaction.
* (baseapp) [#18486](https://github.com/cosmos/cosmos-sdk/pull/18486) Fixed FinalizeBlock calls not being passed to ABCIListeners.
* (baseapp) [#18627](https://github.com/cosmos/cosmos-sdk/pull/18627) Post handlers are run on non successful transaction executions too.
* (baseapp) [#18654](https://github.com/cosmos/cosmos-sdk/pull/18654) Fixes an issue in which `gogoproto.Merge` does not work with gogoproto messages with custom types.

## [v0.50.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.1) - 2023-11-07

> v0.50.0 has been retracted due to a mistake in tagging the release. Please use v0.50.1 instead.

### Features

* (baseapp) [#18071](https://github.com/cosmos/cosmos-sdk/pull/18071) Add hybrid handlers to `MsgServiceRouter`.
* (server) [#18162](https://github.com/cosmos/cosmos-sdk/pull/18162) Start gRPC & API server in standalone mode.
* (baseapp & types) [#17712](https://github.com/cosmos/cosmos-sdk/pull/17712) Introduce `PreBlock`, which runs before begin blocker other modules, and allows to modify consensus parameters, and the changes are visible to the following state machine logics. Additionally it can be used for vote extensions.
* (genutil) [#17571](https://github.com/cosmos/cosmos-sdk/pull/17571) Allow creation of `AppGenesis` without a file lookup.
* (codec) [#17042](https://github.com/cosmos/cosmos-sdk/pull/17042) Add `CollValueV2` which supports encoding of protov2 messages in collections.
* (x/gov) [#16976](https://github.com/cosmos/cosmos-sdk/pull/16976) Add `failed_reason` field to `Proposal` under `x/gov` to indicate the reason for a failed proposal. Referenced from [#238](https://github.com/bnb-chain/greenfield-cosmos-sdk/pull/238) under `bnb-chain/greenfield-cosmos-sdk`.
* (baseapp) [#16898](https://github.com/cosmos/cosmos-sdk/pull/16898) Add `preFinalizeBlockHook` to allow vote extensions persistence.
* (cli) [#16887](https://github.com/cosmos/cosmos-sdk/pull/16887) Add two new CLI commands: `<appd> tx simulate` for simulating a transaction; `<appd> query block-results` for querying CometBFT RPC for block results.
* (x/bank) [#16852](https://github.com/cosmos/cosmos-sdk/pull/16852) Add `DenomMetadataByQueryString` query in bank module to support metadata query by query string.
* (baseapp) [#16581](https://github.com/cosmos/cosmos-sdk/pull/16581) Implement Optimistic Execution as an experimental feature (not enabled by default).
* (types) [#16257](https://github.com/cosmos/cosmos-sdk/pull/16257) Allow setting the base denom in the denom registry.
* (baseapp) [#16239](https://github.com/cosmos/cosmos-sdk/pull/16239) Add Gas Limits to allow node operators to resource bound queries.
* (cli) [#16209](https://github.com/cosmos/cosmos-sdk/pull/16209) Make `StartCmd` more customizable.
* (types/simulation) [#16074](https://github.com/cosmos/cosmos-sdk/pull/16074) Add generic SimulationStoreDecoder for modules using collections.
* (genutil) [#16046](https://github.com/cosmos/cosmos-sdk/pull/16046) Add "module-name" flag to genutil `add-genesis-account` to enable intializing module accounts at genesis.* [#15970](https://github.com/cosmos/cosmos-sdk/pull/15970) Enable SIGN_MODE_TEXTUAL.
* (types) [#15958](https://github.com/cosmos/cosmos-sdk/pull/15958) Add `module.NewBasicManagerFromManager` for creating a basic module manager from a module manager.
* (types/module) [#15829](https://github.com/cosmos/cosmos-sdk/pull/15829) Add new endblocker interface to handle valset updates.
* (runtime) [#15818](https://github.com/cosmos/cosmos-sdk/pull/15818) Provide logger through `depinject` instead of appBuilder.
* (types) [#15735](https://github.com/cosmos/cosmos-sdk/pull/15735) Make `ValidateBasic() error` method of `Msg` interface optional. Modules should validate messages directly in their message handlers ([RFC 001](https://docs.cosmos.network/main/rfc/rfc-001-tx-validation)).
* (x/genutil) [#15679](https://github.com/cosmos/cosmos-sdk/pull/15679) Allow applications to specify a custom genesis migration function for the `genesis migrate` command.
* (telemetry) [#15657](https://github.com/cosmos/cosmos-sdk/pull/15657) Emit more data (go version, sdk version, upgrade height) in prom metrics.
* (client) [#15597](https://github.com/cosmos/cosmos-sdk/pull/15597) Add status endpoint for clients.
* (testutil/integration) [#15556](https://github.com/cosmos/cosmos-sdk/pull/15556) Introduce `testutil/integration` package for module integration testing.
* (runtime) [#15547](https://github.com/cosmos/cosmos-sdk/pull/15547) Allow runtime to pass event core api service to modules.
* (client) [#15458](https://github.com/cosmos/cosmos-sdk/pull/15458) Add a `CmdContext` field to client.Context initialized to cobra command's context.
* (x/genutil) [#15301](https://github.com/cosmos/cosmos-sdk/pull/15031) Add application genesis. The genesis is now entirely managed by the application and passed to CometBFT at note instantiation. Functions that were taking a `cmttypes.GenesisDoc{}` now takes a `genutiltypes.AppGenesis{}`.
* (core) [#15133](https://github.com/cosmos/cosmos-sdk/pull/15133) Implement RegisterServices in the module manager.
* (x/bank) [#14894](https://github.com/cosmos/cosmos-sdk/pull/14894) Return a human readable denomination for IBC vouchers when querying bank balances. Added a `ResolveDenom` parameter to `types.QueryAllBalancesRequest` and `--resolve-denom` flag to `GetBalancesCmd()`.
* (core) [#14860](https://github.com/cosmos/cosmos-sdk/pull/14860) Add `Precommit` and `PrepareCheckState` AppModule callbacks.
* (x/gov) [#14720](https://github.com/cosmos/cosmos-sdk/pull/14720) Upstream expedited proposals from Osmosis.
* (cli) [#14659](https://github.com/cosmos/cosmos-sdk/pull/14659) Added ability to query blocks by events with queries directly passed to Tendermint, which will allow for full query operator support, e.g. `>`.
* (x/auth) [#14650](https://github.com/cosmos/cosmos-sdk/pull/14650) Add Textual SignModeHandler. Enable `SIGN_MODE_TEXTUAL` by following the [UPGRADING.md](./UPGRADING.md) instructions.
* (x/crisis) [#14588](https://github.com/cosmos/cosmos-sdk/pull/14588) Use CacheContext() in AssertInvariants().
* (mempool) [#14484](https://github.com/cosmos/cosmos-sdk/pull/14484) Add priority nonce mempool option for transaction replacement.
* (query) [#14468](https://github.com/cosmos/cosmos-sdk/pull/14468) Implement pagination for collections.
* (x/gov) [#14373](https://github.com/cosmos/cosmos-sdk/pull/14373) Add new proto field `constitution` of type `string` to gov module genesis state, which allows chain builders to lay a strong foundation by specifying purpose.
* (client) [#14342](https://github.com/cosmos/cosmos-sdk/pull/14342) Add `<app> config` command is now a sub-command, for setting, getting and migrating Cosmos SDK configuration files.
* (x/distribution) [#14322](https://github.com/cosmos/cosmos-sdk/pull/14322) Introduce a new gRPC message handler, `DepositValidatorRewardsPool`, that allows explicit funding of a validator's reward pool.
* (x/bank) [#14224](https://github.com/cosmos/cosmos-sdk/pull/14224) Allow injection of restrictions on transfers using `AppendSendRestriction` or `PrependSendRestriction`.

### Improvements

* (x/gov) [#18189](https://github.com/cosmos/cosmos-sdk/pull/18189) Limit the accepted deposit coins for a proposal to the minimum proposal deposit denoms.
* (x/staking) [#18049](https://github.com/cosmos/cosmos-sdk/pull/18049) Return early if Slash encounters zero tokens to burn.
* (x/staking) [#18035](https://github.com/cosmos/cosmos-sdk/pull/18035) Hoisted out of the redelegation loop, the non-changing validator and delegator addresses parsing.
* (keyring) [#17913](https://github.com/cosmos/cosmos-sdk/pull/17913) Add `NewAutoCLIKeyring` for creating an AutoCLI keyring from a SDK keyring.
* (x/consensus) [#18041](https://github.com/cosmos/cosmos-sdk/pull/18041) Let `ToProtoConsensusParams()` return an error.
* (x/gov) [#17780](https://github.com/cosmos/cosmos-sdk/pull/17780) Recover panics and turn them into errors when executing x/gov proposals.
* (baseapp) [#17667](https://github.com/cosmos/cosmos-sdk/pull/17667) Close databases opened by SDK in `baseApp.Close()`.
* (types/module) [#17554](https://github.com/cosmos/cosmos-sdk/pull/17554) Introduce `HasABCIGenesis` which is implemented by a module only when a validatorset update needs to be returned.
* (cli) [#17389](https://github.com/cosmos/cosmos-sdk/pull/17389) gRPC CometBFT commands have been added under `<aapd> q consensus comet`. CometBFT commands placement in the SDK has been simplified. See the exhaustive list below.
    * `client/rpc.StatusCommand()` is now at `server.StatusCommand()`
* (testutil) [#17216](https://github.com/cosmos/cosmos-sdk/issues/17216) Add `DefaultContextWithKeys` to `testutil` package.
* (cli) [#17187](https://github.com/cosmos/cosmos-sdk/pull/17187) Do not use `ctx.PrintObjectLegacy` in commands anymore.
    * `<appd> q gov proposer [proposal-id]` now returns a proposal id as int instead of string.
* (x/staking) [#17164](https://github.com/cosmos/cosmos-sdk/pull/17164) Add `BondedTokensAndPubKeyByConsAddr` to the keeper to enable vote extension verification.
* (x/group, x/gov) [#17109](https://github.com/cosmos/cosmos-sdk/pull/17109) Let proposal summary be 40x longer than metadata limit.
* (version) [#17096](https://github.com/cosmos/cosmos-sdk/pull/17096) Improve `getSDKVersion()` to handle module replacements.
* (types) [#16890](https://github.com/cosmos/cosmos-sdk/pull/16890) Remove `GetTxCmd() *cobra.Command` and `GetQueryCmd() *cobra.Command` from `module.AppModuleBasic` interface.
* (x/authz) [#16869](https://github.com/cosmos/cosmos-sdk/pull/16869) Improve error message when grant not found.
* (all) [#16497](https://github.com/cosmos/cosmos-sdk/pull/16497) Removed all exported vestiges of `sdk.MustSortJSON` and `sdk.SortJSON`.
* (server) [#16238](https://github.com/cosmos/cosmos-sdk/pull/16238) Don't setup p2p node keys if starting a node in GRPC only mode.
* (cli) [#16206](https://github.com/cosmos/cosmos-sdk/pull/16206) Make ABCI handshake profileable.
* (types) [#16076](https://github.com/cosmos/cosmos-sdk/pull/16076) Optimize `ChainAnteDecorators`/`ChainPostDecorators` to instantiate the functions once instead of on every invocation of the returned `AnteHandler`/`PostHandler`.
* (server) [#16071](https://github.com/cosmos/cosmos-sdk/pull/16071) When `mempool.max-txs` is set to a negative value, use a no-op mempool (effectively disable the app mempool).
* (types/query) [#16041](https://github.com/cosmos/cosmos-sdk/pull/16041) Change pagination max limit to a variable in order to be modifed by application devs.
* (simapp) [#15958](https://github.com/cosmos/cosmos-sdk/pull/15958) Refactor SimApp for removing the global basic manager.
* (all modules) [#15901](https://github.com/cosmos/cosmos-sdk/issues/15901) All core Cosmos SDK modules query commands have migrated to [AutoCLI](https://docs.cosmos.network/main/core/autocli), ensuring parity between gRPC and CLI queries.
* (x/auth) [#15867](https://github.com/cosmos/cosmos-sdk/pull/15867) Support better logging for signature verification failure.
* (store/cachekv) [#15767](https://github.com/cosmos/cosmos-sdk/pull/15767) Reduce peak RAM usage during and after `InitGenesis`.
* (x/bank) [#15764](https://github.com/cosmos/cosmos-sdk/pull/15764) Speedup x/bank `InitGenesis`.
* (x/slashing) [#15580](https://github.com/cosmos/cosmos-sdk/pull/15580) Refactor the validator's missed block signing window to be a chunked bitmap instead of a "logical" bitmap, significantly reducing the storage footprint.
* (x/gov) [#15554](https://github.com/cosmos/cosmos-sdk/pull/15554) Add proposal result log in `active_proposal` event. When a proposal passes but fails to execute, the proposal result is logged in the `active_proposal` event.
* (x/consensus) [#15553](https://github.com/cosmos/cosmos-sdk/pull/15553) Migrate consensus module to use collections.
* (server) [#15358](https://github.com/cosmos/cosmos-sdk/pull/15358) Add `server.InterceptConfigsAndCreateContext` as alternative to `server.InterceptConfigsPreRunHandler` which does not set the server context and the default SDK logger.
* (mempool) [#15328](https://github.com/cosmos/cosmos-sdk/pull/15328) Improve the `PriorityNonceMempool`:
    * Support generic transaction prioritization, instead of `ctx.Priority()`
    * Improve construction through the use of a single `PriorityNonceMempoolConfig` instead of option functions
* (x/authz) [#15164](https://github.com/cosmos/cosmos-sdk/pull/15164) Add `MsgCancelUnbondingDelegation` to staking authorization.
* (server) [#15041](https://github.com/cosmos/cosmos-sdk/pull/15041) Remove unnecessary sleeps from gRPC and API server initiation. The servers will start and accept requests as soon as they're ready.
* (baseapp) [#15023](https://github.com/cosmos/cosmos-sdk/pull/15023) & [#15213](https://github.com/cosmos/cosmos-sdk/pull/15213) Add `MessageRouter` interface to baseapp and pass it to authz, gov and groups instead of concrete type.
* [#15011](https://github.com/cosmos/cosmos-sdk/pull/15011) Introduce `cosmossdk.io/log` package to provide a consistent logging interface through the SDK. CometBFT logger is now replaced by `cosmossdk.io/log.Logger`.
* (x/staking) [#14864](https://github.com/cosmos/cosmos-sdk/pull/14864) `<appd> tx staking create-validator` CLI command now takes a json file as an arg instead of using required flags.
* (x/auth) [#14758](https://github.com/cosmos/cosmos-sdk/pull/14758) Allow transaction event queries to directly passed to Tendermint, which will allow for full query operator support, e.g. `>`.
* (x/evidence) [#14757](https://github.com/cosmos/cosmos-sdk/pull/14757) Evidence messages do not need to implement a `.Type()` anymore.
* (x/auth/tx) [#14751](https://github.com/cosmos/cosmos-sdk/pull/14751) Remove `.Type()` and `Route()` methods from all msgs and `legacytx.LegacyMsg` interface.
* (cli) [#14659](https://github.com/cosmos/cosmos-sdk/pull/14659) Added ability to query blocks by either height/hash `<app> q block --type=height|hash <height|hash>`.
* (x/staking) [#14590](https://github.com/cosmos/cosmos-sdk/pull/14590) Return undelegate amount in MsgUndelegateResponse.
* [#14529](https://github.com/cosmos/cosmos-sdk/pull/14529) Add new property `BondDenom` to `SimulationState` struct.
* (store) [#14439](https://github.com/cosmos/cosmos-sdk/pull/14439) Remove global metric gatherer from store.
    * By default store has a no op metric gatherer, the application developer must set another metric gatherer or us the provided one in `store/metrics`.
* (store) [#14438](https://github.com/cosmos/cosmos-sdk/pull/14438) Pass logger from baseapp to store.
* (baseapp) [#14417](https://github.com/cosmos/cosmos-sdk/pull/14417) The store package no longer has a dependency on baseapp.
* (module) [#14415](https://github.com/cosmos/cosmos-sdk/pull/14415) Loosen assertions in SetOrderBeginBlockers() and SetOrderEndBlockers().
* (store) [#14410](https://github.com/cosmos/cosmos-sdk/pull/14410) `rootmulti.Store.loadVersion` has validation to check if all the module stores' height is correct, it will error if any module store has incorrect height.
* [#14406](https://github.com/cosmos/cosmos-sdk/issues/14406) Migrate usage of `types/store.go` to `store/types/..`.
* (context)[#14384](https://github.com/cosmos/cosmos-sdk/pull/14384) Refactor(context): Pass EventManager to the context as an interface.
* (types) [#14354](https://github.com/cosmos/cosmos-sdk/pull/14354) Improve performance on Context.KVStore and Context.TransientStore by 40%.
* (crypto/keyring) [#14151](https://github.com/cosmos/cosmos-sdk/pull/14151) Move keys presentation from `crypto/keyring` to `client/keys`
* (signing) [#14087](https://github.com/cosmos/cosmos-sdk/pull/14087) Add SignModeHandlerWithContext interface with a new `GetSignBytesWithContext` to get the sign bytes using `context.Context` as an argument to access state.
* (server) [#14062](https://github.com/cosmos/cosmos-sdk/pull/14062) Remove rosetta from server start.
* (crypto) [#3129](https://github.com/cosmos/cosmos-sdk/pull/3129) New armor and keyring key derivation uses aead and encryption uses chacha20poly.

### State Machine Breaking

* (x/gov) [#18146](https://github.com/cosmos/cosmos-sdk/pull/18146) Add denom check to reject denoms outside of those listed in `MinDeposit`. A new `MinDepositRatio` param is added (with a default value of `0.001`) and now deposits are required to be at least `MinDepositRatio*MinDeposit` to be accepted.
* (x/group,x/gov) [#16235](https://github.com/cosmos/cosmos-sdk/pull/16235) A group and gov proposal is rejected if the proposal metadata title and summary do not match the proposal title and summary.
* (baseapp) [#15930](https://github.com/cosmos/cosmos-sdk/pull/15930) change vote info provided by prepare and process proposal to the one in the block.
* (x/staking) [#15731](https://github.com/cosmos/cosmos-sdk/pull/15731) Introducing a new index to retrieve the delegations by validator efficiently.
* (x/staking) [#15701](https://github.com/cosmos/cosmos-sdk/pull/15701) The `HistoricalInfoKey` has been updated to use a binary format.
* (x/slashing) [#15580](https://github.com/cosmos/cosmos-sdk/pull/15580) The validator slashing window now stores "chunked" bitmap entries for each validator's signing window instead of a single boolean entry per signing window index.
* (x/staking) [#14590](https://github.com/cosmos/cosmos-sdk/pull/14590) `MsgUndelegateResponse` now includes undelegated amount. `x/staking` module's `keeper.Undelegate` now returns 3 values (completionTime,undelegateAmount,error) instead of 2.
* (x/feegrant) [#14294](https://github.com/cosmos/cosmos-sdk/pull/14294) Moved the logic of rejecting duplicate grant from `msg_server` to `keeper` method.

### API Breaking Changes

* (x/auth) [#17787](https://github.com/cosmos/cosmos-sdk/pull/17787) Remove Tip functionality.
* (types) `module.EndBlockAppModule` has been replaced by Core API `appmodule.HasEndBlocker` or `module.HasABCIEndBlock` when needing validator updates.
* (types) `module.BeginBlockAppModule` has been replaced by Core API `appmodule.HasBeginBlocker`.
* (types) [#17358](https://github.com/cosmos/cosmos-sdk/pull/17358) Remove deprecated `sdk.Handler`, use `baseapp.MsgServiceHandler` instead.
* (client) [#17197](https://github.com/cosmos/cosmos-sdk/pull/17197) `keys.Commands` does not take a home directory anymore. It is inferred from the root command.
* (x/staking) [#17157](https://github.com/cosmos/cosmos-sdk/pull/17157) `GetValidatorsByPowerIndexKey` and `ValidateBasic` for historical info takes a validator address codec in order to be able to decode/encode addresses.
    * `GetOperator()` now returns the address as it is represented in state, by default this is an encoded address
    * `GetConsAddr() ([]byte, error)` returns `[]byte` instead of sdk.ConsAddres.
    * `FromABCIEvidence` & `GetConsensusAddress(consAc address.Codec)` now take a consensus address codec to be able to decode the incoming address.
    * (x/distribution) `Delegate` & `SlashValidator` helper function added the mock staking keeper as a parameter passed to the function
* (x/staking) [#17098](https://github.com/cosmos/cosmos-sdk/pull/17098) `NewMsgCreateValidator`, `NewValidator`, `NewMsgCancelUnbondingDelegation`, `NewMsgUndelegate`, `NewMsgBeginRedelegate`, `NewMsgDelegate` and `NewMsgEditValidator`  takes a string instead of `sdk.ValAddress` or `sdk.AccAddress`:
    * `NewRedelegation` and `NewUnbondingDelegation` takes a validatorAddressCodec and a delegatorAddressCodec in order to decode the addresses.
    * `NewRedelegationResponse` takes a string instead of `sdk.ValAddress` or `sdk.AccAddress`.
    * `NewMsgCreateValidator.Validate()` takes an address codec in order to decode the address.
    * `BuildCreateValidatorMsg` takes a ValidatorAddressCodec in order to decode addresses.
* (x/slashing) [#17098](https://github.com/cosmos/cosmos-sdk/pull/17098) `NewMsgUnjail` takes a string instead of `sdk.ValAddress`
* (x/genutil) [#17098](https://github.com/cosmos/cosmos-sdk/pull/17098) `GenAppStateFromConfig`, AddGenesisAccountCmd and `GenTxCmd` takes an addresscodec to decode addresses.
* (x/distribution) [#17098](https://github.com/cosmos/cosmos-sdk/pull/17098) `NewMsgDepositValidatorRewardsPool`, `NewMsgFundCommunityPool`, `NewMsgWithdrawValidatorCommission` and `NewMsgWithdrawDelegatorReward` takes a string instead of `sdk.ValAddress` or `sdk.AccAddress`.
* (x/staking) [#16959](https://github.com/cosmos/cosmos-sdk/pull/16959) Add validator and consensus address codec as staking keeper arguments.
* (x/staking) [#16958](https://github.com/cosmos/cosmos-sdk/pull/16958) DelegationI interface `GetDelegatorAddr` & `GetValidatorAddr` have been migrated to return string instead of sdk.AccAddress and sdk.ValAddress respectively. stakingtypes.NewDelegation takes a string instead of sdk.AccAddress and sdk.ValAddress.
* (testutil) [#16899](https://github.com/cosmos/cosmos-sdk/pull/16899) The *cli testutil* `QueryBalancesExec` has been removed. Use the gRPC or REST query instead.
* (x/staking) [#16795](https://github.com/cosmos/cosmos-sdk/pull/16795) `DelegationToDelegationResponse`, `DelegationsToDelegationResponses`, `RedelegationsToRedelegationResponses` are no longer exported.
* (x/auth/vesting) [#16741](https://github.com/cosmos/cosmos-sdk/pull/16741) Vesting account constructor now return an error with the result of their validate function.
* (x/auth) [#16650](https://github.com/cosmos/cosmos-sdk/pull/16650) The *cli testutil* `QueryAccountExec` has been removed. Use the gRPC or REST query instead.
* (x/auth) [#16621](https://github.com/cosmos/cosmos-sdk/pull/16621) Pass address codec to auth new keeper constructor.
* (x/auth) [#16423](https://github.com/cosmos/cosmos-sdk/pull/16423) `helpers.AddGenesisAccount` has been moved to `x/genutil` to remove the cyclic dependency between `x/auth` and `x/genutil`.
* (baseapp) [#16342](https://github.com/cosmos/cosmos-sdk/pull/16342) NewContext was renamed to NewContextLegacy. The replacement (NewContext) now does not take a header, instead you should set the header via `WithHeaderInfo` or `WithBlockHeight`. Note that `WithBlockHeight` will soon be depreacted and its recommneded to use `WithHeaderInfo`.
* (x/mint) [#16329](https://github.com/cosmos/cosmos-sdk/pull/16329) Use collections for state management:
    * Removed: keeper `GetParams`, `SetParams`, `GetMinter`, `SetMinter`.
* (x/crisis) [#16328](https://github.com/cosmos/cosmos-sdk/pull/16328) Use collections for state management:
    * Removed: keeper `GetConstantFee`, `SetConstantFee`
* (x/staking) [#16324](https://github.com/cosmos/cosmos-sdk/pull/16324) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`, and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context` and return an `error`. Notable changes:
    * `Validator` method now returns `types.ErrNoValidatorFound` instead of `nil` when not found.
* (x/distribution) [#16302](https://github.com/cosmos/cosmos-sdk/pull/16302) Use collections for FeePool state management.
    * Removed: keeper `GetFeePool`, `SetFeePool`, `GetFeePoolCommunityCoins`
* (types) [#16272](https://github.com/cosmos/cosmos-sdk/pull/16272) `FeeGranter` in the `FeeTx` interface returns `[]byte` instead of `string`.
* (x/gov) [#16268](https://github.com/cosmos/cosmos-sdk/pull/16268) Use collections for proposal state management (part 2):
    * this finalizes the gov collections migration
    * Removed: types all the key related functions
    * Removed: keeper `InsertActiveProposalsQueue`, `RemoveActiveProposalsQueue`, `InsertInactiveProposalsQueue`, `RemoveInactiveProposalsQueue`, `IterateInactiveProposalsQueue`, `IterateActiveProposalsQueue`, `ActiveProposalsQueueIterator`, `InactiveProposalsQueueIterator`
* (x/slashing) [#16246](https://github.com/cosmos/cosmos-sdk/issues/16246) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`, and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context` and return an `error`. `GetValidatorSigningInfo` now returns an error instead of a `found bool`, the error can be `nil` (found), `ErrNoSigningInfoFound` (not found) and any other error.
* (module) [#16227](https://github.com/cosmos/cosmos-sdk/issues/16227) `manager.RunMigrations()` now take a `context.Context` instead of a `sdk.Context`.
* (x/crisis) [#16216](https://github.com/cosmos/cosmos-sdk/issues/16216) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`, methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context` and return an `error` instead of panicking.
* (x/distribution) [#16211](https://github.com/cosmos/cosmos-sdk/pull/16211) Use collections for params state management.
* (cli) [#16209](https://github.com/cosmos/cosmos-sdk/pull/16209) Add API `StartCmdWithOptions` to create customized start command.
* (x/mint) [#16179](https://github.com/cosmos/cosmos-sdk/issues/16179) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`, and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context` and return an `error`.
* (x/gov) [#16171](https://github.com/cosmos/cosmos-sdk/pull/16171) Use collections for proposal state management (part 1):
    * Removed: keeper: `GetProposal`, `UnmarshalProposal`, `MarshalProposal`, `IterateProposal`, `GetProposal`, `GetProposalFiltered`, `GetProposals`, `GetProposalID`, `SetProposalID`
    * Removed: errors unused errors
* (x/gov) [#16164](https://github.com/cosmos/cosmos-sdk/pull/16164) Use collections for vote state management:
    * Removed: types `VoteKey`, `VoteKeys`
    * Removed: keeper `IterateVotes`, `IterateAllVotes`, `GetVotes`, `GetVote`, `SetVote`
* (sims) [#16155](https://github.com/cosmos/cosmos-sdk/pull/16155)
    * `simulation.NewOperationMsg` now marshals the operation msg as proto bytes instead of legacy amino JSON bytes.
    * `simulation.NewOperationMsg` is now 2-arity instead of 3-arity with the obsolete argument `codec.ProtoCodec` removed.
    * The field `OperationMsg.Msg` is now of type `[]byte` instead of `json.RawMessage`.
* (x/gov) [#16127](https://github.com/cosmos/cosmos-sdk/pull/16127) Use collections for deposit state management:
    * The following methods are removed from the gov keeper: `GetDeposit`, `GetAllDeposits`, `IterateAllDeposits`.
    * The following functions are removed from the gov types: `DepositKey`, `DepositsKey`.
* (x/gov) [#16118](https://github.com/cosmos/cosmos-sdk/pull/16118/) Use collections for constituion and params state management.
* (x/gov) [#16106](https://github.com/cosmos/cosmos-sdk/pull/16106) Remove gRPC query methods from gov keeper.
* (x/*all*) [#16052](https://github.com/cosmos/cosmos-sdk/pull/16062) `GetSignBytes` implementations on messages and global legacy amino codec definitions have been removed from all modules.
* (sims) [#16052](https://github.com/cosmos/cosmos-sdk/pull/16062) `GetOrGenerate` no longer requires a codec argument is now 4-arity instead of 5-arity.
* (types/math) [#16040](https://github.com/cosmos/cosmos-sdk/pull/16798) Remove aliases in `types/math.go` (part 2).
* (types/math) [#16040](https://github.com/cosmos/cosmos-sdk/pull/16040) Remove aliases in `types/math.go` (part 1).
* (x/auth) [#16016](https://github.com/cosmos/cosmos-sdk/pull/16016) Use collections for accounts state management:
    * removed: keeper `HasAccountByID`, `AccountAddressByID`, `SetParams
* (x/genutil) [#15999](https://github.com/cosmos/cosmos-sdk/pull/15999) Genutil now takes the `GenesisTxHanlder` interface instead of deliverTx. The interface is implemented on baseapp
* (x/gov) [#15988](https://github.com/cosmos/cosmos-sdk/issues/15988) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`, methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context` and return an `error` (instead of panicking or returning a `found bool`). Iterators callback functions now return an error instead of a `bool`.
* (x/auth) [#15985](https://github.com/cosmos/cosmos-sdk/pull/15985) The `AccountKeeper` does not expose the `QueryServer` and `MsgServer` APIs anymore.
* (x/authz) [#15962](https://github.com/cosmos/cosmos-sdk/issues/15962) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`, methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context`. The `Authorization` interface's `Accept` method now takes a `context.Context` instead of a `sdk.Context`.
* (x/distribution) [#15948](https://github.com/cosmos/cosmos-sdk/issues/15948) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey` and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context`. Keeper methods also now return an `error`.
* (x/bank) [#15891](https://github.com/cosmos/cosmos-sdk/issues/15891) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey` and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context`. Also `FundAccount` and `FundModuleAccount` from the `testutil` package accept a `context.Context` instead of a `sdk.Context`, and it's position was moved to the first place.
* (x/slashing) [#15875](https://github.com/cosmos/cosmos-sdk/pull/15875) `x/slashing.NewAppModule` now requires an `InterfaceRegistry` parameter.
* (x/crisis) [#15852](https://github.com/cosmos/cosmos-sdk/pull/15852) Crisis keeper now takes a instance of the address codec to be able to decode user addresses
* (x/auth) [#15822](https://github.com/cosmos/cosmos-sdk/pull/15822) The type of struct field `ante.HandlerOptions.SignModeHandler` has been changed to `x/tx/signing.HandlerMap`.
* (client) [#15822](https://github.com/cosmos/cosmos-sdk/pull/15822) The return type of the interface method `TxConfig.SignModeHandler` has been changed to `x/tx/signing.HandlerMap`.
    * The signature of `VerifySignature` has been changed to accept a `x/tx/signing.HandlerMap` and other structs from `x/tx` as arguments.
    * The signature of `NewTxConfigWithTextual` has been deprecated and its signature changed to accept a `SignModeOptions`.
    * The signature of `NewSigVerificationDecorator` has been changed to accept a `x/tx/signing.HandlerMap`.
* (x/bank) [#15818](https://github.com/cosmos/cosmos-sdk/issues/15818) `BaseViewKeeper`'s `Logger` method now doesn't require a context. `NewBaseKeeper`, `NewBaseSendKeeper` and `NewBaseViewKeeper` now also require a `log.Logger` to be passed in.
* (x/genutil) [#15679](https://github.com/cosmos/cosmos-sdk/pull/15679) `MigrateGenesisCmd` now takes a `MigrationMap` instead of having the SDK genesis migration hardcoded.
* (client) [#15673](https://github.com/cosmos/cosmos-sdk/pull/15673) Move `client/keys.OutputFormatJSON` and `client/keys.OutputFormatText` to `client/flags` package.
* (x/*all*) [#15648](https://github.com/cosmos/cosmos-sdk/issues/15648) Make `SetParams` consistent across all modules and validate the params at the message handling instead of `SetParams` method.
* (codec) [#15600](https://github.com/cosmos/cosmos-sdk/pull/15600) [#15873](https://github.com/cosmos/cosmos-sdk/pull/15873) add support for getting signers to `codec.Codec` and `InterfaceRegistry`:
    * `InterfaceRegistry` is has unexported methods and implements `protodesc.Resolver` plus the `RangeFiles` and `SigningContext` methods. All implementations of `InterfaceRegistry` by other users must now embed the official implementation.
    * `Codec` has new methods `InterfaceRegistry`, `GetMsgAnySigners`, `GetMsgV1Signers`, and `GetMsgV2Signers` as well as unexported methods. All implementations of `Codec` by other users must now embed an official implementation from the `codec` package.
    * `AminoCodec` is marked as deprecated and no longer implements `Codec.
* (client) [#15597](https://github.com/cosmos/cosmos-sdk/pull/15597) `RegisterNodeService` now requires a config parameter.
* (x/nft) [#15588](https://github.com/cosmos/cosmos-sdk/pull/15588) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey` and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context`.
* (baseapp) [#15568](https://github.com/cosmos/cosmos-sdk/pull/15568) `SetIAVLLazyLoading` is removed from baseapp.
* (x/genutil) [#15567](https://github.com/cosmos/cosmos-sdk/pull/15567) `CollectGenTxsCmd` & `GenTxCmd` takes a address.Codec to be able to decode addresses.
* (x/bank) [#15567](https://github.com/cosmos/cosmos-sdk/pull/15567) `GenesisBalance.GetAddress` now returns a string instead of `sdk.AccAddress`
    * `MsgSendExec` test helper function now takes a address.Codec
* (x/auth) [#15520](https://github.com/cosmos/cosmos-sdk/pull/15520) `NewAccountKeeper` now takes a `KVStoreService` instead of a `StoreKey` and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context`.
* (baseapp) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) `runTxMode`s were renamed to `execMode`. `ModeDeliver` as changed to `ModeFinalize` and a new `ModeVoteExtension` was added for vote extensions.
* (baseapp) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) Writing of state to the multistore was moved to `FinalizeBlock`. `Commit` still handles the committing values to disk.
* (baseapp) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) Calls to BeginBlock and EndBlock have been replaced with core api beginblock & endblock.
* (baseapp) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) BeginBlock and EndBlock are now internal to baseapp. For testing, user must call `FinalizeBlock`. BeginBlock and EndBlock calls are internal to Baseapp.
* (baseapp) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) All calls to ABCI methods now accept a pointer of the abci request and response types
* (x/consensus) [#15517](https://github.com/cosmos/cosmos-sdk/pull/15517) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`.
* (x/bank) [#15477](https://github.com/cosmos/cosmos-sdk/pull/15477) `banktypes.NewMsgMultiSend` and `keeper.InputOutputCoins` only accept one input.
* (server) [#15358](https://github.com/cosmos/cosmos-sdk/pull/15358) Remove `server.ErrorCode` that was not used anywhere.
* (x/capability) [#15344](https://github.com/cosmos/cosmos-sdk/pull/15344) Capability module was removed and is now housed in [IBC-GO](https://github.com/cosmos/ibc-go).
* (mempool) [#15328](https://github.com/cosmos/cosmos-sdk/pull/15328) The `PriorityNonceMempool` is now generic over type `C comparable` and takes a single `PriorityNonceMempoolConfig[C]` argument. See `DefaultPriorityNonceMempoolConfig` for how to construct the configuration and a `TxPriority` type.
* [#15299](https://github.com/cosmos/cosmos-sdk/pull/15299) Remove `StdTx` transaction and signing APIs. No SDK version has actually supported `StdTx` since before Stargate.
* [#15284](https://github.com/cosmos/cosmos-sdk/pull/15284)
* (x/gov) [#15284](https://github.com/cosmos/cosmos-sdk/pull/15284) `NewKeeper` now requires `codec.Codec`.
* (x/authx) [#15284](https://github.com/cosmos/cosmos-sdk/pull/15284) `NewKeeper` now requires `codec.Codec`.
    * `types/tx.Tx` no longer implements `sdk.Tx`.
    * `sdk.Tx` now requires a new method `GetMsgsV2()`.
    * `sdk.Msg.GetSigners` was deprecated and is no longer supported. Use the `cosmos.msg.v1.signer` protobuf annotation instead.
    * `TxConfig` has a new method `SigningContext() *signing.Context`.
    * `SigVerifiableTx.GetSigners()` now returns `([][]byte, error)` instead of `[]sdk.AccAddress`.
    * `AccountKeeper` now has an `AddressCodec() address.Codec` method and the expected `AccountKeeper` for `x/auth/ante` expects this method.
* [#15211](https://github.com/cosmos/cosmos-sdk/pull/15211) Remove usage of `github.com/cometbft/cometbft/libs/bytes.HexBytes` in favor of `[]byte` thorough the SDK.
* (crypto) [#15070](https://github.com/cosmos/cosmos-sdk/pull/15070) `GenerateFromPassword` and `Cost` from `bcrypt.go` now take a `uint32` instead of a `int` type.
* (types) [#15067](https://github.com/cosmos/cosmos-sdk/pull/15067) Remove deprecated alias from `types/errors`. Use `cosmossdk.io/errors` instead.
* (server) [#15041](https://github.com/cosmos/cosmos-sdk/pull/15041) Refactor how gRPC and API servers are started to remove unnecessary sleeps:
    * `api.Server#Start` now accepts a `context.Context`. The caller is responsible for ensuring that the context is canceled such that the API server can gracefully exit. The caller does not need to stop the server.
    * To start the gRPC server you must first create the server via `NewGRPCServer`, after which you can start the gRPC server via `StartGRPCServer` which accepts a `context.Context`. The caller is responsible for ensuring that the context is canceled such that the gRPC server can gracefully exit. The caller does not need to stop the server.
    * Rename `WaitForQuitSignals` to `ListenForQuitSignals`. Note, this function is no longer blocking. Thus the caller is expected to provide a `context.CancelFunc` which indicates that when a signal is caught, that any spawned processes can gracefully exit.
    * Remove `ServerStartTime` constant.
* [#15011](https://github.com/cosmos/cosmos-sdk/pull/15011) All functions that were taking a CometBFT logger, now take `cosmossdk.io/log.Logger` instead.
* (simapp) [#14977](https://github.com/cosmos/cosmos-sdk/pull/14977) Move simulation helpers functions (`AppStateFn` and `AppStateRandomizedFn`) to `testutil/sims`. These takes an extra genesisState argument which is the default state of the app.
* (x/bank) [#14894](https://github.com/cosmos/cosmos-sdk/pull/14894) Allow a human readable denomination for coins when querying bank balances. Added a `ResolveDenom` parameter to `types.QueryAllBalancesRequest`.
* [#14847](https://github.com/cosmos/cosmos-sdk/pull/14847) App and ModuleManager methods `InitGenesis`, `ExportGenesis`, `BeginBlock` and `EndBlock` now also return an error.
* (x/upgrade) [#14764](https://github.com/cosmos/cosmos-sdk/pull/14764) The `x/upgrade` module is extracted to have a separate go.mod file which allows it to be a standalone module.
* (x/auth) [#14758](https://github.com/cosmos/cosmos-sdk/pull/14758) Refactor transaction searching:
    * Refactor `QueryTxsByEvents` to accept a `query` of type `string` instead of `events` of type `[]string`
    * Refactor CLI methods to accept `--query` flag instead of `--events`
    * Pass `prove=false` to Tendermint's `TxSearch` RPC method
* (simulation) [#14751](https://github.com/cosmos/cosmos-sdk/pull/14751) Remove the `MsgType` field from `simulation.OperationInput` struct.
* (store) [#14746](https://github.com/cosmos/cosmos-sdk/pull/14746) Extract Store in its own go.mod and rename the package to `cosmossdk.io/store`.
* (x/nft) [#14725](https://github.com/cosmos/cosmos-sdk/pull/14725) Extract NFT in its own go.mod and rename the package to `cosmossdk.io/x/nft`.
* (x/gov) [#14720](https://github.com/cosmos/cosmos-sdk/pull/14720) Add an expedited field in the gov v1 proposal and `MsgNewMsgProposal`.
* (x/feegrant) [#14649](https://github.com/cosmos/cosmos-sdk/pull/14649) Extract Feegrant in its own go.mod and rename the package to `cosmossdk.io/x/feegrant`.
* (tx) [#14634](https://github.com/cosmos/cosmos-sdk/pull/14634) Move the `tx` go module to `x/tx`.
* (store/streaming)[#14603](https://github.com/cosmos/cosmos-sdk/pull/14603) `StoreDecoderRegistry` moved from store to `types/simulations` this breaks the `AppModuleSimulation` interface.
* (snapshots) [#14597](https://github.com/cosmos/cosmos-sdk/pull/14597) Move `snapshots` to `store/snapshots`, rename and bump proto package to v1.
* (x/staking) [#14590](https://github.com/cosmos/cosmos-sdk/pull/14590) `MsgUndelegateResponse` now includes undelegated amount. `x/staking` module's `keeper.Undelegate` now returns 3 values (completionTime,undelegateAmount,error)  instead of 2.
* (crypto/keyring) [#14151](https://github.com/cosmos/cosmos-sdk/pull/14151) Move keys presentation from `crypto/keyring` to `client/keys`
* (baseapp) [#14050](https://github.com/cosmos/cosmos-sdk/pull/14050) Refactor `ABCIListener` interface to accept Go contexts.
* (x/auth) [#13850](https://github.com/cosmos/cosmos-sdk/pull/13850/) Remove `MarshalYAML` methods from module (`x/...`) types.
* (modules) [#13850](https://github.com/cosmos/cosmos-sdk/pull/13850) and [#14046](https://github.com/cosmos/cosmos-sdk/pull/14046) Remove gogoproto stringer annotations. This removes the custom `String()` methods on all types that were using the annotations.
* (x/evidence) [14724](https://github.com/cosmos/cosmos-sdk/pull/14724) Extract Evidence in its own go.mod and rename the package to `cosmossdk.io/x/evidence`.
* (crypto/keyring) [#13734](https://github.com/cosmos/cosmos-sdk/pull/13834) The keyring's `Sign` method now takes a new `signMode` argument. It is only used if the signing key is a Ledger hardware device. You can set it to 0 in all other cases.
* (snapshots) [14048](https://github.com/cosmos/cosmos-sdk/pull/14048) Move the Snapshot package to the store package. This is done in an effort group all storage related logic under one package.
* (signing) [#13701](https://github.com/cosmos/cosmos-sdk/pull/) Add `context.Context` as an argument `x/auth/signing.VerifySignature`.
* (store) [#11825](https://github.com/cosmos/cosmos-sdk/pull/11825) Make extension snapshotter interface safer to use, renamed the util function `WriteExtensionItem` to `WriteExtensionPayload`.

### Client Breaking Changes

* (x/gov) [#17910](https://github.com/cosmos/cosmos-sdk/pull/17910) Remove telemetry for counting votes and proposals. It was incorrectly counting votes. Use alternatives, such as state streaming.
* (abci) [#15845](https://github.com/cosmos/cosmos-sdk/pull/15845) Remove duplicating events in `logs`.
* (abci) [#15845](https://github.com/cosmos/cosmos-sdk/pull/15845) Add `msg_index` to all event attributes to associate events and messages.
* (x/staking) [#15701](https://github.com/cosmos/cosmos-sdk/pull/15701) `HistoricalInfoKey` now has a binary format.
* (store/streaming) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) State Streaming removed emitting of beginblock, endblock and delivertx in favour of emitting FinalizeBlock.
* (baseapp) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) BeginBlock & EndBlock events have begin or endblock in the events in order to identify which stage they are emitted from since they are returned to comet as FinalizeBlock events.
* (grpc-web) [#14652](https://github.com/cosmos/cosmos-sdk/pull/14652) Use same port for gRPC-Web and the API server.

### CLI Breaking Changes

* (all) The migration of modules to [AutoCLI](https://docs.cosmos.network/main/core/autocli) led to no changes in UX but a [small change in CLI outputs](https://github.com/cosmos/cosmos-sdk/issues/16651) where results can be nested.
* (all) Query pagination flags have been renamed with the migration to AutoCLI:
    * `--reverse` -> `--page-reverse`
    * `--offset` -> `--page-offset`
    * `--limit` -> `--page-limit`
    * `--count-total` -> `--page-count-total`
* (cli) [#17184](https://github.com/cosmos/cosmos-sdk/pull/17184) All json keys returned by the `status` command are now snake case instead of pascal case.
* (server) [#17177](https://github.com/cosmos/cosmos-sdk/pull/17177) Remove `iavl-lazy-loading` configuration.
* (x/gov) [#16987](https://github.com/cosmos/cosmos-sdk/pull/16987) In `<appd> query gov proposals` the proposal status flag have renamed from `--status` to `--proposal-status`. Additionally, that flags now uses the ENUM values: `PROPOSAL_STATUS_DEPOSIT_PERIOD`, `PROPOSAL_STATUS_VOTING_PERIOD`, `PROPOSAL_STATUS_PASSED`, `PROPOSAL_STATUS_REJECTED`, `PROPOSAL_STATUS_FAILED`.
* (x/bank) [#16899](https://github.com/cosmos/cosmos-sdk/pull/16899) With the migration to AutoCLI some bank commands have been split in two:
    * Use `total-supply` (or `total`) for querying the total supply and `total-supply-of` for querying the supply of a specific denom.
    * Use `denoms-metadata` for querying all denom metadata and `denom-metadata` for querying a specific denom metadata.
* (rosetta) [#16276](https://github.com/cosmos/cosmos-sdk/issues/16276) Rosetta migration to standalone repo.
* (cli) [#15826](https://github.com/cosmos/cosmos-sdk/pull/15826) Remove `<appd> q account` command. Use `<appd> q auth account` instead.
* (cli) [#15299](https://github.com/cosmos/cosmos-sdk/pull/15299) Remove `--amino` flag from `sign` and `multi-sign` commands. Amino `StdTx` has been deprecated for a while. Amino JSON signing still works as expected.
* (x/gov) [#14880](https://github.com/cosmos/cosmos-sdk/pull/14880) Remove `<app> tx gov submit-legacy-proposal cancel-software-upgrade` and `software-upgrade` commands. These commands are now in the `x/upgrade` module and using gov v1. Use `tx upgrade software-upgrade` instead.
* (x/staking) [#14864](https://github.com/cosmos/cosmos-sdk/pull/14864) `<appd> tx staking create-validator` CLI command now takes a json file as an arg instead of using required flags.
* (cli) [#14659](https://github.com/cosmos/cosmos-sdk/pull/14659) `<app> q block <height>` is removed as it just output json. The new command allows either height/hash and is `<app> q block --type=height|hash <height|hash>`.
* (grpc-web) [#14652](https://github.com/cosmos/cosmos-sdk/pull/14652) Remove `grpc-web.address` flag.
* (client) [#14342](https://github.com/cosmos/cosmos-sdk/pull/14342) `<app> config` command is now a sub-command using Confix. Use `<app> config --help` to learn more.

### Bug Fixes

* (server) [#18254](https://github.com/cosmos/cosmos-sdk/pull/18254) Don't hardcode gRPC address to localhost.
* (x/gov) [#18173](https://github.com/cosmos/cosmos-sdk/pull/18173) Gov hooks now return an error and are *blocking* when they fail. Expect for `AfterProposalFailedMinDeposit` and `AfterProposalVotingPeriodEnded` which log the error and continue.
* (x/gov) [#17873](https://github.com/cosmos/cosmos-sdk/pull/17873) Fail any inactive and active proposals that cannot be decoded.
* (x/slashing) [#18016](https://github.com/cosmos/cosmos-sdk/pull/18016) Fixed builder function for missed blocks key (`validatorMissedBlockBitArrayPrefixKey`) in slashing/migration/v4.
* (x/bank) [#18107](https://github.com/cosmos/cosmos-sdk/pull/18107) Add missing keypair of SendEnabled to restore legacy param set before migration.
* (baseapp) [#17769](https://github.com/cosmos/cosmos-sdk/pull/17769) Ensure we respect block size constraints in the `DefaultProposalHandler`'s `PrepareProposal` handler when a nil or no-op mempool is used. We provide a `TxSelector` type to assist in making transaction selection generalized. We also fix a comparison bug in tx selection when `req.maxTxBytes` is reached.
* (mempool) [#17668](https://github.com/cosmos/cosmos-sdk/pull/17668) Fix `PriorityNonceIterator.Next()` nil pointer ref for min priority at the end of iteration.
* (config) [#17649](https://github.com/cosmos/cosmos-sdk/pull/17649) Fix `mempool.max-txs` configuration is invalid in `app.config`.
* (baseapp) [#17518](https://github.com/cosmos/cosmos-sdk/pull/17518) Utilizing voting power from vote extensions (CometBFT) instead of the current bonded tokens (x/staking) to determine if a set of vote extensions are valid.
* (baseapp) [#17251](https://github.com/cosmos/cosmos-sdk/pull/17251) VerifyVoteExtensions and ExtendVote initialize their own contexts/states, allowing VerifyVoteExtensions being called without ExtendVote.
* (x/distribution) [#17236](https://github.com/cosmos/cosmos-sdk/pull/17236) Using "validateCommunityTax" in "Params.ValidateBasic", preventing panic when field "CommunityTax" is nil.
* (x/bank) [#17170](https://github.com/cosmos/cosmos-sdk/pull/17170) Avoid empty spendable error message on send coins.
* (x/group) [#17146](https://github.com/cosmos/cosmos-sdk/pull/17146) Rename x/group legacy ORM package's error codespace from "orm" to "legacy_orm", preventing collisions with ORM's error codespace "orm".
* (types/query) [#16905](https://github.com/cosmos/cosmos-sdk/pull/16905) Collections Pagination now applies proper count when filtering results.
* (x/bank) [#16841](https://github.com/cosmos/cosmos-sdk/pull/16841) Correctly process legacy `DenomAddressIndex` values.
* (x/auth/vesting) [#16733](https://github.com/cosmos/cosmos-sdk/pull/16733) Panic on overflowing and negative EndTimes when creating a PeriodicVestingAccount.
* (x/consensus) [#16713](https://github.com/cosmos/cosmos-sdk/pull/16713) Add missing ABCI param in `MsgUpdateParams`.
* (baseapp) [#16700](https://github.com/cosmos/cosmos-sdk/pull/16700) Fix consensus failure in returning no response to malformed transactions.
* [#16639](https://github.com/cosmos/cosmos-sdk/pull/16639) Make sure we don't execute blocks beyond the halt height.
* (baseapp) [#16613](https://github.com/cosmos/cosmos-sdk/pull/16613) Ensure each message in a transaction has a registered handler, otherwise `CheckTx` will fail.
* (baseapp) [#16596](https://github.com/cosmos/cosmos-sdk/pull/16596) Return error during `ExtendVote` and `VerifyVoteExtension` if the request height is earlier than `VoteExtensionsEnableHeight`.
* (baseapp) [#16259](https://github.com/cosmos/cosmos-sdk/pull/16259) Ensure the `Context` block height is correct after `InitChain` and prior to the second block.
* (x/gov) [#16231](https://github.com/cosmos/cosmos-sdk/pull/16231) Fix Rawlog JSON formatting of proposal_vote option field.* (cli) [#16138](https://github.com/cosmos/cosmos-sdk/pull/16138) Fix snapshot commands panic if snapshot don't exists.
* (x/staking) [#16043](https://github.com/cosmos/cosmos-sdk/pull/16043) Call `AfterUnbondingInitiated` hook for new unbonding entries only and fix `UnbondingDelegation` entries handling. This is a behavior change compared to Cosmos SDK v0.47.x, now the hook is called only for new unbonding entries.
* (types) [#16010](https://github.com/cosmos/cosmos-sdk/pull/16010) Let `module.CoreAppModuleBasicAdaptor` fallback to legacy genesis handling.
* (types) [#15691](https://github.com/cosmos/cosmos-sdk/pull/15691) Make `Coin.Validate()` check that `.Amount` is not nil.
* (x/crypto) [#15258](https://github.com/cosmos/cosmos-sdk/pull/15258) Write keyhash file with permissions 0600 instead of 0555.
* (x/auth) [#15059](https://github.com/cosmos/cosmos-sdk/pull/15059) `ante.CountSubKeys` returns 0 when passing a nil `Pubkey`.
* (x/capability) [#15030](https://github.com/cosmos/cosmos-sdk/pull/15030) Prevent `x/capability` from consuming `GasMeter` gas during `InitMemStore`
* (types/coin) [#14739](https://github.com/cosmos/cosmos-sdk/pull/14739) Deprecate the method `Coin.IsEqual` in favour of  `Coin.Equal`. The difference between the two methods is that the first one results in a panic when denoms are not equal. This panic lead to unexpected behavior.

### Deprecated

* (types) [#16980](https://github.com/cosmos/cosmos-sdk/pull/16980) Deprecate `IntProto` and `DecProto`. Instead, `math.Int` and `math.LegacyDec` should be used respectively. Both types support `Marshal` and `Unmarshal` for binary serialization.
* (x/staking) [#14567](https://github.com/cosmos/cosmos-sdk/pull/14567) The `delegator_address` field of `MsgCreateValidator` has been deprecated.
  The validator address bytes and delegator address bytes refer to the same account while creating validator (defer only in bech32 notation).

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/cosmos-sdk/blob/main/CHANGELOG.md#v0470---2023-03-14).
