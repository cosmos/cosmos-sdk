<!--
order: 1
-->

# App and Modules Migration

The following document describes the changes to update your app and modules from Cosmos SDK v0.39 to v0.40,
a.k.a. Stargate release. {synopsis}

## Update Tooling

Make sure to have the following dependencies before updating your app to v0.40:

- Go 1.15+
- Docker
- Node.js v12.0+ (optional, for generating Swagger docs)

In Cosmos-SDK we manage the project using Makefile. Your own app can use a similar Makefile to the [Cosmos SDK's one](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/Makefile). More specifically, below are some _make_ commands that might be useful for your own app, related to the introduction of Protocol Buffers:

- `proto-update-deps` - To download/update the required thirdparty `proto` definitions.
- `proto-gen` - To auto generate proto code.
- `proto-check-breaking` - To check proto breaking changes.
- `proto-format` - To format proto files.

## Updating Modules

This section outlines how to upgrade your module to v0.40. There is also a whole section about [building modules](../building-modules/README.md) from scratch, it might serve as a useful guide.

### Protocol Buffers

As outlined in our [encoding guide](../core/encoding.md), one of the most significant improvements introduced in Cosmos SDK v0.40 is Protobuf.

The rule of thumb is that any object that needs to be serialized (into binary or JSON) must implement `proto.Message` and must be serializable into Protobuf format. The easiest way to do it is to use Protobuf type definition and `protoc` compiler to generate the structures and functions for you. In practice, the three following categories of types must be converted to Protobuf messages:

- client-facing types: `Msg`s, query requests and responses. This is because client will send these types over the wire to your app.
- objects that are stored in state. This is because the SDK stores the binary representation of these types in state.
- genesis types. These are used when importing and exporting state snapshots during chain upgrades.

Let's have a look at [x/auth's](../../x/auth/spec/README.md) `BaseAccount` objects, which are stored in a state. The migration looks like:

```diff
// We were definining `MsgSend` as a Go struct in v0.39.
- // https://github.com/cosmos/cosmos-sdk/blob/v0.39.2/x/bank/internal/types/msgs.go#L12-L16
- type BaseAccount struct {
- 	Address       sdk.AccAddress `json:"address" yaml:"address"`
-	Coins         sdk.Coins      `json:"coins" yaml:"coins"`
- 	PubKey        crypto.PubKey  `json:"public_key" yaml:"public_key"`
- 	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
- 	Sequence      uint64         `json:"sequence" yaml:"sequence"`
- }

// And it should be converted to a Protobuf message in v0.40.
+ // https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/proto/cosmos/auth/v1beta1/auth.proto#L13-L25
+ message BaseAccount {
+  string              address = 1;
+   google.protobuf.Any pub_key = 2
+       [(gogoproto.jsontag) = "public_key,omitempty", (gogoproto.moretags) = "yaml:\"public_key\""];
+   uint64 account_number = 3 [(gogoproto.moretags) = "yaml:\"account_number\""];
+   uint64 sequence       = 4;
+ }
}
```

In general, we recommend to put all the Protobuf definitions in your module's subdirectory under a root `proto/` folder, as described in [ADR-023](../architecture/adr-023-protobuf-naming.md). This ADR also contains other useful information on naming conventions.

You might have noticed that the `PubKey` interface in v0.39's `BaseAccount` has been transformed into an `Any`. For storing interfaces, we use Protobuf's `Any` message, which is a struct that can hold arbitrary content. Please refer to the [encoding FAQ](../core/encoding.md#faq) to learn how to handle interfaces and `Any`s.

Once all your Protobuf messages are defined, use the `make proto-gen` command defined in the [tooling section](#tooling) to generate Go structs. These structs will be generated into `*.pb.go` files. As a quick example, here is the generated Go struct for the Protobuf BaseAccount we defined above:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/x/auth/types/auth.pb.go#L28-L36

There might be some back and forth removing old Go structs/interfaces and defining new Protobuf messages before your Go app compiles and your tests pass.

### Create `Msg` and `Query` Services

Cosmos SDK v0.40 uses Protobuf services to define state transitions (`Msg`s) and state queries, please read [the building modules guide on those services](../building-modules/messages-and-queries.md) for an overview.

#### `Msg` Service

For migrating `Msg`s, the handler pattern (inside the `handler.go` file) is deprecated. You may still keep it if you wish to support `Msg`s defined in older versions of the SDK. However, it is strongly recommended to add a `Msg` service to your Protobuf files, and each old `Msg` should be converted into a service method. Taking [x/bank's](../../x/bank/spec/README.md) `MsgSend` as an example, we have a corresponding `cosmos.bank.v1beta1.Msg/Send` service method:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/proto/cosmos/bank/v1beta1/tx.proto#L10-L31

A state transition is therefore modelized as a Protobuf service method, with a method request, and an (optionally empty) method response.

After defining your `Msg` service, run the `make proto-gen` script again to generate `Msg` server interfaces. The name of this interface is simply `MsgServer`. The implementation of this interface should follow exactly the implementation of the old `Msg` handlers, which, in most cases, defers the actual state transition logic to the [keeper](../building-modules/keeper.md). You may implement a `MsgServer` directly on the keeper, or you can do it using a new struct (e.g. called `msgServer`) that references the module's keeper.

For more information, please check our [`Msg` service guide](../building-modules/msg-services.md).

#### `Query` Service

For migrating state queries, the querier pattern (inside the `querier.go` file) is deprecated. You may still keep this file to support legacy queries, but it is strongly recommended to use a Protobuf `Query` service to handle state queries.

Each query endpoint is now defined as a separate service method in the `Query` service. Still taking `x/bank` as an example, here are the queries to fetch an account's balances:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/proto/cosmos/bank/v1beta1/query.proto#L12-L23

Each query has its own `Request` and `Response` types. Please also note the `google.api.http` option (coming from [`grpc-gateway`](https://github.com/grpc-ecosystem/grpc-gateway)) on each service method. `grpc-gateway` is a tool that exposes `Query` service methods as REST endpoints. Adding this annotation will expose these endpoints not only as gRPC endpoints, but also as REST endpoints. An overview of gRPC versus REST can be found [here](../core/grpc_rest.md).

After defining the `Query` Protobuf service, run the `make proto-gen` command to generate corresponding interfaces. The interface that needs to be implemented by your module is `QueryServer`. This interface can be implemented on the [keeper](../building-modules/keeper.md) directly, or on a struct (e.g. called `queryServer`) that references the module's keeper. The logic of the implementation, namely the logic that fetches data from the module's store and performs unmarshalling, can be deferred to the keeper.

Cosmos SDK v0.40 also comes with an efficient pagination, it now uses `Prefix` stores to make queries. There are 2 helpers for pagination, [`Paginate`](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/types/query/pagination.go#L40-L42), [`FilteredPaginate`](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/types/query/filtered_pagination.go#L9-L17).

For more information, please check our [`Query` service guide](../building-modules/query-services.md).

#### Wiring up `Msg` and `Query` Services

We added a new `RegisterServices` method that registers a module's `Msg` service and a `Query` service. It should be implemented by all modules, using a `Configurator` object:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/x/bank/module.go#L99-L103

If you wish to expose your `Query` endpoints as REST endpoints (as proposed in the [`Query` Services paragraph](#query-services)), make sure to also implement the `RegisterGRPCGatewayRoutes` method:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/x/bank/module.go#L69-L72

### Codec

If you still use Amino (which is deprecated since Stargate), you must register related types using the `RegisterLegacyAminoCodec(cdc *codec.LegacyAmino)` method (previously it was called  `RegisterCodec(cdc *codec.Codec)`).

Moreover, a new `RegisterInterfaces` method has been added to the `AppModule` interface that all modules must implement. This method must register the interfaces that Protobuf messages implement, as well as the service `Msg`s used in the module. An example from x/bank is given below:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/x/bank/types/codec.go#L21-L34

### Keeper

The `Keeper` constructor now takes a `codec.Marshaler` instead of a concrete Amino codec. This `codec.Marshaler` is used to encode types as binary and save the bytes into the state. With an interface, you can define `codec.Marshaler` to be Amino or Protobuf on an app level, and keepers will use that encoding library to encode state. Please note that Amino is deprecated in v0.40, and we strongly recommend to use Protobuf. Internally, the SDK still uses Amino in a couple of places (legacy REST API, x/params, keyring...), but Amino is planned to be removed in a future release.

Keeping Amino for now can be useful if you wish to update to SDK v0.40 without doing a chain upgrade with a genesis migration. This will not require updating the stores, so you can migrate to Cosmos SDK v0.40 with less effort. However, as Amino will be removed in a future release, the chain migration with genesis export/import will need to be performed then.

Related to the keepers, each module's `AppModuleBasic` now also includes this `codec.Marshaler`:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/x/bank/module.go#L35-L38

### CLI

Each modules may optionally expose CLI commands, and some changes are needed in these commands.

First, `context.CLIContext` is renamed to `client.Context` and moved to `github.com/cosmos/cosmos-sdk/client`.

Second, the global `viper` usage is removed from client and is replaced with Cobra' `cmd.Flags()`. There are two helpers to read common flags for CLI txs and queries:

```go
clientCtx, err := client.GetClientQueryContext(cmd)
clientCtx, err := client.GetClientTxContext(cmd)
```

Some other flags helper functions are transformed: `flags.PostCommands(cmds ...*cobra.Command) []*cobra.Command` and `flags.GetCommands(...)` usage is now replaced by `flags.AddTxFlagsToCmd(cmd *cobra.Command)` and `flags.AddQueryFlagsToCmd(cmd *cobra.Command)` respectively.

Moreover, new CLI commands don't take any codec as input anymore. Instead, the `clientCtx` can be retrieved from the `cmd` itself using the `GetClient{Query,Tx}Context` function above, and the codec as `clientCtx.JSONMarshaler`.

```diff
// v0.39
- func SendTxCmd(cdc *codec.Codec) *cobra.Command {
- 	cdc.MarshalJSON(...)
- }

// v0.40
+ func NewSendTxCmd() *cobra.Command {
+	clientCtx, err := client.GetClientTxContext(cmd)
+	clientCtx.JSONMarshaler.MarshalJSON(...)
+}
```

Finally, once your [`Query` services](#query-service) are wired up, the CLI commands should preferably use gRPC to communicate with the node. The gist is to create a `Query` or `Msg` client using the command's `clientCtx`, and perform the request using Protobuf's generated code. An example for querying x/bank balances is given here:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/x/bank/client/cli/query.go#L66-L94

### Miscelleanous

A number of other smaller breaking changes are also noteworthy.

| Before                                                                        | After                                                                                                | Comment                                                                                                                                                                                                               |
| ----------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `alias.go` file                                                               | Removed                                                                                              | `alias` usage is removed, please see [#6311](https://github.com/cosmos/cosmos-sdk/issues/6311) for details.                                                                                                           |
| `codec.New()`                                                                 | `codec.NewLegacyAmino()`                                                                             | Simple rename.                                                                                                                                                                                                        |
| `DefaultGenesis()`                                                            | `DefaultGenesis(cdc codec.JSONMarshaler)`                                                            | `DefaultGenesis` takes a codec argument now                                                                                                                                                                           |
| `ValidateGenesis()`                                                           | `ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage)`       | `ValidateGenesis` now requires `Marshaler`, `TxEncodingConfig`, `json.RawMessage` as input.                                                                                                                           |
| `Route() string`                                                              | `Route() sdk.Route`                                                                                  | For legacy handlers, return type of `Route()` method is changed from `string` to `"github.com/cosmos/cosmos-sdk/types".Route`. It should return a `NewRoute()` which includes `RouterKey` and `NewHandler` as params. |
| `QuerierHandler`                                                              | `LegacyQuerierHandler`                                                                               | Simple rename.                                                                                                                                                                                                        |
| `InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate`   | InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate   | `InitGenesis` now takes a codec input.                                                                                                                                                                                |
| `ExportGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate` | ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate | `ExportGenesis` now takes a codec input.                                                                                                                                                                              |

## Updating Your App

For a reference implementation used for demo purposes, you can refer to the SDK's [SimApp](https://github.com/cosmos/cosmos-sdk/tree/v0.40.0-rc6/simapp) for your app's migration. The most important changes are described in this section.

### Creating Codecs

With the introduction of Protobuf, each app needs to define the encoding library (Amino or Protobuf) to be used throughout the app. There is a central struct, [`EncodingConfig`](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/simapp/params/encoding.go#L9-L11), which defines all information necessary for codecs. In your app, an example `EncodingConfig` with Protobuf as default codec might look like:

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

These codecs are used to populate the following fields on your app (again, we are using SimApp for demo purposes):

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/simapp/app.go#L146-L153

As explained in the [modules migration section](#updating-modules), some functions and structs in modules require an additional `codec.Marshaler` argument. You should pass `app.appCodec` in these cases, and this will be the default codec used throughout the app.

### Registering Non-Module Protobuf Services

We described in the [modules migration section](#updating-modules) `Query` and `Msg` services defined in each module. The SDK also exposes two more module-agnostic services:

- the [Tx Service](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/proto/cosmos/tx/v1beta1/service.proto), to perform operations on transactions,
- the [Tendermint service](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/proto/cosmos/base/tendermint/v1beta1/query.proto), to have a more idiomatic interface to the [Tendermint RPC](https://docs.tendermint.com/master/rpc/).

These services are optional, if you wish to use them, or if you wish to add more module-agnostic Protobuf services into your app, then you need to add them inside `app.go`:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/simapp/app.go#L577-L585

### Registering `grpc-gateway` Routes

The exising `RegisterAPIRoutes` method on the `app` only registers [Legacy API routes](../core/grpc_rest.md#legacy-rest-api-routes). If you are using `grpc-gateway` REST endpoints as described [above](#query-service), then these endpoints need to be wired up to a HTTP server:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/simapp/app.go#L555-L575

## Next {hide}

Learn how to perform a [chain upgrade](./chain-upgrade-guide-040.md) to 0.40.
