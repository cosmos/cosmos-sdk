<!--
order: 5
-->

# Query Services

A Protobuf Query service processes [`queries`](./messages-and-queries.md#queries). Query services are specific to the module in which they are defined, and only process `queries` defined within said module. They are called from `BaseApp`'s [`Query` method](../core/baseapp.md#query). {synopsis}

## Pre-requisite Readings

* [Module Manager](./module-manager.md) {prereq}
* [Messages and Queries](./messages-and-queries.md) {prereq}

## `Querier` type

The `querier` type defined in the Cosmos SDK will be deprecated in favor of [gRPC Services](#grpc-service). It specifies the typical structure of a `querier` function:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/types/queryable.go#L9

Let us break it down:

* The [`Context`](../core/context.md) contains all the necessary information needed to process the `query`, as well as a branch of the latest state. It is primarily used by the [`keeper`](./keeper.md) to access the state.
* The `path` is an array of `string`s that contains the type of the query, and that can also contain `query` arguments. See [`queries`](./messages-and-queries.md#queries) for more information.
* The `req` itself is primarily used to retrieve arguments if they are too large to fit in the `path`. This is done using the `Data` field of `req`.
* The result in `[]byte` returned to `BaseApp`, marshalled using the application's [`codec`](../core/encoding.md).

## Implementation of a module query service

### gRPC Service

When defining a Protobuf `Query` service, a `QueryServer` interface is generated for each module with all the service methods:

```go
type QueryServer interface {
	QueryBalance(context.Context, *QueryBalanceParams) (*types.Coin, error)
	QueryAllBalances(context.Context, *QueryAllBalancesParams) (*QueryAllBalancesResponse, error)
}
```

These custom queries methods should be implemented by a module's keeper, typically in `./keeper/grpc_query.go`. The first parameter of these methods is a generic `context.Context`, whereas querier methods generally need an instance of `sdk.Context` to read
from the store. Therefore, the Cosmos SDK provides a function `sdk.UnwrapSDKContext` to retrieve the `sdk.Context` from the provided
`context.Context`.

Here's an example implementation for the bank module:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/x/bank/keeper/grpc_query.go

## Next {hide}

Learn about [`BeginBlocker` and `EndBlocker`](./beginblock-endblock.md) {hide}
