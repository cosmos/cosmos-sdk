<!--
order: 1
-->

# App Migration

The following document describes the changes to update your app and modules to use Cosmos SDK v0.40,
a.k.a. Stargate release. {synopsis}

## Update Tooling

Make sure to have the following dependencies before updating your app to v0.40:

- Go 1.15+
- Docker
- Node.js v12.0+ (optional, for generating Swagger docs)

A list of handy `make` commands are configured [here](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/Makefile#L355-L443). In general, your own app can use a similar Makefile to the Cosmos SDK's one. Specifically, these are some Makefile commands that might be useful for your own app:

- `proto-update-deps` - To download/update the required thirdparty `proto` definitions.
- `proto-gen` - To auto generate proto code.
- `proto-check-breaking` - To check proto breaking changes.
- `proto-format` - To format proto files.

## Updating Modules

This section outlines how to upgrade your module to v0.40. There is also a whole section of [building modules](../building-modules/README.md), please refer to it for more details.

### Protocol Buffers

As outlined in our [encoding guide](../core/encoding.md), one of the most significant improvements introduced in Cosmos SDK v0.40 is Protobuf. This means that instead of defining your serializable types using Go structs, you should define them as Protobuf messages.

The rule of thumb is that if you need to serialize a type (into binary or JSON), then it should be defined as a Protobuf message. This means that pure domain types can be kept as Go structs and interfaces. In practice, the three following categories of types must be converted to Protobuf message:

- client-facing types: `Msg`s, query requests and responses. This is because client will send these types over the wire to the node.
- types that are stored in state. This is because we store the binary representation of these types in state.
- genesis types. These are used when importing and exporting states during chain upgrades.

An example of type that is stored in state is [x/auth's](../../x/auth/spec/README.md) `BaseAccount` type. Its migration looks like:

```diff
// We were definining `MsgSend` as a Go struct in v0.39.
- // https://github.com/cosmos/cosmos-sdk/blob/v0.39.2/x/bank/internal/types/msgs.go#L12-L16
- type BaseAccount struct {
- 	Address       sdk.AccAddress `json:"address" yaml:"address"`
-	  Coins         sdk.Coins      `json:"coins" yaml:"coins"`
- 	PubKey        crypto.PubKey  `json:"public_key" yaml:"public_key"`
- 	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
- 	Sequence      uint64         `json:"sequence" yaml:"sequence"`
- }

// And it should be converted to a Protobuf message in v0.40.
+ // https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/proto/cosmos/auth/v1beta1/auth.proto#L13-L25
+ message BaseAccount {
+  string              address = 1;
+   google.protobuf.Any pub_key = 2
+       [(gogoproto.jsontag) = "public_key,omitempty", (gogoproto.moretags) = "yaml:\"public_key\""];
+   uint64 account_number = 3 [(gogoproto.moretags) = "yaml:\"account_number\""];
+   uint64 sequence       = 4;
+ }
}
```

In general, we recommend to put all the Protobuf definitions in a single directory `proto/`, as stated in [ADR-023](../architecture/adr-023-protobuf-naming.md). This ADR contains other useful information on naming conventions.

You might have noticed that the `PubKey` interface in v0.39's `BaseAccount` has been transformed into an `Any`. For migrating interfaces, we use Protobuf's `Any` message, which is a struct that can hold arbitrary content. Please refer to the [encoding FAQ](../core/encoding.md#faq) to learn how to achieve that.

Once all your Protobuf messages are defined, use the `make proto-gen` command defined in the [tooling section](#tooling) to generate Go structs. These structs will be generated in `*.pb.go` files. As a quick example, here's the generated Go struct for the Protobuf BaseAccount we defined above:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/x/auth/types/auth.pb.go#L28-L36

There might be back and forth removing old Go structs/interfaces and defining new Protobuf messages before your Go app compiles and before your tests pass.

### Create `Msg` and `Query` Services

Cosmos SDK v0.40 uses Protobuf services to define state transitions (`Msg`s) and state queries, please read [the building modules guide on those services](../building-modules/messages-and-queries.md) for an overview.

#### `Msg` Service

For migrating `Msg`s, the handler pattern (inside the `handler.go` file) is deprecated. You may still keep it if you wish to support `Msg`s defined in older versions of the SDK. However, it is strongly recommended to add a `Msg` service to your Protobuf files, and each old `Msg` should be converted to a service method. Taking [x/bank's](../../x/bank/spec/README.md) `MsgSend` as an example, we have a corresponding `cosmos.bank.v1beta1.Msg/Send` service method:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/proto/cosmos/bank/v1beta1/tx.proto#L10-L31

A state transition is therefore modelized as a Protobuf service method, with a method request, and an (optionally empty) method response.

After defining your `Msg` service, run the `make proto-gen` script again to generate `Msg` server interfaces. The naming of this interface follows a simple convention, for x/bank's `Msg` service we just defined, it will be called `MsgServer`. The implementation of this interface should follow exactly the implementation of the old `Msg` handlers, which, in most cases, defer the actual state transition logic to the [keeper](../building-modules/keeper.md). You may implement this `MsgServer` directly on the keeper, of you can also do it on a new struct (e.g. called `msgServer`) that references the module's keeper.

For more information, please check our [`Msg` service guide](../building-modules/msg-services.md).

#### `Query` Service

For migrating state queries, the querier pattern (inside the `querier.go` file) is deprecated. You may still keep this file to support legacy queries, but it is strongly recommended to use a Protobuf `Query` service to handle state queries.

Each query endpoint is now defined as a separate service method in the `Query` service. Still taking `x/bank` as an example, here are the queries to fetch an account's balances:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/proto/cosmos/bank/v1beta1/query.proto#L12-L23

Each query has its own `Request` and `Response` types. Please also note the `google.api.http` option (from [`grpc-gateway`](https://github.com/grpc-ecosystem/grpc-gateway)) on each service method. `grpc-gateway` is a tool that exposes `Query` service methods as REST endpoints.

After defining the `Query` Protobuf service, run the `make proto-gen` command to generate correspond interfaces. The interface that needs to be implemented is `QueryServer`. This interface can be implemented on the [keeper](../building-modules/keeper.md) directly, or on a struct (e.g. called `queryServer`) that references the module's keeper. The logic of the implementation (i.e. actually fetching from the module's store) can be deferred to the keeper.

Cosmos SDK v0.40 also comes with an efficient pagination, it now uses `Prefix` stores to query. There are 2 helpers for pagination, [`Paginate`](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/types/query/pagination.go#L40-L42), [`FilteredPaginate`](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/types/query/filtered_pagination.go#L9-L17).

For more information, please check our [`Query` service guide](../building-modules/query-services.md).

#### Wiring up `Msg` and `Query` Services

The `RegisterServices` method is newly added and registers module's `MsgServer` and gRPC's `QueryServer`. It should be implemented of all your modules.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/x/bank/module.go#L99-L103

If you wish to expose your `Query` endpoints as REST endpoints (see [`Query` Services](#query-services)), make sure to also implement the `RegisterGRPCGatewayRoutes` function:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/x/bank/module.go#L69-L72

### Codec

For registering module-specific types into the Amino codec, the `RegisterCodec(cdc *codec.Codec)` method has been renamed to `RegisterLegacyAminoCodec(cdc *codec.LegacyAmino)`. Similarly, the `codec.New()` has been renamed to `codec.NewLegacyAmino()`.

Moreover, a new `RegisterInterfaces` method has been added to all modules. This method should register all the interfaces that Protobuf messages implement, as well as the service `Msg`s used in the module. An example of implementation for x/bank is given below:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/x/bank/types/codec.go#L21-L34

### Keeper

The `Keeper` constructor now takes a `codec.Marshaler` instead of a concrete Amino codec. This marshaler is used to encode types as binary and save the bytes into the state. With an interface, you can define the codec to use (Amino or Protobuf) on an app level, and keepers will use the correct encoding library to encode state.

This is useful is you wish to update to SDK v0.40 without doing a chain upgrade with a genesis export/import.

As such, each module's AppModuleBasic now also takes a `codec.Marshaler` field:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/x/bank/module.go#L35-L38

### CLI

Each modules may optionally expose CLI commands, and some changes are needed in these commands.

First, `context.CLIContext` is renamed to `client.Context` and moved to `github.com/cosmos/cosmos-sdk/client`. The global `viper` usage is removed from client and is replaced with Cobra' `cmd.Flags()`. There are two helpers to read common flags for CLI txs and queries:

```go
clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
```

Some other flags helper functions are transformed:

- `flags.PostCommands(cmds ...*cobra.Command) []*cobra.Command` and `flags.GetCommands(...)` usage
  is now replaced by `flags.AddTxFlagsToCmd(cmd *cobra.Command)` and `flags.AddQueryFlagsToCmd(cmd *cobra.Command)`
  respectively.
- new CLI tx commands doesn't take `codec` as an input now, the `clientCtx` can be retrieved from the `cmd` itself.

```diff
// v0.39
- func SendTxCmd(cdc *codec.Codec) *cobra.Command {...}

// v0.40
+ func NewSendTxCmd() *cobra.Command {...}
```

Finally, once your [`Query` services](#query-service) are wired up, the CLI commands should prefer to communicate with the node via gRPC. The gist is to create a `Query` or `Msg` client using the command's `clientCtx`, and perform the request using Protobuf's generated code. An example for querying x/bank balances is given here:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/x/bank/client/cli/query.go#L66-L94

### Miscelleanous

A number of other smaller breaking changes are also noteworthy.

| Before                                                                        | After                                                                                                | Comment                                                                                                                                                                                                               |
| ----------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `alias.go` file                                                               | Removed                                                                                              | `alias` usage is removed, please see [#6311](https://github.com/cosmos/cosmos-sdk/issues/6311) for details.                                                                                                           |
| `DefaultGenesis()`                                                            | `DefaultGenesis(cdc codec.JSONMarshaler)`                                                            | `DefaultGenesis` takes a codec argument now                                                                                                                                                                           |
| `ValidateGenesis()`                                                           | `ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage)`       | `ValidateGenesis` now requires `Marshaler`, `TxEncodingConfig`, `json.RawMessage` as input.                                                                                                                           |
| `Route() string`                                                              | `Route() sdk.Route`                                                                                  | For legacy handlers, return type of `Route()` method is changed from `string` to `"github.com/cosmos/cosmos-sdk/types".Route`. It should return a `NewRoute()` which includes `RouterKey` and `NewHandler` as params. |
| `QuerierHandler`                                                              | `LegacyQuerierHandler`                                                                               | Simple rename.                                                                                                                                                                                                        |
| `InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate`   | InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate   | `InitGenesis` now takes a codec input.                                                                                                                                                                                |
| `ExportGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate` | ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate | `ExportGenesis` now takes a codec input.                                                                                                                                                                              |

## Updating Your App

For a reference implementation used for demo purposes, you can refer to the SDK's [SimApp]() for your app's migration. The most notable changes are described in this section.

### Creating Codecs

With the introduction of Protobuf, each app needs to define the encoding library (Amino or Protobuf) to be used throughout the app. There is a central struct, [`EncodingConfig`](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/simapp/params/encoding.go#L9-L11), which defines all information necessary for codecs. In your app, an example `EncodingConfig` with Protobuf as default codec might look like:

```go
// MakeEncodingConfig creates an EncodingConfig
func MakeEncodingConfig() params.EncodingConfig {
	amino := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	encodingConfig := params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          txCfg,
		Amino:             amino,
	}
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
  ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	return encodingConfig
}
```

These codecs are used to populate the following fields on your app (we are using SimApp for demo purposes):

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/simapp/app.go#L146-L153

As explained in the [modules migration section](#updating-modules), some functions and structs in modules require an additional `codec.Marshaler` argument. You should pass `app.appCodec` in these cases, and this will be the default codec used throughout the app.

### Registering Not-Module-Related Protobuf Services

We described in the [modules migration section](#updating-modules) `Query` and `Msg` services defined in each module. The SDK also exposes two more module-agnostic services:

- the [Tx Service](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/proto/cosmos/tx/v1beta1/service.proto), to perform operations on transactions,
- the [Tendermint service](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/proto/cosmos/base/tendermint/v1beta1/query.proto), to have a more idiomatic interface to the [Tendermint RPC](https://docs.tendermint.com/master/rpc/).

These services are optional, if you wish to use these two Protobuf services, or if you wish to add more module-agnostic Protobuf services, then they need to be added inside `app.go`:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/simapp/app.go#L577-L585

### Registering `grpc-gateway` Routes

The exising `RegisterAPIRoutes` method on the `app` only registers [Legacy API routes](../core/grpc_rest.md#legacy-rest-api-routes). If you are using `grpc-gateway` REST endpoints as described [above](#query-service), then these endpoints need to be wired up to a HTTP server:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/simapp/app.go#L555-L575

## Next {hide}

Learn how to perform a [chain upgrade](./chain-upgrade-guide-040.md) to 0.40.
