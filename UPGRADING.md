# Upgrading Cosmos SDK

This guide provides instructions for upgrading to specific versions of Cosmos SDK.
Note, always read the **SimApp** section for more information on application wiring updates.

<<<<<<< HEAD
## [v0.50.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.0)
=======
## [Unreleased]

### BaseApp

#### Nested Messages Simulation

Now it is possible to simulate the nested messages of a message, providing developers with a powerful tool for
testing and predicting the behavior of complex transactions. This feature allows for a more comprehensive
evaluation of gas consumption, state changes, and potential errors that may occur when executing nested
messages. However, it's important to note that while the simulation can provide valuable insights, it does not
guarantee the correct execution of the nested messages in the future. Factors such as changes in the
blockchain state or updates to the protocol could potentially affect the actual execution of these nested
messages when the transaction is finally processed on the network.

For example, consider a governance proposal that includes nested messages to update multiple protocol
parameters. At the time of simulation, the blockchain state may be suitable for executing all these nested
messages successfully. However, by the time the actual governance proposal is executed (which could be days or
weeks later), the blockchain state might have changed significantly. As a result, while the simulation showed
a successful execution, the actual governance proposal might fail when it's finally processed.

By default, when simulating transactions, the gas cost of nested messages is not calculated. This means that
only the gas cost of the top-level message is considered. However, this behavior can be customized using the
`SetIncludeNestedMsgsGas` option when building the BaseApp. By providing a list of message types to this option,
you can specify which messages should have their nested message gas costs included in the simulation. This
allows for more accurate gas estimation for transactions involving specific message types that contain nested
messages, while maintaining the default behavior for other message types.

Here is an example on how `SetIncludeNestedMsgsGas` option could be set to calculate the gas of a gov proposal
nested messages:

```go
baseAppOptions = append(baseAppOptions, baseapp.SetIncludeNestedMsgsGas([]sdk.Message{&gov.MsgSubmitProposal{}}))
// ...
app.App = appBuilder.Build(db, traceStore, baseAppOptions...)
```

## [v0.52.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.52.0-alpha.0)

Documentation to migrate an application from v0.50.x to server/v2 is available elsewhere.
It is additional to the changes described here.

### SimApp

In this section we describe the changes made in Cosmos SDK' SimApp.
**These changes are directly applicable to your application wiring.**
Please read this section first, but for an exhaustive list of changes, refer to the [CHANGELOG](./simapp/CHANGELOG.md).

#### Client (`root.go`)

The `client` package has been refactored to make use of the address codecs (address, validator address, consensus address, etc.)
and address bech32 prefixes (address and validator address).
This is part of the work of abstracting the SDK from the global bech32 config.

This means the address codecs and prefixes must be provided in the `client.Context` in the application client (usually `root.go`).

```diff
clientCtx = clientCtx.
+ WithAddressCodec(addressCodec).
+ WithValidatorAddressCodec(validatorAddressCodec).
+ WithConsensusAddressCodec(consensusAddressCodec).
+ WithAddressPrefix("cosmos").
+ WithValidatorPrefix("cosmosvaloper")
```

**When using `depinject` / `app_di`, the client codecs can be provided directly from application config.**

Refer to SimApp `root_di.go` and `root.go` for an example with an app di and a legacy app.

Additionally, a simplification of the start command leads to the following change:

```diff
- server.AddCommands(rootCmd, newApp, func(startCmd *cobra.Command) {})
+ server.AddCommands(rootCmd, newApp, server.StartCmdOptions[servertypes.Application]{})
```

#### Server (`app.go`)

##### Module Manager

The basic module manager has been deleted. It was not necessary anymore and was simplified to use the `module.Manager` directly.
It can be removed from your `app.go`.

For depinject users, it isn't necessary anymore to supply a `map[string]module.AppModuleBasic` for customizing the app module basic instantiation.
The custom parameters (such as genutil message validator or gov proposal handler, or evidence handler) can directly be supplied.
When requiring a module manager in `root.go`, inject `*module.Manager` using `depinject.Inject`. 

For non depinject users, simply call `RegisterLegacyAminoCodec` and `RegisterInterfaces` on the module manager:

```diff
-app.BasicModuleManager = module.NewBasicManagerFromManager(...)
-app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
-app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)
+app.ModuleManager.RegisterLegacyAminoCodec(legacyAmino)
+app.ModuleManager.RegisterInterfaces(interfaceRegistry)
```

Additionally, thanks to the genesis simplification, as explained in [the genesis interface update](#genesis-interface), the module manager `InitGenesis` and `ExportGenesis` methods do not require the codec anymore.

##### GRPC-WEB

Grpc-web embedded client has been removed from the server. If you would like to use grpc-web, you can use the [envoy proxy](https://www.envoyproxy.io/docs/envoy/latest/start/start).

##### AnteHandlers

The `GasConsumptionDecorator` and `IncreaseSequenceDecorator` have been merged with the SigVerificationDecorator, so you'll
need to remove them both from your app.go code, they will yield to unresolvable symbols when compiling.

#### Unordered Transactions

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
		ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, options.TxManager, options.Environment),
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
[ADR-070](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-070-unordered-transactions.md)
for more details.

### Depinject `app_config.go` / `app.yml`

With the introduction of [environment in modules](#core-api), depinject automatically creates the environment for all modules.
Learn more about environment [here](https://example.com) <!-- TODO -->. Given the fields of environment, this means runtime creates a kv store service for all modules by default. It can happen that some modules do not have a store necessary (such as `x/auth/tx` for instance). In this case, the store creation should be skipped in `app_config.go`:

```diff
InitGenesis: []string{
	"..."
},
+ // SkipStoreKeys is an optional list of store keys to skip when constructing the
+ // module's keeper. This is useful when a module does not have a store key.
+ SkipStoreKeys: []string{
+ 	"tx",
+ },
```

### Protobuf

The `cosmossdk.io/api/tendermint` package has been removed as CometBFT now publishes its protos to `buf.build/tendermint` and `buf.build/cometbft`.
There is no longer a need for the Cosmos SDK to host these protos for itself and its dependencies.
That package containing proto v2 generated code, but the SDK now uses [buf generated go SDK instead](https://buf.build/docs/bsr/generated-sdks/go).
If you were depending on `cosmossdk.io/api/tendermint`, please use the buf generated go SDK instead, or ask CometBFT host the generated proto v2 code.

The `codectypes.Any` has moved to `github.com/cosmos/gogoproto/types/any`. Module developers need to update the `buf.gen.gogo.yaml` configuration files by adjusting the corresponding `opt` option to `Mgoogle/protobuf/any.proto=github.com/cosmos/gogoproto/types/any` for directly mapping the`Any` type to its new location:

```diff
version: v1
plugins:
  - name: gocosmos
    out: ..
- 	 opt: plugins=grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types,Mcosmos/orm/v1/orm.proto=cosmossdk.io/orm
+    opt: plugins=grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/gogoproto/types/any,Mcosmos/orm/v1/orm.proto=cosmossdk.io/orm
  - name: grpc-gateway
    out: ..
    opt: logtostderr=true,allow_colon_final_segments=true

```

Also, any usages of the interfaces `AnyUnpacker` and `UnpackInterfacesMessage` must be replaced with the interfaces of the same name in the `github.com/cosmos/gogoproto/types/any` package.

### Modules

#### `**all**`

##### Core API

Core API has been introduced for modules since v0.47. With the deprecation of `sdk.Context`, we strongly recommend to use the `cosmossdk.io/core/appmodule` interfaces for the modules. This will allow the modules to work out of the box with server/v2 and baseapp, as well as limit their dependencies on the SDK.

Additionally, the `appmodule.Environment` struct is introduced to fetch different services from the application.
This should be used as an alternative to using `sdk.UnwrapContext(ctx)` to fetch the services.
It needs to be passed into a module at instantiation (or depinject will inject the correct environment). 

`x/circuit` is used as an example:

```go
app.CircuitKeeper = circuitkeeper.NewKeeper(runtime.NewEnvironment(runtime.NewKVStoreService(keys[circuittypes.StoreKey]), logger.With(log.ModuleKey, "x/circuit")), appCodec, authtypes.NewModuleAddress(govtypes.ModuleName).String(), app.AuthKeeper.AddressCodec())
```

If your module requires a message server or query server, it should be passed in the environment as well.

```diff
-govKeeper := govkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[govtypes.StoreKey]), app.AuthKeeper, app.BankKeeper,app.StakingKeeper, app.PoolKeeper, app.MsgServiceRouter(), govConfig, authtypes.NewModuleAddress(govtypes.ModuleName).String())
+govKeeper := govkeeper.NewKeeper(appCodec, runtime.NewEnvironment(runtime.NewKVStoreService(keys[govtypes.StoreKey]), logger.With(log.ModuleKey, "x/circuit"), runtime.EnvWithMsgRouterService(app.MsgServiceRouter()), runtime.EnvWithQueryRouterService(app.GRPCQueryRouter())), app.AuthKeeper, app.BankKeeper, app.StakingKeeper, app.PoolKeeper, govConfig, authtypes.NewModuleAddress(govtypes.ModuleName).String())
```

The signature of the extension interface `HasRegisterInterfaces` has been changed to accept a `cosmossdk.io/core/registry.InterfaceRegistrar` instead of a `codec.InterfaceRegistry`.   `HasRegisterInterfaces` is now a part of `cosmossdk.io/core/appmodule`.  Modules should update their `HasRegisterInterfaces` implementation to accept a `cosmossdk.io/core/registry.InterfaceRegistrar` interface.

```diff
-func (AppModule) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
+func (AppModule) RegisterInterfaces(registry registry.InterfaceRegistrar) {
```

The signature of the extension interface `HasAminoCodec` has been changed to accept a `cosmossdk.io/core/legacy.Amino` instead of a `codec.LegacyAmino`. Modules should update their `HasAminoCodec` implementation to accept a `cosmossdk.io/core/legacy.Amino` interface.

```diff
-func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
+func (AppModule) RegisterLegacyAminoCodec(cdc legacy.Amino) {
```

##### Simulation

`MsgSimulatorFn` has been updated to return an error. Its context argument has been removed, and an address.Codec has
been added to avoid the use of the Accounts.String() method.

```diff
-type MsgSimulatorFn func(r *rand.Rand, ctx sdk.Context, accs []Account) sdk.Msg
+type MsgSimulatorFn func(r *rand.Rand, accs []Account, cdc address.Codec) (sdk.Msg, error)
```

##### Depinject

Previously `cosmossdk.io/core` held functions `Invoke`, `Provide` and `Register` were moved to `cosmossdk.io/depinject/appconfig`.
All modules using dependency injection must update their imports.

##### Params

Previous module migrations have been removed. It is required to migrate to v0.50 prior to upgrading to v0.51 for not missing any module migrations.

##### Genesis Interface

All genesis interfaces have been migrated to take `context.Context` instead of `sdk.Context`.
Secondly, the codec is no longer passed in by the framework. The codec is now passed in by the module.
Lastly, all InitGenesis and ExportGenesis functions now return an error.

```go
// InitGenesis performs genesis initialization for the module.
func (am AppModule) InitGenesis(ctx context.Context, data json.RawMessage) error {
}

// ExportGenesis returns the exported genesis state as raw bytes for the module.
func (am AppModule) ExportGenesis(ctx context.Context) (json.RawMessage, error) {
}
```

##### Migration to Collections

Most of Cosmos SDK modules have migrated to [collections](https://docs.cosmos.network/main/build/packages/collections).
Many functions have been removed due to this changes as the API can be smaller thanks to collections.
For modules that have migrated, verify you are checking against `collections.ErrNotFound` when applicable.

#### `x/accounts`

Accounts's AccountNumber will be used as a global account number tracking replacing Auth legacy AccountNumber. Must set accounts's AccountNumber with auth's AccountNumber value in upgrade handler. This is done through auth keeper MigrateAccountNumber function.

```go
import authkeeper "cosmossdk.io/x/auth/keeper" 
...
err := authkeeper.MigrateAccountNumberUnsafe(ctx, &app.AuthKeeper)
if err != nil {
	return nil, err
}
```

#### `x/auth`

Auth was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/auth`

#### `x/authz`

Authz was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/authz`

#### `x/bank`

Bank was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/bank`

### `x/crsis`

The Crisis module was removed due to it not being supported or functional any longer. 

#### `x/distribution`

Distribution was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/distribution`

The existing chains using x/distribution module needs to add the new x/protocolpool module.

#### `x/group`

Group was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/group`

#### `x/gov`

Gov was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/gov`

Gov v1beta1 proposal handler has been changed to take in a `context.Context` instead of `sdk.Context`.
This change was made to allow legacy proposals to be compatible with server/v2.
If you wish to migrate to server/v2, you should update your proposal handler to take in a `context.Context` and use services.
On the other hand, if you wish to keep using baseapp, simply unwrap the sdk context in your proposal handler.

#### `x/mint`

Mint was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/mint`

#### `x/slashing`

Slashing was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/slashing`

#### `x/staking`

Staking was spun out into its own `go.mod`. To import it use `cosmossdk.io/x/staking`

#### `x/params`

A standalone Go module was created and it is accessible at "cosmossdk.io/x/params".

#### `x/protocolpool`

Introducing a new `x/protocolpool` module to handle community pool funds. Its store must be added while upgrading to v0.51.x.

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
>>>>>>> 651868a17 (docs: rename app v2 to app di when talking about runtime v0 (#21329))

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
and `VerifyVoteExtensionHandler` respectively. Please see [here](https://docs.cosmos.network/v0.50/build/building-apps/vote-extensions)
for more info.

#### Set PreBlocker

A `SetPreBlocker` method has been added to BaseApp. This is essential for BaseApp to run `PreBlock` which runs before begin blocker other modules, and allows to modify consensus parameters, and the changes are visible to the following state machine logics.
Read more about other use cases [here](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-068-preblock.md).

`depinject` / app di users need to add `x/upgrade` in their `app_config.go` / `app.yml`:

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

**Users using `depinject` / app di do not need any changes, this is abstracted for them.**

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

Additionally, `depinject` / app di users must now supply a logger through the main `depinject.Supply` function instead of passing it to `appBuilder.Build`.

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

This is automatically done for `depinject` / app di users, however for supplying different app module implementation, pass them via `depinject.Supply` in the main `AppConfig` (`app_config.go`):

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
