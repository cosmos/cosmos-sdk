# Upgrading Cosmos SDK

This guide provides instructions for upgrading to specific versions of Cosmos SDK.
Note, always read the **SimApp** section for more information on application wiring updates.

## [Unreleased]

### Unordered Transactions

The Cosmos SDK now supports unordered transactions. This means that transactions
can be executed in any order and doesn't require the client to deal with or manage
nonces. This also means the order of execution is not guaranteed. To enable unordered
transactions in your application:

* Update the `App` constructor to create, load, and save the unordered transaction
  manager.

	```go
	func NewApp(...) *App {
		// ...

		// create, start, and load the unordered tx manager
		utxDataDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data")
		app.UnorderedTxManager = unorderedtx.NewManager(utxDataDir)
		app.UnorderedTxManager.Start()

		if err := app.UnorderedTxManager.OnInit(); err != nil {
			panic(fmt.Errorf("failed to initialize unordered tx manager: %w", err))
		}
	}
	```

* Add the decorator to the existing AnteHandler chain, which should be as early
  as possible.

	```go
	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(),
		// ...
		ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxUnOrderedTTL, app.UnorderedTxManager),
		// ...
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
	```

* If the App has a SnapshotManager defined, you must also register the extension
  for the TxManager.

	```go
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(unorderedtx.NewSnapshotter(app.UnorderedTxManager))
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}
	```

* Create or update the App's `Close()` method to close the unordered tx manager.
  Note, this is critical as it ensures the manager's state is written to file
  such that when the node restarts, it can recover the state to provide replay
  protection.

	```go
	func (app *App) Close() error {
		// ...

		// close the unordered tx manager
		if e := app.UnorderedTxManager.Close(); e != nil {
			err = errors.Join(err, e)
		}

		return err
	}
	```

To submit an unordered transaction, the client must set the `unordered` flag to
`true` and ensure a reasonable `timeout_height` is set. The `timeout_height` is
used as a TTL for the transaction and is used to provide replay protection. See
[ADR-070](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-070-unordered-account.md)
for more details.

### Params

* Params Migrations were removed. It is required to migrate to 0.50 prior to upgrading to .51.

### SimApp

In this section we describe the changes made in Cosmos SDK' SimApp.
**These changes are directly applicable to your application wiring.**

#### AnteHandlers

The GasConsumptionDecorator and IncreaseSequenceDecorator have been merged with the SigVerificationDecorator, so you'll
need to remove them both from your app.go code, they will yield to unresolvable symbols when compiling.

#### Client (`root.go`)

The `client` package has been refactored to make use of the address codecs (address, validator address, consensus address, etc.).
This is part of the work of abstracting the SDK from the global bech32 config.

This means the address codecs must be provided in the `client.Context` in the application client (usually `root.go`).

```diff
clientCtx = clientCtx.
+ WithAddressCodec(addressCodec).
+ WithValidatorAddressCodec(validatorAddressCodec).
+ WithConsensusAddressCodec(consensusAddressCodec)
```

**When using `depinject` / `app v2`, the client codecs can be provided directly from application config.**

Refer to SimApp `root_v2.go` and `root.go` for an example with an app v2 and a legacy app.

### Modules

#### `**all**`

##### Genesis Interface

All genesis interfaces have been migrated to take context.Context instead of sdk.Context.

```golang
// InitGenesis performs genesis initialization for the authz module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, data json.RawMessage) {
}

// ExportGenesis returns the exported genesis state as raw bytes for the authz
// module.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
}
```

##### Migration to Collections

Most of Cosmos SDK modules have migrated to [collections](https://docs.cosmos.network/main/packages/collections).
Many functions have been removed due to this changes as the API can be smaller thanks to collections.
For modules that have migrated, verify you are checking against `collections.ErrNotFound` when applicable.

#### `x/auth`

Auth was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/auth`


#### `x/authz`

Authz was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/authz`

#### `x/bank`

Bank was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/bank`

#### `x/distribution`

Distribution was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/distribution`

The existing chains using x/distribution module needs to add the new x/protocolpool module.

#### `x/group`

Group was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/group`

#### `x/gov`

Gov was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/gov`

#### `x/mint`

Mint was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/mint`

#### `x/slashing`

Slashing was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/slashing`

#### `x/staking`

Staking was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/staking`


#### `x/params`

A standalone Go module was created and it is accessible at "cosmossdk.io/x/params".

#### `x/protocolpool`

Introducing a new `x/protocolpool` module to handle community pool funds. Its store must be added while upgrading to v0.51.x

Example:

```go
func (app SimApp) RegisterUpgradeHandlers() {
  	app.UpgradeKeeper.SetUpgradeHandler(
 		UpgradeName,
 		func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
 			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
 		},
 	)

  // ...
}
```

Add `x/protocolpool` store while upgrading to v0.51.x:

```go
storetypes.StoreUpgrades{
			Added: []string{
				protocolpooltypes.ModuleName,
			},
}
```

## [v0.50.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.0-alpha.0)

### Migration to CometBFT (Part 2)

The Cosmos SDK has migrated in its previous versions, to CometBFT.
Some functions have been renamed to reflect the naming change.

Following an exhaustive list:

* `client.TendermintRPC` -> `client.CometRPC`
* `clitestutil.MockTendermintRPC` -> `clitestutil.MockCometRPC`
* `clitestutilgenutil.CreateDefaultTendermintConfig` -> `clitestutilgenutil.CreateDefaultCometConfig`
* Package `client/grpc/tmservice` -> `client/grpc/cmtservice`

Additionally, the commands and flags mentioning `tendermint` have been renamed to `comet`.
These commands and flags are still supported for backward compatibility.

For backward compatibility, the `**/tendermint/**` gRPC services are still supported.

Additionally, the SDK is starting its abstraction from CometBFT Go types through the codebase:

* The usage of the CometBFT logger has been replaced by the Cosmos SDK logger interface (`cosmossdk.io/log.Logger`).
* The usage of `github.com/cometbft/cometbft/libs/bytes.HexByte` has been replaced by `[]byte`.
* Usage of an application genesis (see [genutil](#xgenutil)).

#### Enable Vote Extensions

:::tip
This is an optional feature that is disabled by default.
:::

Once all the code changes required to implement Vote Extensions are in place,
they can be enabled by setting the consensus param `Abci.VoteExtensionsEnableHeight`
to a value greater than zero.

In a new chain, this can be done in the `genesis.json` file.

For existing chains this can be done in two ways:

* During an upgrade the value is set in an upgrade handler.
* A governance proposal that changes the consensus param **after a coordinated upgrade has taken place**.

### BaseApp

All ABCI methods now accept a pointer to the request and response types defined
by CometBFT. In addition, they also return errors. An ABCI method should only
return errors in cases where a catastrophic failure has occurred and the application
should halt. However, this is abstracted away from the application developer. Any
handler that an application can define or set that returns an error, will gracefully
by handled by `BaseApp` on behalf of the application.

BaseApp calls of `BeginBlock` & `Endblock` are now private but are still exposed
to the application to define via the `Manager` type. `FinalizeBlock` is public
and should be used in order to test and run operations. This means that although
`BeginBlock` & `Endblock` no longer exist in the ABCI interface, they are automatically
called by `BaseApp` during `FinalizeBlock`. Specifically, the order of operations
is `BeginBlock` -> `DeliverTx` (for all txs) -> `EndBlock`.

ABCI++ 2.0 also brings `ExtendVote` and `VerifyVoteExtension` ABCI methods. These
methods allow applications to extend and verify pre-commit votes. The Cosmos SDK
allows an application to define handlers for these methods via `ExtendVoteHandler`
and `VerifyVoteExtensionHandler` respectively. Please see [here](https://docs.cosmos.network/v0.50/build/abci/03-vote-extensions)
for more info.

#### Set PreBlocker

A `SetPreBlocker` method has been added to BaseApp. This is essential for BaseApp to run `PreBlock` which runs before begin blocker other modules, and allows to modify consensus parameters, and the changes are visible to the following state machine logics.
Read more about other use cases [here](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-068-preblock.md).

`depinject` / app v2 users need to add `x/upgrade` in their `app_config.go` / `app.yml`:

```diff
+ PreBlockers: []string{
+	upgradetypes.ModuleName,
+ },
BeginBlockers: []string{
-	upgradetypes.ModuleName,
	minttypes.ModuleName,
}
```

When using (legacy) application wiring, the following must be added to `app.go`:

```diff
+app.ModuleManager.SetOrderPreBlockers(
+	upgradetypes.ModuleName,
+)

app.ModuleManager.SetOrderBeginBlockers(
-	upgradetypes.ModuleName,
)

+ app.SetPreBlocker(app.PreBlocker)

// ... //

+func (app *SimApp) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
+	return app.ModuleManager.PreBlock(ctx, req)
+}
```

#### Events

The log section of `abci.TxResult` is not populated in the case of successful
msg(s) execution. Instead a new attribute is added to all messages indicating
the `msg_index` which identifies which events and attributes relate the same
transaction.

`BeginBlock` & `EndBlock` Events are now emitted through `FinalizeBlock` but have
an added attribute, `mode=BeginBlock|EndBlock`, to identify if the event belongs
to `BeginBlock` or `EndBlock`.

### Config files

Confix is a new SDK tool for modifying and migrating configuration of the SDK.
It is the replacement of the `config.Cmd` command from the `client/config` package.

Use the following command to migrate your configuration:

```bash
simd config migrate v0.50
```

If you were using `<appd> config [key]` or `<appd> config [key] [value]` to set and get values from the `client.toml`, replace it with `<appd> config get client [key]` and `<appd> config set client [key] [value]`. The extra verbosity is due to the extra functionalities added in config.

More information about [confix](https://docs.cosmos.network/main/tooling/confix) and how to add it in your application binary in the [documentation](https://docs.cosmos.network/main/tooling/confix).

#### gRPC-Web

gRPC-Web is now listening to the same address and port as the gRPC Gateway API server (default: `localhost:1317`).
The possibility to listen to a different address has been removed, as well as its settings.
Use `confix` to clean-up your `app.toml`. A nginx (or alike) reverse-proxy can be set to keep the previous behavior.

#### Database Support

ClevelDB, BoltDB and BadgerDB are not supported anymore. To migrate from a unsupported database to a supported database please use a database migration tool.

### Protobuf

With the deprecation of the Amino JSON codec defined in [cosmos/gogoproto](https://github.com/cosmos/gogoproto) in favor of the protoreflect powered x/tx/aminojson codec, module developers are encouraged verify that their messages have the correct protobuf annotations to deterministically produce identical output from both codecs.

For core SDK types equivalence is asserted by generative testing of [SignableTypes](https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-beta.0/tests/integration/rapidgen/rapidgen.go#L102) in [TestAminoJSON_Equivalence](https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-beta.0/tests/integration/tx/aminojson/aminojson_test.go#L94).

**TODO: summarize proto annotation requirements.**

#### Stringer

The `gogoproto.goproto_stringer = false` annotation has been removed from most proto files. This means that the `String()` method is being generated for types that previously had this annotation. The generated `String()` method uses `proto.CompactTextString` for _stringifying_ structs.
[Verify](https://github.com/cosmos/cosmos-sdk/pull/13850#issuecomment-1328889651) the usage of the modified `String()` methods and double-check that they are not used in state-machine code.

### SimApp

In this section we describe the changes made in Cosmos SDK' SimApp.
**These changes are directly applicable to your application wiring.**

#### Module Assertions

Previously, all modules were required to be set in `OrderBeginBlockers`, `OrderEndBlockers` and `OrderInitGenesis / OrderExportGenesis` in `app.go` / `app_config.go`. This is no longer the case, the assertion has been loosened to only require modules implementing, respectively, the `appmodule.HasBeginBlocker`, `appmodule.HasEndBlocker` and `appmodule.HasGenesis` / `module.HasGenesis` interfaces.

#### Module wiring

The following modules `NewKeeper` function now take a `KVStoreService` instead of a `StoreKey`:

* `x/auth`
* `x/authz`
* `x/bank`
* `x/consensus`
* `x/crisis`
* `x/distribution`
* `x/evidence`
* `x/feegrant`
* `x/gov`
* `x/mint`
* `x/nft`
* `x/slashing`
* `x/upgrade`

**Users using `depinject` / app v2 do not need any changes, this is abstracted for them.**

Users manually wiring their chain need to use the `runtime.NewKVStoreService` method to create a `KVStoreService` from a `StoreKey`:

```diff
app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
  appCodec,
- keys[consensusparamtypes.StoreKey]
+ runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
  authtypes.NewModuleAddress(govtypes.ModuleName).String(),
)
```

#### Logger

Replace all your CometBFT logger imports by `cosmossdk.io/log`.

Additionally, `depinject` / app v2 users must now supply a logger through the main `depinject.Supply` function instead of passing it to `appBuilder.Build`.

```diff
appConfig = depinject.Configs(
	AppConfig,
	depinject.Supply(
		// supply the application options
		appOpts,
+		logger,
	...
```

```diff
- app.App = appBuilder.Build(logger, db, traceStore, baseAppOptions...)
+ app.App = appBuilder.Build(db, traceStore, baseAppOptions...)
```

User manually wiring their chain need to add the logger argument when creating the `x/bank` keeper.

#### Module Basics

Previously, the `ModuleBasics` was a global variable that was used to register all modules' `AppModuleBasic` implementation.
The global variable has been removed and the basic module manager can be now created from the module manager.

This is automatically done for `depinject` / app v2 users, however for supplying different app module implementation, pass them via `depinject.Supply` in the main `AppConfig` (`app_config.go`):

```go
depinject.Supply(
			// supply custom module basics
			map[string]module.AppModuleBasic{
				genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
				govtypes.ModuleName: gov.NewAppModuleBasic(
					[]govclient.ProposalHandler{
						paramsclient.ProposalHandler,
					},
				),
			},
		)
```

Users manually wiring their chain need to use the new `module.NewBasicManagerFromManager` function, after the module manager creation, and pass a `map[string]module.AppModuleBasic` as argument for optionally overriding some module's `AppModuleBasic`.

#### AutoCLI

[`AutoCLI`](https://docs.cosmos.network/main/core/autocli) has been implemented by the SDK for all its module CLI queries. This means chains must add the following in their `root.go` to enable `AutoCLI` in their application:

```go
if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
	panic(err)
}
```

Where `autoCliOpts` is the autocli options of the app, containing all modules and codecs.
That value can injected by depinject ([see root_v2.go](https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-beta.0/simapp/simd/cmd/root_v2.go#L49-L67)) or manually provided by the app ([see legacy app.go](https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-beta.0/simapp/app.go#L636-L655)).

:::warning
Not doing this will result in all core SDK modules queries not to be included in the binary.
:::

Additionally `AutoCLI` automatically adds the custom modules commands to the root command for all modules implementing the [`appmodule.AppModule`](https://pkg.go.dev/cosmossdk.io/core/appmodule#AppModule) interface.
This means, after ensuring all the used modules implement this interface, the following can be removed from your `root.go`:

```diff
func txCommand() *cobra.Command {
	....
- appd.ModuleBasics.AddTxCommands(cmd)
}
```

```diff
func queryCommand() *cobra.Command {
	....
- appd.ModuleBasics.AddQueryCommands(cmd)
}
```

### Packages

#### Math

References to `types/math.go` which contained aliases for math types aliasing the `cosmossdk.io/math` package have been removed.
Import directly the `cosmossdk.io/math` package instead.

#### Store

References to `types/store.go` which contained aliases for store types have been remapped to point to appropriate `store/types`, hence the `types/store.go` file is no longer needed and has been removed.

##### Extract Store to a standalone module

The `store` module is extracted to have a separate go.mod file which allows it be a standalone module.
All the store imports are now renamed to use `cosmossdk.io/store` instead of `github.com/cosmos/cosmos-sdk/store` across the SDK.

##### Streaming

[ADR-38](https://docs.cosmos.network/main/architecture/adr-038-state-listening) has been implemented in the SDK.

To continue using state streaming, replace `streaming.LoadStreamingServices` by the following in your `app.go`:

```go
if err := app.RegisterStreamingServices(appOpts, app.kvStoreKeys()); err != nil {
	panic(err)
}
```

#### Client

The return type of the interface method `TxConfig.SignModeHandler()` has been changed from `x/auth/signing.SignModeHandler` to `x/tx/signing.HandlerMap`. This change is transparent to most users as the `TxConfig` interface is typically implemented by private `x/auth/tx.config` struct (as returned by `auth.NewTxConfig`) which has been updated to return the new type. If users have implemented their own `TxConfig` interface, they will need to update their implementation to return the new type.

##### Textual sign mode

A new sign mode is available in the SDK that produces more human readable output, currently only available on Ledger
devices but soon to be implemented in other UIs. 

:::tip
This sign mode does not allow offline signing
:::

When using (legacy) application wiring, the following must be added to `app.go` after setting the app's bank keeper:

```go
	enabledSignModes := append(tx.DefaultSignModes, sigtypes.SignMode_SIGN_MODE_TEXTUAL)
	txConfigOpts := tx.ConfigOptions{
		EnabledSignModes:           enabledSignModes,
		TextualCoinMetadataQueryFn: txmodule.NewBankKeeperCoinMetadataQueryFn(app.BankKeeper),
	}
	txConfig, err := tx.NewTxConfigWithOptions(
		appCodec,
		txConfigOpts,
	)
	if err != nil {
		log.Fatalf("Failed to create new TxConfig with options: %v", err)
	}
	app.txConfig = txConfig
```

When using `depinject` / `app v2`, **it's enabled by default** if there's a bank keeper present.

And in the application client (usually `root.go`):

```go
	if !clientCtx.Offline {
		txConfigOpts.EnabledSignModes = append(txConfigOpts.EnabledSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
		txConfigOpts.TextualCoinMetadataQueryFn = txmodule.NewGRPCCoinMetadataQueryFn(clientCtx)
		txConfigWithTextual, err := tx.NewTxConfigWithOptions(
			codec.NewProtoCodec(clientCtx.InterfaceRegistry),
			txConfigOpts,
		)
		if err != nil {
			return err
		}
		clientCtx = clientCtx.WithTxConfig(txConfigWithTextual)
	}
```

When using `depinject` / `app v2`, the a tx config should be recreated from the `txConfigOpts` to use `NewGRPCCoinMetadataQueryFn` instead of depending on the bank keeper (that is used in the server).

To learn more see the [docs](https://docs.cosmos.network/main/learn/advanced/transactions#sign_mode_textual) and the [ADR-050](https://docs.cosmos.network/main/build/architecture/adr-050-sign-mode-textual).

### Modules

#### `**all**`

* [RFC 001](https://docs.cosmos.network/main/rfc/rfc-001-tx-validation) has defined a simplification of the message validation process for modules.
  The `sdk.Msg` interface has been updated to not require the implementation of the `ValidateBasic` method.
  It is now recommended to validate message directly in the message server. When the validation is performed in the message server, the `ValidateBasic` method on a message is no longer required and can be removed.

* Messages no longer need to implement the `LegacyMsg` interface and implementations of `GetSignBytes` can be deleted. Because of this change, global legacy Amino codec definitions and their registration in `init()` can safely be removed as well.

* The `AppModuleBasic` interface has been simplified. Defining `GetTxCmd() *cobra.Command` and `GetQueryCmd() *cobra.Command` is no longer required. The module manager detects when module commands are defined. If AutoCLI is enabled, `EnhanceRootCommand()` will add the auto-generated commands to the root command, unless a custom module command is defined and register that one instead.

* The following modules' `Keeper` methods now take in a `context.Context` instead of `sdk.Context`. Any module that has an interfaces for them (like "expected keepers") will need to update and re-generate mocks if needed:

    * `x/authz`
    * `x/bank`
    * `x/mint`
    * `x/crisis`
    * `x/distribution`
    * `x/evidence`
    * `x/gov`
    * `x/slashing`
    * `x/upgrade`

* `BeginBlock` and `EndBlock` have changed their signature, so it is important that any module implementing them are updated accordingly.

```diff
- BeginBlock(sdk.Context, abci.RequestBeginBlock)
+ BeginBlock(context.Context) error
```

```diff
- EndBlock(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
+ EndBlock(context.Context) error
```

In case a module requires to return `abci.ValidatorUpdate` from `EndBlock`, it can use the `HasABCIEndBlock` interface instead.

```diff
- EndBlock(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
+ EndBlock(context.Context) ([]abci.ValidatorUpdate, error)
```

:::tip
It is possible to ensure that a module implements the correct interfaces by using compiler assertions in your `x/{moduleName}/module.go`:

```go
var (
	_ module.AppModuleBasic      = (*AppModule)(nil)
	_ module.AppModuleSimulation = (*AppModule)(nil)
	_ module.HasGenesis          = (*AppModule)(nil)

	_ appmodule.AppModule        = (*AppModule)(nil)
	_ appmodule.HasBeginBlocker  = (*AppModule)(nil)
	_ appmodule.HasEndBlocker    = (*AppModule)(nil)
	...
)
```

Read more on those interfaces [here](https://docs.cosmos.network/v0.50/building-modules/module-manager#application-module-interfaces).

:::

* `GetSigners()` is no longer required to be implemented on `Msg` types. The SDK will automatically infer the signers from the `Signer` field on the message. The signer field is required on all messages unless using a custom signer function.

To find out more please read the [signer field](../../build/building-modules/05-protobuf-annotations.md#signer) & [here](https://github.com/cosmos/cosmos-sdk/blob/7352d0bce8e72121e824297df453eb1059c28da8/docs/docs/build/building-modules/02-messages-and-queries.md#L40) documentation.
<!-- Link to docs once redeployed -->

#### `x/auth`

For ante handler construction via `ante.NewAnteHandler`, the field `ante.HandlerOptions.SignModeHandler` has been updated to `x/tx/signing/HandlerMap` from `x/auth/signing/SignModeHandler`. Callers typically fetch this value from `client.TxConfig.SignModeHandler()` (which is also changed) so this change should be transparent to most users.

#### `x/capability`

Capability has been moved to [IBC Go](https://github.com/cosmos/ibc-go). IBC v8 will contain the necessary changes to incorporate the new module location.

#### `x/genutil`

The Cosmos SDK has migrated from a CometBFT genesis to a application managed genesis file.
The genesis is now fully handled by `x/genutil`. This has no consequences for running chains:

* Importing a CometBFT genesis is still supported.
* Exporting a genesis now exports the genesis as an application genesis.

When needing to read an application genesis, use the following helpers from the `x/genutil/types` package:

```go
// AppGenesisFromReader reads the AppGenesis from the reader.
func AppGenesisFromReader(reader io.Reader) (*AppGenesis, error)

// AppGenesisFromFile reads the AppGenesis from the provided file.
func AppGenesisFromFile(genFile string) (*AppGenesis, error)
```

#### `x/gov`

##### Expedited Proposals

The `gov` v1 module now supports expedited governance proposals. When a proposal is expedited, the voting period will be shortened to `ExpeditedVotingPeriod` parameter. An expedited proposal must have an higher voting threshold than a classic proposal, that threshold is defined with the `ExpeditedThreshold` parameter.

##### Cancelling Proposals

The `gov` module now supports cancelling governance proposals. When a proposal is canceled, all the deposits of the proposal are either burnt or sent to `ProposalCancelDest` address. The deposits burn rate will be determined by a new parameter called `ProposalCancelRatio` parameter.

```text
1. deposits * proposal_cancel_ratio will be burned or sent to `ProposalCancelDest` address , if `ProposalCancelDest` is empty then deposits will be burned.
2. deposits * (1 - proposal_cancel_ratio) will be sent to depositors.
```

By default, the new `ProposalCancelRatio` parameter is set to `0.5` during migration and `ProposalCancelDest` is set to empty string (i.e. burnt).

#### `x/evidence`

##### Extract evidence to a standalone module

The `x/evidence` module is extracted to have a separate go.mod file which allows it be a standalone module.
All the evidence imports are now renamed to use `cosmossdk.io/x/evidence` instead of `github.com/cosmos/cosmos-sdk/x/evidence` across the SDK.

#### `x/nft`

##### Extract nft to a standalone module

The `x/nft` module is extracted to have a separate go.mod file which allows it to be a standalone module.
All the evidence imports are now renamed to use `cosmossdk.io/x/nft` instead of `github.com/cosmos/cosmos-sdk/x/nft` across the SDK.

#### x/feegrant

##### Extract feegrant to a standalone module

The `x/feegrant` module is extracted to have a separate go.mod file which allows it to be a standalone module.
All the feegrant imports are now renamed to use `cosmossdk.io/x/feegrant` instead of `github.com/cosmos/cosmos-sdk/x/feegrant` across the SDK.

#### `x/upgrade`

##### Extract upgrade to a standalone module

The `x/upgrade` module is extracted to have a separate go.mod file which allows it to be a standalone module.
All the upgrade imports are now renamed to use `cosmossdk.io/x/upgrade` instead of `github.com/cosmos/cosmos-sdk/x/upgrade` across the SDK.

### Tooling

#### Rosetta

Rosetta has moved to it's own [repo](https://github.com/cosmos/rosetta) and not imported by the Cosmos SDK SimApp by default.
Any user who is interested on using the tool can connect it standalone to any node without the need to add it as part of the node binary.
The rosetta tool also allows multi chain connections.

## [v0.47.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.0)

### Migration to CometBFT (Part 1)

The Cosmos SDK has migrated to CometBFT, as its default consensus engine.
CometBFT is an implementation of the Tendermint consensus algorithm, and the successor of Tendermint Core.
Due to the import changes, this is a breaking change. Chains need to remove **entirely** their imports of Tendermint Core in their codebase, from direct and indirects imports in their `go.mod`.

* Replace `github.com/tendermint/tendermint` by `github.com/cometbft/cometbft`
* Replace `github.com/tendermint/tm-db` by `github.com/cometbft/cometbft-db`
* Verify `github.com/tendermint/tendermint` is not an indirect or direct dependency
* Run `make proto-gen`

Other than that, the migration should be seamless.
On the SDK side, clean-up of variables, functions to reflect the new name will only happen from v0.50 (part 2).

Note: It is possible that these steps must first be performed by your dependencies before you can perform them on your own codebase.

### Simulation

Remove `RandomizedParams` from `AppModuleSimulation` interface. Previously, it used to generate random parameter changes during simulations, however, it does so through ParamChangeProposal which is now legacy. Since all modules were migrated, we can now safely remove this from `AppModuleSimulation` interface.

Moreover, to support the `MsgUpdateParams` governance proposals for each modules, `AppModuleSimulation` now defines a `AppModule.ProposalMsgs` method in addition to `AppModule.ProposalContents`. That method defines the messages that can be used to submit a proposal and that should be tested in simulation.

When a module has no proposal messages or proposal content to be tested by simulation, the `AppModule.ProposalMsgs` and `AppModule.ProposalContents` methods can be deleted.

### gRPC

A new gRPC service, `proto/cosmos/base/node/v1beta1/query.proto`, has been introduced
which exposes various operator configuration. App developers should be sure to
register the service with the gRPC-gateway service via
`nodeservice.RegisterGRPCGatewayRoutes` in their application construction, which
is typically found in `RegisterAPIRoutes`.

### AppModule Interface

Support for the `AppModule` `Querier`, `Route` and `LegacyQuerier` methods has been entirely removed from the `AppModule`
interface. This removes and fully deprecates all legacy queriers. All modules no longer support the REST API previously
known as the LCD, and the `sdk.Msg#Route` method won't be used anymore.

Most other existing `AppModule` methods have been moved to extension interfaces in preparation for the migration
to the `cosmossdk.io/core/appmodule` API in the next release. Most `AppModule` implementations should not be broken
by this change.

### SimApp

The `simapp` package **should not be imported in your own app**. Instead, you should import the `runtime.AppI` interface, that defines an `App`, and use the [`simtestutil` package](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/testutil/sims) for application testing.

#### App Wiring

SimApp's `app_v2.go` is using [App Wiring](https://docs.cosmos.network/main/building-apps/app-go-v2), the dependency injection framework of the Cosmos SDK.
This means that modules are injected directly into SimApp thanks to a [configuration file](https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/simapp/app_config.go).
The previous behavior, without the dependency injection framework, is still present in [`app.go`](https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/simapp/app.go) and is not going anywhere.

If you are using a `app.go` without dependency injection, add the following lines to your `app.go` in order to provide newer gRPC services:

```go
autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

reflectionSvc, err := runtimeservices.NewReflectionService()
if err != nil {
    panic(err)
}
reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)
```

#### Constructor

The constructor, `NewSimApp` has been simplified:

* `NewSimApp` does not take encoding parameters (`encodingConfig`) as input, instead the encoding parameters are injected (when using app wiring), or directly created in the constructor. Instead, we can instantiate `SimApp` for getting the encoding configuration.
* `NewSimApp` now uses `AppOptions` for getting the home path (`homePath`) and the invariant checks period (`invCheckPeriod`). These were unnecessary given as arguments as they were already present in the `AppOptions`.

#### Encoding

`simapp.MakeTestEncodingConfig()` was deprecated and has been removed. Instead you can use the `TestEncodingConfig` from the `types/module/testutil` package.
This means you can replace your usage of `simapp.MakeTestEncodingConfig` in tests to `moduletestutil.MakeTestEncodingConfig`, which takes a series of relevant `AppModuleBasic` as input (the module being tested and any potential dependencies).

#### Export

`ExportAppStateAndValidators` takes an extra argument, `modulesToExport`, which is a list of module names to export.
That argument should be passed to the module maanager `ExportGenesisFromModules` method.

#### Replaces

The `GoLevelDB` version must pinned to `v1.0.1-0.20210819022825-2ae1ddf74ef7` in the application, following versions might cause unexpected behavior.
This can be done adding `replace github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7` to the `go.mod` file.

* [issue #14949 on cosmos-sdk](https://github.com/cosmos/cosmos-sdk/issues/14949)
* [issue #25413 on go-ethereum](https://github.com/ethereum/go-ethereum/pull/25413)

### Protobuf

The SDK has migrated from `gogo/protobuf` (which is currently unmaintained), to our own maintained fork, [`cosmos/gogoproto`](https://github.com/cosmos/gogoproto).

This means you should replace all imports of `github.com/gogo/protobuf` to `github.com/cosmos/gogoproto`.
This allows you to remove the replace directive `replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1` from your `go.mod` file.

Please use the `ghcr.io/cosmos/proto-builder` image (version >= `0.11.5`) for generating protobuf files.

See which buf commit for `cosmos/cosmos-sdk` to pin in your `buf.yaml` file [here](https://github.com/cosmos/cosmos-sdk/blob/main/proto/README.md).

#### Gogoproto Import Paths

The SDK made a [patch fix](https://github.com/cosmos/gogoproto/pull/32) on its gogoproto repository to require that each proto file's package name matches its OS import path (relatively to a protobuf root import path, usually the root `proto/` folder, set by the `protoc -I` flag).

For example, assuming you put all your proto files in subfolders inside your root `proto/` folder, then a proto file with package name `myapp.mymodule.v1` should be found in the `proto/myapp/mymodule/v1/` folder. If it is in another folder, the proto generation command will throw an error.

If you are using a custom folder structure for your proto files, please reorganize them so that their OS path matches their proto package name.

This is to allow the proto FileDescriptSets to be correctly registered, and this standardized OS import paths allows [Hubl](https://github.com/cosmos/cosmos-sdk/tree/main/tools/hubl) to reflectively talk to any chain.

#### `{accepts,implements}_interface` proto annotations

The SDK is normalizing the strings inside the Protobuf `accepts_interface` and `implements_interface` annotations. We require them to be fully-scoped names. They will soon be used by code generators like Pulsar and Telescope to match which messages can or cannot be packed inside `Any`s.

Here are the following replacements that you need to perform on your proto files:

```diff
- "Content"
+ "cosmos.gov.v1beta1.Content"
- "Authorization"
+ "cosmos.authz.v1beta1.Authorization"
- "sdk.Msg"
+ "cosmos.base.v1beta1.Msg"
- "AccountI"
+ "cosmos.auth.v1beta1.AccountI"
- "ModuleAccountI"
+ "cosmos.auth.v1beta1.ModuleAccountI"
- "FeeAllowanceI"
+ "cosmos.feegrant.v1beta1.FeeAllowanceI"
```

Please also check that in your own app's proto files that there are no single-word names for those two proto annotations. If so, then replace them with fully-qualified names, even though those names don't actually resolve to an actual protobuf entity.

For more information, see the [encoding guide](https://github.com/cosmos/cosmos-sdk/blob/main/docs/learn/advanced/05-encoding.md).

### Transactions

#### Broadcast Mode

Broadcast mode `block` was deprecated and has been removed. Please use `sync` mode
instead. When upgrading your tests from `block` to `sync` and checking for a
transaction code, you need to query the transaction first (with its hash) to get
the correct code.

### Modules

#### `**all**`

`EventTypeMessage` events, with `sdk.AttributeKeyModule` and `sdk.AttributeKeySender` are now emitted directly at message execution (in `baseapp`).
This means that the following boilerplate should be removed from all your custom modules:

```go
ctx.EventManager().EmitEvent(
	sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, `signer/sender`),
	),
)
```

The module name is assumed by `baseapp` to be the second element of the message route: `"cosmos.bank.v1beta1.MsgSend" -> "bank"`.
In case a module does not follow the standard message path, (e.g. IBC), it is advised to keep emitting the module name event.
`Baseapp` only emits that event if the module has not already done so.

#### `x/params`

The `params` module was deprecated since v0.46. The Cosmos SDK has migrated away from `x/params` for its own modules.
Cosmos SDK modules now store their parameters directly in its respective modules.
The `params` module will be removed in `v0.50`, as mentioned [in v0.46 release](https://github.com/cosmos/cosmos-sdk/blob/v0.46.1/UPGRADING.md#xparams). It is strongly encouraged to migrate away from `x/params` before `v0.50`.

When performing a chain migration, the params table must be initizalied manually. This was done in the modules keepers in previous versions.
Have a look at `simapp.RegisterUpgradeHandlers()` for an example.

#### `x/crisis`

With the migrations of all modules away from `x/params`, the crisis module now has a store.
The store must be created during a chain upgrade to v0.47.x.

```go
storetypes.StoreUpgrades{
			Added: []string{
				crisistypes.ModuleName,
			},
}
```

#### `x/gov`

##### Minimum Proposal Deposit At Time of Submission

The `gov` module has been updated to support a minimum proposal deposit at submission time. It is determined by a new
parameter called `MinInitialDepositRatio`. When multiplied by the existing `MinDeposit` parameter, it produces
the necessary proportion of coins needed at the proposal submission time. The motivation for this change is to prevent proposal spamming.

By default, the new `MinInitialDepositRatio` parameter is set to zero during migration. The value of zero signifies that this
feature is disabled. If chains wish to utilize the minimum proposal deposits at time of submission, the migration logic needs to be
modified to set the new parameter to the desired value.

##### New `Proposal.Proposer` field

The `Proposal` proto has been updated with proposer field. For proposal state migraton developers can call `v4.AddProposerAddressToProposal` in their upgrade handler to update all existing proposal and make them compatible and **this migration is optional**.

```go
import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	v4 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v4"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func (app SimApp) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(UpgradeName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			// this migration is optional
			// add proposal ids with proposers which are active (deposit or voting period)
			proposals := make(map[uint64]string)
			proposals[1] = "cosmos1luyncewxk4lm24k6gqy8y5dxkj0klr4tu0lmnj" ...
			v4.AddProposerAddressToProposal(ctx, sdk.NewKVStoreKey(v4.ModuleName), app.appCodec, proposals)
			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		})
}

```

#### `x/consensus`

Introducing a new `x/consensus` module to handle managing Tendermint consensus
parameters. For migration it is required to call a specific migration to migrate
existing parameters from the deprecated `x/params` to `x/consensus` module. App
developers should ensure to call `baseapp.MigrateParams` in their upgrade handler.

Example:

```go
func (app SimApp) RegisterUpgradeHandlers() {
 	----> baseAppLegacySS := app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable()) <----

 	app.UpgradeKeeper.SetUpgradeHandler(
 		UpgradeName,
 		func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
 			// Migrate Tendermint consensus parameters from x/params module to a
 			// dedicated x/consensus module.
 			----> baseapp.MigrateParams(ctx, baseAppLegacySS, &app.ConsensusParamsKeeper) <----

			// ...

 			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
 		},
 	)

  // ...
}
```

The `x/params` module should still be imported in your app.go in order to handle this migration.

Because the `x/consensus` module is a new module, its store must be added while upgrading to v0.47.x:

```go
storetypes.StoreUpgrades{
			Added: []string{
				consensustypes.ModuleName,
			},
}
```

##### `app.go` changes

When using an `app.go` without App Wiring, the following changes are required:

```diff
- bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable()))
+ app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, keys[consensusparamstypes.StoreKey], authtypes.NewModuleAddress(govtypes.ModuleName).String())
+ bApp.SetParamStore(&app.ConsensusParamsKeeper)
```

When using App Wiring, the parameter store is automatically set for you.

#### `x/nft`

The SDK does not validate anymore the `classID` and `nftID` of an NFT, for extra flexibility in your NFT implementation.
This means chain developers need to validate the `classID` and `nftID` of an NFT.

### Ledger

Ledger support has been generalized to enable use of different apps and keytypes that use `secp256k1`. The Ledger interface remains the same, but it can now be provided through the Keyring `Options`, allowing higher-level chains to connect to different Ledger apps or use custom implementations. In addition, higher-level chains can provide custom key implementations around the Ledger public key, to enable greater flexibility with address generation and signing.

This is not a breaking change, as all values will default to use the standard Cosmos app implementation unless specified otherwise.

## [v0.46.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.0)

### Go API Changes

The `replace google.golang.org/grpc` directive can be removed from the `go.mod`, it is no more required to block the version.

A few packages that were deprecated in the previous version are now removed.

For instance, the REST API, deprecated in v0.45, is now removed. If you have not migrated yet, please follow the [instructions](https://docs.cosmos.network/v0.45/migrations/rest.html).

To improve clarity of the API, some renaming and improvements has been done:

| Package   | Previous                           | Current                              |
| --------- | ---------------------------------- | ------------------------------------ |
| `simapp`  | `encodingConfig.Marshaler`         | `encodingConfig.Codec`               |
| `simapp`  | `FundAccount`, `FundModuleAccount` | Functions moved to `x/bank/testutil` |
| `types`   | `AccAddressFromHex`                | `AccAddressFromHexUnsafe`            |
| `x/auth`  | `MempoolFeeDecorator`              | Use `DeductFeeDecorator` instead     |
| `x/bank`  | `AddressFromBalancesStore`         | `AddressAndDenomFromBalancesStore`   |
| `x/gov`   | `keeper.DeleteDeposits`            | `keeper.DeleteAndBurnDeposits`       |
| `x/gov`   | `keeper.RefundDeposits`            | `keeper.RefundAndDeleteDeposits`     |
| `x/{mod}` | package `legacy`                   | package `migrations`                 |

For the exhaustive list of API renaming, please refer to the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/main/CHANGELOG.md).

#### new packages

Additionally, new packages have been introduced in order to further split the codebase. Aliases are available for a new API breaking migration, but it is encouraged to migrate to this new packages:

* `errors` should replace `types/errors` when registering errors or wrapping SDK errors.
* `math` contains the `Int` or `Uint` types that are used in the SDK.
* `x/nft` an NFT base module.
* `x/group` a group module allowing to create DAOs, multisig and policies. Greatly composes with `x/authz`.

#### `x/authz`

* `authz.NewMsgGrant` `expiration` is now a pointer. When `nil` is used, then no expiration will be set (grant won't expire).
* `authz.NewGrant` takes a new argument: block time, to correctly validate expire time.

### Keyring

The keyring has been refactored in v0.46.

* The `Unsafe*` interfaces have been removed from the keyring package. Please use interface casting if you wish to access those unsafe functions.
* The keys' implementation has been refactored to be serialized as proto.
* `keyring.NewInMemory` and `keyring.New` takes now a `codec.Codec`.
* Take `keyring.Record` instead of `Info` as first argument in:
  _ `MkConsKeyOutput`
  _ `MkValKeyOutput` \* `MkAccKeyOutput`
* Rename:
  _ `SavePubKey` to `SaveOfflineKey` and remove the `algo` argument.
  _ `NewMultiInfo`, `NewLedgerInfo` to `NewLegacyMultiInfo`, `newLegacyLedgerInfo` respectively. \* `NewOfflineInfo` to `newLegacyOfflineInfo` and move it to `migration_test.go`.

### PostHandler

A `postHandler` is like an `antehandler`, but is run _after_ the `runMsgs` execution. It is in the same store branch that `runMsgs`, meaning that both `runMsgs` and `postHandler`. This allows to run a custom logic after the execution of the messages.

### IAVL

v0.19.0 IAVL introduces a new "fast" index. This index represents the latest state of the
IAVL laid out in a format that preserves data locality by key. As a result, it allows for faster queries and iterations
since data can now be read in lexicographical order that is frequent for Cosmos-SDK chains.

The first time the chain is started after the upgrade, the aforementioned index is created. The creation process
might take time and depends on the size of the latest state of the chain. For example, Osmosis takes around 15 minutes to rebuild the index.

While the index is being created, node operators can observe the following in the logs:
"Upgrading IAVL storage for faster queries + execution on the live state. This may take a while". The store
key is appended to the message. The message is printed for every module that has a non-transient store.
As a result, it gives a good indication of the progress of the upgrade.

There is also downgrade and re-upgrade protection. If a node operator chooses to downgrade to IAVL pre-fast index, and then upgrade again, the index is rebuilt from scratch. This implementation detail should not be relevant in most cases. It was added as a safeguard against operator
mistakes.

### Modules

#### `x/params`

* The `x/params` module has been deprecated in favour of each module housing and providing way to modify their parameters. Each module that has parameters that are changeable during runtime have an authority, the authority can be a module or user account. The Cosmos SDK team recommends migrating modules away from using the param module. An example of how this could look like can be found [here](https://github.com/cosmos/cosmos-sdk/pull/12363).
* The Param module will be maintained until April 18, 2023. At this point the module will reach end of life and be removed from the Cosmos SDK.

#### `x/gov`

The `gov` module has been greatly improved. The previous API has been moved to `v1beta1` while the new implementation is called `v1`.

In order to submit a proposal with `submit-proposal` you now need to pass a `proposal.json` file.
You can still use the old way by using `submit-legacy-proposal`. This is not recommended.
More information can be found in the gov module [client documentation](https://docs.cosmos.network/v0.46/modules/gov/07_client.html).

#### `x/staking`

The `staking module` added a new message type to cancel unbonding delegations. Users that have unbonded by accident or wish to cancel a undelegation can now specify the amount and valdiator they would like to cancel the unbond from

### Protobuf

The `third_party/proto` folder that existed in [previous version](https://github.com/cosmos/cosmos-sdk/tree/v0.45.3/third_party/proto) now does not contains directly the [proto files](https://github.com/cosmos/cosmos-sdk/tree/release/v0.46.x/third_party/proto).

Instead, the SDK uses [`buf`](https://buf.build). Clients should have their own [`buf.yaml`](https://docs.buf.build/configuration/v1/buf-yaml) with `buf.build/cosmos/cosmos-sdk` as dependency, in order to avoid having to copy paste these files.

The protos can as well be downloaded using `buf export buf.build/cosmos/cosmos-sdk:8cb30a2c4de74dc9bd8d260b1e75e176 --output <some_folder>`.

Cosmos message protobufs should be extended with `cosmos.msg.v1.signer`:

```protobuf
message MsgSetWithdrawAddress {
  option (cosmos.msg.v1.signer) = "delegator_address"; ++

  option (gogoproto.equal)           = false;
  option (gogoproto.goproto_getters) = false;

  string delegator_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string withdraw_address  = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

When clients interact with a node they are required to set a codec in in the grpc.Dial. More information can be found in this [doc](https://docs.cosmos.network/v0.46/run-node/interact-node.html#programmatically-via-go).
