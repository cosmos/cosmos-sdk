# Upgrading Cosmos SDK

This guide provides instructions for upgrading to specific versions of Cosmos SDK.

## [Unreleased]

### Simulation

Remove `RandomizedParams` from `AppModuleSimulation` interface. Previously, it used to generate random parameter changes during simulations, however, it does so through ParamChangeProposal which is now legacy. Since all modules were migrated, we can now safely remove this from `AppModuleSimulation` interface.

### AppModule Interface

Remove `Querier`, `Route` and `LegacyQuerier` from the app module interface. This removes and fully deprecates all legacy queriers. All modules no longer support the REST API previously known as the LCD, and the `sdk.Msg#Route` method won't be used anymore.

### SimApp

The `simapp` package **should not be imported in your own app**. Instead, you should import the `runtime.AppI` interface, that defines an `App`, and use the [`simtestutil` package](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/testutil/sims) for application testing.

#### App Wiring

SimApp's `app.go` is now using [App Wiring](https://docs.cosmos.network/main/building-chain/depinject.html), the dependency injection framework of the Cosmos SDK.
This means that modules are injected directly into SimApp thanks to a [configuration file](https://github.com/cosmos/cosmos-sdk/blob/main/simapp/app_config.go).
The old behavior is preserved and can still be used, without the dependency injection framework, as shows [`app_legacy.go`](https://github.com/cosmos/cosmos-sdk/blob/main/simapp/app_legacy.go).

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

### Protobuf

The SDK has migrated from `gogo/protobuf` (which is currently unmaintained), to our own maintained fork, [`cosmos/gogoproto`](https://github.com/cosmos/gogoproto).

This means you should replace all imports of `github.com/gogo/protobuf` to `github.com/cosmos/gogoproto`.
This allows you to remove the replace directive `replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1` from your `go.mod` file.

Please use the `ghcr.io/cosmos/proto-builder` image (version >= `0.11.0`) for generating protobuf files.

### Transactions

#### Broadcast Mode

Broadcast mode `block` was deprecated and has been removed. Please use `sync` mode
instead. When upgrading your tests from `block` to `sync` and checking for a
transaction code, you need to query the transaction first (with its hash) to get
the correct code.

### Modules

#### `x/gov`

##### Minimum Proposal Deposit At Time of Submission

The `gov` module has been updated to support a minimum proposal deposit at submission time. It is determined by a new
parameter called `MinInitialDepositRatio`. When multiplied by the existing `MinDeposit` parameter, it produces
the necessary proportion of coins needed at the proposal submission time. The motivation for this change is to prevent proposal spamming.

By default, the new `MinInitialDepositRatio` parameter is set to zero during migration. The value of zero signifies that this 
feature is disabled. If chains wish to utilize the minimum proposal deposits at time of submission, the migration logic needs to be 
modified to set the new parameter to the desired value.

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

The old params module is required to still be imported in your app.go in order to handle this migration. 

##### App.go Changes

Previous:

```go
bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable()))
```

After:

```go
app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, keys[upgradetypes.StoreKey], authtypes.NewModuleAddress(govtypes.ModuleName).String())
bApp.SetParamStore(&app.ConsensusParamsKeeper)
```

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
        * `MkConsKeyOutput`
        * `MkValKeyOutput`
        * `MkAccKeyOutput`
* Rename:
        * `SavePubKey` to `SaveOfflineKey` and remove the `algo` argument.
        * `NewMultiInfo`, `NewLedgerInfo`  to `NewLegacyMultiInfo`, `newLegacyLedgerInfo` respectively.
        * `NewOfflineInfo` to `newLegacyOfflineInfo` and move it to `migration_test.go`.

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

* The `x/param` module has been depreacted in favour of each module housing and providing way to modify their parameters. Each module that has parameters that are changable during runtime have an authority, the authority can be a module or user account. The Cosmos-SDK team recommends migrating modules away from using the param module. An example of how this could look like can be found [here](https://github.com/cosmos/cosmos-sdk/pull/12363). 
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

```proto
message MsgSetWithdrawAddress {
  option (cosmos.msg.v1.signer) = "delegator_address"; ++

  option (gogoproto.equal)           = false;
  option (gogoproto.goproto_getters) = false;

  string delegator_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string withdraw_address  = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

<!-- todo: cosmos.scalar types -->

When clients interract with a node they are required to set a codec in in the grpc.Dial. More information can be found in this [doc](https://docs.cosmos.network/v0.46/run-node/interact-node.html#programmatically-via-go).
