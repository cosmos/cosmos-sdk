# Upgrading Cosmos SDK

This guide provides instructions for upgrading to specific versions of Cosmos SDK.
Note, always read the **SimApp** section for more information on application wiring updates.

## [v0.50.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.0)

### Migration to CometBFT (Part 2)

The Cosmos SDK has migrated in its previous versions, to CometBFT.
Some functions have been renamed to reflect the naming change.

Following an exhaustive list:

* `client.TendermintRPC` -> `client.CometRPC`
* `clitestutil.MockTendermintRPC` -> `clitestutil.MockCometRPC`
* `clitestutilgenutil.CreateDefaultTendermintConfig` -> `clitestutilgenutil.CreateDefaultCometConfig`
* Package `client/grpc/tmservice` -> `client/grpc/cmtservice`

Additionally, the commands and flags mentioning `tendermint` have been renamed to `comet`.
However, these commands and flags are still supported for backward compatibility.

For backward compatibility, the `**/tendermint/**` gRPC services are still supported.

Additionally, the SDK is starting its abstraction from CometBFT Go types thorought the codebase:

* The usage of the CometBFT logger has been replaced by the Cosmos SDK logger interface (`cosmossdk.io/log.Logger`).
* The usage of `github.com/cometbft/cometbft/libs/bytes.HexByte` has been replaced by `[]byte`.

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
and `VerifyVoteExtensionHandler` respectively. Please see [here](https://docs.cosmos.network/v0.50/building-apps/vote-extensions)
for more info.

#### Upgrade

**Users using `depinject` / app v2 do not need any changes, this is abstracted for them.**
```diff
+ app.BaseApp.SetMigrationModuleManager(app.ModuleManager)
```
BaseApp added `SetMigrationModuleManager` for apps to set their ModuleManager which implements `RunMigrationBeginBlock`. This is essential for BaseApp to run `BeginBlock` of upgrade module and inject `ConsensusParams` to context for `beginBlocker` during `beginBlock`.

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

Previously, the `ModuleBasics` was a global variable that was used to register all modules's `AppModuleBasic` implementation.
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

Users manually wiring their chain need to use the new `module.NewBasicManagerFromManager` function, after the module manager creation, and pass a `map[string]module.AppModuleBasic` as argument for optionally overridding some module's `AppModuleBasic`.

#### AutoCLI

[`AutoCLI`](https://docs.cosmos.network/main/building-modules/autocli) has been implemented by the SDK for all its module CLI queries. This means chains must add the following in their `root.go` to enable `AutoCLI` in their application:

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

The return type of the interface method `TxConfig.SignModeHandler()` has been changed from `x/auth/signing.SignModeHandler` to `x/tx/signing.HandlerMap`. This change is transparent to most users as the `TxConfig` interface is typically implemented by private `x/auth/tx.config` struct (as returned by `auth.NewTxConfig`) which has been updated to return the new type.  If users have implemented their own `TxConfig` interface, they will need to update their implementation to return the new type.

### Modules

#### `**all**`

* [RFC 001](https://docs.cosmos.network/main/rfc/rfc-001-tx-validation) has defined a simplification of the message validation process for modules.
The `sdk.Msg` interface has been updated to not require the implementation of the `ValidateBasic` method.
It is now recommended to validate message directly in the message server. When the validation is performed in the message server, the `ValidateBasic` method on a message is no longer required and can be removed.

* Messages no longer need to implement the `LegacyMsg` interface and implementations of `GetSignBytes` can be deleted. Because of this change, global legacy Amino codec definitions and their registration in `init()` can safely be removed as well.  

* The `AppModuleBasic` interface has been simplifed. Defining `GetTxCmd() *cobra.Command` and `GetQueryCmd() *cobra.Command` is no longer required. The module manager detects when module commands are defined. If AutoCLI is enabled, `EnhanceRootCommand()` will add the auto-generated commands to the root command, unless a custom module command is defined and register that one instead.

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

#### `x/auth`

For ante handler construction via `ante.NewAnteHandler`, the field `ante.HandlerOptions.SignModeHandler` has been updated to `x/tx/signing/HandlerMap` from `x/auth/signing/SignModeHandler`.  Callers typically fetch this value from `client.TxConfig.SignModeHandler()` (which is also changed) so this change should be transparent to most users.

#### `x/capability`

Capability has been moved to [IBC-GO](https://github.com/cosmos/ibc-go). IBC v8 will contain the necessary changes to incorporate the new module location.

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
