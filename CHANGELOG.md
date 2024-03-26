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

Every module contains its own CHANGELOG.md. Please refer to the module you are interested in.

### Features

* (runtime) [#19571](https://github.com/cosmos/cosmos-sdk/pull/19571) Implement `core/router.Service` it in runtime. This service is present in all modules (when using depinject).
* (types) [#19164](https://github.com/cosmos/cosmos-sdk/pull/19164) Add a ValueCodec for the math.Uint type that can be used in collections maps.
* (types) [#19281](https://github.com/cosmos/cosmos-sdk/pull/19281) Added a new method, `IsGT`, for `types.Coin`. This method is used to check if a `types.Coin` is greater than another `types.Coin`.
* (client) [#18557](https://github.com/cosmos/cosmos-sdk/pull/18557) Add `--qrcode` flag to `keys show` command to support displaying keys address QR code.
* (client) [#18101](https://github.com/cosmos/cosmos-sdk/pull/18101) Add a `keyring-default-keyname` in `client.toml` for specifying a default key name, and skip the need to use the `--from` flag when signing transactions.
* (tests) [#17868](https://github.com/cosmos/cosmos-sdk/pull/17868) Added helper method `SubmitTestTx` in testutil to broadcast test txns to test e2e tests.
* (client) [#17513](https://github.com/cosmos/cosmos-sdk/pull/17513) Allow overwriting `client.toml`. Use `client.CreateClientConfig` in place of `client.ReadFromClientConfig` and provide a custom template and a custom config.
* (runtime) [#18475](https://github.com/cosmos/cosmos-sdk/pull/18475) Adds an implementation for core.branch.Service.
* (baseapp) [#18499](https://github.com/cosmos/cosmos-sdk/pull/18499) Add `MsgRouter` response type from message name function.
* (types) [#18768](https://github.com/cosmos/cosmos-sdk/pull/18768) Add MustValAddressFromBech32 function.
* (gRPC) [#19049](https://github.com/cosmos/cosmos-sdk/pull/19049) Add debug log prints for each gRPC request.
* (x/consensus) [#19483](https://github.com/cosmos/cosmos-sdk/pull/19483) Add consensus messages registration to consensus module.
* (types) [#19759](https://github.com/cosmos/cosmos-sdk/pull/19759) Align SignerExtractionAdapter in PriorityNonceMempool Remove.

### Improvements

* (types) [#19672](https://github.com/cosmos/cosmos-sdk/pull/19672) `PreBlock` now returns only an error for consistency with server/v2. The SDK has upgraded x/upgrade accordingly. `ResponsePreBlock` hence has been removed.
* (server) [#19455](https://github.com/cosmos/cosmos-sdk/pull/19455) Allow calling back into the application struct in PostSetup.
* (types) [#19512](https://github.com/cosmos/cosmos-sdk/pull/19512) The notion of basic manager does not exist anymore (and all related helpers).
    * The module manager now can do everything that the basic manager was doing.
    * `AppModuleBasic` has been deprecated for extension interfaces.
    * Modules can now implement `appmodule.HasRegisterInterfaces`, `modue.HasGRPCGateway` and `module.HasAminoCodec` when relevant.
    * SDK modules now directly implement those extension interfaces on `AppModule` instead of `AppModuleBasic`.
* (client/keys) [#18950](https://github.com/cosmos/cosmos-sdk/pull/18950) Improve `<appd> keys add`, `<appd> keys import` and `<appd> keys rename` by checking name validation.
* (client/keys) [#18745](https://github.com/cosmos/cosmos-sdk/pull/18745) Improve `<appd> keys export` and `<appd> keys mnemonic` by adding --yes option to skip interactive confirmation.
* (client/keys) [#18743](https://github.com/cosmos/cosmos-sdk/pull/18743) Improve `<appd> keys add -i` by hiding inputting of bip39 passphrase.
* (client/keys) [#18703](https://github.com/cosmos/cosmos-sdk/pull/18703) Improve `<appd> keys add` and `<appd> keys show` by checking whether there are duplicate keys in the multisig case.
    * Usage of `Must...` kind of functions are avoided in keeper methods.
* (client/keys) [#18687](https://github.com/cosmos/cosmos-sdk/pull/18687) Improve `<appd> keys mnemonic` by displaying mnemonic discreetly on an alternate screen and adding `--indiscreet` option to disable it.
* (client/keys) [#18684](https://github.com/cosmos/cosmos-sdk/pull/18684) Improve `<appd> keys export` by displaying unarmored hex private key discreetly on an alternate screen and adding `--indiscreet` option to disable it.
* (client/keys) [#18663](https://github.com/cosmos/cosmos-sdk/pull/18663) Improve `<appd> keys add` by displaying mnemonic discreetly on an alternate screen and adding `--indiscreet` option to disable it.
* (types) [#18440](https://github.com/cosmos/cosmos-sdk/pull/18440) Add `AmountOfNoValidation` to `sdk.DecCoins`.
* (client) [#17503](https://github.com/cosmos/cosmos-sdk/pull/17503) Add `client.Context{}.WithAddressCodec`, `WithValidatorAddressCodec`, `WithConsensusAddressCodec` to provide address codecs to the client context. See the [UPGRADING.md](./UPGRADING.md) for more details.
* (crypto/keyring) [#17503](https://github.com/cosmos/cosmos-sdk/pull/17503) Simplify keyring interfaces to use `[]byte` instead of `sdk.Address` for addresses.
* (all) [#16537](https://github.com/cosmos/cosmos-sdk/pull/16537) Properly propagated `fmt.Errorf` errors and using `errors.New` where appropriate.
* (rpc) [#17470](https://github.com/cosmos/cosmos-sdk/pull/17470) Avoid open 0.0.0.0 to public by default and add `listen-ip-address` argument for `testnet init-files` cmd.
* (types) [#17670](https://github.com/cosmos/cosmos-sdk/pull/17670) Use `ctx.CometInfo` in place of `ctx.VoteInfos`
* [#17733](https://github.com/cosmos/cosmos-sdk/pull/17733) Ensure `buf export` exports all proto dependencies
* (crypto/keys) [#18026](https://github.com/cosmos/cosmos-sdk/pull/18026) Made public key generation constant time on `secp256k1`
* (crypto | x/auth) [#14372](https://github.com/cosmos/cosmos-sdk/pull/18194) Key checks on signatures antehandle.
* (types) [#18963](https://github.com/cosmos/cosmos-sdk/pull/18963) Swap out amino json encoding of `ABCIMessageLogs` for std lib json encoding
* (x/auth) [#19651](https://github.com/cosmos/cosmos-sdk/pull/19651) Allow empty public keys in `GetSignBytesAdapter`.
* (x/genutil) [#19735](https://github.com/cosmos/cosmos-sdk/pull/19735) Update genesis api to match new `appmodule.HasGenesis` interface.

### Bug Fixes

* (baseapp) [#18727](https://github.com/cosmos/cosmos-sdk/pull/18727) Ensure that `BaseApp.Init` firstly returns any errors from a nil commit multistore instead of panicking on nil dereferencing and before sealing the app.
* (client) [#18622](https://github.com/cosmos/cosmos-sdk/pull/18622) Fixed a potential under/overflow from `uint64->int64` when computing gas fees as a LegacyDec.
* (client/keys) [#18562](https://github.com/cosmos/cosmos-sdk/pull/18562) `keys delete` won't terminate when a key is not found.
* (baseapp) [#18383](https://github.com/cosmos/cosmos-sdk/pull/18383) Fixed a data race inside BaseApp.getContext, found by end-to-end (e2e) tests.
* (client/server) [#18345](https://github.com/cosmos/cosmos-sdk/pull/18345) Consistently set viper prefix in client and server. It defaults for the binary name for both client and server.
* (simulation) [#17911](https://github.com/cosmos/cosmos-sdk/pull/17911) Fix all problems with executing command `make test-sim-custom-genesis-fast` for simulation test.
* (simulation) [#18196](https://github.com/cosmos/cosmos-sdk/pull/18196) Fix the problem of `validator set is empty after InitGenesis` in simulation test.
* (baseapp) [#18551](https://github.com/cosmos/cosmos-sdk/pull/18551) Fix SelectTxForProposal the calculation method of tx bytes size is inconsistent with CometBFT
* (server) [#18994](https://github.com/cosmos/cosmos-sdk/pull/18994) Update server context directly rather than a reference to a sub-object
* (crypto) [#19691](https://github.com/cosmos/cosmos-sdk/pull/19691) Fix tx sign doesn't throw an error when incorrect Ledger is used.
* [#19833](https://github.com/cosmos/cosmos-sdk/pull/19833) Fix some places in which we call Remove inside a Walk.
* [#19851](https://github.com/cosmos/cosmos-sdk/pull/19851) Fix some places in which we call Remove inside a Walk (x/staking and x/gov).

### API Breaking Changes

* (types) [#19792](https://github.com/cosmos/cosmos-sdk/pull/19792) In `MsgSimulatorFn` `sdk.Context` argument is replaced for an `address.Codec`. It also returns an error.
* (types) [#19742](https://github.com/cosmos/cosmos-sdk/pull/19742) Removes the use of `Accounts.String`
    * `SimulationState` now has address and validator codecs as fields.
* (types) [#19447](https://github.com/cosmos/cosmos-sdk/pull/19447) `module.testutil.MakeTestEncodingConfig` now takes `CodecOptions` as argument.
* (types) [#19512](https://github.com/cosmos/cosmos-sdk/pull/19512) Remove basic manager and all related functions (`module.BasicManager`, `module.NewBasicManager`, `module.NewBasicManagerFromManager`, `NewGenesisOnlyAppModule`).
    * The module manager now can do everything that the basic manager was doing.
    * When using runtime, just inject the module manager when needed using your app config.
    * All `AppModuleBasic` structs have been removed.
* (x/consensus) [#19488](https://github.com/cosmos/cosmos-sdk/pull/19488) Consensus module creation takes `appmodule.Environment` instead of individual services.
* (server) [#18303](https://github.com/cosmos/cosmos-sdk/pull/18303) `x/genutil` now handles the application export. `server.AddCommands` does not take an `AppExporter` but instead `genutilcli.Commands` does.
* (x/gov/testutil) [#17986](https://github.com/cosmos/cosmos-sdk/pull/18036) `MsgDeposit` has been removed because of AutoCLI migration.
* (x/staking/testutil) [#17986](https://github.com/cosmos/cosmos-sdk/pull/17986) `MsgRedelegateExec`, `MsgUnbondExec` has been removed because of AutoCLI migration.
* (x/bank/testutil) [#17868](https://github.com/cosmos/cosmos-sdk/pull/17868) `MsgSendExec` has been removed because of AutoCLI migration.
* (app) [#17838](https://github.com/cosmos/cosmos-sdk/pull/17838) Params module was removed from simapp and all imports of the params module removed throughout the repo.
    * The Cosmos SDK has migrated away from using params, if your app still uses it, then you can leave it plugged into your app
* (types/simulation) [#17737](https://github.com/cosmos/cosmos-sdk/pull/17737) Remove unused parameter from `RandomFees`
* (client/keys) [#17503](https://github.com/cosmos/cosmos-sdk/pull/17503) `clientkeys.NewKeyOutput`, `MkConsKeyOutput`, `MkValKeyOutput`, `MkAccKeyOutput`, `MkAccKeysOutput` now take their corresponding address codec instead of using the global SDK config.
* (types) `module.BeginBlockAppModule` has been replaced by Core API `appmodule.HasBeginBlocker`.
* (types) `module.EndBlockAppModule` has been replaced by Core API `appmodule.HasEndBlocker` or `module.HasABCIEndBlock` when needing validator updates.
* (client) [#17259](https://github.com/cosmos/cosmos-sdk/pull/17259) Remove deprecated `clientCtx.PrintObjectLegacy`. Use `clientCtx.PrintProto` or `clientCtx.PrintRaw` instead.
* (types) [#16918](https://github.com/cosmos/cosmos-sdk/pull/16918) Remove `IntProto` and `DecProto`. Instead, `math.Int` and `math.LegacyDec` should be used respectively. Both types support `Marshal` and `Unmarshal` which should be used for binary marshaling.
* (client) [#17215](https://github.com/cosmos/cosmos-sdk/pull/17215) `server.StartCmd`,`server.ExportCmd`,`server.NewRollbackCmd`,`pruning.Cmd`,`genutilcli.InitCmd`,`genutilcli.GenTxCmd`,`genutilcli.CollectGenTxsCmd`,`genutilcli.AddGenesisAccountCmd`, do not take a home directory anymore. It is inferred from the root command.
* (baseapp) [#16244](https://github.com/cosmos/cosmos-sdk/pull/16244) `SetProtocolVersion` has been renamed to `SetAppVersion`. It now updates the consensus params in baseapp's `ParamStore`.
* (types) [#17348](https://github.com/cosmos/cosmos-sdk/pull/17348) Remove the `WrapServiceResult` function.
    * The `*sdk.Result` returned by the msg server router will not contain the `.Data` field.
* (types) [#17426](https://github.com/cosmos/cosmos-sdk/pull/17426) `NewContext` does not take a `cmtproto.Header{}` any longer.
    * `WithChainID` / `WithBlockHeight` / `WithBlockHeader` must be used to set values on the context
* (types) [#17738](https://github.com/cosmos/cosmos-sdk/pull/17738) `WithBlockTime()` was removed & `BlockTime()` were deprecated in favor of `WithHeaderInfo()` & `HeaderInfo()`. `BlockTime` now gets data from `HeaderInfo()` instead of `BlockHeader()`.
* (client) [#17746](https://github.com/cosmos/cosmos-sdk/pull/17746) `txEncodeAmino` & `txDecodeAmino` txs via grpc and rest were removed
    * `RegisterLegacyAmino` was removed from `AppModuleBasic`
* (types) [#17885](https://github.com/cosmos/cosmos-sdk/pull/17885) `InitGenesis` & `ExportGenesis` now take `context.Context` instead of `sdk.Context`
* (x/group) [#17937](https://github.com/cosmos/cosmos-sdk/pull/17937) Groups module was moved to its own go.mod `cosmossdk.io/x/group`
* (x/gov) [#18197](https://github.com/cosmos/cosmos-sdk/pull/18197) Gov module was moved to its own go.mod `cosmossdk.io/x/gov`
* (x/distribution) [#18199](https://github.com/cosmos/cosmos-sdk/pull/18199) Distribution module was moved to its own go.mod `cosmossdk.io/x/distribution`
* (x/slashing) [#18201](https://github.com/cosmos/cosmos-sdk/pull/18201) Slashing module was moved to its own go.mod `cosmossdk.io/x/slashing`
* (x/staking) [#18257](https://github.com/cosmos/cosmos-sdk/pull/18257) Staking module was moved to its own go.mod `cosmossdk.io/x/staking`
* (x/authz) [#18265](https://github.com/cosmos/cosmos-sdk/pull/18265) Authz module was moved to its own go.mod `cosmossdk.io/x/authz`
* (x/mint) [#18283](https://github.com/cosmos/cosmos-sdk/pull/18283) Mint module was moved to its own go.mod `cosmossdk.io/x/mint`
* (x/consensus) [#18041](https://github.com/cosmos/cosmos-sdk/pull/18041) `ToProtoConsensusParams()` returns an error
* (x/slashing) [#18115](https://github.com/cosmos/cosmos-sdk/pull/18115) `NewValidatorSigningInfo` takes strings instead of `sdk.AccAddress`
* (types) [#18268](https://github.com/cosmos/cosmos-sdk/pull/18268) Remove global setting of basedenom. Use the staking module parameter instead
* (x/auth) [#18351](https://github.com/cosmos/cosmos-sdk/pull/18351) Auth module was moved to its own go.mod `cosmossdk.io/x/auth`
* (types) [#18372](https://github.com/cosmos/cosmos-sdk/pull/18372) Removed global configuration for coin type and purpose. Setters and getters should be removed and access directly to defined types.
* (types) [#18695](https://github.com/cosmos/cosmos-sdk/pull/18695) Removed global configuration for txEncoder.
* (types) [#18607](https://github.com/cosmos/cosmos-sdk/pull/18607) Removed address verifier from global config, moved verifier function to bech32 codec.
* (server) [#18909](https://github.com/cosmos/cosmos-sdk/pull/18909) Remove configuration endpoint on grpc reflection endpoint in favour of auth module bech32prefix endpoint already exposed.
* (crypto) [#19541](https://github.com/cosmos/cosmos-sdk/pull/19541) The deprecated `FromTmProtoPublicKey`, `ToTmProtoPublicKey`, `FromTmPubKeyInterface` and `ToTmPubKeyInterface` functions have been removed. Use their replacements (`Cmt` instead of `Tm`) instead.
* (types) [#19652](https://github.com/cosmos/cosmos-sdk/pull/19652) and [#19758](https://github.com/cosmos/cosmos-sdk/pull/19758)
    * Moved`types/module.HasRegisterInterfaces` to `cosmossdk.io/core`.
    * Moved `RegisterInterfaces` and `RegisterImplementations` from `InterfaceRegistry` to `cosmossdk.io/core/registry.InterfaceRegistrar` interface.
* (types) [#19627](https://github.com/cosmos/cosmos-sdk/pull/19627) and [#19735](https://github.com/cosmos/cosmos-sdk/pull/19735) All genesis interfaces now don't take `codec.JsonCodec`.
    * Every module has the codec already, passing it created an unneeded dependency.
    * Additionally, to reflect this change, the module manager does not take a codec either.
* (runtime) [#19747](https://github.com/cosmos/cosmos-sdk/pull/19747) `runtime.ValidatorAddressCodec` and `runtime.ConsensusAddressCodec` have been moved to `core`.

### Client Breaking Changes

* (runtime) [#19040](https://github.com/cosmos/cosmos-sdk/pull/19040) Simplify app config implementation and deprecate `/cosmos/app/v1alpha1/config` query.

### CLI Breaking Changes

* (server) [#18303](https://github.com/cosmos/cosmos-sdk/pull/18303) `appd export` has moved with other genesis commands, use `appd genesis export` instead.

### Deprecated

* (simapp) [#19146](https://github.com/cosmos/cosmos-sdk/pull/19146) Replace `--v` CLI option with `--validator-count`/`-n`.
* (module) [#19370](https://github.com/cosmos/cosmos-sdk/pull/19370) Deprecate `module.Configurator`, use `appmodule.HasMigrations` and `appmodule.HasServices` instead from Core API.

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

## [v0.50.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.4) - 2023-02-19

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

## [v0.50.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.3) - 2023-01-15

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

* (baseapp) [#
](https://github.com/cosmos/cosmos-sdk/pull/19058) Fix baseapp posthandler branch would fail if the `runMsgs` had returned an error.
* (baseapp) [#18609](https://github.com/cosmos/cosmos-sdk/issues/18609) Fixed accounting in the block gas meter after module's beginBlock and before DeliverTx, ensuring transaction processing always starts with the expected zeroed out block gas meter.
* (baseapp) [#18895](https://github.com/cosmos/cosmos-sdk/pull/18895) Fix de-duplicating vote extensions during validation in ValidateVoteExtensions.

## [v0.50.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.2) - 2023-12-11

### Features

* (debug) [#18219](https://github.com/cosmos/cosmos-sdk/pull/18219) Add debug commands for application codec types.
* (client/keys) [#17639](https://github.com/cosmos/cosmos-sdk/pull/17639) Allows using and saving public keys encoded as base64.
* (server) [#17094](https://github.com/cosmos/cosmos-sdk/pull/17094) Add a `shutdown-grace` flag for waiting a given time before exit.

### Improvements

* (telemetry) [#18646](https://github.com/cosmos/cosmos-sdk/pull/18646) Enable statsd and dogstatsd telemetry sinks.
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
* (genutil) [#16046](https://github.com/cosmos/cosmos-sdk/pull/16046) Add "module-name" flag to genutil `add-genesis-account` to enable initializing module accounts at genesis.* [#15970](https://github.com/cosmos/cosmos-sdk/pull/15970) Enable SIGN_MODE_TEXTUAL.
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

* [#18204](https://github.com/cosmos/cosmos-sdk/pull/18204) Use streaming json parser to parse chain-id from genesis file.
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
* (types/query) [#16041](https://github.com/cosmos/cosmos-sdk/pull/16041) Change pagination max limit to a variable in order to be modified by application devs.
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
* (baseapp) [#16342](https://github.com/cosmos/cosmos-sdk/pull/16342) NewContext was renamed to NewContextLegacy. The replacement (NewContext) now does not take a header, instead you should set the header via `WithHeaderInfo` or `WithBlockHeight`. Note that `WithBlockHeight` will soon be deprecated and its recommended to use `WithHeaderInfo`.
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
* (x/gov) [#16118](https://github.com/cosmos/cosmos-sdk/pull/16118/) Use collections for constitution and params state management.
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

## [v0.47.10](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.10) - 2024-02-27

### Bug Fixes

* (x/staking) Fix a possible bypass of delagator slashing: [GHSA-86h5-xcpx-cfqc](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-86h5-xcpx-cfqc)
* (server) [#19573](https://github.com/cosmos/cosmos-sdk/pull/19573) Use proper `db_backend` type when reading chain-id

## [v0.47.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.9) - 2024-02-19

### Bug Fixes

* (x/auth/vesting) [GHSA-4j93-fm92-rp4m](#bug-fixes) Add `BlockedAddr` check in `CreatePeriodicVestingAccount`.
* (baseapp) [#19177](https://github.com/cosmos/cosmos-sdk/pull/19177) Fix baseapp `DefaultProposalHandler` same-sender non-sequential sequence.

## [v0.47.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.8) - 2024-01-22

### Improvements

* (client/tx) [#18852](https://github.com/cosmos/cosmos-sdk/pull/18852) Add `WithFromName` to tx factory.
* (types) [#18875](https://github.com/cosmos/cosmos-sdk/pull/18875) Speedup coins.Sort() if len(coins) <= 1.
* (types) [#18888](https://github.com/cosmos/cosmos-sdk/pull/18888) Speedup DecCoin.Sort() if len(coins) <= 1
* (testutil) [#18930](https://github.com/cosmos/cosmos-sdk/pull/18930) Add NodeURI for clientCtx.

### Bug Fixes

* [#19106](https://github.com/cosmos/cosmos-sdk/pull/19106) Allow empty public keys when setting signatures. Public keys aren't needed for every transaction.
* (server) [#18920](https://github.com/cosmos/cosmos-sdk/pull/18920) Fixes consensus failure while restart node with wrong `chainId` in genesis.

## [v0.47.7](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.7) - 2023-12-20

### Improvements

* (x/gov) [#18707](https://github.com/cosmos/cosmos-sdk/pull/18707) Improve genesis validation.
* (server) [#18478](https://github.com/cosmos/cosmos-sdk/pull/18478) Add command flag to disable colored logs.

### Bug Fixes

* (baseapp) [#18609](https://github.com/cosmos/cosmos-sdk/issues/18609) Fixed accounting in the block gas meter after BeginBlock and before DeliverTx, ensuring transaction processing always starts with the expected zeroed out block gas meter.
* (server) [#18537](https://github.com/cosmos/cosmos-sdk/pull/18537) Fix panic when defining minimum gas config as `100stake;100uatom`. Use a `,` delimiter instead of `;`. Fixes the server config getter to use the correct delimiter.
* (client/tx) [#18472](https://github.com/cosmos/cosmos-sdk/pull/18472) Utilizes the correct Pubkey when simulating a transaction.

## [v0.47.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.6) - 2023-11-14

### Features

* (server) [#18110](https://github.com/cosmos/cosmos-sdk/pull/18110) Start gRPC & API server in standalone mode.

### Improvements

* (baseapp) [#17954](https://github.com/cosmos/cosmos-sdk/issues/17954) Add `Mempool()` method on `BaseApp` to allow access to the mempool.
* (x/gov) [#17780](https://github.com/cosmos/cosmos-sdk/pull/17780) Recover panics and turn them into errors when executing x/gov proposals.

### Bug Fixes

* (server) [#18254](https://github.com/cosmos/cosmos-sdk/pull/18254) Don't hardcode gRPC address to localhost.
* (server) [#18251](https://github.com/cosmos/cosmos-sdk/pull/18251) Call `baseapp.Close()` when app started as grpc only.
* (baseapp) [#17769](https://github.com/cosmos/cosmos-sdk/pull/17769) Ensure we respect block size constraints in the `DefaultProposalHandler`'s `PrepareProposal` handler when a nil or no-op mempool is used. We provide a `TxSelector` type to assist in making transaction selection generalized. We also fix a comparison bug in tx selection when `req.maxTxBytes` is reached.
* (config) [#17649](https://github.com/cosmos/cosmos-sdk/pull/17649) Fix `mempool.max-txs` configuration is invalid in `app.config`.
* (mempool) [#17668](https://github.com/cosmos/cosmos-sdk/pull/17668) Fix `PriorityNonceIterator.Next()` nil pointer ref for min priority at the end of iteration.
* (x/auth) [#17902](https://github.com/cosmos/cosmos-sdk/pull/17902) Remove tip posthandler.
* (x/bank) [#18107](https://github.com/cosmos/cosmos-sdk/pull/18107) Add missing keypair of SendEnabled to restore legacy param set before migration.

### Client Breaking Changes

* (x/gov) [#17910](https://github.com/cosmos/cosmos-sdk/pull/17910) Remove telemetry for counting votes and proposals. It was incorrectly counting votes. Use alternatives, such as state streaming.

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

* (cli) [#16856](https://github.com/cosmos/cosmos-sdk/pull/16856) Improve `simd prune` UX by using the app default home directory and set pruning method as first variable argument (defaults to default). `pruning.PruningCmd` rest unchanged for API compatibility, use `pruning.Cmd` instead.
* (testutil) [#16704](https://github.com/cosmos/cosmos-sdk/pull/16704) Make app config configurator for testing configurable with external modules.
* (deps) [#16565](https://github.com/cosmos/cosmos-sdk/pull/16565) Bump CometBFT to [v0.37.2](https://github.com/cometbft/cometbft/blob/v0.37.2/CHANGELOG.md).

### Bug Fixes

* (x/auth) [#16994](https://github.com/cosmos/cosmos-sdk/pull/16994) Fix regression where querying transactions events with `<=` or `>=` would not work.
* (server) [#16827](https://github.com/cosmos/cosmos-sdk/pull/16827) Properly use `--trace` flag (before it was setting the trace level instead of displaying the stacktraces).
* (x/auth) [#16554](https://github.com/cosmos/cosmos-sdk/pull/16554) `ModuleAccount.Validate` now reports a nil `.BaseAccount` instead of panicking.
* [#16588](https://github.com/cosmos/cosmos-sdk/pull/16588) Propagate snapshotter failures to the caller, (it would create an empty snapshot silently before).
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
* (x/slashing, x/staking) [#14363](https://github.com/cosmos/cosmos-sdk/pull/14363) Add the infraction a validator committed type as an argument to a `SlashWithInfractionReason` keeper method.
* (client) [#14051](https://github.com/cosmos/cosmos-sdk/pull/14051) Add `--grpc` client option.
* (x/genutil) [#14149](https://github.com/cosmos/cosmos-sdk/pull/14149) Add `genutilcli.GenesisCoreCommand` command, which contains all genesis-related sub-commands.
* (x/evidence) [#13740](https://github.com/cosmos/cosmos-sdk/pull/13740) Add new proto field `hash` of type `string` to `QueryEvidenceRequest` which helps to decode the hash properly while using query API.
* (core) [#13306](https://github.com/cosmos/cosmos-sdk/pull/13306) Add a `FormatCoins` function to in `core/coins` to format sdk Coins following the Value Renderers spec.
* (math) [#13306](https://github.com/cosmos/cosmos-sdk/pull/13306) Add `FormatInt` and `FormatDec` functions in `math` to format integers and decimals following the Value Renderers spec.
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

## [v0.46.16](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.16) - 2023-11-07

EOL notice. This is the last release of the `v0.46.x` line. Per this version, the v0.46.x line reached its end-of-life.

### Bug Fixes

* (server) [#18254](https://github.com/cosmos/cosmos-sdk/pull/18254) Don't hardcode gRPC address to localhost.

## [v0.46.15](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.14) - 2023-08-21

### Improvements

* (x/gov) [#17387](https://github.com/cosmos/cosmos-sdk/pull/17387) Add `MsgSubmitProposal` `SetMsgs` method.
* (x/gov) [#17354](https://github.com/cosmos/cosmos-sdk/issues/17354) Emit `VoterAddr` in `proposal_vote` event.
* (x/genutil) [#17296](https://github.com/cosmos/cosmos-sdk/pull/17296) Add `MigrateHandler` to allow reuse migrate genesis related function.
    * In v0.46, v0.47 this function is additive to the `genesis migrate` command. However in v0.50+, adding custom migrations to the `genesis migrate` command is directly possible.

## Bug Fixes

* (server) [#17181](https://github.com/cosmos/cosmos-sdk/pull/17181) Fix `db_backend` lookup fallback from `config.toml`.

## [v0.46.14](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.14) - 2023-07-17

### Features

* (sims) [#16656](https://github.com/cosmos/cosmos-sdk/pull/16656) Add custom max gas for block for sim config with unlimited as default.

### Improvements

* (cli) [#16856](https://github.com/cosmos/cosmos-sdk/pull/16856) Improve `simd prune` UX by using the app default home directory and set pruning method as first variable argument (defaults to default). `pruning.PruningCmd` rest unchanged for API compatibility, use `pruning.Cmd` instead.
* (deps) [#16553](https://github.com/cosmos/cosmos-sdk/pull/16553) Bump CometBFT to [v0.34.29](https://github.com/cometbft/cometbft/blob/v0.34.29/CHANGELOG.md#v03429).

### Bug Fixes

* (x/auth) [#16994](https://github.com/cosmos/cosmos-sdk/pull/16994) Fix regression where querying transactions events with `<=` or `>=` would not work.
* (x/auth) [#16554](https://github.com/cosmos/cosmos-sdk/pull/16554) `ModuleAccount.Validate` now reports a nil `.BaseAccount` instead of panicking.
* [#16588](https://github.com/cosmos/cosmos-sdk/pull/16588) Propagate snapshotter failures to the caller, (it would create an empty snapshot silently before).
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
* (x/distribution) [#15462](https://github.com/cosmos/cosmos-sdk/pull/15462) Add delegator address to the event for withdrawing delegation rewards.
* [#14019](https://github.com/cosmos/cosmos-sdk/issues/14019) Remove the interface casting to allow other implementations of a `CommitMultiStore`.

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

* (store/cache) [#13881](https://github.com/cosmos/cosmos-sdk/pull/13881) Optimize iteration on nested cached KV stores and other operations in general.
* (deps) [#14846](https://github.com/cosmos/cosmos-sdk/pull/14846) Bump btcd.
* (deps) Bump Tendermint version to [v0.34.26](https://github.com/informalsystems/tendermint/releases/tag/v0.34.26).
* (store/cache) [#14189](https://github.com/cosmos/cosmos-sdk/pull/14189) Add config `iavl-lazy-loading` to enable lazy loading of iavl store, to improve start up time of archive nodes, add method `SetLazyLoading` to `CommitMultiStore` interface.
    * A new field has been added to the app.toml. This allows nodes with larger databases to startup quicker

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

* (store/cache) [#13881](https://github.com/cosmos/cosmos-sdk/pull/13881) Optimize iteration on nested cached KV stores and other operations in general.
* (x/gov) [#14347](https://github.com/cosmos/cosmos-sdk/pull/14347) Support `v1.Proposal` message in `v1beta1.Proposal.Content`.
* (deps) Use Informal System fork of Tendermint version to [v0.34.24](https://github.com/informalsystems/tendermint/releases/tag/v0.34.24).

### Bug Fixes

* (x/group) [#14526](https://github.com/cosmos/cosmos-sdk/pull/14526) Fix wrong address set in `EventUpdateGroupPolicy`.
* (ante) [#14448](https://github.com/cosmos/cosmos-sdk/pull/14448) Return anteEvents when postHandler fail.

### API Breaking Changes

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
    * For applying the patch please refer to the [RELEASE NOTES](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.3)
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
* [#13301](https://github.com/cosmos/cosmos-sdk/pull/13301) Keep the balance query endpoint compatible with legacy blocks
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
* [#11089](https://github.com/cosmos/cosmos-sdk/pull/11089) interacting with the node through `grpc.Dial` requires clients to pass a codec refer to [doc](https://docs.cosmos.network/main/user/run-node/interact-node).
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
* [#11089](https://github.com/cosmos/cosmos-sdk/pull/11089) Now cosmos-sdk consumers can upgrade gRPC to its newest versions.
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
* (cli) [#11337](https://github.com/cosmos/cosmos-sdk/pull/11337) Fixes `show-address` cli cmd
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

## [v0.45.16](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.16) - 2023-05-11

### Security Bug Fixes

* (x/feegrant) [#16097](https://github.com/cosmos/cosmos-sdk/pull/16097) Fix infinite feegrant allowance bug.

## [v0.45.15](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.15) - 2023-03-22

### Improvements

* (deps) Migrate to [CometBFT](https://github.com/cometbft/cometbft). Follow the instructions in the [release process](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.15).
* (deps) [#15127](https://github.com/cosmos/cosmos-sdk/pull/15127) Bump btcd.
* (store) [#14410](https://github.com/cosmos/cosmos-sdk/pull/14410) `rootmulti.Store.loadVersion` has validation to check if all the module stores' height is correct, it will error if any module store has incorrect height.

## [v0.45.14](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.14) - 2023-02-16

### Features

* [#14583](https://github.com/cosmos/cosmos-sdk/pull/14583) Add support for Core API.

## v0.45.13 - 2023-02-08

### Improvements

* (deps) Bump Tendermint version to [v0.34.26](https://github.com/informalsystems/tendermint/releases/tag/v0.34.26).

### Bug Fixes

* (store) [#14798](https://github.com/cosmos/cosmos-sdk/pull/14798) Copy btree to avoid the problem of modify while iteration.

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

* (store) [#12945](https://github.com/cosmos/cosmos-sdk/pull/12945) Fix nil end semantics in store/cachekv/iterator when iterating a dirty cache.
* (store) [#13516](https://github.com/cosmos/cosmos-sdk/pull/13516) Fix state listener that was observing writes at wrong time.

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

* [#13588](https://github.com/cosmos/cosmos-sdk/pull/13588) Fix regression in distribution.WithdrawDelegationRewards when rewards are zero.
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
* (store) [#13540](https://github.com/cosmos/cosmos-sdk/pull/13540) Default fastnode migration to false to prevent surprises. Operators must enable it, unless they have it enabled already.

### API Breaking Changes

* (cli) [#13089](https://github.com/cosmos/cosmos-sdk/pull/13089) Fix rollback command don't actually delete multistore versions, added method `RollbackToVersion` to interface `CommitMultiStore` and added method `CommitMultiStore` to `Application` interface.

### Bug Fixes

* Implement dragonberry security patch.
* For applying the patch please refer to the [RELEASE PROCESS](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.9)
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
* (x/bank) [#8479](https://github.com/cosmos/cosmos-sdk/pull/8479) Additional client denom metadata validation for `base` and `display` denoms.
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
    `SnapshotVersion` and `FlushVersion` accept a version argument and determine if the version should be
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
    * (x/capability) [#5828](https://github.com/cosmos/cosmos-sdk/pull/5828) Capability module integration as outlined in [ADR 3 - Dynamic Capability Store](https://docs.cosmos.network/main/build/architecture/adr-003-dynamic-capability-store).
    * (x/crisis) `x/crisis` has a new function: `AddModuleInitFlags`, which will register optional crisis module flags for the start command.
    * (x/ibc) [#5277](https://github.com/cosmos/cosmos-sdk/pull/5277) `x/ibc` changes from IBC alpha. See ICS specs below:
        * [ICS 002 - Client Semantics](https://github.com/cosmos/ibc/tree/main/spec/core/ics-002-client-semantics) subpackage
        * [ICS 003 - Connection Semantics](https://github.com/cosmos/ibc/tree/main/spec/core/ics-003-connection-semantics) subpackage
        * [ICS 004 - Channel and Packet Semantics](https://github.com/cosmos/ibc/tree/main/spec/core/ics-004-channel-and-packet-semantics) subpackage
        * [ICS 005 - Port Allocation](https://github.com/cosmos/ibc/tree/main/spec/core/ics-005-port-allocation) subpackage
        * [ICS 006 - Solo Machine Client](https://github.com/cosmos/ibc/tree/main/spec/client/ics-006-solo-machine-client) subpackage
        * [ICS 007 - Tendermint Client](https://github.com/cosmos/ibc/tree/main/spec/client/ics-007-tendermint-client) subpackage
        * [ICS 009 - Loopback Client](https://github.com/cosmos/ibc/tree/main/spec/client/ics-009-loopback-cilent) subpackage
        * [ICS 020 - Fungible Token Transfer](https://github.com/cosmos/ibc/tree/main/spec/app/ics-020-fungible-token-transfer) subpackage
        * [ICS 023 - Vector Commitments](https://github.com/cosmos/ibc/tree/main/spec/core/ics-023-vector-commitments) subpackage
        * [ICS 024 - Host State Machine Requirements](https://github.com/cosmos/ibc/tree/main/spec/core/ics-024-host-requirements) subpackage
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
        * Allow uppercase letters and numbers in denominations to support ADR 001
        * Added `Validate` function that returns a descriptive error
    * (types) [#5581](https://github.com/cosmos/cosmos-sdk/pull/5581) Add convenience functions {,Must}Bech32ifyAddressBytes.
    * (types/module) [#5724](https://github.com/cosmos/cosmos-sdk/issues/5724) The `types/module` package does no longer depend on `x/simulation`.
    * (types) [#5585](https://github.com/cosmos/cosmos-sdk/pull/5585) IBC additions:
        * `Coin` denomination max length has been increased to 32.
        * Added `CapabilityKey` alias for `StoreKey` to match IBC spec.
    * (types/rest) [#5900](https://github.com/cosmos/cosmos-sdk/pull/5900) Add Check\*Error function family to spare developers from replicating tons of boilerplate code.
    * (types) [#6128](https://github.com/cosmos/cosmos-sdk/pull/6137) Add `String()` method to `GasMeter`.
    * (types) [#6195](https://github.com/cosmos/cosmos-sdk/pull/6195) Add codespace to broadcast(sync/async) response.
    * (types) [#6897](https://github.com/cosmos/cosmos-sdk/issues/6897) Add KV type from tendermint to `types` directory.
    * (version) [#7848](https://github.com/cosmos/cosmos-sdk/pull/7848) [#7941](https://github.com/cosmos/cosmos-sdk/pull/7941)
    `version --long` output now shows the list of build dependencies and replaced build dependencies.

## Previous Releases

[CHANGELOG of previous versions](https://github.com/cosmos/cosmos-sdk/blob/c17c3caab86a1426a1eef4541e8203f5f54a1a54/CHANGELOG.md#v0391---2020-08-11) (pre Stargate).
