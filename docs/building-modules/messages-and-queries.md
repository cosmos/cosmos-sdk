<!--
order: 3
-->

# Messages and Queries

`Message`s and `Queries` are the two primary objects handled by modules. Most of the core components defined in a module, like `handler`s, `keeper`s and `querier`s, exist to process `message`s and `queries`. {synopsis}

## Pre-requisite Readings

- [Introduction to SDK Modules](./intro.md) {prereq}

## Messages

`Message`s are objects whose end-goal is to trigger state-transitions. They are wrapped in [transactions](../core/transactions.md), which may contain one or multiple of them. 

When a transaction is relayed from the underlying consensus engine to the SDK application, it is first decoded by [`baseapp`](../core/baseapp.md). Then, each `message` contained in the transaction is extracted and routed to the appropriate module via `baseapp`'s `router` so that it can be processed by the module's [`handler`](./handler.md). For a more detailed explanation of the lifecycle of a transaction, click [here](../basics/tx-lifecycle.md). 

Defining `message`s is the responsibility of module developers. Typically, they are defined as protobuf messages in a `proto/` directory (see more info about [conventions and naming](../core/encoding.md#faq)). The `message`'s definition usually includes a list of parameters needed to process the message that will be provided by end-users when they want to create a new transaction containing said `message`.

Here's an example of a protobuf message definition:

+++ https://github.com/cosmos/cosmos-sdk/blob/d55c1a26657a0af937fa2273b38dcfa1bb3cff9f/proto/cosmos/gov/v1beta1/tx.proto#L15-L27

The `Msg` is typically accompanied by a standard constructor function, that is called from one of the [module's interface](./module-interfaces.md). `message`s also need to implement the [`Msg`] interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/4a1b2fba43b1052ca162b3a1e0b6db6db9c26656/types/tx_msg.go#L10-L33

It extends `proto.Message` and contains the following methods:

- `Route() string`: Name of the route for this message. Typically all `message`s in a module have the same route, which is most often the module's name.
- `Type() string`: Type of the message, used primarly in [events](../core/events.md). This should return a message-specific `string`, typically the denomination of the message itself.
- `ValidateBasic() error`: This method is called by `baseapp` very early in the processing of the `message` (in both [`CheckTx`](../core/baseapp.md#checktx) and [`DeliverTx`](../core/baseapp.md#delivertx)), in order to discard obviously invalid messages. `ValidateBasic` should only include *stateless* checks, i.e. checks that do not require access to the state. This usually consists in checking that the message's parameters are correctly formatted and valid (i.e. that the `amount` is strictly positive for a transfer).
- `GetSignBytes() []byte`: Return the canonical byte representation of the message. Used to generate a signature. 
- `GetSigners() []AccAddress`: Return the list of signers. The SDK will make sure that each `message` contained in a transaction is signed by all the signers listed in the list returned by this method. 

See an example implementation of a `message` from the `gov` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/master/x/gov/types/msgs.go#L94-L136

## Queries

A `query` is a request for information made by end-users of applications through an interface and processed by a full-node. A `query` is received by a full-node through its consensus engine and relayed to the application via the ABCI. It is then routed to the appropriate module via `baseapp`'s `queryrouter` so that it can be processed by the module's query service (./query-services.md). For a deeper look at the lifecycle of a `query`, click [here](../interfaces/query-lifecycle.md). 

### gRPC Queries

Starting from v0.40, the prefered way to define queries is by using [Protobuf services](https://developers.google.com/protocol-buffers/docs/proto#services). A `Query` service should be created per module in `query.proto`. This service lists endpoints starting with `rpc`. 

Here's an example of such a `Query` service definition:

+++ https://github.com/cosmos/cosmos-sdk/blob/d55c1a26657a0af937fa2273b38dcfa1bb3cff9f/proto/cosmos/auth/v1beta1/query.proto#L12-L23

As `proto.Message`s, generated `Response` types implement by default `String()` method of [`fmt.Stringer`](https://golang.org/pkg/fmt/#Stringer).

A `RegisterQueryServer` method is also generated and should be used to register the module's query server in `RegisterQueryService` method from the [`AppModule` interface](./module-manager.md#appmodule).

### Legacy Queries

Before the introduction of Protobuf and gRPC in the SDK, there was usually no specific `query` object defined by module developers, contrary to `message`s. Instead, the SDK took the simpler approach of using a simple `path` to define each `query`. The `path` contains the `query` type and all the arguments needed in order to process it. For most module queries, the `path` should look like the following:

```
queryCategory/queryRoute/queryType/arg1/arg2/...
```

where:

- `queryCategory` is the category of the `query`, typically `custom` for module queries. It is used to differentiate between different kinds of queries within `baseapp`'s [`Query` method](../core/baseapp.md#query).
- `queryRoute` is used by `baseapp`'s [`queryRouter`](../core/baseapp.md#query-routing) to map the `query` to its module. Usually, `queryRoute` should be the name of the module.
- `queryType` is used by the module's [`querier`](./query-services.md#legacy-queriers) to map the `query` to the appropriate `querier function` within the module. 
- `args` are the actual arguments needed to process the `query`. They are filled out by the end-user. Note that for bigger queries, you might prefer passing arguments in the `Data` field of the request `req` instead of the `path`. 

The `path` for each `query` must be defined by the module developer in the module's [command-line interface file](./module-interfaces.md#query-commands).Overall, there are 3 mains components module developers need to implement in order to make the subset of the state defined by their module queryable:

- A [`querier`](./query-services.md#legacy-queriers), to process the `query` once it has been [routed to the module](../core/baseapp.md#query-routing). 
- [Query commands](./module-interfaces.md#query-commands) in the module's CLI file, where the `path` for each `query` is specified. 
- `query` return types. Typically defined in a file `types/querier.go`, they specify the result type of each of the module's `queries`. These custom types must implement the `String()` method of [`fmt.Stringer`](https://golang.org/pkg/fmt/#Stringer). 

### Store Queries

Store queries query directly for store keys. They use `clientCtx.QueryABCI(req abci.RequestQuery)` to return the full `abci.ResponseQuery` with inclusion Merkle proofs.

See following examples:

+++ https://github.com/cosmos/cosmos-sdk/blob/080fcf1df25ccdf97f3029b6b6f83caaf5a235e4/x/ibc/core/client/query.go#L36-L46

+++ https://github.com/cosmos/cosmos-sdk/blob/080fcf1df25ccdf97f3029b6b6f83caaf5a235e4/baseapp/abci.go#L722-L749

## Next {hide}

Learn about [`handler`s](./handler.md) {hide}