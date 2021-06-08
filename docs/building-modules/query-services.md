<!--
order: 5
-->

# Query Services

A query service processes [`queries`](./messages-and-queries.md#queries). Query services are specific to the module in which they are defined, and only process `queries` defined within said module. They are called from `BaseApp`'s [`Query` method](../core/baseapp.md#query). {synopsis}

## Pre-requisite Readings

- [Module Manager](./module-manager.md) {prereq}
- [Messages and Queries](./messages-and-queries.md) {prereq}

## `Querier` type

The `querier` type defined in the Cosmos SDK will be deprecated in favor of [gRPC Services](#grpc-service). It specifies the typical structure of a `querier` function:

+++ https://github.com/cosmos/cosmos-sdk/blob/9a183ffbcc0163c8deb71c7fd5f8089a83e58f05/types/queryable.go#L9

Let us break it down:

- The `path` is an array of `string`s that contains the type of the query, and that can also contain `query` arguments. See [`queries`](./messages-and-queries.md#queries) for more information.
- The `req` itself is primarily used to retrieve arguments if they are too large to fit in the `path`. This is done using the `Data` field of `req`.
- The [`Context`](../core/context.md) contains all the necessary information needed to process the `query`, as well as a branch of the latest state. It is primarily used by the [`keeper`](./keeper.md) to access the state.
- The result `res` returned to `BaseApp`, marshalled using the application's [`codec`](../core/encoding.md).

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
from the store. Therefore, the SDK provides a function `sdk.UnwrapSDKContext` to retrieve the `sdk.Context` from the provided
`context.Context`.

Here's an example implementation for the bank module:

+++ https://github.com/cosmos/cosmos-sdk/blob/d55c1a26657a0af937fa2273b38dcfa1bb3cff9f/x/bank/keeper/grpc_query.go

### Legacy Queriers

Module legacy `querier`s are typically implemented in a `./keeper/querier.go` file inside the module's folder. The [module manager](./module-manager.md) is used to add the module's `querier`s to the [application's `queryRouter`](../core/baseapp.md#query-routing) via the `NewQuerier()` method. Typically, the manager's `NewQuerier()` method simply calls a `NewQuerier()` method defined in `keeper/querier.go`, which looks like the following:

```go
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryType1:
			return queryType1(ctx, path[1:], req, keeper)

		case QueryType2:
			return queryType2(ctx, path[1:], req, keeper)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}
	}
}
```

This simple switch returns a `querier` function specific to the type of the received `query`. At this point of the [query lifecycle](../basics/query-lifecycle.md), the first element of the `path` (`path[0]`) contains the type of the query. The following elements are either empty or contain arguments needed to process the query.

The `querier` functions themselves are pretty straighforward. They generally fetch a value or values from the state using the [`keeper`](./keeper.md). Then, they marshall the value(s) using the [`codec`](../core/encoding.md) and return the `[]byte` obtained as result. 

For a deeper look at `querier`s, see this [example implementation of a `querier` function](https://github.com/cosmos/cosmos-sdk/blob/7f59723d889b69ca19966167f0b3a7fec7a39e53/x/gov/keeper/querier.go) from the bank module. 

## Next {hide}

Learn about [`BeginBlocker` and `EndBlocker`](./beginblock-endblock.md) {hide}
