# Upgrading Cosmos SDK

This guide provides instructions for upgrading to specific versions of Cosmos SDK.
Note, always read the **SimApp** section for more information on application wiring updates.

## [Unreleased]

<!-- ### BaseApp

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

To be able to simulate nested messages within a transaction, message types containing nested messages must implement the
`HasNestedMsgs` interface. This interface requires a single method: `GetMsgs() ([]sdk.Msg, error)`, which should return
the nested messages. By implementing this interface, the BaseApp can simulate these nested messages during
transaction simulation. -->

### Modules

#### `x/params`

The `x/params` module has been removed from the Cosmos SDK.  The following [migration](https://github.com/cosmos/cosmos-sdk/blob/828fcf2f05db0c4759ed370852b6dacc589ea472/x/mint/migrations/v2/migrate.go) 
and [PR](https://github.com/cosmos/cosmos-sdk/pull/12363) can be used as a reference for migrating a legacy params module to the supported module-managed params paradigm.

More information can be found in the [deprecation notice](https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md#xparams).

## [v0.52.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.52.0-beta.1)

Documentation to migrate an application from v0.50.x to server/v2 is available elsewhere.
It is additional to the changes described here.

### SimApp

In this section we describe the changes made in Cosmos SDK' SimApp.
**These changes are directly applicable to your application wiring.**
Please read this section first, but for an exhaustive list of changes, refer to the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.52.x/simapp/CHANGELOG.md).

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

##### gRPC Web

Grpc-web embedded client has been removed from the server. If you would like to use grpc-web, you can use the [envoy proxy](https://www.envoyproxy.io/docs/envoy/latest/start/start). Here's how to set it up:

<details>
<summary>Step by step guide</summary>

1. Install Envoy following the [official installation guide](https://www.envoyproxy.io/docs/envoy/latest/start/install).

2. Create an Envoy configuration file named `envoy.yaml` with the following content:

   ```yaml
	static_resources:
	listeners:
	- name: listener_0
		address:
		socket_address: { address: 0.0.0.0, port_value: 8080 }
		filter_chains:
		- filters:
		- name: envoy.filters.network.http_connection_manager
			typed_config:
			"@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
			codec_type: auto
			stat_prefix: ingress_http
			route_config:
				name: local_route
				virtual_hosts:
				- name: local_service
				domains: ["*"]
				routes:
				- match: { prefix: "/" }
					route:
					cluster: grpc_service
					timeout: 0s
					max_stream_duration:
						grpc_timeout_header_max: 0s
				cors:
					allow_origin_string_match:
					- prefix: "*"
					allow_methods: GET, PUT, DELETE, POST, OPTIONS
					allow_headers: keep-alive,user-agent,cache-control,content-type,content-transfer-encoding,custom-header-1,x-accept-content-transfer-encoding,x-accept-response-streaming,x-user-agent,x-grpc-web,grpc-timeout
					max_age: "1728000"
					expose_headers: custom-header-1,grpc-status,grpc-message
			http_filters:
			- name: envoy.filters.http.grpc_web
				typed_config:
				"@type": type.googleapis.com/envoy.extensions.filters.http.grpc_web.v3.GrpcWeb
			- name: envoy.filters.http.cors
				typed_config:
				"@type": type.googleapis.com/envoy.extensions.filters.http.cors.v3.Cors
			- name: envoy.filters.http.router
				typed_config:
				"@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
	clusters:
	- name: grpc_service
		connect_timeout: 0.25s
		type: logical_dns
		http2_protocol_options: {}
		lb_policy: round_robin
		load_assignment:
		cluster_name: cluster_0
		endpoints:
			- lb_endpoints:
				- endpoint:
					address:
					socket_address:
						address: 0.0.0.0
						port_value: 9090
   ```

   This configuration tells Envoy to listen on port 8080 and forward requests to your gRPC service on port 9090. Note that this configuration is a starting point and can be modified according to your specific needs and preferences. You may need to adjust ports, addresses, or add additional settings based on your particular setup and requirements.

3. Start your Cosmos SDK application, ensuring it's configured to serve gRPC on port 9090.

4. Start Envoy with the configuration file:

   ```bash
   envoy -c envoy.yaml
   ```

5. If Envoy starts successfully, you should see output similar to this:

   ```bash
   [2024-08-29 10:47:08.753][6281320][info][config] [source/common/listener_manager/listener_manager_impl.cc:930] all dependencies initialized. starting workers
   [2024-08-29 10:47:08.754][6281320][info][main] [source/server/server.cc:978] starting main dispatch loop
   ```

   This indicates that Envoy has started and is ready to proxy requests.

	<!-- markdown-link-check-disable -->
6. Update your client applications to connect to Envoy (http://localhost:8080 by default).
	<!-- markdown-link-check-enable -->

</details>

By following these steps, Envoy will handle the translation between gRPC-Web and gRPC, allowing your existing gRPC-Web clients to continue functioning without modifications to your Cosmos SDK application.

To test the setup, you can use a tool like [grpcurl](https://github.com/fullstorydev/grpcurl). For example:

```bash
grpcurl -plaintext localhost:8080 cosmos.base.tendermint.v1beta1.Service/GetLatestBlock
```

##### AnteHandlers

The `GasConsumptionDecorator` and `IncreaseSequenceDecorator` have been merged with the SigVerificationDecorator, so you'll
need to remove them both from your app.go code, they will yield to unresolvable symbols when compiling.

#### Unordered Transactions

The Cosmos SDK now supports unordered transactions. This means that transactions
can be executed in any order and doesn't require the client to deal with or manage
nonces. This also means the order of execution is not guaranteed.

Unordered transactions are automatically enabled when using `depinject` / app di, simply supply the `servertypes.AppOptions` in `app.go`:

```diff
	depinject.Supply(
+		// supply the application options
+		appOpts,
		// supply the logger
		logger,
	)
```

<details>
<summary>Step-by-step Wiring </summary>
If you are still using the legacy wiring, you must enable unordered transactions manually:

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

* Create or update the App's `Preblocker()` method to call the unordered tx
  manager's `OnNewBlock()` method.

	```go
	...
	app.SetPreblocker(app.PreBlocker)
	...

	func (app *SimApp) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		app.UnorderedTxManager.OnNewBlock(ctx.BlockTime())
		return app.ModuleManager.PreBlock(ctx, req)
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

</details>

To submit an unordered transaction, the client must set the `unordered` flag to
`true` and ensure a reasonable `timeout_height` is set. The `timeout_height` is
used as a TTL for the transaction and is used to provide replay protection. See
[ADR-070](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-070-unordered-transactions.md)
for more details.

#### Sign Mode Textual

With the split of `x/auth/tx/config` in two (x/auth/tx/config as depinject module for txconfig and tx options) and `x/validate`, sign mode textual is no more automatically configured when using runtime (it was previously the case).
For the same instructions than for legacy app wiring to enable sign mode textual (see in v0.50 UPGRADING documentation).

### Depinject `app_config.go` / `app.yml`

With the introduction of [environment in modules](#core-api), depinject automatically creates the environment for all modules.
Learn more about environment in the [core/appmodule/v2/environment.go](core/appmodule/v2/environment.go) file. The Environment struct provides essential services to modules including logging, event handling, gas metering, header access, routing, and both KV and memory store services. Given the fields of environment, this means runtime creates a kv store service for all modules by default. It can happen that some modules do not have a store necessary (such as `x/auth/tx` for instance). In this case, the store creation should be skipped in `app_config.go`:

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

All modules (expect `auth`) were spun out into their own `go.mod`. Replace their imports by `cosmossdk.io/x/{moduleName}`.

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

The signature of the extension interface `HasAminoCodec` has been changed to accept a `cosmossdk.io/core/registry.AminoRegistrar` instead of a `codec.LegacyAmino`. Modules should update their `HasAminoCodec` implementation to accept a `cosmossdk.io/core/registry.AminoRegistrar` interface.

```diff
-func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
+func (AppModule) RegisterLegacyAminoCodec(registrar registry.AminoRegistrar) {
```

##### Simulation

`MsgSimulatorFn` has been updated to return an error. Its context argument has been removed, and an address.Codec has
been added to avoid the use of the Accounts.String() method.

```diff
-type MsgSimulatorFn func(r *rand.Rand, ctx sdk.Context, accs []Account) sdk.Msg
+type MsgSimulatorFn func(r *rand.Rand, accs []Account, cdc address.Codec) (sdk.Msg, error)
```

The interface `HasProposalMsgs` has been renamed to `HasLegacyProposalMsgs`, as we've introduced a new simulation framework, simpler and easier to use, named [simsx](https://github.com/cosmos/cosmos-sdk/blob/main/simsx/README.md).

##### Depinject

Previously `cosmossdk.io/core` held functions `Invoke`, `Provide` and `Register` were moved to `cosmossdk.io/depinject/appconfig`.
All modules using dependency injection must update their imports.

##### Params

Previous module migrations have been removed. It is required to migrate to v0.50 prior to upgrading to v0.52 for not missing any module migrations.

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

#### `x/auth`

Vesting accounts messages (and CLIs) have been removed. Existing vesting accounts will keep working but no new vesting accounts can be created.
Use `x/accounts` lockup accounts or implement an `x/accounts` vesting account instead.

#### `x/accounts`

Accounts's AccountNumber will be used as a global account number tracking replacing Auth legacy AccountNumber. Must set accounts's AccountNumber with auth's AccountNumber value in upgrade handler. This is done through auth keeper MigrateAccountNumber function.

```go
import authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper" 
...
err := authkeeper.MigrateAccountNumberUnsafe(ctx, &app.AuthKeeper)
if err != nil {
	return nil, err
}
```

##### TX Decoder

In order to support x/accounts properly we need to init a `TxDecoder`, modify your `app.go`:

```diff
import (
+ 	txdecode "cosmossdk.io/x/tx/decode"
)
+	txDecoder, err := txdecode.NewDecoder(txdecode.Options{
+		SigningContext: signingCtx,
+		ProtoCodec:     appCodec,
+	})
+	if err != nil {
+		panic(err)
+	}
```

#### `x/crisis`

The `x/crisis` module was removed due to it not being supported or functional any longer. 

#### `x/distribution`

Existing chains using `x/distribution` module must add the new `x/protocolpool` module.

#### `x/gov`

Gov v1beta1 proposal handler has been changed to take in a `context.Context` instead of `sdk.Context`.
This change was made to allow legacy proposals to be compatible with server/v2.
If you wish to migrate to server/v2, you should update your proposal handler to take in a `context.Context` and use services.
On the other hand, if you wish to keep using baseapp, simply unwrap the sdk context in your proposal handler.

#### `x/mint`

The `x/mint` module has been updated to work with a mint function [`MintFn`](https://docs.cosmos.network/v0.52/build/modules/mint#mintfn).

When using the default inflation calculation function and runtime, no change is required. The depinject configuration of mint automatically sets it if none is provided. However, when not using runtime, the mint function must be set in on the mint keeper:

```diff
+ mintKeeper.SetMintFn(keeper.DefaultMintFn(types.DefaultInflationCalculationFn, stakingKeeper, mintKeeper))
```

#### `x/protocolpool`

Introducing a new `x/protocolpool` module to handle community pool funds. Its store must be added while upgrading to v0.52.x.

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

Add `x/protocolpool` store while upgrading to v0.52.x:

```go
storetypes.StoreUpgrades{
			Added: []string{
				protocolpooltypes.ModuleName,
			},
}
```

#### `x/validate`

Introducing `x/validate` a module that is solely used for registering default ante/post handlers and global tx validators when using runtime and runtime/v2. If you wish to set your custom ante/post handlers, no need to use this module.
You can however always extend them by adding extra tx validators (see `x/validate` documentation).

#### `tools/benchmark`

Introducing [`tools/benchmark`](https://github.com/cosmos/cosmos-sdk/tree/main/tools/benchmark) a Cosmos SDK module for benchmarking your chain. It is a standalone module that can be added to your chain to stress test it. This module should NOT be added in a production environment.

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
and `VerifyVoteExtensionHandler` respectively. Please see [here](https://docs.cosmos.network/v0.50/build/abci/vote-extensions)
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

More information about [confix](https://docs.cosmos.network/main/build/tooling/confix) and how to add it in your application binary in the [documentation](https://docs.cosmos.network/main/build/tooling/confix).

#### gRPC-Web

gRPC-Web is now listening to the same address and port as the gRPC Gateway API server (default: `localhost:1317`).
The possibility to listen to a different address has been removed, as well as its settings.
Use `confix` to clean-up your `app.toml`. A nginx (or alike) reverse-proxy can be set to keep the previous behavior.

#### Database Support

ClevelDB, BoltDB and BadgerDB are not supported anymore. To migrate from an unsupported database to a supported database please use a database migration tool.

### Protobuf

With the deprecation of the Amino JSON codec defined in [cosmos/gogoproto](https://github.com/cosmos/gogoproto) in favor of the protoreflect powered x/tx/aminojson codec, module developers are encouraged verify that their messages have the correct protobuf annotations to deterministically produce identical output from both codecs.

For core SDK types equivalence is asserted by generative testing of [SignableTypes](https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-beta.0/tests/integration/rapidgen/rapidgen.go#L102) in [TestAminoJSON_Equivalence](https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-beta.0/tests/integration/tx/aminojson/aminojson_test.go#L94).

Read more about the available annotations [here](https://docs.cosmos.network/v0.50/build/building-modules/protobuf-annotations).

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
