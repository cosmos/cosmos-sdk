<!--
order: 1
-->

# App Migration

The following document describes the changes to update your app and modules to use Cosmos SDK v0.40,
a.k.a. Stargate release. {synopsis}

## Tooling

Make sure to have the following dependencies when updating your app to v0.40:

- Go 1.15+
- Docker
- Node.js v12.0+ (Optional, for generating docs)

A list of handy `make` commands are configured [here](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc5/Makefile#L355-L443), they will be useful for your own app.

- `proto-update-deps` - To download/update the required thirdparty `proto` definitions.
- `proto-gen` - To auto generate proto code.
- `proto-check-breaking` - To check proto breaking changes.
- `proto-gen-any` - To generate the SDK's custom wrapper for google.protobuf.Any. It should only be run manually when needed.
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

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40-rc5/proto/cosmos/bank/v1beta1/query.proto#L12-L23

Each query has its own `Request` and `Response` types.

After defining the `Query` Protobuf service, run the `make proto-gen` command to generate correspond interfaces. The interface that needs to be implemented is `QueryServer`. This interface can be implemented on the [keeper](../building-modules/keeper.md) directly, or on a struct (e.g. called `queryServer`) that references the module's keeper. The logic of the implementation (i.e. actually fetching from the module's store) can be deferred to the keeper.

For more information, please check our [`Query` service guide](../building-modules/query-services.md).

#### Wiring up `Msg` and `Query` Services

The `RegisterServices` method is newly added and registers module's `MsgServer` and gRPC's `QueryServer`. It should be implemented of all your modules.

+++ https://github.com/cosmos/cosmos-sdk/blob/7c1da3d9988c361d6165d26d33bed47352072366/x/bank/module.go#L99-L103

---

## Updating the App

## Updating a module to use Cosmos-SDK v0.40

This section covers the changes in modules from `v0.39.x` to `v0.40`.

### Miscelleanous

- `internal` package is removed and `types`, `keeper` are moved to module level
- `alias` usage is removed [#6311](https://github.com/cosmos/cosmos-sdk/issues/6311)

#### types/codec.go

- `codec.New()` is changed to `codec.NewLegacyAmino()`.
- Added `RegisterInterfaces` method in which we add `RegisterImplementations` and `RegisterInterfaces` based on msgs and interfaces in module.

  ```go
  func RegisterInterfaces(registry types.InterfaceRegistry) {
      registry.RegisterImplementations((*sdk.Msg)(nil),
          &MsgSend{},
          &MsgMultiSend{},
      )

      registry.RegisterInterface(
          "cosmos.bank.v1beta1.SupplyI",
          (*exported.SupplyI)(nil),
          &Supply{},
      )
  }
  ```

* `init()` changed from:
  ```go
  func init() {
  )
  }
  ```
  to:
  ```go
  func init() {
  )
  )
  }
  ```

#### Msgs

SDK now leverages protobuf service definitions for defining Msgs which will give us significant developer UX
improvements in terms of the code that is generated and the fact that return types will now be well defined.
`sdk.Msg`'s are now replaced by protobuf services. Now every sdk.Msg is defined as a protobuf service method.
Example:

```proto
package cosmos.bank;

service Msg {
  rpc Send(MsgSend) returns (MsgSendResponse);
}
```

```proto
message MsgSend {
  string   from_address                    = 1 [(gogoproto.moretags) = "yaml:\"from_address\""];
  string   to_address                      = 2 [(gogoproto.moretags) = "yaml:\"to_address\""];
  repeated cosmos.base.v1beta1.Coin amount = 3
      [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];
}

message MsgSendResponse { }
```

We use protobuf service definitions for defining Msgs as well as the code generated by them as a replacement for Msg handlers.

_Encoding:_
Currently, we are encoding Msgs as Any in Txs which involves packing the binary-encoded Msg with its type URL.

The type URL for MsgSend based on the proto3 spec is /cosmos.bank.MsgSubmitProposal.

The fully-qualified name for the SubmitProposal service method above (also based on the proto3 and gRPC specs) is
/cosmos.bank.Msg/Send which varies by a single / character. The generated .pb.go files for protobuf services
include names of this form and any compliant protobuf/gRPC code generator will generate the same name.
In order to encode service methods in transactions, we encode them as Anys in the same TxBody.messages field as
other Msgs. We simply set Any.type_url to the full-qualified method name (ex. /cosmos.bank.Msg/Send) and
set Any.value to the protobuf encoding of the request message (MsgSend in this case).

_Decoding:_
When decoding, TxBody.UnpackInterfaces will need a special case to detect if Any type URLs match the service method format (ex. /cosmos.gov.Msg/SubmitProposal) by checking for two / characters. Messages that are method names plus request parameters instead of a normal Any messages will get unpacked into the ServiceMsg struct:

type ServiceMsg struct {
// MethodName is the fully-qualified service name
MethodName string
// Request is the request payload
Request MsgRequest
}

#### Keeper

The `Keeper` constructor now takes a `codec.Marshaler` instead of a concrete Amino codec. The exact type provided is
specified by `ModuleCdc`.

#### module.go

- `type AppModuleBasic struct{}` is updated to:

  ```go
  type AppModuleBasic struct {
      cdc codec.Marshaler
  }
  ```

- `RegisterCodec(cdc *codec.Codec)` method is changed to `RegisterLegacyAminoCodec(cdc *codec.LegacyAmino)`
- Added `RegisterInterfaces` method which implements `AppModuleBasic` which takes one parameter of type
  `"github.com/cosmos/cosmos-sdk/codec/types".InterfaceRegistry`. This method is used for registering interface types of module.
  ```go
  func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
  // register all interfaces in module.
  types.RegisterInterfaces(registry) //module's types/codec.go
  }
  ``
  ```
- `DefaultGenesis()` takes codec input now
  ```go
  func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {}
  ```
- `ValidateGenesis` now requires `Marshaler`, `TxEncodingConfig`, `json.RawMessage` as input.
  ```go
  func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage) error {}
  ```
- `GetQueryCmd(cdc *codec.Codec)`,`GetTxCmd(cdc *codec.Codec)` is changed to `GetQueryCmd()`,`GetTxCmd()` respectively.
- Return type of `Route()` method which implements `AppModule` is changed from `string` to `"github.com/cosmos/cosmos-sdk/types".Route`. We will return a NewRoute which includes `RouterKey` and `NewHandler` as params.
  ```go
  func (am AppModule) Route() sdk.Route {
      return sdk.NewRoute(types.RouterKey, handler.NewHandler(am.keeper))
  }
  ```
- `QuerierHandler` is renamed to `LegacyQuerierHandler`.

  ```go
    func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {}
  ```

* `RegisterGRPCGatewayRoutes` is newly added for registering module's gRPC gateway routes with API Server.`RegisterQueryHandlerClient` is auto generated by proto code generator.
  ```go
  // RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the bank module.
  func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
      types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
  }
  ```
* `InitGenesis` and `ExportGenesis` require explicit codec input.
  ```go
  func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {}
  func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {}
  ```

### Querier

Stargate comes with new API service, gRPC. Module keeper implements the gRPC's `QueryServer` interface.
Every module now has a `keeper/grpc_query.go` which contains the querier implementations. All the module query services
are defined in module's `query.proto` file. Here's an example for defining the query service:

/_cosmos/bank/v1beta1/query.proto_/

```go
// Query defines the gRPC querier service.
service Query {
    // AllBalances queries the balance of all coins for a single account.
    rpc AllBalances(QueryAllBalancesRequest) returns (QueryAllBalancesResponse) {
        option (google.api.http).get = "/cosmos/bank/v1beta1/balances/{address}";
    }
  ...
}

// QueryAllBalancesRequest is the request type for the Query/AllBalances RPC method.
message QueryAllBalancesRequest {
  option (gogoproto.equal)           = false;
  option (gogoproto.goproto_getters) = false;

  // address is the address to query balances for.
  string address = 1;

  // pagination defines an optional pagination for the request.
  cosmos.base.query.v1beta1.PageRequest pagination = 2;
}

// QueryAllBalancesResponse is the response type for the Query/AllBalances RPC
// method.
message QueryAllBalancesResponse {
  // balances is the balances of all the coins.
  repeated cosmos.base.v1beta1.Coin balances = 1
      [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];

  // pagination defines the pagination in the response.
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```

`0.40` comes with an efficient querier pagination, it now uses `Prefix` stores to query.
There are 2 helpers for `pagination`, 1) `Paginate` 2) `FilteredPaginate`.

### Client

- `context.CLIContext` is renamed to `client.Context` and moved to `github.com/cosmos/cosmos-sdk/client`
- The global `viper` usage is removed from client and is replaced with Cobra' `cmd.Flags()`. There are two helpers
  to read common flags for CLI txs and queries.

```go
clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
```

- Flags helper functions `flags.PostCommands(cmds ...*cobra.Command) []*cobra.Command`, `flags.GetCommands(...)` usage
  is now replaced by `flags.AddTxFlagsToCmd(cmd *cobra.Command)` and `flags.AddQueryFlagsToCmd(cmd *cobra.Command)`
  respectively.
- New CLI tx commands doesn't take `codec` as an input now.

```go
// v0.39.x
func SendTxCmd(cdc *codec.Codec) *cobra.Command {
	...
}

// v0.40
func NewSendTxCmd() *cobra.Command {
	...
}
```

_Sample code for new tx command:_

```go
// NewSendTxCmd returns a CLI command handler for creating a MsgSend transaction.
func NewSendTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "send [from_key_or_address] [to_address] [amount]",
		Short: `Send funds from one account to another. Note, the'--from' flag is
ignored as it is implied from [from_key_or_address].`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])

			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			toAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgSend(clientCtx.GetFromAddress(), toAddr, coins)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
```

## Updating a cosmos-sdk chain to use SDK 0.40

This section covers the changes to `BaseApp` and related changes in `app.go`, `config`

#### BaseApp

`BaseApp` now has two new fields, `GRPCQueryRouter` and `MsgServiceRouter`

- `GRPCQueryRouter` routes ABCI Query requests to GRPC handlers.
- `GRPCQueryHandler` defines a function type which handles ABCI Query requests using gRPC

##### gRPC Router (baseapp/grpcrouter.go)

It has `GRPCQueryRouter` and `GRPCQueryHandler`. `GRPCQueryRouter` routes ABCI Query requests to respective GRPC
handlers and is used in abci `Route`. `GRPCQueryHandler` defines a function type which handles ABCI Query requests
using gRPC.

##### Server

`GRPCRouter` and `Telemetry` are added newly to `Server`.

```go
// Server defines the server's API interface.
type Server struct {
	Router     *mux.Router
	GRPCRouter *runtime.ServeMux
	ClientCtx  client.Context

	logger   log.Logger
	metrics  *telemetry.Metrics
	listener net.Listener
}
```

`CustomGRPCHeaderMatcher` is an interceptor for gRPC gateway requests. It is useful for mapping request headers to
GRPC metadata. HTTP headers that start with 'Grpc-Metadata-' are automatically mapped to gRPC metadata after
removing prefix 'Grpc-Metadata-'. We can use this CustomGRPCHeaderMatcher if headers don't start with `Grpc-Metadata-`.

- API is made `in-process` with the node now. Enabling/disabling the API server and Swagger can now be configured from `app.toml`
  Both legacy REST API and gRPC gateway API are using the same server. Swagger can be accessed via `{baseurl}/swagger/`

```yaml
...
[api]

# Enable defines if the API server should be enabled.
enable = true

# Swagger defines if swagger documentation should automatically be registered.
swagger = true

# Address defines the API server to listen on.
address = "tcp://0.0.0.0:1317"
...
```

#### REST Queries and Swagger Generation

[grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) is a project that translates REST calls into GRPC calls
using special annotations on service methods. Modules that want to expose REST queries should add
`google.api.http` annotations to their `rpc` methods

## Upgrading a live chain to v0.40

[How to upgrade a chain from `0.39` to `0.40`](./chain-upgrade-guide-040.md)

References:

- [ADR 019 - Protobuf State Encoding](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-019-protobuf-state-encoding.md)
- [ADR 020 - Protobuf Transaction Encoding](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-020-protobuf-transaction-encoding.md)
- [ADR 021 - Protobuf Query Encoding](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-021-protobuf-query-encoding.md)
- [ADR 023 - Protobuf Naming](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-023-protobuf-naming.md)
- [ADR 031 - Protobuf Msg Services](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-031-msg-service.md)

```

```
