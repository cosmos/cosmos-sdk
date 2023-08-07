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

<<<<<<< HEAD
=======
### Features

* (x/bank) [#16795](https://github.com/cosmos/cosmos-sdk/pull/16852) Add `DenomMetadataByQueryString` query in bank module to support metadata query by query string.

### Improvements

* (x/group, x/gov) [#17220](https://github.com/cosmos/cosmos-sdk/pull/17220) Add `--skip-metadata` flag in `draft-proposal` to skip metadata prompt.
* (x/group, x/gov) [#17109](https://github.com/cosmos/cosmos-sdk/pull/17109) Let proposal summary be 40x longer than metadata limit.
* (all) [#16537](https://github.com/cosmos/cosmos-sdk/pull/16537) Properly propagated `fmt.Errorf` errors and using `errors.New` where appropriate.
* (version) [#17096](https://github.com/cosmos/cosmos-sdk/pull/17096) Improve `getSDKVersion()` to handle module replacements
* (x/staking) [#17164](https://github.com/cosmos/cosmos-sdk/pull/17164) Add `BondedTokensAndPubKeyByConsAddr` to the keeper to enable vote extension verification.
* (x/genutil) [#17296](https://github.com/cosmos/cosmos-sdk/pull/17296) Add `MigrateHandler` to allow reuse migrate genesis related function.

### Bug Fixes

* (baseapp) [#17251](https://github.com/cosmos/cosmos-sdk/pull/17251) VerifyVoteExtensions and ExtendVote initialize their own contexts/states, allowing VerifyVoteExtensions being called without ExtendVote.
* (x/auth) [#17209](https://github.com/cosmos/cosmos-sdk/pull/17209) Internal error on AccountInfo when account's public key is not set.
* (baseapp) [#17159](https://github.com/cosmos/cosmos-sdk/pull/17159) Validators can propose blocks that exceed the gas limit.
* (x/group) [#17146](https://github.com/cosmos/cosmos-sdk/pull/17146) Rename x/group legacy ORM package's error codespace from "orm" to "legacy_orm", preventing collisions with ORM's error codespace "orm".
* (x/bank) [#17170](https://github.com/cosmos/cosmos-sdk/pull/17170) Avoid empty spendable error message on send coins.
* (x/distribution) [#17236](https://github.com/cosmos/cosmos-sdk/pull/17236) Using "validateCommunityTax" in "Params.ValidateBasic", preventing panic when field "CommunityTax" is nil.

### API Breaking Changes

* (x/staking) [#17256](https://github.com/cosmos/cosmos-sdk/pull/17256) Use collections for `UnbondingID`.
* (x/staking) [#17260](https://github.com/cosmos/cosmos-sdk/pull/17260) Use collections for `ValidatorByConsAddr`:
    * remove from `types`: `GetValidatorByConsAddrKey`
* (x/staking) [#17248](https://github.com/cosmos/cosmos-sdk/pull/17248) Use collections for `UnbondingType`.
    * remove from `types`: `GetUnbondingTypeKey`.
* (client) [#17259](https://github.com/cosmos/cosmos-sdk/pull/17259) Remove deprecated `clientCtx.PrintObjectLegacy`. Use `clientCtx.PrintProto` or `clientCtx.PrintRaw` instead.
* (x/distribution) [#17115](https://github.com/cosmos/cosmos-sdk/pull/17115) Use collections for `PreviousProposer` and `ValidatorSlashEvents`:
    * remove from `Keeper`: `GetPreviousProposerConsAddr`, `SetPreviousProposerConsAddr`, `GetValidatorHistoricalReferenceCount`, `GetValidatorSlashEvent`, `SetValidatorSlashEvent`.
* (x/feegrant) [#16535](https://github.com/cosmos/cosmos-sdk/pull/16535) Use collections for `FeeAllowance`, `FeeAllowanceQueue`.
* (x/staking) [#17063](https://github.com/cosmos/cosmos-sdk/pull/17063) Use collections for `HistoricalInfo`:
    * remove `Keeper`: `GetHistoricalInfo`, `SetHistoricalInfo`,
* (x/staking) [#17062](https://github.com/cosmos/cosmos-sdk/pull/17062) Use collections for `ValidatorUpdates`:
    * remove `Keeper`: `SetValidatorUpdates`, `GetValidatorUpdates`
* (x/slashing) [#17023](https://github.com/cosmos/cosmos-sdk/pull/17023) Use collections for `ValidatorSigningInfo`:
    * remove `Keeper`: `SetValidatorSigningInfo`, `GetValidatorSigningInfo`, `IterateValidatorSigningInfos`
* (x/staking) [#17026](https://github.com/cosmos/cosmos-sdk/pull/17026) Use collections for `LastTotalPower`:
    * remove `Keeper`: `SetLastTotalPower`, `GetLastTotalPower`
* (x/distribution) [#16440](https://github.com/cosmos/cosmos-sdk/pull/16440) use collections for `DelegatorWithdrawAddresState`:
    * remove `Keeper`: `SetDelegatorWithdrawAddr`, `DeleteDelegatorWithdrawAddr`, `IterateDelegatorWithdrawAddrs`.
* (x/distribution) [#16459](https://github.com/cosmos/cosmos-sdk/pull/16459) use collections for `ValidatorCurrentRewards` state management:
    * remove `Keeper`: `IterateValidatorCurrentRewards`, `GetValidatorCurrentRewards`, `SetValidatorCurrentRewards`, `DeleteValidatorCurrentRewards`
* (x/authz) [#16509](https://github.com/cosmos/cosmos-sdk/pull/16509) `AcceptResponse` has been moved to sdk/types/authz and the `Updated` field is now of the type `sdk.Msg` instead of `authz.Authorization`.
* (x/distribution) [#16483](https://github.com/cosmos/cosmos-sdk/pull/16483) use collections for `DelegatorStartingInfo` state management:
    * remove `Keeper`: `IterateDelegatorStartingInfo`, `GetDelegatorStartingInfo`, `SetDelegatorStartingInfo`, `DeleteDelegatorStartingInfo`, `HasDelegatorStartingInfo`
* (x/distribution) [#16571](https://github.com/cosmos/cosmos-sdk/pull/16571) use collections for `ValidatorAccumulatedCommission` state management:
    * remove `Keeper`: `IterateValidatorAccumulatedCommission`, `GetValidatorAccumulatedCommission`, `SetValidatorAccumulatedCommission`, `DeleteValidatorAccumulatedCommission`
* (x/distribution) [#16590](https://github.com/cosmos/cosmos-sdk/pull/16590) use collections for `ValidatorOutstandingRewards` state management:
    * remove `Keeper`: `IterateValidatorOutstandingRewards`, `GetValidatorOutstandingRewards`, `SetValidatorOutstandingRewards`, `DeleteValidatorOutstandingRewards`
* (x/distribution) [#16607](https://github.com/cosmos/cosmos-sdk/pull/16607) use collections for `ValidatorHistoricalRewards` state management:
    * remove `Keeper`: `IterateValidatorHistoricalRewards`, `GetValidatorHistoricalRewards`, `SetValidatorHistoricalRewards`, `DeleteValidatorHistoricalRewards`, `DeleteValidatorHistoricalReward`, `DeleteAllValidatorHistoricalRewards`
* (x/slashing) [#16441](https://github.com/cosmos/cosmos-sdk/pull/16441) Params state is migrated to collections. `GetParams` has been removed.
* (types) [#16918](https://github.com/cosmos/cosmos-sdk/pull/16918) Remove `IntProto` and `DecProto`. Instead, `math.Int` and `math.LegacyDec` should be used respectively. Both types support `Marshal` and `Unmarshal` which should be used for binary marshaling.
* (x/staking) [#17098](https://github.com/cosmos/cosmos-sdk/pull/17098) `NewMsgCreateValidator`, `NewValidator`, `NewMsgCancelUnbondingDelegation`, `NewMsgUndelegate`, `NewMsgBeginRedelegate`, `NewMsgDelegate` and `NewMsgEditValidator`  takes a string instead of `sdk.ValAddress` or `sdk.AccAddress`
    * `NewMsgCreateValidator.Validate()` takes an address codec in order to decode the address
    * `NewRedelegationResponse` takes a string instead of `sdk.ValAddress` or `sdk.AccAddress`
    * `NewRedelegation` and `NewUnbondingDelegation` takes a validatorAddressCodec and a delegatorAddressCodec in order to decode the addresses
    * `BuildCreateValidatorMsg` takes a ValidatorAddressCodec in order to decode addresses
* (x/slashing) [#17098](https://github.com/cosmos/cosmos-sdk/pull/17098) `NewMsgUnjail` takes a string instead of `sdk.ValAddress`
* (x/genutil) [#17098](https://github.com/cosmos/cosmos-sdk/pull/17098) `GenAppStateFromConfig`, AddGenesisAccountCmd and `GenTxCmd` takes an addresscodec to decode addresses
* (x/distribution) [#17098](https://github.com/cosmos/cosmos-sdk/pull/17098) `NewMsgDepositValidatorRewardsPool`, `NewMsgFundCommunityPool`, `NewMsgWithdrawValidatorCommission` and `NewMsgWithdrawDelegatorReward` takes a string instead of `sdk.ValAddress` or `sdk.AccAddress`
* (client) [#17215](https://github.com/cosmos/cosmos-sdk/pull/17215) `server.StartCmd`,`server.ExportCmd`,`server.NewRollbackCmd`,`pruning.Cmd`,`genutilcli.InitCmd`,`genutilcli.GenTxCmd`,`genutilcli.CollectGenTxsCmd`,`genutilcli.AddGenesisAccountCmd`, `genutilcli.GenesisCoreCommand` do not take a home directory anymore. It is inferred from the root command.

### CLI Breaking Changes

* (server) [#17177](https://github.com/cosmos/cosmos-sdk/pull/17177) Remove `iavl-lazy-loading` configuration.
* (rosetta) [#16276](https://github.com/cosmos/cosmos-sdk/issues/16276) Rosetta migration to standalone repo.
* (cli) [#17184](https://github.com/cosmos/cosmos-sdk/pull/17184) All json keys returned by the `status` command are now snake case instead of pascal case.

### State Machine Breaking

* (x/distribution) [#17115](https://github.com/cosmos/cosmos-sdk/pull/17115) Migrate `PreviousProposer` to collections.

## [v0.50.0-beta.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.0-beta.0) - 2023-07-19

### Features

* (codec) [#17042](https://github.com/cosmos/cosmos-sdk/pull/17042) Add `CollValueV2` which supports encoding of protov2 messages in collections.
* (baseapp) [#16898](https://github.com/cosmos/cosmos-sdk/pull/16898) Add `preFinalizeBlockHook` to allow vote extensions persistence.
* (cli) [#16887](https://github.com/cosmos/cosmos-sdk/pull/16887) Add two new CLI commands: `<appd> tx simulate` for simulating a transaction; `<appd> query block-results` for querying CometBFT RPC for block results.
* (x/gov) [#16976](https://github.com/cosmos/cosmos-sdk/pull/16976) Add `failed_reason` field to `Proposal` under `x/gov` to indicate the reason for a failed proposal. Referenced from [#238](https://github.com/bnb-chain/greenfield-cosmos-sdk/pull/238) under `bnb-chain/greenfield-cosmos-sdk`.
* [#15970](https://github.com/cosmos/cosmos-sdk/pull/15970) Enable SIGN_MODE_TEXTUAL.
* (types) [#15958](https://github.com/cosmos/cosmos-sdk/pull/15958) Add `module.NewBasicManagerFromManager` for creating a basic module manager from a module manager.
* (runtime) [#15818](https://github.com/cosmos/cosmos-sdk/pull/15818) Provide logger through `depinject` instead of appBuilder.
* (client) [#15597](https://github.com/cosmos/cosmos-sdk/pull/15597) Add status endpoint for clients.
* (testutil/integration) [#15556](https://github.com/cosmos/cosmos-sdk/pull/15556) Introduce `testutil/integration` package for module integration testing.
* (types) [#15735](https://github.com/cosmos/cosmos-sdk/pull/15735) Make `ValidateBasic() error` method of `Msg` interface optional. Modules should validate messages directly in their message handlers ([RFC 001](https://docs.cosmos.network/main/rfc/rfc-001-tx-validation)).
* (x/genutil) [#15679](https://github.com/cosmos/cosmos-sdk/pull/15679) Allow applications to specify a custom genesis migration function for the `genesis migrate` command.
* (client) [#15458](https://github.com/cosmos/cosmos-sdk/pull/15458) Add a `CmdContext` field to client.Context initialized to cobra command's context.
* (core) [#15133](https://github.com/cosmos/cosmos-sdk/pull/15133) Implement RegisterServices in the module manager.
* (x/gov) [#14373](https://github.com/cosmos/cosmos-sdk/pull/14373) Add new proto field `constitution` of type `string` to gov module genesis state, which allows chain builders to lay a strong foundation by specifying purpose.
* (x/genutil) [#15301](https://github.com/cosmos/cosmos-sdk/pull/15031) Add application genesis. The genesis is now entirely managed by the application and passed to CometBFT at note instantiation. Functions that were taking a `cmttypes.GenesisDoc{}` now takes a `genutiltypes.AppGenesis{}`.
* (cli) [#14659](https://github.com/cosmos/cosmos-sdk/pull/14659) Added ability to query blocks by events with queries directly passed to Tendermint, which will allow for full query operator support, e.g. `>`.
* (x/gov) [#14720](https://github.com/cosmos/cosmos-sdk/pull/14720) Upstream expedited proposals from Osmosis.
* (x/auth) [#14650](https://github.com/cosmos/cosmos-sdk/pull/14650) Add Textual SignModeHandler. It is however **NOT** enabled by default, and should only be used for **TESTING** purposes until `SIGN_MODE_TEXTUAL` is fully released.
* (x/crisis) [#14588](https://github.com/cosmos/cosmos-sdk/pull/14588) Use CacheContext() in AssertInvariants().
* (client) [#14342](https://github.com/cosmos/cosmos-sdk/pull/14342) Add `<app> config` command is now a sub-command, for setting, getting and migrating Cosmos SDK configuration files.
* (query) [#14468](https://github.com/cosmos/cosmos-sdk/pull/14468) Implement pagination for collections.
* (x/distribution) [#14322](https://github.com/cosmos/cosmos-sdk/pull/14322) Introduce a new gRPC message handler, `DepositValidatorRewardsPool`, that allows explicit funding of a validator's reward pool.
* [#13473](https://github.com/cosmos/cosmos-sdk/pull/13473) ADR-038: Go plugin system proposal
* (mempool) [#14484](https://github.com/cosmos/cosmos-sdk/pull/14484) Add priority nonce mempool option for transaction replacement.
* (x/bank) [#14894](https://github.com/cosmos/cosmos-sdk/pull/14894) Return a human readable denomination for IBC vouchers when querying bank balances. Added a `ResolveDenom` parameter to `types.QueryAllBalancesRequest` and `--resolve-denom` flag to `GetBalancesCmd()`.
* (runtime) [#15547](https://github.com/cosmos/cosmos-sdk/pull/15547) Allow runtime to pass event core api service to modules
* (telemetry) [#15657](https://github.com/cosmos/cosmos-sdk/pull/15657) Emit more data (go version, sdk version, upgrade height) in prom metrics
* (types/module) [#15829](https://github.com/cosmos/cosmos-sdk/pull/15829) Add new endblocker interface to handle valset updates.
* (core) [#14860](https://github.com/cosmos/cosmos-sdk/pull/14860) Add `Precommit` and `PrepareCheckState` AppModule callbacks.
* (types/simulation) [#16074](https://github.com/cosmos/cosmos-sdk/pull/16074) Add generic SimulationStoreDecoder for modules using collections.
* (cli) [#16209](https://github.com/cosmos/cosmos-sdk/pull/16209) Make `StartCmd` more customizable.
* (types) [#16257](https://github.com/cosmos/cosmos-sdk/pull/16257) Allow setting the base denom in the denom registry.
* (genutil) [#16046](https://github.com/cosmos/cosmos-sdk/pull/16046) Add "module-name" flag to genutil `add-genesis-account` to enable intializing module accounts at genesis.

### Improvements

* (all modules) [#15901](https://github.com/cosmos/cosmos-sdk/issues/15901) All core Cosmos SDK modules query commands have migrated to [AutoCLI](https://docs.cosmos.network/main/building-modules/autocli), ensuring parity between gRPC and CLI queries.
* (types) [#16890](https://github.com/cosmos/cosmos-sdk/pull/16890) Remove `GetTxCmd() *cobra.Command` and `GetQueryCmd() *cobra.Command` from `module.AppModuleBasic` interface.
* (cli) [#16856](https://github.com/cosmos/cosmos-sdk/pull/16856) Improve `simd prune` UX by using the app default home directory and set pruning method as first variable argument (defaults to default).
* (x/authz) [#16869](https://github.com/cosmos/cosmos-sdk/pull/16869) Improve error message when grant not found.
* (all) [#16497](https://github.com/cosmos/cosmos-sdk/pull/16497) Removed all exported vestiges of `sdk.MustSortJSON` and `sdk.SortJSON`.
* (cli) [#16206](https://github.com/cosmos/cosmos-sdk/pull/16206) Make ABCI handshake profileable.
* (types) [#16076](https://github.com/cosmos/cosmos-sdk/pull/16076) Optimize `ChainAnteDecorators`/`ChainPostDecorators` to instantiate the functions once instead of on every invocation of the returned `AnteHandler`/`PostHandler`.
* (server) [#16071](https://github.com/cosmos/cosmos-sdk/pull/16071) When `mempool.max-txs` is set to a negative value, use a no-op mempool (effectively disable the app mempool).
* (simapp) [#15958](https://github.com/cosmos/cosmos-sdk/pull/15958) Refactor SimApp for removing the global basic manager.
* (crypto) [#3129](https://github.com/cosmos/cosmos-sdk/pull/3129) New armor and keyring key derivation uses aead and encryption uses chacha20poly
* (x/slashing) [#15580](https://github.com/cosmos/cosmos-sdk/pull/15580) Refactor the validator's missed block signing window to be a chunked bitmap instead of a "logical" bitmap, significantly reducing the storage footprint.
* (x/gov) [#15554](https://github.com/cosmos/cosmos-sdk/pull/15554) Add proposal result log in `active_proposal` event. When a proposal passes but fails to execute, the proposal result is logged in the `active_proposal` event.
* (mempool) [#15328](https://github.com/cosmos/cosmos-sdk/pull/15328) Improve the `PriorityNonceMempool`
    * Support generic transaction prioritization, instead of `ctx.Priority()`
    * Improve construction through the use of a single `PriorityNonceMempoolConfig` instead of option functions
* (x/authz) [#15164](https://github.com/cosmos/cosmos-sdk/pull/15164) Add `MsgCancelUnbondingDelegation` to staking authorization
* (server) [#15358](https://github.com/cosmos/cosmos-sdk/pull/15358) Add `server.InterceptConfigsAndCreateContext` as alternative to `server.InterceptConfigsPreRunHandler` which does not set the server context and the default SDK logger.
* [#15011](https://github.com/cosmos/cosmos-sdk/pull/15011) Introduce `cosmossdk.io/log` package to provide a consistent logging interface through the SDK. CometBFT logger is now replaced by `cosmossdk.io/log.Logger`.
* (x/auth) [#14758](https://github.com/cosmos/cosmos-sdk/pull/14758) Allow transaction event queries to directly passed to Tendermint, which will allow for full query operator support, e.g. `>`.
* (server) [#15041](https://github.com/cosmos/cosmos-sdk/pull/15041) Remove unnecessary sleeps from gRPC and API server initiation. The servers will start and accept requests as soon as they're ready.
* (x/staking) [#14864](https://github.com/cosmos/cosmos-sdk/pull/14864) `create-validator` CLI command now takes a json file as an arg instead of having a bunch of required flags to it.
* (cli) [#14659](https://github.com/cosmos/cosmos-sdk/pull/14659) Added ability to query blocks by either height/hash `<app> q block --type=height|hash <height|hash>`.
* (store) [#14410](https://github.com/cosmos/cosmos-sdk/pull/14410) `rootmulti.Store.loadVersion` has validation to check if all the module stores' height is correct, it will error if any module store has incorrect height.
* (x/evidence) [#14757](https://github.com/cosmos/cosmos-sdk/pull/14757) Evidence messages do not need to implement a `.Type()` anymore.
* (x/auth/tx) [#14751](https://github.com/cosmos/cosmos-sdk/pull/14751) Remove `.Type()` and `Route()` methods from all msgs and `legacytx.LegacyMsg` interface.
* [#14529](https://github.com/cosmos/cosmos-sdk/pull/14529) Add new property `BondDenom` to `SimulationState` struct.
* (module) [#14415](https://github.com/cosmos/cosmos-sdk/pull/14415) Loosen assertions in SetOrderBeginBlockers() and SetOrderEndBlockers()
* (context)[#14384](https://github.com/cosmos/cosmos-sdk/pull/14384) Refactor(context): Pass EventManager to the context as an interface.
* (types) [#14354](https://github.com/cosmos/cosmos-sdk/pull/14354) Improve performance on Context.KVStore and Context.TransientStore by 40%.
* (crypto/keyring) [#14151](https://github.com/cosmos/cosmos-sdk/pull/14151) Move keys presentation from `crypto/keyring` to `client/keys`
* (signing) [#14087](https://github.com/cosmos/cosmos-sdk/pull/14087) Add SignModeHandlerWithContext interface with a new `GetSignBytesWithContext` to get the sign bytes using `context.Context` as an argument to access state.
* (server) [#14062](https://github.com/cosmos/cosmos-sdk/pull/14062) Remove rosetta from server start.
* (baseapp) [#14417](https://github.com/cosmos/cosmos-sdk/pull/14417) `SetStreamingService` accepts appOptions, AppCodec and Storekeys needed to set streamers.  
    * Store pacakge no longer has a dependency on baseapp. 
* (store) [#14438](https://github.com/cosmos/cosmos-sdk/pull/14438) Pass logger from baseapp to store. 
* (store) [#14439](https://github.com/cosmos/cosmos-sdk/pull/14439) Remove global metric gatherer from store. 
    * By default store has a no op metric gatherer, the application developer must set another metric gatherer or us the provided one in `store/metrics`.
* [#14406](https://github.com/cosmos/cosmos-sdk/issues/14406) Migrate usage of types/store.go to store/types/..
* (x/staking) [#14590](https://github.com/cosmos/cosmos-sdk/pull/14590) Return undelegate amount in MsgUndelegateResponse.
* (baseapp) [#15023](https://github.com/cosmos/cosmos-sdk/pull/15023) & [#15213](https://github.com/cosmos/cosmos-sdk/pull/15213) Add `MessageRouter` interface to baseapp and pass it to authz, gov and groups instead of concrete type. 
* (x/consensus) [#15553](https://github.com/cosmos/cosmos-sdk/pull/15553) Migrate consensus module to use collections
* (store/cachekv) [#15767](https://github.com/cosmos/cosmos-sdk/pull/15767) Reduce peak RAM usage during and after InitGenesis
* (x/bank) [#15764](https://github.com/cosmos/cosmos-sdk/pull/15764) Speedup x/bank InitGenesis
* (x/auth) [#15867](https://github.com/cosmos/cosmos-sdk/pull/15867) Support better logging for signature verification failure.
* (types/query) [#16041](https://github.com/cosmos/cosmos-sdk/pull/16041) change pagination max limit to a variable in order to be modifed by application devs
* (server) [#16238](https://github.com/cosmos/cosmos-sdk/pull/16238) Don't setup p2p node keys if starting a node in GRPC only mode.

### State Machine Breaking

* (x/group,x/gov) [#16235](https://github.com/cosmos/cosmos-sdk/pull/16235) A group and gov proposal is rejected if the proposal metadata title and summary do not match the proposal title and summary.
* (x/staking) [#15701](https://github.com/cosmos/cosmos-sdk/pull/15701) The `HistoricalInfoKey` has been updated to use a binary format.
* (x/slashing) [#15580](https://github.com/cosmos/cosmos-sdk/pull/15580) The validator slashing window now stores "chunked" bitmap entries for each validator's signing window instead of a single boolean entry per signing window index.
* (x/feegrant) [#14294](https://github.com/cosmos/cosmos-sdk/pull/14294) Moved the logic of rejecting duplicate grant from `msg_server` to `keeper` method.
* (x/staking) [#14590](https://github.com/cosmos/cosmos-sdk/pull/14590) `MsgUndelegateResponse` now includes undelegated amount. `x/staking` module's `keeper.Undelegate` now returns 3 values (completionTime,undelegateAmount,error) instead of 2.
* (x/staking) [#15731](https://github.com/cosmos/cosmos-sdk/pull/15731) Introducing a new index to retrieve the delegations by validator efficiently.
* (baseapp) [#15930](https://github.com/cosmos/cosmos-sdk/pull/15930) change vote info provided by prepare and process proposal to the one in the block 

### API Breaking Changes

* (x/staking) [#16958](https://github.com/cosmos/cosmos-sdk/pull/16958) DelegationI interface `GetDelegatorAddr` & `GetValidatorAddr` have been migrated to return string instead of sdk.AccAddress and sdk.ValAddress respectively. stakingtypes.NewDelegation takes a string instead of sdk.AccAddress and sdk.ValAddress.
* (x/staking) [#16959](https://github.com/cosmos/cosmos-sdk/pull/16959) Add validator and consensus address codec as staking keeper arguments.
* (types) [#16272](https://github.com/cosmos/cosmos-sdk/pull/16272) `FeeGranter` in the `FeeTx` interface returns `[]byte` instead of `string`. 
* (testutil) [#16899](https://github.com/cosmos/cosmos-sdk/pull/16899) The *cli testutil* `QueryBalancesExec` has been removed. Use the gRPC or REST query instead.
* (x/auth) [#16650](https://github.com/cosmos/cosmos-sdk/pull/16650) The *cli testutil* `QueryAccountExec` has been removed. Use the gRPC or REST query instead.
* (types/math) [#16040](https://github.com/cosmos/cosmos-sdk/pull/16798) Remove aliases in `types/math.go` (part 2).
* (x/staking) [#16795](https://github.com/cosmos/cosmos-sdk/pull/16795) `DelegationToDelegationResponse`, `DelegationsToDelegationResponses`, `RedelegationsToRedelegationResponses` are no longer exported.
* (x/staking) [#16324](https://github.com/cosmos/cosmos-sdk/pull/16324) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`, and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context` and return an `error`. Notable changes:
    * `Validator` method now returns `types.ErrNoValidatorFound` instead of `nil` when not found.
* (x/auth) [#16621](https://github.com/cosmos/cosmos-sdk/pull/16621) Pass address codec to auth new keeper constructor.
* (x/auth/vesting) [#16741](https://github.com/cosmos/cosmos-sdk/pull/16741) Vesting account constructor now return an error with the result of their validate function.
* (baseapp) [#15568](https://github.com/cosmos/cosmos-sdk/pull/15568) `SetIAVLLazyLoading` is removed from baseapp.
* (x/slashing) [#16246](https://github.com/cosmos/cosmos-sdk/issues/16246) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`, and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context` and return an `error`. `GetValidatorSigningInfo` now returns an error instead of a `found bool`, the error can be `nil` (found), `ErrNoSigningInfoFound` (not found) and any other error.
* (module) [#16227](https://github.com/cosmos/cosmos-sdk/issues/16227) `manager.RunMigrations()` now take a `context.Context` instead of a `sdk.Context`.
* (x/mint) [#16179](https://github.com/cosmos/cosmos-sdk/issues/16179) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`, and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context` and return an `error`.
* (x/crisis) [#16216](https://github.com/cosmos/cosmos-sdk/issues/16216) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`, methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context` and return an `error` instead of panicking.
* (x/gov) [#15988](https://github.com/cosmos/cosmos-sdk/issues/15988) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`, methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context` and return an `error` (instead of panicking or returning a `found bool`). Iterators callback functions now return an error instead of a `bool`.
* (x/auth) [#15985](https://github.com/cosmos/cosmos-sdk/pull/15985) The `AccountKeeper` does not expose the `QueryServer` and `MsgServer` APIs anymore.
* (x/authz) [#15962](https://github.com/cosmos/cosmos-sdk/issues/15962) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`, methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context`. The `Authorization` interface's `Accept` method now takes a `context.Context` instead of a `sdk.Context`.
* (x/distribution) [#15948](https://github.com/cosmos/cosmos-sdk/issues/15948) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey` and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context`. Keeper methods also now return an `error`.
* (x/bank) [#15891](https://github.com/cosmos/cosmos-sdk/issues/15891) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey` and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context`. Also `FundAccount` and `FundModuleAccount` from the `testutil` package accept a `context.Context` instead of a `sdk.Context`, and it's position was moved to the first place.
* (x/bank) [#15818](https://github.com/cosmos/cosmos-sdk/issues/15818) `BaseViewKeeper`'s `Logger` method now doesn't require a context. `NewBaseKeeper`, `NewBaseSendKeeper` and `NewBaseViewKeeper` now also require a `log.Logger` to be passed in.
* (client) [#15597](https://github.com/cosmos/cosmos-sdk/pull/15597) `RegisterNodeService` now requires a config parameter.
* (x/*all*) [#15648](https://github.com/cosmos/cosmos-sdk/issues/15648) Make `SetParams` consistent across all modules and validate the params at the message handling instead of `SetParams` method.
* (x/genutil) [#15679](https://github.com/cosmos/cosmos-sdk/pull/15679) `MigrateGenesisCmd` now takes a `MigrationMap` instead of having the SDK genesis migration hardcoded.
* (client) [#15673](https://github.com/cosmos/cosmos-sdk/pull/15673) Move `client/keys.OutputFormatJSON` and `client/keys.OutputFormatText` to `client/flags` package.
* (x/nft) [#15588](https://github.com/cosmos/cosmos-sdk/pull/15588) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey` and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context`. 
* (x/auth) [#15520](https://github.com/cosmos/cosmos-sdk/pull/15520) `NewAccountKeeper` now takes a `KVStoreService` instead of a `StoreKey` and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context`. 
* (x/consensus) [#15517](https://github.com/cosmos/cosmos-sdk/pull/15517) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey`.
* (x/bank) [#15477](https://github.com/cosmos/cosmos-sdk/pull/15477) `banktypes.NewMsgMultiSend` and `keeper.InputOutputCoins` only accept one input.
* (mempool) [#15328](https://github.com/cosmos/cosmos-sdk/pull/15328) The `PriorityNonceMempool` is now generic over type `C comparable` and takes a single `PriorityNonceMempoolConfig[C]` argument. See `DefaultPriorityNonceMempoolConfig` for how to construct the configuration and a `TxPriority` type.
* (server) [#15358](https://github.com/cosmos/cosmos-sdk/pull/15358) Remove `server.ErrorCode` that was not used anywhere.
* [#15211](https://github.com/cosmos/cosmos-sdk/pull/15211) Remove usage of `github.com/cometbft/cometbft/libs/bytes.HexBytes` in favor of `[]byte` thorough the SDK.
* [#15011](https://github.com/cosmos/cosmos-sdk/pull/15011) All functions that were taking a CometBFT logger, now take `cosmossdk.io/log.Logger` instead.
* (x/auth) [#14758](https://github.com/cosmos/cosmos-sdk/pull/14758) Refactor transaction searching:
    * Refactor `QueryTxsByEvents` to accept a `query` of type `string` instead of `events` of type `[]string`
    * Pass `prove=false` to Tendermint's `TxSearch` RPC method
    * Refactor CLI methods to accept `--query` flag instead of `--events`
* (server) [#15041](https://github.com/cosmos/cosmos-sdk/pull/15041) Refactor how gRPC and API servers are started to remove unnecessary sleeps:
    * Remove `ServerStartTime` constant.
    * Rename `WaitForQuitSignals` to `ListenForQuitSignals`. Note, this function is no longer blocking. Thus the caller is expected to provide a `context.CancelFunc` which indicates that when a signal is caught, that any spawned processes can gracefully exit.
    * `api.Server#Start` now accepts a `context.Context`. The caller is responsible for ensuring that the context is canceled such that the API server can gracefully exit. The caller does not need to stop the server.
    * To start the gRPC server you must first create the server via `NewGRPCServer`, after which you can start the gRPC server via `StartGRPCServer` which accepts a `context.Context`. The caller is responsible for ensuring that the context is canceled such that the gRPC server can gracefully exit. The caller does not need to stop the server.
* (types) [#15067](https://github.com/cosmos/cosmos-sdk/pull/15067) Remove deprecated alias from `types/errors`. Use `cosmossdk.io/errors` instead.
* (simapp) [#14977](https://github.com/cosmos/cosmos-sdk/pull/14977) Move simulation helpers functions (`AppStateFn` and `AppStateRandomizedFn`) to `testutil/sims`. These takes an extra genesisState argument which is the default state of the app.
* (x/gov) [#14720](https://github.com/cosmos/cosmos-sdk/pull/14720) Add an expedited field in the gov v1 proposal and `MsgNewMsgProposal`.
* [#14847](https://github.com/cosmos/cosmos-sdk/pull/14847) App and ModuleManager methods `InitGenesis`, `ExportGenesis`, `BeginBlock` and `EndBlock` now also return an error.
* (x/upgrade) [#14764](https://github.com/cosmos/cosmos-sdk/pull/14764) The `x/upgrade` module is extracted to have a separate go.mod file which allows it to be a standalone module. 
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
* (store) [#11825](https://github.com/cosmos/cosmos-sdk/pull/11825) Make extension snapshotter interface safer to use, renamed the util function `WriteExtensionItem` to `WriteExtensionPayload`.
* (signing) [#13701](https://github.com/cosmos/cosmos-sdk/pull/) Add `context.Context` as an argument `x/auth/signing.VerifySignature`.
* (snapshots) [14048](https://github.com/cosmos/cosmos-sdk/pull/14048) Move the Snapshot package to the store package. This is done in an effort group all storage related logic under one package.
* (baseapp) [#14050](https://github.com/cosmos/cosmos-sdk/pull/14050) Refactor `ABCIListener` interface to accept Go contexts.
* (store/streaming)[#14603](https://github.com/cosmos/cosmos-sdk/pull/14603) `StoreDecoderRegistry` moved from store to `types/simulations` this breaks the `AppModuleSimulation` interface. 
* (x/staking) [#14590](https://github.com/cosmos/cosmos-sdk/pull/14590) `MsgUndelegateResponse` now includes undelegated amount. `x/staking` module's `keeper.Undelegate` now returns 3 values (completionTime,undelegateAmount,error)  instead of 2.
* (x/feegrant) [#14649](https://github.com/cosmos/cosmos-sdk/pull/14649) Extract Feegrant in its own go.mod and rename the package to `cosmossdk.io/x/feegrant`.
* (x/bank) [#14894](https://github.com/cosmos/cosmos-sdk/pull/14894) Allow a human readable denomination for coins when querying bank balances. Added a `ResolveDenom` parameter to `types.QueryAllBalancesRequest`.
* (crypto) [#15070](https://github.com/cosmos/cosmos-sdk/pull/15070) `GenerateFromPassword` and `Cost` from `bcrypt.go` now take a `uint32` instead of a `int` type.  
* (x/capability) [#15344](https://github.com/cosmos/cosmos-sdk/pull/15344) Capability module was removed and is now housed in [IBC-GO](https://github.com/cosmos/ibc-go). 
* [#15299](https://github.com/cosmos/cosmos-sdk/pull/15299) Remove `StdTx` transaction and signing APIs. No SDK version has actually supported `StdTx` since before Stargate.
* (codec) [#15600](https://github.com/cosmos/cosmos-sdk/pull/15600) [#15873](https://github.com/cosmos/cosmos-sdk/pull/15873) add support for getting signers to `codec.Codec` and `InterfaceRegistry`:
    * `Codec` has new methods `InterfaceRegistry`, `GetMsgAnySigners`, `GetMsgV1Signers`, and `GetMsgV2Signers` as well as unexported methods. All implementations of `Codec` by other users must now embed an official implementation from the `codec` package.
    * `InterfaceRegistry` is has unexported methods and implements `protodesc.Resolver` plus the `RangeFiles` and `SigningContext` methods. All implementations of `InterfaceRegistry` by other users must now embed the official implementation.
    * `AminoCodec` is marked as deprecated and no longer implements `Codec.
* (x/crisis) [#15852](https://github.com/cosmos/cosmos-sdk/pull/15852) Crisis keeper now takes a instance of the address codec to be able to decode user addresses
* (x/slashing) [#15875](https://github.com/cosmos/cosmos-sdk/pull/15875) `x/slashing.NewAppModule` now requires an `InterfaceRegistry` parameter.
* (client) [#15822](https://github.com/cosmos/cosmos-sdk/pull/15822) The return type of the interface method `TxConfig.SignModeHandler` has been changed to `x/tx/signing.HandlerMap`.
* (x/auth) [#15822](https://github.com/cosmos/cosmos-sdk/pull/15822) The type of struct field `ante.HandlerOptions.SignModeHandler` has been changed to `x/tx/signing.HandlerMap`.
    * The signature of `NewSigVerificationDecorator` has been changed to accept a `x/tx/signing.HandlerMap`.
    * The signature of `VerifySignature` has been changed to accept a `x/tx/signing.HandlerMap` and other structs from `x/tx` as arguments.
    * The signature of `NewTxConfigWithTextual` has been deprecated and its signature changed to accept a `SignModeOptions`.
* (x/bank) [#15567](https://github.com/cosmos/cosmos-sdk/pull/15567) `GenesisBalance.GetAddress` now returns a string instead of `sdk.AccAddress`
    * `MsgSendExec` test helper function now takes a address.Codec 
* (x/genutil) [#15567](https://github.com/cosmos/cosmos-sdk/pull/15567) `CollectGenTxsCmd` & `GenTxCmd` takes a address.Codec to be able to decode addresses
* (x/genutil) [#15999](https://github.com/cosmos/cosmos-sdk/pull/15999) Genutil now takes the `GenesisTxHanlder` interface instead of deliverTx. The interface is implemented on baseapp
* (types/math) [#16040](https://github.com/cosmos/cosmos-sdk/pull/16040) Remove aliases in `types/math.go` (part 1).
* (x/gov) [#16106](https://github.com/cosmos/cosmos-sdk/pull/16106) Remove gRPC query methods from Keeper.
* (x/gov) [#16118](https://github.com/cosmos/cosmos-sdk/pull/16118/) Use collections for constituion and params state management.
* (x/gov) [#16127](https://github.com/cosmos/cosmos-sdk/pull/16127) Use collections for deposit state management:
    * The following methods are removed from the gov keeper: `GetDeposit`, `GetAllDeposits`, `IterateAllDeposits`.
    * The following functions are removed from the gov types: `DepositKey`, `DepositsKey`.
* (x/gov) [#16164](https://github.com/cosmos/cosmos-sdk/pull/16164) Use collections for vote state management:
    * Removed: types `VoteKey`, `VoteKeys`
    * Removed: keeper `IterateVotes`, `IterateAllVotes`, `GetVotes`, `GetVote`, `SetVote`
* (x/gov) [#16171](https://github.com/cosmos/cosmos-sdk/pull/16171) Use collections for proposal state management (part 1):
    * Removed: keeper: `GetProposal`, `UnmarshalProposal`, `MarshalProposal`, `IterateProposal`, `GetProposal`, `GetProposalFiltered`, `GetProposals`, `GetProposalID`, `SetProposalID`
    * Removed: errors unused errors
* (sims) [#16155](https://github.com/cosmos/cosmos-sdk/pull/16155) 
    * `simulation.NewOperationMsg` now marshals the operation msg as proto bytes instead of legacy amino JSON bytes.
    * `simulation.NewOperationMsg` is now 2-arity instead of 3-arity with the obsolete argument `codec.ProtoCodec` removed.
    * The field `OperationMsg.Msg` is now of type `[]byte` instead of `json.RawMessage`.
* (cli) [#16209](https://github.com/cosmos/cosmos-sdk/pull/16209) Add API `StartCmdWithOptions` to create customized start command.
* (x/auth) [#16016](https://github.com/cosmos/cosmos-sdk/pull/16016) Use collections for accounts state management:
    * removed: keeper `HasAccountByID`, `AccountAddressByID`, `SetParams
* (x/distribution) [#16211](https://github.com/cosmos/cosmos-sdk/pull/16211) Use collections for params state management.
* [#15284](https://github.com/cosmos/cosmos-sdk/pull/15284)
    * `sdk.Msg.GetSigners` was deprecated and is no longer supported. Use the `cosmos.msg.v1.signer` protobuf annotation instead.
    * `sdk.Tx` now requires a new method `GetMsgsV2()`.
    * `types/tx.Tx` no longer implements `sdk.Tx`.
    * `TxConfig` has a new method `SigningContext() *signing.Context`.
    * `AccountKeeper` now has an `AddressCodec() address.Codec` method and the expected `AccountKeeper` for `x/auth/ante` expects this method.
    * `SigVerifiableTx.GetSigners()` now returns `([][]byte, error)` instead of `[]sdk.AccAddress`.
* (x/authx) [#15284](https://github.com/cosmos/cosmos-sdk/pull/15284) `NewKeeper` now requires `codec.Codec`.
* (x/gov) [#15284](https://github.com/cosmos/cosmos-sdk/pull/15284) `NewKeeper` now requires `codec.Codec`.
* (x/distribution) [#16302](https://github.com/cosmos/cosmos-sdk/pull/16302) Use collections for FeePool state management.
    * Removed: keeper `GetFeePool`, `SetFeePool`, `GetFeePoolCommunityCoins`
* (x/gov) [#16268](https://github.com/cosmos/cosmos-sdk/pull/16268) Use collections for proposal state management (part 2):
    * this finalizes the gov collections migration
    * Removed: keeper `InsertActiveProposalsQueue`, `RemoveActiveProposalsQueue`, `InsertInactiveProposalsQueue`, `RemoveInactiveProposalsQueue`, `IterateInactiveProposalsQueue`, `IterateActiveProposalsQueue`, `ActiveProposalsQueueIterator`, `InactiveProposalsQueueIterator`
    * Removed: types all the key related functions
* (baseapp) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) BeginBlock and EndBlock are now internal to baseapp. For testing, user must call `FinalizeBlock`. BeginBlock and EndBlock calls are internal to Baseapp. 
* (baseapp) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) Writing of state to the multistore was moved to FinalizeBlock. Commit still handles the commiting values to disk. 
* (baseapp) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) `runTxMode`s were renamed to `execMode`. ModeDeliver as changed to `ModeFinalize` and a new `ModeVoteExtension` was added for vote extensions.
* (baseapp) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) All calls to ABCI methods now accept a pointer of the abci request and response types
* (baseapp) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) Calls to BeginBlock and EndBlock have been replaced with core api beginblock & endblock. 
* (x/crisis) [#16328](https://github.com/cosmos/cosmos-sdk/pull/16328) Use collections for state management:
    * Removed: keeper `GetConstantFee`, `SetConstantFee` 
* (x/mint) [#16329](https://github.com/cosmos/cosmos-sdk/pull/16329) Use collections for state management:
    * Removed: keeper `GetParams`, `SetParams`, `GetMinter`, `SetMinter`.
* (x/*all*) [#16052](https://github.com/cosmos/cosmos-sdk/pull/16062) `GetSignBytes` implementations on messages and global legacy amino codec definitions have been removed from all modules.
* (sims) [#16052](https://github.com/cosmos/cosmos-sdk/pull/16062) `GetOrGenerate` no longer requires a codec argument is now 4-arity instead of 5-arity.
* (baseapp) [#16342](https://github.com/cosmos/cosmos-sdk/pull/16342) NewContext was renamed to NewContextLegacy. The replacement (NewContext) now does not take a header, instead you should set the header via `WithHeaderInfo` or `WithBlockHeight`. Note that `WithBlockHeight` will soon be depreacted and its recommneded to use `WithHeaderInfo`.
* (x/auth) [#16423](https://github.com/cosmos/cosmos-sdk/pull/16423) `helpers.AddGenesisAccount` has been moved to `x/genutil` to remove the cyclic dependency between `x/auth` and `x/genutil`.

### Client Breaking Changes

* (x/staking) [#15701](https://github.com/cosmos/cosmos-sdk/pull/15701) `HistoricalInfoKey` now has a binary format.
* (grpc-web) [#14652](https://github.com/cosmos/cosmos-sdk/pull/14652) Use same port for gRPC-Web and the API server.
* (abci) [#15845](https://github.com/cosmos/cosmos-sdk/pull/15845) Add `msg_index` to all event attributes to associate events and messages
* (abci) [#15845](https://github.com/cosmos/cosmos-sdk/pull/15845) Remove duplicating events in `logs`
* (baseapp) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) BeginBlock & EndBlock events have begin or endblock in the events in order to identify which stage they are emitted from since they are returned to comet as FinalizeBlock events, 
* (store/streaming) [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519/files) State Streaming removed emitting of beginblock, endblock and delivertx in favour of emitting FinalizeBlock. 
* (x/bank) [#17273](https://github.com/cosmos/cosmos-sdk/pull/17273) Remove message events including `sender` attribute whose information is already present in the relevant events

### CLI Breaking Changes

* (all) The migration of modules to [AutoCLI](https://docs.cosmos.network/main/building-modules/autocli) led to no changes in UX but a [small change in CLI outputs](https://github.com/cosmos/cosmos-sdk/issues/16651) where results can be nested.
* (all) Query pagination flags have been renamed with the migration to AutoCLI:
    * `--limit` -> `--page-limit`
    * `--offset` -> `--page-offset`
    * `--count-total` -> `--page-count-total`
    * `--reverse` -> `--page-reverse`
* (x/gov) [#16987](https://github.com/cosmos/cosmos-sdk/pull/16987) In `<appd> query gov proposals` the proposal status flag have renamed from `--status` to `--proposal-status`. Additonally, that flags now uses the ENUM values: `PROPOSAL_STATUS_DEPOSIT_PERIOD`, `PROPOSAL_STATUS_VOTING_PERIOD`, `PROPOSAL_STATUS_PASSED`, `PROPOSAL_STATUS_REJECTED`, `PROPOSAL_STATUS_FAILED`.
* (x/bank) [#16899](https://github.com/cosmos/cosmos-sdk/pull/16899) With the migration to AutoCLI some bank commands have been split in two: 
    * Use `denoms-metadata` for querying all denom metadata and `denom-metadata` for querying a specific denom metadata.
    * Use `total-supply` (or `total`) for querying the total supply and `total-supply-of` for querying the supply of a specific denom. 
* (cli) [#15826](https://github.com/cosmos/cosmos-sdk/pull/15826) Remove `<appd> q account` command. Use `<appd> q auth account` instead.
* (x/staking) [#14864](https://github.com/cosmos/cosmos-sdk/pull/14864) `create-validator` CLI command now takes a json file as an arg instead of having a bunch of required flags to it.
* (cli) [#14659](https://github.com/cosmos/cosmos-sdk/pull/14659) `<app> q block <height>` is removed as it just output json. The new command allows either height/hash and is `<app> q block --type=height|hash <height|hash>`. 
* (x/gov) [#14880](https://github.com/cosmos/cosmos-sdk/pull/14880) Remove `<app> tx gov submit-legacy-proposal cancel-software-upgrade` and `software-upgrade` commands. These commands are now in the `x/upgrade` module and using gov v1. Use `tx upgrade software-upgrade` instead.
* (grpc-web) [#14652](https://github.com/cosmos/cosmos-sdk/pull/14652) Remove `grpc-web.address` flag.
* (client) [#14342](https://github.com/cosmos/cosmos-sdk/pull/14342) `<app> config` command is now a sub-command. Use `<app> config --help` to learn more.
* (cli) [#15299](https://github.com/cosmos/cosmos-sdk/pull/15299) Remove `--amino` flag from `sign` and `multi-sign` commands. Amino `StdTx` has been deprecated for a while. Amino JSON signing still works as expected. 

### Bug Fixes

* (x/bank) [#16841](https://github.com/cosmos/cosmos-sdk/pull/16841) Correctly process legacy `DenomAddressIndex` values.
* (types/query) [#16905](https://github.com/cosmos/cosmos-sdk/pull/16905) Collections Pagination now applies proper count when filtering results.
* (x/consensus) [#16713](https://github.com/cosmos/cosmos-sdk/pull/16713) Add missing ABCI param in `MsgUpdateParams`.
* [#16547](https://github.com/cosmos/cosmos-sdk/pull/16547) Ensure a transaction's gas limit cannot exceed the block gas limit.
* (baseapp) [#16613](https://github.com/cosmos/cosmos-sdk/pull/16613) Ensure each message in a transaction has a registered handler, otherwise `CheckTx` will fail.
* [#16639](https://github.com/cosmos/cosmos-sdk/pull/16639) Make sure we don't execute blocks beyond the halt height.
* (x/auth/vesting) [#16733](https://github.com/cosmos/cosmos-sdk/pull/16733) Panic on overflowing and negative EndTimes when creating a PeriodicVestingAccount.
* (baseapp) [#16700](https://github.com/cosmos/cosmos-sdk/pull/16700) Fix consensus failure in returning no response to malformed transactions.
* (baseapp) [#16596](https://github.com/cosmos/cosmos-sdk/pull/16596) Return error during `ExtendVote` and `VerifyVoteExtension` if the request height is earlier than `VoteExtensionsEnableHeight`.
* (baseapp) [#16259](https://github.com/cosmos/cosmos-sdk/pull/16259) Ensure the `Context` block height is correct after `InitChain` and prior to the second block.
* (x/staking) [#16043](https://github.com/cosmos/cosmos-sdk/pull/16043) Call `AfterUnbondingInitiated` hook for new unbonding entries only and fix `UnbondingDelegation` entries handling. This is a behavior change compared to Cosmos SDK v0.47.x, now the hook is called only for new unbonding entries.
* (types) [#16010](https://github.com/cosmos/cosmos-sdk/pull/16010) Let `module.CoreAppModuleBasicAdaptor` fallback to legacy genesis handling.
* (types) [#15691](https://github.com/cosmos/cosmos-sdk/pull/15691) Make `Coin.Validate()` check that `.Amount` is not nil.
* (x/auth) [#15059](https://github.com/cosmos/cosmos-sdk/pull/15059) `ante.CountSubKeys` returns 0 when passing a nil `Pubkey`.
* (x/capability) [#15030](https://github.com/cosmos/cosmos-sdk/pull/15030) Prevent `x/capability` from consuming `GasMeter` gas during `InitMemStore`
* (types/coin) [#14739](https://github.com/cosmos/cosmos-sdk/pull/14739) Deprecate the method `Coin.IsEqual` in favour of  `Coin.Equal`. The difference between the two methods is that the first one results in a panic when denoms are not equal. This panic lead to unexpected behavior.
* (x/crypto) [#15258](https://github.com/cosmos/cosmos-sdk/pull/15258) Write keyhash file with permissions 0600 instead of 0555.
* (cli) [#16138](https://github.com/cosmos/cosmos-sdk/pull/16138) Fix snapshot commands panic if snapshot don't exists.
* (x/gov) [#16231](https://github.com/cosmos/cosmos-sdk/pull/16231) Fix Rawlog JSON formatting of proposal_vote option field.

### Deprecated

* (types) [#16980](https://github.com/cosmos/cosmos-sdk/pull/16980) Deprecate `IntProto` and `DecProto`. Instead, `math.Int` and `math.LegacyDec` should be used respectively. Both types support `Marshal` and `Unmarshal` for binary serialization.
* (x/staking) [#14567](https://github.com/cosmos/cosmos-sdk/pull/14567) The `delegator_address` field of `MsgCreateValidator` has been deprecated.
   The validator address bytes and delegator address bytes refer to the same account while creating validator (defer only in bech32 notation).

## [v0.47.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.4) - 2023-07-17

### Features

* (sims) [#16656](https://github.com/cosmos/cosmos-sdk/pull/16656) Add custom max gas for block for sim config with unlimited as default.

### Improvements

* (cli) [#16856](https://github.com/cosmos/cosmos-sdk/pull/16856) Improve `simd prune` UX by using the app default home directory and set pruning method as first variable argument (defaults to default). `pruning.PruningCmd` rest unchanged for API compability, use `pruning.Cmd` instead.
* (testutil) [#16704](https://github.com/cosmos/cosmos-sdk/pull/16704) Make app config configurator for testing configurable with external modules.
* (deps) [#16565](https://github.com/cosmos/cosmos-sdk/pull/16565) Bump CometBFT to [v0.37.2](https://github.com/cometbft/cometbft/blob/v0.37.2/CHANGELOG.md).

### Bug Fixes

* (x/auth) [#16994](https://github.com/cosmos/cosmos-sdk/pull/16994) Fix regression where querying transactions events with `<=` or `>=` would not work.
* (server) [#16827](https://github.com/cosmos/cosmos-sdk/pull/16827) Properly use `--trace` flag (before it was setting the trace level instead of displaying the stacktraces).
* (x/auth) [#16554](https://github.com/cosmos/cosmos-sdk/pull/16554) `ModuleAccount.Validate` now reports a nil `.BaseAccount` instead of panicking.
* [#16588](https://github.com/cosmos/cosmos-sdk/pull/16588) Propogate snapshotter failures to the caller, (it would create an empty snapshot silently before).
* (x/slashing) [#16784](https://github.com/cosmos/cosmos-sdk/pull/16784) Emit event with the correct reason in `SlashWithInfractionReason`.

## [v0.47.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.3) - 2023-06-08

### Features

* (baseapp) [#16290](https://github.com/cosmos/cosmos-sdk/pull/16290) Add circuit breaker setter in baseapp.
* (x/group) [#16191](https://github.com/cosmos/cosmos-sdk/pull/16191) Add EventProposalPruned event to group module whenever a proposal is pruned.
* (tx) [#15992](https://github.com/cosmos/cosmos-sdk/pull/15992) Add `WithExtensionOptions` in tx Factory to allow `SetExtensionOptions` with given extension options.

### Improvements

* (baseapp) [#16407](https://github.com/cosmos/cosmos-sdk/pull/16407) Make `DefaultProposalHandler.ProcessProposalHandler` return a ProcessProposal NoOp when using none or a NoOp mempool.
* (deps) [#16083](https://github.com/cosmos/cosmos-sdk/pull/16083) Bumps `proto-builder` image to 0.13.0.
* (client) [#16075](https://github.com/cosmos/cosmos-sdk/pull/16075) Partly revert [#15953](https://github.com/cosmos/cosmos-sdk/issues/15953) and `factory.Prepare` now does nothing in offline mode.
* (server) [#15984](https://github.com/cosmos/cosmos-sdk/pull/15984) Use `cosmossdk.io/log` package for logging instead of CometBFT logger. NOTE: v0.45 and v0.46 were not using CometBFT logger either. This keeps the same underlying logger (zerolog) as in v0.45.x+ and v0.46.x+ but now properly supporting filtered logging.
* (gov) [#15979](https://github.com/cosmos/cosmos-sdk/pull/15979) Improve gov error message when failing to convert v1 proposal to v1beta1.
* (store) [#16067](https://github.com/cosmos/cosmos-sdk/pull/16067) Add local snapshots management commands.
* (server) [#16061](https://github.com/cosmos/cosmos-sdk/pull/16061) Add Comet bootstrap command.
* (snapshots) [#16060](https://github.com/cosmos/cosmos-sdk/pull/16060) Support saving and restoring snapshot locally.
* (x/staking) [#16068](https://github.com/cosmos/cosmos-sdk/pull/16068) Update simulation to allow non-EOA accounts to stake.
* (server) [#16142](https://github.com/cosmos/cosmos-sdk/pull/16142) Remove JSON Indentation from the GRPC to REST gateway's responses. (Saving bandwidth)
* (types) [#16145](https://github.com/cosmos/cosmos-sdk/pull/16145) Rename interface `ExtensionOptionI` back to `TxExtensionOptionI` to avoid breaking change.
* (baseapp) [#16193](https://github.com/cosmos/cosmos-sdk/pull/16193) Add `Close` method to `BaseApp` for custom app to cleanup resource in graceful shutdown.

### Bug Fixes

* Fix [barberry](https://forum.cosmos.network/t/cosmos-sdk-security-advisory-barberry/10825) security vulnerability.
* (server) [#16395](https://github.com/cosmos/cosmos-sdk/pull/16395) Do not override some Comet config is purposely set differently in `InterceptConfigsPreRunHandler`.
* (store) [#16449](https://github.com/cosmos/cosmos-sdk/pull/16449) Fix StateSync Restore by excluding memory store.
* (cli) [#16312](https://github.com/cosmos/cosmos-sdk/pull/16312) Allow any addresses in `client.ValidatePromptAddress`.
* (x/group) [#16017](https://github.com/cosmos/cosmos-sdk/pull/16017) Correctly apply account number in group v2 migration.

### API Breaking Changes

* (testutil) [#14991](https://github.com/cosmos/cosmos-sdk/pull/14991) The `testutil/testdata_pulsar` package has moved to `testutil/testdata/testpb`.  Chains will not notice this breaking change as this package contains testing utilities only used by the SDK internally.

## [v0.47.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.2) - 2023-04-27

### Improvements

* (x/evidence) [#15908](https://github.com/cosmos/cosmos-sdk/pull/15908) Update the equivocation handler to work with ICS by removing a pubkey check that was performing a no-op for consumer chains.
* (x/slashing) [#15908](https://github.com/cosmos/cosmos-sdk/pull/15908) Remove the validators' pubkey check in the signature handler in order to work with ICS.
* (deps) [#15957](https://github.com/cosmos/cosmos-sdk/pull/15957) Bump CometBFT to [v0.37.1](https://github.com/cometbft/cometbft/blob/v0.37.1/CHANGELOG.md#v0371).
* (store) [#15683](https://github.com/cosmos/cosmos-sdk/pull/15683) `rootmulti.Store.CacheMultiStoreWithVersion` now can handle loading archival states that don't persist any of the module stores the current state has.
* [#15448](https://github.com/cosmos/cosmos-sdk/pull/15448) Automatically populate the block timestamp for historical queries. In contexts where the block timestamp is needed for previous states, the timestamp will now be set. Note, when querying against a node it must be re-synced in order to be able to automatically populate the block timestamp. Otherwise, the block timestamp will be populated for heights going forward once upgraded.
* [#14019](https://github.com/cosmos/cosmos-sdk/issues/14019) Remove the interface casting to allow other implementations of a `CommitMultiStore`.
* (simtestutil) [#15903](https://github.com/cosmos/cosmos-sdk/pull/15903) Add `AppStateFnWithExtendedCbs` with moduleStateCb callback function to allow access moduleState.

### Bug Fixes

* (baseapp) [#15789](https://github.com/cosmos/cosmos-sdk/pull/15789) Ensure `PrepareProposal` and `ProcessProposal` respect `InitialHeight` set by CometBFT when set to a value greater than 1.
* (types) [#15433](https://github.com/cosmos/cosmos-sdk/pull/15433) Allow disabling of account address caches (for printing bech32 account addresses).
* (client/keys) [#15876](https://github.com/cosmos/cosmos-sdk/pull/15876) Fix the JSON output `<appd> keys list --output json` when there are no keys.

## [v0.47.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.1) - 2023-03-23

### Features

* (x/bank) [#15265](https://github.com/cosmos/cosmos-sdk/pull/15265) Update keeper interface to include `GetAllDenomMetaData`.
* (x/groups) [#14879](https://github.com/cosmos/cosmos-sdk/pull/14879) Add `Query/Groups` query to get all the groups.
* (x/gov,cli) [#14718](https://github.com/cosmos/cosmos-sdk/pull/14718) Added `AddGovPropFlagsToCmd` and `ReadGovPropFlags` functions.
* (cli) [#14655](https://github.com/cosmos/cosmos-sdk/pull/14655) Add a new command to list supported algos.
* (x/genutil,cli) [#15147](https://github.com/cosmos/cosmos-sdk/pull/15147) Add `--initial-height` flag to cli init cmd to provide `genesis.json` with user-defined initial block height.

### Improvements

* (x/distribution) [#15462](https://github.com/cosmos/cosmos-sdk/pull/15462) Add delegator address to the event for withdrawing delegation rewards.
* [#14609](https://github.com/cosmos/cosmos-sdk/pull/14609) Add `RetryForBlocks` method to use in tests that require waiting for a transaction to be included in a block.

### Bug Fixes

* (baseapp) [#15487](https://github.com/cosmos/cosmos-sdk/pull/15487) Reset state before calling PrepareProposal and ProcessProposal.
* (cli) [#15123](https://github.com/cosmos/cosmos-sdk/pull/15123) Fix the CLI `offline` mode behavior to be really offline. The API of `clienttx.NewFactoryCLI` is updated to return an error. 

### Deprecated

* (x/genutil) [#15316](https://github.com/cosmos/cosmos-sdk/pull/15316) Remove requirement on node & IP being included in a gentx.

## [v0.47.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.0) - 2023-03-14

### Features

* (x/gov) [#15151](https://github.com/cosmos/cosmos-sdk/pull/15151) Add `burn_vote_quorum`, `burn_proposal_deposit_prevote` and `burn_vote_veto` params to allow applications to decide if they would like to burn deposits
* (client) [#14509](https://github.com/cosmos/cosmos-sdk/pull/#14509) Added `AddKeyringFlags` function.
* (x/bank) [#14045](https://github.com/cosmos/cosmos-sdk/pull/14045) Add CLI command `spendable-balances`, which also accepts the flag `--denom`.
* (x/slashing, x/staking) [#14363](https://github.com/cosmos/cosmos-sdk/pull/14363) Add the infraction a validator commited type as an argument to a `SlashWithInfractionReason` keeper method.
* (client) [#14051](https://github.com/cosmos/cosmos-sdk/pull/14051) Add `--grpc` client option.
* (x/genutil) [#14149](https://github.com/cosmos/cosmos-sdk/pull/14149) Add `genutilcli.GenesisCoreCommand` command, which contains all genesis-related sub-commands.
* (x/evidence) [#13740](https://github.com/cosmos/cosmos-sdk/pull/13740) Add new proto field `hash` of type `string` to `QueryEvidenceRequest` which helps to decode the hash properly while using query API.
* (core) [#13306](https://github.com/cosmos/cosmos-sdk/pull/13306) Add a `FormatCoins` function to in `core/coins` to format sdk Coins following the Value Renderers spec.
* (math) [#13306](https://github.com/cosmos/cosmos-sdk/pull/13306) Add `FormatInt` and `FormatDec` functiosn in `math` to format integers and decimals following the Value Renderers spec.
* (x/staking) [#13122](https://github.com/cosmos/cosmos-sdk/pull/13122) Add `UnbondingCanComplete` and `PutUnbondingOnHold` to `x/staking` module.
* [#13437](https://github.com/cosmos/cosmos-sdk/pull/13437) Add new flag `--modules-to-export` in `simd export` command to export only selected modules.
* [#13298](https://github.com/cosmos/cosmos-sdk/pull/13298) Add `AddGenesisAccount` helper func in x/auth module which helps adding accounts to genesis state.
* (x/authz) [#12648](https://github.com/cosmos/cosmos-sdk/pull/12648) Add an allow list, an optional list of addresses allowed to receive bank assets via authz MsgSend grant.
* (sdk.Coins) [#12627](https://github.com/cosmos/cosmos-sdk/pull/12627) Make a Denoms method on sdk.Coins.
* (testutil) [#12973](https://github.com/cosmos/cosmos-sdk/pull/12973) Add generic `testutil.RandSliceElem` function which selects a random element from the list.
* (client) [#12936](https://github.com/cosmos/cosmos-sdk/pull/12936) Add capability to preprocess transactions before broadcasting from a higher level chain.
* (cli) [#13064](https://github.com/cosmos/cosmos-sdk/pull/13064) Add `debug prefixes` to list supported HRP prefixes via .
* (ledger) [#12935](https://github.com/cosmos/cosmos-sdk/pull/12935) Generalize Ledger integration to allow for different apps or keytypes that use SECP256k1.
* (x/bank) [#11981](https://github.com/cosmos/cosmos-sdk/pull/11981) Create the `SetSendEnabled` endpoint for managing the bank's SendEnabled settings.
* (x/auth) [#13210](https://github.com/cosmos/cosmos-sdk/pull/13210) Add `Query/AccountInfo` endpoint for simplified access to basic account info.
* (x/consensus) [#12905](https://github.com/cosmos/cosmos-sdk/pull/12905) Create a new `x/consensus` module that is now responsible for maintaining Tendermint consensus parameters instead of `x/param`. Legacy types remain in order to facilitate parameter migration from the deprecated `x/params`. App developers should ensure that they execute `baseapp.MigrateParams` during their chain upgrade. These legacy types will be removed in a future release.
* (client/tx) [#13670](https://github.com/cosmos/cosmos-sdk/pull/13670) Add validation in `BuildUnsignedTx` to prevent simple inclusion of valid mnemonics

### Improvements

* [#14995](https://github.com/cosmos/cosmos-sdk/pull/14995) Allow unknown fields in `ParseTypedEvent`.
* (store) [#14931](https://github.com/cosmos/cosmos-sdk/pull/14931) Exclude in-memory KVStores, i.e. `StoreTypeMemory`, from CommitInfo commitments.
* (cli) [#14919](https://github.com/cosmos/cosmos-sdk/pull/14919) Fix never assigned error when write validators.
* (x/group) [#14923](https://github.com/cosmos/cosmos-sdk/pull/14923) Fix error while using pagination in `x/group` from CLI.
* (types/coin) [#14715](https://github.com/cosmos/cosmos-sdk/pull/14715) `sdk.Coins.Add` now returns an empty set of coins `sdk.Coins{}` if both coins set are empty.
    * This is a behavior change, as previously `sdk.Coins.Add` would return `nil` in this case.
* (reflection) [#14838](https://github.com/cosmos/cosmos-sdk/pull/14838) We now require that all proto files' import path (i.e. the OS path) matches their fully-qualified package name. For example, proto files with package name `cosmos.my.pkg.v1` should live in the folder `cosmos/my/pkg/v1/*.proto` relatively to the protoc import root folder (usually the root `proto/` folder).
* (baseapp) [#14505](https://github.com/cosmos/cosmos-sdk/pull/14505) PrepareProposal and ProcessProposal now use deliverState for the first block in order to access changes made in InitChain.
* (x/group) [#14527](https://github.com/cosmos/cosmos-sdk/pull/14527) Fix wrong address set in `EventUpdateGroupPolicy`.
* (cli) [#14509](https://github.com/cosmos/cosmos-sdk/pull/14509) Added missing options to keyring-backend flag usage.
* (server) [#14441](https://github.com/cosmos/cosmos-sdk/pull/14441) Fix `--log_format` flag not working.
* (ante) [#14448](https://github.com/cosmos/cosmos-sdk/pull/14448) Return anteEvents when postHandler fail.
* (baseapp) [#13983](https://github.com/cosmos/cosmos-sdk/pull/13983) Don't emit duplicate ante-handler events when a post-handler is defined.
* (x/staking) [#14064](https://github.com/cosmos/cosmos-sdk/pull/14064) Set all fields in `redelegation.String()`.
* (x/upgrade) [#13936](https://github.com/cosmos/cosmos-sdk/pull/13936) Make downgrade verification work again.
* (x/group) [#13742](https://github.com/cosmos/cosmos-sdk/pull/13742) Fix `validate-genesis` when group policy accounts exist.
* (store) [#13516](https://github.com/cosmos/cosmos-sdk/pull/13516) Fix state listener that was observing writes at wrong time.
* (simstestutil) [#15305](https://github.com/cosmos/cosmos-sdk/pull/15305) Add `AppStateFnWithExtendedCb` with callback function to extend rawState.
* (simapp) [#14977](https://github.com/cosmos/cosmos-sdk/pull/14977) Move simulation helpers functions (`AppStateFn` and `AppStateRandomizedFn`) to `testutil/sims`. These takes an extra genesisState argument which is the default state of the app.
* (cli) [#14953](https://github.com/cosmos/cosmos-sdk/pull/14953) Enable profiling block replay during abci handshake with `--cpu-profile`.
* (store) [#14410](https://github.com/cosmos/cosmos-sdk/pull/14410) `rootmulti.Store.loadVersion` has validation to check if all the module stores' height is correct, it will error if any module store has incorrect height.
* (store) [#14189](https://github.com/cosmos/cosmos-sdk/pull/14189) Add config `iavl-lazy-loading` to enable lazy loading of iavl store, to improve start up time of archive nodes, add method `SetLazyLoading` to `CommitMultiStore` interface.
* (deps) [#14830](https://github.com/cosmos/cosmos-sdk/pull/14830) Bump to IAVL `v0.19.5-rc.1`.
* (tools) [#14793](https://github.com/cosmos/cosmos-sdk/pull/14793) Dockerfile optimization.
* (x/gov) [#13010](https://github.com/cosmos/cosmos-sdk/pull/13010) Partial cherry-pick of this issue for adding proposer migration.
* [#14691](https://github.com/cosmos/cosmos-sdk/pull/14691) Change behavior of `sdk.StringifyEvents` to not flatten events attributes by events type.
    * This change only affects ABCI message logs, and not the events field.
* [#14692](https://github.com/cosmos/cosmos-sdk/pull/14692) Improve RPC queries error message when app is at height 0.
* [#14017](https://github.com/cosmos/cosmos-sdk/pull/14017) Simplify ADR-028 and `address.Module`.
    * This updates the [ADR-028](https://docs.cosmos.network/main/architecture/adr-028-public-key-addresses) and enhance the `address.Module` API to support module addresses and sub-module addresses in a backward compatible way.
* (snapshots) [#14608](https://github.com/cosmos/cosmos-sdk/pull/14608/) Deprecate unused structs `SnapshotKVItem` and `SnapshotSchema`.
* [#15243](https://github.com/cosmos/cosmos-sdk/pull/15243) `LatestBlockResponse` & `BlockByHeightResponse` types' field `sdk_block` was incorrectly cast `proposer_address` bytes to validator operator address, now to consensus address
* (x/group, x/gov) [#14483](https://github.com/cosmos/cosmos-sdk/pull/14483) Add support for `[]string` and `[]int` in `draft-proposal` prompt.
* (protobuf) [#14476](https://github.com/cosmos/cosmos-sdk/pull/14476) Clean up protobuf annotations `{accepts,implements}_interface`.
* (x/gov, x/group) [#14472](https://github.com/cosmos/cosmos-sdk/pull/14472) The recommended metadata format for x/gov and x/group proposals now uses an array of strings (instead of a single string) for the `authors` field.
* (crypto) [#14460](https://github.com/cosmos/cosmos-sdk/pull/14460) Check the signature returned by a ledger device against the public key in the keyring.
* [#14356](https://github.com/cosmos/cosmos-sdk/pull/14356) Add `events.GetAttributes` and `event.GetAttribute` methods to simplify the retrieval of an attribute from event(s).
* (types) [#14332](https://github.com/cosmos/cosmos-sdk/issues/14332) Reduce state export time by 50%.
* (types) [#14163](https://github.com/cosmos/cosmos-sdk/pull/14163) Refactor `(coins Coins) Validate()` to avoid unnecessary map.
* [#13881](https://github.com/cosmos/cosmos-sdk/pull/13881) Optimize iteration on nested cached KV stores and other operations in general.
* (x/gov) [#14347](https://github.com/cosmos/cosmos-sdk/pull/14347) Support `v1.Proposal` message in `v1beta1.Proposal.Content`.
* [#13882](https://github.com/cosmos/cosmos-sdk/pull/13882) Add tx `encode` and `decode` endpoints to amino tx service.
  > Note: These endpoints encodes and decodes only amino txs.
* (config) [#13894](https://github.com/cosmos/cosmos-sdk/pull/13894) Support state streaming configuration in `app.toml` template and default configuration.
* (x/nft) [#13836](https://github.com/cosmos/cosmos-sdk/pull/13836) Remove the validation for `classID` and `nftID` from the NFT module.
* [#13789](https://github.com/cosmos/cosmos-sdk/pull/13789) Add tx `encode` and `decode` endpoints to tx service.
  > Note: These endpoints will only encode and decode proto messages, Amino encoding and decoding is not supported.
* [#13619](https://github.com/cosmos/cosmos-sdk/pull/13619) Add new function called LogDeferred to report errors in defers. Use the function in x/bank files.
* (deps) [#13397](https://github.com/cosmos/cosmos-sdk/pull/13397) Bump Go version minimum requirement to `1.19`.
* [#13070](https://github.com/cosmos/cosmos-sdk/pull/13070) Migrate from `gogo/protobuf` to `cosmos/gogoproto`.
* [#12995](https://github.com/cosmos/cosmos-sdk/pull/12995) Add `FormatTime` and `ParseTimeString` methods.
* [#12952](https://github.com/cosmos/cosmos-sdk/pull/12952) Replace keyring module to Cosmos fork.
* [#12352](https://github.com/cosmos/cosmos-sdk/pull/12352) Move the `RegisterSwaggerAPI` logic into a separate helper function in the server package.
* [#12876](https://github.com/cosmos/cosmos-sdk/pull/12876) Remove proposer-based rewards.
* [#12846](https://github.com/cosmos/cosmos-sdk/pull/12846) Remove `RandomizedParams` from the `AppModuleSimulation` interface which is no longer needed.
* (ci) [#12854](https://github.com/cosmos/cosmos-sdk/pull/12854) Use ghcr.io to host the proto builder image. Update proto builder image to go 1.19
* (x/bank) [#12706](https://github.com/cosmos/cosmos-sdk/pull/12706) Added the `chain-id` flag to the `AddTxFlagsToCmd` API. There is no longer a need to explicitly register this flag on commands whens `AddTxFlagsToCmd` is already called.
* [#12717](https://github.com/cosmos/cosmos-sdk/pull/12717) Use injected encoding params in simapp.
* [#12634](https://github.com/cosmos/cosmos-sdk/pull/12634) Move `sdk.Dec` to math package.
* [#12187](https://github.com/cosmos/cosmos-sdk/pull/12187) Add batch operation for x/nft module.
* [#12455](https://github.com/cosmos/cosmos-sdk/pull/12455) Show attempts count in error for signing.
* [#13101](https://github.com/cosmos/cosmos-sdk/pull/13101) Remove weights from `simapp/params` and `testutil/sims`. They are now in their respective modules.
* [#12398](https://github.com/cosmos/cosmos-sdk/issues/12398) Refactor all `x` modules to unit-test via mocks and decouple `simapp`.
* [#13144](https://github.com/cosmos/cosmos-sdk/pull/13144) Add validator distribution info grpc gateway get endpoint.
* [#13168](https://github.com/cosmos/cosmos-sdk/pull/13168) Migrate tendermintdev/proto-builder to ghcr.io. New image `ghcr.io/cosmos/proto-builder:0.8`
* [#13178](https://github.com/cosmos/cosmos-sdk/pull/13178) Add `cosmos.msg.v1.service` protobuf annotation to allow tooling to distinguish between Msg and Query services via reflection.
* [#13236](https://github.com/cosmos/cosmos-sdk/pull/13236) Integrate Filter Logging
* [#13528](https://github.com/cosmos/cosmos-sdk/pull/13528) Update `ValidateMemoDecorator` to only check memo against `MaxMemoCharacters` param when a memo is present.
* [#13651](https://github.com/cosmos/cosmos-sdk/pull/13651) Update `server/config/config.GetConfig` function.
* [#13781](https://github.com/cosmos/cosmos-sdk/pull/13781) Remove `client/keys.KeysCdc`.
* [#13802](https://github.com/cosmos/cosmos-sdk/pull/13802) Add --output-document flag to the export CLI command to allow writing genesis state to a file.
* [#13794](https://github.com/cosmos/cosmos-sdk/pull/13794) `types/module.Manager` now supports the
`cosmossdk.io/core/appmodule.AppModule` API via the new `NewManagerFromMap` constructor.
* [#14175](https://github.com/cosmos/cosmos-sdk/pull/14175) Add `server.DefaultBaseappOptions(appopts)` function to reduce boiler plate in root.go. 

### State Machine Breaking

* (baseapp, x/auth/posthandler) [#13940](https://github.com/cosmos/cosmos-sdk/pull/13940) Update `PostHandler` to receive the `runTx` success boolean.
* (store) [#14378](https://github.com/cosmos/cosmos-sdk/pull/14378) The `CacheKV` store is thread-safe again, which includes improved iteration and deletion logic. Iteration is on a strictly isolated view now, which is breaking from previous behavior.
* (x/bank) [#14538](https://github.com/cosmos/cosmos-sdk/pull/14538) Validate denom in bank balances GRPC queries.
* (x/group) [#14465](https://github.com/cosmos/cosmos-sdk/pull/14465) Add title and summary to proposal struct.
* (x/gov) [#14390](https://github.com/cosmos/cosmos-sdk/pull/14390) Add title, proposer and summary to proposal struct.
* (x/group) [#14071](https://github.com/cosmos/cosmos-sdk/pull/14071) Don't re-tally proposal after voting period end if they have been marked as ACCEPTED or REJECTED.
* (x/group) [#13742](https://github.com/cosmos/cosmos-sdk/pull/13742) Migrate group policy account from module accounts to base account.
* (x/auth)[#13780](https://github.com/cosmos/cosmos-sdk/pull/13780) `id` (type of int64) in `AccountAddressByID` grpc query is now deprecated, update to account-id(type of uint64) to use `AccountAddressByID`.
* (codec) [#13307](https://github.com/cosmos/cosmos-sdk/pull/13307) Register all modules' `Msg`s with group's ModuleCdc so that Amino sign bytes are correctly generated.* (x/gov) 
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

### API Breaking Changes

* Migrate to CometBFT. Follow the migration instructions in the [upgrade guide](./UPGRADING.md#migration-to-cometbft-part-1).
* (simulation) [#14728](https://github.com/cosmos/cosmos-sdk/pull/14728) Rename the `ParamChanges` field to `LegacyParamChange` and `Contents` to `LegacyProposalContents` in `simulation.SimulationState`. Additionally it adds a `ProposalMsgs` field to `simulation.SimulationState`.
* (x/gov) [#14782](https://github.com/cosmos/cosmos-sdk/pull/14782) Move the `metadata` argument in `govv1.NewProposal` alongside `title` and `summary`.
* (x/upgrade) [#14216](https://github.com/cosmos/cosmos-sdk/pull/14216) Change upgrade keeper receiver to upgrade keeper pointers.
* (x/auth) [#13780](https://github.com/cosmos/cosmos-sdk/pull/13780) Querying with `id` (type of int64) in `AccountAddressByID` grpc query now throws error, use account-id(type of uint64) instead.
* (store) [#13516](https://github.com/cosmos/cosmos-sdk/pull/13516) Update State Streaming APIs:
    * Add method `ListenCommit` to `ABCIListener`
    * Move `ListeningEnabled` and  `AddListener` methods to `CommitMultiStore`
    * Remove `CacheWrapWithListeners` from `CacheWrap` and `CacheWrapper` interfaces
    * Remove listening APIs from the caching layer (it should only listen to the `rootmulti.Store`)
    * Add three new options to file streaming service constructor.
    * Modify `ABCIListener` such that any error from any method will always halt the app via `panic`
* (x/auth) [#13877](https://github.com/cosmos/cosmos-sdk/pull/13877) Rename `AccountKeeper`'s `GetNextAccountNumber` to `NextAccountNumber`.
* (x/evidence) [#13740](https://github.com/cosmos/cosmos-sdk/pull/13740) The `NewQueryEvidenceRequest` function now takes `hash` as a HEX encoded `string`.
* (server) [#13485](https://github.com/cosmos/cosmos-sdk/pull/13485) The `Application` service now requires the `RegisterNodeService` method to be implemented.
* [#13437](https://github.com/cosmos/cosmos-sdk/pull/13437) Add a list of modules to export argument in `ExportAppStateAndValidators`.
* (simapp) [#13402](https://github.com/cosmos/cosmos-sdk/pull/13402) Move simulation flags to `x/simulation/client/cli`.
* (simapp) [#13402](https://github.com/cosmos/cosmos-sdk/pull/13402) Move simulation helpers functions (`SetupSimulation`, `SimulationOperations`, `CheckExportSimulation`, `PrintStats`, `GetSimulationLog`) to `testutil/sims`.
* (simapp) [#13402](https://github.com/cosmos/cosmos-sdk/pull/13402) Move `testutil/rest` package to `testutil`.
* (types) [#13380](https://github.com/cosmos/cosmos-sdk/pull/13380) Remove deprecated `sdk.NewLevelDB`.
* (simapp) [#13378](https://github.com/cosmos/cosmos-sdk/pull/13378) Move `simapp.App` to `runtime.AppI`.
* (tx) [#12659](https://github.com/cosmos/cosmos-sdk/pull/12659) Remove broadcast mode `block`.
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
* (store) [#11825](https://github.com/cosmos/cosmos-sdk/pull/11825) Make extension snapshotter interface safer to use, renamed the util function `WriteExtensionItem` to `WriteExtensionPayload`.
* (x/genutil)[#12956](https://github.com/cosmos/cosmos-sdk/pull/12956) `genutil.AppModuleBasic` has a new attribute: genesis transaction validation function. The existing validation logic is implemented in `genutiltypes.DefaultMessageValidator`. Use `genutil.NewAppModuleBasic` to create a new genutil Module Basic.
* (codec) [#12964](https://github.com/cosmos/cosmos-sdk/pull/12964) `ProtoCodec.MarshalInterface` now returns an error when serializing unregistered types and a subsequent `ProtoCodec.UnmarshalInterface` would fail.
* (x/staking) [#12973](https://github.com/cosmos/cosmos-sdk/pull/12973) Removed `stakingkeeper.RandomValidator`. Use `testutil.RandSliceElem(r, sk.GetAllValidators(ctx))` instead.
* (x/gov) [#13160](https://github.com/cosmos/cosmos-sdk/pull/13160) Remove custom marshaling of proposl and voteoption.
* (types) [#13430](https://github.com/cosmos/cosmos-sdk/pull/13430) Remove unused code `ResponseCheckTx` and `ResponseDeliverTx`
* (store) [#13529](https://github.com/cosmos/cosmos-sdk/pull/13529) Add method `LatestVersion` to `MultiStore` interface, add method `SetQueryMultiStore` to baesapp to support alternative `MultiStore` implementation for query service.
* (pruning) [#13609](https://github.com/cosmos/cosmos-sdk/pull/13609) Move pruning package to be under store package
* [#13794](https://github.com/cosmos/cosmos-sdk/pull/13794) Most methods on `types/module.AppModule` have been moved to 
extension interfaces. `module.Manager.Modules` is now of type `map[string]interface{}` to support in parallel the new 
`cosmossdk.io/core/appmodule.AppModule` API.

### CLI Breaking Changes

* (genesis) [#14149](https://github.com/cosmos/cosmos-sdk/pull/14149) Add `simd genesis` command, which contains all genesis-related sub-commands.
* (x/genutil) [#13535](https://github.com/cosmos/cosmos-sdk/pull/13535) Replace in `simd init`, the `--staking-bond-denom` flag with `--default-denom` which is used for all default denomination in the genesis, instead of only staking.

### Bug Fixes

* (x/auth/vesting) [#15373](https://github.com/cosmos/cosmos-sdk/pull/15373) Add extra checks when creating a periodic vesting account.
* (x/auth) [#13838](https://github.com/cosmos/cosmos-sdk/pull/13838) Fix calling `String()` and `MarshalYAML` panics when pubkey is set on a `BaseAccount``. 
* (x/evidence) [#13740](https://github.com/cosmos/cosmos-sdk/pull/13740) Fix evidence query API to decode the hash properly.
* (bank) [#13691](https://github.com/cosmos/cosmos-sdk/issues/13691) Fix unhandled error for vesting account transfers, when total vesting amount exceeds total balance.
* [#13553](https://github.com/cosmos/cosmos-sdk/pull/13553) Ensure all parameter validation for decimal types handles nil decimal values.
* [#13145](https://github.com/cosmos/cosmos-sdk/pull/13145) Fix panic when calling `String()` to a Record struct type.
* [#13116](https://github.com/cosmos/cosmos-sdk/pull/13116) Fix a dead-lock in the `Group-TotalWeight` `x/group` invariant.
* (types) [#12154](https://github.com/cosmos/cosmos-sdk/pull/12154) Add `baseAccountGetter` to avoid invalid account error when create vesting account.
* (x/staking) [#12303](https://github.com/cosmos/cosmos-sdk/pull/12303) Use bytes instead of string comparison in delete validator queue
* (store/rootmulti) [#12487](https://github.com/cosmos/cosmos-sdk/pull/12487) Fix non-deterministic map iteration.
* (sdk/dec_coins) [#12903](https://github.com/cosmos/cosmos-sdk/pull/12903) Fix nil `DecCoin` creation when converting `Coins` to `DecCoins`
* (store) [#12945](https://github.com/cosmos/cosmos-sdk/pull/12945) Fix nil end semantics in store/cachekv/iterator when iterating a dirty cache.
* (x/gov) [#13051](https://github.com/cosmos/cosmos-sdk/pull/13051) In SubmitPropsal, when a legacy msg fails it's handler call, wrap the error as ErrInvalidProposalContent (instead of ErrNoProposalHandlerExists).
* (snapshot) [#13400](https://github.com/cosmos/cosmos-sdk/pull/13400) Fix snapshot checksum issue in golang 1.19. 
* (server) [#13778](https://github.com/cosmos/cosmos-sdk/pull/13778) Set Cosmos SDK default endpoints to localhost to avoid unknown exposure of endpoints.
* (x/auth) [#13877](https://github.com/cosmos/cosmos-sdk/pull/13877) Handle missing account numbers during `InitGenesis`.
* (x/gov) [#13918](https://github.com/cosmos/cosmos-sdk/pull/13918) Propagate message errors when executing a proposal.

### Deprecated

* (x/evidence) [#13740](https://github.com/cosmos/cosmos-sdk/pull/13740) The `evidence_hash` field of `QueryEvidenceRequest` has been deprecated and now contains a new field `hash` with type `string`.
* (x/bank) [#11859](https://github.com/cosmos/cosmos-sdk/pull/11859) The Params.SendEnabled field is deprecated and unusable.
  The information can now be accessed using the BankKeeper.
  Setting can be done using MsgSetSendEnabled as a governance proposal.
  A SendEnabled query has been added to both GRPC and CLI.

>>>>>>> 7a778f5c9 (refactor: add MigrateHandler to allow reuse migrate genesis related function  (#17296))
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

* (types) [\#12201](https://github.com/cosmos/cosmos-sdk/pull/12201) Add `MustAccAddressFromBech32` util function
* [\#11696](https://github.com/cosmos/cosmos-sdk/pull/11696) Rename `helpers.GenTx` to `GenSignedMockTx` to avoid confusion with genutil's `GenTxCmd`.
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
