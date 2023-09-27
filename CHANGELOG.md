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

* (x/protocolpool) [#17657](https://github.com/cosmos/cosmos-sdk/pull/17657) Create a new `x/protocolpool` module that is responsible for handling community pool funds. This module is split out into a new module from x/distribution.
* (baseapp) [#16581](https://github.com/cosmos/cosmos-sdk/pull/16581) Implement Optimistic Execution as an experimental feature (not enabled by default).
* (client/keys) [#17639](https://github.com/cosmos/cosmos-sdk/pull/17639) Allows using and saving public keys encoded as base64
* (client) [#17513](https://github.com/cosmos/cosmos-sdk/pull/17513) Allow overwritting `client.toml`. Use `client.CreateClientConfig` in place of `client.ReadFromClientConfig` and provide a custom template and a custom config.
* (x/bank) [#17569](https://github.com/cosmos/cosmos-sdk/pull/17569) Introduce a new message type, `MsgBurn `, to burn coins.
* (server) [#17094](https://github.com/cosmos/cosmos-sdk/pull/17094) Add duration `shutdown-grace` for resource clean up (closing database handles) before exit.

### Improvements

* (client) [#17503](https://github.com/cosmos/cosmos-sdk/pull/17503) Add `client.Context{}.WithAddressCodec`, `WithValidatorAddressCodec`, `WithConsensusAddressCodec` to provide address codecs to the client context. See the [UPGRADING.md](./UPGRADING.md) for more details.
* (crypto/keyring) [#17503](https://github.com/cosmos/cosmos-sdk/pull/17503) Simplify keyring interfaces to use `[]byte` instead of `sdk.Address` for addresses.
* (all) [#16537](https://github.com/cosmos/cosmos-sdk/pull/16537) Properly propagated `fmt.Errorf` errors and using `errors.New` where appropriate.
* (rpc) [#17470](https://github.com/cosmos/cosmos-sdk/pull/17470) Avoid open 0.0.0.0 to public by default and add `listen-ip-address` argument for `testnet init-files` cmd.
* (types) [#17670](https://github.com/cosmos/cosmos-sdk/pull/17670) Use `ctx.CometInfo` in place of `ctx.VoteInfos`
* [#17733](https://github.com/cosmos/cosmos-sdk/pull/17733) Ensure `buf export` exports all proto dependencies
  
### Bug Fixes

* (x/gov) [#17873](https://github.com/cosmos/cosmos-sdk/pull/17873) Fail any inactive and active proposals whose messages cannot be decoded.

### API Breaking Changes

* (x/distribution) [#17657](https://github.com/cosmos/cosmos-sdk/pull/17657) The `FundCommunityPool` and `DistributeFromFeePool` keeper methods are now removed from x/distribution.
* (x/distribution) [#17657](https://github.com/cosmos/cosmos-sdk/pull/17657) The distribution module keeper now takes a new argument `PoolKeeper` in addition.
* (app) [#17838](https://github.com/cosmos/cosmos-sdk/pull/17838) Params module was removed from simapp and all imports of the params module removed throughout the repo. 
    * The Cosmos SDK has migrated aay from using params, if you're app still uses it, then you can leave it plugged into your app
* (x/staking) [#17778](https://github.com/cosmos/cosmos-sdk/pull/17778) Use collections for `Params`
    * remove from `Keeper`: `GetParams`, `SetParams`
* (types/simulation) [#17737](https://github.com/cosmos/cosmos-sdk/pull/17737) Remove unused parameter from `RandomFees`
* (x/staking) [#17486](https://github.com/cosmos/cosmos-sdk/pull/17486) Use collections for `RedelegationQueueKey`:
    * remove from `types`: `GetRedelegationTimeKey`
    * remove from `Keeper`: `RedelegationQueueIterator`
* (x/staking) [#17562](https://github.com/cosmos/cosmos-sdk/pull/17562) Use collections for `ValidatorQueue`
    * remove from `types`: `GetValidatorQueueKey`, `ParseValidatorQueueKey`
    * remove from `Keeper`: `ValidatorQueueIterator`
* (x/staking) [#17498](https://github.com/cosmos/cosmos-sdk/pull/17498) Use collections for `LastValidatorPower`:
    * remove from `types`: `GetLastValidatorPowerKey`
    * remove from `Keeper`: `LastValidatorsIterator`, `IterateLastValidators`
* (x/staking) [#17291](https://github.com/cosmos/cosmos-sdk/pull/17291) Use collections for `UnbondingDelegationByValIndex`:
    * remove from `types`: `GetUBDKeyFromValIndexKey`, `GetUBDsByValIndexKey`, `GetUBDByValIndexKey`
* (x/slashing) [#17568](https://github.com/cosmos/cosmos-sdk/pull/17568) Use collections for `ValidatorMissedBlockBitmap`:
    * remove from `types`: `ValidatorMissedBlockBitmapPrefixKey`, `ValidatorMissedBlockBitmapKey`
* (x/staking) [#17481](https://github.com/cosmos/cosmos-sdk/pull/17481) Use collections for `UnbondingQueue`:
    * remove from `Keeper`: `UBDQueueIterator`
    * remove from `types`: `GetUnbondingDelegationTimeKey`
* (x/staking) [#17123](https://github.com/cosmos/cosmos-sdk/pull/17123) Use collections for `Validators`
* (x/staking) [#17270](https://github.com/cosmos/cosmos-sdk/pull/17270) Use collections for `UnbondingDelegation`:
    * remove from `types`: `GetUBDsKey`
    * remove from `Keeper`: `IterateUnbondingDelegations`, `IterateDelegatorUnbondingDelegations`
* (client/keys) [#17503](https://github.com/cosmos/cosmos-sdk/pull/17503) `clientkeys.NewKeyOutput`, `MkConsKeyOutput`, `MkValKeyOutput`, `MkAccKeyOutput`, `MkAccKeysOutput` now take their corresponding address codec instead of using the global SDK config.
* (x/staking) [#17336](https://github.com/cosmos/cosmos-sdk/pull/17336) Use collections for `RedelegationByValDstIndexKey`:
    * remove from `types`: `GetREDByValDstIndexKey`, `GetREDsToValDstIndexKey`
* (x/staking) [#17332](https://github.com/cosmos/cosmos-sdk/pull/17332) Use collections for `RedelegationByValSrcIndexKey`:
    * remove from `types`: `GetREDKeyFromValSrcIndexKey`, `GetREDsFromValSrcIndexKey`
* (x/staking) [#17315](https://github.com/cosmos/cosmos-sdk/pull/17315) Use collections for `RedelegationKey`:
    * remove from `keeper`: `GetRedelegation`
* (types) `module.BeginBlockAppModule` has been replaced by Core API `appmodule.HasBeginBlocker`.
* (types) `module.EndBlockAppModule` has been replaced by Core API `appmodule.HasEndBlocker` or `module.HasABCIEndBlock` when needing validator updates.
* (x/slashing) [#17044](https://github.com/cosmos/cosmos-sdk/pull/17044) Use collections for `AddrPubkeyRelation`:
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
* (client) [#17215](https://github.com/cosmos/cosmos-sdk/pull/17215) `server.StartCmd`,`server.ExportCmd`,`server.NewRollbackCmd`,`pruning.Cmd`,`genutilcli.InitCmd`,`genutilcli.GenTxCmd`,`genutilcli.CollectGenTxsCmd`,`genutilcli.AddGenesisAccountCmd`, do not take a home directory anymore. It is inferred from the root command.
* (baseapp) [#16244](https://github.com/cosmos/cosmos-sdk/pull/16244) `SetProtocolVersion` has been renamed to `SetAppVersion`. It now updates the consensus params in baseapp's `ParamStore`.
* (types) [#17348](https://github.com/cosmos/cosmos-sdk/pull/17348) Remove the `WrapServiceResult` function.
    * The `*sdk.Result` returned by the msg server router will not contain the `.Data` field.
* (x/staking) [#17335](https://github.com/cosmos/cosmos-sdk/pull/17335) Remove usage of `"github.com/cosmos/cosmos-sdk/x/staking/types".Infraction_*` in favour of `"cosmossdk.io/api/cosmos/staking/v1beta1".Infraction_` in order to remove dependency between modules on staking
* (x/gov) [#17496](https://github.com/cosmos/cosmos-sdk/pull/17469) in `x/gov/types/v1beta1/vote.go` `NewVote` was removed, constructing the struct is required for this type
* (types) [#17426](https://github.com/cosmos/cosmos-sdk/pull/17426) `NewContext` does not take a `cmtproto.Header{}` any longer. 
    * `WithChainID` / `WithBlockHeight` / `WithBlockHeader` must be used to set values on the context
* (x/bank) [#17569](https://github.com/cosmos/cosmos-sdk/pull/17569) `BurnCoins` takes an address instead of a module name
* (x/distribution) [#17670](https://github.com/cosmos/cosmos-sdk/pull/17670) `AllocateTokens` takes `comet.VoteInfos` instead of `[]abci.VoteInfo`
* (types) [#17738](https://github.com/cosmos/cosmos-sdk/pull/17738) `WithBlockTime()` was removed & `BlockTime()` were deprecated in favor of `WithHeaderInfo()` & `HeaderInfo()`. `BlockTime` now gets data from `HeaderInfo()` instead of `BlockHeader()`.
* (client) [#17746](https://github.com/cosmos/cosmos-sdk/pull/17746) `txEncodeAmino` & `txDecodeAmino` txs via grpc and rest were removed
    * `RegisterLegacyAmino` was removed from `AppModuleBasic`
* (x/staking) [#17655](https://github.com/cosmos/cosmos-sdk/pull/17655) `QueryHistoricalInfo` was adjusted to return `HistoricalRecord` and marked `Hist` as deprecated.
* (types) [#17885](https://github.com/cosmos/cosmos-sdk/pull/17885) `InitGenesis` & `ExportGenesis` now take `context.Context` instead of `sdk.Context`

### CLI Breaking Changes

### State Machine Breaking

* (x/distribution) [#17657](https://github.com/cosmos/cosmos-sdk/pull/17657) Migrate community pool funds from x/distribution to x/protocolpool.
* (x/distribution) [#17115](https://github.com/cosmos/cosmos-sdk/pull/17115) Migrate `PreviousProposer` to collections.
* (x/upgrade) [#16244](https://github.com/cosmos/cosmos-sdk/pull/16244) upgrade module no longer stores the app version but gets and sets the app version stored in the `ParamStore` of baseapp.
* (x/staking) [#17655](https://github.com/cosmos/cosmos-sdk/pull/17655) `HistoricalInfo` was replaced with `HistoricalRecord`, it removes the validator set and comet header and only keep what is needed for IBC. 

### Client Breaking Changes

* (x/distribution) [#17657](https://github.com/cosmos/cosmos-sdk/pull/17657) Deprecate `CommunityPool` and `FundCommunityPool` rpc methods. Use x/protocolpool module's rpc methods instead.
* (x/gov) [#17910](https://github.com/cosmos/cosmos-sdk/pull/17910) remove telemetry for counting votes and proposals

## [v0.50.0-rc.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.0-rc.1) - 2023-09-25

### Features

* (baseapp & types) [#17712](https://github.com/cosmos/cosmos-sdk/pull/17712) Introduce `PreBlock`, which runs before begin blocker other modules, and allows to modify consensus parameters, and the changes are visible to the following state machine logics. Additionally it can be used for vote extensions.
* (genutil) [#17571](https://github.com/cosmos/cosmos-sdk/pull/17571) Allow creation of `AppGenesis` without a file lookup.
* (keyring) [#17424](https://github.com/cosmos/cosmos-sdk/pull/17424) Allows to import private keys encoded in hex. 
* (client/rpc) [#17274](https://github.com/cosmos/cosmos-sdk/pull/17274) Add `QueryEventForTxCmd` cmd to subscribe and wait event for transaction by hash.
* (codec) [#17042](https://github.com/cosmos/cosmos-sdk/pull/17042) Add `CollValueV2` which supports encoding of protov2 messages in collections.
* (x/gov) [#16976](https://github.com/cosmos/cosmos-sdk/pull/16976) Add `failed_reason` field to `Proposal` under `x/gov` to indicate the reason for a failed proposal. Referenced from [#238](https://github.com/bnb-chain/greenfield-cosmos-sdk/pull/238) under `bnb-chain/greenfield-cosmos-sdk`.
* (baseapp) [#16898](https://github.com/cosmos/cosmos-sdk/pull/16898) Add `preFinalizeBlockHook` to allow vote extensions persistence.
* (cli) [#16887](https://github.com/cosmos/cosmos-sdk/pull/16887) Add two new CLI commands: `<appd> tx simulate` for simulating a transaction; `<appd> query block-results` for querying CometBFT RPC for block results.
* (x/bank) [#16852](https://github.com/cosmos/cosmos-sdk/pull/16852) Add `DenomMetadataByQueryString` query in bank module to support metadata query by query string.
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
* (x/auth) [#14650](https://github.com/cosmos/cosmos-sdk/pull/14650) Add Textual SignModeHandler. It is however **NOT** enabled by default, and should only be used for **TESTING** purposes until `SIGN_MODE_TEXTUAL` is fully released.
* (x/crisis) [#14588](https://github.com/cosmos/cosmos-sdk/pull/14588) Use CacheContext() in AssertInvariants().
* (mempool) [#14484](https://github.com/cosmos/cosmos-sdk/pull/14484) Add priority nonce mempool option for transaction replacement.
* (query) [#14468](https://github.com/cosmos/cosmos-sdk/pull/14468) Implement pagination for collections.
* (x/gov) [#14373](https://github.com/cosmos/cosmos-sdk/pull/14373) Add new proto field `constitution` of type `string` to gov module genesis state, which allows chain builders to lay a strong foundation by specifying purpose.
* (client) [#14342](https://github.com/cosmos/cosmos-sdk/pull/14342) Add `<app> config` command is now a sub-command, for setting, getting and migrating Cosmos SDK configuration files.
* (x/distribution) [#14322](https://github.com/cosmos/cosmos-sdk/pull/14322) Introduce a new gRPC message handler, `DepositValidatorRewardsPool`, that allows explicit funding of a validator's reward pool.
* (x/bank) [#14224](https://github.com/cosmos/cosmos-sdk/pull/14224) Allow injection of restrictions on transfers using `AppendSendRestriction` or `PrependSendRestriction`.
* [#13473](https://github.com/cosmos/cosmos-sdk/pull/13473) ADR-038: Go plugin system proposal.

>>>>>>> 5e53fb224 (refactor(x/gov): remove gov vote and proposal based telemetry (#17910))
### Improvements

* (x/gov) [#17780](https://github.com/cosmos/cosmos-sdk/pull/17780) Recover panics and turn them into errors when executing x/gov proposals.

### Bug Fixes

* (baseapp) [#17769](https://github.com/cosmos/cosmos-sdk/pull/17769) Ensure we respect block size constraints in the `DefaultProposalHandler`'s `PrepareProposal` handler when a nil or no-op mempool is used. We provide a `TxSelector` type to assist in making transaction selection generalized. We also fix a comparison bug in tx selection when `req.maxTxBytes` is reached.
* (config) [#17649](https://github.com/cosmos/cosmos-sdk/pull/17649) Fix `mempool.max-txs` configuration is invalid in `app.config`.
* (mempool) [#17668](https://github.com/cosmos/cosmos-sdk/pull/17668) Fix `PriorityNonceIterator.Next()` nil pointer ref for min priority at the end of iteration.
* (x/auth) [#17902](https://github.com/cosmos/cosmos-sdk/pull/17902) Remove tip posthandler 

## [v0.47.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.5) - 2023-09-01

### Features

* (client/rpc) [#17274](https://github.com/cosmos/cosmos-sdk/pull/17274) Add `QueryEventForTxCmd` cmd to subscribe and wait event for transaction by hash.
* (keyring) [#17424](https://github.com/cosmos/cosmos-sdk/pull/17424) Allows to import private keys encoded in hex. 

### Improvements

* (x/gov) [#17387](https://github.com/cosmos/cosmos-sdk/pull/17387) Add `MsgSubmitProposal` `SetMsgs` method. 
* (x/gov) [#17354](https://github.com/cosmos/cosmos-sdk/issues/17354) Emit `VoterAddr` in `proposal_vote` event.
* (x/group, x/gov) [#17220](https://github.com/cosmos/cosmos-sdk/pull/17220) Add `--skip-metadata` flag in `draft-proposal` to skip metadata prompt.
* (x/genutil) [#17296](https://github.com/cosmos/cosmos-sdk/pull/17296) Add `MigrateHandler` to allow reuse migrate genesis related function.
    * In v0.46, v0.47 this function is additive to the `genesis migrate` command. However in v0.50+, adding custom migrations to the `genesis migrate` command is directly possible.

### Bug Fixes

* (server) [#17181](https://github.com/cosmos/cosmos-sdk/pull/17181) Fix `db_backend` lookup fallback from `config.toml`.
* (runtime) [#17284](https://github.com/cosmos/cosmos-sdk/pull/17284) Properly allow to combine depinject-enabled modules and non-depinject-enabled modules in app v2.
* (baseapp) [#17159](https://github.com/cosmos/cosmos-sdk/pull/17159) Validators can propose blocks that exceed the gas limit.
* (baseapp) [#16547](https://github.com/cosmos/cosmos-sdk/pull/16547) Ensure a transaction's gas limit cannot exceed the block gas limit.
* (x/gov,x/group) [#17220](https://github.com/cosmos/cosmos-sdk/pull/17220) Do not try validate `msgURL` as web URL in `draft-proposal` command.
* (cli) [#17188](https://github.com/cosmos/cosmos-sdk/pull/17188) Fix `--output-document` flag in `tx multi-sign`.
* (x/auth) [#17209](https://github.com/cosmos/cosmos-sdk/pull/17209) Internal error on AccountInfo when account's public key is not set.

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
