# Upgrading Cosmos SDK

This guide provides instructions for upgrading to specific versions of Cosmos SDK.

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

#### `x/authz`

* `authz.NewMsgGrant` `expiration` is now a pointer. When `nil` is used, then no expiration will be set (grant won't expire).
* `authz.NewGrant` takes a new argument: block time, to correctly validate expire time.

### gRPC

A new gRPC service, `proto/cosmos/base/node/v1beta1/query.proto`, has been introduced
which exposes various operator configuration. App developers should be sure to
register the service with the gRPC-gateway service via
`nodeservice.RegisterGRPCGatewayRoutes` in their application construction, which
is typically found in `RegisterAPIRoutes`.

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

### Protobuf

The `third_party/proto` folder that existed in [previous version](https://github.com/cosmos/cosmos-sdk/tree/v0.45.3/third_party/proto) now does not contains directly the [proto files](https://github.com/cosmos/cosmos-sdk/tree/release/v0.46.x/third_party/proto).

Instead, the SDK uses [`buf`](https://buf.build). Clients should have their own [`buf.yaml`](https://docs.buf.build/configuration/v1/buf-yaml) with `buf.build/cosmos/cosmos-sdk` as dependency, in order to avoid having to copy paste these files.

The protos can as well be downloaded using `buf export buf.build/cosmos/cosmos-sdk:8cb30a2c4de74dc9bd8d260b1e75e176 --output <some_folder>`.
