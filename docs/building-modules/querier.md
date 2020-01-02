<!--
order: 5
synopsis: "A `Querier` designates a function that processes [`queries`](./messages-and-queries.md#queries). `querier`s are specific to the module in which they are defined, and only process `queries` defined within said module. They are called from `baseapp`'s [`Query` method](../core/baseapp.md#query)."
-->

# Queriers

## Pre-requisite Readings {hide}

- [Module Manager](./module-manager.md) {prereq}
- [Messages and Queries](./messages-and-queries.md) {prereq}

## `Querier` type

The `querier` type defined in the Cosmos SDK specifies the typical structure of a `querier` function:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/queryable.go#L6

Let us break it down:

- The `path` is an array of `string`s that contains the type of the query, and that can also contain `query` arguments. See [`queries`](./messages-and-queries.md#queries) for more information.
- The `req` itself is primarily used to retrieve arguments if they are too large to fit in the `path`. This is done using the `Data` field of `req`. 
- The [`Context`](../core/context.md) contains all the necessary information needed to process the `query`, as well as a cache-wrapped copy of the latest state. It is primarily used by the [`keeper`](./keeper.md) to access the state. 
- The result `res` returned to `baseapp`, marhsalled using the application's [`codec`](../core/encoding.md). 

## Implementation of a module `querier`s

Module `querier`s are typically implemented in a `./internal/keeper/querier.go` file inside the module's folder. The [module manager](./module-manager.md) is used to add the module's `querier`s to the [application's `queryRouter`](../core/baseapp.md#query-routing) via the `NewQuerier()` method. Typically, the manager's `NewQuerier()` method simply calls a `NewQuerier()` method defined in `keeper/querier.go`, which looks like the following:

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

This simple switch returns a `querier` function specific to the type of the received `query`. At this point of the [query lifecycle](../interfaces/query-lifecycle.md), the first element of the `path` (`path[0]`) contains the type of the query. The following elements are either empty or contain arguments needed to process the query. 

The `querier` functions themselves are pretty straighforward. They generally fetch a value or values from the state using the [`keeper`](./keeper.md). Then, they marshall the value(s) using the [`codec`](../core/encoding.md) and return the `[]byte` obtained as result. 

For a deeper look at `querier`s, see this [example implementation of a `querier` function](https://github.com/cosmos/sdk-application-tutorial/blob/c6754a1e313eb1ed973c5c91dcc606f2fd288811/x/nameservice/internal/keeper/querier.go) from the nameservice tutorial. 

## Next {hide}

Learn about [`BeginBlocker` and `EndBlocker`](./beginblock-endblock.md) {hide}
