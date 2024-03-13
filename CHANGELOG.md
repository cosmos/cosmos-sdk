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

### API Breaking Changes

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
* (types) [#19652](https://github.com/cosmos/cosmos-sdk/pull/19652)
    * Moved`types/module.HasRegisterInterfaces` to `cosmossdk.io/core`.
    * Moved `RegisterInterfaces` and `RegisterImplementations` from `InterfaceRegistry` to `cosmossdk.io/core/registry.LegacyRegistry` interface.
* (types) [#19627](https://github.com/cosmos/cosmos-sdk/pull/19627) and [#19735](https://github.com/cosmos/cosmos-sdk/pull/19735) All genesis interfaces now don't take `codec.JsonCodec`.
    * Every module has the codec already, passing it created an unneeded dependency.
    * Additionally, to reflect this change, the module manager does not take a codec either.

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

>>>>>>> 5424b55c5 (fix(crypto): error if incorrect ledger public key (#19691))
### Bug Fixes

* (x/gov) [#19725](https://github.com/cosmos/cosmos-sdk/pull/19725) Fetch a failed proposal tally from proposal.FinalTallyResult in the gprc query.

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
