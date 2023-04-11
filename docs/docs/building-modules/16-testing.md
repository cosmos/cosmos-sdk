---
sidebar_position: 1
---

# Testing

The Cosmos SDK contains different types of [tests](https://martinfowler.com/articles/practical-test-pyramid.html).
These tests have different goals and are used at different stages of the development cycle.
We advice, as a general rule, to use tests at all stages of the development cycle.
It is adviced, as a chain developer, to test your application and modules in a similar way than the SDK.

The rationale behind testing can be found in [ADR-59](https://docs.cosmos.network/main/architecture/adr-059-test-scopes.html).

## Unit Tests

Unit tests are the lowest test category of the [test pyramid](https://martinfowler.com/articles/practical-test-pyramid.html).
All packages and modules should have unit test coverage. Modules should have their dependencies mocked: this means mocking keepers.

The SDK uses `mockgen` to generate mocks for keepers:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/scripts/mockgen.sh#L3-L6
```

You can read more about mockgen [here](https://github.com/golang/mock).

### Example

As an example, we will walkthrough the [keeper tests](https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/x/gov/keeper/keeper_test.go) of the `x/gov` module.

The `x/gov` module has a `Keeper` type requires a few external dependencies (ie. imports outside `x/gov` to work properly).

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/x/gov/keeper/keeper.go#L61-L65
```

In order to only test `x/gov`, we mock the [expected keepers](https://docs.cosmos.network/v0.46/building-modules/keeper.html#type-definition) and instantiate the `Keeper` with the mocked dependencies. Note that we may need to configure the mocked dependencies to return the expected values:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/x/gov/keeper/common_test.go#L67-L81
```

This allows us to test the `x/gov` module without having to import other modules.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/x/gov/keeper/keeper_test.go#L3-L35
```

We can test then create unit tests using the newly created `Keeper` instance.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/x/gov/keeper/keeper_test.go#L73-L91
```

## Integration Tests

Integration tests are at the second level of the [test pyramid](https://martinfowler.com/articles/practical-test-pyramid.html).
In the SDK, we locate our integration tests under [`/tests/integrations`](https://github.com/cosmos/cosmos-sdk/tree/main/tests/integration).

The goal of these integration tests is to test how a component interacts with other dependencies. Compared to unit tests, integration tests do not mock dependencies. Instead, they use the direct dependencies of the component. This differs as well from end-to-end tests, which test the component with a full application.

Integration tests interact with the tested module via the defined `Msg` and `Query` services. The result of the test can be verified by checking the state of the application, by checking the emitted events or the response. It is adviced to combine two of these methods to verify the result of the test.

The SDK provides small helpers for quickly setting up an integration tests. These helpers can be found at <https://github.com/cosmos/cosmos-sdk/blob/main/testutil/integration>.

### Example

```go reference
https://github.com/cosmos/cosmos-sdk/blob/29e22b3bdb05353555c8e0b269311bbff7b8deca/testutil/integration/example_test.go#L22-L89
```

## Deterministic and Regression tests	

Tests are written for queries in the Cosmos SDK which have `module_query_safe` Protobuf annotation.

Each query is tested using 2 methods:

* Use property-based testing with the [`rapid`](https://pkg.go.dev/pgregory.net/rapid@v0.5.3) library. The property that is tested is that the query response and gas consumption are the same upon 1000 query calls.
* Regression tests are written with hardcoded responses and gas, and verify they don't change upon 1000 calls and between SDK patch versions.

Here's an example of regression tests:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/tests/integration/bank/keeper/deterministic_test.go#L102-L115
```

## Simulations

Simulations uses as well a minimal application, built with [`depinject`](../packages/01-depinject.md):

:::note
You can as well use the `AppConfig` `configurator` for creating an `AppConfig` [inline](https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/x/slashing/app_test.go#L54-L62). There is no difference between those two ways, use whichever you prefer.
:::

Following is an example for `x/gov/` simulations:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/x/gov/simulation/operations_test.go#L292-L310
```

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/x/gov/simulation/operations_test.go#L69-L111
```

## End-to-end Tests

End-to-end tests are at the top of the [test pyramid](https://martinfowler.com/articles/practical-test-pyramid.html).
They must test the whole application flow, from the user perspective (for instance, CLI tests). They are located under [`/tests/e2e`](https://github.com/cosmos/cosmos-sdk/tree/main/tests/e2e).

<!-- @julienrbrt: makes more sense to use an app wired app to have 0 simapp dependencies -->
For that, the SDK is using `simapp` but you should use your own application (`appd`).
Here are some examples:

* SDK E2E tests: <https://github.com/cosmos/cosmos-sdk/tree/main/tests/e2e>.
* Cosmos Hub E2E tests: <https://github.com/cosmos/gaia/tree/main/tests/e2e>.
* Osmosis E2E tests: <https://github.com/osmosis-labs/osmosis/tree/main/tests/e2e>.

:::note warning
The SDK is in the process of creating its E2E tests, as defined in [ADR-59](https://docs.cosmos.network/main/architecture/adr-059-test-scopes.html). This page will eventually be updated with better examples.
:::

## Learn More

Learn more about testing scope in [ADR-59](https://docs.cosmos.network/main/architecture/adr-059-test-scopes.html).
