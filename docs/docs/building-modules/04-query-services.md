---
sidebar_position: 1
---

# Query Services

:::note Synopsis
A Protobuf Query service processes [`queries`](./02-messages-and-queries.md#queries). Query services are specific to the module in which they are defined, and only process `queries` defined within said module. They are called from `BaseApp`'s [`Query` method](../core/00-baseapp.md#query).
:::

:::note

### Pre-requisite Readings

* [Module Manager](./01-module-manager.md)
* [Messages and Queries](./02-messages-and-queries.md)

:::

## `Querier` type

The `querier` type defined in the Cosmos SDK will be deprecated in favor of [gRPC Services](#grpc-service). It specifies the typical structure of a `querier` function:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/types/queryable.go#L9
```

Let us break it down:

* The [`Context`](../core/02-context.md) contains all the necessary information needed to process the `query`, as well as a branch of the latest state. It is primarily used by the [`keeper`](./06-keeper.md) to access the state.
* The `path` is an array of `string`s that contains the type of the query, and that can also contain `query` arguments. See [`queries`](./02-messages-and-queries.md#queries) for more information.
* The `req` itself is primarily used to retrieve arguments if they are too large to fit in the `path`. This is done using the `Data` field of `req`.
* The result in `[]byte` returned to `BaseApp`, marshalled using the application's [`codec`](../core/05-encoding.md).

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

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/x/bank/keeper/grpc_query.go
```

### Calling queries from the State Machine

The Cosmos SDK v0.47 introduces a new `cosmos.query.v1.module_query_safe` Protobuf annotation which is used to state that a query that is safe to be called from within the state machine, for example:

* a Keeper's query function can be called from another module's Keeper,
* ADR-033 intermodule query calls,
* CosmWasm contracts can also directly interact with these queries.

If the `module_query_safe` annotation set to `true`, it means:

* The query is deterministic: given a block height it will return the same response upon multiple calls, and doesn't introduce any state-machine breaking changes across SDK patch versions.
* Gas consumption never fluctuates across calls and across patch versions.

If you are a module developer and want to use `module_query_safe` annotation for your own query, you have to ensure the following things:

* the query is deterministic and won't introduce state-machine-breaking changes without coordinated upgrades
* it has its gas tracked, to avoid the attack vector where no gas is accounted for
 on potentially high-computation queries.
