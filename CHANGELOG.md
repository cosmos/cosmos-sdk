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

## [v0.47.16](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.16) - 2025-02-20

### Bug Fixes

<<<<<<< HEAD
* [GHSA-x5vx-95h7-rv4p](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-x5vx-95h7-rv4p) Fix Group module can halt chain when handling a malicious proposal
=======
* (x/auth) [#23741](https://github.com/cosmos/cosmos-sdk/pull/23741) Support legacy global AccountNumber.
* (types/query) [#23880](https://github.com/cosmos/cosmos-sdk/pull/23880) Fix NPE in query pagination.

### Removed

* (tools/hub) [#23562](https://github.com/cosmos/cosmos-sdk/pull/23562) Remove `tools/hubl`. A similar tool will be maintained in [ignite](https://www.github.com/ignite/cli).

### API Breaking Changes

* (x/params) [#22995](https://github.com/cosmos/cosmos-sdk/pull/22995) Remove `x/params`.  Migrate to the new params system introduced in `v0.47` as demonstrated [here](https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md#xparams).
* (testutil) [#22392](https://github.com/cosmos/cosmos-sdk/pull/22392) Remove `testutil/network` package. Use the integration framework or systemtests framework instead.

#### Removal of v0 components

This subsection lists the API breaking changes that are [part of the removal of v0 components](https://github.com/cosmos/cosmos-sdk/issues/22904). The v0 components were deprecated in `v0.52` and are now removed.

* (simapp) [#23009](https://github.com/cosmos/cosmos-sdk/pull/23009) Simapp has been removed. Check-out Simapp/v2 instead.
* (server) [#23018](https://github.com/cosmos/cosmos-sdk/pull/23018) [#23238](https://github.com/cosmos/cosmos-sdk/pull/23238) The server package has been removed. Use server/v2 instead
* (x/genutil) [#23238](https://github.com/cosmos/cosmos-sdk/pull/23238) Genutil commands specific to a baseapp chain have been deleted.
* (client) [#22904](https://github.com/cosmos/cosmos-sdk/issues/22904) v1 specific client commands have been removed.

## [v0.52.0-rc.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.52.0-rc.2) - 2025-01-23

Every module contains its own CHANGELOG.md. Please refer to the module you are interested in.

### Features

* (sims) [#23013](https://github.com/cosmos/cosmos-sdk/pull/23013) Integration with app v2
* (x/auth/ante) [#23128](https://github.com/cosmos/cosmos-sdk/pull/23128) Allow custom verifyIsOnCurve when validate tx for public key like ethsecp256k1.
* (server) [#23321](https://github.com/cosmos/cosmos-sdk/pull/23321) Add custom rollback command option. In order to use it, you need to implement the Rollback interface and remove the default rollback command with `cmd.RemoveCommand(cmd.RollbackCmd)` and then add it back with `cmd.AddCommand(cmd.NewCustomRollbackCmd(appCreator, rollbackable))`.

### Improvements

* [#23470](https://github.com/cosmos/cosmos-sdk/pull/23470) Converge to use of one single sign mode type and signer data:
  * Use api's signmode throughout the SDK to align with `cosmossdk.io/tx`. This allows developer not to juggle between sign mode types
  * Deprecate `authsigning.SignerData` in favor of txsigning.SignerData and replace its usage
  * Remove `APISignModeToInternal` from `x/auth` as no conversion is necessary by the user anymore
* (all) [#23445](https://github.com/cosmos/cosmos-sdk/pull/23445) Remove `v2` code from codebase.
* (codec) [#22988](https://github.com/cosmos/cosmos-sdk/pull/22988) Improve edge case handling for recursion limits.
* (proto) [#23437](https://github.com/cosmos/cosmos-sdk/pull/23437) Deprecate `Block` field from `GetBlockByHeightResponse` and return empty comet block for `GetBlockByHeight`.
* (module) [#23488](https://github.com/cosmos/cosmos-sdk/pull/23488) Remove CoreAppModuleAdaptor which is no longer used and add HasRegisterServices public interface.

### Bug Fixes

* (x/auth/tx) [#23492](https://github.com/cosmos/cosmos-sdk/pull/23492) Add missing timeoutTimestamp in newBuilderFromDecodedTx.
* (query) [#23002](https://github.com/cosmos/cosmos-sdk/pull/23002) Fix collection filtered pagination.
* (x/auth/tx) [#23170](https://github.com/cosmos/cosmos-sdk/pull/23170) Avoid panic from `newWrapperFromDecodedTx` when `AuthInfo.Fee` is optional in decodedTx.
* (x/auth/tx) [#23144](https://github.com/cosmos/cosmos-sdk/pull/23144) Add missing `CacheWithValue` for `ExtensionOptions`.
* (x/auth/tx) [#23148](https://github.com/cosmos/cosmos-sdk/pull/23148) Avoid panic from `intoAnyV2` when v1.PublicKey is optional.
* (server) [#23244](https://github.com/cosmos/cosmos-sdk/pull/23244) Allow align block header with skip check header in grpc server.
* (x/auth) [#23357](https://github.com/cosmos/cosmos-sdk/pull/23357) Fixes accessibility of the AddressStringToBytes HTTP binding and adds another binding to AddressBytesToString.

### Deprecated

* (modules) [#22994](https://github.com/cosmos/cosmos-sdk/pull/22994) Deprecate `Invariants` and associated methods.

## [v0.52.0-rc.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.52.0-rc.1) - 2024-12-18

Every module contains its own CHANGELOG.md. Please refer to the module you are interested in.

### Features

* (client) [#17513](https://github.com/cosmos/cosmos-sdk/pull/17513) Allow overwriting `client.toml`. Use `client.CreateClientConfig` in place of `client.ReadFromClientConfig` and provide a custom template and a custom config.
* (tests) [#17868](https://github.com/cosmos/cosmos-sdk/pull/17868) Added helper method `SubmitTestTx` in testutil to broadcast test txns to test e2e tests.
* (client) [#18101](https://github.com/cosmos/cosmos-sdk/pull/18101) Add a `keyring-default-keyname` in `client.toml` for specifying a default key name, and skip the need to use the `--from` flag when signing transactions.
* (runtime) [#18475](https://github.com/cosmos/cosmos-sdk/pull/18475) Adds an implementation for core.branch.Service.
* (baseapp) [#18499](https://github.com/cosmos/cosmos-sdk/pull/18499) Add `MsgRouter` response type from message name function.
* (client) [#18557](https://github.com/cosmos/cosmos-sdk/pull/18557) Add `--qrcode` flag to `keys show` command to support displaying keys address QR code.
* (types) [#18768](https://github.com/cosmos/cosmos-sdk/pull/18768) Add MustValAddressFromBech32 function.
* (gRPC) [#19049](https://github.com/cosmos/cosmos-sdk/pull/19049) Add debug log prints for each gRPC request.
* (types) [#19164](https://github.com/cosmos/cosmos-sdk/pull/19164) Add a ValueCodec for the math.Uint type that can be used in collections maps.
* (types) [#19281](https://github.com/cosmos/cosmos-sdk/pull/19281) Added a new method, `IsGT`, for `types.Coin`. This method is used to check if a `types.Coin` is greater than another `types.Coin`.
* (runtime) [#19571](https://github.com/cosmos/cosmos-sdk/pull/19571) Implement `core/router.Service` in runtime. This service is present in all modules (when using depinject).
* (types) [#19759](https://github.com/cosmos/cosmos-sdk/pull/19759) Align SignerExtractionAdapter in PriorityNonceMempool Remove.
* (client) [#19870](https://github.com/cosmos/cosmos-sdk/pull/19870) Add new query command `wait-tx`. Alias `event-query-tx-for` to `wait-tx` for backward compatibility.
* (client) [#19905](https://github.com/cosmos/cosmos-sdk/pull/19905) Add grpc client config to `client.toml`.
* (genutil) [#19971](https://github.com/cosmos/cosmos-sdk/pull/19971) Allow manually setting the consensus key type in genesis
* (runtime) [#19953](https://github.com/cosmos/cosmos-sdk/pull/19953) Implement `core/transaction.Service` in runtime.
* (runtime) [#18475](https://github.com/cosmos/cosmos-sdk/pull/18475) Adds an implementation for `core.branch.Service`.
* (runtime) [#19004](https://github.com/cosmos/cosmos-sdk/pull/19004) Adds an implementation for `core/header.Service` in runtime.
* (runtime) [#20238](https://github.com/cosmos/cosmos-sdk/pull/20238) Adds an implementation for `core/comet.Service` in runtime.
* (tests) [#20013](https://github.com/cosmos/cosmos-sdk/pull/20013) Introduce system tests to run multi node local testnet in CI
* (crypto/keyring) [#20212](https://github.com/cosmos/cosmos-sdk/pull/20212) Expose the db keyring used in the keystore.
* (client/tx) [#20870](https://github.com/cosmos/cosmos-sdk/pull/20870) Add `timeout-timestamp` field for tx body defines time based timeout.Add `WithTimeoutTimestamp` to tx factory. Increased gas cost for processing newly added timeout timestamp field in tx body.
* (client) [#21074](https://github.com/cosmos/cosmos-sdk/pull/21074) Add auto cli for node service
* (x/validate) [#21822](https://github.com/cosmos/cosmos-sdk/pull/21822) New module solely responsible for providing ante/post handlers and tx validators for v2. It can be extended by the app developer to provide extra tx validators.
    * In comparison to x/auth/tx/config, there is no app config to skip ante/post handlers, as overwriting them in baseapp or not injecting the x/validate module has the same effect.
* (baseapp) [#21979](https://github.com/cosmos/cosmos-sdk/pull/21979) Create CheckTxHandler to allow extending the logic of CheckTx.

### Improvements

* RocksDB libraries have been upgraded to support RockDB v9 instead of v8.
* (all) [#16537](https://github.com/cosmos/cosmos-sdk/pull/16537) Properly propagated `fmt.Errorf` errors and using `errors.New` where appropriate.
* (client) [#17503](https://github.com/cosmos/cosmos-sdk/pull/17503) Add `client.Context{}.WithAddressCodec`, `WithValidatorAddressCodec`, `WithConsensusAddressCodec` to provide address codecs to the client context. See the [UPGRADING.md](./UPGRADING.md) for more details.
* (crypto/keyring) [#17503](https://github.com/cosmos/cosmos-sdk/pull/17503) Simplify keyring interfaces to use `[]byte` instead of `sdk.Address` for addresses.
* (rpc) [#17470](https://github.com/cosmos/cosmos-sdk/pull/17470) Avoid open 0.0.0.0 to public by default and add `listen-ip-address` argument for `testnet init-files` cmd.
* (types) [#17670](https://github.com/cosmos/cosmos-sdk/pull/17670) Use `ctx.CometInfo` in place of `ctx.VoteInfos`
* [#17733](https://github.com/cosmos/cosmos-sdk/pull/17733) Ensure `buf export` exports all proto dependencies
* (crypto/keys) [#18026](https://github.com/cosmos/cosmos-sdk/pull/18026) Made public key generation constant time on `secp256k1`
* (crypto | x/auth) [#14372](https://github.com/cosmos/cosmos-sdk/pull/18194) Key checks on signatures antehandle.
* (types) [#18440](https://github.com/cosmos/cosmos-sdk/pull/18440) Add `AmountOfNoValidation` to `sdk.DecCoins`.
* (client/keys) [#18663](https://github.com/cosmos/cosmos-sdk/pull/18663) Improve `<appd> keys add` by displaying mnemonic discreetly on an alternate screen and adding `--indiscreet` option to disable it.
* (client/keys) [#18684](https://github.com/cosmos/cosmos-sdk/pull/18684) Improve `<appd> keys export` by displaying unarmored hex private key discreetly on an alternate screen and adding `--indiscreet` option to disable it.
* (client/keys) [#18687](https://github.com/cosmos/cosmos-sdk/pull/18687) Improve `<appd> keys mnemonic` by displaying mnemonic discreetly on an alternate screen and adding `--indiscreet` option to disable it.
* (client/keys) [#18703](https://github.com/cosmos/cosmos-sdk/pull/18703) Improve `<appd> keys add` and `<appd> keys show` by checking whether there are duplicate keys in the multisig case.
    * Usage of `Must...` kind of functions are avoided in keeper methods.
* (client/keys) [#18743](https://github.com/cosmos/cosmos-sdk/pull/18743) Improve `<appd> keys add -i` by hiding inputting of bip39 passphrase.
* (client/keys) [#18745](https://github.com/cosmos/cosmos-sdk/pull/18745) Improve `<appd> keys export` and `<appd> keys mnemonic` by adding --yes option to skip interactive confirmation.
* (client/keys) [#18950](https://github.com/cosmos/cosmos-sdk/pull/18950) Improve `<appd> keys add`, `<appd> keys import` and `<appd> keys rename` by checking name validation.
* (types) [#18963](https://github.com/cosmos/cosmos-sdk/pull/18963) Swap out amino json encoding of `ABCIMessageLogs` for std lib json encoding
* (types) [#19512](https://github.com/cosmos/cosmos-sdk/pull/19512) The notion of basic manager does not exist anymore (and all related helpers).
    * The module manager now can do everything that the basic manager was doing.
    * `AppModuleBasic` has been deprecated for extension interfaces.
    * Modules can now implement `appmodule.HasRegisterInterfaces`, `module.HasGRPCGateway` and `module.HasAminoCodec` when relevant.
    * SDK modules now directly implement those extension interfaces on `AppModule` instead of `AppModuleBasic`.
* (server) [#19455](https://github.com/cosmos/cosmos-sdk/pull/19455) Allow calling back into the application struct in PostSetup.
* (types) [#19672](https://github.com/cosmos/cosmos-sdk/pull/19672) `PreBlock` now returns only an error for consistency with server/v2. The SDK has upgraded x/upgrade accordingly. `ResponsePreBlock` hence has been removed.
* (x/auth) [#19651](https://github.com/cosmos/cosmos-sdk/pull/19651) Allow empty public keys in `GetSignBytesAdapter`.
* (x/genutil) [#19735](https://github.com/cosmos/cosmos-sdk/pull/19735) Update genesis api to match new `appmodule.HasGenesis` interface.
* (types) [#19869](https://github.com/cosmos/cosmos-sdk/pull/19869) Removed `Any` type from `codec/types` and replaced it with an alias for `cosmos/gogoproto/types/any`.
* (server) [#19854](https://github.com/cosmos/cosmos-sdk/pull/19854) Add customizability to start command.
    * Add `StartCmdOptions` in `server.AddCommands` instead of `servertypes.ModuleInitFlags`. To set custom flags set them in the `StartCmdOptions` struct on the `AddFlags` field.
    * Add `StartCommandHandler` to `StartCmdOptions` to allow custom start command handlers. Users now have total control over how the app starts.
* (server) [#19966](https://github.com/cosmos/cosmos-sdk/pull/19966) Return BlockHeader by shallow copy in server Context.
* (codec) [#20122](https://github.com/cosmos/cosmos-sdk/pull/20122) Added a cache to address codec.
* (proto) [#20098](https://github.com/cosmos/cosmos-sdk/pull/20098) Use cosmos_proto added_in annotation instead of // Since comments.
* (baseapp) [#20208](https://github.com/cosmos/cosmos-sdk/pull/20208) Skip running validateBasic for rechecking txs.
* (baseapp) [#20380](https://github.com/cosmos/cosmos-sdk/pull/20380) Enhanced OfferSnapshot documentation.
* (client) [#20771](https://github.com/cosmos/cosmos-sdk/pull/20771) Remove `ReadDefaultValuesFromDefaultClientConfig` from `client` package. (It was introduced in `v0.50.6` as a quick fix).
* (grpcserver) [#20945](https://github.com/cosmos/cosmos-sdk/pull/20945) Adds error handling for out-of-gas panics in grpc query handlers.
* (internal) [#21412](https://github.com/cosmos/cosmos-sdk/pull/21412) Using unsafe.String and unsafe.SliceData.
* (client) [#21436](https://github.com/cosmos/cosmos-sdk/pull/21436) Use `address.Codec` from client.Context in `tx.Sign`.
* (x/genutil) [#21249](https://github.com/cosmos/cosmos-sdk/pull/21249) Incremental JSON parsing for AppGenesis where possible.
* (genutil) [#21701](https://github.com/cosmos/cosmos-sdk/pull/21701) Improved error messages for genesis validation.
* (runtime) [#21704](https://github.com/cosmos/cosmos-sdk/pull/21704) Move `upgradetypes.StoreLoader` to runtime and alias it in upgrade for backward compatibility.
* (sims)[#21613](https://github.com/cosmos/cosmos-sdk/pull/21613) Add sims2 framework and factory methods for simpler message factories in modules
* (modules) [#21963](https://github.com/cosmos/cosmos-sdk/pull/21963) Duplicatable metrics are no more collected in modules. They were unnecessary overhead.
* (crypto/ledger) [#22116](https://github.com/cosmos/cosmos-sdk/pull/22116) Improve error message when deriving paths using index >100
* (testutil/integration) [#22616](https://github.com/cosmos/cosmos-sdk/pull/22616) Remove double context in integration tests v1.
    * Use `integrationApp.Context()` instead of creating a context prior.
* (version) [#22807](https://github.com/cosmos/cosmos-sdk/pull/22807) Return server/v2 information in the `version` functions and commands.
* [#22826](https://github.com/cosmos/cosmos-sdk/pull/22826) Simplify testing frameworks by removing `testutil/cmdtest`.

### Bug Fixes

* (baseapp) [#18383](https://github.com/cosmos/cosmos-sdk/pull/18383) Fixed a data race inside BaseApp.getContext, found by end-to-end (e2e) tests.
* (client/server) [#18345](https://github.com/cosmos/cosmos-sdk/pull/18345) Consistently set viper prefix in client and server. It defaults for the binary name for both client and server.
* (baseapp) [#18551](https://github.com/cosmos/cosmos-sdk/pull/18551) Fix SelectTxForProposal the calculation method of tx bytes size is inconsistent with CometBFT
* (client/keys) [#18562](https://github.com/cosmos/cosmos-sdk/pull/18562) `keys delete` won't terminate when a key is not found.
* (client) [#18622](https://github.com/cosmos/cosmos-sdk/pull/18622) Fixed a potential under/overflow from `uint64->int64` when computing gas fees as a LegacyDec.
* (baseapp) [#18727](https://github.com/cosmos/cosmos-sdk/pull/18727) Ensure that `BaseApp.Init` firstly returns any errors from a nil commit multistore instead of panicking on nil dereferencing and before sealing the app.
* (server) [#18994](https://github.com/cosmos/cosmos-sdk/pull/18994) Update server context directly rather than a reference to a sub-object
* [#19833](https://github.com/cosmos/cosmos-sdk/pull/19833) Fix some places in which we call Remove inside a Walk.
* [#19851](https://github.com/cosmos/cosmos-sdk/pull/19851) Fix some places in which we call Remove inside a Walk (x/staking and x/gov).
* (sims) [#21952](https://github.com/cosmos/cosmos-sdk/pull/21952) Use liveness matrix for validator sign status in sims
* (baseapp) [#21003](https://github.com/cosmos/cosmos-sdk/pull/21003) Align block header when query with latest height.
* (sims) [#21906](https://github.com/cosmos/cosmos-sdk/pull/21906) Skip sims test when running dry on validators
* (cli) [#21919](https://github.com/cosmos/cosmos-sdk/pull/21919) Query address-by-acc-num by account_id instead of id.
* (cli) [#22656](https://github.com/cosmos/cosmos-sdk/pull/22656) Prune cmd should disable async pruning.

### API Breaking Changes

* (baseapp) [#16244](https://github.com/cosmos/cosmos-sdk/pull/16244) `SetProtocolVersion` has been renamed to `SetAppVersion`. It now updates the consensus params in baseapp's `ParamStore`.
* (types) [#16918](https://github.com/cosmos/cosmos-sdk/pull/16918), [#22925](https://github.com/cosmos/cosmos-sdk/pull/22925) Deprecate `IntProto` and `DecProto`. Instead, `math.Int` and `math.LegacyDec` should be used respectively. Both types support `Marshal` and `Unmarshal` which should be used for binary marshaling.
* (client) [#17215](https://github.com/cosmos/cosmos-sdk/pull/17215) `server.StartCmd`,`server.ExportCmd`,`server.NewRollbackCmd`,`pruning.Cmd`,`genutilcli.InitCmd`,`genutilcli.GenTxCmd`,`genutilcli.CollectGenTxsCmd`,`genutilcli.AddGenesisAccountCmd`, do not take a home directory anymore. It is inferred from the root command.
* (client) [#17259](https://github.com/cosmos/cosmos-sdk/pull/17259) Remove deprecated `clientCtx.PrintObjectLegacy`. Use `clientCtx.PrintProto` or `clientCtx.PrintRaw` instead.
* (types) [#17348](https://github.com/cosmos/cosmos-sdk/pull/17348) Remove the `WrapServiceResult` function.
    * The `*sdk.Result` returned by the msg server router will not contain the `.Data` field. 
* (types) [#17426](https://github.com/cosmos/cosmos-sdk/pull/17426) `NewContext` does not take a `cmtproto.Header{}` any longer.
    * `WithChainID` / `WithBlockHeight` / `WithBlockHeader` must be used to set values on the context
* (client/keys) [#17503](https://github.com/cosmos/cosmos-sdk/pull/17503) `clientkeys.NewKeyOutput`, `MkConsKeyOutput`, `MkValKeyOutput`, `MkAccKeyOutput`, `MkAccKeysOutput` now take their corresponding address codec instead of using the global SDK config.
* (types/simulation) [#17737](https://github.com/cosmos/cosmos-sdk/pull/17737) Remove unused parameter from `RandomFees`
* (types) [#17738](https://github.com/cosmos/cosmos-sdk/pull/17738) `WithBlockTime()` was removed & `BlockTime()` were deprecated in favor of `WithHeaderInfo()` & `HeaderInfo()`. `BlockTime` now gets data from `HeaderInfo()` instead of `BlockHeader()`.
* (client) [#17746](https://github.com/cosmos/cosmos-sdk/pull/17746) `txEncodeAmino` & `txDecodeAmino` txs via grpc and rest were removed
* (app) [#17838](https://github.com/cosmos/cosmos-sdk/pull/17838) Params module was removed from simapp and all imports of the params module removed throughout the repo.
    * The Cosmos SDK has migrated away from using params, if your app still uses it, then you can leave it plugged into your app
* (x/bank/testutil) [#17868](https://github.com/cosmos/cosmos-sdk/pull/17868) `MsgSendExec` has been removed because of AutoCLI migration.
* (types) [#17885](https://github.com/cosmos/cosmos-sdk/pull/17885) `InitGenesis` & `ExportGenesis` now take `context.Context` instead of `sdk.Context`
* (x/gov/testutil) [#17986](https://github.com/cosmos/cosmos-sdk/pull/18036) `MsgDeposit` has been removed because of AutoCLI migration.
* (x/staking/testutil) [#17986](https://github.com/cosmos/cosmos-sdk/pull/17986) `MsgRedelegateExec`, `MsgUnbondExec` has been removed because of AutoCLI migration.
* (x/group) [#17937](https://github.com/cosmos/cosmos-sdk/pull/17937) Groups module was moved to its own go.mod `cosmossdk.io/x/group`
* (x/gov) [#18197](https://github.com/cosmos/cosmos-sdk/pull/18197) Gov module was moved to its own go.mod `cosmossdk.io/x/gov`
* (x/distribution) [#18199](https://github.com/cosmos/cosmos-sdk/pull/18199) Distribution module was moved to its own go.mod `cosmossdk.io/x/distribution`
* (x/slashing) [#18201](https://github.com/cosmos/cosmos-sdk/pull/18201) Slashing module was moved to its own go.mod `cosmossdk.io/x/slashing`
* (x/staking) [#18257](https://github.com/cosmos/cosmos-sdk/pull/18257) Staking module was moved to its own go.mod `cosmossdk.io/x/staking`
* (types) [#18268](https://github.com/cosmos/cosmos-sdk/pull/18268) Remove global setting of basedenom. Use the staking module parameter instead
* (x/authz) [#18265](https://github.com/cosmos/cosmos-sdk/pull/18265) Authz module was moved to its own go.mod `cosmossdk.io/x/authz`
* (x/mint) [#18283](https://github.com/cosmos/cosmos-sdk/pull/18283) Mint module was moved to its own go.mod `cosmossdk.io/x/mint`
* (server) [#18303](https://github.com/cosmos/cosmos-sdk/pull/18303) `x/genutil` now handles the application export. `server.AddCommands` does not take an `AppExporter` but instead `genutilcli.Commands` does.
* (types) [#18372](https://github.com/cosmos/cosmos-sdk/pull/18372) Removed global configuration for coin type and purpose. Setters and getters should be removed and access directly to defined types.
* (types) [#18607](https://github.com/cosmos/cosmos-sdk/pull/18607) Removed address verifier from global config, moved verifier function to bech32 codec.
* (types) [#18695](https://github.com/cosmos/cosmos-sdk/pull/18695) Removed global configuration for txEncoder.
* (server) [#18909](https://github.com/cosmos/cosmos-sdk/pull/18909) Remove configuration endpoint on grpc reflection endpoint in favour of auth module bech32prefix endpoint already exposed.
* (crypto) [#19541](https://github.com/cosmos/cosmos-sdk/pull/19541) The deprecated `FromTmProtoPublicKey`, `ToTmProtoPublicKey`, `FromTmPubKeyInterface` and `ToTmPubKeyInterface` functions have been removed. Use their replacements (`Cmt` instead of `Tm`) instead.
* (types) [#19512](https://github.com/cosmos/cosmos-sdk/pull/19512) Remove basic manager and all related functions (`module.BasicManager`, `module.NewBasicManager`, `module.NewBasicManagerFromManager`, `NewGenesisOnlyAppModule`).
    * The module manager now can do everything that the basic manager was doing.
    * When using runtime, just inject the module manager when needed using your app config.
    * All `AppModuleBasic` structs have been removed.
* (types) [#19627](https://github.com/cosmos/cosmos-sdk/pull/19627) and [#19735](https://github.com/cosmos/cosmos-sdk/pull/19735) All genesis interfaces now don't take `codec.JsonCodec`:
    * Every module has the codec already, passing it created an unneeded dependency.
    * Additionally, to reflect this change, the module manager does not take a codec either.
* (types) [#19652](https://github.com/cosmos/cosmos-sdk/pull/19652) and [#19758](https://github.com/cosmos/cosmos-sdk/pull/19758)
    * Moved`types/module.HasRegisterInterfaces` to `cosmossdk.io/core`.
    * Moved `RegisterInterfaces` and `RegisterImplementations` from `InterfaceRegistry` to `cosmossdk.io/core/registry.InterfaceRegistrar` interface.
* (types) [#19742](https://github.com/cosmos/cosmos-sdk/pull/19742) Removes the use of `Accounts.String`
    * `SimulationState` now has address and validator codecs as fields.
* (runtime) [#19747](https://github.com/cosmos/cosmos-sdk/pull/19747) `runtime.ValidatorAddressCodec` and `runtime.ConsensusAddressCodec` have been moved to `core`.
* (all) [#19726](https://github.com/cosmos/cosmos-sdk/pull/19726) Integrate comet v1
* [#19833](https://github.com/cosmos/cosmos-sdk/pull/19833) Fix some places in which we call Remove inside a Walk.
* [#19839](https://github.com/cosmos/cosmos-sdk/pull/19839) `Tx.GetMsgsV2` has been replaced with `Tx.GetReflectMessages`, and `Codec.GetMsgV1Signers` and `Codec.GetMsgV2Signers` have been replaced with `GetMsgSigners` and `GetReflectMsgSigners` respectively. These API changes clear up confusion as to the use and purpose of these methods.
* [#19851](https://github.com/cosmos/cosmos-sdk/pull/19851) Fix some places in which we call Remove inside a Walk (x/staking and x/gov).
* (server) [#19854](https://github.com/cosmos/cosmos-sdk/pull/19854) Remove `servertypes.ModuleInitFlags` types and from `server.AddCommands` as `StartCmdOptions` already achieves the same goal.
* (x/genutil) [#19926](https://github.com/cosmos/cosmos-sdk/pull/19926) Removal of the `Address.String()` method and related changes:
    * Added an address codec as an argument to `CollectTxs`, `GenAppStateFromConfig`, and `AddGenesisAccount`.
    * Removed the `ValidatorAddressCodec` argument from `CollectGenTxsCmd`, now utilizing the context for this purpose.
    * Changed `ValidateAccountInGenesis` to accept a string instead of an `AccAddress`.
* (baseapp) [#19993](https://github.com/cosmos/cosmos-sdk/pull/19993) Indicate pruning with error code "not found" rather than "invalid request".
* (x/consensus) [#20010](https://github.com/cosmos/cosmos-sdk/pull/20010) Move consensus module to be its own go.mod
* (x/crisis) [#20043](https://github.com/cosmos/cosmos-sdk/pull/20043) Changed `NewMsgVerifyInvariant` to accept a string as argument instead of an `AccAddress`.
* (x/simulation)[#20056](https://github.com/cosmos/cosmos-sdk/pull/20056) `SimulateFromSeed` now takes an address codec as argument.
* (server) [#20140](https://github.com/cosmos/cosmos-sdk/pull/20140) Remove embedded grpc-web proxy in favor of standalone grpc-web proxy. [Envoy Proxy](https://www.envoyproxy.io/docs/envoy/latest/start/start)
* (client) [#20255](https://github.com/cosmos/cosmos-sdk/pull/20255) Use comet proofOp proto type instead of sdk version to avoid needing to translate to later be proven in the merkle proof runtime. 
* (types)[#20369](https://github.com/cosmos/cosmos-sdk/pull/20369) The signature of `HasAminoCodec` has changed to accept a `core/legacy.Amino` interface instead of `codec.LegacyAmino`.
* (server) [#20422](https://github.com/cosmos/cosmos-sdk/pull/20422) Deprecated `ServerContext`. To get `cmtcfg.Config` from cmd, use `client.GetCometConfigFromCmd(cmd)` instead of `server.GetServerContextFromCmd(cmd).Config`
* (x/genutil) [#20740](https://github.com/cosmos/cosmos-sdk/pull/20740) Update `genutilcli.Commands` and `genutilcli.CommandsWithCustomMigrationMap` to take the genesis module and abstract the module manager.
* (types/errors) [#20756](https://github.com/cosmos/cosmos-sdk/pull/20756) Remove `ResponseCheckTxWithEvents`, `ResponseExecTxResultWithEvents` & `QueryResult` from types/errors pkg. They have been moved to `baseapp/errors.go` and made private.
* (client) [#20976](https://github.com/cosmos/cosmos-sdk/pull/20976) Simplified command initialization by removing unnecessary parameters such as `txConfig` and `addressCodec`.
    * Remove parameter `txConfig` from `genutilcli.Commands`,`genutilcli.CommandsWithCustomMigrationMap`,`genutilcli.GenTxCmd`.
    * Remove parameter `addressCodec` from `genutilcli.GenTxCmd`,`genutilcli.AddGenesisAccountCmd`,`stakingcli.BuildCreateValidatorMsg`.
* (sims) [#21039](https://github.com/cosmos/cosmos-sdk/pull/21039): Remove Baseapp from sims by a new interface `simtypes.AppEntrypoint`.
* (x/genutil) [#21372](https://github.com/cosmos/cosmos-sdk/pull/21372) Remove `AddGenesisAccount` for `AddGenesisAccounts`.
* (baseapp) [#21413](https://github.com/cosmos/cosmos-sdk/pull/21413) Add `SelectBy` method to `Mempool` interface, which is thread-safe to use.
* (types/mempool) [#21744](https://github.com/cosmos/cosmos-sdk/pull/21744) Update types/mempool.Mempool interface to take decoded transactions. This avoid to decode the transaction twice.
* (x/auth/tx/config)  [#21822](https://github.com/cosmos/cosmos-sdk/pull/21822) Sign mode textual is no more automatically added to tx config when using runtime. Should be added manually on the server side.
* (x/auth/tx/config)  [#21822](https://github.com/cosmos/cosmos-sdk/pull/21822) This depinject module now only provide txconfig and tx config options. `x/validate` now handles the providing of ante/post handlers, alongside tx validators for v2. The corresponding app config options have been removed from the depinject module config.
* (x/crisis) [#20809](https://github.com/cosmos/cosmos-sdk/pull/20809) Crisis module was removed from the Cosmos SDK.
* (client) [#22775](https://github.com/cosmos/cosmos-sdk/pull/22775) Removed client prompt validations.

### Client Breaking Changes

* (runtime) [#19040](https://github.com/cosmos/cosmos-sdk/pull/19040) Simplify app config implementation and deprecate `/cosmos/app/v1alpha1/config` query. 

### CLI Breaking Changes

* (server) [#18303](https://github.com/cosmos/cosmos-sdk/pull/18303) `appd export` has moved with other genesis commands, use `appd genesis export` instead.
* (perf)[#20490](https://github.com/cosmos/cosmos-sdk/pull/20490) Sims: Replace runsim command with Go stdlib testing. CLI: `Commit` default true, `Lean`, `SimulateEveryOperation`, `PrintAllInvariants`, `DBBackend` params removed
* (client/tx) [#20870](https://github.com/cosmos/cosmos-sdk/pull/20870) Removed `timeout-height` flag replace with `timeout-timestamp` flag for a time based timeout.

### Deprecated

* (simapp) [#19146](https://github.com/cosmos/cosmos-sdk/pull/19146) Replace `--v` CLI option with `--validator-count`/`-n`.
* (module) [#19370](https://github.com/cosmos/cosmos-sdk/pull/19370) Deprecate `module.Configurator`, use `appmodule.HasMigrations` and `appmodule.HasServices` instead from Core API.
* (types) [#21435](https://github.com/cosmos/cosmos-sdk/pull/21435) The `String()` method on `AccAddress`, `ValAddress` and `ConsAddress` have been deprecated. This is done because those are still using the deprecated global `sdk.Config`. Use an `address.Codec` instead.

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

## Features

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
>>>>>>> 44bd60f7e (fix: Fix npe in pagination (#23880))

## [v0.47.15](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.15) - 2024-12-16

### Bug Fixes

* Bump `cosmossdk.io/math` to v1.4.
* Fix [ABS-0043/ABS-0044](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-8wcc-m6j2-qxvm) Limit recursion depth for unknown field detection and unpack any

## [v0.47.14](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.14) - 2024-09-20

### Improvements

* [#21295](https://github.com/cosmos/cosmos-sdk/pull/21295) Bump to gogoproto v1.7.0.
* [#21295](https://github.com/cosmos/cosmos-sdk/pull/21295) Remove usage of `slices.SortFunc` due to API break in used versions.

## [v0.47.13](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.13) - 2024-07-15

### Bug Fixes

* (client) [#20912](https://github.com/cosmos/cosmos-sdk/pull/20912) Fix `math.LegacyDec` type deserialization in GRPC queries.
* (x/group) [#20750](https://github.com/cosmos/cosmos-sdk/pull/20750) x/group shouldn't claim "orm" error codespace. This prevents any chain Cosmos SDK `v0.47` chain to use the ORM module.

## [v0.47.12](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.12) - 2024-06-10

## Improvements

* (x/authz,x/feegrant) [#20590](https://github.com/cosmos/cosmos-sdk/pull/20590) Provide updated keeper in depinject for authz and feegrant modules.

### Bug Fixes

* (baseapp) [#20144](https://github.com/cosmos/cosmos-sdk/pull/20144) Remove txs from mempool when AnteHandler fails in recheck.
* (testutil/sims) [#20151](https://github.com/cosmos/cosmos-sdk/pull/20151) Set all signatures and don't overwrite the previous one in `GenSignedMockTx`.

## [v0.47.11](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.11) - 2024-04-22

### Bug Fixes

* (x/feegrant,x/authz) [#20114](https://github.com/cosmos/cosmos-sdk/pull/20114) Follow up of [GHSA-4j93-fm92-rp4m](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-4j93-fm92-rp4m). The same issue was found in `x/feegrant` and `x/authz` modules.
* (crypto) [#20027](https://github.com/cosmos/cosmos-sdk/pull/20027) secp256r1 keys now implement gogoproto's customtype interface.
* (x/gov) [#19725](https://github.com/cosmos/cosmos-sdk/pull/19725) Fetch a failed proposal tally from `proposal.FinalTallyResult` in the gprc query.
* (crypto) [#19691](https://github.com/cosmos/cosmos-sdk/pull/19746) Throw an error when signing with incorrect Ledger.

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
