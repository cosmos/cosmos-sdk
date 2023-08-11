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
* (x/group, x/gov) [#17220](https://github.com/cosmos/cosmos-sdk/pull/17220) Add `--skip-metadata` flag in `draft-proposal` to skip metadata prompt.
* (x/genutil) [#17296](https://github.com/cosmos/cosmos-sdk/pull/17296) Add `MigrateHandler` to allow reuse migrate genesis related function.
    * In v0.46, v0.47 this function is additive to the `genesis migrate` command. However in v0.50+, adding custom migrations to the `genesis migrate` command is directly possible.

### Bug Fixes

* (runtime) [#17284](https://github.com/cosmos/cosmos-sdk/pull/17284) Properly allow to combine depinject-enabled modules and non-depinject-enabled modules in app v2.
* (baseapp) [#17159](https://github.com/cosmos/cosmos-sdk/pull/17159) Validators can propose blocks that exceed the gas limit.
<<<<<<< HEAD
* (baseapp) [#16547](https://github.com/cosmos/cosmos-sdk/pull/16547) Ensure a transaction's gas limit cannot exceed the block gas limit.
* (x/gov,x/group) [#17220](https://github.com/cosmos/cosmos-sdk/pull/17220) Do not try validate `msgURL` as web URL in `draft-proposal` command.
* (cli) [#17188](https://github.com/cosmos/cosmos-sdk/pull/17188) Fix `--output-document` flag in `tx multi-sign`.
* (x/auth) [#17209](https://github.com/cosmos/cosmos-sdk/pull/17209) Internal error on AccountInfo when account's public key is not set.
=======
* (x/group) [#17146](https://github.com/cosmos/cosmos-sdk/pull/17146) Rename x/group legacy ORM package's error codespace from "orm" to "legacy_orm", preventing collisions with ORM's error codespace "orm".
* (x/bank) [#17170](https://github.com/cosmos/cosmos-sdk/pull/17170) Avoid empty spendable error message on send coins.
* (x/distribution) [#17236](https://github.com/cosmos/cosmos-sdk/pull/17236) Using "validateCommunityTax" in "Params.ValidateBasic", preventing panic when field "CommunityTax" is nil.
* (migrations) [#16583](https://github.com/cosmos/cosmos-sdk/pull/16583) Add `MigrationModuleManager` to handle migration of upgrade module before other modules, ensuring access to the updated context with consensus parameters within the same block that executes the migration.

### API Breaking Changes

* (types) [#17358](https://github.com/cosmos/cosmos-sdk/pull/17358) Remove deprecated `sdk.Handler`, use `baseapp.MsgServiceHandler` instead.
* (x/slashing) [17044](https://github.com/cosmos/cosmos-sdk/pull/17044) Use collections for `AddrPubkeyRelation`:
    * remove from `types`: `AddrPubkeyRelationKey`
    * remove from `Keeper`: `AddPubkey`
* (x/staking) [#17260](https://github.com/cosmos/cosmos-sdk/pull/17260) Use collections for `DelegationKey`:
    * remove from `types`: `GetDelegationKey`, `GetDelegationsKey`
* (x/staking) [#17288](https://github.com/cosmos/cosmos-sdk/pull/17288) Use collections for `UnbondingIndex`:
    * remove from `types`: `GetUnbondingIndexKey`.
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
    * remove `Keeper`: `GetHistoricalInfo`, `SetHistoricalInfo`
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
* (client) [#17215](https://github.com/cosmos/cosmos-sdk/pull/17215) `server.StartCmd`,`server.ExportCmd`,`server.NewRollbackCmd`,`pruning.Cmd`,`genutilcli.InitCmd`,`genutilcli.GenTxCmd`,`genutilcli.CollectGenTxsCmd`,`genutilcli.AddGenesisAccountCmd`, do not take a home directory anymore. It is inferred from the root command.

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
* (baseapp) [#14417](https://github.com/cosmos/cosmos-sdk/pull/14417) The store pacakge no longer has a dependency on baseapp. 
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
>>>>>>> 0c1f6fc16 (fix: Add MigrationModuleManager to handle migration of upgrade module before other modules (#16583))

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

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/cosmos-sdk/blob/main/CHANGELOG.md#v0460---2022-07-26).
