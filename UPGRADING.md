# Upgrading Cosmos SDK

This guide provides instructions for upgrading to specific versions of Cosmos SDK.

## v0.46

### Client Changes

### `x/gov` v1

The `gov` module has been greatly improved. The previous API has been moved to `v1beta1` while the new implementation is called `v1`.

In order to submit a proposal with `submit-proposal` you now need to pass a `proposal.json` file.
You can still use the old way by using `submit-legacy-proposal`. This is not recommended.
More information can be found in the gov module [client documentation](https://docs.cosmos.network/v0.46/modules/gov/07_client.html).

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

### new packages

Additionally, new packages have been introduced in order to further split the codebase. Aliases are available for a new API breaking migration, but it is encouraged to migrate to this new packages:

* `errors` should replace `types/errors` when registering errors or wrapping SDK errors.
* `math` contains the `Int` or `Uint` types that are used in the SDK.

### `x/authz`

* `authz.NewMsgGrant` `expiration` is now a pointer. When `nil` is used, then no expiration will be set (grant won't expire).
* `authz.NewGrant` takes a new argument: block time, to correctly validate expire time.

### State Machine Changes

#### PostHandler

`postHandler` is like an `antehandler`, but is run _after_ the `runMsgs` execution. It is in the same store branch that `runMsgs`, meaning that both `runMsgs` and `postHandler`. This allows to run a custom logic after the execution of the messages.
